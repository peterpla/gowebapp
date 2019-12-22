package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-playground/validator"
	"github.com/julienschmidt/httprouter"
	"github.com/spf13/viper"

	"github.com/peterpla/lead-expert/pkg/adding"
	"github.com/peterpla/lead-expert/pkg/appengine"
	"github.com/peterpla/lead-expert/pkg/config"
	"github.com/peterpla/lead-expert/pkg/middleware"
	"github.com/peterpla/lead-expert/pkg/serviceInfo"
)

var prefix = "TaskTagging"
var logPrefix = "tagging.main.init(),"
var cfg config.Config

// use a single instance of Validate, it caches struct info
var validate *validator.Validate

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

	validate = validator.New()

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

		incomingRequest := adding.Request{}
		if err := incomingRequest.ReadRequest(w, r, p, validate); err != nil {
			// ReadRequest called http.Error so we just return
			return
		}

		newRequest := incomingRequest

		// TODO: implement tagging processing
		// E.g., select which ML tagging service to use and submit that request.
		//
		// The current default selection is TBD
		// so TaskTaggingWriteToQ and TaskTaggingNextSvcToHandleReq
		// reflect "tagging" as the next stage in the pipeline.

		// add timestamps and get duration
		var duration time.Duration
		var err error
		if duration, err = newRequest.AddTimestamps("BeginTagging", startTime, "EndTagging"); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// create task on the next pipeline stage's queue with updated request
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
