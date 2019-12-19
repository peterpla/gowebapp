package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	speech "cloud.google.com/go/speech/apiv1"
	"github.com/julienschmidt/httprouter"
	"github.com/spf13/viper"
	speechpb "google.golang.org/genproto/googleapis/cloud/speech/v1"

	"github.com/peterpla/lead-expert/pkg/adding"
	"github.com/peterpla/lead-expert/pkg/appengine"
	"github.com/peterpla/lead-expert/pkg/check"
	"github.com/peterpla/lead-expert/pkg/config"
	"github.com/peterpla/lead-expert/pkg/middleware"
	"github.com/peterpla/lead-expert/pkg/serviceInfo"
)

var prefix = "TaskTranscriptionGCP"
var logPrefix = "transcription-gcp.main.init(),"
var cfg config.Config

func init() {
	if err := config.GetConfig(&cfg); err != nil {
		msg := fmt.Sprintf(logPrefix+" GetConfig error: %v", err)
		panic(msg)
	}
	// set ServiceName and QueueName appropriately
	cfg.ServiceName = viper.GetString(prefix + "SvcName")
	cfg.QueueName = viper.GetString(prefix + "WriteToQ")
	cfg.NextServiceName = viper.GetString(prefix + "NextSvcToHandleReq")

	// make ServiceName and QueueName available to other packages
	serviceInfo.RegisterServiceName(cfg.ServiceName)
	serviceInfo.RegisterQueueName(cfg.QueueName)
	serviceInfo.RegisterNextServiceName(cfg.NextServiceName)

	config.SetConfigPointer(&cfg)
}

func main() {
	// Creating App Engine task handlers: https://cloud.google.com/tasks/docs/creating-appengine-handlers

	router := httprouter.New()
	router.POST("/task_handler", taskHandler(cfg.Adder))
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

	log.Printf("Service %s listening on port %s, requests will be added to queue %s",
		cfg.ServiceName, port, cfg.QueueName)
	log.Fatal(http.ListenAndServe(":"+port, middleware.LogReqResp(router)))
}

