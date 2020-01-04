package request

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

func TestToMap(t *testing.T) {
	earlier := time.Now().Add(time.Microsecond * -3)

	request := Request{
		RequestID:    uuid.New(),
		CustomerID:   1234567,
		MediaFileURI: "http://www.dropbox.com/test.mp3",
		Status:       Pending,
		AcceptedAt:   earlier.Format(time.RFC3339Nano),
	}
	if _, err := request.AddTimestamps("BeginTest", earlier.Format(time.RFC3339Nano), "EndTest"); err != nil {
		t.Errorf("AddTimestamps error: %v", err)
	}

	expected := make(map[string]interface{})
	expected["RequestID"] = request.RequestID
	expected["CustomerID"] = request.CustomerID
	expected["MediaFileURI"] = request.MediaFileURI
	expected["Status"] = request.Status
	expected["AcceptedAt"] = request.AcceptedAt
	expected["Timestamps:[BeginTest]"] = request.Timestamps["BeginTest"]
	expected["Timestamps:[EndTest]"] = request.Timestamps["EndTest"]

	result, err := request.ToMap()
	if err != nil {
		t.Errorf("ToMap error: %v", err)
	}

	if !cmp.Equal(expected, result) {
		t.Errorf("ToMap mismatch, expected %v, got %v", expected, result)
	}

}
