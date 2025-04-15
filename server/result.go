package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type ResultDetail struct {
	ID   string `json:"id"`
	Text string `json:"text"`
	URL  string `json:"url"`
}

func (s *Server) addResultRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/result/{id}", s.jsonResultHandler)
}

func (s *Server) jsonResultHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "ID parameter is required", http.StatusBadRequest)
		return
	}

	// Convert string ID to int for database
	idInt, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	// Get the document from the database
	doc, err := s.querier.GetDocumentByID(r.Context(), int64(idInt))
	if err != nil {
		http.Error(w, "Document not found", http.StatusNotFound)
		return
	}

	// Sanitize the text
	sanitizedText := s.sanitizerPolicy.Sanitize(doc.Text)

	result := ResultDetail{
		ID:   id,
		Text: sanitizedText,
		URL:  doc.Url,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(result); err != nil {
		fmt.Println("error", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
