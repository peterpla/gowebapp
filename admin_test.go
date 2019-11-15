package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAdminHandler(t *testing.T) {
	type test struct {
		name    string
		url     string
		params  map[string]string
		status  int
		content string
	}

	tests := []test{
		{name: "loggedIn", url: "/admin", params: map[string]string{"loggedIn": "true"}, status: http.StatusOK, content: "<h1>Admin</h1>"},
		{name: "notLoggedIn", url: "/admin", status: http.StatusNotFound},
	}

	srv := NewServer()

	for _, tc := range tests {
		// log.Printf("%s: testing URL: %s", tc.name, tc.url)
		req, err := http.NewRequest("GET", tc.url, nil)
		if err != nil {
			t.Fatal(err)
		}

		// populate the query params
		q := req.URL.Query()
		for k, v := range tc.params {
			q.Add(k, v)
		}
		req.URL.RawQuery = q.Encode()

		rr := httptest.NewRecorder()
		srv.adminOnly(rr, req)

		if got := rr.Code; got != tc.status {
			t.Errorf("%s: expected status %v, got %v",
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
