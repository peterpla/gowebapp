package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/peterpla/lead-expert/pkg/request"
	"github.com/peterpla/lead-expert/pkg/serviceInfo"
)

// requestRepository implements the request.RequestRepository interface
type requestRepository struct {
	ProjectID  string
	Collection string
}

func NewFirestoreRequestRepository(projID string, coll string) request.RequestRepository {
	sn := serviceInfo.GetNextServiceName()
	// log.Printf("%s.fstore.NewFirestoreRequestRepository, projID: %q, coll: %q\n",
	// 	sn, projID, coll)

	repo := requestRepository{
		projID,
		coll,
	}
	// log.Printf("%s.fstore.NewFirestoreRequestRepository, returning repo: %+v\n", sn, repo)
	_ = sn

	return repo
}

// Create writes the Request to the database
func (r requestRepository) Create(request *request.Request) error {
	sn := serviceInfo.GetNextServiceName()

	// TODO: lock the request while it's being written?

	// get a map corresponding to the Request
	// reqMap, err := request.ToMap()
	// if err != nil {
	// 	log.Printf("%s.fstore.Create, ToMap err: %v\n", sn, err)
	// 	return err
	// }

	// TODO: normalize Request fields

	// block creation of zero UUID requests
	zeroUUID := uuid.UUID{}
	if request.RequestID == zeroUUID {
		log.Printf("%s.fstore.Create, zero UUID not allowed\n", sn)
		return ErrZeroUUIDError
	}

	// prepare to talk to Firestore
	// log.Printf("%s.fstore.Create, repo.ProjectID: %s, repo.Collection: %s\n", sn, r.ProjectID, r.Collection)
	ctx := context.Background()
	projID := r.ProjectID
	client, err := firestore.NewClient(ctx, projID)
	if err != nil {
		log.Printf("%s.fstore.Create, NewClient returned err: %v\n", sn, err)
		return ErrCreateError
	}
	defer client.Close()

	col := r.Collection
	// log.Printf("%s.fstore.Create, client: %+v, colRef: %+v\n", sn, client, col)

	// request UUID = document ID, we'll search by the UUID later
	docID := request.RequestID.String()

	// log.Printf("%s.fstore.Create, calling Set() with reqMap: %+v\n", sn, reqMap)
	// _, err = client.Collection(col).Doc(docID).Set(ctx, reqMap)
	_, err = client.Collection(col).Doc(docID).Set(ctx, request)
	if err != nil {
		log.Printf("%s.fstore.Create, Set returned err %+v\n", sn, err)
		return ErrCreateError
	}
	docRef := client.Collection(col).Doc(docID)

	// read the Request back from the database, and return it
	docsnap, err := docRef.Get(ctx)
	if err != nil {
		log.Printf("%s.fstore.Create, Get returned err: %+v\n", sn, err)
		return ErrCreateError
	}

	createdRequest := make(map[string]interface{})
	if err := docsnap.DataTo(&createdRequest); err != nil {
		log.Printf("%s.fstore.Create, DataTo returned err: %+v", sn, err)
		return ErrCreateError
	}
	// log.Printf("%s.fstore.Create, createdRequest: +%v\n", sn, createdRequest)

	return nil
}

