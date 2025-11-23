package service

import (
	"context"

	"vector-database/db/document"
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
		InsertImage(ctx context.Context, input model.ImageInput) (model.ImageDocument, error)
		SearchImages(ctx context.Context, query model.ImageQuery) ([]model.ImageDocument, error)
	}

	AnalyzeService interface {
		TagMessage()
	}
)

type searchImp struct {
	store   document.Store
	encoder EncoderService
	dim     int
}

type encoderImp struct {
	Dimension int
}

func NewEncoder(dimension int) (EncoderService, error) {
	return &encoderImp{Dimension: dimension}, nil
}

func NewSearch(store document.Store, encoder EncoderService, dimension int) (SearchService, error) {
	return &searchImp{
		store:   store,
		encoder: encoder,
		dim:     dimension,
	}, nil
}
