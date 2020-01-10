package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/go-playground/validator"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"

	"github.com/peterpla/lead-expert/pkg/config"
	"github.com/peterpla/lead-expert/pkg/database"
	"github.com/peterpla/lead-expert/pkg/queue"
	"github.com/peterpla/lead-expert/pkg/request"
)

// var validate *validator.Validate
var createdUUID = uuid.UUID{}

func TestDefaultPost(t *testing.T) {

	cfg := config.GetConfigPointer()
	servicePrefix := ""
	port := cfg.TaskDefaultPort
	repo = database.NewFirestoreRequestRepository(cfg.ProjectID, cfg.DatabaseRequests)

	validate = validator.New()

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

	qi = queue.QueueInfo{}
	q = queue.NewNullQueue(&qi) // use null queue, requests thrown away on exit
	// q = queue.NewGCTQueue(&qi) // use Google Cloud Tasks
	qs = queue.NewService(q)

	apiPrefix := "/api/v1"

	prefix := fmt.Sprintf("http://localhost:%s%s", port, apiPrefix)
	if cfg.IsGAE {
		prefix = fmt.Sprintf("https://%s%s.appspot.com%s", servicePrefix, os.Getenv("PROJECT_ID"), apiPrefix)
	}

	for _, tc := range tests {
		url := prefix + tc.endpoint
		// log.Printf("Test %s: %s", tc.name, url)

		router := httprouter.New()
		router.POST("/api/v1/requests", postHandler(q))

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

		var b []byte

		if tc.respBody != "" {
			if b, err = ioutil.ReadAll(rr.Body); err != nil {
				t.Fatalf("%s: ReadAll error: %v", tc.name, err)
			}

			if !strings.Contains(string(b), tc.respBody) {
				t.Errorf("%s: expected %q, not found (in %q)", tc.name, tc.respBody, string(b))
			}
		}

		if tc.name == "valid POST requests" {
			// save the created RequestID
			var response request.PostResponse
			if err := json.Unmarshal(b, &response); err != nil {
				t.Errorf("%s: json.Unmarshal error: %+v", tc.name, err)
			}
			createdUUID = response.RequestID
			// log.Printf("createdUUID: %q\n", createdUUID)
		}
	}
}

func TestDefaultGetStatus(t *testing.T) {

	// skip this test if createdUUID was not set by TestDefaultPost (has zero value)
	var zeroUUID = uuid.UUID{}
	if createdUUID == zeroUUID {
		t.Skip("skipping as createdUUID not set.")
	}

	cfg := config.GetConfigPointer()
	servicePrefix := ""
	port := cfg.TaskDefaultPort

	validate = validator.New()

	apiPrefix := "/api/v1"

	req := request.Request{
		CustomerID:   1234567,
		MediaFileURI: "gs://elated-practice-224603.appspot.com/audio_uploads/audio-01.mp3",
	}

	prefix := fmt.Sprintf("https://%s%s.appspot.com%s", servicePrefix, os.Getenv("PROJECT_ID"), apiPrefix)
	if !cfg.IsGAE {
		// prefix := fmt.Sprintf("http://localhost:%s%s", port, apiPrefix)
		prefix = apiPrefix
		_ = port // suppress not-used warning
	}

	router := httprouter.New()
	router.GET(prefix+"/status/:uuid", getStatusHandler())

	// build the GET request with custom header
	url := prefix + "/status/" + createdUUID.String()

	body, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("json.Marshal error: %+v", err)
	}

	// log.Printf("GET /status/%q, body: %q", url, string(body))
	theRequest, err := http.NewRequest("GET", url, bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}

	// response recorder
	rr := httptest.NewRecorder()

	// send the request
	router.ServeHTTP(rr, theRequest)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code %v, got %v", http.StatusOK, rr.Code)
	}

	if rr.Body != nil {
		var b []byte
		if b, err = ioutil.ReadAll(rr.Body); err != nil {
			t.Fatalf("ReadAll error: %v", err)
		}
		// log.Printf("response body: %q\n", string(b))

		var response request.GetStatusResponse
		if err := json.Unmarshal(b, &response); err != nil {
			t.Fatalf("json.Unmarshal error: %+v", err)
		}

		if response.OriginalRequestID != createdUUID {
			t.Errorf("RequestID mismatch, expected %q, got %q", createdUUID, response.OriginalRequestID)
		}
	}
}

