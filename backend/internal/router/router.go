package router

import (
	"time"

	"cinemaos-backend/internal/config"
	"cinemaos-backend/internal/handler"
	"cinemaos-backend/internal/middleware"
	"cinemaos-backend/internal/pkg/logger"

	"github.com/gin-gonic/gin"
)

// Router holds all route dependencies
type Router struct {
	cfg            *config.Config
	logger         *logger.Logger
	authMiddleware *middleware.AuthMiddleware
	authHandler    *handler.AuthHandler
	healthHandler  *handler.HealthHandler
	movieHandler   *handler.MovieHandler
	cinemaHandler  *handler.CinemaHandler
	showtimeHandler *handler.ShowtimeHandler
}

// NewRouter creates a new router
func NewRouter(
	cfg *config.Config,
	logger *logger.Logger,
	authMiddleware *middleware.AuthMiddleware,
	authHandler *handler.AuthHandler,
	healthHandler *handler.HealthHandler,
	movieHandler *handler.MovieHandler,
	cinemaHandler *handler.CinemaHandler,
	showtimeHandler *handler.ShowtimeHandler,
) *Router {
	return &Router{
		cfg:            cfg,
		logger:         logger,
		authMiddleware: authMiddleware,
		authHandler:    authHandler,
		healthHandler:  healthHandler,
		movieHandler:   movieHandler,
		cinemaHandler:  cinemaHandler,
		showtimeHandler: showtimeHandler,
	}
}

// Setup configures the Gin router with all routes and middleware
func (r *Router) Setup() *gin.Engine {
	// Set Gin mode
	if r.cfg.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create router
	router := gin.New()

	// Global middleware
	router.Use(middleware.RecoveryMiddleware(r.logger))
	router.Use(middleware.RequestIDMiddleware())
	router.Use(middleware.LoggingMiddleware(r.logger))
	router.Use(middleware.ResponseTimeMiddleware(r.logger))
	router.Use(middleware.CORSMiddleware(r.cfg.CORS))
	router.Use(middleware.SecureHeadersMiddleware())

	// Rate limiting (100 requests per minute per IP)
	rateLimiter := middleware.NewRateLimiter(100, time.Minute)
	router.Use(rateLimiter.RateLimit())

	// Health check routes (no auth required)
	router.GET("/health", r.healthHandler.Health)
	router.GET("/health/ready", r.healthHandler.HealthDetailed)
	router.GET("/health/live", r.healthHandler.Live)
	router.GET("/info", r.healthHandler.Info)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Auth routes
		auth := v1.Group("/auth")
		{
			// Public routes
			auth.POST("/register", r.authHandler.Register)
			auth.POST("/login", r.authHandler.Login)
			auth.POST("/refresh", r.authHandler.RefreshToken)
			auth.POST("/forgot-password", r.authHandler.ForgotPassword)
			auth.POST("/reset-password", r.authHandler.ResetPassword)

			// Protected routes
			auth.POST("/logout", r.authMiddleware.Authenticate(), r.authHandler.Logout)
			auth.POST("/change-password", r.authMiddleware.Authenticate(), r.authHandler.ChangePassword)
			auth.GET("/me", r.authMiddleware.Authenticate(), r.authHandler.GetCurrentUser)
			auth.PATCH("/me", r.authMiddleware.Authenticate(), r.authHandler.UpdateProfile)
		}

		// Movies routes
		movies := v1.Group("/movies")
		{
			movies.GET("", r.movieHandler.List)
			movies.GET("/:id", r.movieHandler.GetByID)
			movies.GET("/now-showing", r.movieHandler.GetNowShowing)
			movies.GET("/coming-soon", r.movieHandler.GetComingSoon)
			movies.GET("/:id/showtimes", r.movieHandler.GetShowtimes)
			
			// Admin only
			movies.POST("", r.authMiddleware.Authenticate(), r.authMiddleware.RequireAdmin(), r.movieHandler.Create)
			movies.PUT("/:id", r.authMiddleware.Authenticate(), r.authMiddleware.RequireAdmin(), r.movieHandler.Update)
			movies.DELETE("/:id", r.authMiddleware.Authenticate(), r.authMiddleware.RequireAdmin(), r.movieHandler.Delete)
		}

		// Cinemas routes
		cinemas := v1.Group("/cinemas")
		{
			cinemas.GET("", r.cinemaHandler.List)
			cinemas.GET("/:id", r.cinemaHandler.GetByID)
			// cinemas.GET("/:id/showtimes", r.cinemaHandler.GetShowtimes) // To be implemented with Showtime module

			// Admin only
			cinemas.POST("", r.authMiddleware.Authenticate(), r.authMiddleware.RequireAdmin(), r.cinemaHandler.Create)
			cinemas.POST("/:id/screens", r.authMiddleware.Authenticate(), r.authMiddleware.RequireAdmin(), r.cinemaHandler.AddScreen)
		}

		// Showtime routes
		showtimes := v1.Group("/showtimes")
		{
			showtimes.GET("", r.showtimeHandler.List)
			showtimes.GET("/:id", r.showtimeHandler.GetByID)
			
			// Admin only
			showtimes.POST("", r.authMiddleware.Authenticate(), r.authMiddleware.RequireAdmin(), r.showtimeHandler.Create)
			showtimes.PUT("/:id", r.authMiddleware.Authenticate(), r.authMiddleware.RequireAdmin(), r.showtimeHandler.Update)
			showtimes.DELETE("/:id", r.authMiddleware.Authenticate(), r.authMiddleware.RequireAdmin(), r.showtimeHandler.Delete)
		}

		// Bookings routes (to be implemented)
		// bookings := v1.Group("/bookings")
		// {
		// 	bookings.Use(authMiddleware.Authenticate())
		// 	bookings.POST("/hold", bookingHandler.HoldSeats)
		// 	bookings.POST("/confirm", bookingHandler.ConfirmBooking)
		// 	bookings.GET("", bookingHandler.GetUserBookings)
		// 	bookings.GET("/:id", bookingHandler.GetByID)
		// 	bookings.POST("/:id/cancel", bookingHandler.Cancel)
		// }
	}

	// Not found handler
	router.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "NOT_FOUND",
				"message": "Route not found",
			},
		})
	})

	return router
}
