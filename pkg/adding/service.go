package adding

import (
	"errors"
)

// ErrCustomerNotFound - provided customer ID was not found
var ErrCustomerNotFound = errors.New("customer ID not found")

// Service provides adding operations
type Service interface {
	AddRequest(Request)
}

// Repository - access Request repository
type Repository interface {
	AddRequest(Request) error
}

type service struct {
	bR Repository
}

// New Service - creates an adding service with its dependencies
func NewService(r Repository) Service {
	return &service{r}
}

// AddRequest adds the request to the database
func (s *service) AddRequest(req Request) {
	// log.Printf("adding.AddRequest - enter\n")
	// TODO: validation

	// TODO: error handling
	_ = s.bR.AddRequest(req)
	// log.Printf("adding.AddRequest - exit\n")
}
