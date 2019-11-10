package main

import (
	"os"
	"testing"
)

var srv *server

func TestMain(m *testing.M) {
	srv = NewServer()
	os.Exit(m.Run())
}
