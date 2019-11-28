package config

import (
	"reflect"
	"testing"
)

func TestLoadFlagsAndConfig(t *testing.T) {

	var cfg Config

	defaultResult := Config{
		AppName:                  "gowebapp",
		ConfigFile:               "config.yaml",
		Description:              "Describe gowebapp here",
		EncryptedBucket:          "elated-practice-224603-gowebapp-secret",
		KmsKey:                   "config",
		KmsKeyRing:               "devkeyring",
		KmsLocation:              "us-west2",
		Port:                     8080,
		ProjectID:                "elated-practice-224603",
		StorageLocation:          "us-west2",
		TasksLocation:            "us-west2",
		TasksQRequests:           "wInitialRequest",
		TasksServiceRequestsPort: "8081",
		Verbose:                  false,
		Version:                  "0.1.0",
		Help:                     false,
	}

	// guard against calling twice, which will trigger panic with "flag redefined"
	if cfg.Port == 0 { // uninitialized port value
		if err := LoadFlagsAndConfig(&cfg); err != nil {
			t.Fatalf("error from LoadFlagsAndConfig: %v", err)
		}
		if !reflect.DeepEqual(defaultResult, cfg) {
			t.Fatalf("expected %v, got %v", defaultResult, cfg)
		}
	}
}
