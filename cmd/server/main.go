package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/alexmorten/patchy/db"
	"github.com/alexmorten/patchy/server"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	domain := getDomain()
	
	if domain != "" {
		fmt.Printf("Starting server with autocert for domain: %s\n", domain)
	} else {
		fmt.Println("Starting server in HTTP mode (no domain provided)")
	}
	
	dbPool := setupDatabaseConnection()
	defer dbPool.Close()
	
	frontendDir := getFrontendDir()
	meilisearchURL := getEnvOrDefault("MEILISEARCH_URL", "http://localhost:7700")
	certCacheDir := getEnvOrDefault("CERT_CACHE_DIR", "certs")
	host := getEnvOrDefault("HOST", "0.0.0.0")
	port := getEnvOrDefault("PORT", "7788")
	
	querier := db.New(dbPool)
	
	config := server.ServerConfig{
		Domain:         domain,
		Querier:        querier,
		FrontendDir:    frontendDir,
		MeilisearchURL: meilisearchURL,
		CertCacheDir:   certCacheDir,
		Host:           host,
		Port:           port,
	}
	
	srv := server.NewServer(config)
	
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func getDomain() string {
	domain := flag.String("domain", "", "Domain name for HTTPS with Let's Encrypt (e.g., example.com)")
	flag.Parse()
	
	if *domain == "" {
		return os.Getenv("DOMAIN")
	}
	return *domain
}


func setupDatabaseConnection() *pgxpool.Pool {
	connString := getConnectionString()
	config := createPoolConfig(connString)
	
	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	
	return pool
}

func getEnvOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getFrontendDir() string {
	if _, err := os.Stat("/app/frontend/dist"); err == nil {
		return "/app/frontend/dist"
	}
	return "./frontend/dist"
}

func getConnectionString() string {
	return getEnvOrDefault("POSTGRES_CONNECTION_STRING", "postgresql://localhost:5432/patchy")
}

func createPoolConfig(connString string) *pgxpool.Config {
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		log.Fatalf("Invalid database connection string: %v", err)
	}
	
	config.MaxConns = 10
	config.MinConns = 2
	
	return config
}
