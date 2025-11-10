package model

import (
	"errors"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Document represents the MongoDB shape of a stored document.
type Document struct {
	ID        primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	Content   string                 `bson:"content" json:"content"`
	Embedding []float32              `bson:"embedding" json:"embedding"`
	Metadata  map[string]interface{} `bson:"metadata,omitempty" json:"metadata,omitempty"`
	Score     float64                `bson:"score,omitempty" json:"score,omitempty"`
}

// DocumentInput is the data provided by callers before an embedding is generated.
type DocumentInput struct {
	Content  string
	Metadata map[string]interface{}
}

// Validate ensures the document contains the minimum payload.
func (d DocumentInput) Validate() error {
	if d.Content == "" {
		return errors.New("content is required")
	}
	return nil
}
