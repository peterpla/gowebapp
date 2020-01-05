package request

import (
	"log"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

func TestToMap(t *testing.T) {
	earlier := time.Now().Add(time.Microsecond * -3)

	tmpUUID := uuid.New()
	request := Request{
		RequestID:    tmpUUID,
		CustomerID:   1234567,
		MediaFileURI: "http://www.dropbox.com/test.mp3",
		Status:       Pending,
		AcceptedAt:   earlier.Format(time.RFC3339Nano),
	}
	if _, err := request.AddTimestamps("BeginTest", earlier.Format(time.RFC3339Nano), "EndTest"); err != nil {
		t.Errorf("AddTimestamps error: %v", err)
	}

	expected := make(map[string]interface{})
	expected["request_id"] = tmpUUID.String()
	expected["customer_id"] = float64(1234567)
	expected["media_uri"] = request.MediaFileURI
	expected["status"] = request.Status
	expected["original_status"] = float64(request.OriginalStatus)
	expected["accepted_at"] = request.AcceptedAt
	expected["completed_at"] = request.CompletedAt
	expected["working_transcript"] = request.WorkingTranscript
	expected["final_transcript"] = request.FinalTranscript

	timestamps := make(map[string]interface{})
	timestamps["BeginTest"] = request.Timestamps["BeginTest"]
	timestamps["EndTest"] = request.Timestamps["EndTest"]
	expected["timestamps"] = timestamps

	result, err := request.ToMap()
	if err != nil {
		t.Errorf("ToMap error: %v", err)
	}

	if !cmp.Equal(expected, result) {
		findMismatch(t, expected, result)

		t.Errorf("ToMap mismatch, expected %v, got %v", expected, result)
	}
}

func findMismatch(t *testing.T, expected map[string]interface{}, got map[string]interface{}) {

	var foundMismatch = false

	if present(t, got, "request_id", "got") && present(t, expected, "request_id", "expected") {
		if expected["request_id"].(string) != got["request_id"].(string) {
			foundMismatch = true
			t.Errorf("RequestID: expected %q, got %q", expected["request_id"].(string), got["request_id"].(string))
		}
	}

	if present(t, got, "customer_id", "got") && present(t, expected, "customer_id", "expected") {
		eCustID := expected["customer_id"].(float64)
		gCustID := got["customer_id"].(float64)
		if eCustID != gCustID {
			foundMismatch = true
			t.Errorf("CustomerID: expected %d, got %d", int(eCustID), int(gCustID))
		}
	}

	if present(t, got, "media_uri", "got") && present(t, expected, "media_uri", "expected") {
		if expected["media_uri"].(string) != got["media_uri"].(string) {
			foundMismatch = true
			t.Errorf("MediaFileURI: expected %q, got %q", expected["media_uri"].(string), got["media_uri"].(string))
		}
	}

	if present(t, got, "status", "got") && present(t, expected, "status", "expected") {
		if expected["status"].(string) != got["status"].(string) {
			foundMismatch = true
			t.Errorf("Status: expected %q, got %q", expected["status"].(string), got["status"].(string))
		}
	}

	if present(t, got, "original_status", "got") && present(t, expected, "original_status", "expected") {
		eOrigStatus := expected["original_status"].(float64)
		gOrigStatus := got["original_status"].(float64)
		if eOrigStatus != gOrigStatus {
			foundMismatch = true
			t.Errorf("OriginalStatus: expected %f got %f", eOrigStatus, gOrigStatus)
		}
	}

	if expected["accepted_at"].(string) != got["accepted_at"].(string) {
		foundMismatch = true
		t.Errorf("AcceptedAt: expected %q, got %q", expected["accepted_at"].(string), got["accepted_at"].(string))
	}
	if expected["completed_at"].(string) != got["completed_at"].(string) {
		foundMismatch = true
		t.Errorf("CompletedAt: expected %q, got %q", expected["completed_at"].(string), got["completed_at"].(string))
	}
	if expected["working_transcript"].(string) != got["working_transcript"].(string) {
		foundMismatch = true
		t.Errorf("WorkingTranscript: expected %q, got %q", expected["working_transcript"].(string), got["working_transcript"].(string))
	}
	if expected["final_transcript"].(string) != got["final_transcript"].(string) {
		foundMismatch = true
		t.Errorf("FinalTranscript: expected %q, got %q", expected["final_transcript"].(string), got["final_transcript"].(string))
	}

	eBegin, eEnd := getBeginEndTimestamps(t, expected)
	gBegin, gEnd := getBeginEndTimestamps(t, got)
	if eBegin != gBegin {
		foundMismatch = true
		t.Errorf("timestamps[\"BeginTest\"]: expected %q, got %q", eBegin, gBegin)
	}
	if eEnd != gEnd {
		foundMismatch = true
		t.Errorf("timestamps[\"EndTest\"]: expected %q, got %q", eEnd, gEnd)
	}

	if !foundMismatch {
		log.Println("findMismatch: mismatch NOT found")
	}
}

func present(t *testing.T, theMap map[string]interface{}, key string, label string) bool {
	_, ok := theMap[key]
	if !ok {
		t.Errorf("%s missing %q", label, key)
	}
	return ok
}

func getBeginEndTimestamps(t *testing.T, theMap map[string]interface{}) (begin string, end string) {
	var temp map[string]interface{}
	var ok bool

	temp, ok = theMap["timestamps"].(map[string]interface{})
	if !ok {
		t.Errorf("\"Timestamps\" not present")
	}

	begin, ok = temp["BeginTest"].(string)
	if !ok {
		t.Errorf("\"BeginTest\" not present")
	}

	end, ok = temp["EndTest"].(string)
	if !ok {
		t.Errorf("\"EndTest\" not present")
	}

	return begin, end
}
