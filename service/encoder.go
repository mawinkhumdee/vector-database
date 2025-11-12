package service

import (
	"context"
	"errors"
	"hash/fnv"
	"math"
	"strings"
)

func (h *encoderImp) Encode(_ context.Context, text string) ([]float32, error) {
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
