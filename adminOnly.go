package main

import (
	"fmt"
	"net/http"
)

func (srv *server) adminOnly(w http.ResponseWriter, r *http.Request) {
	if !currentUserIsAdmin(r) {
		http.NotFound(w, r)
		return
	}
	fmt.Fprintf(w, "<h1>Admin</h1>")
}

// func (srv *server) adminOnly(next http.HandlerFunc) http.HandlerFunc {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		if !currentUserIsAdmin(r) {
// 			http.NotFound(w, r)
// 			return
// 		}
// 		next.ServeHTTP(w, r)
// 	})
// }

func currentUserIsAdmin(r *http.Request) bool {
	// MOCK: pretend user is an admin EXCEPT when accessing "/admin"
	return r.URL.RawQuery == "loggedIn=true"
}
