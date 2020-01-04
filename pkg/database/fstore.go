package database

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"

	"github.com/peterpla/lead-expert/pkg/request"
	"github.com/peterpla/lead-expert/pkg/serviceInfo"
)

type requestRepository struct {
	ProjectID  string
	Collection string
}

var ErrCreateError = fmt.Errorf("fstore Create error")

func NewFirestoreRequestRepository(projID string, coll string) request.RequestRepository {
	sn := serviceInfo.GetNextServiceName()
	log.Printf("%s.fstore.NewFirestoreRequestRepository, projID: %q, coll: %q\n",
		sn, projID, coll)

	return requestRepository{
		projID,
		coll,
	}
}

// Create writes the Request to the database
func (r requestRepository) Create(request *request.Request) error {
	sn := serviceInfo.GetNextServiceName()
	// See Exercise as example: https://github.com/peterpla/exercise/blob/master/backend/

	// TODO: lock the request while it's being written?

	// get a map corresponding to the Request
	userMap, err := request.ToMap()
	if err != nil {
		log.Printf("%s.fstore.Create, ToMap err: %v\n", sn, err)
		return err
	}

	// TODO: normalize Request fields

	// prepare to talk to Firestore
	log.Printf("%s.fstore.Create, repo.ProjectID: %s, repo.Collection: %s\n", sn, r.ProjectID, r.Collection)
	ctx := context.Background()
	projID := r.ProjectID
	client, err := firestore.NewClient(ctx, projID)
	if err != nil {
		log.Printf("%s.fstore.Create, NewClient returned err: %v\n", sn, err)
		return ErrCreateError
	}
	defer client.Close()

	col := r.Collection
	log.Printf("%s.fstore.Create, client: %+v, colRef: %+v\n", sn, client, col)

	// create the new User document using Set, with document ID set to CustomerID
	uid := strconv.Itoa(request.CustomerID)
	log.Printf("%s.fstore.Create, calling Set() with userMap: %+v\n", sn, userMap)
	_, err = client.Collection(col).Doc(uid).Set(ctx, userMap)
	if err != nil {
		log.Printf("%s.fstore.Create, Set returned err %+v\n", sn, err)
		return ErrCreateError
	}
	docRef := client.Collection(col).Doc(uid)

	// read the user back from the database, and return it
	docsnap, err := docRef.Get(ctx)
	if err != nil {
		log.Printf("%s.fstore.Create, Get returned err: %+v\n", sn, err)
		return ErrCreateError
	}

	createdUser := make(map[string]interface{})
	if err := docsnap.DataTo(&createdUser); err != nil {
		log.Printf("r.createUser, DataTo returned err: %+v", err)
		return ErrCreateError
	}
	log.Printf("%s.fstore.Create, createdUser: +%v\n", sn, createdUser)

	return nil
}

func (r requestRepository) FindByID(reqID uuid.UUID) (*request.Request, error) {
	// See Exercise as example: https://github.com/peterpla/exercise/blob/master/backend/
	var emptyRequest = request.Request{}

	// TODO: set key fields of userMap from Request
	// TODO: query the database
	// TODO: handle Not Found and other errors
	// TODO: populate Request from returned data

	return &emptyRequest, nil
}
