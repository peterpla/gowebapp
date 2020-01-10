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

	"github.com/peterpla/lead-expert/pkg/config"
	"github.com/peterpla/lead-expert/pkg/database"
	"github.com/peterpla/lead-expert/pkg/middleware"
	"github.com/peterpla/lead-expert/pkg/queue"
	"github.com/peterpla/lead-expert/pkg/request"
	"github.com/peterpla/lead-expert/pkg/serviceInfo"
)

var prefix = "TaskDefault"
var initLogPrefix = "default.main.init(),"
var cfg config.Config
var apiPrefix = "/api/v1"
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

	// register them for access by other packages in this service
	serviceInfo.RegisterServiceName(cfg.ServiceName)
	serviceInfo.RegisterQueueName(cfg.QueueName)
	serviceInfo.RegisterNextServiceName(cfg.NextServiceName)
}

func main() {
	// log.Printf("Enter default.main\n")

	defer catch()

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
	router.POST(apiPrefix+"/requests", postHandler(q))
	router.GET(apiPrefix+"/status/:uuid", getStatusHandler())
	router.GET(apiPrefix+"/transcripts/:uuid", getTranscriptsHandler())
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
		serviceInfo.GetServiceName(), port, cfg.QueueName)
	log.Fatal(http.ListenAndServe(":"+port, middleware.LogReqResp(router)))
}

// catch recover() and log it
func catch() {
	defer func() {
		if r := recover(); r != nil {
			log.Fatalf("=====> RECOVER in %s.main.catch, recover() returned: %v\n", serviceInfo.GetServiceName(), r)
		}
	}()
}

