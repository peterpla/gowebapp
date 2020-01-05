package request

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-playground/validator"
	"github.com/golang/gddo/httputil/header"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"

	"github.com/peterpla/lead-expert/pkg/serviceInfo"
)

const Pending string = "PENDING"
const Error string = "ERROR"
const Completed string = "COMPLETED"

// special UUIDs used for testing purposes
var PendingUUIDStr = "da4ae569-484d-4f59-bc52-c876058252d8"
var PendingUUID = uuid.MustParse(PendingUUIDStr)
var ErrorUUIDStr = "aa4073c3-5ae8-4344-9c29-41e15414e609"
var ErrorUUID = uuid.MustParse(ErrorUUIDStr)
var CompletedUUIDStr = "6697be3b-bdfa-4438-9e2a-ea1511dd0e40"
var CompletedUUID = uuid.MustParse(CompletedUUIDStr)

// ErrTimestampsKeyExists - key provided already exists
var ErrTimestampsKeyExists = errors.New("Timestamps key exists")

// ErrInvalidTime - time provided does not parse
var ErrInvalidTime = errors.New("Invalid time value cannot be parsed")

// Request defines properties of an incoming transcription request
// to be added
type Request struct {
	RequestID         uuid.UUID         `json:"request_id" firestore:"-"` // redundant when Firestore docID = RequestID
	CustomerID        int               `json:"customer_id" firestore:"customer_id" validate:"required,gte=1,lt=10000000"`
	MediaFileURI      string            `json:"media_uri" firestore:"media_uri" validate:"required,uri"`
	Status            string            `json:"status" firestore:"status"`                             // one of "PENDING", "ERROR", "COMPLETED"
	OriginalStatus    int               `json:"original_status" firestore:"original_status,omitempty"` // as reported throughout the pipeline
	AcceptedAt        string            `json:"accepted_at" firestore:"accepted_at"`
	UpdatedAt         string            `json:"updated_at" firestore:"updated_at,omitempty"`
	CompletedAt       string            `json:"completed_at" firestore:"completed_at,omitempty"`
	WorkingTranscript string            `json:"working_transcript" firestore:"working_transcript,omitempty"`
	FinalTranscript   string            `json:"final_transcript" firestore:"final_transcript,omitempty"`
	Timestamps        map[string]string `json:"timestamps" firestore:"timestamps"`
}

type RequestRepository interface {
	Create(request *Request) error
	FindByID(reqID uuid.UUID) (*Request, error)
	Update(request *Request) error
}

