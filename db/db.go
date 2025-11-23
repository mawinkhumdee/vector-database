package db

import (
	"context"
	"fmt"

	"vector-database/config"
	"vector-database/db/document"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Database struct {
	client    *mongo.Client
	Documents document.Store
}

func New(ctx context.Context, cfg config.MongoDB) (*Database, error) {
	clientOpts := options.Client().ApplyURI(cfg.URI)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, fmt.Errorf("connect mongo: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("ping mongo: %w", err)
	}

	db := client.Database(cfg.Database)
	docCollection := db.Collection(cfg.Collection.Document)
	if err := document.EnsureIndexes(ctx, docCollection, cfg); err != nil {
		return nil, err
	}

	return &Database{
		client:    client,
		Documents: document.NewStore(docCollection, cfg),
	}, nil
}

// Close releases the Mongo client resources.
func (d *Database) Close(ctx context.Context) error {
	return d.client.Disconnect(ctx)
}
