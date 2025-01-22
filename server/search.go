package server

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/alexmorten/livereload"
	"github.com/meilisearch/meilisearch-go"
)

func (s *Server) addSearchRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /search", s.searchHandler)
	mux.HandleFunc("GET /", s.indexHandler)
}

func (s *Server) indexHandler(w http.ResponseWriter, _ *http.Request) {
	type Page struct {
		EnableLiveReload bool
		LiveReloadScript template.HTML
		Hits             []map[string]interface{}
	}

	page := &Page{
		EnableLiveReload: s.IsDev,
		LiveReloadScript: livereload.LiveReloadScriptHTML(),
		Hits: []map[string]interface{}{
			{"ID": 1,
				"Text": "Hello World",
				"URL":  "http://example.com",
			}},
	}
	fmt.Println("index")
	s.getTemplate().ExecuteTemplate(w, "index.html", page)
}

func (s *Server) searchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	searchRes, err := s.searchClient.Index("documents").Search(query,
		&meilisearch.SearchRequest{
			// AttributesToHighlight: []string{"Text"},
			Limit: 10,
		})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.getTemplate().ExecuteTemplate(w, "index.html", searchRes)
}