// taskHandler processes task requests.
func taskHandler(a adding.Service) httprouter.Handle {
	sn := cfg.ServiceName

	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		// log.Printf("%s.taskHandler, request: %+v, params: %+v\n", sn, r, p)
		startTime := time.Now().UTC().Format(time.RFC3339Nano)

		// pull task and queue names from App Engine headers
		taskName, queueName := appengine.GetAppEngineInfo(w, r)

		// Extract the request body for further task details.
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("%s.main, ReadAll error: %v", sn, err)
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			return
		}
		// log.Printf("%s.taskHandler, body: %+v\n", sn, string(body))

		// decode incoming request
		var incomingRequest adding.Request

		decoder := json.NewDecoder(bytes.NewReader(body))
		err = decoder.Decode(&incomingRequest)
		if err != nil {
			log.Printf("%s.taskHandler, json.Decode error: %v", sn, err)
			http.Error(w, err.Error(), http.StatusBadRequest)
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

		var newRequest adding.Request

		// submit transcription request
		if newRequest, err = googleSpeechToText(incomingRequest); err != nil {
			log.Printf("%s.taskHandler, googleSpeechToText error: %v", sn, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// add timestamps and get duration
		var duration time.Duration
		if duration, err = newRequest.AddTimestamps("BeginTranscriptionGCP", startTime, "EndTranscriptionGCP"); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// create task on the next pipeline stage's queue with updated Request
		a.AddRequest(newRequest)

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

func googleSpeechToText(req adding.Request) (adding.Request, error) {
	var badRequest adding.Request
	var err error

	sn := serviceInfo.GetServiceName()
	// log.Printf("%s.googleSpeechToText, request: %+v\n", sn, req)

	// Overall flow:
	//   1. extract file URI from Request (only files in GCS buckets supported at this point)
	//   2. if needed, convert file to a supported format
	// 	 3. submit file to Speech-to-Text service
	//   4. update Request with the provided transcription(s)
	//   5. add Request to next queue in the pipeline

	// TODO: convert the media file if needed
	//
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

	// TODO: copy file into GCS bucket

	// !!! HACK !!! confirm audio file already in GCS bucket
	if len(req.MediaFileURI) < 5 {
		log.Printf("%s.taskHandler, MediaFileURI too short: %q", sn, req.MediaFileURI)
		// http.Error(w, ErrBadMediaFileURI.Error(), http.StatusBadRequest)
		return badRequest, ErrBadMediaFileURI
	}
	var intro = req.MediaFileURI[0:5]
	if intro != "gs://" {
		log.Printf("%s.taskHandler, only \"gs://\" URIs supported (temporary): %q", sn, req.MediaFileURI)
		// http.Error(w, ErrBadMediaFileURI.Error(), http.StatusBadRequest)
		return badRequest, ErrBadMediaFileURI
	}

	var ctx context.Context
	var client *speech.Client
	var gSTTreq *speechpb.LongRunningRecognizeRequest

	ctx, client, gSTTreq, err = prepareGoogleSTT(req.MediaFileURI)
	if err != nil {
		log.Printf("%s.googleSpeechToText, prepareGoogleSTT error: %v", sn, err)
		// http.Error(w, ErrBadMediaFileURI.Error(), http.StatusBadRequest)
		return badRequest, ErrBadMediaFileURI
	}

	var resp *speechpb.LongRunningRecognizeResponse
	// func getGoogleSTTResponse(ctx context.Context, client *speech.Client, req *speechpb.LongRunningRecognizeRequest) (*speechpb.LongRunningRecognizeResponse, error) {
	if resp, err = getGoogleSTTResponse(ctx, client, gSTTreq); err != nil {
		// http.Error(w, "Internal Error", http.StatusInternalServerError)
		return badRequest, err
	}

	newRequest := processTranscription(req, resp)

	// log.Printf("%s.googleSpeechToText exiting, request %s, WorkingTranscript: %+v\n", sn, newRequest.RequestID, newRequest.WorkingTranscript)

	return newRequest, nil
}

func prepareGoogleSTT(gcsURI string) (context.Context, *speech.Client, *speechpb.LongRunningRecognizeRequest, error) {
	sn := serviceInfo.GetServiceName()
	// log.Printf("%s.prepareGoogleSTT, URI: %s\n", sn, gcsURI)

	ctx := context.Background()
	client, err := speech.NewClient(ctx)
	if err != nil {
		log.Printf("%s.taskHandler, speech.NewClient() error: %v", sn, err)
		return nil, nil, nil, err
	}

	// Send the contents of the audio file for transcription.
	req := &speechpb.LongRunningRecognizeRequest{
		Config: &speechpb.RecognitionConfig{
			// for MP3, DO NOT include Encoding or SampleRateHertz
			// Encoding:        speechpb.RecognitionConfig_LINEAR16,
			// SampleRateHertz: 48000,
			LanguageCode: "en-US",
			UseEnhanced:  true, // phone model requires enhanced service
			Model:        "phone_call",
			// adds punctuation to recognition result
			EnableAutomaticPunctuation: true,
			// recognize different speakers and what they say
			DiarizationConfig: &speechpb.SpeakerDiarizationConfig{
				EnableSpeakerDiarization: true,
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

func processTranscription(req adding.Request, resp *speechpb.LongRunningRecognizeResponse) adding.Request {
	// sn := serviceInfo.GetServiceName()
	// log.Printf("%s.processTranscription, request: %+v\n", sn, req)

	// modify a copy of the incoming request
	newRequest := req

	// save Alternatives in the Request struct all alternative translations
	for _, result := range resp.Results {
		for _, alt := range result.Alternatives {
			// log.Printf("%s.processTranscription, resp.Results.Alternative: %+v\n", sn, alt)
			tempResults := new(adding.RawResults)
			tempResults.Transcript = alt.GetTranscript()
			tempResults.Confidence = alt.GetConfidence()
			newRequest.RawTranscript = append(newRequest.RawTranscript, *tempResults)

			newRequest.RawWords = alt.GetWords()
			newRequest.AttributedStrings = wordsToAttributedStrings(newRequest.RawWords)
		}
	}
	// log.Printf("%s.processTranscription, request %s after ML transcription: %+v\n", sn, newRequest.RequestID, newRequest)

	persistAndTrim(&newRequest)

	return newRequest
}

func wordsToAttributedStrings(rawWords []*speechpb.WordInfo) []string {
	// sn := serviceInfo.GetServiceName()
	// log.Printf("%s.wordsToAttributedStrings, rawWords: %+v\n", sn, rawWords)

	// use | instead of \n to keep log entries cleaner
	// completionProcessing replaces | with \n
	const separator = "|"
	strings := []string{}

	var speaker = 1
	var tmpString = "[Speaker 1]"
	var wordCount int

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

		// policy: add space in front of the word we're adding (avoids
		// trailing spaces)
		tmpString = tmpString + " " + tmpWord
		wordCount++
	}

	// end the final string, append it
	tmpString = tmpString + "\n"
	strings = append(strings, tmpString)
	// log.Printf("%s.transcriptionGCP.wordsToAttributedStrings, appended tmpString: %q\n",
	// 	sn, tmpString)

	// log.Printf("%s.transcriptionGCP.wordsToAttributedStrings, returning %d words: %q\n",
	// 	sn, wordCount, strings)

	return strings
}

func persistAndTrim(req *adding.Request) {
	sn := serviceInfo.GetServiceName()
	// consolidate transcript into one string, WorkingTranscript
	req.WorkingTranscript = strings.Join(req.AttributedStrings, "")

	// TODO: persist fields that are about to be trimmed
	log.Printf("%s.persistAndTrim, TODO: ==> persist <=== before deleting from Request: RawTranscript, RawWords, AttributedStrings\n", sn)

	// trim (remove) fields in Request we no longer need
	var emptyRawTranscript = []adding.RawResults{}
	req.RawTranscript = emptyRawTranscript

	var emptyRawWords = []*speechpb.WordInfo{}
	req.RawWords = emptyRawWords

	var emptyAttributedStrings = []string{}
	req.AttributedStrings = emptyAttributedStrings

	// log.Printf("%s.persistAndTrim exiting, req: %+v\n", sn, req)
}
