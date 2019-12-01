package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"

	"github.com/peterpla/gowebapp/pkg/adding"
	"github.com/peterpla/gowebapp/pkg/middleware"
	"github.com/peterpla/gowebapp/pkg/server"
)

var serviceName, queueName string

func main() {
	// Creating App Engine task handlers: https://cloud.google.com/tasks/docs/creating-appengine-handlers
	// log.Printf("Enter service-dispatch.main\n")

	s := server.NewServer() // processes env vars and config file
	serviceName := s.Cfg.TaskServiceDispatchSvc
	queueName := s.Cfg.TaskServiceDispatchWriteToQ

	router := httprouter.New()
	s.Router = router

	// Default endpoint Cloud Tasks sends to is /task_handler
	router.POST("/task_handler", taskHandler(s.Adder, serviceName))

	// custom NotFound handler
	router.NotFound = http.HandlerFunc(myNotFound)

	// Allow confirmation the task handling service is running.
	router.GET("/", indexHandler)

	port := os.Getenv("PORT") // Google App Engine complains if "PORT" env var isn't checked
	if !s.IsGAE {
		port = os.Getenv("TASK_SERVICE_DISPATCH_PORT")
	}
	if port == "" {
		port = s.Cfg.TaskInitialRequestPort
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Service %s listening on port %s, requests will be added to queue %s", serviceName, port, queueName)
	log.Fatal(http.ListenAndServe(":"+port, middleware.LogReqResp(router)))
}

// indexHandler responds to requests with "service running"
func indexHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	log.Printf("Enter %s.indexHandler\n", serviceName)
	if r.URL.Path != "/" {
		log.Printf("%s.indexHandler, r.URL.Path: %s, will respond NotFound\n", serviceName, r.URL.Path)
		http.NotFound(w, r)
		return
	}
	// indicate service is running
	fmt.Fprintf(w, "%s service running\n", serviceName)
}

// taskHandler processes task requests.
func taskHandler(a adding.Service, serviceName string) httprouter.Handle {
	log.Printf("%s.taskHandler - enter/exit\n", serviceName)
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		// log.Printf("%s.taskHandler - enter handler\n", serviceName)
		// log.Printf("... request: %+v\n", r)
		// log.Printf("... params: %+v\n", p)

		t, ok := r.Header["X-Appengine-Taskname"]
		if !ok || len(t[0]) == 0 {
			// You may use the presence of the X-Appengine-Taskname header to validate
			// the request comes from Cloud Tasks.
			log.Printf("%s: Invalid Task: No X-Appengine-Taskname request header found\n", serviceName)
			http.Error(w, "Bad Request - Invalid Task", http.StatusBadRequest)
			return
		}
		taskName := t[0]

		// Pull useful headers from Task request.
		q, ok := r.Header["X-Appengine-Queuename"]
		queueName = ""
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

		// decode incoming request
		var incomingRequest adding.Request

		decoder := json.NewDecoder(bytes.NewReader(body))
		err = decoder.Decode(&incomingRequest)
		if err != nil {
			log.Printf("%s.taskHandler, json.Decode error: %v", serviceName, err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Printf("%s.taskHandler - decoded request: %+v\n", serviceName, incomingRequest)

		// TODO: validation incoming request

		// TODO: create task on TranscriptionGCP queue with updated request
		newRequest := incomingRequest
		a.AddRequest(newRequest)

		// Log & output details of the task.
		output := fmt.Sprintf("%s.taskHandler completed: queue %q, task %q\n... payload: %+v",
			serviceName, queueName, taskName, newRequest)
		log.Println(output)

		// Set a non-2xx status code to indicate a failure in task processing that should be retried.
		// For example, http.Error(w, "Internal Server Error: Task Processing", http.StatusInternalServerError)
		w.WriteHeader(http.StatusOK)

		// log.Printf("%s.taskHandler - exit hander\n", serviceName)
	}
}

func myNotFound(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s.myNotFound, request for %s not routed\n", serviceName, r.URL.Path)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	_, _ = w.Write([]byte("<h2>404 Not Foundw</h2>"))
}
