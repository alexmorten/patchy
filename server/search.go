package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/meilisearch/meilisearch-go"
)

type SearchResult struct {
	ID   string `json:"id"`
	Text string `json:"text"`
	URL  string `json:"url"`
}

func (s *Server) addSearchRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/search", s.jsonSearchHandler)
}

func (s *Server) jsonSearchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	searchRes, err := s.searchClient.Index("documents").Search(query,
		&meilisearch.SearchRequest{
			AttributesToHighlight: []string{"Text"},
			Limit:                 10,
		})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	results := make([]SearchResult, 0, len(searchRes.Hits))
	for _, hit := range searchRes.Hits {
		hitMap := hit.(map[string]interface{})
		formatted := hitMap["_formatted"].(map[string]interface{})
		formattedText := formatted["Text"].(string)
		sanitizedText := s.sanitizerPolicy.Sanitize(formattedText)

		results = append(results, SearchResult{
			ID:   fmt.Sprintf("%d", int(hitMap["ID"].(float64))),
			Text: sanitizedText,
			URL:  hitMap["Url"].(string),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(results); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
