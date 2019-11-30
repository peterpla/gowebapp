package server

import (
	"log"
	"net/http"
	"os"

	"github.com/peterpla/gowebapp/pkg/adding"
	"github.com/peterpla/gowebapp/pkg/config"
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
	IsGAE       bool
}

// NewServer initializes the app-wide Server struct
func NewServer() *Server {
	s := &Server{}
	s.IsGAE = false
	s.storageType = Memory // TODO: configurable storage type
	if os.Getenv("GAE_ENV") != "" {
		s.IsGAE = true
		s.storageType = GCTQueue
	}
	// log.Printf("NewServer before LoadFlagsAndConfig, s: %+v\n", s)
	s.Cfg = &config.Config{}
	if err := config.LoadFlagsAndConfig(s.Cfg); err != nil {
		log.Fatalf("Error loading flags and configuration: %v", err)
	}

	switch s.storageType {
	case Memory:
		storage := new(memory.Storage)
		s.Adder = adding.NewService(storage)

	case GCTQueue:
		storage := new(queue.GCT)
		s.Adder = adding.NewService(storage)

	default:
		panic("unsupported storageType")
	}

	log.Printf("NewServer exiting\n... s: %+v\n... s.Cfg: %+v\n", s, s.Cfg)
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("Enter server.ServeHTTP\n")
	h := s.Router
	h.ServeHTTP(w, r)
	log.Printf("Exit server.ServeHTTP\n")
}
