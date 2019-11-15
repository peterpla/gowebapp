package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGzipMiddleware(t *testing.T) {
	type test struct {
		name        string
		url         string
		header      string
		headerValue string
		status      int
		minSize     int
		maxSize     int
	}

	// NOTE: must adjust minSize and maxSize values below to reflect expected sizes
	tests := []test{
		{name: "no-gzip", url: "/index.html", status: http.StatusOK, minSize: 700, maxSize: 720},
		{name: "gzip", url: "/index.html", header: "Accept-Encoding", headerValue: "gzip", status: http.StatusOK, minSize: 400, maxSize: 420},
	}

	srv := NewServer()

	for _, tc := range tests {
		// log.Printf("%s: testing URL: %s", tc.name, tc.url)
		req, err := http.NewRequest("GET", tc.url, nil)
		if err != nil {
			t.Fatal(err)
		}
		if tc.header != "" {
			req.Header.Add(tc.header, tc.headerValue)
		}

		rr := httptest.NewRecorder()
		// handler := GzipMiddleware(srv.handleAdmin())
		handler := GzipMiddleware(srv.testHandler())

		handler.ServeHTTP(rr, req)

		if got := rr.Code; got != tc.status {
			t.Errorf("%s: expected status code: %v, got %v",
				tc.name, tc.status, got)
		}

		len := rr.Body.Len()
		// log.Printf("ContentLength: %d", len)
		if len < tc.minSize || len > tc.maxSize {
			t.Errorf("%s: ContentLength expected %d - %d, got %d\nContent: %s\n", tc.name, tc.minSize, tc.maxSize, len, rr.Body)
		}
	}
}

// testHandler returns a http.HandlerFunc that writes a response that
// reveals whether gzip compression was performed
func (srv *server) testHandler() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		// write lorem ipsum text as response
		fmt.Fprintf(rw,
			`Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nunc tristique, est ac porttitor tincidunt, arcu turpis tempor elit, in suscipit velit sapien sed velit. Mauris vitae elit mollis, auctor turpis et, pretium urna. In elementum fermentum convallis. Mauris euismod sollicitudin pulvinar. Ut ligula velit, pretium vitae facilisis eget, volutpat eu justo. Vestibulum vestibulum efficitur vestibulum. Duis lacinia vitae libero dictum tempus. Duis in tellus vitae sem venenatis maximus.
		
		Donec varius efficitur ex et vestibulum. Praesent ut scelerisque nisi. Morbi sed turpis molestie, elementum odio at, hendrerit dolor. Nam placerat nulla non tortor viverra commodo. Morbi pellentesque sodales nibh, at congue.`)
	}
}
