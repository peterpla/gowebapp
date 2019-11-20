package adding

// Request defines properties of an incoming transcription request
// to be added
type Request struct {
	RequestID    int    `json:"request_id"`
	CustomerID   string `json:"customer_id"`
	MediaFileURL string `json:"media_url"`
	CustomConfig bool   `json:"custom_config"`
}
