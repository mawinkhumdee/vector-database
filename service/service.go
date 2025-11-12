package service

import (
	"context"

	"vector-database/db"
	"vector-database/model"
)

type (
	EncoderService interface {
		Encode(ctx context.Context, text string) ([]float32, error)
	}

	SearchService interface {
		IndexDocument(ctx context.Context, input model.DocumentInput) (model.Document, error)
		SearchByText(ctx context.Context, text string, limit int) ([]model.Document, error)
		SearchByVector(ctx context.Context, query model.VectorQuery) ([]model.Document, error)
	}
)

type searchImp struct {
	store   db.VectorDB
	encoder EncoderService
	dim     int
}

type encoderImp struct {
	Dimension int
}

func NewEncoder(dimension int) (EncoderService, error) {
	return &encoderImp{Dimension: dimension}, nil
}

func NewSearch(store db.VectorDB, encoder EncoderService, dimension int) (SearchService, error) {
	return &searchImp{
		store:   store,
		encoder: encoder,
		dim:     dimension,
	}, nil
}
