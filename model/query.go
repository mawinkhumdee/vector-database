package model

import (
	"errors"
	"fmt"
)

// VectorQuery describes a similarity search request.
type VectorQuery struct {
	QueryVector   []float32
	Limit         int
	NumCandidates int
	Filter        map[string]interface{}
}

// Validate ensures the query has all information before it hits the db layer.
func (q VectorQuery) Validate(expectedDim int) error {
	if len(q.QueryVector) != expectedDim {
		return fmt.Errorf("query vector dimension mismatch: expected %d, got %d", expectedDim, len(q.QueryVector))
	}
	if q.Limit <= 0 {
		return errors.New("limit must be positive")
	}
	if q.NumCandidates != 0 && q.NumCandidates < q.Limit {
		return errors.New("numCandidates must be >= limit")
	}
	return nil
}

// Candidates returns a safe default when the caller does not specify a value.
func (q VectorQuery) Candidates() int {
	if q.NumCandidates > 0 {
		return q.NumCandidates
	}
	if q.Limit > 0 {
		return q.Limit * 5
	}
	return 50
}
