package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-playground/validator"
	"github.com/julienschmidt/httprouter"
	"github.com/spf13/viper"

	"github.com/peterpla/lead-expert/pkg/appengine"
	"github.com/peterpla/lead-expert/pkg/config"
	"github.com/peterpla/lead-expert/pkg/middleware"
	"github.com/peterpla/lead-expert/pkg/queue"
	"github.com/peterpla/lead-expert/pkg/request"
	"github.com/peterpla/lead-expert/pkg/serviceInfo"
)

var prefix = "TaskCompletionProcessing"
var initLogPrefix = "completion-processing.main.init(),"
var cfg config.Config
var q queue.Queue
var qi = queue.QueueInfo{}
var qs queue.QueueService

// use a single instance of Validate, it caches struct info
var validate *validator.Validate

func init() {
	if err := config.GetConfig(&cfg, prefix); err != nil {
		msg := fmt.Sprintf(initLogPrefix+" GetConfig error: %v", err)
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
	router.POST("/task_handler", taskHandler()) // default endpoint Cloud Tasks POSTs to
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

// handler for Cloud Tasks POSTs
func taskHandler() httprouter.Handle {
	sn := cfg.ServiceName

	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		// log.Printf("%s.taskHandler, request: %+v, params: %+v\n", sn, r, p)
		startTime := time.Now().UTC()

		// pull task and queue names from App Engine headers
		taskName, queueName := appengine.GetAppEngineInfo(w, r)

		incomingRequest := request.Request{}
		if err := incomingRequest.ReadRequest(w, r, p, validate); err != nil {
			// ReadRequest called http.Error so we just return
			return
		}

		// TODO: establish what constitutes "completion processing"

		// replace | with \n in WorkingTranscript
		incomingRequest.FinalTranscript = strings.Replace(incomingRequest.WorkingTranscript, "|", "\n", -1)

		// TODO: communicate status to the client.
		// For detailed discussions of how to return status to the client upon completion of a long-running request, see:
		// - "REST and long-running jobs", https://farazdagi.com/2014/rest-and-long-running-jobs/
		// - "Long running REST API with queues", https://stackoverflow.com/a/33011965/10649045 .

		// Set a non-2xx status code to indicate a failure in task processing that should be retried.
		// For example, http.Error(w, "Internal Server Error: Task Processing", http.StatusInternalServerError)
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")

		// !!! HACK !!! write the response to the client as if responding to the original POST request

		// populate a CompletionResponse struct for the HTTP response, with
		// selected fields of Request
		incomingRequest.CompletedAt = time.Now().UTC().Format(time.RFC3339Nano)
		response := request.GetTranscriptResponse{
			RequestID:    incomingRequest.RequestID,
			CustomerID:   incomingRequest.CustomerID,
			MediaFileURI: incomingRequest.MediaFileURI,
			AcceptedAt:   incomingRequest.AcceptedAt,
			CompletedAt:  incomingRequest.CompletedAt,
			Transcript:   incomingRequest.WorkingTranscript,
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("%s.postHandler, json.NewEncoder.Encode error: +%v\n", sn, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// service duration
		serviceDuration := time.Now().UTC().Sub(startTime)

		// total request duration
		requestDuration, err := incomingRequest.RequestDuration()
		if err != nil {
			log.Printf("%s.postHandler, error: %v\n", sn, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Printf("%s.taskHandler completed in %v =====> Request Processed in %v <==== : queue %q, task %q, response: %+v",
			sn, serviceDuration, requestDuration, queueName, taskName, response)
	}
}

// indexHandler serves as a health check, responding "service running"
func indexHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	sn := cfg.ServiceName
	// log.Printf("Enter %s.indexHandler\n", sn)
	if r.URL.Path != "/" {
		// log.Printf("%s.indexHandler, r.URL.Path: %s, will respond NotFound\n", sn, r.URL.Path)
		http.NotFound(w, r)
		return
	}
	// indicate service is running
	fmt.Fprintf(w, "%q service running\n", sn)
}

func myNotFound(w http.ResponseWriter, r *http.Request) {
	// log.Printf("%s.myNotFound, request for %s not routed\n", sn, r.URL.Path)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	_, _ = w.Write([]byte("<h2>404 Not Foundw</h2>"))
}
