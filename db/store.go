package db

import (
	"context"

	"vector-database/config"
	"vector-database/model"
)

// VectorStore describes the behaviour the service layer expects from any database implementation.
type VectorStore interface {
	InsertDocument(ctx context.Context, doc model.DocumentInput, embedding []float32) (model.Document, error)
	SimilaritySearch(ctx context.Context, query model.VectorQuery) ([]model.Document, error)
	Close(ctx context.Context) error
}

// InitMongoVectorStore wires up the Mongo implementation behind the VectorStore interface.
func InitMongoVectorStore(ctx context.Context, cfg config.MongoConfig) (VectorStore, error) {
	return NewMongoVectorStore(ctx, cfg)
}