func (r requestRepository) FindByID(reqID uuid.UUID) (*request.Request, error) {
	sn := serviceInfo.GetNextServiceName()
	// See Exercise as example: https://github.com/peterpla/exercise/blob/master/backend/

	var emptyRequest = request.Request{}
	var foundRequest request.Request

	zeroUUID := uuid.UUID{}
	if reqID == zeroUUID {
		log.Printf("%s.fstore.FindByID, zero UUID not allowed\n", sn)
		return &emptyRequest, ErrZeroUUIDError
	}

	// prepare to talk to Firestore
	// log.Printf("%s.fstore.FindByID, repo.ProjectID: %s, repo.Collection: %s\n", repo.ProjectID, repo.Collection)
	ctx := context.Background()
	projID := r.ProjectID
	// log.Printf("%s.fstore.FindByID, projID: %s\n", projID)
	client, err := firestore.NewClient(ctx, projID)
	if err != nil {
		log.Printf("%s.fstore.FindByID, failed to create client: %v\n", sn, err)
		return &emptyRequest, ErrFindError
	}
	defer client.Close()
	// log.Printf("firestore client: %+v\n", client)

	col := client.Collection(r.Collection)
	// log.Printf("%s.fstore.FindByID, col: %+v\n", col)

	// search by UUID
	docID := reqID.String()
	docRef := col.Doc(docID)
	// log.Printf("%s.fstore.FindByID, docRef: %+v", docRef)

	// read the user back from the database
	docsnap, err := docRef.Get(ctx)
	if err != nil {
		st, _ := status.FromError(err)
		if st.Code() == codes.NotFound {
			// UUID not found
			log.Printf("%s.fstore.FindByID, doc with UUID=%q not found\n", sn, docID)
			return &emptyRequest, ErrNotFoundError
		}
		// some other error, return it
		log.Printf("%s.fstore.FindByID, Get returned err: %+v\n", sn, err)
		return &emptyRequest, ErrFindError
	}

	// extract data into the temporary map
	if err := docsnap.DataTo(&foundRequest); err != nil {
		log.Printf("%s.fstore.FindByID, DataTo returned err: %+v", sn, err)
		return &emptyRequest, ErrFindError
	}

	// save the UUID in RequestID as expected elsewhere
	foundRequest.RequestID = reqID

	// log.Printf("%s.fstore.FindByID, foundRequest: %+v\n", sn, foundRequest)

	return &foundRequest, nil
}

// Update writes an updated Request to the database
func (r requestRepository) Update(request *request.Request) error {
	sn := serviceInfo.GetNextServiceName()

	// TODO: lock the request while it's being written?

	// block update of zero UUID requests
	zeroUUID := uuid.UUID{}
	if request.RequestID == zeroUUID {
		log.Printf("%s.fstore.Update, zero UUID not allowed\n", sn)
		return ErrZeroUUIDError
	}

	// to use MergeAll we must provide a map, so get a map corresponding to the Request
	reqMap, err := request.ToMap()
	if err != nil {
		log.Printf("%s.fstore.Create, ToMap err: %v\n", sn, err)
		return err
	}

	// request.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	reqMap["updated_at"] = time.Now().UTC().Format(time.RFC3339Nano)
	// log.Printf("reqMap: %+v\n", reqMap)

	// prepare to talk to Firestore
	// log.Printf("%s.fstore.Update, repo.ProjectID: %s, repo.Collection: %s\n", sn, repo.ProjectID, repo.Collection)
	ctx := context.Background()
	projID := r.ProjectID
	client, err := firestore.NewClient(ctx, projID)
	if err != nil {
		log.Printf("%s.fstore.Update, failed to create client: %v\n", sn, err)
		return ErrUpdateError
	}
	defer client.Close()
	// log.Printf("firestore client: %+v\n", client)

	// request UUID = document ID, we'll search by the UUID later
	docID := request.RequestID.String()

	col := client.Collection(r.Collection)
	docRef := col.Doc(docID)
	// log.Printf("%s.fstore.Update, docRef: %+v\n", sn, docRef)

	// use "set with merge" (i.e., with MergeAll SetOption) - provided
	// fields overwrite corresponding fields in the existing document
	_, err = docRef.Set(ctx, reqMap, firestore.MergeAll)
	if err != nil {
		// "Set creates or overwrites the document with the given data."
		// I.e., Not Found is not a concern
		log.Printf("%s.fstore.Update, Firestore Set (with MergeAll) returned err: %v\n", sn, err)
		return ErrUpdateError
	}

	// read back the complete, updated document
	// docsnap, err := docRef.Get(ctx)
	// if err != nil {
	// 	log.Printf("%s.fstore.Update, docRef.Get returned err: %+v\n", sn, err)
	// 	return ErrUpdateError
	// }
	// log.Printf("after Set, Get docsnap: %+v\n", docsnap)

	// extract data into temp map
	// var tmpMap = make(map[string]interface{})
	// if err := docsnap.DataTo(&tmpMap); err != nil {
	// 	log.Printf("%s.fstore.Update, DataTo returned err: %+v\n", sn, err)
	// 	return ErrUpdateError
	// }

	// log.Printf("%s.fstore.Update, updated request: +%v\n", sn, request)

	return nil
}

var ErrCreateError = fmt.Errorf("fstore Create error")
var ErrZeroUUIDError = fmt.Errorf("fstore zero UUID error")
var ErrUpdateError = fmt.Errorf("fstore Update error")
var ErrNotFoundError = fmt.Errorf("fstore Not Found error")
var ErrFindError = fmt.Errorf("fstore Find error")
