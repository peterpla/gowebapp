package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/julienschmidt/httprouter"

	"github.com/peterpla/lead-expert/pkg/adding"
	"github.com/peterpla/lead-expert/pkg/config"
	"github.com/peterpla/lead-expert/pkg/storage/memory"
)

func TestDefaultPost(t *testing.T) {

	cfg := config.GetConfigPointer()
	servicePrefix := ""
	port := cfg.TaskDefaultPort

	type test struct {
		name     string
		endpoint string
		body     string
		respBody string
		status   int
	}

	tests := []test{
		// valid
		{name: "valid POST requests",
			endpoint: "/requests",
			body:     `{ "customer_id": 1234567, "media_uri": "gs://elated-practice-224603.appspot.com/audio_uploads/audio-01.mp3" }`,
			respBody: "accepted_at",
			status:   http.StatusAccepted},
		// bad customer_id
		{name: "string customer_id",
			endpoint: "/requests",
			body:     `{ "customer_id": "nope", "media_uri": "gs://elated-practice-224603.appspot.com/audio_uploads/audio-01.mp3" }`,
			respBody: "invalid value for the \"customer_id\"",
			status:   http.StatusBadRequest},
		{name: "zero customer_id",
			endpoint: "/requests",
			body:     `{ "customer_id": 0, "media_uri": "gs://elated-practice-224603.appspot.com/audio_uploads/audio-01.mp3" }`,
			respBody: "Error:Field validation for 'CustomerID'",
			status:   http.StatusBadRequest},
		{name: "negative customer_id",
			endpoint: "/requests",
			body:     `{ "customer_id": -1, "media_uri": "gs://elated-practice-224603.appspot.com/audio_uploads/audio-01.mp3" }`,
			respBody: "Error:Field validation for 'CustomerID'",
			status:   http.StatusBadRequest},
		{name: "too big customer_id",
			endpoint: "/requests",
			body:     `{ "customer_id": 12345678, "media_uri": "gs://elated-practice-224603.appspot.com/audio_uploads/audio-01.mp3" }`,
			respBody: "Error:Field validation for 'CustomerID'",
			status:   http.StatusBadRequest},
		// bad media_uri
		{name: "invalid media_uri",
			endpoint: "/requests",
			body:     `{ "customer_id": 1234567, "media_uri": "lollipop" }`,
			respBody: "Error:Field validation for 'MediaFileURI'",
			status:   http.StatusBadRequest},
	}

	storage := new(memory.Storage)
	adder := adding.NewService(storage)

	apiPrefix := "/api/v1"

	prefix := fmt.Sprintf("http://localhost:%s%s", port, apiPrefix)
	if cfg.IsGAE {
		prefix = fmt.Sprintf("https://%s%s.appspot.com%s", servicePrefix, os.Getenv("PROJECT_ID"), apiPrefix)
	}

	for _, tc := range tests {
		url := prefix + tc.endpoint
		// log.Printf("Test %s: %s", tc.name, url)

		router := httprouter.New()
		router.POST("/api/v1/requests", postHandler(adder))

		// build the POST request with custom header
		theRequest, err := http.NewRequest("POST", url, strings.NewReader(tc.body))
		if err != nil {
			t.Fatal(err)
		}

		// response recorder
		rr := httptest.NewRecorder()

		// send the request
		router.ServeHTTP(rr, theRequest)

		if tc.status != rr.Code {
			t.Errorf("%s: %q expected status code %v, got %v", tc.name, tc.endpoint, tc.status, rr.Code)
		}

		if tc.respBody != "" {
			var b []byte
			if b, err = ioutil.ReadAll(rr.Body); err != nil {
				t.Fatalf("%s: ReadAll error: %v", tc.name, err)
			}

			if !strings.Contains(string(b), tc.respBody) {
				t.Errorf("%s: expected %q, not found (in %q)", tc.name, tc.respBody, string(b))
			}
		}
	}
}
