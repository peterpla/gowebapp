package main

import (
	"log"
	"os"
	"time"
)

func main() {
	for i := 0; i < 10; i++ {
		log.Println("Hello world from the wInitialRequest service!")
		time.Sleep(6 * time.Second)
	}
	os.Exit(0)
}
