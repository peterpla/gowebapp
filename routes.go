package main

import (
	"net/http"
)

func (srv *server) routes() {
	http.HandleFunc("/", srv.handleHomeOld) // not equivalent to s.router.HandleFunc ???
	http.HandleFunc("/home", srv.handleHomeOld)
	http.HandleFunc("/about", srv.handleAbout)
}
