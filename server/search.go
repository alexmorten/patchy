package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/url"

	"github.com/alexmorten/livereload"
	"github.com/meilisearch/meilisearch-go"
)

type SearchResult struct {
	ID   string `json:"id"`
	Text string `json:"text"`
	URL  string `json:"url"`
}

func (s *Server) addSearchRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /search", s.searchHandler)
	mux.HandleFunc("GET /", s.indexHandler)
	mux.HandleFunc("GET /api/search", s.jsonSearchHandler)
}

type Page struct {
	EnableLiveReload bool
	LiveReloadScript template.HTML
	Hits             []interface{}
	Query            string
}

func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")

	page := &Page{
		EnableLiveReload: s.IsDev,
		LiveReloadScript: livereload.LiveReloadScriptHTML(),
		Hits:             []interface{}{},
	}
	if query != "" {
		searchRes, err := s.searchClient.Index("documents").Search(query,
			&meilisearch.SearchRequest{
				AttributesToHighlight: []string{"Text"},
				Limit:                 100,
			})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		page.Hits = s.convertHits(searchRes.Hits)
		page.Query = query
	}

	err := s.getTemplate().ExecuteTemplate(w, "index.html", page)
	if err != nil {
		fmt.Println(err)
	}
}

func (s *Server) searchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.FormValue("query")
	searchRes, err := s.searchClient.Index("documents").Search(query,
		&meilisearch.SearchRequest{
			AttributesToHighlight: []string{"Text"},
			Limit:                 100,
		})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	page := Page{
		Hits:  s.convertHits(searchRes.Hits),
		Query: query,
	}
	w.Header().Add("HX-Replace-Url", "/?q="+url.QueryEscape(query))
	w.WriteHeader(200)
	err = s.getTemplate().ExecuteTemplate(w, "searchResponse.html", page)
	if err != nil {
		fmt.Println(err)
	}

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

		results = append(results, SearchResult{
			ID:   fmt.Sprintf("%d", int(hitMap["ID"].(float64))),
			Text: formattedText,
			URL:  hitMap["Url"].(string),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(results); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (s *Server) convertHits(hits []interface{}) []interface{} {
	var res []interface{}
	for _, hit := range hits {
		formatted := hit.(map[string]interface{})["_formatted"]
		formattedText := formatted.(map[string]interface{})["Text"].(string)
		sanitizedText := s.sanitizerPolicy.Sanitize(formattedText)
		safeHtml := template.HTML(sanitizedText)
		convertedHit := map[string]interface{}{
			"ID":   hit.(map[string]interface{})["ID"],
			"Text": safeHtml,
			"URL":  hit.(map[string]interface{})["Url"],
		}
		res = append(res, convertedHit)
	}
	return res
}
