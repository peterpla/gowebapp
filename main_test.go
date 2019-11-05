package main

import (
	"testing"
)

func TestLoadFlagsAndConfig(t *testing.T) {

	var Cfg config

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

	if err := loadFlagsAndConfig(&Cfg); err != nil {
		t.Fatalf("error from loadFlagsAndConfig: %v", err)
	}
<<<<<<< HEAD
	// log.Printf("config file: %q, port: %d, verbose: %t\n", Cfg.configFile, Cfg.port, Cfg.verbose)
>>>>>>> a888b49ed2ba4a5f84d77e5fbaed994151e29dc4
	if Cfg.configFile != defaultResult.configFile ||
		Cfg.port != defaultResult.port ||
		Cfg.verbose != defaultResult.verbose ||
		Cfg.help != defaultResult.help {
		t.Fatalf("expected %v, got %v", defaultResult, Cfg)
	}
}
