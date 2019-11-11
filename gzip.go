package main

import (
	"compress/gzip"
	"io"
	"net/http"
	"os"
	"strings"
)

func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if os.Getenv("GAE_ENV") == "" {
			// not running on Google App Engine, we must handle gzip'ing
			encodings := r.Header.Get("Accept-Encoding")
			if !strings.Contains(encodings, "gzip") { // client does not support gzip
				next.ServeHTTP(w, r)
				return
			}
			w.Header().Add("Content-Encoding", "gzip")
			gzipWriter := gzip.NewWriter(w)
			defer gzipWriter.Close()

			// use an instance of gzipResponseWriter using the ResponseWriter we received,
			// and the gzipWriter we created below
			grw := gzipResponseWriter{
				ResponseWriter: w,
				Writer:         gzipWriter,
			}

			// now we can invoke the next middleware, passing it our gzipResponseWriter
			next.ServeHTTP(grw, r)
		} else {
			// pass the request along - Google App Engine auto-gzip's if the client accepts it
			next.ServeHTTP(w, r)
		}
	})
}

// to pass the gzip writer down the middleware chain, create a new
// ResponseWriter whose Write method delegates to the gzip Write method
type gzipResponseWriter struct {
	http.ResponseWriter
	io.Writer
}

// define our own Write method, to overwrite the embedded Write methods
func (grw gzipResponseWriter) Write(data []byte) (int, error) {
	return grw.Writer.Write(data)
}
