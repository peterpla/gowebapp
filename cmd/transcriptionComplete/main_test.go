package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-playground/validator"
	"github.com/julienschmidt/httprouter"

	"github.com/peterpla/lead-expert/pkg/config"
	"github.com/peterpla/lead-expert/pkg/database"
	"github.com/peterpla/lead-expert/pkg/queue"
)

func TestTranscriptionComplete(t *testing.T) {

	cfg := config.GetConfigPointer()
	servicePrefix := "transcription-complete-dot-" // <---- change to match service!!
	port := cfg.TaskTranscriptionCompletePort      // <---- change to match service!!
	repo = database.NewFirestoreRequestRepository(cfg.ProjectID, cfg.DatabaseRequests)

	validate = validator.New()

	type test struct {
		name     string
		method   string
		endpoint string
		body     string
		respBody string
		status   int
	}

	jsonBody := fmt.Sprintf("{ \"customer_id\": %7d, \"media_uri\": %q, \"accepted_at\": %q }",
		1234567, "gs://elated-practice-224603.appspot.com/audio_uploads/audio-02.mp3", time.Now().UTC().Format(time.RFC3339Nano))

	tests := []test{
		// valid
		{name: "valid POST /task_handler",
			method:   "POST",
			endpoint: "/task_handler",
			body:     jsonBody,
			status:   http.StatusOK},
		{name: "valid GET /",
			method:   "GET",
			endpoint: "/",
			respBody: "service running",
			status:   http.StatusOK},
		{name: "invalid GET /nope",
			method:   "GET",
			endpoint: "/nope",
			respBody: "not found",
			status:   http.StatusNotFound},
	}

	qi = queue.QueueInfo{}
	q = queue.NewNullQueue(&qi) // use null queue, requests thrown away on exit
	// q = queue.NewGCTQueue(&qi) // use Google Cloud Tasks

	prefix := fmt.Sprintf("http://localhost:%s", port)
	if cfg.IsGAE {
		prefix = fmt.Sprintf("https://%s%s.appspot.com", servicePrefix, os.Getenv("PROJECT_ID"))
	}

	for _, tc := range tests {
		url := prefix + tc.endpoint
		// log.Printf("Test %s: %s", tc.name, url)

		router := httprouter.New()
		router.POST("/task_handler", taskHandler(q))
		router.GET("/", indexHandler)

		// build the POST request with custom header
		theRequest, err := http.NewRequest(tc.method, url, strings.NewReader(tc.body))
		if err != nil {
			t.Fatal(err)
		}
		theRequest.Header.Set("X-Appengine-Taskname", "localTask")
		theRequest.Header.Set("X-Appengine-Queuename", "localQueue")

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
			// log.Printf("%s: rr.Body: %q\n", tc.name, string(b))

			if !strings.Contains(string(b), tc.respBody) {
				t.Errorf("%s: expected %q, not found (in %q)", tc.name, tc.respBody, string(b))
			}
		}
	}
}
