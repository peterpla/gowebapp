package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/peterpla/gowebapp/pkg/adding"
	"github.com/peterpla/gowebapp/pkg/config"
	"github.com/peterpla/gowebapp/pkg/http/rest"
	"github.com/peterpla/gowebapp/pkg/middleware"
	"github.com/peterpla/gowebapp/pkg/storage/memory"
	"github.com/peterpla/gowebapp/pkg/storage/queue"
)

// StorageType defines available storage types
type Type int

const (
	// Memory - store data in memory
	Memory Type = iota
	// Cloud Tasks queue - add data to Google Cloud Tasks queue
	GCTQueue
)

type Server struct {
	Cfg         *config.Config
	Router      http.Handler
	storageType Type
	Adder       adding.Service
	isGAE       bool
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("Enter server.ServeHTTP\n")
	h := s.Router
	h.ServeHTTP(w, r)
	log.Printf("Exit server.ServeHTTP\n")
}

// NewServer initializes the app-wide Server struct
func NewServer() *Server {
	s := &Server{}
	s.storageType = GCTQueue // TODO: configurable storage type
	s.isGAE = false
	if os.Getenv("GAE_ENV") != "" {
		s.isGAE = true
	}
	s.Cfg = &config.Config{}
	return s
}

var srv *Server

func main() {
	srv = NewServer()
	if err := config.LoadFlagsAndConfig(srv.Cfg); err != nil {
		log.Fatalf("Error loading flags and configuration: %v", err)
	}
	log.Printf("main, config: %+v\n", srv.Cfg)

	switch srv.storageType {
	case Memory:
		storage := new(memory.Storage)
		srv.Adder = adding.NewService(storage)

	case GCTQueue:
		storage := new(queue.GCT)
		srv.Adder = adding.NewService(storage)

	default:
		panic("unsupported storageType")
	}

	newRouter := rest.Routes(srv.Adder)

	port := os.Getenv("PORT") // Google App Engine complains if "PORT" env var isn't checked
	if port == "" {
		port = strconv.Itoa(srv.Cfg.Port)
	}
	log.Printf("Service default listening on port %s\n", port)

	srv.Router = newRouter
	err := http.ListenAndServe(":"+port, middleware.LogReqResp(newRouter))

	log.Printf("Error return from http.ListenAndServe: %v", err)
}
