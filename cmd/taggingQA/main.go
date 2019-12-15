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
	"github.com/peterpla/lead-expert/pkg/config"
	"github.com/peterpla/lead-expert/pkg/middleware"
	"github.com/peterpla/lead-expert/pkg/serviceInfo"
)

var Config config.Config

func init() {
	logPrefix := "tagging-qa.main.init(),"
	if err := config.GetConfig(&Config); err != nil {
		msg := fmt.Sprintf(logPrefix+" GetConfig error: %v", err)
		panic(msg)
	}
	// log.Printf(logPrefix+" Config: %+v", Config)
}

func main() {
	// Creating App Engine task handlers: https://cloud.google.com/tasks/docs/creating-appengine-handlers
	// log.Printf("Enter tagging-complete.main, Config: %+v\n", Config)

	// set ServiceName and QueueName appropriately
	prefix := "TaskTaggingQA"
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
		// log.Printf("%s.indexHandler, r.URL.Path: %s, will respond NotFound\n", serviceName, r.URL.Path)
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

		// TODO: validation incoming request

		newRequest := incomingRequest

		// add timestamps and get duration
		var duration time.Duration
		if duration, err = newRequest.AddTimestamps("BeginTaggingQA", startTime, "EndTaggingQA"); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// create task on the next pipeline stage's queue with updated request
		a.AddRequest(newRequest)

		// Log & output details of the created task.
		output := fmt.Sprintf("%s.taskHandler completed in %v: queue %q, task %q, newRequest: %+v",
			serviceName, duration, queueName, taskName, newRequest)
		log.Println(output)

		// Set a non-2xx status code to indicate a failure in task processing that should be retried.
		// For example, http.Error(w, "Internal Server Error: Task Processing", http.StatusInternalServerError)
		w.WriteHeader(http.StatusOK)
	}
}

func myNotFound(w http.ResponseWriter, r *http.Request) {
	// log.Printf("%s.myNotFound, request for %s not routed\n", serviceName, r.URL.Path)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	_, _ = w.Write([]byte("<h2>404 Not Foundw</h2>"))
}
