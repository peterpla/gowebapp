package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	speech "cloud.google.com/go/speech/apiv1"
	"github.com/go-playground/validator"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/spf13/viper"
	speechpb "google.golang.org/genproto/googleapis/cloud/speech/v1"

	"github.com/peterpla/lead-expert/pkg/appengine"
	"github.com/peterpla/lead-expert/pkg/check"
	"github.com/peterpla/lead-expert/pkg/config"
	"github.com/peterpla/lead-expert/pkg/middleware"
	"github.com/peterpla/lead-expert/pkg/queue"
	"github.com/peterpla/lead-expert/pkg/request"
	"github.com/peterpla/lead-expert/pkg/serviceInfo"
)

var prefix = "TaskTranscriptionGCP"
var logPrefix = "transcription-gcp.main.init(),"
var cfg config.Config
var q queue.Queue
var qi = queue.QueueInfo{}
var qs queue.QueueService

// use a single instance of Validate, it caches struct info
var validate *validator.Validate

const MAX_ALTERNATIVES = 2 // max alternatives from Google Speech-to-Text

func init() {
	if err := config.GetConfig(&cfg, prefix); err != nil {
		msg := fmt.Sprintf(logPrefix+" GetConfig error: %v", err)
		panic(msg)
	}

	// make ServiceName and QueueName available to other packages
	serviceInfo.RegisterServiceName(cfg.ServiceName)
	serviceInfo.RegisterQueueName(cfg.QueueName)
	serviceInfo.RegisterNextServiceName(cfg.NextServiceName)
}

func main() {
	// Creating App Engine task handlers: https://cloud.google.com/tasks/docs/creating-appengine-handlers

	if cfg.IsGAE {
		q = queue.NewGCTQueue(&qi) // use Google Cloud Tasks for queueing
	} else {
		q = queue.NewNullQueue(&qi) // use null queue, requests thrown away on exit
	}

	qs = queue.NewService(q)
	_ = qs

	router := httprouter.New()
	router.POST("/task_handler", taskHandler(q))
	router.GET("/", indexHandler)
	router.NotFound = http.HandlerFunc(myNotFound)
	cfg.Router = router

	port := os.Getenv("PORT") // Google App Engine complains if "PORT" env var isn't checked
	if !cfg.IsGAE {
		port = viper.GetString(prefix + "Port")
	}
	if port == "" {
		panic("PORT undefined")
	}

	validate = validator.New()

	log.Printf("Service %s listening on port %s, requests will be added to queue %s",
		cfg.ServiceName, port, cfg.QueueName)
	log.Fatal(http.ListenAndServe(":"+port, middleware.LogReqResp(router)))
}

