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

	// Wrap the mux with CORS middleware
	handler := corsMiddleware(mux)
	return http.ListenAndServe("127.0.0.1:7788", handler)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173") // Vite's default dev server port
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
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
