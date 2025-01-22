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

	err = s.getTemplate().ExecuteTemplate(w, "searchResponse.html", page)
	if err != nil {
		fmt.Println(err)
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
