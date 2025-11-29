package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"connectrpc.com/connect"
	"github.com/cinemaos/backend/internal/cache"
	"github.com/cinemaos/backend/internal/database"
	"github.com/cinemaos/backend/internal/middleware"
	"github.com/cinemaos/backend/internal/services"
	"github.com/joho/godotenv"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Connect to database
	if err := database.Connect(); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer database.Close()

	// Run migrations
	if err := database.AutoMigrate(); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// Connect to Redis
	if err := cache.Connect(); err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}
	defer cache.Close()

	// Create HTTP mux
	mux := http.NewServeMux()

	// Create interceptors
	interceptors := connect.WithInterceptors(middleware.AuthInterceptor())

	// Register Connect RPC services
	authService := services.NewAuthService()
	moviesService := services.NewMoviesService()
	showtimesService := services.NewShowtimesService()
	bookingsService := services.NewBookingsService()
	pricingService := services.NewPricingService()

	// Register service handlers (these will be implemented)
	mux.Handle(services.NewAuthServiceHandler(authService, interceptors))
	mux.Handle(services.NewMoviesServiceHandler(moviesService, interceptors))
	mux.Handle(services.NewShowtimesServiceHandler(showtimesService, interceptors))
	mux.Handle(services.NewBookingsServiceHandler(bookingsService, interceptors))
	mux.Handle(services.NewPricingServiceHandler(pricingService, interceptors))

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Enable CORS
	handler := corsMiddleware(mux)

	// Get port from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	// Start server with HTTP/2 support
	addr := fmt.Sprintf(":%s", port)
	log.Printf("🚀 Server starting on http://localhost%s", addr)
	log.Printf("📡 Connect RPC services ready")

	if err := http.ListenAndServe(addr, h2c.NewHandler(handler, &http2.Server{})); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := os.Getenv("CORS_ORIGIN")
		if origin == "" {
			origin = "http://localhost:3000"
		}

		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Connect-Protocol-Version, Connect-Timeout-Ms")
		w.Header().Set("Access-Control-Expose-Headers", "Connect-Protocol-Version, Connect-Timeout-Ms")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
