package server

import (
	"html/template"
	"net/http"

	"github.com/alexmorten/livereload"
	"github.com/alexmorten/patchy/db"
	"github.com/meilisearch/meilisearch-go"
	"github.com/microcosm-cc/bluemonday"
)

type Server struct {
	querier      *db.Queries
	getTemplate  func() *template.Template
	searchClient meilisearch.ServiceManager
	IsDev        bool

	sanitizerPolicy *bluemonday.Policy
}

var templateNames = []string{"server/pages/index.css", "server/pages/result.html", "server/pages/index.html", "server/pages/searchForm.html", "server/pages/searchResponse.html"}

func NewServer(isDev bool, querier *db.Queries) *Server {
	return &Server{
		IsDev:           isDev,
		querier:         querier,
		getTemplate:     templateGetter(isDev),
		searchClient:    meilisearch.New("http://localhost:7700"),
		sanitizerPolicy: bluemonday.NewPolicy().AllowElements("em"),
	}
}

func (s *Server) ListenAndServe() error {
	mux := http.NewServeMux()

	mux.Handle("GET /assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("server/pages/assets"))))
	s.addSearchRoutes(mux)
	if s.IsDev {
		livereload.HandleLiveReload(mux, "server/pages", "server/pages/assets")
	}

	return http.ListenAndServe("127.0.0.1:7788", mux)
}

func templateGetter(isDev bool) func() *template.Template {
	if isDev {
		return func() *template.Template {
			return template.Must(template.ParseFiles(templateNames...))
		}
	}

	t := template.Must(template.ParseFiles(templateNames...))
	return func() *template.Template {
		return t
	}
}
