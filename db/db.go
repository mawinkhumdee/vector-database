package db

import (
	"context"
	"fmt"

	"vector-database/config"
	"vector-database/model"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type VectorDB interface {
	InsertDocument(ctx context.Context, doc model.DocumentInput, embedding []float32) (model.Document, error)
	SimilaritySearch(ctx context.Context, query model.VectorQuery) ([]model.Document, error)
	Close(ctx context.Context) error
}

type vectorDB struct {
	client     *mongo.Client
	collection *mongo.Collection
	cfg        config.MongoDB
}

func NewVectorDB(ctx context.Context, cfg config.MongoDB) (VectorDB, error) {
	clientOpts := options.Client().ApplyURI(cfg.URI)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, fmt.Errorf("connect mongo: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("ping mongo: %w", err)
	}

	collection := client.Database(cfg.Database).Collection(cfg.Collection)

	if err := ensureVectorIndex(ctx, collection, cfg); err != nil {
		return nil, err
	}

	return &vectorDB{
		client:     client,
		collection: collection,
		cfg:        cfg,
	}, nil
}

func (m *vectorDB) Close(ctx context.Context) error {
	return m.client.Disconnect(ctx)
}
