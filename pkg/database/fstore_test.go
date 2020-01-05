package database

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/go-playground/validator"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"

	"github.com/peterpla/lead-expert/pkg/config"
	"github.com/peterpla/lead-expert/pkg/request"
)

var cfg config.Config
var testProject string
var testColl = "leadexperts-requests-TEST"

var repo requestRepository

var validate *validator.Validate

var testUUID = uuid.UUID{}
var expectedReq request.Request
var updatedReq request.Request
var emptyRequest = request.Request{}

func TestFirestore(t *testing.T) {

	initTestCollection()

	t.Run("TestNewFirestoreRequestRepository", func(t *testing.T) {

		// same inputs we use in initRepo() should produce equal result
		got := NewFirestoreRequestRepository(testProject, testColl)

		if !cmp.Equal(repo, got) {
			t.Errorf("NewFirestoreRequestRepository mismatch, expected %v, got %v", repo, got)
		}
	})

	t.Run("TestCreate", func(t *testing.T) {

		testUUID = uuid.New()

		startTime := time.Now().UTC()
		jsonBody := `{ "customer_id": 1234567, "media_uri": "gs://elated-practice-224603.appspot.com/audio_uploads/audio-01.mp3" }`

		/* ********** ********** ********** ********** ********** */
		// Mimic the work done by cmd/server/main.go/postHandler, so the
		// Request looks the same

		// create a http.ResponseWriter and http.Request, to pass to ReadRequest
		rw := httptest.NewRecorder()
		httpReq, err := http.NewRequest("POST", "api/v1/requests", strings.NewReader(jsonBody))
		if err != nil {
			t.Fatal(err)
		}

		// ReadRequest processes/validates the http.Request, populates the Request struct
		if err = expectedReq.ReadRequest(rw, httpReq, nil, validate); err != nil {
			t.Fatal(err)
		}
		// set Request fields like cmd/server/main.go/postHandler
		expectedReq.RequestID = testUUID
		expectedReq.AcceptedAt = time.Now().UTC().Format(time.RFC3339Nano)
		expectedReq.Status = request.Pending

		// add timestamps and get duration
		if _, err := expectedReq.AddTimestamps("BeginDefault", startTime.Format(time.RFC3339Nano), "EndDefault"); err != nil {
			t.Errorf("Addstamps returned err: %v", err)
		}

		// Ready to write a realistic Request to the database
		if err := repo.Create(&expectedReq); err != nil {
			t.Errorf("Create returned err: %v", err)
		}

		// read it back
		var gotReq *request.Request
		if gotReq, err = repo.FindByID(testUUID); err != nil {
			t.Errorf("FindByID returned err: %v", err)
		}
		if !cmp.Equal(expectedReq, *gotReq) {
			findMismatch(t, expectedReq, *gotReq)

			t.Errorf("Expected %v, got %v", expectedReq, *gotReq)
		}

		// test that creating with zero UUID is blocked
		// start with a copy of expectedReq that we just used successfully
		tmpReq := expectedReq
		tmpReq.RequestID = uuid.UUID{}
		if err := repo.Create(&tmpReq); err != ErrZeroUUIDError {
			t.Errorf("TestCreate, zero UUID, expected %v, got %v", ErrZeroUUIDError, err)
		}
	})

	t.Run("TestUpdate", func(t *testing.T) {
		var err error
		var completedTime = time.Now().UTC().Format(time.RFC3339Nano)

		// modify a copy of expectedReq, used earlier so it has AcceptedAt, Status, etc. values set
		updatedReq = expectedReq
		updatedReq.CompletedAt = completedTime
		if err := repo.Update(&updatedReq); err != nil {
			t.Errorf("TestUpdate, Update error: %v\n", err)
		}

		time.Sleep(time.Millisecond * 100) // guess at delay for updated values to be available

		// read back the updated Request and compare to what we wrote
		var gotReq *request.Request
		if gotReq, err = repo.FindByID(updatedReq.RequestID); err != nil {
			t.Errorf("TestUpdate, FindByID error: %v\n", err)
		}
		// ensure the CompletedAt value we updated was preserved
		if gotReq.CompletedAt != updatedReq.CompletedAt {
			t.Errorf("TestUpdate, CompletedAt expected %q, got %q\n", updatedReq.CompletedAt, gotReq.CompletedAt)
		}

		updatedReq.UpdatedAt = gotReq.UpdatedAt // now updatedReq should equal gotReq
		if !cmp.Equal(updatedReq, *gotReq) {
			findMismatch(t, updatedReq, *gotReq)

			t.Errorf("TestUpdate, expected %+v, got %+v", updatedReq, *gotReq)
		}
	})

	t.Run("TestFindByID", func(t *testing.T) {
		zeroUUID := uuid.UUID{}

		type test struct {
			name     string
			testID   uuid.UUID
			resultID uuid.UUID
			err      error
			result   request.Request
		}

		tests := []test{
			{name: "TestFindByID, zero UUID",
				testID:   zeroUUID,
				resultID: zeroUUID,
				err:      ErrZeroUUIDError,
				result:   emptyRequest,
			},
			{name: "TestFindByID, expected UUID",
				testID:   testUUID,
				resultID: testUUID,
				err:      nil,
				result:   updatedReq, // as Update'd above
			},
			{name: "TestFindByID, random UUID",
				testID:   uuid.New(),
				resultID: zeroUUID,
				err:      ErrNotFoundError,
				result:   emptyRequest,
			},
		}

		for _, tc := range tests {
			var gotReq *request.Request
			var err error
			gotReq, err = repo.FindByID(tc.testID)

			if tc.err != err {
				t.Errorf("%s: err expected %v, got %v", tc.name, tc.err, err)
			}
			if tc.resultID != gotReq.RequestID {
				t.Errorf("%s: uuid expected %v, got %v", tc.name, tc.resultID, gotReq.RequestID)
			}
			if !cmp.Equal(tc.result, *gotReq) {
				findMismatch(t, tc.result, *gotReq)

				t.Errorf("%s: request expected %+v, got %+v", tc.name, tc.result, gotReq)
			}
		}
	})

	// delete test collection
	deleteTestCollection()
}

