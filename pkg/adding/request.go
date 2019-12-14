package adding

import (
	"time"

	"github.com/google/uuid"
)

// Request defines properties of an incoming transcription request
// to be added
type Request struct {
	RequestID       uuid.UUID    `json:"request_id"`
	CustomerID      string       `json:"customer_id"`
	MediaFileURI    string       `json:"media_uri"`
	CustomConfig    bool         `json:"custom_config"`
	AcceptedAt      time.Time    `json:"accepted_at"`
	RawTranscript   []RawResults `json:"raw_transcript"`
	FinalTranscript string       `json:"final_transcript"`
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
	CustomerID   string    `json:"customer_id"`
	MediaFileURI string    `json:"media_uri"`
	AcceptedAt   string    `json:"accepted_at"`
}

type CompletionResponse struct {
	RequestID       uuid.UUID `json:"request_id"`
	CustomerID      string    `json:"customer_id"`
	MediaFileURI    string    `json:"media_uri"`
	AcceptedAt      string    `json:"accepted_at"`
	FinalTranscript string    `json:"final_transcript"`
	CompletedAt     string    `json:"completed_at"`
}
