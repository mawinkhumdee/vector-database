package service

import (
	"context"
	"fmt"

	"vector-database/db"
	"vector-database/model"
)

// EmbeddingService defines the operations exposed to the rest of the application.
type EmbeddingService interface {
	IndexDocument(ctx context.Context, input model.DocumentInput) (model.Document, error)
	SearchByText(ctx context.Context, text string, limit int) ([]model.Document, error)
	SearchByVector(ctx context.Context, query model.VectorQuery) ([]model.Document, error)
}

type embeddingService struct {
	store   db.VectorStore
	encoder VectorEncoder
	dim     int
}

// NewEmbeddingService wires the db layer with an encoder implementation.
func NewEmbeddingService(store db.VectorStore, encoder VectorEncoder, dimension int) (EmbeddingService, error) {
	if store == nil {
		return nil, fmt.Errorf("vector store cannot be nil")
	}
	if encoder == nil {
		return nil, fmt.Errorf("encoder cannot be nil")
	}
	if dimension <= 0 {
		return nil, fmt.Errorf("dimension must be positive")
	}

	return &embeddingService{
		store:   store,
		encoder: encoder,
		dim:     dimension,
	}, nil
}

func (s *embeddingService) IndexDocument(ctx context.Context, input model.DocumentInput) (model.Document, error) {
	if err := input.Validate(); err != nil {
		return model.Document{}, err
	}

	vector, err := s.encoder.Encode(ctx, input.Content)
	if err != nil {
		return model.Document{}, fmt.Errorf("encode content: %w", err)
	}

	return s.store.InsertDocument(ctx, input, vector)
}

func (s *embeddingService) SearchByText(ctx context.Context, text string, limit int) ([]model.Document, error) {
	if limit <= 0 {
		return nil, fmt.Errorf("limit must be positive")
	}

	vector, err := s.encoder.Encode(ctx, text)
	if err != nil {
		return nil, fmt.Errorf("encode query: %w", err)
	}

	query := model.VectorQuery{
		QueryVector: vector,
		Limit:       limit,
	}

	return s.store.SimilaritySearch(ctx, query)
}

func (s *embeddingService) SearchByVector(ctx context.Context, query model.VectorQuery) ([]model.Document, error) {
	if err := query.Validate(s.dim); err != nil {
		return nil, err
	}
	return s.store.SimilaritySearch(ctx, query)
}
