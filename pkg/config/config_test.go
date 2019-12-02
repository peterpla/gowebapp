package config

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestGetConfig(t *testing.T) {

	var cfg Config

	defaultResult := Config{
		Adder:           nil,
		AppName:         "gowebapp",
		ConfigFile:      "config.yaml",
		Description:     "Describe gowebapp here",
		IsGAE:           false,
		QueueName:       "",
		Router:          nil,
		ServiceName:     "",
		NextServiceName: "",
		StorageType:     Memory,
		// Key Management Service for encrypted config
		EncryptedBucket: "elated-practice-224603-gowebapp-secret",
		KmsKey:          "config",
		KmsKeyRing:      "devkeyring",
		KmsLocation:     "us-west2",
		//
		ProjectID:       "elated-practice-224603",
		StorageLocation: "us-west2",
		TasksLocation:   "us-west2",
		// port number used by each service
		TaskDefaultPort:         "8080",
		TaskInitialRequestPort:  "8081",
		TaskServiceDispatchPort: "8082",
		// queue name used by each services
		TaskDefaultWriteToQ:         "InitialRequest",
		TaskInitialRequestWriteToQ:  "ServiceDispatch",
		TaskServiceDispatchWriteToQ: "TranscriptionGCP",
		// service name of each service
		TaskDefaultSvcName:         "default",
		TaskInitialRequestSvcName:  "initial-request",
		TaskServiceDispatchSvcName: "service-dispatch",
		// next service in the chain to handle requests
		TaskDefaultNextSvcToHandleReq:         "initial-request",
		TaskInitialRequestNextSvcToHandleReq:  "service-dispatch",
		TaskServiceDispatchNextSvcToHandleReq: "transcription-gcp",
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
			t.Fatalf("expected %+v, got %+v", defaultResult, cfg)
		}
	}
}
