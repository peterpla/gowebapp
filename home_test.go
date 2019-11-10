package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHomeHandler(t *testing.T) {

	type test struct {
		name    string
		url     string
		status  int
		content string
	}

	tests := []test{
		{name: "empty", url: "", status: http.StatusOK, content: "nav-link active\">Home</a>"},
		{name: "slash", url: "/", status: http.StatusOK, content: "nav-link active\">Home</a>"},
		{name: "home", url: "/home", status: http.StatusOK, content: "nav-link active\">Home</a>"},
	}

	for _, tc := range tests {
		// log.Printf("%s: testing URL: %s", tc.name, tc.url)
		req, err := http.NewRequest("GET", tc.url, nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(srv.handleHomeOld)

		handler.ServeHTTP(rr, req)

		if got := rr.Code; got != tc.status {
			t.Errorf("%s: expected status code: %v, got %v",
				tc.name, tc.status, got)
		}

		var b []byte
		if b, err = ioutil.ReadAll(rr.Body); err != nil {
			t.Fatalf("%s: ReadAll(rr.Body) error: %v", tc.name, err)
		}
		h1 := strings.Index(string(b), tc.content)
		if h1 < 0 {
			t.Errorf("%s: expected '%s', not found", tc.name, tc.content)
		}
	}
}
