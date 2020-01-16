package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-playground/validator"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"

	"github.com/peterpla/lead-expert/pkg/config"
	"github.com/peterpla/lead-expert/pkg/database"
	"github.com/peterpla/lead-expert/pkg/queue"
	"github.com/peterpla/lead-expert/pkg/request"
)

func TestTaggingQAPost(t *testing.T) {

	cfg := config.GetConfigPointer()
	// servicePrefix := "tagging-qa-dot-" // <---- change to match service!!
	port := cfg.TaskTaggingQAPort // <---- change to match service!!
	repo = database.NewFirestoreRequestRepository(cfg.ProjectID, cfg.DatabaseRequests)

	validate = validator.New()

	type test struct {
		name     string
		endpoint string
		body     string
		// respTags map[string]request.Tags
		status int
	}

	var customerID = 1234567
	var media = "gs://elated-practice-224603.appspot.com/audio_uploads/audio-01.mp3"
	var timeNow = time.Now().UTC().Format(time.RFC3339Nano)

	jsonSingleTag := createSingleTag(t, customerID, media, timeNow)

	tests := []test{
		// valid
		{name: "single Tag",
			endpoint: "/task_handler",
			body:     string(jsonSingleTag),
			// respTags: nil,
			status: http.StatusOK},
	}

	qi = queue.QueueInfo{}
	q = queue.NewNullQueue(&qi) // use null queue, requests thrown away on exit
	// q = queue.NewGCTQueue(&qi) // use Google Cloud Tasks
	qs = queue.NewService(q)

	prefix := fmt.Sprintf("http://localhost:%s", port)
	// prefix := fmt.Sprintf("https://%s%s.appspot.com", servicePrefix, cfg.ProjectID)

	for _, tc := range tests {
		url := prefix + tc.endpoint
		// log.Printf("Test %s: %s", tc.name, url)

		router := httprouter.New()
		router.POST("/task_handler", taskHandler(q))

		// build the POST request with custom header
		theRequest, err := http.NewRequest("POST", url, strings.NewReader(tc.body))
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
	}
}

// createSingleTag returns the JSON of a request with a single Tag
func createSingleTag(t *testing.T, customerID int, media string, timeNow string) []byte {
	var jsonString []byte
	var err error

	var req = request.Request{}
	req.RequestID = uuid.New()
	req.CustomerID = customerID
	req.MediaFileURI = media
	req.AcceptedAt = timeNow

	var m = make(map[string]request.Tags)
	var k = "123 Main Street"
	var it = "ADDRESS"
	m[k] = request.Tags{InfoType: it, Likelihood: 4, BeginByteOffset: 0, EndByteOffset: len(k)}
	req.MatchedTags = m

	if jsonString, err = json.Marshal(req); err != nil {
		t.Fatalf("createSingleTag, json.Marshal err: %v\n", err)
	}

	return jsonString
}

// ********** ********** ********** ********** ********** **********

func TestTaggingQAGet(t *testing.T) {

	cfg := config.GetConfigPointer()
	// servicePrefix := "tagging-qa-dot-" // <---- change to match service!!
	port := cfg.TaskTaggingQAPort // <---- change to match service!!
	repo = database.NewFirestoreRequestRepository(cfg.ProjectID, cfg.DatabaseRequests)

	validate = validator.New()

	type test struct {
		name     string
		method   string
		endpoint string
		respBody string
		status   int
	}

	// var jsonBody []byte
	// var err error

	tests := []test{
		{name: "valid GET /",
			method:   "GET",
			endpoint: "/",
			respBody: "service running",
			status:   http.StatusOK},
		{name: "invalid GET /nope",
			method:   "GET",
			endpoint: "/nope",
			respBody: "Not Found",
			status:   http.StatusNotFound},
	}

	qi = queue.QueueInfo{}
	q = queue.NewNullQueue(&qi) // use null queue, requests thrown away on exit
	// q = queue.NewGCTQueue(&qi) // use Google Cloud Tasks
	qs = queue.NewService(q)

	prefix := fmt.Sprintf("http://localhost:%s", port)
	// prefix = fmt.Sprintf("https://%s%s.appspot.com", servicePrefix, cfg.ProjectID)

	for _, tc := range tests {
		url := prefix + tc.endpoint
		// log.Printf("Test %s: %s", tc.name, url)

		router := httprouter.New()
		router.GET("/", indexHandler)
		router.NotFound = http.HandlerFunc(myNotFound)

		// build the POST request with custom header
		theRequest, err := http.NewRequest(tc.method, url, nil)
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
				t.Errorf("%s: expected %q, not found in %q", tc.name, tc.respBody, string(b))
			}
		}
	}
}
