package main

import (
	"reflect"
	"testing"
)

func TestLoadFlagsAndConfig(t *testing.T) {

	var cfg config

	defaultResult := config{
		appName:         "gowebapp",
		configFile:      "config.yaml",
		description:     "Describe gowebapp here",
		encryptedBucket: "elated-practice-224603-gowebapp-secret",
		kmsKey:          "config",
		kmsKeyRing:      "devkeyring",
		kmsLocation:     "us-west2",
		port:            8080,
		projectID:       "elated-practice-224603",
		storageLocation: "us-west2",
		verbose:         false,
		version:         "0.1.0",
		help:            false,
	}

	if err := loadFlagsAndConfig(&cfg); err != nil {
		t.Fatalf("error from loadFlagsAndConfig: %v", err)
	}
	if !reflect.DeepEqual(defaultResult, cfg) {
		t.Fatalf("expected %v, got %v", defaultResult, cfg)
	}
}
