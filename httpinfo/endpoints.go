package httpinfo

import "net/http"

// Endpoint describes an HTTP resource exposed by the server.
type Endpoint struct {
	Method      string
	Path        string
	Description string
}

const (
	// DefaultAddr is the default host:port for the HTTP server.
	DefaultAddr = ":8080"
	basePath    = "/api"
)

var (
	// InsertMessageEndpoint accepts POST payloads for indexing new messages.
	InsertMessageEndpoint = Endpoint{
		Method:      http.MethodPost,
		Path:        basePath + "/messages",
		Description: "Insert a message and store its embedding",
	}
	// GetMessageEndpoint accepts GET requests to fetch the most relevant messages.
	GetMessageEndpoint = Endpoint{
		Method:      http.MethodGet,
		Path:        basePath + "/messages",
		Description: "Retrieve messages via semantic search",
	}
)
