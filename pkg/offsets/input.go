package offsets

const GoStdLib = "go"

// InputLibs key: name of the library, or "go" for the Go standard library
type InputLibs map[string]LibQuery

type LibQuery struct {
	// Versions constraint. E.g. ">= 1.12" will only download versions
	// larger or equal to 1.12
	Versions string `json:"versions"`

	// Fields key: qualified name of the struct.
	// Examples: net/http.Request, google.golang.org/grpc/internal/transport.Stream
	// Value: list of case-sensitive name of the fields whose offsets we want to retrieve
	Fields map[string][]string
}
