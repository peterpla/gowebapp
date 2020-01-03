package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-playground/validator"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/spf13/viper"

	"github.com/peterpla/lead-expert/pkg/adding"
	"github.com/peterpla/lead-expert/pkg/config"
	"github.com/peterpla/lead-expert/pkg/middleware"
	"github.com/peterpla/lead-expert/pkg/queue"
	"github.com/peterpla/lead-expert/pkg/serviceInfo"
)

var prefix = "TaskDefault"
var initLogPrefix = "default.main.init(),"
var cfg config.Config
var apiPrefix = "/api/v1"
var useQueue queue.QueueInfo

// use a single instance of Validate, it caches struct info
var validate *validator.Validate

func init() {
	if err := config.GetConfig(&cfg, prefix); err != nil {
		msg := fmt.Sprintf(initLogPrefix+" GetConfig error: %v", err)
		panic(msg)
	}

	// register them for access by other packages in this service
	serviceInfo.RegisterServiceName(cfg.ServiceName)
	serviceInfo.RegisterQueueName(cfg.QueueName)
	serviceInfo.RegisterNextServiceName(cfg.NextServiceName)
}

func main() {
	// log.Printf("Enter default.main\n")

	if !cfg.IsGAE {
		useQueue.Name = "local"
		useQueue.ServiceToHandle = ""
	} else {
		// Google Cloud Tasks
		useQueue.Name = fmt.Sprintf("projects/%s/locations/%s/queues/%s",
			cfg.ProjectID, cfg.TasksLocation, cfg.QueueName)
		useQueue.ServiceToHandle = cfg.NextServiceName
	}

	router := httprouter.New()
	router.POST(apiPrefix+"/requests", postHandler(cfg.Adder))
	router.GET(apiPrefix+"/queues/:uuid", getQueueHandler(cfg.Adder))
	router.GET(apiPrefix+"/transcripts/:uuid", getTranscriptHandler(cfg.Adder))
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

// postHandler returns the handler func for POST /requests
func postHandler(a adding.Service) httprouter.Handle {
	var err error
	sn := cfg.ServiceName

	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		startTime := time.Now().UTC()
		// log.Printf("%s.main.postHandler, enter\n", sn)

		newRequest := adding.Request{}
		if err = newRequest.ReadRequest(w, r, p, validate); err != nil {
			// log.Printf("%s.postHandler, err: %v\n", sn, err)
			// readRequest calls http.Error() on error
			return
		}
		newRequest.RequestID = uuid.New()
		newRequest.AcceptedAt = time.Now().UTC().Format(time.RFC3339Nano)
		newRequest.Status = adding.Pending
		location := getStatusURI(newRequest.RequestID)

		// add timestamps and get duration
		var duration time.Duration
		if duration, err = newRequest.AddTimestamps("BeginDefault", startTime.Format(time.RFC3339Nano), "EndDefault"); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// add the request (e.g., to a queue) for subsequent processing
		returnedReq := a.AddRequest(newRequest)

		// TODO: save returnedRequest to database

		// provide selected fields of Request as the HTTP response
		response := adding.PostResponse{
			RequestID:    returnedReq.RequestID,
			CustomerID:   returnedReq.CustomerID,
			MediaFileURI: returnedReq.MediaFileURI,
			AcceptedAt:   returnedReq.AcceptedAt,
			Location:     location,
		}

		// send response to client
		w.WriteHeader(http.StatusAccepted)
		w.Header().Set("Content-Type", "application/json")
		if err = json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("%s.postHandler, json.NewEncoder.Encode error: +%v\n", sn, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Printf("%s.postHandler, completed in %v, newRequest: %+v\n", sn, duration, newRequest)
	}
}

func getStatusURI(reqID uuid.UUID) string {
	return apiPrefix + "/queues/" + reqID.String()
}

// ********** ********** ********** ********** ********** **********

// getQueueHandler returns the handler func for GET /queue
func getQueueHandler(a adding.Service) httprouter.Handle {
	sn := cfg.ServiceName
	// log.Printf("%s.getQueueHandler, enter/exit\n", sn)

	// var err error

	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		startTime := time.Now().UTC()
		// log.Printf("%s.getQueueHandler, enter\n", sn)

		var err error
		reqForStatus := adding.Request{}
		if err = reqForStatus.ReadRequest(w, r, p, validate); err != nil {
			log.Printf("%s.getQueueHandler, err: %v\n", sn, err)
			// readRequest calls http.Error() on error
			return
		}
		reqForStatus.RequestID = uuid.New()
		reqForStatus.AcceptedAt = time.Now().UTC().Format(time.RFC3339Nano)

		var requestedUUID uuid.UUID
		paramUUID := p.ByName("uuid")
		if requestedUUID, err = uuid.Parse(paramUUID); err != nil {
			log.Printf("%s.getQueueHandler, bad UUID err: %v\n", sn, err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// !!! HACK !!! - fake a response, since we don't have the database running - !!! HACK !!!
		acceptedAt := startTime.Add(time.Second * -1)
		log.Printf("%s.getQueueHandler, =====> PLACEHOLDER <===== query database for status of %s\n",
			sn, requestedUUID)

		// !!! HACK !!! - should get this from database - !!! HACK !!!
		originalRequest := adding.Request{
			RequestID:    requestedUUID,
			CustomerID:   reqForStatus.CustomerID,
			MediaFileURI: reqForStatus.MediaFileURI,
			Status:       adding.Pending,
			AcceptedAt:   acceptedAt.Format(time.RFC3339Nano),
		}

		// handle special UUIDs used for testing
		if requestedUUID.String() == adding.PendingUUIDStr {
			originalRequest.RequestID = adding.PendingUUID
			originalRequest.Status = adding.Pending
			originalRequest.OriginalStatus = 0
		}
		if requestedUUID.String() == adding.CompletedUUIDStr {
			originalRequest.RequestID = adding.CompletedUUID
			originalRequest.Status = adding.Completed
			originalRequest.OriginalStatus = http.StatusOK
		}
		if requestedUUID.String() == adding.ErrorUUIDStr {
			originalRequest.RequestID = adding.ErrorUUID
			originalRequest.Status = adding.Error
			originalRequest.OriginalStatus = http.StatusBadRequest
		}

		// provide selected fields of Request as the HTTP response
		response := adding.GetQueueResponse{
			RequestID:         reqForStatus.RequestID,
			CustomerID:        originalRequest.CustomerID,
			MediaFileURI:      originalRequest.MediaFileURI,
			AcceptedAt:        originalRequest.AcceptedAt,
			OriginalRequestID: originalRequest.RequestID,
		}

		switch originalRequest.Status {
		case adding.Error:
			response.OriginalStatus = originalRequest.OriginalStatus
		case adding.Pending:
			etaTime := time.Now().UTC()
			etaTime = etaTime.Add(time.Second * 45) // TODO: calculate multiplier based on recent processing time
			response.ETA = etaTime.Format(time.RFC3339Nano)
			response.OriginalStatus = originalRequest.OriginalStatus
		case adding.Completed:
			response.Location = getLocationURI(requestedUUID)
			response.OriginalStatus = originalRequest.OriginalStatus
		default:
			log.Printf("%s.getQueueHandler, invalid originalRequest.Status: %v\n", sn, originalRequest.Status)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// add timestamps and get duration
		var duration time.Duration
		if duration, err = reqForStatus.AddTimestamps("BeginDefault", startTime.Format(time.RFC3339Nano), "EndDefault"); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// send response to client
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		if err = json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("%s.postHandler, json.NewEncoder.Encode error: +%v\n", sn, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Printf("%s.getQueueHandler, completed in %v, response: %+v\n", sn, duration, response)
	}
}

func getLocationURI(reqID uuid.UUID) string {
	return apiPrefix + "/transcripts/" + reqID.String()

}

// ********** ********** ********** ********** ********** **********

// getTranscriptHandler returns the handler func for GET /queue
func getTranscriptHandler(a adding.Service) httprouter.Handle {
	sn := cfg.ServiceName
	// log.Printf("%s.getTranscriptHandler, enter/exit\n", sn)

	// var err error

	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		startTime := time.Now().UTC()
		// log.Printf("%s.getTranscriptHandler, enter\n", sn)

		var err error
		reqForTranscript := adding.Request{}
		if err = reqForTranscript.ReadRequest(w, r, p, validate); err != nil {
			log.Printf("%s.getTranscriptHandler, err: %v\n", sn, err)
			// readRequest calls http.Error() on error
			return
		}
		reqForTranscript.RequestID = uuid.New()
		reqForTranscript.AcceptedAt = time.Now().UTC().Format(time.RFC3339Nano)

		var requestedUUID uuid.UUID
		paramUUID := p.ByName("uuid")
		if requestedUUID, err = uuid.Parse(paramUUID); err != nil {
			log.Printf("%s.getTranscriptHandler, bad UUID err: %v\n", sn, err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// !!! HACK !!! - fake a response, since we don't have the database running - !!! HACK !!!
		acceptedAt := startTime.Add(time.Second * -47)
		completedAt := startTime.Add(time.Second * -2)
		completedAt = completedAt.Add(time.Millisecond * -37521)
		log.Printf("%s.getTranscriptHandler, =====> PLACEHOLDER <===== query database for status of %s\n",
			sn, requestedUUID)

		// !!! HACK !!! - should get this from database - !!! HACK !!!
		completedRequest := adding.Request{
			RequestID:       requestedUUID,
			CustomerID:      reqForTranscript.CustomerID,
			MediaFileURI:    reqForTranscript.MediaFileURI,
			Status:          "COMPLETED",
			AcceptedAt:      acceptedAt.Format(time.RFC3339Nano),
			CompletedAt:     completedAt.Format(time.RFC3339Nano),
			FinalTranscript: "[Speaker 1] Thank you for calling Park flooring.\n[Speaker 2] Hi, my name is Yuri.\n",
		}

		// add timestamps and get duration
		var duration time.Duration
		if duration, err = reqForTranscript.AddTimestamps("BeginDefault", startTime.Format(time.RFC3339Nano), "EndDefault"); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// provide selected fields of Request as the HTTP response
		response := adding.GetTranscriptResponse{
			RequestID:    reqForTranscript.RequestID, // this request for transcript
			CustomerID:   completedRequest.CustomerID,
			MediaFileURI: completedRequest.MediaFileURI,
			AcceptedAt:   completedRequest.AcceptedAt,
			CompletedAt:  completedRequest.CompletedAt,
			CompletedID:  completedRequest.RequestID, // the request that produced the transcript
			Transcript:   completedRequest.FinalTranscript,
		}

		// send response to client
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		if err = json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("%s.getTranscriptHandler, json.NewEncoder.Encode error: +%v\n", sn, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Printf("%s.getTranscriptHandler, completed in %v, response: %+v\n", sn, duration, response)
	}
}

// ********** ********** ********** ********** ********** **********

// indexHandler serves as a health check, responding "service running"
func indexHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	sn := cfg.ServiceName
	// log.Printf("Enter %s.indexHandler\n", sn)
	if r.URL.Path != "/" {
		log.Printf("%s.indexHandler, r.URL.Path: %s, will respond NotFound\n", sn, r.URL.Path)
		http.NotFound(w, r)
		return
	}
	// indicate service is running
	fmt.Fprintf(w, "%q service running\n", sn)
}

// ********** ********** ********** ********** ********** **********

func myNotFound(w http.ResponseWriter, r *http.Request) {
	// sn := cfg.ServiceName
	// log.Printf("%s.myNotFound, request for %s not routed\n", sn, r.URL.Path)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	_, _ = w.Write([]byte("<h2>404 Not Foundw</h2>"))
}
