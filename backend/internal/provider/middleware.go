package provider

import (
	"cinemaos-backend/internal/app/authinfra"
	"cinemaos-backend/internal/middleware"
	"cinemaos-backend/internal/pkg/logger"
)

// ProvideAuthMiddleware creates and returns an auth middleware
func ProvideAuthMiddleware(
	jwtManager *authinfra.JWTManager,
	logger *logger.Logger,
) *middleware.AuthMiddleware {
	return middleware.NewAuthMiddleware(jwtManager, logger)
}
