package main

import (
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
		{name: "no-gzip", url: "/home", status: http.StatusOK, minSize: 2 * 1014, maxSize: 3 * 1024},
		{name: "gzip", url: "/home", header: "Accept-Encoding", headerValue: "gzip", status: http.StatusOK, minSize: 750, maxSize: 1 * 1024},
	}

	srv := &server{}

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
		handler := GzipMiddleware(srv.handleAdmin())

		handler.ServeHTTP(rr, req)

		if got := rr.Code; got != tc.status {
			t.Errorf("%s: expected status code: %v, got %v",
				tc.name, tc.status, got)
		}

		len := rr.Body.Len()
		// log.Printf("ContentLength: %d", len)
		if len < tc.minSize || len > tc.maxSize {
			t.Errorf("%s: ContentLength expected %d - %d, got %d", tc.name, tc.minSize, tc.maxSize, len)
		}
	}
}
