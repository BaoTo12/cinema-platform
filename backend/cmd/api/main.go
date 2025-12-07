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

	"cinemaos-backend/internal/config"
	authapp "cinemaos-backend/internal/app/auth"
	"cinemaos-backend/internal/app/authinfra"
	cinemaapp "cinemaos-backend/internal/app/cinema"
	movieapp "cinemaos-backend/internal/app/movie"
	showtimeapp "cinemaos-backend/internal/app/showtime"
	"cinemaos-backend/internal/app/postgres"
	"cinemaos-backend/internal/app/redis"
	"cinemaos-backend/internal/handler"
	"cinemaos-backend/internal/middleware"
	"cinemaos-backend/internal/router"
	httpserver "cinemaos-backend/internal/server"
	"cinemaos-backend/internal/pkg/logger"
	"cinemaos-backend/internal/pkg/tracer"
	"cinemaos-backend/internal/pkg/validator"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

// @title           CinemaOS API
// @version         1.0
// @description     Cinema Operating System Backend API
// @termsOfService  http://swagger.io/terms/

// @contact.name    API Support
// @contact.url     http://www.swagger.io/support
// @contact.email   support@cinemaos.com

// @license.name    Apache 2.0
// @license.url     http://www.apache.org/licenses/LICENSE-2.0.html

// @host            localhost:8080
// @BasePath        /api/v1
// @schemes         http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "", "path to config file")
	flag.Parse()

	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found, using environment variables")
	}

	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	// Initialize logger
	log, err := logger.New(logger.Config{
		Level:      cfg.Logger.Level,
		Format:     cfg.Logger.Format,
		Output:     cfg.Logger.Output,
		TimeFormat: cfg.Logger.TimeFormat,
	})
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	log.Info("Starting CinemaOS Backend",
		zap.String("version", cfg.App.Version),
		zap.String("environment", cfg.App.Environment),
	)

	// Initialize tracer
	tp, err := tracer.New(tracer.Config{
		Enabled:     cfg.Tracer.Enabled,
		ServiceName: cfg.Tracer.ServiceName,
		Endpoint:    cfg.Tracer.Endpoint,
		Insecure:    cfg.Tracer.Insecure,
		SampleRate:  cfg.Tracer.SampleRate,
		Environment: cfg.App.Environment,
		Version:     cfg.App.Version,
	})
	if err != nil {
		log.Fatal("Failed to initialize tracer", zap.Error(err))
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Error("Failed to shutdown tracer", zap.Error(err))
		}
	}()

	// Initialize Database
	db, err := postgres.New(cfg.Database, log)
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	if err := db.AutoMigrate(); err != nil {
		log.Fatal("Failed to run migrations", zap.Error(err))
	}

	// Initialize Redis
	// In a real scenario, you'd likely want to handle redis connection failure gracefully if it's optional
	// But let's assume it's critical for auth caching
	redisClient, err := redis.New(cfg.Redis, log)
	if err != nil {
		log.Error("Failed to connect to Redis", zap.Error(err))
		// continue without redis or exit depending on requirements?
		// for now, let's log and continue, but auth token revocation/caching might be affected
	} else {
		defer redisClient.Close()
	}

	// Repositories
	userRepo := postgres.NewUserRepository(db)
	refreshRepo := postgres.NewRefreshTokenRepository(db)
	resetTokenRepo := postgres.NewPasswordResetTokenRepository(db)
	movieRepo := postgres.NewMovieRepository(db)
	cinemaRepo := postgres.NewCinemaRepository(db)
	screenRepo := postgres.NewScreenRepository(db)
	seatRepo := postgres.NewSeatRepository(db)
	showtimeRepo := postgres.NewShowtimeRepository(db)

	// Infrastructure Services
	jwtManager := authinfra.NewJWTManager(cfg.JWT)
	passwordManager := authinfra.NewPasswordManager()

	// Application Services
	authService := authapp.NewService(
		userRepo,
		refreshRepo,
		resetTokenRepo,
		jwtManager,
		passwordManager,
		log,
		cfg.Email.FrontendURL,
	)
	movieService := movieapp.NewService(movieRepo, log)
	cinemaService := cinemaapp.NewService(cinemaRepo, screenRepo, seatRepo, log)
	showtimeService := showtimeapp.NewService(showtimeRepo, movieRepo, cinemaRepo, screenRepo, log)

	// Validator
	requestValidator := validator.New()

	// Handlers
	authHandler := handler.NewAuthHandler(authService, requestValidator)
	healthHandler := handler.NewHealthHandler(cfg, db, redisClient)
	movieHandler := handler.NewMovieHandler(movieService, requestValidator)
	cinemaHandler := handler.NewCinemaHandler(cinemaService, requestValidator)
	showtimeHandler := handler.NewShowtimeHandler(showtimeService, requestValidator)

	// Middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtManager, log)

	// Router
	appRouter := router.NewRouter(
		cfg,
		log,
		authMiddleware,
		authHandler,
		healthHandler,
		movieHandler,
		cinemaHandler,
		showtimeHandler,
	)

	// Server
	srv := httpserver.NewServer(cfg.Server, appRouter.Setup(), log)

	// Graceful shutdown
	go func() {
		if err := srv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) { // http.ErrServerClosed
			// http.NewServer wrapper might return a different error depending on implementation
			// let's check standard http error
			log.Fatal("Server failed to start", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown", zap.Error(err))
	}

	log.Info("Server exited properly")
}