func initTestCollection() {
	if err := config.GetConfig(&cfg, "/api/v1"); err != nil {
		msg := fmt.Sprintf("GetConfig error: %v", err)
		panic(msg)
	}
	testProject = cfg.ProjectID
	repo = requestRepository{testProject, testColl}

	// setup validator while we're at it
	validate = validator.New()
}

func deleteTestCollection() {

	const BATCHSIZE = 20

	ctx := context.Background()
	client, err := firestore.NewClient(ctx, testProject)
	if err != nil {
		log.Printf("failed to create Client: %v\n", err)
		return
	}
	defer client.Close()

	col := client.Collection(testColl)

	for {
		// Get a batch of documents
		iter := col.Limit(BATCHSIZE).Documents(ctx)
		numDeleted := 0

		// Iterate through the documents, adding
		// a delete operation for each one to a
		// WriteBatch.
		batch := client.Batch()
		for {
			doc, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				return
			}

			batch.Delete(doc.Ref)
			numDeleted++
			log.Printf("deleting record %d\n", numDeleted)
		}

		// If there are no documents to delete,
		// the process is over.
		if numDeleted == 0 {
			return
		}

		_, err := batch.Commit(ctx)
		if err != nil {
			return
		}

	}
}

/* ********** ********** ********** ********** ********** ********** */

func findMismatch(t *testing.T, expected request.Request, got request.Request) {

	var foundMismatch = false

	if expected.RequestID != got.RequestID {
		foundMismatch = true
		t.Errorf("RequestID: expected %q, got %q", expected.RequestID, got.RequestID)
	}
	if expected.CustomerID != got.CustomerID {
		foundMismatch = true
		t.Errorf("CustomerID: expected %d, got %d", expected.CustomerID, got.CustomerID)
	}
	if expected.MediaFileURI != got.MediaFileURI {
		foundMismatch = true
		t.Errorf("MediaFileURI: expected %q, got %q", expected.MediaFileURI, got.MediaFileURI)
	}
	if expected.Status != got.Status {
		foundMismatch = true
		t.Errorf("Status: expected %q, got %q", expected.Status, got.Status)
	}
	if expected.OriginalStatus != got.OriginalStatus {
		foundMismatch = true
		t.Errorf("OriginalStatus: expected %d got %d", expected.OriginalStatus, got.OriginalStatus)
	}
	if expected.AcceptedAt != got.AcceptedAt {
		foundMismatch = true
		t.Errorf("AcceptedAt: expected %q, got %q", expected.AcceptedAt, got.AcceptedAt)
	}
	if expected.CompletedAt != got.CompletedAt {
		foundMismatch = true
		t.Errorf("CompletedAt: expected %q, got %q", expected.CompletedAt, got.CompletedAt)
	}
	if expected.WorkingTranscript != got.WorkingTranscript {
		foundMismatch = true
		t.Errorf("WorkingTranscript: expected %q, got %q", expected.WorkingTranscript, got.WorkingTranscript)
	}
	if expected.FinalTranscript != got.FinalTranscript {
		foundMismatch = true
		t.Errorf("FinalTranscript: expected %q, got %q", expected.FinalTranscript, got.FinalTranscript)
	}

	if !foundMismatch {
		log.Println("findMismatch: mismatch NOT found")
	}
}

// func present(t *testing.T, theMap map[string]interface{}, key string, label string) bool {
// 	_, ok := theMap[key]
// 	if !ok {
// 		t.Errorf("%s missing %q", label, key)
// 	}
// 	return ok
// }

// func getBeginEndTimestamps(t *testing.T, theMap map[string]interface{}) (begin string, end string) {
// 	var temp map[string]interface{}
// 	var ok bool

// 	temp, ok = theMap["timestamps"].(map[string]interface{})
// 	if !ok {
// 		t.Errorf("\"Timestamps\" not present")
// 	}

// 	begin, ok = temp["BeginTest"].(string)
// 	if !ok {
// 		t.Errorf("\"BeginTest\" not present")
// 	}

// 	end, ok = temp["EndTest"].(string)
// 	if !ok {
// 		t.Errorf("\"EndTest\" not present")
// 	}

// 	return begin, end
// }
