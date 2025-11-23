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
	"vector-database/service"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	database, err := db.New(ctx, cfg.MongoDB)
	if err != nil {
		log.Fatalf("init mongo vector store: %v", err)
	}
	defer func() {
		_ = database.Close(context.Background())
	}()

	encoder, err := service.NewEncoder(cfg.MongoDB.EmbeddingDimension)
	if err != nil {
		log.Fatalf("init encoder: %v", err)
	}

	embeddingService, err := service.NewSearch(database.Documents, encoder, cfg.MongoDB.EmbeddingDimension)
	if err != nil {
		log.Fatalf("init embedding service: %v", err)
	}

	messageHandler := handler.NewMessageHandler(embeddingService)
	imageHandler := handler.NewImageHandler(embeddingService)
	mux := http.NewServeMux()
	messageHandler.Register(mux)
	imageHandler.Register(mux)

	log.Printf("HTTP server listening on %s", httpinfo.DefaultAddr)
	if err := http.ListenAndServe(httpinfo.DefaultAddr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
