package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-playground/validator"
	"github.com/julienschmidt/httprouter"
	"github.com/spf13/viper"

	"github.com/peterpla/lead-expert/pkg/appengine"
	"github.com/peterpla/lead-expert/pkg/config"
	"github.com/peterpla/lead-expert/pkg/database"
	"github.com/peterpla/lead-expert/pkg/middleware"
	"github.com/peterpla/lead-expert/pkg/queue"
	"github.com/peterpla/lead-expert/pkg/request"
	"github.com/peterpla/lead-expert/pkg/serviceInfo"
)

var prefix = "TaskCompletionProcessing"
var initLogPrefix = "completion-processing.main.init(),"
var cfg config.Config
var repo request.RequestRepository
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
	sn := serviceInfo.GetServiceName()
	// Creating App Engine task handlers: https://cloud.google.com/tasks/docs/creating-appengine-handlers

	defer catch() // implements recover so panics reported

	// connect to the Request database
	repo = database.NewFirestoreRequestRepository(cfg.ProjectID, cfg.DatabaseRequests)

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

	log.Printf("Starting service %s listening on port %s, requests will be added to queue %s\n",
		sn, port, cfg.QueueName)
	// run ListenAndServe in a separate go routine so main can listen for signals
	go startListening(":"+port, middleware.LogReqResp(router))

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	s := <-signals
	log.Printf("\n%s.main, received signal %s, terminating", sn, s.String())
}

func startListening(addr string, handler http.Handler) {
	if err := http.ListenAndServe(addr, handler); err != http.ErrServerClosed {
		log.Fatalf("%s.startListening, ListenAndServe returned err: %+v\n", serviceInfo.GetServiceName(), err)
	}
}

// catch recover() and log it
func catch() {
	defer func() {
		if r := recover(); r != nil {
			log.Fatalf("=====> RECOVER in %s.main.catch, recover() returned: %v\n", serviceInfo.GetServiceName(), r)
		}
	}()
}

// ********** ********** ********** ********** ********** **********

// handler for Cloud Tasks POSTs
func taskHandler() httprouter.Handle {
	sn := serviceInfo.GetServiceName()

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

		// TODO: implement whatever constitutes "completion processing"

		// replace | with \n in WorkingTranscript
		incomingRequest.FinalTranscript = strings.Replace(incomingRequest.WorkingTranscript, "|", "\n", -1)
		incomingRequest.Status = request.Completed
		incomingRequest.CompletedAt = time.Now().UTC().Format(time.RFC3339Nano)

		// add timestamps and get duration
		_, err := incomingRequest.AddTimestamps("BeginCompletionProcessing", startTime.Format(time.RFC3339Nano), "EndCompletionProcessing")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// write completed Request to the Requests database
		if err := repo.Update(&incomingRequest); err != nil {
			log.Printf("%s.postHandler, repo.Update error: %+v\n", sn, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Set a non-2xx status code to indicate a failure in task processing that should be retried.
		// For example, http.Error(w, "Internal Server Error: Task Processing", http.StatusInternalServerError)
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")

		// populate a CompletionResponse struct for the HTTP response, with
		// selected fields of Request
		response := request.GetTranscriptResponse{
			RequestID:    incomingRequest.RequestID,
			CustomerID:   incomingRequest.CustomerID,
			MediaFileURI: incomingRequest.MediaFileURI,
			AcceptedAt:   incomingRequest.AcceptedAt,
			CompletedAt:  incomingRequest.CompletedAt,
			Transcript:   incomingRequest.WorkingTranscript,
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("%s.postHandler, json.NewEncoder.Encode error: %+v\n", sn, err)
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

// ********** ********** ********** ********** ********** **********

// indexHandler serves as a health check, responding "service running"
func indexHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	sn := serviceInfo.GetServiceName()
	// log.Printf("Enter %s.indexHandler\n", sn)
	if r.URL.Path != "/" {
		// log.Printf("%s.indexHandler, r.URL.Path: %s, will respond NotFound\n", sn, r.URL.Path)
		http.NotFound(w, r)
		return
	}
	// indicate service is running
	fmt.Fprintf(w, "%q service running\n", sn)
}

// ********** ********** ********** ********** ********** **********

func myNotFound(w http.ResponseWriter, r *http.Request) {
	// log.Printf("%s.myNotFound, request for %s not routed\n", sn, r.URL.Path)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	_, _ = w.Write([]byte("<h2>404 Not Foundw</h2>"))
}
