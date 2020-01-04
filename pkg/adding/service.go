package adding

import (
	"errors"
)

// ErrCustomerNotFound - provided customer ID was not found
var ErrCustomerNotFound = errors.New("customer ID not found")

// Service provides adding operations
// type Service interface {
// 	// AddRequest() takes a Request and returns a (possibly modified)
// 	// Request
// 	AddRequest(Request) Request
// }

// Repository provides persistent adding services
// type Repository interface {
// 	// AddRequest() takes a request and returns an error with
// 	// the status of the Add operation
// 	AddRequest(Request) error
// }

// type service struct {
// 	bR Repository
// }

// New Service - creates an adding service with its dependencies
// func NewService(r Repository) Service {
// 	panic("adding.NewService - should not be called\n")
// 	// return &service{r}
// }

// AddRequest adds the request to be processed
// func (s *service) AddRequest(req Request) Request {
// 	sn := serviceInfo.GetServiceName()
// 	msg := fmt.Sprintf("%s.adding.NewService - should not be called\n", sn)
// 	panic(msg)

// log.Printf("%s.adding.AddRequest enter, req: %+v\n", sn, req)

// newReq := req
// TODO: validate req object

// TODO: error handling
// if err := s.bR.AddRequest(req); err != nil {
// 	log.Printf("%s.adding.AddRequest, bR.AddRequest error: %+v, req: +%v\n", sn, err, req)
// }

// // log.Printf("%s.adding.AddRequest exiting, newReq: %+v\n", sn, newReq)
// return newReq
// }
