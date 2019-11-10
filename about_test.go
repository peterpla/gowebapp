package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAboutHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/about", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(srv.handleHomeOld)

	handler.ServeHTTP(rr, req)

	if got := rr.Code; got != http.StatusOK {
		t.Errorf("expected status code %v, got %v",
			http.StatusOK, got)
	}

	var b []byte
	if b, err = ioutil.ReadAll(rr.Body); err != nil {
		t.Fatalf("ReadAll(rr.Body) error: %v", err)
	}
	content := `nav-link">About</a>`
	h1 := strings.Index(string(b), content)
	if h1 < 0 {
		t.Errorf("expected '%s', not found", content)
	}
}
