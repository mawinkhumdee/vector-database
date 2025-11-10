package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"vector-database/config"
	"vector-database/db"
	"vector-database/handler"
	"vector-database/httpinfo"
	"vector-database/model"
	"vector-database/service"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	store, err := db.InitMongoVectorStore(ctx, cfg.Mongo)
	if err != nil {
		log.Fatalf("init mongo vector store: %v", err)
	}
	defer func() {
		_ = store.Close(context.Background())
	}()

	encoder, err := service.NewHashEncoder(cfg.Mongo.EmbeddingDimension)
	if err != nil {
		log.Fatalf("init encoder: %v", err)
	}

	embeddingService, err := service.NewEmbeddingService(store, encoder, cfg.Mongo.EmbeddingDimension)
	if err != nil {
		log.Fatalf("init embedding service: %v", err)
	}

	seed(ctx, embeddingService)

	messageHandler := handler.NewMessageHandler(embeddingService)
	mux := http.NewServeMux()
	messageHandler.Register(mux)

	log.Printf("HTTP server listening on %s", httpinfo.DefaultAddr)
	if err := http.ListenAndServe(httpinfo.DefaultAddr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func seed(ctx context.Context, svc service.EmbeddingService) {
	seedDocuments := []model.DocumentInput{
		{Content: "Vectors power semantic search for knowledge bases.", Metadata: map[string]interface{}{"topic": "knowledge-base"}},
		{Content: "Recommendation systems match users with relevant products.", Metadata: map[string]interface{}{"topic": "recommendation"}},
		{Content: "Chatbots rely on embeddings to understand context.", Metadata: map[string]interface{}{"topic": "chatbot"}},
	}

	for _, doc := range seedDocuments {
		if _, err := svc.IndexDocument(ctx, doc); err != nil {
			log.Printf("seed warning: %v", err)
			return
		}
	}
}
