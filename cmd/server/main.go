package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/spf13/viper"

	"github.com/peterpla/lead-expert/pkg/adding"
	"github.com/peterpla/lead-expert/pkg/config"
	"github.com/peterpla/lead-expert/pkg/middleware"
	"github.com/peterpla/lead-expert/pkg/serviceInfo"
)

var Config config.Config

func init() {
	logPrefix := "default.main.init(),"
	if err := config.GetConfig(&Config); err != nil {
		msg := fmt.Sprintf(logPrefix+" GetConfig error: %v", err)
		panic(msg)
	}

	// if Config.IsGAE {
	// 	log.Printf(logPrefix+" GOOGLE_CLOUD_PROJECT %q, Config: %+v", os.Getenv("GOOGLE_CLOUD_PROJECT"), Config)
	// } else {
	// 	log.Printf(logPrefix+" Config: %+v", Config)
	// }
}

func main() {
	// Creating App Engine task handlers: https://cloud.google.com/tasks/docs/creating-appengine-handlers
	// log.Printf("Enter default.main\n")

	// set ServiceName and QueueName appropriately
	prefix := "TaskDefault"
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

	router.POST("/api/v1/requests", postHandler(Config.Adder))

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
	err := http.ListenAndServe(":"+port, middleware.LogReqResp(router))

	log.Printf("Error return from http.ListenAndServe: %v", err)
}

// postHandler returns the handler func for POST /requests
func postHandler(a adding.Service) httprouter.Handle {
	var err error

	sn := serviceInfo.GetServiceName()
	// log.Printf("%s.main.postHandler - enter/exit", sn)
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		// log.Printf("%s.main.postHandler, enter\n", sn)
		startTime := time.Now().UTC().Format(time.RFC3339Nano)

		decoder := json.NewDecoder(r.Body)

		var newRequest adding.Request
		err = decoder.Decode(&newRequest)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// log.Printf("%s.taskHandler - decoded request: %+v\n", sn, newRequest)

		// TODO: pick up custom configuration from request
		// TODO: validate incoming request

		// set RequestID that uniquely identifies this request
		newRequest.RequestID = uuid.New()
		newRequest.AcceptedAt = time.Now().UTC().Format(time.RFC3339Nano)

		// add timestamps and get duration
		var duration time.Duration
		if duration, err = newRequest.AddTimestamps("BeginDefault", startTime, "EndDefault"); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// add the request (e.g., to a queue) for subsequent processing
		returnedReq := a.AddRequest(newRequest)

		w.WriteHeader(http.StatusAccepted)
		w.Header().Set("Content-Type", "application/json")

		// populate a PostResponse struct for the HTTP response, with
		// selected fields of Request (which will have many more fields
		// than we want to return here)
		response := adding.PostResponse{
			RequestID:    returnedReq.RequestID,
			CustomerID:   returnedReq.CustomerID,
			MediaFileURI: returnedReq.MediaFileURI,
			AcceptedAt:   returnedReq.AcceptedAt,
		}

		if err = json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("%s.postHandler, json.NewEncoder.Encode error: +%v\n", sn, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Printf("%s.postHandler, completed in %v, newRequest: %+v\n", sn, duration, newRequest)
	}
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

func myNotFound(w http.ResponseWriter, r *http.Request) {
	// log.Printf("%s.myNotFound, request for %s not routed\n", serviceName, r.URL.Path)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	_, _ = w.Write([]byte("<h2>404 Not Foundw</h2>"))
}