// postHandler returns the handler func for POST /requests
func postHandler(q queue.Queue) httprouter.Handle {
	var err error
	sn := serviceInfo.GetServiceName()

	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		startTime := time.Now().UTC()
		// log.Printf("%s.main.postHandler, enter, repo: %+v\n", sn, repo)

		newRequest := request.Request{}
		if err = newRequest.ReadRequest(w, r, p, validate); err != nil {
			// log.Printf("%s.postHandler, err: %v\n", sn, err)
			// readRequest calls http.Error() on error
			return
		}
		newRequest.RequestID = uuid.New()
		newRequest.AcceptedAt = time.Now().UTC().Format(time.RFC3339Nano)
		newRequest.Status = request.Pending

		// add timestamps and get duration
		var duration time.Duration
		if duration, err = newRequest.AddTimestamps("BeginDefault", startTime.Format(time.RFC3339Nano), "EndDefault"); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// write the Request to the Requests database
		if err := repo.Create(&newRequest); err != nil {
			log.Printf("%s.postHandler, repo.Create error: %+v\n", sn, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// newRequest.CreatedAt = time.Now().UTC().Format(time.RFC3339Nano)

		// create task on the next pipeline stage's queue with request
		if err := q.Add(&qi, &newRequest); err != nil {
			log.Printf("%s.postHandler, q.Add error: %+v\n", sn, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// provide selected fields of Request as the HTTP response
		response := request.PostResponse{
			RequestID:    newRequest.RequestID,
			CustomerID:   newRequest.CustomerID,
			MediaFileURI: newRequest.MediaFileURI,
			AcceptedAt:   newRequest.AcceptedAt,
			PollEndpoint: getStatusURI(newRequest.RequestID),
		}

		// send response to client
		w.WriteHeader(http.StatusAccepted)
		w.Header().Set("Content-Type", "application/json")
		if err = json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("%s.postHandler, json.NewEncoder.Encode error: %+v\n", sn, err)
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

// getStatusHandler returns the handler func for GET /queue
func getStatusHandler() httprouter.Handle {
	sn := serviceInfo.GetServiceName()
	log.Printf("%s.getStatusHandler, enter/exit\n", sn)

	// var err error

	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		startTime := time.Now().UTC()
		log.Printf("%s.main.getStatusHandler, enter, repo: %+v\n", sn, repo)

		var err error
		reqForStatus := request.Request{}
		if err = reqForStatus.ReadRequest(w, r, p, validate); err != nil {
			log.Printf("%s.getStatusHandler, err: %v\n", sn, err)
			// readRequest calls http.Error() on error
			return
		}

		// validate the requested UUID
		var requestedUUID uuid.UUID
		paramUUID := p.ByName("uuid")
		if requestedUUID, err = uuid.Parse(paramUUID); err != nil {
			log.Printf("%s.getStatusHandler, bad UUID err: %v\n", sn, err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var zeroUUID uuid.UUID
		if requestedUUID == zeroUUID {
			log.Printf("%s.getStatusHandler, zero UUID\n", sn)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		reqForStatus.RequestID = uuid.New()
		reqForStatus.AcceptedAt = time.Now().UTC().Format(time.RFC3339Nano)

		var originalRequest request.Request
		var returnedReq *request.Request

		// for special testing UUIDs, hardwire responses
		switch requestedUUID.String() {
		//
		case request.PendingUUIDStr:
			originalRequest.RequestID = request.PendingUUID
			originalRequest.Status = request.Pending

		case request.CompletedUUIDStr:
			originalRequest.RequestID = request.CompletedUUID
			originalRequest.Status = request.Completed
			originalRequest.OriginalStatus = http.StatusOK

		case request.ErrorUUIDStr:
			originalRequest.RequestID = request.ErrorUUID
			originalRequest.Status = request.Error
			originalRequest.OriginalStatus = http.StatusBadRequest

		default:
			// not a special case, find the requested UUID in the database
			returnedReq, err = repo.FindByID(requestedUUID)
			if err == database.ErrNotFoundError {
				log.Printf("%s.getStatusHandler, UUID not found: %q\n", sn, requestedUUID.String())
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			if err != nil {
				log.Printf("%s.getStatusHandler, repo.FindByID error: %+v\n", sn, err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			originalRequest = *returnedReq
		}

		// TODO: validate the same CustomerID is asking for status of the same MediaFileURI, reqForStatus vs. originalRequest

		reqForStatus.CompletedAt = time.Now().UTC().Format(time.RFC3339Nano)

		// provide selected fields of Request as the HTTP response
		response := request.GetStatusResponse{
			RequestID:         reqForStatus.RequestID,
			CustomerID:        originalRequest.CustomerID,
			MediaFileURI:      originalRequest.MediaFileURI,
			AcceptedAt:        originalRequest.AcceptedAt,
			OriginalRequestID: originalRequest.RequestID,
			CompletedAt:       reqForStatus.CompletedAt,
		}

		switch originalRequest.Status {
		case request.Error:
			response.OriginalStatus = originalRequest.OriginalStatus
			response.OriginalCompletedAt = originalRequest.CompletedAt
		case request.Pending:
			response.ETA = getETA().Format(time.RFC3339Nano)
			response.Endpoint = getStatusURI(originalRequest.RequestID)
		case request.Completed:
			response.Endpoint = getLocationURI(originalRequest.RequestID)
			response.OriginalStatus = originalRequest.OriginalStatus
			response.OriginalCompletedAt = originalRequest.CompletedAt
		default:
			log.Printf("%s.getStatusHandler, invalid originalRequest.Status: %v\n", sn, originalRequest.Status)
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
			log.Printf("%s.postHandler, json.NewEncoder.Encode error: %+v\n", sn, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Printf("%s.getStatusHandler, completed in %v, response: %+v\n", sn, duration, response)
	}
}

func getETA() time.Time {
	// TODO: calculate multiplier based on recent processing time
	etaTime := time.Now().UTC()
	etaTime = etaTime.Add(time.Second * 45) // blindly guess 45 seconds from now
	return etaTime
}

func getLocationURI(reqID uuid.UUID) string {
	return apiPrefix + "/transcripts/" + reqID.String()

}

// ********** ********** ********** ********** ********** **********

// getTranscriptsHandler returns the handler func for GET /queue
func getTranscriptsHandler() httprouter.Handle {
	sn := serviceInfo.GetServiceName()
	// log.Printf("%s.getTranscriptsHandler, enter/exit\n", sn)

	// var err error

	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		startTime := time.Now().UTC()
		// log.Printf("%s.getTranscriptsHandler, enter\n", sn)

		var err error
		reqForTranscript := request.Request{}
		if err = reqForTranscript.ReadRequest(w, r, p, validate); err != nil {
			log.Printf("%s.getTranscriptsHandler, err: %v\n", sn, err)
			// readRequest calls http.Error() on error
			return
		}
		reqForTranscript.RequestID = uuid.New()
		reqForTranscript.AcceptedAt = time.Now().UTC().Format(time.RFC3339Nano)

		var requestedUUID uuid.UUID
		paramUUID := p.ByName("uuid")
		if requestedUUID, err = uuid.Parse(paramUUID); err != nil {
			log.Printf("%s.getTranscriptsHandler, bad UUID err: %v\n", sn, err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var requestPointer *request.Request
		requestPointer, err = repo.FindByID(requestedUUID)
		if err == database.ErrNotFoundError {
			log.Printf("%s.getTranscriptsHandler, UUID not found: %q\n", sn, requestedUUID.String())
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if err != nil {
			log.Printf("%s.getTranscriptsHandler, repo.FindByID error: %+v\n", sn, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if requestPointer.Status != request.Completed {
			log.Printf("%s.getTranscriptsHandler, Status not COMPLETED: %q\n", sn, requestPointer.Status)
			w.WriteHeader(http.StatusSeeOther)
			// TODO: implement See Other response with Location: /status/:uuid - client needs to poll until Completed
			return
		}

		returnedRequest := *requestPointer

		// TODO: validate the same CustomerID is asking for status of the same MediaFileURI, reqForStatus vs. originalRequest

		// add timestamps and get duration
		var duration time.Duration
		if duration, err = reqForTranscript.AddTimestamps("BeginDefault", startTime.Format(time.RFC3339Nano), "EndDefault"); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// provide selected fields of Request as the HTTP response
		response := request.GetTranscriptResponse{
			RequestID:           reqForTranscript.RequestID, // this request for transcript
			CustomerID:          returnedRequest.CustomerID,
			MediaFileURI:        returnedRequest.MediaFileURI,
			AcceptedAt:          returnedRequest.AcceptedAt,
			CompletedAt:         returnedRequest.CompletedAt,
			OriginalRequestID:   returnedRequest.RequestID, // the request that produced the transcript
			OriginalAcceptedAt:  returnedRequest.AcceptedAt,
			OriginalCompletedAt: returnedRequest.CompletedAt,
			Transcript:          returnedRequest.FinalTranscript,
		}

		// send response to client
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		if err = json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("%s.getTranscriptsHandler, json.NewEncoder.Encode error: %+v\n", sn, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Printf("%s.getTranscriptsHandler, completed in %v, response: %+v\n", sn, duration, response)
	}
}

// ********** ********** ********** ********** ********** **********

// indexHandler serves as a health check, responding "service running"
func indexHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	sn := serviceInfo.GetServiceName()
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
	sn := serviceInfo.GetServiceName()
	log.Printf("%s.myNotFound, request for %s not routed\n", sn, r.URL.Path)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	_, _ = w.Write([]byte("<h2>404 Not Foundw</h2>"))
}
