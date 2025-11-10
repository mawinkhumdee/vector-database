package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type MongoConfig struct {
	URI                string `yaml:"uri"`
	Database           string `yaml:"database"`
	Collection         string `yaml:"collection"`
	VectorIndex        string `yaml:"vectorIndex"`
	EmbeddingDimension int    `yaml:"embeddingDimension"`
}

type AppConfig struct {
	Mongo MongoConfig `yaml:"mongo"`
}

const (
	defaultConfigPath = "config.yml"
	defaultMongoURI   = "mongodb://vector:secretvector@localhost:27017/?authSource=admin&directConnection=true"
)

func Load() (AppConfig, error) {
	path := getEnv("CONFIG_PATH", defaultConfigPath)

	content, err := os.ReadFile(path)
	if err != nil {
		return AppConfig{}, fmt.Errorf("read config file %s: %w", path, err)
	}

	var cfg AppConfig
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		return AppConfig{}, fmt.Errorf("parse config file %s: %w", path, err)
	}

	applyDefaults(&cfg)

	if err := validate(cfg); err != nil {
		return AppConfig{}, err
	}

	return cfg, nil
}

func applyDefaults(cfg *AppConfig) {
	if cfg.Mongo.URI == "" {
		cfg.Mongo.URI = defaultMongoURI
	}
	if cfg.Mongo.Database == "" {
		cfg.Mongo.Database = "vectors"
	}
	if cfg.Mongo.Collection == "" {
		cfg.Mongo.Collection = "documents"
	}
	if cfg.Mongo.VectorIndex == "" {
		cfg.Mongo.VectorIndex = "vector_index"
	}
	if cfg.Mongo.EmbeddingDimension == 0 {
		cfg.Mongo.EmbeddingDimension = 1536
	}
}

func validate(cfg AppConfig) error {
	if cfg.Mongo.URI == "" {
		return fmt.Errorf("mongo.uri must be provided")
	}
	if cfg.Mongo.EmbeddingDimension <= 0 {
		return fmt.Errorf("mongo.embeddingDimension must be positive, got %d", cfg.Mongo.EmbeddingDimension)
	}
	return nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}
