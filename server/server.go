package server

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/alexmorten/patchy/db"
	"github.com/meilisearch/meilisearch-go"
	"github.com/microcosm-cc/bluemonday"
	"golang.org/x/crypto/acme/autocert"
)

type ServerConfig struct {
	Domain          string
	Querier         *db.Queries
	FrontendDir     string
	MeilisearchURL  string
	CertCacheDir    string
	Host            string
	Port            string
}

type Server struct {
	config          ServerConfig
	searchClient    meilisearch.ServiceManager
	sanitizerPolicy *bluemonday.Policy
}

func NewServer(config ServerConfig) *Server {
	return &Server{
		config:          config,
		searchClient:    meilisearch.New(config.MeilisearchURL),
		sanitizerPolicy: bluemonday.NewPolicy().AllowElements("em", "mark"),
	}
}



func (s *Server) ListenAndServe() error {
	handler := s.setupRoutes()
	
	if s.config.Domain != "" {
		return s.serveHTTPS(handler)
	}
	
	return s.serveHTTP(handler)
}

func (s *Server) setupRoutes() http.Handler {
	mux := http.NewServeMux()
	s.addSearchRoutes(mux)
	s.addResultRoutes(mux)
	mux.HandleFunc("/", s.serveStaticFiles())
	return corsMiddleware(mux)
}

func (s *Server) serveHTTP(handler http.Handler) error {
	addr := fmt.Sprintf("%s:%s", s.config.Host, s.config.Port)
	fmt.Printf("Starting HTTP server on %s\n", addr)
	return http.ListenAndServe(addr, handler)
}

func getEnvOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func (s *Server) serveHTTPS(handler http.Handler) error {
	if err := s.ensureCertDir(); err != nil {
		return err
	}
	
	certManager := s.createCertManager()
	httpsServer := s.createHTTPSServer(handler, certManager)
	httpServer := s.createRedirectServer(certManager)
	
	s.startHTTPRedirectServer(httpServer)
	return s.startHTTPSServer(httpsServer)
}

func (s *Server) ensureCertDir() error {
	if err := os.MkdirAll(s.config.CertCacheDir, 0700); err != nil {
		return fmt.Errorf("failed to create cert cache directory '%s': %v", s.config.CertCacheDir, err)
	}
	return nil
}

func (s *Server) createCertManager() *autocert.Manager {
	manager := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(s.config.Domain),
		Cache:      autocert.DirCache(s.config.CertCacheDir),
	}
	return manager
}

func (s *Server) createHTTPSServer(handler http.Handler, certManager *autocert.Manager) *http.Server {
	return &http.Server{
		Addr:    ":443",
		Handler: handler,
		TLSConfig: &tls.Config{
			GetCertificate: certManager.GetCertificate,
			MinVersion:     tls.VersionTLS12,
		},
	}
}

func (s *Server) createRedirectServer(certManager *autocert.Manager) *http.Server {
	return &http.Server{
		Addr:    ":80",
		Handler: certManager.HTTPHandler(http.HandlerFunc(redirectToHTTPS)),
	}
}

func (s *Server) startHTTPRedirectServer(server *http.Server) {
	go func() {
		fmt.Printf("Starting HTTP server for ACME challenges and redirects on port 80\n")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()
}

func (s *Server) startHTTPSServer(server *http.Server) error {
	fmt.Printf("Starting HTTPS server for domain %s on port 443\n", s.config.Domain)
	return server.ListenAndServeTLS("", "")
}

// redirectToHTTPS redirects HTTP requests to HTTPS
func redirectToHTTPS(w http.ResponseWriter, r *http.Request) {
	httpsURL := "https://" + r.Host + r.RequestURI
	http.Redirect(w, r, httpsURL, http.StatusMovedPermanently)
}

func (s *Server) serveStaticFiles() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestedPath := filepath.Join(s.config.FrontendDir, r.URL.Path)
		
		if s.fileExists(requestedPath) {
			http.ServeFile(w, r, requestedPath)
			return
		}
		
		indexPath := filepath.Join(s.config.FrontendDir, "index.html")
		if s.fileExists(indexPath) {
			http.ServeFile(w, r, indexPath)
			return
		}
		
		http.NotFound(w, r)
	}
}

func (s *Server) fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setCORSHeaders(w)
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}
