package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Configs struct {
	MongoDB MongoDB `yaml:"mongo"`
}

type MongoDB struct {
	URI                string     `yaml:"uri"`
	Database           string     `yaml:"database"`
	Collection         Collection `yaml:"collection"`
	VectorIndex        string     `yaml:"vectorIndex"`
	EmbeddingDimension int        `yaml:"embeddingDimension"`
}

type Collection struct {
	Document string `yaml:"document"`
	analyze  string `yaml:"analyze"`
}

const (
	defaultConfigPath = "config.yml"
	defaultMongoURI   = "mongodb://vector:secretvector@localhost:27017/?authSource=admin&directConnection=true"
)

func Load() (Configs, error) {
	path := getEnv("CONFIG_PATH", defaultConfigPath)

	content, err := os.ReadFile(path)
	if err != nil {
		return Configs{}, fmt.Errorf("read config file %s: %w", path, err)
	}

	var cfg Configs
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		return Configs{}, fmt.Errorf("parse config file %s: %w", path, err)
	}

	if err := validate(cfg); err != nil {
		return Configs{}, err
	}

	return cfg, nil
}

func validate(cfg Configs) error {
	if cfg.MongoDB.URI == "" {
		return fmt.Errorf("mongo.uri must be provided")
	}
	if cfg.MongoDB.EmbeddingDimension <= 0 {
		return fmt.Errorf("mongo.embeddingDimension must be positive, got %d", cfg.MongoDB.EmbeddingDimension)
	}
	return nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}
