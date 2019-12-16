package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/spf13/viper"

	"github.com/peterpla/lead-expert/pkg/adding"
	"github.com/peterpla/lead-expert/pkg/appengine"
	"github.com/peterpla/lead-expert/pkg/config"
	"github.com/peterpla/lead-expert/pkg/middleware"
	"github.com/peterpla/lead-expert/pkg/serviceInfo"
)

var prefix = "TaskInitialRequest"
var logPrefix = "initial-request.main.init(),"
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

		newRequest := incomingRequest

		// add timestamps and get duration
		var duration time.Duration
		if duration, err = newRequest.AddTimestamps("BeginInitialRequest", startTime, "EndInitialRequest"); err != nil {
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
		log.Printf("%s.indexHandler, r.URL.Path: %s, will respond NotFound\n", sn, r.URL.Path)
		http.NotFound(w, r)
		return
	}

	// I'm not dead yet
	fmt.Fprintf(w, "%q service running\n", sn)
}

func myNotFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	_, _ = w.Write([]byte("<h2>404 Not Foundw</h2>"))
}
