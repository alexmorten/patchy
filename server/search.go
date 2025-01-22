package server

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/alexmorten/livereload"
	"github.com/meilisearch/meilisearch-go"
)

func (s *Server) addSearchRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /search", s.searchHandler)
	mux.HandleFunc("GET /", s.indexHandler)
}

type Page struct {
	EnableLiveReload bool
	LiveReloadScript template.HTML
	Hits             []interface{}
	Query            string
}

func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Println("index")

	page := &Page{
		EnableLiveReload: s.IsDev,
		LiveReloadScript: livereload.LiveReloadScriptHTML(),
		Hits:             []interface{}{},
	}

	err := s.getTemplate().ExecuteTemplate(w, "index.html", page)
	if err != nil {
		fmt.Println(err)
	}
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

	page := Page{
		Hits:  searchRes.Hits,
		Query: query,
	}

	err = s.getTemplate().ExecuteTemplate(w, "searchResponse.html", page)
	if err != nil {
		fmt.Println(err)
	}

}