func TestDefaultGetStatusSpecial(t *testing.T) {

	cfg := config.GetConfigPointer()
	servicePrefix := ""
	port := cfg.TaskDefaultPort

	validate = validator.New()

	type test struct {
		name     string
		endpoint string
		uuid     string
		body     string
		respBody string
		status   int
	}

	tests := []test{
		// valid
		{name: "GET status COMPLETED",
			endpoint: "/status/",
			uuid:     request.CompletedUUIDStr,
			body:     `{ "customer_id": 1234567, "media_uri": "gs://elated-practice-224603.appspot.com/audio_uploads/audio-01.mp3" }`,
			respBody: "endpoint",
			status:   http.StatusOK,
		},
		{name: "GET status PENDING",
			endpoint: "/status/",
			uuid:     request.PendingUUIDStr,
			body:     `{ "customer_id": 1234567, "media_uri": "gs://elated-practice-224603.appspot.com/audio_uploads/audio-01.mp3" }`,
			respBody: "eta",
			status:   http.StatusOK,
		},
		{name: "GET status ERROR",
			endpoint: "/status/",
			uuid:     request.ErrorUUIDStr,
			body:     `{ "customer_id": 1234567, "media_uri": "gs://elated-practice-224603.appspot.com/audio_uploads/audio-01.mp3" }`,
			respBody: "original_status",
			status:   http.StatusOK,
		},
	}

	apiPrefix := "/api/v1"

	prefix := fmt.Sprintf("https://%s%s.appspot.com%s", servicePrefix, os.Getenv("PROJECT_ID"), apiPrefix)
	if !cfg.IsGAE {
		// prefix := fmt.Sprintf("http://localhost:%s%s", port, apiPrefix)
		prefix = apiPrefix
		_ = port // suppress not-used warning
	}

	for _, tc := range tests {

		router := httprouter.New()
		router.GET(prefix+tc.endpoint+":uuid", getStatusHandler())

		tempUUID := tc.uuid
		if tc.uuid == "generate" {
			tempUUID = uuid.New().String()
		}
		// build the GET request with custom header
		url := prefix + tc.endpoint + tempUUID
		// log.Printf("Test %s: %s", tc.name, url)

		theRequest, err := http.NewRequest("GET", url, strings.NewReader(tc.body))
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
			// log.Printf("%s: response body: %q\n", tc.name, string(b))

			if !strings.Contains(string(b), tc.respBody) {
				t.Errorf("%s: expected %q, not found (in %q)", tc.name, tc.respBody, string(b))
			}
		}
	}
}

func TestDefaultGetTranscripts(t *testing.T) {

	cfg := config.GetConfigPointer()
	servicePrefix := ""
	port := cfg.TaskDefaultPort

	validate = validator.New()

	type test struct {
		name     string
		endpoint string
		body     string
		respBody string
		status   int
	}

	tests := []test{
		// valid
		{name: "valid GET transcripts",
			endpoint: "/transcripts/",
			body:     `{ "customer_id": 1234567, "media_uri": "gs://elated-practice-224603.appspot.com/audio_uploads/audio-01.mp3" }`,
			respBody: "transcript",
			status:   http.StatusOK,
		},
		// TODO: test for "status_for_req" = "ERROR", "PENDING", "COMPLETE"
	}

	apiPrefix := "/api/v1"

	prefix := fmt.Sprintf("https://%s%s.appspot.com%s", servicePrefix, os.Getenv("PROJECT_ID"), apiPrefix)
	if !cfg.IsGAE {
		// prefix := fmt.Sprintf("http://localhost:%s%s", port, apiPrefix)
		prefix = apiPrefix
		_ = port // suppress not-used warning
	}

	for _, tc := range tests {

		router := httprouter.New()
		router.GET(prefix+tc.endpoint+":uuid", getTranscriptsHandler())

		// build the GET request with custom header
		url := prefix + tc.endpoint + uuid.New().String() // TODO: use a non-random UUID
		// log.Printf("Test %s: %s", tc.name, url)

		theRequest, err := http.NewRequest("GET", url, strings.NewReader(tc.body))
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
			// log.Printf("%s: response body: %q\n", tc.name, string(b))

			if !strings.Contains(string(b), tc.respBody) {
				t.Errorf("%s: expected %q, not found (in %q)", tc.name, tc.respBody, string(b))
			}
		}
	}
}
