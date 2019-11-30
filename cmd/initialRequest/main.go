package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
)

var service = os.Getenv("TASK_INITIAL_REQUEST_WRITE_TO_Q")

func main() {
	// Creating App Engine task handlers: https://cloud.google.com/tasks/docs/creating-appengine-handlers
	log.Printf("Enter initial-request.main\n")

	router := httprouter.New()

	// Default endpoint Cloud Tasks sends to is /task_handler
	router.POST("/task_handler", taskHandler())

	// custom NotFound handler
	router.NotFound = http.HandlerFunc(myNotFound)

	// Allow confirmation the task handling service is running.
	router.GET("/", indexHandler)

	port := os.Getenv("TASK_INITIAL_REQUEST_PORT")
	if port == "" {
		port = "8081"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Service initial-request listening on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}

// indexHandler responds to requests with "service running"
func indexHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	log.Printf("Enter initial-request.indexHandler\n")
	if r.URL.Path != "/" {
		log.Printf("initial-request.indexHandler, r.URL.Path: %s, will respond NotFound\n", r.URL.Path)
		http.NotFound(w, r)
		return
	}
	// indicate service is running
	fmt.Fprint(w, "initial-request service running.")
}

// taskHandler processes task requests.
func taskHandler() httprouter.Handle {
	// log.Printf("initial-request.taskHandler - enter/exit\n")
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		// log.Printf("initial-request.taskHandler - enter handler\n")
		// log.Printf("request: %+v\n", r)
		// log.Printf("params: %+v\n", p)

		t, ok := r.Header["X-Appengine-Taskname"]
		if !ok || len(t[0]) == 0 {
			// You may use the presence of the X-Appengine-Taskname header to validate
			// the request comes from Cloud Tasks.
			log.Println("Invalid Task: No X-Appengine-Taskname request header found")
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
			// log.Printf("initial-request.main, ReadAll error: %v", err)
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			return
		}

		// Log & output details of the task.
		output := fmt.Sprintf("initial-request.taskHandler, completed: queue %s, task %s\n... payload: %s",
			queueName, taskName, string(body))
		log.Println(output)

		// Set a non-2xx status code to indicate a failure in task processing that should be retried.
		// For example, http.Error(w, "Internal Server Error: Task Processing", http.StatusInternalServerError)
		w.WriteHeader(http.StatusOK)

		// log.Printf("initial-request.taskHandler - exit hander\n")
	}
}

func myNotFound(w http.ResponseWriter, r *http.Request) {
	log.Printf("initial-request.myNotFound, request for %s not routed\n", r.URL.Path)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	_, _ = w.Write([]byte("<h2>404 Not Foundw</h2>"))
}
