package httpinfo

import "net/http"

type Endpoint struct {
	Method      string
	Path        string
	Description string
}

const (
	DefaultAddr = ":8080"
	basePath    = "/api"
)

var (
	InsertMessageEndpoint = Endpoint{
		Method:      http.MethodPost,
		Path:        basePath + "/messages",
		Description: "Insert a message and store its embedding",
	}
	GetMessageEndpoint = Endpoint{
		Method:      http.MethodGet,
		Path:        basePath + "/messages",
		Description: "Retrieve messages via semantic search",
	}
)
