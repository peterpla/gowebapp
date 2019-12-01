package main

import (
	"log"
	"net/http"
	"os"

	"github.com/peterpla/gowebapp/pkg/http/rest"
	"github.com/peterpla/gowebapp/pkg/middleware"
	"github.com/peterpla/gowebapp/pkg/server"
)

func main() {
	s := server.NewServer() // processes env vars and config file
	serviceName := s.Cfg.TaskDefaultSvc
	queueName := s.Cfg.TaskDefaultWriteToQ

	newRouter := rest.Routes(s.Adder)
	s.Router = newRouter

	port := os.Getenv("PORT") // Google App Engine complains if "PORT" env var isn't checked
	if !s.IsGAE {
		port = os.Getenv("TASK_DEFAULT_PORT")
	}
	if port == "" {
		port = s.Cfg.TaskDefaultPort
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Service %s listening on port %s, requests will be added to queue %s", serviceName, port, queueName)
	err := http.ListenAndServe(":"+port, middleware.LogReqResp(newRouter))

	log.Printf("Error return from http.ListenAndServe: %v", err)
}
