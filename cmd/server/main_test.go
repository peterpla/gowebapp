package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/julienschmidt/httprouter"

	"github.com/peterpla/lead-expert/pkg/adding"
	"github.com/peterpla/lead-expert/pkg/config"
	"github.com/peterpla/lead-expert/pkg/storage/memory"
)

func TestDefaultPost(t *testing.T) {

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
			body:     `{ "customer_id": 1234567, "media_uri": "gs://elated-practice-224603.appspot.com/audio_uploads/audio-02.mp3" }`,
			respBody: "accepted_at",
			status:   http.StatusAccepted},
		// bad customer_id
		{name: "string customer_id",
			endpoint: "/requests",
			body:     `{ "customer_id": "nope", "media_uri": "gs://elated-practice-224603.appspot.com/audio_uploads/audio-02.mp3" }`,
			respBody: "invalid value for the \"customer_id\"",
			status:   http.StatusBadRequest},
		{name: "zero customer_id",
			endpoint: "/requests",
			body:     `{ "customer_id": 0, "media_uri": "gs://elated-practice-224603.appspot.com/audio_uploads/audio-02.mp3" }`,
			respBody: "Error:Field validation for 'CustomerID'",
			status:   http.StatusBadRequest},
		{name: "negative customer_id",
			endpoint: "/requests",
			body:     `{ "customer_id": -1, "media_uri": "gs://elated-practice-224603.appspot.com/audio_uploads/audio-02.mp3" }`,
			respBody: "Error:Field validation for 'CustomerID'",
			status:   http.StatusBadRequest},
		{name: "too big customer_id",
			endpoint: "/requests",
			body:     `{ "customer_id": 12345678, "media_uri": "gs://elated-practice-224603.appspot.com/audio_uploads/audio-02.mp3" }`,
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

	cfg := config.GetConfigPointer()
	port := cfg.TaskDefaultPort
	prefix := fmt.Sprintf("http://localhost:%s%s", port, apiPrefix)
	if cfg.IsGAE {
		prefix = fmt.Sprintf("https://%s.appspot.com%s", os.Getenv("PROJECT_ID"), apiPrefix)
	}

	for _, tc := range tests {
		url := prefix + tc.endpoint
		// log.Printf("Test %s: %s", tc.name, url)

		router := httprouter.New()
		router.POST("/api/v1/requests", postHandler(adder))

		// POST it
		resp, err := http.Post(url, "application/json", bytes.NewBufferString(tc.body))
		if err != nil {
			t.Fatalf("%s: http.Post error: %v", tc.name, err)
		}

		if tc.status != resp.StatusCode {
			t.Errorf("%s: %q expected status code %v, got %v", tc.name, tc.endpoint, tc.status, resp.StatusCode)
		}

		if tc.respBody != "" {
			var b []byte
			if b, err = ioutil.ReadAll(resp.Body); err != nil {
				t.Fatalf("%s: ReadAll error: %v", tc.name, err)
			}

			if !strings.Contains(string(b), tc.respBody) {
				t.Errorf("%s: expected %q, not found (in %q)", tc.name, tc.respBody, string(b))
			}
		}
	}
}
