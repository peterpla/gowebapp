package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/golang/gddo/httputil/header"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/spf13/viper"

	"github.com/peterpla/lead-expert/pkg/adding"
	"github.com/peterpla/lead-expert/pkg/config"
	"github.com/peterpla/lead-expert/pkg/middleware"
	"github.com/peterpla/lead-expert/pkg/serviceInfo"
)

var prefix = "TaskDefault"
var initLogPrefix = "default.main.init(),"
var cfg config.Config

func init() {
	if err := config.GetConfig(&cfg); err != nil {
		msg := fmt.Sprintf(initLogPrefix+" GetConfig error: %v", err)
		panic(msg)
	}

	// set ServiceName, QueueName and NextServiceName appropriately
	cfg.ServiceName = viper.GetString(prefix + "SvcName")
	cfg.QueueName = viper.GetString(prefix + "WriteToQ")
	cfg.NextServiceName = viper.GetString(prefix + "NextSvcToHandleReq")

	// register them for access by other packages in this service
	serviceInfo.RegisterServiceName(cfg.ServiceName)
	serviceInfo.RegisterQueueName(cfg.QueueName)
	serviceInfo.RegisterNextServiceName(cfg.NextServiceName)

	config.SetConfigPointer(&cfg)
}

func main() {
	// log.Printf("Enter default.main\n")

	router := httprouter.New()
	router.POST("/api/v1/requests", postHandler(cfg.Adder))
	router.GET("/", indexHandler)
	router.NotFound = http.HandlerFunc(myNotFound)
	cfg.Router = router

	port := os.Getenv("PORT") // Google App Engine complains if "PORT" env var isn't checked
	if !cfg.IsGAE {
		port = viper.GetString(prefix + "Port")
	}
	if port == "" {
		panic("PORT undefined")
	}

	log.Printf("Service %s listening on port %s, requests will be added to queue %s",
		cfg.ServiceName, port, cfg.QueueName)
	log.Fatal(http.ListenAndServe(":"+port, middleware.LogReqResp(router)))
}

// postHandler returns the handler func for POST /requests
func postHandler(a adding.Service) httprouter.Handle {
	var err error
	sn := cfg.ServiceName
	// use a single instance of Validate, it caches struct info
	var validate *validator.Validate

	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		// log.Printf("%s.main.postHandler, enter\n", sn)
		startTime := time.Now().UTC()

		var newRequest adding.Request

		err = decodeJSONBody(w, r, &newRequest)
		if err != nil {
			var mr *malformedRequest
			if errors.As(err, &mr) {
				http.Error(w, mr.msg, mr.status)
			} else {
				log.Println("%s.postHandler, " + err.Error())
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}
		// log.Printf("%s.taskHandler - decoded request: %+v\n", sn, newRequest)

		// validate incoming request
		// See https://github.com/go-playground/validator/blob/master/doc.go
		validate = validator.New()
		err := validate.Struct(newRequest)
		if err != nil {
			// log.Printf("%s.main.postHandler, validation error: %v\n", sn, err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		newRequest.RequestID = uuid.New()
		newRequest.AcceptedAt = time.Now().UTC().Format(time.RFC3339Nano)

		// add timestamps and get duration
		var duration time.Duration
		if duration, err = newRequest.AddTimestamps("BeginDefault", startTime.Format(time.RFC3339Nano), "EndDefault"); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// add the request (e.g., to a queue) for subsequent processing
		returnedReq := a.AddRequest(newRequest)

		// provide selected fields of Request as the HTTP response
		response := adding.PostResponse{
			RequestID:    returnedReq.RequestID,
			CustomerID:   returnedReq.CustomerID,
			MediaFileURI: returnedReq.MediaFileURI,
			AcceptedAt:   returnedReq.AcceptedAt,
		}

		// send response to client
		w.WriteHeader(http.StatusAccepted)
		w.Header().Set("Content-Type", "application/json")
		if err = json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("%s.postHandler, json.NewEncoder.Encode error: +%v\n", sn, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Printf("%s.postHandler, completed in %v, newRequest: %+v\n", sn, duration, newRequest)
	}
}

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

// indexHandler serves as a health check, responding "service running"
func indexHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	sn := cfg.ServiceName
	// log.Printf("Enter %s.indexHandler\n", sn)
	if r.URL.Path != "/" {
		log.Printf("%s.indexHandler, r.URL.Path: %s, will respond NotFound\n", sn, r.URL.Path)
		http.NotFound(w, r)
		return
	}
	// indicate service is running
	fmt.Fprintf(w, "%q service running\n", sn)
}

func myNotFound(w http.ResponseWriter, r *http.Request) {
	// sn := cfg.ServiceName
	// log.Printf("%s.myNotFound, request for %s not routed\n", sn, r.URL.Path)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	_, _ = w.Write([]byte("<h2>404 Not Foundw</h2>"))
}
