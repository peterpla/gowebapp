package main

import (
	"net/http"
)

func (srv *server) routes() {
	http.HandleFunc("/", srv.handleHome()) // not equivalent to s.router.HandleFunc ???
	http.HandleFunc("/home", srv.handleHome())
	http.HandleFunc("/about", srv.handleAbout)
	http.HandleFunc("/admin", srv.adminOnly(srv.handleAdmin()))

	// show all routes
	// v := reflect.ValueOf(http.DefaultServeMux).Elem()
	// log.Printf("routes: %+v\n", v.FieldByName("m"))
}
