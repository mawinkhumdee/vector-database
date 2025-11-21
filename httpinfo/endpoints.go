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
	InsertImageEndpoint = Endpoint{
		Method:      http.MethodPost,
		Path:        basePath + "/images",
		Description: "Insert an image with a textual description and store its embedding",
	}
	SearchImageEndpoint = Endpoint{
		Method:      http.MethodPost,
		Path:        basePath + "/images/search",
		Description: "Find images whose embeddings are similar to the provided payload",
	}
)
