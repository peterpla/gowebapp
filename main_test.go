package main

import (
	"log"
	"testing"
)

func TestLoadFlagsAndConfig(t *testing.T) {

	var cfg config

	defaultResult := config{
		configFile: "",
		port:       8080,
		verbose:    false,
		help:       false,
	}

	if err := loadFlagsAndConfig(&cfg); err != nil {
		t.Fatalf("error from loadFlagsAndConfig: %v", err)
	}
	log.Printf("config file: %q, port: %d, verbose: %t\n", cfg.configFile, cfg.port, cfg.verbose)
	if cfg.configFile != defaultResult.configFile ||
		cfg.port != defaultResult.port ||
		cfg.verbose != defaultResult.verbose ||
		cfg.help != defaultResult.help {
		t.Fatalf("expected %v, got %v", defaultResult, cfg)
	}
}
