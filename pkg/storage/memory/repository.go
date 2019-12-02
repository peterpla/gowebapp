package memory

import (
	"log"

	"github.com/peterpla/gowebapp/pkg/adding"
)

// Memory storage keeps data in memory
type Storage struct {
	// In-memory representation of incoming request queue
	requests []adding.Request
}

// Add saves the request to the repository
func (m *Storage) AddRequest(req adding.Request) error {
	// log.Printf("memory.AddRequest - enter\n")

	newRequest := adding.Request{
		RequestID:    len(m.requests) + 1,
		CustomerID:   req.CustomerID,
		MediaFileURL: req.MediaFileURL,
		CustomConfig: false,
	}

	// TODO: pick up custom configuration from request
	// (if CustomConfig == True) or defaults from customer profile

	m.requests = append(m.requests, newRequest)
	log.Printf("memory.AddRequest exiting, requests: %+v\n", m.requests)

	return nil
}
