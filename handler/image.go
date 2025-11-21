package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"vector-database/httpinfo"
	"vector-database/model"
	"vector-database/service"
)

const (
	imageFormField      = "image"
	maxImageUploadBytes = 10 << 20 // 10MB
)

// ImageHandler exposes endpoints for inserting images and searching similar ones.
type ImageHandler struct {
	service service.SearchService
}

// NewImageHandler wires the provided SearchService into HTTP routes.
func NewImageHandler(svc service.SearchService) *ImageHandler {
	return &ImageHandler{service: svc}
}

// Register attaches the image HTTP endpoints to the mux.
func (h *ImageHandler) Register(mux *http.ServeMux) {
	mux.Handle(httpinfo.InsertImageEndpoint.Path, http.HandlerFunc(h.handleInsertImage))
	mux.Handle(httpinfo.SearchImageEndpoint.Path, http.HandlerFunc(h.handleSearchImage))
}

func (h *ImageHandler) handleInsertImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != httpinfo.InsertImageEndpoint.Method {
		w.Header().Set("Allow", httpinfo.InsertImageEndpoint.Method)
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if err := r.ParseMultipartForm(maxImageUploadBytes); err != nil {
		writeError(w, http.StatusBadRequest, "multipart form required")
		return
	}
	defer cleanupMultipart(r)

	imageBytes, err := readImageField(r, imageFormField)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	metadata, err := parseMetadataField(r.FormValue("metadata"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	input := model.ImageInput{
		Description: r.FormValue("description"),
		ImageData:   imageBytes,
		Metadata:    metadata,
	}

	doc, err := h.service.InsertImage(r.Context(), input)
	if err != nil {
		writeError(w, statusFromError(err), err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"image": doc,
	})
}

type searchImageResponse struct {
	Results []model.ImageDocument `json:"results"`
}

func (h *ImageHandler) handleSearchImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != httpinfo.SearchImageEndpoint.Method {
		w.Header().Set("Allow", httpinfo.SearchImageEndpoint.Method)
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if err := r.ParseMultipartForm(maxImageUploadBytes); err != nil {
		writeError(w, http.StatusBadRequest, "multipart form required")
		return
	}
	defer cleanupMultipart(r)

	imageBytes, err := readImageField(r, imageFormField)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	limit, err := parseLimitField(r.FormValue("limit"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	query := model.ImageQuery{
		ImageData:   imageBytes,
		Description: r.FormValue("description"),
		Limit:       limit,
	}

	results, err := h.service.SearchImages(r.Context(), query)
	if err != nil {
		writeError(w, statusFromError(err), err.Error())
		return
	}

	writeJSON(w, http.StatusOK, searchImageResponse{
		Results: results,
	})
}

func statusFromError(err error) int {
	if errors.Is(err, service.ErrInvalidArgument) {
		return http.StatusBadRequest
	}
	return http.StatusInternalServerError
}

func readImageField(r *http.Request, field string) ([]byte, error) {
	file, _, err := r.FormFile(field)
	if err != nil {
		return nil, errors.New("image file is required")
	}
	defer file.Close()

	limited := io.LimitReader(file, maxImageUploadBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, fmt.Errorf("read image file: %w", err)
	}
	if int64(len(data)) > maxImageUploadBytes {
		return nil, fmt.Errorf("image file must be <= %d bytes", maxImageUploadBytes)
	}
	if len(data) == 0 {
		return nil, errors.New("image file is empty")
	}

	return data, nil
}

func parseMetadataField(raw string) (map[string]interface{}, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	var metadata map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &metadata); err != nil {
		return nil, errors.New("metadata must be valid JSON")
	}
	return metadata, nil
}

func parseLimitField(raw string) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return 0, errors.New("limit must be a positive integer")
	}
	return value, nil
}

func cleanupMultipart(r *http.Request) {
	if r.MultipartForm != nil {
		_ = r.MultipartForm.RemoveAll()
	}
}