func (req *Request) ReadRequest(w http.ResponseWriter, r *http.Request, p httprouter.Params, validate *validator.Validate) error {
	sn := serviceInfo.GetServiceName()

	var err error

	err = decodeJSONBody(w, r, req)

	if err != nil {
		var mr *malformedRequest
		if errors.As(err, &mr) {
			http.Error(w, mr.msg, mr.status)
		} else {
			log.Println("%s.request.ReadRequest, " + err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return err
	}

	// log.Printf("%s.readRequest - decoded request: %+v\n", sn, newRequest)

	// validate incoming request
	// See https://github.com/go-playground/validator/blob/master/doc.go
	err = validate.Struct(req)
	if err != nil {
		log.Printf("%s.request.ReadRequest, validation error: %v\n", sn, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}
	// log.Printf("%s.request.ReadRequest - validated request: %+v\n", sn, newRequest)

	return nil
}

// ********** ********** ********** ********** ********** **********

// Alex Edwards, "How to Parse a JSON Request Body in Go"
// https://www.alexedwards.net/blog/how-to-properly-parse-a-json-request-body

type malformedRequest struct {
	status int
	msg    string
}

func (mr *malformedRequest) Error() string {
	return mr.msg
}

func decodeJSONBody(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	if r.Header.Get("Content-Type") != "" {
		value, _ := header.ParseValueAndParams(r.Header, "Content-Type")
		if value != "application/json" {
			msg := "Content-Type header is not application/json"
			return &malformedRequest{status: http.StatusUnsupportedMediaType, msg: msg}
		}
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(&dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		case errors.As(err, &syntaxError):
			msg := fmt.Sprintf("Request body contains badly-formed JSON (at position %d)", syntaxError.Offset)
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case errors.Is(err, io.ErrUnexpectedEOF):
			msg := fmt.Sprintf("Request body contains badly-formed JSON")
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case errors.As(err, &unmarshalTypeError):
			msg := fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			msg := fmt.Sprintf("Request body contains unknown field %s", fieldName)
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case errors.Is(err, io.EOF):
			msg := "Request body must not be empty"
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case err.Error() == "http: request body too large":
			msg := "Request body must not be larger than 1MB"
			return &malformedRequest{status: http.StatusRequestEntityTooLarge, msg: msg}

		default:
			return err
		}
	}

	if dec.More() {
		msg := "Request body must only contain a single JSON object"
		return &malformedRequest{status: http.StatusBadRequest, msg: msg}
	}

	return nil
}

// ********** ********** ********** ********** ********** **********

// PostResponse holds seelcted fields of Result struct to include in
// HTTP response to initial POST request
type PostResponse struct {
	RequestID    uuid.UUID `json:"request_id"`
	CustomerID   int       `json:"customer_id"`
	MediaFileURI string    `json:"media_uri"`
	AcceptedAt   string    `json:"accepted_at"`
	PollEndpoint string    `json:"poll_endpoint,omitempty"`
}

// GetQueueResponse holds seelcted fields of Result struct to include in
// HTTP response to GET /queue/{uuid} request
type GetQueueResponse struct {
	RequestID         uuid.UUID `json:"request_id"`
	CustomerID        int       `json:"customer_id"`
	MediaFileURI      string    `json:"media_uri"`
	AcceptedAt        string    `json:"accepted_at"`
	OriginalRequestID uuid.UUID `json:"original_request_id"`
	ETA               string    `json:"eta,omitempty"`             // time.Time.String()
	Endpoint          string    `json:"endpoint,omitempty"`        // uri
	OriginalStatus    int       `json:"original_status,omitempty"` // http.Status*
}

type GetTranscriptResponse struct {
	RequestID    uuid.UUID `json:"request_id"`
	CustomerID   int       `json:"customer_id" validate:"required,gte=1,lt=10000000"`
	MediaFileURI string    `json:"media_uri"`
	AcceptedAt   string    `json:"accepted_at"`
	CompletedAt  string    `json:"completed_at"`
	CompletedID  uuid.UUID `json:"completed_id"`
	Transcript   string    `json:"transcript"`
}

func (req *Request) AddTimestamps(startKey, startTimestamp, endKey string) (time.Duration, error) {

	var badTime time.Duration
	var startTime time.Time
	var err error

	if startTime, err = time.Parse(time.RFC3339Nano, startTimestamp); err != nil {
		log.Printf("request.AddTimestamps ERROR: startTime %s does not parse (RFC3339Nano)\n", startTimestamp)
		return badTime, ErrTimestampsKeyExists
	}

	// initialize map if needed
	if req.Timestamps == nil {
		// log.Printf("request.AddTimestamps initializing Timestamps map in Request\n")
		req.Timestamps = make(map[string]string)
	}

	// if startKey already exists, return error
	startKeyValue, ok := req.Timestamps[startKey]
	if ok {
		log.Printf("request.AddTimestamps ERROR: key %s exists with value %s\n", startKey, startKeyValue)
		return badTime, ErrTimestampsKeyExists
	}

	// if endKey already exists, return error
	endKeyValue, ok := req.Timestamps[endKey]
	if ok {
		log.Printf("request.AddTimestamp ERROR: key %s exists with value %s\n", endKey, endKeyValue)
		return badTime, ErrTimestampsKeyExists
	}

	// set startKey to startTimestamp
	req.Timestamps[startKey] = startTimestamp

	// set endKey to current time
	now := time.Now().UTC()
	req.Timestamps[endKey] = now.Format(time.RFC3339Nano)
	duration := now.Sub(startTime)

	return duration, nil
}

func (req *Request) RequestDuration() (time.Duration, error) {
	var badDuration time.Duration
	var accepted, completed time.Time
	var err error

	accepted, err = time.Parse(time.RFC3339Nano, req.AcceptedAt)
	if err != nil {
		log.Printf("%s.request.RequestDuration, time.Parse error: %v from AcceptedAt: %v\n", serviceInfo.GetServiceName(), err, req.AcceptedAt)
		return badDuration, err
	}
	completed, _ = time.Parse(time.RFC3339Nano, req.CompletedAt)
	if err != nil {
		log.Printf("%s.request.RequestDuration, time.Parse error: %v from CompletedAt: %v\n", serviceInfo.GetServiceName(), err, req.CompletedAt)
		return badDuration, err
	}

	return completed.Sub(accepted), nil

}

func (req *Request) ToMap() (map[string]interface{}, error) {
	sn := serviceInfo.GetServiceName()

	emptyMap := make(map[string]interface{})
	newMap := make(map[string]interface{})

	var reqJSON []byte
	var err error

	// TODO: more efficient Request->map conversion than JSON marshal/unmarshal
	// first generate JSON representation of Request
	if reqJSON, err = json.Marshal(req); err != nil {
		log.Printf("%s.request.ToMap, json.Marshal error: %v\n", sn, err)
		return emptyMap, err
	}

	// unmarshall the JSON into the map
	if err := json.Unmarshal(reqJSON, &newMap); err != nil {
		log.Printf("%s.request.ToMap, json.Unmarshal error: %v\n", sn, err)
		return emptyMap, err
	}
	// log.Printf("%s.request.ToMap, returning newMap: %+v\n", sn, newMap)

	return newMap, nil
}
