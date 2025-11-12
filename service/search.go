package service

import (
	"context"
	"fmt"
	"vector-database/model"
)

func (s *searchImp) IndexDocument(ctx context.Context, input model.DocumentInput) (model.Document, error) {
	if err := input.Validate(); err != nil {
		return model.Document{}, err
	}

	vector, err := s.encoder.Encode(ctx, input.Content)
	if err != nil {
		return model.Document{}, fmt.Errorf("encode content: %w", err)
	}

	return s.store.InsertDocument(ctx, input, vector)
}

func (s *searchImp) SearchByText(ctx context.Context, text string, limit int) ([]model.Document, error) {
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

func (s *searchImp) SearchByVector(ctx context.Context, query model.VectorQuery) ([]model.Document, error) {
	if err := query.Validate(s.dim); err != nil {
		return nil, err
	}
	return s.store.SimilaritySearch(ctx, query)
}
