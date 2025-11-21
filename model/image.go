package model

import (
	"errors"
	"strings"
)

// ImageInput contains the data required to index an image.
type ImageInput struct {
	Description string
	ImageData   []byte
	Metadata    map[string]interface{}
}

// Validate ensures the insert payload has mandatory fields.
func (i ImageInput) Validate() error {
	if strings.TrimSpace(i.Description) == "" {
		return errors.New("description is required")
	}
	if len(i.ImageData) == 0 {
		return errors.New("image is required")
	}
	return nil
}

// ImageQuery describes an image similarity search and optional textual hint.
type ImageQuery struct {
	ImageData   []byte
	Description string
	Limit       int
}

// Validate ensures the query can be executed safely.
func (q ImageQuery) Validate() error {
	if len(q.ImageData) == 0 {
		return errors.New("image is required")
	}
	if q.Limit < 0 {
		return errors.New("limit must be a positive integer")
	}
	return nil
}

// ImageDocument is the response payload returned by image APIs.
type ImageDocument struct {
	ID          string                 `json:"id,omitempty"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Score       float64                `json:"score,omitempty"`
}
