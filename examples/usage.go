package main

import (
	"log"

	"github.com/grafana/go-offsets-tracker/pkg/offsets"
)

func main() {
	track, err := offsets.Open("./examples/offsets.json")
	if err != nil {
		log.Fatal("opening file", err)
	}
	structName := "google.golang.org/grpc/internal/transport.Stream"
	fieldName := "method"
	version := "1.16.7"
	off, ok := track.Find(structName, fieldName, version)
	if !ok {
		log.Fatal("offsets not found!", structName, fieldName, version)
	}
	log.Printf("offset for %s.%s (%s): %d", structName, fieldName, version, off)
}
