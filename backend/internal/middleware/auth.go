package middleware

import (
	"context"
	"strings"

	"connectrpc.com/connect"
	"cinemaos-backend/internal/utils"
)

const UserContextKey = "user"

type UserContext struct {
	UserID string
	Email  string
	Role   string
}

// AuthInterceptor validates JWT tokens from requests
func AuthInterceptor() connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			// Skip auth for certain procedures
			procedure := req.Spec().Procedure
			if isPublicProcedure(procedure) {
				return next(ctx, req)
			}

			// Get authorization header
			auth := req.Header().Get("Authorization")
			if auth == "" {
				return nil, connect.NewError(connect.CodeUnauthenticated, nil)
			}

			// Extract token
			parts := strings.Split(auth, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return nil, connect.NewError(connect.CodeUnauthenticated, nil)
			}

			token := parts[1]

			// Validate token
			claims, err := utils.ValidateAccessToken(token)
			if err != nil {
				return nil, connect.NewError(connect.CodeUnauthenticated, err)
			}

			// Add user to context
			userCtx := &UserContext{
				UserID: claims.UserID,
				Email:  claims.Email,
				Role:   claims.Role,
			}

			ctx = context.WithValue(ctx, UserContextKey, userCtx)

			return next(ctx, req)
		}
	}
}

func isPublicProcedure(procedure string) bool {
	publicProcedures := []string{
		"/cinema.v1.AuthService/Register",
		"/cinema.v1.AuthService/Login",
		"/cinema.v1.AuthService/RefreshToken",
		"/cinema.v1.MoviesService/ListMovies",
		"/cinema.v1.MoviesService/GetMovie",
		"/cinema.v1.MoviesService/GetNowShowing",
		"/cinema.v1.ShowtimesService/ListShowtimes",
		"/cinema.v1.ShowtimesService/GetShowtime",
		"/cinema.v1.ShowtimesService/GetSeatMap",
	}

	for _, p := range publicProcedures {
		if p == procedure {
			return true
		}
	}
	return false
}

// GetUserFromContext retrieves user context
func GetUserFromContext(ctx context.Context) (*UserContext, bool) {
	user, ok := ctx.Value(UserContextKey).(*UserContext)
	return user, ok
}

// RequireAdmin checks if user has admin/manager role
func RequireAdmin(ctx context.Context) error {
	user, ok := GetUserFromContext(ctx)
	if !ok {
		return connect.NewError(connect.CodeUnauthenticated, nil)
	}

	if user.Role != "ADMIN" && user.Role != "MANAGER" {
		return connect.NewError(connect.CodePermissionDenied, nil)
	}

	return nil
}
