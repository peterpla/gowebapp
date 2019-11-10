package main

import "net/http"

func (s *server) routes() {
	http.HandleFunc("/", s.handleHome)
	http.HandleFunc("/home", s.handleHome)
}
