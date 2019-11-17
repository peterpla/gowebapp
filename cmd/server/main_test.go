package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	// log.Printf("srv: %v\n", srv)
	srv = NewServer()
	if err := loadFlagsAndConfig(srv.cfg); err != nil {
		log.Fatalf("Error loading flags and configuration: %v", err)
	}
	// log.Printf("config: %+v\n", srv.cfg)

	os.Exit(m.Run())
}

func TestEndpoints(t *testing.T) {

	type test struct {
		name    string
		url     string
		status  int
		content string
	}

	tests := []test{
		{name: "no-path", url: "", status: http.StatusOK, content: "<h1>Learn to Create Websites</h1>"},
		{name: "index", url: "index.html", status: http.StatusOK, content: "<h1>Learn to Create Websites</h1>"},
		{name: "about", url: "about.html", status: http.StatusOK, content: "<h1>About</h1>"},
		// {name: "favicon", url: "favicon.ico", status: http.StatusOK},
	}

	// IMPORTANT: comment/uncomment to change where the app is running
	prefix := "http://localhost:" + strconv.Itoa(srv.cfg.port)
	// prefix := fmt.Sprintf("https://%s.appspot.com", srv.cfg.projectID)

	for _, tc := range tests {
		endpoint := tc.url
		url := prefix + "/" + endpoint
		// log.Printf("Test %s: %s", tc.name, url)

		response, err := http.Get(url)
		if err != nil {
			t.Fatalf("%s: http.Get error: %v", tc.name, err)
		}

		if response.StatusCode != tc.status {
			t.Errorf("%s: %q expected status code %v, got %v", tc.name, tc.url, tc.status, response.StatusCode)
		}

		if tc.content != "" {
			var b []byte
			if b, err = ioutil.ReadAll(response.Body); err != nil {
				t.Fatalf("%s: ReadAll error: %v", tc.name, err)
			}

			if !strings.Contains(string(b), tc.content) {
				t.Errorf("%s: expected %q, not found (in %q)", tc.name, tc.content, string(b))
			}
		}
	}
}
