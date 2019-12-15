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
	"github.com/peterpla/lead-expert/pkg/config"
	"github.com/peterpla/lead-expert/pkg/middleware"
	"github.com/peterpla/lead-expert/pkg/serviceInfo"
)

var Config config.Config

func init() {
	logPrefix := "transcription-gcp.main.init(),"
	if err := config.GetConfig(&Config); err != nil {
		msg := fmt.Sprintf(logPrefix+" GetConfig error: %v", err)
		panic(msg)
	}
	// log.Printf(logPrefix+" Config: %+v", Config)
}

func main() {
	// Creating App Engine task handlers: https://cloud.google.com/tasks/docs/creating-appengine-handlers
	// log.Printf("Enter transcription-gcp.main, Config: %+v\n", Config)

	// set ServiceName and QueueName appropriately
	prefix := "TaskTranscriptionGCP"
	Config.ServiceName = viper.GetString(prefix + "SvcName")
	Config.QueueName = viper.GetString(prefix + "WriteToQ")
	Config.NextServiceName = viper.GetString(prefix + "NextSvcToHandleReq")

	// make ServiceName and QueueName available to other packages
	serviceInfo.RegisterServiceName(Config.ServiceName)
	serviceInfo.RegisterQueueName(Config.QueueName)
	serviceInfo.RegisterNextServiceName(Config.NextServiceName)
	// log.Println(serviceInfo.DumpServiceInfo())

	router := httprouter.New()
	Config.Router = router

	// Default endpoint Cloud Tasks sends to is /task_handler
	router.POST("/task_handler", taskHandler(Config.Adder))

	// custom NotFound handler
	router.NotFound = http.HandlerFunc(myNotFound)

	// Allow confirmation the task handling service is running.
	router.GET("/", indexHandler)

	port := os.Getenv("PORT") // Google App Engine complains if "PORT" env var isn't checked
	if !Config.IsGAE {
		port = viper.GetString(prefix + "Port")
	}
	if port == "" {
		panic("PORT undefined")
	}

	log.Printf("Service %s listening on port %s, requests will be added to queue %s",
		Config.ServiceName, port, Config.QueueName)
	log.Fatal(http.ListenAndServe(":"+port, middleware.LogReqResp(router)))
}

// indexHandler responds to requests with "service running"
func indexHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	serviceName := Config.ServiceName
	// log.Printf("Enter %s.indexHandler\n", serviceName)
	if r.URL.Path != "/" {
		log.Printf("%s.indexHandler, r.URL.Path: %s, will respond NotFound\n", serviceName, r.URL.Path)
		http.NotFound(w, r)
		return
	}
	// indicate service is running
	fmt.Fprintf(w, "%q service running\n", serviceName)
}

