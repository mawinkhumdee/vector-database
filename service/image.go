package service

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"vector-database/model"
)

const (
	imageDescriptionMetadataKey = "_image_description"
	imagePayloadMetadataKey     = "_image_payload"
	maxImageSampleBytes         = 4096
	imageChunkSize              = 64
)

// ErrInvalidArgument signals that the caller-provided payload is invalid.
var ErrInvalidArgument = errors.New("invalid argument")

func (s *searchImp) InsertImage(ctx context.Context, input model.ImageInput) (model.ImageDocument, error) {
	if err := input.Validate(); err != nil {
		return model.ImageDocument{}, fmt.Errorf("%w: %v", ErrInvalidArgument, err)
	}

	imageBytes, err := normaliseImageData(input.ImageData)
	if err != nil {
		return model.ImageDocument{}, fmt.Errorf("%w: %v", ErrInvalidArgument, err)
	}

	docInput := model.DocumentInput{
		Content:  composeImageEmbeddingText(input.Description, imageBytes),
		Metadata: cloneMetadata(input.Metadata),
	}
	docInput.Metadata[imageDescriptionMetadataKey] = input.Description
	docInput.Metadata[imagePayloadMetadataKey] = base64.StdEncoding.EncodeToString(imageBytes)

	doc, err := s.IndexDocument(ctx, docInput)
	if err != nil {
		return model.ImageDocument{}, err
	}

	return newImageDocument(doc), nil
}

func (s *searchImp) SearchImages(ctx context.Context, query model.ImageQuery) ([]model.ImageDocument, error) {
	if err := query.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidArgument, err)
	}

	imageBytes, err := normaliseImageData(query.ImageData)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidArgument, err)
	}

	limit := query.Limit
	if limit == 0 {
		limit = 5
	}

	payload := composeImageEmbeddingText(query.Description, imageBytes)
	docs, err := s.SearchByText(ctx, payload, limit)
	if err != nil {
		return nil, err
	}

	results := make([]model.ImageDocument, len(docs))
	for i, doc := range docs {
		results[i] = newImageDocument(doc)
	}

	return results, nil
}

func newImageDocument(doc model.Document) model.ImageDocument {
	resp := model.ImageDocument{
		Description: extractImageDescription(doc),
		Metadata:    sanitizeImageMetadata(doc.Metadata),
		Score:       doc.Score,
	}
	if doc.ID != primitive.NilObjectID {
		resp.ID = doc.ID.Hex()
	}
	return resp
}

func extractImageDescription(doc model.Document) string {
	if val, ok := doc.Metadata[imageDescriptionMetadataKey]; ok {
		if desc, ok := val.(string); ok && desc != "" {
			return desc
		}
	}
	return doc.Content
}

func sanitizeImageMetadata(metadata map[string]interface{}) map[string]interface{} {
	if len(metadata) == 0 {
		return nil
	}
	clean := make(map[string]interface{}, len(metadata))
	for k, v := range metadata {
		if k == imageDescriptionMetadataKey || k == imagePayloadMetadataKey {
			continue
		}
		clean[k] = v
	}
	if len(clean) == 0 {
		return nil
	}
	return clean
}

func cloneMetadata(metadata map[string]interface{}) map[string]interface{} {
	if len(metadata) == 0 {
		return map[string]interface{}{}
	}
	out := make(map[string]interface{}, len(metadata))
	for k, v := range metadata {
		out[k] = v
	}
	return out
}

func normaliseImageData(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("image is required")
	}
	if !isSupportedImageFormat(data) {
		return nil, errors.New("image must be a JPEG or PNG file")
	}
	return data, nil
}

func composeImageEmbeddingText(description string, data []byte) string {
	tokens := imageFeatureTokens(data)
	var builder strings.Builder

	desc := strings.TrimSpace(description)
	if desc != "" {
		builder.WriteString(desc)
	}

	for _, token := range tokens {
		if builder.Len() > 0 {
			builder.WriteByte(' ')
		}
		builder.WriteString(token)
	}

	return builder.String()
}

func imageFeatureTokens(data []byte) []string {
	if len(data) == 0 {
		return nil
	}
	if len(data) > maxImageSampleBytes {
		data = data[:maxImageSampleBytes]
	}

	tokenCount := len(data)/imageChunkSize + 8
	tokens := make([]string, 0, tokenCount)
	for i := 0; i < len(data); i += imageChunkSize {
		end := i + imageChunkSize
		if end > len(data) {
			end = len(data)
		}
		chunk := data[i:end]
		var sum int
		for _, b := range chunk {
			sum += int(b)
		}
		avg := sum / len(chunk)
		tokens = append(tokens, fmt.Sprintf("pix:%02x", avg))
	}

	hash := sha256.Sum256(data)
	for i := 0; i < len(hash); i += 4 {
		end := i + 4
		if end > len(hash) {
			end = len(hash)
		}
		tokens = append(tokens, fmt.Sprintf("sig:%x", hash[i:end]))
	}

	return tokens
}

func isSupportedImageFormat(data []byte) bool {
	if len(data) < 4 {
		return false
	}
	// JPEG magic number: FF D8 FF
	if len(data) >= 3 && data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
		return true
	}
	// PNG magic number: 89 50 4E 47 0D 0A 1A 0A
	if len(data) >= 8 &&
		data[0] == 0x89 &&
		data[1] == 0x50 &&
		data[2] == 0x4E &&
		data[3] == 0x47 &&
		data[4] == 0x0D &&
		data[5] == 0x0A &&
		data[6] == 0x1A &&
		data[7] == 0x0A {
		return true
	}
	return false
}
