package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	dlp "cloud.google.com/go/dlp/apiv2"
	"github.com/go-playground/validator"
	"github.com/julienschmidt/httprouter"
	"github.com/spf13/viper"
	dlppb "google.golang.org/genproto/googleapis/privacy/dlp/v2"

	"github.com/peterpla/lead-expert/pkg/appengine"
	"github.com/peterpla/lead-expert/pkg/config"
	"github.com/peterpla/lead-expert/pkg/database"
	"github.com/peterpla/lead-expert/pkg/middleware"
	"github.com/peterpla/lead-expert/pkg/queue"
	"github.com/peterpla/lead-expert/pkg/request"
	"github.com/peterpla/lead-expert/pkg/serviceInfo"
)

var prefix = "TaskTagging"
var logPrefix = "tagging.main.init(),"
var cfg config.Config
var repo request.RequestRepository
var q queue.Queue
var qi = queue.QueueInfo{}
var qs queue.QueueService

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

	log.Printf("Starting service %s listening on port %s, requests will be added to queue %s",
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

// taskHandler processes task requests.
func taskHandler(q queue.Queue) httprouter.Handle {
	sn := serviceInfo.GetServiceName()

	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		// log.Printf("%s.taskHandler, request: %+v, params: %+v\n", sn, r, p)
		startTime := time.Now().UTC().Format(time.RFC3339Nano)

		// pull task and queue names from App Engine headers
		taskName, queueName := appengine.GetAppEngineInfo(w, r)

		var err error
		incomingRequest := request.Request{}
		if err = incomingRequest.ReadRequest(w, r, p, validate); err != nil {
			// ReadRequest called http.Error so we just return
			return
		}

		newRequest := incomingRequest

		// TODO: implement tagging processing: select the ML tagging service
		// to use and submit that request.
		//
		// The current default selection is Google Data Loss Prevention (DLP)
		// Classification using pre-defined (and eventually, custom) InfoType detectors.
		// See https://cloud.google.com/dlp/docs/concepts-infotypes
		//
		// TODO: to select from additional services, add a tagging-dispatch servive

		if err = gDLPTagging(&newRequest); err != nil {
			log.Printf("%s.taskHandler, gDLPTagging error: %+v\n", sn, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Printf("%s.taskHandler, tags: %+v\n", sn, newRequest.MatchedTags)

		// add timestamps and get duration
		var duration time.Duration
		if duration, err = newRequest.AddTimestamps("BeginTagging", startTime, "EndTagging"); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// TODO: write updated Request to the Requests database
		_ = repo

		// create task on the next pipeline stage's queue with updated request
		if err = q.Add(&qi, &newRequest); err != nil {
			log.Printf("%s.taskHandler, q.Add error: %+v\n", sn, err)
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

func gDLPTagging(req *request.Request) error {
	sn := serviceInfo.GetServiceName()
	// log.Printf("%s.gDLPTagging, enter,  req: %+v\n", sn, req)

	if req.WorkingTranscript == "" {
		// no transcript to tag; all downstream pipeline stages will fail so report error and exit
		log.Printf("%s.gDLPTagging, empty WorkingTranscript: %q\n", sn, req.WorkingTranscript)
		return ErrEmptyTranscript
	}

	// first use of req.MatchedTags, initialize the map
	req.MatchedTags = make(map[string]request.Tags)

	var client *dlp.Client
	var ctx context.Context
	var err error

	if client, ctx, err = gDLPClient(); err != nil {
		log.Printf("%s.gDLPTagging, gDLPClient err: %v\n", sn, err)
		msg := fmt.Sprintf("gDLP error: %v", err)
		return &taggingError{status: http.StatusInternalServerError, msg: msg}
	}
	defer client.Close()

	gDLPReq := gDLPPrepareRequest(req)

	var resp *dlppb.InspectContentResponse
	if resp, err = gDLPInspect(ctx, client, gDLPReq); err != nil {
		log.Printf("%s.gDLPTagging, gDLPInspect err: %v\n", sn, err)
		msg := fmt.Sprintf("gDLP error: %v", err)
		return &taggingError{status: http.StatusInternalServerError, msg: msg}
	}

	// Copy tags matching in transcript into Request's tags map
	gDLPTagsToTagMap(resp.Result, req)

	return nil
}

func gDLPClient() (*dlp.Client, context.Context, error) {
	sn := serviceInfo.GetServiceName()
	// log.Printf("%s.gDLPClient enter\n", sn)

	ctx := context.Background()

	// Initialize client.
	client, err := dlp.NewClient(ctx)
	if err != nil {
		log.Printf("%s.gDLPClient, NewClient err: %v\n", sn, err)
		return nil, nil, ErrDLPError
	}
	return client, ctx, nil
}

func gDLPPrepareRequest(req *request.Request) *dlppb.InspectContentRequest {
	sn := serviceInfo.GetServiceName()
	// log.Printf("%s.gDLPPrepareRequest enter\n", sn)

	// set parameters for request
	input := req.WorkingTranscript
	projectID := config.GetConfigPointer().ProjectID

	minLikelihood := dlppb.Likelihood_POSSIBLE
	includeQuote := true
	// TODO: add/tune list of InfoTypes we want to match
	infoTypes := []*dlppb.InfoType{
		{Name: "PHONE_NUMBER"},
		{Name: "PERSON_NAME"},
		{Name: "STREET_ADDRESS"},
		{Name: "US_STATE"},
	}
	item := &dlppb.ContentItem{
		DataItem: &dlppb.ContentItem_Value{
			Value: input,
		},
	}

	// Create the request
	gDLPReq := &dlppb.InspectContentRequest{
		Parent: "projects/" + projectID,
		Item:   item,
		InspectConfig: &dlppb.InspectConfig{
			InfoTypes:     infoTypes,
			MinLikelihood: minLikelihood,
			IncludeQuote:  includeQuote,
		},
	}
	log.Printf("%s.gDLPPrepareRequest exit, gDLPReq: %+v\n", sn, gDLPReq)

	return gDLPReq
}

func gDLPInspect(ctx context.Context, client *dlp.Client, gDLPReq *dlppb.InspectContentRequest) (*dlppb.InspectContentResponse, error) {
	sn := serviceInfo.GetServiceName()
	// log.Printf("%s.gDLPInspect enter\n", sn)

	resp, err := client.InspectContent(ctx, gDLPReq)
	if err != nil {
		log.Printf("%s.gDLPInspect, InspectContent err: %v\n", sn, err)
		msg := fmt.Sprintf("gDLP error: %v", err)
		return nil, &taggingError{status: http.StatusInternalServerError, msg: msg}
	}
	return resp, nil
}

func gDLPTagsToTagMap(result *dlppb.InspectResult, req *request.Request) {
	sn := serviceInfo.GetServiceName()
	log.Printf("%s.gDLPTagsToTagMap enter, result: %+v\n", sn, result)

	log.Printf("Findings: %d\n", len(result.Findings))
	for _, f := range result.Findings {
		var tag = request.Tags{}

		name := f.GetInfoType().GetName()
		tag.Quote = f.GetQuote()
		tag.Likelihood = int(f.GetLikelihood())
		tag.BeginByteOffset = int(f.Location.GetByteRange().GetStart())
		tag.EndByteOffset = int(f.Location.GetByteRange().GetEnd())

		if _, ok := req.MatchedTags[name]; ok {
			log.Printf("%s.gDLPTagsToTagMap, f[%q]: %+v, compare to existing req.MatchedTags[%q]: %+v\n",
				sn, name, f, name, req.MatchedTags[name])
			// already have this tag, ignore unless it has a higher likelihood
			if tag.Likelihood <= req.MatchedTags[name].Likelihood {
				continue
			}
		}
		// otherwise add this tag to the map
		log.Printf("%s.gDLPTagsToTagMap, added to req.MatchedTags[%q]: %+v\n", sn, name, tag)
		req.MatchedTags[name] = tag
	}
	log.Printf("%s.gDLPTagsToTagMap, exiting, tags: %+v\n", sn, req.MatchedTags)
}

// ********** ********** ********** ********** ********** **********

// indexHandler responds to requests with "service running"
func indexHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	sn := serviceInfo.GetServiceName()

	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// I'm not dead yet
	fmt.Fprintf(w, "%q service running\n", sn)
}

// ********** ********** ********** ********** ********** **********

func myNotFound(w http.ResponseWriter, r *http.Request) {
	var msg404 = []byte("<h2>404 Not Foundw</h2>")

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	_, _ = w.Write(msg404)
}

// ********** ********** ********** ********** ********** **********

type taggingError struct {
	status int
	msg    string
}

func (mr *taggingError) Error() string {
	return mr.msg
}

var ErrEmptyTranscript = &taggingError{status: http.StatusInternalServerError, msg: "empty transcript"}
var ErrDLPError = &taggingError{status: http.StatusInternalServerError, msg: "gDLP error"}
