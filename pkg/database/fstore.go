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
	sn := serviceInfo.GetServiceName()
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
func (r requestRepository) Create(req *request.Request) error {
	sn := serviceInfo.GetServiceName()
	// log.Printf("%s.fstore.Create, repo: %+v, req: %+v\n", sn, r, *req)

	// TODO: lock the request while it's being written?

	// get a map corresponding to the Request
	// reqMap, err := req.ToMap()
	// if err != nil {
	// 	log.Printf("%s.fstore.Create, ToMap err: %v\n", sn, err)
	// 	return err
	// }

	// TODO: normalize Request fields

	// block creation of zero UUID requests
	zeroUUID := uuid.UUID{}
	if req.RequestID == zeroUUID {
		log.Printf("%s.fstore.Create, zero UUID not allowed\n", sn)
		return ErrZeroUUIDError
	}

	// prepare to talk to Firestore
	ctx := context.Background()
	projID := r.ProjectID
	client, err := firestore.NewClient(ctx, projID)
	if err != nil {
		log.Printf("%s.fstore.Create, NewClient returned err: %v\n", sn, err)
		return ErrCreateError
	}
	defer client.Close()

	docID := req.RequestID.String() // request UUID = document ID, we'll search by the UUID later
	col := r.Collection
	colRef := client.Collection(col)
	docRef := colRef.Doc(docID)
	// log.Printf("%s.fstore.Create, calling Set() with client: %+v, col: %+v, colRef: %+v, docID: %+v, docRef: %+v, reqMap: %+v\n",
	// 	sn, client, col, colRef, docID, docRef, reqMap)
	// log.Printf("%s.fstore.Create, calling Set() with client: %+v,\n... col: %+v, colRef: %+v,\n... docID: %+v, docRef: %+v,\n... req: %+v\n",
	// 	sn, client, col, colRef, docID, docRef, *req)
	req.CreatedAt = time.Now().UTC().Format(time.RFC3339Nano)

	// _, err = docRef.Set(ctx, reqMap)
	_, err = docRef.Set(ctx, *req)
	if err != nil {
		log.Printf("%s.fstore.Create, Set returned err %+v\n", sn, err)
		return ErrCreateError
	}

	// read the Request back from the database, and return it
	// docsnap, err := docRef.Get(ctx)
	// if err != nil {
	// 	log.Printf("%s.fstore.Create, Get returned err: %+v\n", sn, err)
	// 	return ErrCreateError
	// }
	//
	// tmpReq := request.Request{}
	// if err := docsnap.DataTo(&tmpReq); err != nil {
	// 	log.Printf("%s.fstore.Create, DataTo returned err: %+v", sn, err)
	// 	return ErrCreateError
	// }

	log.Printf("%s.fstore.Create, req: +%v\n", sn, *req)

	return nil
}

func (r requestRepository) FindByID(reqID uuid.UUID) (*request.Request, error) {
	sn := serviceInfo.GetServiceName()
	// See Exercise as example: https://github.com/peterpla/exercise/blob/master/backend/

	var emptyRequest = request.Request{}

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
		log.Printf("%s.fstore.FindByID, NewClient returned err: %v\n", sn, err)
		return &emptyRequest, ErrFindError
	}
	defer client.Close()
	// log.Printf("firestore client: %+v\n", client)

	// search by UUID
	docID := reqID.String()
	col := r.Collection
	colRef := client.Collection(col)
	docRef := colRef.Doc(docID)
	// log.Printf("%s.fstore.FindByID, calling Get() with client: %+v,\n... col: %+v, colRef: %+v,\n... docID: %+v, docRef: %+v\n",
	// 	sn, client, col, colRef, docID, docRef)

	// read the user back from the database
	docsnap, err := docRef.Get(ctx)
	if err != nil {
		st, _ := status.FromError(err)
		if st.Code() == codes.NotFound {
			// UUID not found
			log.Printf("%s.fstore.FindByID, docID %q not found\n", sn, docID)
			return &emptyRequest, ErrNotFoundError
		}
		// some other error, return it
		log.Printf("%s.fstore.FindByID, Get returned err: %+v\n", sn, err)
		return &emptyRequest, ErrFindError
	}

	// extract data into the Request we'll return
	var foundRequest request.Request
	if err := docsnap.DataTo(&foundRequest); err != nil {
		log.Printf("%s.fstore.FindByID, DataTo returned err: %+v", sn, err)
		return &emptyRequest, ErrFindError
	}

	// save the UUID in RequestID as expected elsewhere
	foundRequest.RequestID = reqID

	log.Printf("%s.fstore.FindByID, foundRequest: %+v\n", sn, foundRequest)

	return &foundRequest, nil
}

// Update writes an updated Request to the database
func (r requestRepository) Update(req *request.Request) error {
	sn := serviceInfo.GetServiceName()

	// TODO: lock the request while it's being written?

	// block update of zero UUID requests
	zeroUUID := uuid.UUID{}
	if req.RequestID == zeroUUID {
		log.Printf("%s.fstore.Update, zero UUID not allowed\n", sn)
		return ErrZeroUUIDError
	}

	// to use MergeAll we must provide a map, so get a map corresponding to the Request
	reqMap, err := req.ToMap()
	if err != nil {
		log.Printf("%s.fstore.Create, ToMap err: %v\n", sn, err)
		return err
	}

	// req.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	reqMap["updated_at"] = time.Now().UTC().Format(time.RFC3339Nano)
	// log.Printf("reqMap: %+v\n", reqMap)

	// prepare to talk to Firestore
	// log.Printf("%s.fstore.Update, repo.ProjectID: %s, repo.Collection: %s\n", sn, repo.ProjectID, repo.Collection)
	ctx := context.Background()
	projID := r.ProjectID
	client, err := firestore.NewClient(ctx, projID)
	if err != nil {
		log.Printf("%s.fstore.Update, NewClient returned err: %v\n", sn, err)
		return ErrUpdateError
	}
	defer client.Close()
	// log.Printf("firestore client: %+v\n", client)

	docID := req.RequestID.String()
	col := r.Collection
	colRef := client.Collection(col)
	docRef := colRef.Doc(docID)
	// log.Printf("%s.fstore.Update, calling Set() with MergeAll, client: %+v,\n... col: %+v, colRef: %+v,\n... docID: %+v, docRef: %+v\n... reqMap: %+v\n",
	// 	sn, client, col, colRef, docID, docRef, reqMap)

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

	log.Printf("%s.fstore.Update, updated request: +%v\n", sn, req)

	return nil
}

var ErrCreateError = fmt.Errorf("fstore Create error")
var ErrZeroUUIDError = fmt.Errorf("fstore zero UUID error")
var ErrUpdateError = fmt.Errorf("fstore Update error")
var ErrNotFoundError = fmt.Errorf("fstore Not Found error")
var ErrFindError = fmt.Errorf("fstore Find error")
