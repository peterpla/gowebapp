package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/peterpla/gowebapp/pkg/middleware"
)

type server struct {
	cfg    *config
	router *http.ServeMux
	isGAE  bool
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("Enter server.ServeHTTP\n")
	h := s.router
	h.ServeHTTP(w, r)
	log.Printf("Exit server.ServeHTTP\n")
}

// NewServer initializes the app-wide server struct
func NewServer() *server {
	s := &server{}
	s.isGAE = false
	if os.Getenv("GAE_ENV") != "" {
		s.isGAE = true
	}
	s.cfg = &config{}
	return s
}

var srv *server

func main() {
	srv = NewServer()
	if err := loadFlagsAndConfig(srv.cfg); err != nil {
		log.Fatalf("Error loading flags and configuration: %v", err)
	}
	// log.Printf("config: %+v\n", srv.cfg)

	http.HandleFunc("/api/v1", apiHandler())
	// FileServer returns a handler that serves HTTP requests with the
	// contents of the file system rooted at http.Dir("/root").
	// As a special case, the returned file server redirects any
	// request ending in "/index.html" to the same path, without
	// the final "index.html".
	http.Handle("/", http.FileServer(http.Dir("../../public/")))
	http.HandleFunc("/favicon.ico", faviconHandler)

	// show all routes
	// v := reflect.ValueOf(http.DefaultServeMux).Elem()
	// log.Printf("routes: %+v\n", v.FieldByName("m"))

	port := os.Getenv("PORT") // Google App Engine complains if "PORT" env var isn't checked
	if port == "" {
		port = strconv.Itoa(srv.cfg.port)
	}
	log.Printf("listening on port %s\n", port)
	srv.router = http.DefaultServeMux
	err := http.ListenAndServe(":"+port, middleware.LogReqResp(http.DefaultServeMux))

	log.Printf("Error return from http.ListenAndServe: %v", err)
}

func apiHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("apiHandler\n")
	}
}

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "../../public/favicon.ico")
}
