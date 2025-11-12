package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"vector-database/httpinfo"
	"vector-database/model"
	"vector-database/service"
)

// MessageHandler wires HTTP endpoints to the embedding service.
type MessageHandler struct {
	service service.SearchService
}

// NewMessageHandler creates a handler ready to register HTTP routes.
func NewMessageHandler(svc service.SearchService) *MessageHandler {
	return &MessageHandler{service: svc}
}

// Register attaches the handler methods to the provided mux.
func (h *MessageHandler) Register(mux *http.ServeMux) {
	mux.Handle(httpinfo.InsertMessageEndpoint.Path, http.HandlerFunc(h.dispatchMessages))
}

type insertMessageRequest struct {
	Content  string                 `json:"content"`
	Metadata map[string]interface{} `json:"metadata"`
}

func (h *MessageHandler) handleInsertMessage(w http.ResponseWriter, r *http.Request) {
	var req insertMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	docInput := model.DocumentInput{
		Content:  req.Content,
		Metadata: req.Metadata,
	}

	doc, err := h.service.IndexDocument(r.Context(), docInput)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"document": toMessageResponse(doc),
	})
}

type getMessageResponse struct {
	Query   string                    `json:"query"`
	Results []messageDocumentResponse `json:"results"`
}

func (h *MessageHandler) handleGetMessage(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		writeError(w, http.StatusBadRequest, "query parameter 'q' is required")
		return
	}

	limit, err := parseLimit(r.URL.Query().Get("limit"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	res, err := h.service.SearchByText(r.Context(), query, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, getMessageResponse{
		Query:   query,
		Results: toMessageResponses(res),
	})
}

func parseLimit(raw string) (int, error) {
	if raw == "" {
		return 5, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return 0, errors.New("limit must be a positive integer")
	}
	return value, nil
}

func methodNotAllowed(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Allow", httpinfo.InsertMessageEndpoint.Method+", "+httpinfo.GetMessageEndpoint.Method)
	writeError(w, http.StatusMethodNotAllowed, "method not allowed")
}

func (h *MessageHandler) dispatchMessages(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case httpinfo.InsertMessageEndpoint.Method:
		h.handleInsertMessage(w, r)
	case httpinfo.GetMessageEndpoint.Method:
		h.handleGetMessage(w, r)
	default:
		methodNotAllowed(w, r)
	}
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

type messageDocumentResponse struct {
	ID       string                 `json:"id,omitempty"`
	Content  string                 `json:"content"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	Score    float64                `json:"score,omitempty"`
}

func toMessageResponse(doc model.Document) messageDocumentResponse {
	resp := messageDocumentResponse{
		Content:  doc.Content,
		Metadata: doc.Metadata,
		Score:    doc.Score,
	}

	if doc.ID != primitive.NilObjectID {
		resp.ID = doc.ID.Hex()
	}

	return resp
}

func toMessageResponses(docs []model.Document) []messageDocumentResponse {
	responses := make([]messageDocumentResponse, len(docs))
	for i, doc := range docs {
		responses[i] = toMessageResponse(doc)
	}
	return responses
}
