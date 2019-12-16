package adding

import (
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
)

// ErrTimestampsKeyExists - key provided already exists
var ErrTimestampsKeyExists = errors.New("Timestamps key exists")

// ErrInvalidTime - time provided does not parse
var ErrInvalidTime = errors.New("Invalid time value cannot be parsed")

// Request defines properties of an incoming transcription request
// to be added
type Request struct {
	RequestID       uuid.UUID         `json:"request_id"`
	CustomerID      int               `json:"customer_id" validate:"required,gte=1,lt=10000000"`
	MediaFileURI    string            `json:"media_uri" validate:"required,uri"`
	AcceptedAt      string            `json:"accepted_at"`
	CompletedAt     string            `json:"completed_at"`
	RawTranscript   []RawResults      `json:"raw_transcript"`
	FinalTranscript string            `json:"final_transcript"`
	Timestamps      map[string]string `json:"timestamps"`
}

// RawResults holds the raw results from ML transcription
type RawResults struct {
	Transcript string
	Confidence float32
}

// PostResponse holds seelcted fields of Result struct to include in
// HTTP response to initial POST request
type PostResponse struct {
	RequestID    uuid.UUID `json:"request_id"`
	CustomerID   int       `json:"customer_id"`
	MediaFileURI string    `json:"media_uri"`
	AcceptedAt   string    `json:"accepted_at"`
}

type CompletionResponse struct {
	RequestID       uuid.UUID `json:"request_id"`
	CustomerID      int       `json:"customer_id" validate:"required,gte=1,lt=10000000"`
	MediaFileURI    string    `json:"media_uri"`
	AcceptedAt      string    `json:"accepted_at"`
	FinalTranscript string    `json:"final_transcript"`
	CompletedAt     string    `json:"completed_at"`
}

func (req *Request) AddTimestamps(startKey, startTimestamp, endKey string) (time.Duration, error) {

	var badTime time.Duration
	var startTime time.Time
	var err error

	if startTime, err = time.Parse(time.RFC3339Nano, startTimestamp); err != nil {
		log.Printf("adding.AddTimestamps ERROR: startTime %s does not parse (RFC3339Nano)\n", startTimestamp)
		return badTime, ErrTimestampsKeyExists
	}

	// initialize map if needed
	if req.Timestamps == nil {
		// log.Printf("adding.AddTimestamps initializing Timestamps map in Request\n")
		req.Timestamps = make(map[string]string)
	}

	// if startKey already exists, return error
	startKeyValue, ok := req.Timestamps[startKey]
	if ok {
		log.Printf("adding.AddTimestamps ERROR: key %s exists with value %s\n", startKey, startKeyValue)
		return badTime, ErrTimestampsKeyExists
	}

	// if endKey already exists, return error
	endKeyValue, ok := req.Timestamps[endKey]
	if ok {
		log.Printf("adding.AddTimestamp ERROR: key %s exists with value %s\n", endKey, endKeyValue)
		return badTime, ErrTimestampsKeyExists
	}

	// set startKey to startTimestamp
	req.Timestamps[startKey] = startTimestamp

	// set endKey to current time
	now := time.Now().UTC()
	req.Timestamps[endKey] = now.Format(time.RFC3339Nano)
	duration := now.Sub(startTime)

	return duration, nil
}

func (req *Request) RequestDuration() time.Duration {
	var accepted, completed time.Time

	accepted, _ = time.Parse(time.RFC3339Nano, req.AcceptedAt)
	completed, _ = time.Parse(time.RFC3339Nano, req.CompletedAt)

	return completed.Sub(accepted)

}
