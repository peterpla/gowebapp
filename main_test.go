package main

import (
	"log"
	"testing"
)

func TestLoadFlagsAndConfig(t *testing.T) {

	var Cfg config

	defaultResult := config{
		configFile: "",
		port:       8080,
		verbose:    false,
		help:       false,
	}

	if err := loadFlagsAndConfig(&Cfg); err != nil {
		t.Fatalf("error from loadFlagsAndConfig: %v", err)
	}
	log.Printf("config file: %q, port: %d, verbose: %t\n", Cfg.configFile, Cfg.port, Cfg.verbose)
	if Cfg.configFile != defaultResult.configFile ||
		Cfg.port != defaultResult.port ||
		Cfg.verbose != defaultResult.verbose ||
		Cfg.help != defaultResult.help {
		t.Fatalf("expected %v, got %v", defaultResult, Cfg)
	}
}
