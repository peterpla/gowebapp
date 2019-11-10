package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	// replaces deprecated google.golang.org/api/cloudkms/v1
)

type server struct {
	cfg    *config
	router *http.ServeMux
}

// NewServer initializes the app-wide server struct
func NewServer() *server {
	s := &server{}
	s.cfg = &config{}
	return s
}

func main() {
	s := NewServer()

	pCfg := s.cfg

	if err := loadFlagsAndConfig(pCfg); err != nil {
		log.Fatalf("Error loading flags and configuration: %v", err)
	}
	// log.Printf("config file: %q, port: %d, verbose: %t\n", Cfg.configFile, Cfg.port, Cfg.verbose)

	// h := NewHome()
	_ = NewHome()
	s.router = http.NewServeMux()

	s.routes()

	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "/Users/peterplamondon/go/src/github.com/peterpla/gowebapp/public/favicon.ico")
	})

	port := os.Getenv("PORT") // Google App Engine complains if "PORT" env var isn't checked
	if port == "" {
		port = strconv.Itoa(s.cfg.port)
	}
	log.Printf("listening on port %s\n", port)
	err := http.ListenAndServe(":"+port, nil)

	log.Printf("Error return from http.ListenAndServe: %v", err)
}
