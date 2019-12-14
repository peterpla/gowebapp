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
	"github.com/peterpla/lead-expert/pkg/storage/memory"
)

func TestRequestsPost(t *testing.T) {

	type test struct {
		name     string
		endpoint string
		body     string
		respBody string
		status   int
	}

	tests := []test{
		{name: "POST requests",
			endpoint: "/requests",
			body:     `{ "customer_id": "12345", "media_url": "gs://elated-practice-224603.appspot.com/audio_uploads/audio-02.mp3", "custom_config": false }`,
			respBody: "accepted_at",
			status:   http.StatusAccepted},
	}

	storage := new(memory.Storage)
	adder := adding.NewService(storage)

	// port := os.Getenv("PORT") // needed for localhost testing, not for GAE
	apiPrefix := "/api/v1"

	// IMPORTANT: comment/uncomment to change where the app is running
	// prefix := fmt.Sprintf("http://localhost:%s%s", port, apiPrefix)
	prefix := fmt.Sprintf("https://%s.appspot.com%s", os.Getenv("PROJECT_ID"), apiPrefix)

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
