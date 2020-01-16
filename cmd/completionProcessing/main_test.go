package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-playground/validator"
	"github.com/julienschmidt/httprouter"

	"github.com/peterpla/lead-expert/pkg/config"
	"github.com/peterpla/lead-expert/pkg/database"
)

func TestCompletionProcessing(t *testing.T) {
	t.Skip()

	cfg := config.GetConfigPointer()
	// servicePrefix := "completion-processing-dot-" // <---- change to match service!!
	port := cfg.TaskCompletionProcessingPort // <---- change to match service!!
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
		1234567, "gs://elated-practice-224603.appspot.com/audio_uploads/audio-01.mp3", time.Now().UTC().Format(time.RFC3339Nano))

	tests := []test{
		// valid
		{name: "valid POST /task_handler",
			method:   "POST",
			endpoint: "/task_handler",
			body:     jsonBody,
			respBody: "accepted_at",
			status:   http.StatusInternalServerError}, // TODO: inject UUID to avoid ErrZeroUUIDError?
		{name: "valid GET /",
			method:   "GET",
			endpoint: "/",
			respBody: "service running",
			status:   http.StatusOK},
		{name: "invalid GET /nope",
			method:   "GET",
			endpoint: "/nope",
			respBody: "not found",
			status:   http.StatusNotFound}}

	prefix := fmt.Sprintf("http://localhost:%s", port)
	// prefix := fmt.Sprintf("https://%s%s.appspot.com", servicePrefix, cfg.ProjectID)

	for _, tc := range tests {
		url := prefix + tc.endpoint
		// log.Printf("Test %s: %s", tc.name, url)

		router := httprouter.New()
		router.POST("/task_handler", taskHandler())
		router.GET("/", indexHandler)

		// build the POST request with custom header
		theRequest, err := http.NewRequest(tc.method, url, strings.NewReader(tc.body))
		if err != nil {
			t.Fatal(err)
		}

		// running locally, add headers as App Engine does, since we check for them elsewhere
		if strings.HasPrefix(prefix, "http://localhost") {
			theRequest.Header.Set("X-Appengine-Taskname", "localTask")
			theRequest.Header.Set("X-Appengine-Queuename", "localQueue")
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
