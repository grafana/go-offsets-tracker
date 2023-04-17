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
	structFields := [][]string{
		{"google.golang.org/grpc/internal/transport.Stream", "method", "1.16.7"},
		{"google.golang.org/genproto/googleapis/rpc/status.Status", "Code", "v0.1.0"},
	}
	for _, sf := range structFields {
		structName, fieldName, version := sf[0], sf[1], sf[2]
		off, ok := track.Find(structName, fieldName, version)
		if !ok {
			log.Fatal("offsets not found!", structName, fieldName, version)
		}
		log.Printf("offset for %s.%s (%s): %d", structName, fieldName, version, off)
	}
}
