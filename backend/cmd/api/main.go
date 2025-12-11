package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "", "path to config file")
	flag.Parse()

	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found, using environment variables")
	}

	// Initialize application using Wire
	app, err := InitializeApplication(configPath)
	if err != nil {
		fmt.Printf("Failed to initialize application: %v\n", err)
		os.Exit(1)
	}

	app.Logger.Info("Starting CinemaOS Backend",
		zap.String("message", "All dependencies injected via Google Wire"),
	)

	// Cleanup resources on shutdown
	defer func() {
		if app.DB != nil {
			app.DB.Close()
		}
		if app.RedisClient != nil {
			app.RedisClient.Close()
		}
		if app.Tracer != nil {
			if err := app.Tracer.Shutdown(context.Background()); err != nil {
				app.Logger.Error("Failed to shutdown tracer", zap.Error(err))
			}
		}
	}()

	// Start server
	go func() {
		if err := app.Server.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			app.Logger.Fatal("Server failed to start", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	app.Logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), app.Config.Server.ShutdownTimeout)
	defer cancel()

	if err := app.Server.Shutdown(ctx); err != nil {
		app.Logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	app.Logger.Info("Server exited properly")
}
