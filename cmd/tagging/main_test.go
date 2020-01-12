package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-playground/validator"
	"github.com/julienschmidt/httprouter"

	"github.com/peterpla/lead-expert/pkg/config"
	"github.com/peterpla/lead-expert/pkg/database"
	"github.com/peterpla/lead-expert/pkg/queue"
)

func TestTagging(t *testing.T) {

	cfg := config.GetConfigPointer()
	// servicePrefix := "tagging-dot-" // <---- change to match service!!
	port := cfg.TaskTaggingPort // <---- change to match service!!
	repo = database.NewFirestoreRequestRepository(cfg.ProjectID, cfg.DatabaseRequests)

	validate = validator.New()

	type test struct {
		name     string
		endpoint string
		body     string
		respBody string
		status   int
	}

	var jsonBody []byte
	var err error
	var testFile = "./../../data/RE7a23da60565501cf1d88f9984b1c6399_transcriptQAComplete.json"

	if jsonBody, err = ioutil.ReadFile(testFile); err != nil {
		msg := fmt.Sprintf("TestTagging cannot read the json file %q, err: %v", testFile, err)
		panic(msg)
	}
	// log.Printf("TestTagging, jsonBody: %q\n", string(jsonBody))

	tests := []test{
		// valid
		{name: "valid POST /task_handler",
			endpoint: "/task_handler",
			body:     string(jsonBody),
			status:   http.StatusOK},
	}

	qi = queue.QueueInfo{}
	q = queue.NewNullQueue(&qi) // use null queue, requests thrown away on exit
	// q = queue.NewGCTQueue(&qi) // use Google Cloud Tasks
	qs = queue.NewService(q)

	prefix := fmt.Sprintf("http://localhost:%s", port)
	// prefix = fmt.Sprintf("https://%s%s.appspot.com", servicePrefix, os.Getenv("PROJECT_ID"))

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
			t.Errorf("%s: expected blank body, got %q", tc.name, string(b))
		}
	}
}
