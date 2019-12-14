package adding

import (
	"time"

	"github.com/google/uuid"
)

// Request defines properties of an incoming transcription request
// to be added
type Request struct {
	RequestID    uuid.UUID `json:"request_id"`
	CustomerID   string    `json:"customer_id"`
	MediaFileURI string    `json:"media_uri"`
	CustomConfig bool      `json:"custom_config"`
	AcceptedAt   time.Time `json:"accepted_at"`
}

type ReqResponse struct {
	RequestID    uuid.UUID `json:"request_id"`
	CustomerID   string    `json:"customer_id"`
	MediaFileURI string    `json:"media_uri"`
	AcceptedAt   string    `json:"accepted_at"`
}
