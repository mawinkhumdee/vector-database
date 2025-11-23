package document

import (
	"context"
	"errors"
	"fmt"

	"vector-database/config"
	"vector-database/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Store defines CRUD and search operations over the documents collection.
type Store interface {
	InsertDocument(ctx context.Context, doc model.DocumentInput, embedding []float32) (model.Document, error)
	SimilaritySearch(ctx context.Context, query model.VectorQuery) ([]model.Document, error)
}

type mongoStore struct {
	collection *mongo.Collection
	cfg        config.MongoDB
}

// NewStore wires the Mongo collection and config into a Store implementation.
func NewStore(collection *mongo.Collection, cfg config.MongoDB) Store {
	return &mongoStore{
		collection: collection,
		cfg:        cfg,
	}
}

func (m *mongoStore) InsertDocument(ctx context.Context, doc model.DocumentInput, embedding []float32) (model.Document, error) {
	if err := doc.Validate(); err != nil {
		return model.Document{}, err
	}
	if len(embedding) != m.cfg.EmbeddingDimension {
		return model.Document{}, fmt.Errorf("embedding dimension mismatch: expected %d, got %d", m.cfg.EmbeddingDimension, len(embedding))
	}

	payload := bson.M{
		"content":   doc.Content,
		"embedding": float32ToFloat64(embedding),
	}
	if len(doc.Metadata) > 0 {
		payload["metadata"] = doc.Metadata
	}

	res, err := m.collection.InsertOne(ctx, payload)
	if err != nil {
		return model.Document{}, fmt.Errorf("insert document: %w", err)
	}

	id, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return model.Document{}, errors.New("failed to convert inserted id to ObjectID")
	}

	return model.Document{
		ID:        id,
		Content:   doc.Content,
		Embedding: embedding,
		Metadata:  doc.Metadata,
	}, nil
}

func (m *mongoStore) SimilaritySearch(ctx context.Context, query model.VectorQuery) ([]model.Document, error) {
	if err := query.Validate(m.cfg.EmbeddingDimension); err != nil {
		return nil, err
	}

	vectorStage := bson.D{
		{Key: "index", Value: m.cfg.VectorIndex},
		{Key: "path", Value: "embedding"},
		{Key: "queryVector", Value: float32ToFloat64(query.QueryVector)},
		{Key: "numCandidates", Value: query.Candidates()},
		{Key: "limit", Value: query.Limit},
	}

	if len(query.Filter) > 0 {
		vectorStage = append(vectorStage, bson.E{Key: "filter", Value: query.Filter})
	}

	pipeline := mongo.Pipeline{
		{{Key: "$vectorSearch", Value: vectorStage}},
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "content", Value: 1},
			{Key: "metadata", Value: 1},
			{Key: "embedding", Value: 1},
			{Key: "score", Value: bson.D{{Key: "$meta", Value: "vectorSearchScore"}}},
		}}},
	}

	cursor, err := m.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("vector search aggregate: %w", err)
	}
	defer cursor.Close(ctx)

	var results []model.Document
	for cursor.Next(ctx) {
		var doc struct {
			ID        primitive.ObjectID     `bson:"_id"`
			Content   string                 `bson:"content"`
			Embedding []float64              `bson:"embedding"`
			Metadata  map[string]interface{} `bson:"metadata"`
			Score     float64                `bson:"score"`
		}
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("decode vector search result: %w", err)
		}

		results = append(results, model.Document{
			ID:        doc.ID,
			Content:   doc.Content,
			Embedding: float64ToFloat32(doc.Embedding),
			Metadata:  doc.Metadata,
			Score:     doc.Score,
		})
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("iterate vector search cursor: %w", err)
	}

	return results, nil
}

func float32ToFloat64(vector []float32) []float64 {
	result := make([]float64, len(vector))
	for i, v := range vector {
		result[i] = float64(v)
	}
	return result
}

func float64ToFloat32(vector []float64) []float32 {
	result := make([]float32, len(vector))
	for i, v := range vector {
		result[i] = float32(v)
	}
	return result
}

// EnsureIndexes creates the Atlas Vector Search index when it does not exist.
func EnsureIndexes(ctx context.Context, coll *mongo.Collection, cfg config.MongoDB) error {
	exists, err := vectorIndexExists(ctx, coll, cfg.VectorIndex)
	if err != nil {
		return fmt.Errorf("list vector indexes: %w", err)
	}
	if exists {
		return nil
	}

	command := bson.D{
		{Key: "createSearchIndexes", Value: coll.Name()},
		{Key: "indexes", Value: bson.A{
			bson.D{
				{Key: "name", Value: cfg.VectorIndex},
				{Key: "definition", Value: bson.D{
					{Key: "mappings", Value: bson.D{
						{Key: "dynamic", Value: false},
						{Key: "fields", Value: bson.D{
							{Key: "embedding", Value: bson.D{
								{Key: "type", Value: "vector"},
								{Key: "similarity", Value: "cosine"},
								{Key: "numDimensions", Value: cfg.EmbeddingDimension},
							}},
						}},
					}},
				}},
			},
		}},
	}

	if err := coll.Database().RunCommand(ctx, command).Err(); err != nil {
		return fmt.Errorf("create vector index: %w", err)
	}
	return nil
}

func vectorIndexExists(ctx context.Context, coll *mongo.Collection, name string) (bool, error) {
	cursor, err := coll.Aggregate(ctx, mongo.Pipeline{
		{{Key: "$listSearchIndexes", Value: bson.D{{Key: "name", Value: name}}}},
	})
	if err != nil {
		return false, err
	}
	defer cursor.Close(ctx)

	return cursor.Next(ctx), cursor.Err()
}
