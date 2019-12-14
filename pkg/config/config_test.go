package config

import (
	"log"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestGetConfig(t *testing.T) {

	var cfg Config

	defaultResult := Config{
		Adder:           nil,
		AppName:         "MyLeadExpert",
		ConfigFile:      "config.yaml",
		Description:     "More leads for local retailers. Generate more sales by routing your existing traffic through a proven conversion process.",
		IsGAE:           false,
		QueueName:       "",
		Router:          nil,
		ServiceName:     "",
		NextServiceName: "",
		StorageType:     Memory,
		// Key Management Service for encrypted config
		EncryptedBucket: "elated-practice-224603-lead-expert-secret",
		KmsKey:          "config",
		KmsKeyRing:      "devkeyring",
		KmsLocation:     "us-west2",
		//
		ProjectID:       "elated-practice-224603",
		StorageLocation: "us-west2",
		TasksLocation:   "us-west2",
		// port number used by each service
		TaskDefaultPort:               "8080",
		TaskInitialRequestPort:        "8081",
		TaskServiceDispatchPort:       "8082",
		TaskTranscriptionGCPPort:      "8083",
		TaskTranscriptionCompletePort: "8084",
		TaskTranscriptQAPort:          "8085",
		TaskTranscriptQACompletePort:  "8086",
		TaskTaggingPort:               "8087",
		TaskTaggingCompletePort:       "8088",
		TaskTaggingQAPort:             "8089",
		TaskTaggingQACompletePort:     "8090",
		TaskCompletionProcessingPort:  "8091",
		// queue name used by each services
		TaskDefaultWriteToQ:               "InitialRequest",
		TaskInitialRequestWriteToQ:        "ServiceDispatch",
		TaskServiceDispatchWriteToQ:       "TranscriptionGCP",
		TaskTranscriptionGCPWriteToQ:      "TranscriptionComplete",
		TaskTranscriptionCompleteWriteToQ: "TranscriptQA",
		TaskTranscriptQAWriteToQ:          "TranscriptQAComplete",
		TaskTranscriptQACompleteWriteToQ:  "Tagging",
		TaskTaggingWriteToQ:               "TaggingComplete",
		TaskTaggingCompleteWriteToQ:       "TaggingQA",
		TaskTaggingQAWriteToQ:             "TaggingQAComplete",
		TaskTaggingQACompleteWriteToQ:     "CompletionProcessing",
		TaskCompletionProcessingWriteToQ:  "no-queue",
		// service name of each service
		TaskDefaultSvcName:               "default",
		TaskInitialRequestSvcName:        "initial-request",
		TaskServiceDispatchSvcName:       "service-dispatch",
		TaskTranscriptionGCPSvcName:      "transcription-gcp",
		TaskTranscriptionCompleteSvcName: "transcription-complete",
		TaskTranscriptQASvcName:          "transcript-qa",
		TaskTranscriptQACompleteSvcName:  "transcript-qa-complete",
		TaskTaggingSvcName:               "tagging",
		TaskTaggingCompleteSvcName:       "tagging-complete",
		TaskTaggingQASvcName:             "tagging-qa",
		TaskTaggingQACompleteSvcName:     "tagging-qa-complete",
		TaskCompletionProcessingSvcName:  "completion-processing",
		// next service in the chain to handle requests
		TaskDefaultNextSvcToHandleReq:               "initial-request",
		TaskInitialRequestNextSvcToHandleReq:        "service-dispatch",
		TaskServiceDispatchNextSvcToHandleReq:       "transcription-gcp",
		TaskTranscriptionGCPNextSvcToHandleReq:      "transcription-complete",
		TaskTranscriptionCompleteNextSvcToHandleReq: "transcript-qa",
		TaskTranscriptQANextSvcToHandleReq:          "transcript-qa-complete",
		TaskTranscriptQACompleteNextSvcToHandleReq:  "tagging",
		TaskTaggingNextSvcToHandleReq:               "tagging-complete",
		TaskTaggingCompleteNextSvcToHandleReq:       "tagging-qa",
		TaskTaggingQANextSvcToHandleReq:             "tagging-qa-complete",
		TaskTaggingQACompleteNextSvcToHandleReq:     "completion-processing",
		TaskCompletionProcessingNextSvcToHandleReq:  "no-service",
		//
		Verbose: false,
		Version: "0.1.0",
		Help:    false,
	}

	// guard against calling twice, which will trigger panic with "flag redefined"
	if cfg.AppName == "" { // uninitialized
		if err := GetConfig(&cfg); err != nil {
			t.Fatalf("error from GetConfig: %v", err)
		}

		// CHEAT: nil-out the actual .Adder
		cfg.Adder = nil

		if !cmp.Equal(defaultResult, cfg) {
			findMismatch(t, defaultResult, cfg)
			t.Fatalf("expected %+v, got %+v", defaultResult, cfg)
		}
	}
}

func findMismatch(t *testing.T, expected Config, got Config) {

	var foundMismatch = false

	if expected.Adder != got.Adder {
		foundMismatch = true
		log.Printf("Adder: expected +%v, got %+v", expected.Adder, got.Adder)
	}
	if expected.AppName != got.AppName {
		foundMismatch = true
		log.Printf("AppName: expected +%v, got %+v", expected.AppName, got.AppName)
	}
	if expected.ConfigFile != got.ConfigFile {
		foundMismatch = true
		log.Printf("ConfigFile: expected +%v, got %+v", expected.ConfigFile, got.ConfigFile)
	}
	if expected.Description != got.Description {
		foundMismatch = true
		log.Printf("Description: expected +%v, got %+v", expected.Description, got.Description)
	}
	if expected.IsGAE != got.IsGAE {
		foundMismatch = true
		log.Printf("IsGAE: expected +%v, got %+v", expected.IsGAE, got.IsGAE)
	}
	if expected.QueueName != got.QueueName {
		foundMismatch = true
		log.Printf("QueueName: expected +%v, got %+v", expected.QueueName, got.QueueName)
	}
	if expected.Router != got.Router {
		foundMismatch = true
		log.Printf("Router: expected +%v, got %+v", expected.Router, got.Router)
	}
	if expected.ServiceName != got.ServiceName {
		foundMismatch = true
		log.Printf("ServiceName: expected +%v, got %+v", expected.ServiceName, got.ServiceName)
	}
	if expected.NextServiceName != got.NextServiceName {
		foundMismatch = true
		log.Printf("NextServiceName: expected +%v, got %+v", expected.NextServiceName, got.NextServiceName)
	}
	if expected.StorageType != got.StorageType {
		foundMismatch = true
		log.Printf("StorageType: expected +%v, got %+v", expected.StorageType, got.StorageType)
	}
	if expected.EncryptedBucket != got.EncryptedBucket {
		foundMismatch = true
		log.Printf("EncryptedBucket: expected +%v, got %+v", expected.EncryptedBucket, got.EncryptedBucket)
	}

	// Key Management Service for encrypted config
	if expected.KmsKey != got.KmsKey {
		foundMismatch = true
		log.Printf("KmsKey: expected +%v, got %+v", expected.KmsKey, got.KmsKey)
	}
	if expected.KmsKeyRing != got.KmsKeyRing {
		foundMismatch = true
		log.Printf("KmsKeyRing: expected +%v, got %+v", expected.KmsKeyRing, got.KmsKeyRing)
	}
	if expected.KmsLocation != got.KmsLocation {
		foundMismatch = true
		log.Printf("KmsLocation: expected +%v, got %+v", expected.KmsLocation, got.KmsLocation)
	}

	//
	if expected.ProjectID != got.ProjectID {
		foundMismatch = true
		log.Printf("ProjectID: expected +%v, got %+v", expected.ProjectID, got.ProjectID)
	}
	if expected.StorageLocation != got.StorageLocation {
		foundMismatch = true
		log.Printf("StorageLocation: expected +%v, got %+v", expected.StorageLocation, got.StorageLocation)
	}
	if expected.TasksLocation != got.TasksLocation {
		foundMismatch = true
		log.Printf("TasksLocation: expected +%v, got %+v", expected.TasksLocation, got.TasksLocation)
	}

	// port used by each service: Task*Port
	if expected.TaskDefaultPort != got.TaskDefaultPort {
		foundMismatch = true
		log.Printf("TaskDefaultPort: expected +%v, got %+v", expected.TaskDefaultPort, got.TaskDefaultPort)
	}
	if expected.TaskInitialRequestPort != got.TaskInitialRequestPort {
		foundMismatch = true
		log.Printf("TaskInitialRequestPort: expected +%v, got %+v", expected.TaskInitialRequestPort, got.TaskInitialRequestPort)
	}
	if expected.TaskServiceDispatchPort != got.TaskServiceDispatchPort {
		foundMismatch = true
		log.Printf("TaskServiceDispatchPort: expected +%v, got %+v", expected.TaskServiceDispatchPort, got.TaskServiceDispatchPort)
	}
	if expected.TaskTranscriptionGCPPort != got.TaskTranscriptionGCPPort {
		foundMismatch = true
		log.Printf("TaskTranscriptionGCPPort: expected +%v, got %+v", expected.TaskTranscriptionGCPPort, got.TaskTranscriptionGCPPort)
	}
	if expected.TaskTranscriptionCompletePort != got.TaskTranscriptionCompletePort {
		foundMismatch = true
		log.Printf("TaskTranscriptionCompletePort: expected +%v, got %+v", expected.TaskTranscriptionCompletePort, got.TaskTranscriptionCompletePort)
	}
	if expected.TaskTranscriptQAPort != got.TaskTranscriptQAPort {
		foundMismatch = true
		log.Printf("TaskTranscriptQAPort: expected +%v, got %+v", expected.TaskTranscriptQAPort, got.TaskTranscriptQAPort)
	}
	if expected.TaskTranscriptQACompletePort != got.TaskTranscriptQACompletePort {
		foundMismatch = true
		log.Printf("TaskTranscriptQACompletePort: expected +%v, got %+v", expected.TaskTranscriptQACompletePort, got.TaskTranscriptQACompletePort)
	}
	if expected.TaskTaggingPort != got.TaskTaggingPort {
		foundMismatch = true
		log.Printf("TaskTaggingPort: expected +%v, got %+v", expected.TaskTaggingPort, got.TaskTaggingPort)
	}
	if expected.TaskTaggingCompletePort != got.TaskTaggingCompletePort {
		foundMismatch = true
		log.Printf("TaskTaggingCompletePort: expected +%v, got %+v", expected.TaskTaggingCompletePort, got.TaskTaggingCompletePort)
	}
	if expected.TaskTaggingQAPort != got.TaskTaggingQAPort {
		foundMismatch = true
		log.Printf("TaskTaggingQAPort: expected +%v, got %+v", expected.TaskTaggingQAPort, got.TaskTaggingQAPort)
	}
	if expected.TaskTaggingQACompletePort != got.TaskTaggingQACompletePort {
		foundMismatch = true
		log.Printf("TaskTaggingQACompletePort: expected +%v, got %+v", expected.TaskTaggingQACompletePort, got.TaskTaggingQACompletePort)
	}
	if expected.TaskCompletionProcessingPort != got.TaskCompletionProcessingPort {
		foundMismatch = true
		log.Printf("TaskCompletionProcessingPort: expected +%v, got %+v", expected.TaskCompletionProcessingPort, got.TaskCompletionProcessingPort)
	}

	// queue name used by each service: Task*WriteToQ
	if expected.TaskDefaultWriteToQ != got.TaskDefaultWriteToQ {
		foundMismatch = true
		log.Printf("TaskDefaultWriteToQ: expected +%v, got %+v", expected.TaskDefaultWriteToQ, got.TaskDefaultWriteToQ)
	}
	if expected.TaskInitialRequestWriteToQ != got.TaskInitialRequestWriteToQ {
		foundMismatch = true
		log.Printf("TaskInitialRequestWriteToQ: expected +%v, got %+v", expected.TaskInitialRequestWriteToQ, got.TaskInitialRequestWriteToQ)
	}
	if expected.TaskServiceDispatchWriteToQ != got.TaskServiceDispatchWriteToQ {
		foundMismatch = true
		log.Printf("TaskServiceDispatchWriteToQ: expected +%v, got %+v", expected.TaskServiceDispatchWriteToQ, got.TaskServiceDispatchWriteToQ)
	}
	if expected.TaskTranscriptionGCPWriteToQ != got.TaskTranscriptionGCPWriteToQ {
		foundMismatch = true
		log.Printf("TaskTranscriptionGCPWriteToQ: expected +%v, got %+v", expected.TaskTranscriptionGCPWriteToQ, got.TaskTranscriptionGCPWriteToQ)
	}
	if expected.TaskTranscriptionCompleteWriteToQ != got.TaskTranscriptionCompleteWriteToQ {
		foundMismatch = true
		log.Printf("TaskTranscriptionCompleteWriteToQ: expected +%v, got %+v", expected.TaskTranscriptionCompleteWriteToQ, got.TaskTranscriptionCompleteWriteToQ)
	}
	if expected.TaskTranscriptQAWriteToQ != got.TaskTranscriptQAWriteToQ {
		foundMismatch = true
		log.Printf("TaskTranscriptQAWriteToQ: expected +%v, got %+v", expected.TaskTranscriptQAWriteToQ, got.TaskTranscriptQAWriteToQ)
	}
	if expected.TaskTranscriptQACompleteWriteToQ != got.TaskTranscriptQACompleteWriteToQ {
		foundMismatch = true
		log.Printf("TaskTranscriptQACompleteWriteToQ: expected +%v, got %+v", expected.TaskTranscriptQACompleteWriteToQ, got.TaskTranscriptQACompleteWriteToQ)
	}
	if expected.TaskTaggingWriteToQ != got.TaskTaggingWriteToQ {
		foundMismatch = true
		log.Printf("TaskTaggingWriteToQ: expected +%v, got %+v", expected.TaskTaggingWriteToQ, got.TaskTaggingWriteToQ)
	}
	if expected.TaskTaggingCompleteWriteToQ != got.TaskTaggingCompleteWriteToQ {
		foundMismatch = true
		log.Printf("TaskTaggingCompleteWriteToQ: expected +%v, got %+v", expected.TaskTaggingCompleteWriteToQ, got.TaskTaggingCompleteWriteToQ)
	}
	if expected.TaskTaggingQAWriteToQ != got.TaskTaggingQAWriteToQ {
		foundMismatch = true
		log.Printf("TaskTaggingQAWriteToQ: expected +%v, got %+v", expected.TaskTaggingQAWriteToQ, got.TaskTaggingQAWriteToQ)
	}
	if expected.TaskTaggingQACompleteWriteToQ != got.TaskTaggingQACompleteWriteToQ {
		foundMismatch = true
		log.Printf("TaskTaggingQACompleteWriteToQ: expected +%v, got %+v", expected.TaskTaggingQACompleteWriteToQ, got.TaskTaggingQACompleteWriteToQ)
	}
	if expected.TaskCompletionProcessingWriteToQ != got.TaskCompletionProcessingWriteToQ {
		foundMismatch = true
		log.Printf("TaskCompletionProcessingWriteToQ: expected +%v, got %+v", expected.TaskCompletionProcessingWriteToQ, got.TaskCompletionProcessingWriteToQ)
	}

	// service name of each service: Task*SvcName
	if expected.TaskDefaultSvcName != got.TaskDefaultSvcName {
		foundMismatch = true
		log.Printf("TaskDefaultSvcName: expected +%v, got %+v", expected.TaskDefaultSvcName, got.TaskDefaultSvcName)
	}
	if expected.TaskInitialRequestSvcName != got.TaskInitialRequestSvcName {
		foundMismatch = true
		log.Printf("TaskInitialRequestSvcName: expected +%v, got %+v", expected.TaskInitialRequestSvcName, got.TaskInitialRequestSvcName)
	}
	if expected.TaskServiceDispatchSvcName != got.TaskServiceDispatchSvcName {
		foundMismatch = true
		log.Printf("TaskServiceDispatchSvcName: expected +%v, got %+v", expected.TaskServiceDispatchSvcName, got.TaskServiceDispatchSvcName)
	}
	if expected.TaskTranscriptionGCPSvcName != got.TaskTranscriptionGCPSvcName {
		foundMismatch = true
		log.Printf("TaskTranscriptionGCPSvcName: expected +%v, got %+v", expected.TaskTranscriptionGCPSvcName, got.TaskTranscriptionGCPSvcName)
	}
	if expected.TaskTranscriptionCompleteSvcName != got.TaskTranscriptionCompleteSvcName {
		foundMismatch = true
		log.Printf("TaskTranscriptionCompleteSvcName: expected +%v, got %+v", expected.TaskTranscriptionCompleteSvcName, got.TaskTranscriptionCompleteSvcName)
	}
	if expected.TaskTranscriptQASvcName != got.TaskTranscriptQASvcName {
		foundMismatch = true
		log.Printf("TaskTranscriptQASvcName: expected +%v, got %+v", expected.TaskTranscriptQASvcName, got.TaskTranscriptQASvcName)
	}
	if expected.TaskTranscriptQACompleteSvcName != got.TaskTranscriptQACompleteSvcName {
		foundMismatch = true
		log.Printf("TaskTranscriptQACompleteSvcName: expected +%v, got %+v", expected.TaskTranscriptQACompleteSvcName, got.TaskTranscriptQACompleteSvcName)
	}
	if expected.TaskTaggingSvcName != got.TaskTaggingSvcName {
		foundMismatch = true
		log.Printf("TaskTaggingSvcName: expected +%v, got %+v", expected.TaskTaggingSvcName, got.TaskTaggingSvcName)
	}
	if expected.TaskTaggingCompleteSvcName != got.TaskTaggingCompleteSvcName {
		foundMismatch = true
		log.Printf("TaskTaggingCompleteSvcName: expected +%v, got %+v", expected.TaskTaggingCompleteSvcName, got.TaskTaggingCompleteSvcName)
	}
	if expected.TaskTaggingQASvcName != got.TaskTaggingQASvcName {
		foundMismatch = true
		log.Printf("TaskTaggingQASvcName: expected +%v, got %+v", expected.TaskTaggingQASvcName, got.TaskTaggingQASvcName)
	}
	if expected.TaskTaggingQACompleteSvcName != got.TaskTaggingQACompleteSvcName {
		foundMismatch = true
		log.Printf("TaskTaggingQACompleteSvcName: expected +%v, got %+v", expected.TaskTaggingQACompleteSvcName, got.TaskTaggingQACompleteSvcName)
	}
	if expected.TaskCompletionProcessingSvcName != got.TaskCompletionProcessingSvcName {
		foundMismatch = true
		log.Printf("TaskCompletionProcessingSvcName: expected +%v, got %+v", expected.TaskCompletionProcessingSvcName, got.TaskCompletionProcessingSvcName)
	}

	// next service in the chain to handle requests: Task*NextSvcToHandleReq
	if expected.TaskDefaultNextSvcToHandleReq != got.TaskDefaultNextSvcToHandleReq {
		foundMismatch = true
		log.Printf("TaskDefaultNextSvcToHandleReq: expected +%v, got %+v", expected.TaskDefaultNextSvcToHandleReq, got.TaskDefaultNextSvcToHandleReq)
	}
	if expected.TaskInitialRequestNextSvcToHandleReq != got.TaskInitialRequestNextSvcToHandleReq {
		foundMismatch = true
		log.Printf("TaskInitialRequestNextSvcToHandleReq: expected +%v, got %+v", expected.TaskInitialRequestNextSvcToHandleReq, got.TaskInitialRequestNextSvcToHandleReq)
	}
	if expected.TaskServiceDispatchNextSvcToHandleReq != got.TaskServiceDispatchNextSvcToHandleReq {
		foundMismatch = true
		log.Printf("TaskServiceDispatchNextSvcToHandleReq: expected +%v, got %+v", expected.TaskServiceDispatchNextSvcToHandleReq, got.TaskServiceDispatchNextSvcToHandleReq)
	}
	if expected.TaskTranscriptionGCPNextSvcToHandleReq != got.TaskTranscriptionGCPNextSvcToHandleReq {
		foundMismatch = true
		log.Printf("TaskTranscriptionGCPNextSvcToHandleReq: expected +%v, got %+v", expected.TaskTranscriptionGCPNextSvcToHandleReq, got.TaskTranscriptionGCPNextSvcToHandleReq)
	}
	if expected.TaskTranscriptionCompleteNextSvcToHandleReq != got.TaskTranscriptionCompleteNextSvcToHandleReq {
		foundMismatch = true
		log.Printf("TaskTranscriptionCompleteNextSvcToHandleReq: expected +%v, got %+v", expected.TaskTranscriptionCompleteNextSvcToHandleReq, got.TaskTranscriptionCompleteNextSvcToHandleReq)
	}
	if expected.TaskTranscriptQANextSvcToHandleReq != got.TaskTranscriptQANextSvcToHandleReq {
		foundMismatch = true
		log.Printf("TaskTranscriptQANextSvcToHandleReq: expected +%v, got %+v", expected.TaskTranscriptQANextSvcToHandleReq, got.TaskTranscriptQANextSvcToHandleReq)
	}
	if expected.TaskTranscriptQACompleteNextSvcToHandleReq != got.TaskTranscriptQACompleteNextSvcToHandleReq {
		foundMismatch = true
		log.Printf("TaskTranscriptQACompleteNextSvcToHandleReq: expected +%v, got %+v", expected.TaskTranscriptQACompleteNextSvcToHandleReq, got.TaskTranscriptQACompleteNextSvcToHandleReq)
	}
	if expected.TaskTaggingNextSvcToHandleReq != got.TaskTaggingNextSvcToHandleReq {
		foundMismatch = true
		log.Printf("TaskTaggingNextSvcToHandleReq: expected +%v, got %+v", expected.TaskTaggingNextSvcToHandleReq, got.TaskTaggingNextSvcToHandleReq)
	}
	if expected.TaskTaggingCompleteNextSvcToHandleReq != got.TaskTaggingCompleteNextSvcToHandleReq {
		foundMismatch = true
		log.Printf("TaskTaggingCompleteNextSvcToHandleReq: expected +%v, got %+v", expected.TaskTaggingCompleteNextSvcToHandleReq, got.TaskTaggingCompleteNextSvcToHandleReq)
	}
	if expected.TaskTaggingQANextSvcToHandleReq != got.TaskTaggingQANextSvcToHandleReq {
		foundMismatch = true
		log.Printf("TaskTaggingQANextSvcToHandleReq: expected +%v, got %+v", expected.TaskTaggingQANextSvcToHandleReq, got.TaskTaggingQANextSvcToHandleReq)
	}
	if expected.TaskTaggingQACompleteNextSvcToHandleReq != got.TaskTaggingQACompleteNextSvcToHandleReq {
		foundMismatch = true
		log.Printf("TaskTaggingQACompleteNextSvcToHandleReq: expected +%v, got %+v", expected.TaskTaggingQACompleteNextSvcToHandleReq, got.TaskTaggingQACompleteNextSvcToHandleReq)
	}
	if expected.TaskCompletionProcessingNextSvcToHandleReq != got.TaskCompletionProcessingNextSvcToHandleReq {
		foundMismatch = true
		log.Printf("TaskCompletionProcessingNextSvcToHandleReq: expected +%v, got %+v", expected.TaskCompletionProcessingNextSvcToHandleReq, got.TaskCompletionProcessingNextSvcToHandleReq)
	}

	//
	if expected.Verbose != got.Verbose {
		foundMismatch = true
		log.Printf("Verbose: expected +%v, got %+v", expected.Verbose, got.Verbose)
	}
	if expected.Version != got.Version {
		foundMismatch = true
		log.Printf("Version: expected +%v, got %+v", expected.Version, got.Version)
	}

	if !foundMismatch {
		log.Println("findMismatch: mismatch NOT found")
	}
}
