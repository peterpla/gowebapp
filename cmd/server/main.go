package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
	"github.com/spf13/viper"

	"github.com/peterpla/gowebapp/pkg/adding"
	"github.com/peterpla/gowebapp/pkg/config"
	"github.com/peterpla/gowebapp/pkg/middleware"
	"github.com/peterpla/gowebapp/pkg/serviceInfo"
)

var Config config.Config

func init() {
	logPrefix := "default.main.init(),"
	if err := config.GetConfig(&Config); err != nil {
		msg := fmt.Sprintf(logPrefix+" GetConfig error: %v", err)
		panic(msg)
	}
	// log.Printf(logPrefix+" Config: %+v", Config)
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

	router.POST("/api/v1/requests", addRequest(Config.Adder))

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

// addRequest returns a handler for POST /requests
func addRequest(a adding.Service) httprouter.Handle {
	serviceName := Config.ServiceName
	// log.Printf("%s.main.AddRequest - enter/exit", serviceName)
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		// log.Printf("%s.main.AddRequest handler - enter\n", serviceName)
		decoder := json.NewDecoder(r.Body)

		var newRequest adding.Request
		err := decoder.Decode(&newRequest)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Printf("%s.taskHandler - decoded request: %+v\n", serviceName, newRequest)

		a.AddRequest(newRequest)

		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode("New request added.")

		// log.Printf("%s.AddRequest handler - exit\n", serviceName)
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
