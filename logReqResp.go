package main

import (
	"log"
	"net/http"
	"os"
)

func LogReqResp(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// log.Printf("Enter LogReqResp, next: %p\n", next)
		rec := &statusRecorder{w, 200}

		next.ServeHTTP(rec, r)

		if os.Getenv("GAE_ENV") == "" {
			log.Printf("%s %s %s %d\n", r.RemoteAddr, r.Method, r.URL, rec.status)
		}
		// log.Printf("Exit LogReqResp\n")
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

// WriteHeader wraps the passed-in ResponseWriter's WriteHeader
func (rec *statusRecorder) WriteHeader(code int) {
	rec.status = code                    // save the status code
	rec.ResponseWriter.WriteHeader(code) // pass it on to wrapped method
}