// taskHandler processes task requests.
func taskHandler(q queue.Queue) httprouter.Handle {
	sn := cfg.ServiceName

	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		// log.Printf("%s.taskHandler, request: %+v, params: %+v\n", sn, r, p)
		startTime := time.Now().UTC().Format(time.RFC3339Nano)

		// pull task and queue names from App Engine headers
		taskName, queueName := appengine.GetAppEngineInfo(w, r)

		incomingRequest := request.Request{}
		if err := incomingRequest.ReadRequest(w, r, p, validate); err != nil {
			// ReadRequest called http.Error so we just return
			return
		}

		if cfg.IsGAE {
			// check for zero-value UUID, likely indicates failure to
			// proogate the Request object
			//
			// guard with IsGAE because local execution currently
			// does not propogate requests between services, so
			// although cmd/server/main.go assigns a UUID, all
			// subsequent services will see a zero-value UUID
			if err := check.RequestID(incomingRequest); err != nil {
				log.Printf("%s.main, check.RequestID error: %v", sn, err)
				http.Error(w, "Internal Error", http.StatusInternalServerError)
				return
			}
		}

		// log.Printf("%s.taskHandler - decoded request: %+v\n", sn, incomingRequest)

		var newRequest request.Request
		var err error

		// submit transcription request
		if newRequest, err = googleSpeechToText(incomingRequest); err != nil {
			log.Printf("%s.taskHandler, googleSpeechToText error: %v", sn, err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// add timestamps and get duration
		var duration time.Duration
		if duration, err = newRequest.AddTimestamps("BeginTranscriptionGCP", startTime, "EndTranscriptionGCP"); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// create task on the next pipeline stage's queue with updated Request
		if err := q.Add(&qi, &newRequest); err != nil {
			log.Printf("%s.taskHandler, q.Add error: +%v\n", sn, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// send response to Cloud Tasks
		w.WriteHeader(http.StatusOK)
		// Set a non-2xx status code to indicate a failure in task processing that should be retried.
		// For example, http.Error(w, "Internal Server Error: Task Processing", http.StatusInternalServerError)

		log.Printf("%s.taskHandler completed in %v: queue %q, task %q, newRequest: %+v",
			sn, duration, queueName, taskName, newRequest)
	}
}

// indexHandler responds to requests with "service running"
func indexHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	sn := cfg.ServiceName

	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// I'm not dead yet
	fmt.Fprintf(w, "%q service running\n", sn)
}

func myNotFound(w http.ResponseWriter, r *http.Request) {
	var msg404 = []byte("<h2>404 Not Foundw</h2>")

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	_, _ = w.Write(msg404)
}

// ********** ********** ********** ********** ********** **********

// ErrBadMediaFileURI
var ErrBadMediaFileURI = errors.New("Bad media_uri")

func googleSpeechToText(req request.Request) (request.Request, error) {
	// sn := serviceInfo.GetServiceName()
	// log.Printf("%s.googleSpeechToText, request: %+v\n", sn, req)

	var emptyRequest = request.Request{}
	var badRequest request.Request
	var err error

	// Overall flow:
	//   1. copy the media file to Google Cloud Storage (only files already in GCS buckets supported at this point)
	//   2. if needed, convert file to a supported format (only .MP3 files supported at this point)
	// 	 3. submit file to Speech-to-Text service
	//   4. capture the transcription for use by later pipeline stages

	if err := copyAndConvertMediaFile(req); err != nil {
		return badRequest, ErrBadMediaFileURI
	}

	// prepare the request
	ctx, client, gSTTreq, err := prepareGoogleSTTRequest(req.MediaFileURI)
	if err != nil {
		return emptyRequest, err
	}

	var resp *speechpb.LongRunningRecognizeResponse
	// submit the request, get the response
	if resp, err = getGoogleSTTResponse(ctx, client, gSTTreq); err != nil {
		return emptyRequest, err
	}

	// process the response, capture the working transcript for later pipeline stages
	newRequest := processTranscriptionResponse(req, resp)

	// log.Printf("%s.googleSpeechToText exiting, WorkingTranscript: %q, request %s\n", sn, newRequest.WorkingTranscript, newRequest.RequestID)

	return newRequest, nil
}

func prepareGoogleSTTRequest(gcsURI string) (context.Context, *speech.Client, *speechpb.LongRunningRecognizeRequest, error) {
	sn := serviceInfo.GetServiceName()
	// log.Printf("%s.prepareGoogleSTTRequest, URI: %s\n", sn, gcsURI)

	ctx := context.Background()
	client, err := speech.NewClient(ctx)
	if err != nil {
		log.Printf("%s.taskHandler, speech.NewClient() error: %v", sn, err)
		return nil, nil, nil, err
	}

	// "By using the [classes] in your recognition config, Cloud
	// Speech-to-Text is more likely to correctly transcribe audio
	// that includes [those classes]""
	phrases := []string{"$MONEY", "$MONTH", "$POSTALCODE", "$FULLPHONENUM"}
	speechContext := speechpb.SpeechContext{Phrases: phrases}

	// Send the contents of the audio file for transcription.
	req := &speechpb.LongRunningRecognizeRequest{
		Config: &speechpb.RecognitionConfig{
			// for MP3, DO NOT include Encoding or SampleRateHertz
			// Encoding:        speechpb.RecognitionConfig_LINEAR16,
			// SampleRateHertz: 48000,
			LanguageCode:    "en-US",
			UseEnhanced:     true, // phone model requires enhanced service
			Model:           "phone_call",
			MaxAlternatives: MAX_ALTERNATIVES,
			// adds punctuation to recognition result
			EnableAutomaticPunctuation: true,
			// recognize different speakers and what they say
			DiarizationConfig: &speechpb.SpeakerDiarizationConfig{
				EnableSpeakerDiarization: true,
			},
			SpeechContexts: []*speechpb.SpeechContext{
				&speechContext,
			},
		},
		Audio: &speechpb.RecognitionAudio{
			// where to find the audio file
			AudioSource: &speechpb.RecognitionAudio_Uri{Uri: gcsURI},
		},
	}

	return ctx, client, req, nil
}

func getGoogleSTTResponse(ctx context.Context, client *speech.Client, req *speechpb.LongRunningRecognizeRequest) (*speechpb.LongRunningRecognizeResponse, error) {
	// "Transcribing long audio files", https://cloud.google.com/speech-to-text/docs/async-recognize
	sn := serviceInfo.GetServiceName()

	op, err := client.LongRunningRecognize(ctx, req)
	if err != nil {
		log.Printf("%s.submitgoogleSpeechToText, error from LongRunningRecognize(req: %+v), error: %v", sn, req, err)
		return nil, err
	}
	resp, err := op.Wait(ctx)
	if err != nil {
		log.Printf("%s.submitgoogleSpeechToText, Wait() error: %v", sn, err)
		return nil, err
	}
	// log.Printf("%s.submitgoogleSpeechToText, resp.Results: %+v", sn, resp.Results)

	return resp, nil
}

// copyAndConvertMediaFile ensures the media file is available on Google Cloud Storage
func copyAndConvertMediaFile(req request.Request) error {
	sn := serviceInfo.GetServiceName()

	uri := req.MediaFileURI

	// TODO: copy file into GCS bucket

	// !!! HACK !!! confirm media file is already in GCS bucket
	if len(uri) < 5 || uri[0:5] != "gs://" {
		log.Printf("%s.copyAndConvertMediaFile, only \"gs://\" files supported (temporary): %q, RequestID: %s\n",
			sn, uri, req.RequestID.String())
		return ErrBadMediaFileURI
	}

	// TODO: convert the media file if needed

	// Libraries to investigate re: MP3 -> WAV
	//  https://github.com/giorgisio/goav - Golang bindings for FFmpeg
	//  https://www.ffmpeg.org/ffmpeg.html - ffmpeg is a very fast video and audio converter
	//
	//  https://github.com/nareix/joy4/cgo/ffmpeg - Golang audio/video library and streaming server
	//  https://github.com/xfrr/goffmpeg - FFMPEG wrapper written in GO
	//
	//  https://github.com/go-audio/examples/blob/master/format-converter/main.go - Generic Go package designed to
	//	define a common interface to analyze and/or process audio data
	//
	//	https://github.com/faiface/beep - A little package ... Suitable for playback and audio-processing.
	//
	// Potentially relevent write-ups:
	//  "Scalable Video Transcoding With App Engine Flexible",
	//	https://medium.com/google-cloud/scalable-video-transcoding-with-app-engine-flexible-621f6e7fdf56
	//
	//  "ffmpeg won't execute properly in google app engine standard nodejs",
	//  https://stackoverflow.com/questions/57350148/ffmpeg-wont-execute-properly-in-google-app-engine-standard-nodejs

	// !!! HACK !!! - only work with MP3 files

	// confirm filename ends in ".MP3" (case insensitive)
	if strings.ToLower(filepath.Ext(uri)) != ".mp3" {
		log.Printf("%s.copyAndConvertMediaFile, only \".MP3\" files supported (temporary): %q", sn, uri)
		return ErrBadMediaFileURI
	}

	return nil
}

// ********** ********** ********** ********** ********** **********

// rawWordInfo holds a copy of the WordInfo slice from STT
type rawWordInfo struct {
	wordInfo []*speechpb.WordInfo
}

// Transcript holds data during transcription processing
type Transcript struct {
	requestID         uuid.UUID
	mediaFileURI      string
	rawTranscript     []string
	rawConfidence     []float32
	rawWords          []rawWordInfo
	attributedStrings [][]string
	workingTranscript string
}

func processTranscriptionResponse(req request.Request, resp *speechpb.LongRunningRecognizeResponse) request.Request {
	sn := serviceInfo.GetServiceName()
	// log.Printf("%s.processTranscriptionResponse, request: %+v, LongRunningRecognizeResponse: %+v\n",
	// 	sn, req, resp)

	// modify a copy of the incoming request
	newRequest := req

	transcript := newTranscript(newRequest)

	// var transcript = Transcript{
	// 	requestID:    newRequest.RequestID,
	// 	mediaFileURI: newRequest.MediaFileURI,
	// }
	// // log.Printf("%s.processTranscriptionResponse, transcript: %+v\n", sn, transcript)
	// transcript.rawTranscript = make([]string, MAX_ALTERNATIVES+1)
	// transcript.rawConfidence = make([]float32, MAX_ALTERNATIVES+1)
	// transcript.rawWords = make([]rawWordInfo, MAX_ALTERNATIVES+1)
	// transcript.attributedStrings = make([][]string, 16*(MAX_ALTERNATIVES+1))
	// // log.Printf("%s.processTranscriptionResponse, transcript: %+v\n", sn, transcript)

	var a int
	var ac int // alternative count
	// var r int
	// var result *speechpb.SpeechRecognitionResult

	for _, result := range resp.Results {

		var alt *speechpb.SpeechRecognitionAlternative
		for a, alt = range result.Alternatives {
			if len(alt.GetWords()) == 0 {
				// alternatives with no WordInfo records are ignored
				continue
			}
			transcript.rawTranscript[a] = alt.GetTranscript()
			transcript.rawConfidence[a] = alt.GetConfidence()
			transcript.rawWords[a].wordInfo = alt.GetWords()
			transcript.attributedStrings[a] = wordsToAttributedStrings(transcript.rawWords[a].wordInfo)
			ac++

			// log.Printf("%s.processTranscriptionResponse, processed Results [%d] Alternative[%d], attributedStrings begins: %16q, Confidence: %6.4f, RequestID: %s\n",
			// 	sn, r, a, transcript.attributedStrings[a][0], transcript.rawConfidence[a], req.RequestID.String())
		}
	}

	transcript.workingTranscript = strings.Join(transcript.attributedStrings[0], "")
	newRequest.WorkingTranscript = transcript.workingTranscript

	// log.Printf("%s.processTranscriptionResponse, processed %d results, %d alternatives, WorkingTranscript begins: %16q, RequestID: %s\n",
	// 	sn, r, ac, newRequest.WorkingTranscript, req.RequestID.String())
	log.Printf("%s.processTranscriptionResponse, TODO: ==> persist transcript struct <===, RequestID: %s\n",
		sn, req.RequestID.String())

	return newRequest
}

func newTranscript(req request.Request) Transcript {
	var transcript = Transcript{
		requestID:    req.RequestID,
		mediaFileURI: req.MediaFileURI,
	}

	transcript.rawTranscript = make([]string, MAX_ALTERNATIVES+1)
	transcript.rawConfidence = make([]float32, MAX_ALTERNATIVES+1)
	transcript.rawWords = make([]rawWordInfo, MAX_ALTERNATIVES+1)
	transcript.attributedStrings = make([][]string, 16*(MAX_ALTERNATIVES+1))
	// log.Printf("%s.processTranscriptionResponse, transcript: %+v\n", sn, transcript)

	return transcript
}

func wordsToAttributedStrings(rawWords []*speechpb.WordInfo) []string {
	// sn := serviceInfo.GetServiceName()
	// log.Printf("%s.wordsToAttributedStrings, rawWords: %+v\n", sn, rawWords)

	// use | instead of \n to keep log entries cleaner
	// completionProcessing replaces | with \n
	const separator = "|"
	var initString = "[Speaker 1]"

	strings := []string{}
	emptyStrings := strings

	var speaker = 1
	var wordCount int

	var tmpString = initString
	for _, word := range rawWords {
		tmpWord := word.GetWord()
		tmpSpeaker := int(word.GetSpeakerTag())
		// tmpConfidence = word.GetConfidence()
		// tmpStart = word.GetStartTime()
		// tmpEnd = word.GetEndTime()

		if tmpSpeaker != speaker {
			// changed speakers - end the current string
			tmpString = tmpString + separator
			strings = append(strings, tmpString)
			// log.Printf("%s.transcriptionGCP.wordsToAttributedStrings, appended tmpString: %+v\n",
			// 	sn, tmpString)

			// reset tmpString, capture new speaker
			tmpString = fmt.Sprintf("[Speaker %d]", tmpSpeaker)
			speaker = tmpSpeaker
		}

		// POLICY: add space in front of the word we're adding (avoids
		// trailing spaces)
		tmpString = tmpString + " " + tmpWord
		wordCount++
	}

	if tmpString != initString {
		// end the final string, append it
		tmpString = tmpString + "\n"
		strings = append(strings, tmpString)
	} else {
		// didn't get any words from GetWord(), return empty []string
		strings = emptyStrings
	}

	// log.Printf("%s.transcriptionGCP.wordsToAttributedStrings, returning %d words: %q\n",
	// 	sn, wordCount, strings)

	return strings
}
