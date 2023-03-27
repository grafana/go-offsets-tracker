# Go offsets tracker

This tool allows generating a JSON description of the byte offsets of a set of field structs,
for each version or Go or third-party library.

It is useful to inspect some known fields in executable files that do not embed Debug information.

This is a standalone, modified version of the [Open Telemetry offsets tracker tool](https://github.com/open-telemetry/opentelemetry-go-instrumentation),
and both are licensed under [Apache Software License 2.0](./LICENSE).

## How to install

```
go install github.com/grafana/go-offsets-tracker/cmd/go-offsets-tracker@latest
```

## How to generate offsets

Specify the library/struct/field that you want to track, for a version range in an input JSON
file. Check [examples/input_file.json](./examples/input_file.json) to understand the schema:

```
go-offsets-tracker -i examples/input_file.json examples/offsets.json
```

If the output file ([examples/offsets.json](./examples/offsets.json)) in the above example)
already exists, the program will reuse these known offsets as a cache, to not have to retrieve
the information again from the internet.

If you need to regenerate completely the output file, remove it or use an output file that
does not exist.

## How to read offsets from a program

Use `offsets.Open` or `offsets.Read` to load an (`offsets.Track`).

Use the `Find` method of the `offsets.Track` to get the offsets, given the struct, field and version names:

```go
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
```

Output:

```
offset for google.golang.org/grpc/internal/transport.Stream.method (1.16.7): 64
```
