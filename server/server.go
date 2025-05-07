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

type Server struct {
	querier         *db.Queries
	searchClient    meilisearch.ServiceManager
	domain          string
	sanitizerPolicy *bluemonday.Policy
	frontendDir     string
	certCacheDir    string // Added field for configurable certificate cache directory
}

func NewServer(domain string, querier *db.Queries) *Server {
	// Default frontend directory for development
	frontendDir := "./frontend/dist"

	// Check if we're running in a Docker container
	if _, err := os.Stat("/app/frontend/dist"); err == nil {
		frontendDir = "/app/frontend/dist"
	}
	fmt.Println("Frontend directory:", frontendDir)

	// Use environment variables if provided
	meilisearchURL := "http://localhost:7700"
	if envURL := os.Getenv("MEILISEARCH_URL"); envURL != "" {
		meilisearchURL = envURL
	}

	// Get certificate cache directory from environment variable or use default
	certCacheDir := "certs"
	if envCertDir := os.Getenv("CERT_CACHE_DIR"); envCertDir != "" {
		certCacheDir = envCertDir
	}

	return &Server{
		domain:          domain,
		querier:         querier,
		searchClient:    meilisearch.New(meilisearchURL),
		sanitizerPolicy: bluemonday.NewPolicy().AllowElements("em", "mark"),
		frontendDir:     frontendDir,
		certCacheDir:    certCacheDir,
	}
}

func (s *Server) ListenAndServe() error {
	mux := http.NewServeMux()

	// API routes
	s.addSearchRoutes(mux)
	s.addResultRoutes(mux)

	// Serve frontend static files
	mux.HandleFunc("/", s.serveStaticFiles())

	handler := corsMiddleware(mux)

	// Check if domain is provided to enable HTTPS with autocert
	if s.domain != "" {
		return s.serveHTTPS(handler)
	}

	// Otherwise, serve regular HTTP
	host := os.Getenv("HOST")
	if host == "" {
		host = "0.0.0.0"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "7788"
	}

	fmt.Printf("Starting HTTP server on %s:%s\n", host, port)
	return http.ListenAndServe(host+":"+port, handler)
}

// serveHTTPS sets up an HTTPS server with automatic certificate management via Let's Encrypt
func (s *Server) serveHTTPS(handler http.Handler) error {
	// Create certificate cache directory if it doesn't exist
	if err := os.MkdirAll(s.certCacheDir, 0700); err != nil {
		return fmt.Errorf("failed to create cert cache directory '%s': %v", s.certCacheDir, err)
	}
	fmt.Printf("Using certificate cache directory: %s\n", s.certCacheDir)

	// Set up the autocert manager
	certManager := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(s.domain),  // Only allow our domain
		Cache:      autocert.DirCache(s.certCacheDir), // Use the configurable cert directory
	}

	// Set up the HTTPS server
	httpsServer := &http.Server{
		Addr:    ":443",
		Handler: handler,
		TLSConfig: &tls.Config{
			GetCertificate: certManager.GetCertificate,
			MinVersion:     tls.VersionTLS12, // Improve cert reputation score
		},
	}

	// Set up an HTTP server to redirect to HTTPS and handle ACME challenges
	httpServer := &http.Server{
		Addr:    ":80",
		Handler: certManager.HTTPHandler(http.HandlerFunc(redirectToHTTPS)),
	}

	// Start HTTP server in a goroutine
	go func() {
		fmt.Printf("Starting HTTP server for ACME challenges and redirects on port 80\n")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	// Start HTTPS server
	fmt.Printf("Starting HTTPS server for domain %s on port 443\n", s.domain)
	return httpsServer.ListenAndServeTLS("", "") // Certificates come from Let's Encrypt
}

// redirectToHTTPS redirects HTTP requests to HTTPS
func redirectToHTTPS(w http.ResponseWriter, r *http.Request) {
	httpsURL := "https://" + r.Host + r.RequestURI
	http.Redirect(w, r, httpsURL, http.StatusMovedPermanently)
}

// serveStaticFiles returns a handler that serves static files from the frontend build directory
func (s *Server) serveStaticFiles() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// First, try to serve the requested path as a file
		path := filepath.Join(s.frontendDir, r.URL.Path)

		// Check if the file exists
		if _, err := os.Stat(path); err == nil {
			http.ServeFile(w, r, path)
			return
		}

		// For SPA (Single Page Application) routing, serve index.html for paths that don't match files
		indexPath := filepath.Join(s.frontendDir, "index.html")
		if _, err := os.Stat(indexPath); err == nil {
			http.ServeFile(w, r, indexPath)
			return
		}

		// If index.html doesn't exist, return 404
		http.NotFound(w, r)
	}
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
