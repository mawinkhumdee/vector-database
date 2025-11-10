package service

import (
	"context"
	"errors"
	"hash/fnv"
	"math"
	"strings"
)

// VectorEncoder abstract the embedding generation logic. It can be backed by any ML model or API.
type VectorEncoder interface {
	Encode(ctx context.Context, text string) ([]float32, error)
}

// HashEncoder is a deterministic, low fidelity encoder that is useful for demos and tests.
// It hashes each token into a bucket (dimension) and normalises the vector to unit length.
type HashEncoder struct {
	Dimension int
}

// NewHashEncoder ensures the encoder always has a positive dimension.
func NewHashEncoder(dimension int) (*HashEncoder, error) {
	if dimension <= 0 {
		return nil, errors.New("encoder dimension must be positive")
	}
	return &HashEncoder{Dimension: dimension}, nil
}

func (h *HashEncoder) Encode(_ context.Context, text string) ([]float32, error) {
	if h.Dimension <= 0 {
		return nil, errors.New("encoder dimension must be positive")
	}

	vector := make([]float32, h.Dimension)
	if text == "" {
		return vector, nil
	}

	tokens := strings.Fields(strings.ToLower(text))
	for _, token := range tokens {
		hasher := fnv.New64a()
		_, _ = hasher.Write([]byte(token))
		idx := hasher.Sum64() % uint64(h.Dimension)
		vector[idx] += 1
	}

	normalise(vector)
	return vector, nil
}

func normalise(vec []float32) {
	var sum float64
	for _, v := range vec {
		sum += float64(v * v)
	}

	if sum == 0 {
		return
	}

	magnitude := float32(math.Sqrt(sum))
	for i, v := range vec {
		vec[i] = v / magnitude
	}
}