// taskHandler processes task requests.
func taskHandler(a adding.Service) httprouter.Handle {
	serviceName := Config.ServiceName
	// log.Printf("%s.taskHandler - enter/exit\n", serviceName)

	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		// log.Printf("%s.taskHandler, request: %+v, params: %+v\n", serviceName, r, p)
		startTime := time.Now().UTC().Format(time.RFC3339Nano)

		// var taskName string
		t, ok := r.Header["X-Appengine-Taskname"]
		if !ok || len(t[0]) == 0 {
			// You may use the presence of the X-Appengine-Taskname header to validate
			// the request comes from Cloud Tasks.
			log.Printf("%s Invalid Task: No X-Appengine-Taskname request header found\n", serviceName)
			http.Error(w, "Bad Request - Invalid Task", http.StatusBadRequest)
			return
		}
		taskName := t[0]

		// Pull useful headers from Task request.
		q, ok := r.Header["X-Appengine-Queuename"]
		queueName := ""
		if ok {
			queueName = q[0]
		}

		// Extract the request body for further task details.
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("%s.main, ReadAll error: %v", serviceName, err)
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			return
		}
		// log.Printf("%s.taskHandler, body: %+v\n", serviceName, string(body))

		// decode incoming request
		var incomingRequest adding.Request

		decoder := json.NewDecoder(bytes.NewReader(body))
		err = decoder.Decode(&incomingRequest)
		if err != nil {
			log.Printf("%s.taskHandler, json.Decode error: %v", serviceName, err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// log.Printf("%s.taskHandler - decoded request: %+v\n", serviceName, incomingRequest)

		// ********** ********** ********** ********** **********

		// Overall flow:
		//   1. extract file URI from Request (only files in GCS buckets supported at this point)
		//   2. if needed, convert file to a supported format
		// 	 3. submit file to Speech-to-Text service
		//   4. update Request with the provided transcription(s)
		//   5. add Request to next queue in the pipeline

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

		// TODO: convert the media file if needed

		// !!! HACK !!! confirm audio file already in GCS bucket
		if len(incomingRequest.MediaFileURI) < 5 {
			log.Printf("%s.taskHandler, MediaFileURI too short: %q", serviceName, incomingRequest.MediaFileURI)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		var intro = incomingRequest.MediaFileURI[0:5]
		if intro != "gs://" {
			log.Printf("%s.taskHandler, only \"gs://\" URIs supported (temporary): %q", serviceName, incomingRequest.MediaFileURI)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		ctx := context.Background()
		client, err := speech.NewClient(ctx)
		if err != nil {
			log.Printf("%s.taskHandler, speech.NewClient() error: %v", serviceName, err)
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			return
		}

		var resp *speechpb.LongRunningRecognizeResponse
		if resp, err = submitGoogleSpeechToText(client, incomingRequest.MediaFileURI); err != nil {
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			return
		}

		// create new Request struct, add transcription results to it
		newRequest := incomingRequest

		// save in the Request struct all alternative translations
		for _, result := range resp.Results {
			for _, alt := range result.Alternatives {
				temp := new(adding.RawResults)
				temp.Transcript = alt.Transcript
				temp.Confidence = alt.Confidence
				newRequest.RawTranscript = append(newRequest.RawTranscript, *temp)
				// fmt.Fprintf(w, "\"%v\" (confidence=%3f)\n", alt.Transcript, alt.Confidence)
			}
		}
		// log.Printf("%s.taskHandler, request %s, ML transcription: %+v\n", serviceName, newRequest.RequestID, newRequest.RawTranscript)

		// add timestamps and get duration
		var duration time.Duration
		if duration, err = newRequest.AddTimestamps("BeginTranscriptionGCP", startTime, "EndTranscriptionGCP"); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// create task on the next pipeline stage's queue with updated Request
		a.AddRequest(newRequest)

		// Log details of the created task.
		output := fmt.Sprintf("%s.taskHandler completed in %v: queue %q, task %q, newRequest: %+v",
			serviceName, duration, queueName, taskName, newRequest)
		log.Println(output)

		// Report success/failure to Cloud Tasks.
		// Set a non-2xx status code to indicate a failure in task processing that should be retried.
		// For example, http.Error(w, "Internal Server Error: Task Processing", http.StatusInternalServerError)
		w.WriteHeader(http.StatusOK)
	}
}

func submitGoogleSpeechToText(client *speech.Client, gcsURI string) (*speechpb.LongRunningRecognizeResponse, error) {
	// "Transcribing long audio files", https://cloud.google.com/speech-to-text/docs/async-recognize
	serviceName := Config.ServiceName
	// log.Printf("%s.submitGoogleSpeechToText enter, gcsURI %q\n", serviceName, gcsURI)

	// Send the contents of the audio file with the encoding and
	// and sample rate information to be transcripted.
	// For MP3, DO NOT include Encoding or SampleRateHertz
	ctx := context.Background()
	req := &speechpb.LongRunningRecognizeRequest{
		Config: &speechpb.RecognitionConfig{
			// Encoding:        speechpb.RecognitionConfig_LINEAR16,
			// SampleRateHertz: 48000,
			LanguageCode: "en-US",
			// use phone model, requires enhanced service
			UseEnhanced: true,
			Model:       "phone_call",
		},
		Audio: &speechpb.RecognitionAudio{
			AudioSource: &speechpb.RecognitionAudio_Uri{Uri: gcsURI},
		},
	}

	op, err := client.LongRunningRecognize(ctx, req)
	if err != nil {
		log.Printf("%s.submitGoogleSpeechToText, error from LongRunningRecognize(req: %+v), error: %v", serviceName, req, err)
		return nil, err
	}
	resp, err := op.Wait(ctx)
	if err != nil {
		log.Printf("%s.submitGoogleSpeechToText, Wait() error: %v", serviceName, err)
		return nil, err
	}
	// log.Printf("%s.submitGoogleSpeechToText, resp.Results: %+v", serviceName, resp.Results)

	return resp, nil
}

func myNotFound(w http.ResponseWriter, r *http.Request) {
	// log.Printf("%s.myNotFound, request for %s not routed\n", serviceName, r.URL.Path)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	_, _ = w.Write([]byte("<h2>404 Not Foundw</h2>"))
}
