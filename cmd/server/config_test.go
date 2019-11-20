package main

import (
	"reflect"
	"testing"
)

func TestLoadFlagsAndConfig(t *testing.T) {

	var cfg config

	defaultResult := config{
		appName:                  "gowebapp",
		configFile:               "config.yaml",
		description:              "Describe gowebapp here",
		encryptedBucket:          "elated-practice-224603-gowebapp-secret",
		kmsKey:                   "config",
		kmsKeyRing:               "devkeyring",
		kmsLocation:              "us-west2",
		Port:                     8080,
		projectID:                "elated-practice-224603",
		storageLocation:          "us-west2",
		tasksLocation:            "us-west2",
		tasksQRequests:           "wInitialRequest",
		tasksServiceRequestsPort: "8081",
		verbose:                  false,
		version:                  "0.1.0",
		help:                     false,
	}

	// guard against calling twice, which will trigger panic with "flag redefined"
	if cfg.Port == 0 { // uninitialized port value
		if err := loadFlagsAndConfig(&cfg); err != nil {
			t.Fatalf("error from loadFlagsAndConfig: %v", err)
		}
		if !reflect.DeepEqual(defaultResult, cfg) {
			t.Fatalf("expected %v, got %v", defaultResult, cfg)
		}
	}
}
