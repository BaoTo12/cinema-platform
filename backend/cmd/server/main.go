package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"cinemaos-backend/internal/cache"
	"cinemaos-backend/internal/database"
	
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  No .env file found, using environment variables")
	}

	log.Println("🚀 Starting CinemaOS Backend Server...")

	// Connect to database
	log.Println("📊 Connecting to PostgreSQL...")
	if err := database.Connect(); err != nil {
		log.Fatal("❌ Failed to connect to database:", err)
	}
	defer database.Close()

	// Run migrations
	log.Println("🔧 Running database migrations...")
	if err := database.AutoMigrate(); err != nil {
		log.Fatal("❌ Failed to run migrations:", err)
	}

	// Connect to Redis
	log.Println("🔴 Connecting to Redis...")
	if err := cache.Connect(); err != nil {
		log.Fatal("❌ Failed to connect to Redis:", err)
	}
	defer cache.Close()

	// Create HTTP mux
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","database":"connected","redis":"connected"}`))
	})

	// API info endpoint
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message":"CinemaOS API Server","version":"1.0.0","docs":"/health"}`))
	})

	// Enable CORS
	handler := corsMiddleware(mux)

	// Get port from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	// Start server
	addr := fmt.Sprintf(":%s", port)
	log.Printf("✅ Server ready!")
	log.Printf("🌐 Listening on http://localhost%s", addr)
	log.Printf("💚 Health check: http://localhost%s/health", addr)
	log.Println("==========================================")

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal("❌ Server failed to start:", err)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := os.Getenv("CORS_ORIGIN")
		if origin == "" {
			origin = "*"
		}

		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
