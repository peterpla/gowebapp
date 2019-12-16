package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	speech "cloud.google.com/go/speech/apiv1"
	"github.com/julienschmidt/httprouter"
	"github.com/spf13/viper"
	speechpb "google.golang.org/genproto/googleapis/cloud/speech/v1"

	"github.com/peterpla/lead-expert/pkg/adding"
	"github.com/peterpla/lead-expert/pkg/appengine"
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
		// log.Printf("%s.taskHandler - decoded request: %+v\n", sn, incomingRequest)

		// TODO: convert the media file if needed
		//
		// Overall flow:
		//   1. extract file URI from Request (only files in GCS buckets supported at this point)
		//   2. if needed, convert file to a supported format
		// 	 3. submit file to Speech-to-Text service
		//   4. update Request with the provided transcription(s)
		//   5. add Request to next queue in the pipeline
		//
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
		if len(incomingRequest.MediaFileURI) < 5 {
			log.Printf("%s.taskHandler, MediaFileURI too short: %q", sn, incomingRequest.MediaFileURI)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		var intro = incomingRequest.MediaFileURI[0:5]
		if intro != "gs://" {
			log.Printf("%s.taskHandler, only \"gs://\" URIs supported (temporary): %q", sn, incomingRequest.MediaFileURI)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var resp *speechpb.LongRunningRecognizeResponse
		if resp, err = submitGoogleSpeechToText(incomingRequest.MediaFileURI); err != nil {
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			return
		}

		newRequest := incomingRequest

		// save in the Request struct all alternative translations
		for _, result := range resp.Results {
			for _, alt := range result.Alternatives {
				temp := new(adding.RawResults)
				temp.Transcript = alt.Transcript
				temp.Confidence = alt.Confidence
				newRequest.RawTranscript = append(newRequest.RawTranscript, *temp)
			}
		}
		// log.Printf("%s.taskHandler, request %s, ML transcription: %+v\n", sn, newRequest.RequestID, newRequest.RawTranscript)

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

func submitGoogleSpeechToText(gcsURI string) (*speechpb.LongRunningRecognizeResponse, error) {
	// "Transcribing long audio files", https://cloud.google.com/speech-to-text/docs/async-recognize
	sn := cfg.ServiceName

	ctx := context.Background()
	client, err := speech.NewClient(ctx)
	if err != nil {
		log.Printf("%s.taskHandler, speech.NewClient() error: %v", sn, err)
		return nil, err
	}

	// Send the contents of the audio file with the encoding and
	// and sample rate information to be transcripted.
	// For MP3, DO NOT include Encoding or SampleRateHertz
	req := &speechpb.LongRunningRecognizeRequest{
		Config: &speechpb.RecognitionConfig{
			LanguageCode: "en-US",
			UseEnhanced:  true, // phone model requires enhanced service
			Model:        "phone_call",
			// Encoding:        speechpb.RecognitionConfig_LINEAR16,
			// SampleRateHertz: 48000,
		},
		Audio: &speechpb.RecognitionAudio{
			AudioSource: &speechpb.RecognitionAudio_Uri{Uri: gcsURI},
		},
	}

	op, err := client.LongRunningRecognize(ctx, req)
	if err != nil {
		log.Printf("%s.submitGoogleSpeechToText, error from LongRunningRecognize(req: %+v), error: %v", sn, req, err)
		return nil, err
	}
	resp, err := op.Wait(ctx)
	if err != nil {
		log.Printf("%s.submitGoogleSpeechToText, Wait() error: %v", sn, err)
		return nil, err
	}
	// log.Printf("%s.submitGoogleSpeechToText, resp.Results: %+v", sn, resp.Results)

	return resp, nil
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
