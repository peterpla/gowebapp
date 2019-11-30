package config

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestLoadFlagsAndConfig(t *testing.T) {

	var cfg Config

	defaultResult := Config{
		AppName:     "gowebapp",
		ConfigFile:  "config.yaml",
		Description: "Describe gowebapp here",
		// Key Management Service for encrypted config
		EncryptedBucket: "elated-practice-224603-gowebapp-secret",
		KmsKey:          "config",
		KmsKeyRing:      "devkeyring",
		KmsLocation:     "us-west2",
		//
		Port:            8080,
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
		TaskDefaultSvc:         "default",
		TaskInitialRequestSvc:  "initial-request",
		TaskServiceDispatchSvc: "service-dispatch",
		//
		Verbose: false,
		Version: "0.1.0",
		Help:    false,
	}

	// guard against calling twice, which will trigger panic with "flag redefined"
	if cfg.Port == 0 { // uninitialized port value
		if err := LoadFlagsAndConfig(&cfg); err != nil {
			t.Fatalf("error from LoadFlagsAndConfig: %v", err)
		}
		// if !reflect.DeepEqual(defaultResult, cfg) {
		if !cmp.Equal(defaultResult, cfg) {
			t.Fatalf("expected %+v, got %+v", defaultResult, cfg)
		}
	}
}
