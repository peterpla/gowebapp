package fstore

import (
	"github.com/google/uuid"

	"github.com/peterpla/lead-expert/pkg/request"
)

type requestRepository struct {
	ProjectID  string
	Collection string
}

func NewFirestoreRequestRepository(projID string, coll string) request.RequestRepository {
	return &requestRepository{
		projID,
		coll,
	}
}

func (r *requestRepository) Create(request *request.Request) error {
	// See Exercise as example: https://github.com/peterpla/exercise/blob/master/backend/
	var userMap = make(map[string]interface{})

	// TODO: unmarshal Request into userMap
	// TODO: Create in database
	// TODO: update Request from marshalling returned Result

	return nil
}

func (r *requestRepository) FindByID(reqID uuid.UUID) (*request.Request, error) {
	// See Exercise as example: https://github.com/peterpla/exercise/blob/master/backend/
	var emptyRequest = request.Request{}
	var userMap = make(map[string]interface{})

	// TODO: set key fields of userMap from Request
	// TODO: query the database
	// TODO: handle Not Found and other errors
	// TODO: populate Request from returned data

	return &emptyRequest, nil
}
