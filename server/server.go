package server

import (
	"net/http"

	"github.com/alexmorten/patchy/db"
	"github.com/meilisearch/meilisearch-go"
	"github.com/microcosm-cc/bluemonday"
)

type Server struct {
	querier         *db.Queries
	searchClient    meilisearch.ServiceManager
	IsDev           bool
	sanitizerPolicy *bluemonday.Policy
}

func NewServer(isDev bool, querier *db.Queries) *Server {
	return &Server{
		IsDev:           isDev,
		querier:         querier,
		searchClient:    meilisearch.New("http://localhost:7700"),
		sanitizerPolicy: bluemonday.NewPolicy().AllowElements("em", "mark"),
	}
}

func (s *Server) ListenAndServe() error {
	mux := http.NewServeMux()
	s.addSearchRoutes(mux)
	s.addResultRoutes(mux)
	handler := corsMiddleware(mux)
	return http.ListenAndServe("127.0.0.1:7788", handler)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
