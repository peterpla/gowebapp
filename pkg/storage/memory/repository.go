package memory

import (
	"github.com/peterpla/gowebapp/pkg/adding"
)

// Memory storage keeps data in memory
type Storage struct {
	requests []adding.Request
}

// Add saves the request to the repository
func (m *Storage) AddRequest(req adding.Request) error {
	// log.Printf("memory.AddRequest - enter\n")

	// TODO: check for duplicate?

	newRequest := adding.Request{
		RequestID:    len(m.requests) + 1,
		CustomerID:   req.CustomerID,
		MediaFileURL: req.MediaFileURL,
		CustomConfig: false,
	}

	// TODO: pick up custom configuration from request
	// (if CustomConfig == True) or defaults from customer profile

	// TODO: create different struct(s) for incoming requests,
	// stored requests, others?

	m.requests = append(m.requests, newRequest)
	// log.Printf("memory.AddRequest - exit, requests: %+v\n", m.requests)

	return nil
}
