package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/peterpla/gowebapp/pkg/http/rest"
	"github.com/peterpla/gowebapp/pkg/middleware"
	"github.com/peterpla/gowebapp/pkg/server"
)

var queueName = os.Getenv("TASKS_Q_REQUESTS")     // queue to add new requests to
var serviceName = os.Getenv("TASKS_SVC_REQUESTS") // service we're running in

var srv *server.Server

func main() {
	srv = server.NewServer()

	newRouter := rest.Routes(srv.Adder)

	port := os.Getenv("TASKS_PORT_REQUESTS") // Google App Engine complains if "PORT" env var isn't checked
	if port == "" {
		port = strconv.Itoa(srv.Cfg.Port)
	}
	log.Printf("Service %s listening on port %s, requests will be added to queue %s", serviceName, port, queueName)

	srv.Router = newRouter
	err := http.ListenAndServe(":"+port, middleware.LogReqResp(newRouter))

	log.Printf("Error return from http.ListenAndServe: %v", err)
}
