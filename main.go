package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	// replaces deprecated google.golang.org/api/cloudkms/v1
)

func main() {
	if err := loadFlagsAndConfig(&Cfg); err != nil {
		log.Fatalf("Error loading flags and configuration: %v", err)
	}
	// log.Printf("config file: %q, port: %d, verbose: %t\n", Cfg.configFile, Cfg.port, Cfg.verbose)

	h := NewHome()
	h.registerRoutes()
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "/Users/peterplamondon/go/src/github.com/peterpla/gowebapp/public/favicon.ico")
	})

	port := os.Getenv("PORT") // Google App Engine complains if "PORT" env var isn't checked
	if port == "" {
		port = strconv.Itoa(Cfg.port)
	}
	log.Printf("listening on port %s\n", port)
	err := http.ListenAndServe(":"+port, nil)

	log.Printf("Error return from http.ListenAndServe: %v", err)
}
