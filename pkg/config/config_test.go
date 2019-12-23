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
		QueueName:       "InitialRequest",
		Router:          nil,
		ServiceName:     "default",
		NextServiceName: "initial-request",
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
	if cfg.AppName == "" {
		// uninitialized
		if err := GetConfig(&cfg, "TaskDefault"); err != nil {
			t.Fatalf("error from GetConfig: %v", err)
		}

		// CHEAT: nil-out the actual .Adder, hard to compare addresses
		cfg.Adder = nil

		if !cmp.Equal(defaultResult, cfg) {
			findMismatch(t, defaultResult, cfg)
		}
	}
}

func findMismatch(t *testing.T, expected Config, got Config) {

	var foundMismatch = false

	if expected.Adder != got.Adder {
		foundMismatch = true
		t.Errorf("Adder: expected %v, got %v", expected.Adder, got.Adder)
	}
	if expected.AppName != got.AppName {
		foundMismatch = true
		t.Errorf("AppName: expected %q, got %q", expected.AppName, got.AppName)
	}
	if expected.ConfigFile != got.ConfigFile {
		foundMismatch = true
		t.Errorf("ConfigFile: expected %q, got %q", expected.ConfigFile, got.ConfigFile)
	}
	if expected.Description != got.Description {
		foundMismatch = true
		t.Errorf("Description: expected %q, got %q", expected.Description, got.Description)
	}
	if expected.IsGAE != got.IsGAE {
		foundMismatch = true
		t.Errorf("IsGAE: expected %t, got %v", expected.IsGAE, got.IsGAE)
	}
	if expected.QueueName != got.QueueName {
		foundMismatch = true
		t.Errorf("QueueName: expected %q, got %q", expected.QueueName, got.QueueName)
	}
	if expected.Router != got.Router {
		foundMismatch = true
		t.Errorf("Router: expected %v, got %v", expected.Router, got.Router)
	}
	if expected.ServiceName != got.ServiceName {
		foundMismatch = true
		t.Errorf("ServiceName: expected %q, got %q", expected.ServiceName, got.ServiceName)
	}
	if expected.NextServiceName != got.NextServiceName {
		foundMismatch = true
		t.Errorf("NextServiceName: expected %q, got %q", expected.NextServiceName, got.NextServiceName)
	}
	if expected.StorageType != got.StorageType {
		foundMismatch = true
		t.Errorf("StorageType: expected %v, got %v", expected.StorageType, got.StorageType)
	}
	if expected.EncryptedBucket != got.EncryptedBucket {
		foundMismatch = true
		t.Errorf("EncryptedBucket: expected %q, got %q", expected.EncryptedBucket, got.EncryptedBucket)
	}

	// Key Management Service for encrypted config
	if expected.KmsKey != got.KmsKey {
		foundMismatch = true
		t.Errorf("KmsKey: expected %q, got %q", expected.KmsKey, got.KmsKey)
	}
	if expected.KmsKeyRing != got.KmsKeyRing {
		foundMismatch = true
		t.Errorf("KmsKeyRing: expected %q, got %q", expected.KmsKeyRing, got.KmsKeyRing)
	}
	if expected.KmsLocation != got.KmsLocation {
		foundMismatch = true
		t.Errorf("KmsLocation: expected %q, got %q", expected.KmsLocation, got.KmsLocation)
	}

	//
	if expected.ProjectID != got.ProjectID {
		foundMismatch = true
		t.Errorf("ProjectID: expected %q, got %q", expected.ProjectID, got.ProjectID)
	}
	if expected.StorageLocation != got.StorageLocation {
		foundMismatch = true
		t.Errorf("StorageLocation: expected %q, got %q", expected.StorageLocation, got.StorageLocation)
	}
	if expected.TasksLocation != got.TasksLocation {
		foundMismatch = true
		t.Errorf("TasksLocation: expected %q, got %q", expected.TasksLocation, got.TasksLocation)
	}

	// port used by each service: Task*Port
	if expected.TaskDefaultPort != got.TaskDefaultPort {
		foundMismatch = true
		t.Errorf("TaskDefaultPort: expected %q, got %q", expected.TaskDefaultPort, got.TaskDefaultPort)
	}
	if expected.TaskInitialRequestPort != got.TaskInitialRequestPort {
		foundMismatch = true
		t.Errorf("TaskInitialRequestPort: expected %q, got %q", expected.TaskInitialRequestPort, got.TaskInitialRequestPort)
	}
	if expected.TaskServiceDispatchPort != got.TaskServiceDispatchPort {
		foundMismatch = true
		t.Errorf("TaskServiceDispatchPort: expected %q, got %q", expected.TaskServiceDispatchPort, got.TaskServiceDispatchPort)
	}
	if expected.TaskTranscriptionGCPPort != got.TaskTranscriptionGCPPort {
		foundMismatch = true
		t.Errorf("TaskTranscriptionGCPPort: expected %q, got %q", expected.TaskTranscriptionGCPPort, got.TaskTranscriptionGCPPort)
	}
	if expected.TaskTranscriptionCompletePort != got.TaskTranscriptionCompletePort {
		foundMismatch = true
		t.Errorf("TaskTranscriptionCompletePort: expected %q, got %q", expected.TaskTranscriptionCompletePort, got.TaskTranscriptionCompletePort)
	}
	if expected.TaskTranscriptQAPort != got.TaskTranscriptQAPort {
		foundMismatch = true
		t.Errorf("TaskTranscriptQAPort: expected %q, got %q", expected.TaskTranscriptQAPort, got.TaskTranscriptQAPort)
	}
	if expected.TaskTranscriptQACompletePort != got.TaskTranscriptQACompletePort {
		foundMismatch = true
		t.Errorf("TaskTranscriptQACompletePort: expected %q, got %q", expected.TaskTranscriptQACompletePort, got.TaskTranscriptQACompletePort)
	}
	if expected.TaskTaggingPort != got.TaskTaggingPort {
		foundMismatch = true
		t.Errorf("TaskTaggingPort: expected %q, got %q", expected.TaskTaggingPort, got.TaskTaggingPort)
	}
	if expected.TaskTaggingCompletePort != got.TaskTaggingCompletePort {
		foundMismatch = true
		t.Errorf("TaskTaggingCompletePort: expected %q, got %q", expected.TaskTaggingCompletePort, got.TaskTaggingCompletePort)
	}
	if expected.TaskTaggingQAPort != got.TaskTaggingQAPort {
		foundMismatch = true
		t.Errorf("TaskTaggingQAPort: expected %q, got %q", expected.TaskTaggingQAPort, got.TaskTaggingQAPort)
	}
	if expected.TaskTaggingQACompletePort != got.TaskTaggingQACompletePort {
		foundMismatch = true
		t.Errorf("TaskTaggingQACompletePort: expected %q, got %q", expected.TaskTaggingQACompletePort, got.TaskTaggingQACompletePort)
	}
	if expected.TaskCompletionProcessingPort != got.TaskCompletionProcessingPort {
		foundMismatch = true
		t.Errorf("TaskCompletionProcessingPort: expected %q, got %q", expected.TaskCompletionProcessingPort, got.TaskCompletionProcessingPort)
	}

	// queue name used by each service: Task*WriteToQ
	if expected.TaskDefaultWriteToQ != got.TaskDefaultWriteToQ {
		foundMismatch = true
		t.Errorf("TaskDefaultWriteToQ: expected %q, got %q", expected.TaskDefaultWriteToQ, got.TaskDefaultWriteToQ)
	}
	if expected.TaskInitialRequestWriteToQ != got.TaskInitialRequestWriteToQ {
		foundMismatch = true
		t.Errorf("TaskInitialRequestWriteToQ: expected %q, got %q", expected.TaskInitialRequestWriteToQ, got.TaskInitialRequestWriteToQ)
	}
	if expected.TaskServiceDispatchWriteToQ != got.TaskServiceDispatchWriteToQ {
		foundMismatch = true
		t.Errorf("TaskServiceDispatchWriteToQ: expected %q, got %q", expected.TaskServiceDispatchWriteToQ, got.TaskServiceDispatchWriteToQ)
	}
	if expected.TaskTranscriptionGCPWriteToQ != got.TaskTranscriptionGCPWriteToQ {
		foundMismatch = true
		t.Errorf("TaskTranscriptionGCPWriteToQ: expected %q, got %q", expected.TaskTranscriptionGCPWriteToQ, got.TaskTranscriptionGCPWriteToQ)
	}
	if expected.TaskTranscriptionCompleteWriteToQ != got.TaskTranscriptionCompleteWriteToQ {
		foundMismatch = true
		t.Errorf("TaskTranscriptionCompleteWriteToQ: expected %q, got %q", expected.TaskTranscriptionCompleteWriteToQ, got.TaskTranscriptionCompleteWriteToQ)
	}
	if expected.TaskTranscriptQAWriteToQ != got.TaskTranscriptQAWriteToQ {
		foundMismatch = true
		t.Errorf("TaskTranscriptQAWriteToQ: expected %q, got %q", expected.TaskTranscriptQAWriteToQ, got.TaskTranscriptQAWriteToQ)
	}
	if expected.TaskTranscriptQACompleteWriteToQ != got.TaskTranscriptQACompleteWriteToQ {
		foundMismatch = true
		t.Errorf("TaskTranscriptQACompleteWriteToQ: expected %q, got %q", expected.TaskTranscriptQACompleteWriteToQ, got.TaskTranscriptQACompleteWriteToQ)
	}
	if expected.TaskTaggingWriteToQ != got.TaskTaggingWriteToQ {
		foundMismatch = true
		t.Errorf("TaskTaggingWriteToQ: expected %q, got %q", expected.TaskTaggingWriteToQ, got.TaskTaggingWriteToQ)
	}
	if expected.TaskTaggingCompleteWriteToQ != got.TaskTaggingCompleteWriteToQ {
		foundMismatch = true
		t.Errorf("TaskTaggingCompleteWriteToQ: expected %q, got %q", expected.TaskTaggingCompleteWriteToQ, got.TaskTaggingCompleteWriteToQ)
	}
	if expected.TaskTaggingQAWriteToQ != got.TaskTaggingQAWriteToQ {
		foundMismatch = true
		t.Errorf("TaskTaggingQAWriteToQ: expected %q, got %q", expected.TaskTaggingQAWriteToQ, got.TaskTaggingQAWriteToQ)
	}
	if expected.TaskTaggingQACompleteWriteToQ != got.TaskTaggingQACompleteWriteToQ {
		foundMismatch = true
		t.Errorf("TaskTaggingQACompleteWriteToQ: expected %q, got %q", expected.TaskTaggingQACompleteWriteToQ, got.TaskTaggingQACompleteWriteToQ)
	}
	if expected.TaskCompletionProcessingWriteToQ != got.TaskCompletionProcessingWriteToQ {
		foundMismatch = true
		t.Errorf("TaskCompletionProcessingWriteToQ: expected %q, got %q", expected.TaskCompletionProcessingWriteToQ, got.TaskCompletionProcessingWriteToQ)
	}

	// service name of each service: Task*SvcName
	if expected.TaskDefaultSvcName != got.TaskDefaultSvcName {
		foundMismatch = true
		t.Errorf("TaskDefaultSvcName: expected %q, got %q", expected.TaskDefaultSvcName, got.TaskDefaultSvcName)
	}
	if expected.TaskInitialRequestSvcName != got.TaskInitialRequestSvcName {
		foundMismatch = true
		t.Errorf("TaskInitialRequestSvcName: expected %q, got %q", expected.TaskInitialRequestSvcName, got.TaskInitialRequestSvcName)
	}
	if expected.TaskServiceDispatchSvcName != got.TaskServiceDispatchSvcName {
		foundMismatch = true
		t.Errorf("TaskServiceDispatchSvcName: expected %q, got %q", expected.TaskServiceDispatchSvcName, got.TaskServiceDispatchSvcName)
	}
	if expected.TaskTranscriptionGCPSvcName != got.TaskTranscriptionGCPSvcName {
		foundMismatch = true
		t.Errorf("TaskTranscriptionGCPSvcName: expected %q, got %q", expected.TaskTranscriptionGCPSvcName, got.TaskTranscriptionGCPSvcName)
	}
	if expected.TaskTranscriptionCompleteSvcName != got.TaskTranscriptionCompleteSvcName {
		foundMismatch = true
		t.Errorf("TaskTranscriptionCompleteSvcName: expected %q, got %q", expected.TaskTranscriptionCompleteSvcName, got.TaskTranscriptionCompleteSvcName)
	}
	if expected.TaskTranscriptQASvcName != got.TaskTranscriptQASvcName {
		foundMismatch = true
		t.Errorf("TaskTranscriptQASvcName: expected %q, got %q", expected.TaskTranscriptQASvcName, got.TaskTranscriptQASvcName)
	}
	if expected.TaskTranscriptQACompleteSvcName != got.TaskTranscriptQACompleteSvcName {
		foundMismatch = true
		t.Errorf("TaskTranscriptQACompleteSvcName: expected %q, got %q", expected.TaskTranscriptQACompleteSvcName, got.TaskTranscriptQACompleteSvcName)
	}
	if expected.TaskTaggingSvcName != got.TaskTaggingSvcName {
		foundMismatch = true
		t.Errorf("TaskTaggingSvcName: expected %q, got %q", expected.TaskTaggingSvcName, got.TaskTaggingSvcName)
	}
	if expected.TaskTaggingCompleteSvcName != got.TaskTaggingCompleteSvcName {
		foundMismatch = true
		t.Errorf("TaskTaggingCompleteSvcName: expected %q, got %q", expected.TaskTaggingCompleteSvcName, got.TaskTaggingCompleteSvcName)
	}
	if expected.TaskTaggingQASvcName != got.TaskTaggingQASvcName {
		foundMismatch = true
		t.Errorf("TaskTaggingQASvcName: expected %q, got %q", expected.TaskTaggingQASvcName, got.TaskTaggingQASvcName)
	}
	if expected.TaskTaggingQACompleteSvcName != got.TaskTaggingQACompleteSvcName {
		foundMismatch = true
		t.Errorf("TaskTaggingQACompleteSvcName: expected %q, got %q", expected.TaskTaggingQACompleteSvcName, got.TaskTaggingQACompleteSvcName)
	}
	if expected.TaskCompletionProcessingSvcName != got.TaskCompletionProcessingSvcName {
		foundMismatch = true
		t.Errorf("TaskCompletionProcessingSvcName: expected %q, got %q", expected.TaskCompletionProcessingSvcName, got.TaskCompletionProcessingSvcName)
	}

	// next service in the chain to handle requests: Task*NextSvcToHandleReq
	if expected.TaskDefaultNextSvcToHandleReq != got.TaskDefaultNextSvcToHandleReq {
		foundMismatch = true
		t.Errorf("TaskDefaultNextSvcToHandleReq: expected %q, got %q", expected.TaskDefaultNextSvcToHandleReq, got.TaskDefaultNextSvcToHandleReq)
	}
	if expected.TaskInitialRequestNextSvcToHandleReq != got.TaskInitialRequestNextSvcToHandleReq {
		foundMismatch = true
		t.Errorf("TaskInitialRequestNextSvcToHandleReq: expected %q, got %q", expected.TaskInitialRequestNextSvcToHandleReq, got.TaskInitialRequestNextSvcToHandleReq)
	}
	if expected.TaskServiceDispatchNextSvcToHandleReq != got.TaskServiceDispatchNextSvcToHandleReq {
		foundMismatch = true
		t.Errorf("TaskServiceDispatchNextSvcToHandleReq: expected %q, got %q", expected.TaskServiceDispatchNextSvcToHandleReq, got.TaskServiceDispatchNextSvcToHandleReq)
	}
	if expected.TaskTranscriptionGCPNextSvcToHandleReq != got.TaskTranscriptionGCPNextSvcToHandleReq {
		foundMismatch = true
		t.Errorf("TaskTranscriptionGCPNextSvcToHandleReq: expected %q, got %q", expected.TaskTranscriptionGCPNextSvcToHandleReq, got.TaskTranscriptionGCPNextSvcToHandleReq)
	}
	if expected.TaskTranscriptionCompleteNextSvcToHandleReq != got.TaskTranscriptionCompleteNextSvcToHandleReq {
		foundMismatch = true
		t.Errorf("TaskTranscriptionCompleteNextSvcToHandleReq: expected %q, got %q", expected.TaskTranscriptionCompleteNextSvcToHandleReq, got.TaskTranscriptionCompleteNextSvcToHandleReq)
	}
	if expected.TaskTranscriptQANextSvcToHandleReq != got.TaskTranscriptQANextSvcToHandleReq {
		foundMismatch = true
		t.Errorf("TaskTranscriptQANextSvcToHandleReq: expected %q, got %q", expected.TaskTranscriptQANextSvcToHandleReq, got.TaskTranscriptQANextSvcToHandleReq)
	}
	if expected.TaskTranscriptQACompleteNextSvcToHandleReq != got.TaskTranscriptQACompleteNextSvcToHandleReq {
		foundMismatch = true
		t.Errorf("TaskTranscriptQACompleteNextSvcToHandleReq: expected %q, got %q", expected.TaskTranscriptQACompleteNextSvcToHandleReq, got.TaskTranscriptQACompleteNextSvcToHandleReq)
	}
	if expected.TaskTaggingNextSvcToHandleReq != got.TaskTaggingNextSvcToHandleReq {
		foundMismatch = true
		t.Errorf("TaskTaggingNextSvcToHandleReq: expected %q, got %q", expected.TaskTaggingNextSvcToHandleReq, got.TaskTaggingNextSvcToHandleReq)
	}
	if expected.TaskTaggingCompleteNextSvcToHandleReq != got.TaskTaggingCompleteNextSvcToHandleReq {
		foundMismatch = true
		t.Errorf("TaskTaggingCompleteNextSvcToHandleReq: expected %q, got %q", expected.TaskTaggingCompleteNextSvcToHandleReq, got.TaskTaggingCompleteNextSvcToHandleReq)
	}
	if expected.TaskTaggingQANextSvcToHandleReq != got.TaskTaggingQANextSvcToHandleReq {
		foundMismatch = true
		t.Errorf("TaskTaggingQANextSvcToHandleReq: expected %q, got %q", expected.TaskTaggingQANextSvcToHandleReq, got.TaskTaggingQANextSvcToHandleReq)
	}
	if expected.TaskTaggingQACompleteNextSvcToHandleReq != got.TaskTaggingQACompleteNextSvcToHandleReq {
		foundMismatch = true
		t.Errorf("TaskTaggingQACompleteNextSvcToHandleReq: expected %q, got %q", expected.TaskTaggingQACompleteNextSvcToHandleReq, got.TaskTaggingQACompleteNextSvcToHandleReq)
	}
	if expected.TaskCompletionProcessingNextSvcToHandleReq != got.TaskCompletionProcessingNextSvcToHandleReq {
		foundMismatch = true
		t.Errorf("TaskCompletionProcessingNextSvcToHandleReq: expected %q, got %q", expected.TaskCompletionProcessingNextSvcToHandleReq, got.TaskCompletionProcessingNextSvcToHandleReq)
	}

	//
	if expected.Verbose != got.Verbose {
		foundMismatch = true
		t.Errorf("Verbose: expected %t, got %v", expected.Verbose, got.Verbose)
	}
	if expected.Version != got.Version {
		foundMismatch = true
		t.Errorf("Version: expected %q, got %q", expected.Version, got.Version)
	}

	if !foundMismatch {
		log.Println("findMismatch: mismatch NOT found")
	}
}
