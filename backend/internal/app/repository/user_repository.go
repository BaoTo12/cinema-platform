package repository

import (
	"context"

	"cinemaos-backend/internal/app/entity"
	"github.com/google/uuid"
)

// UserRepository defines the interface for user data access
type UserRepository interface {
	// Create creates a new user
	Create(ctx context.Context, user *entity.User) error
	
	// GetByID retrieves a user by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	
	// GetByEmail retrieves a user by email
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	
	// Update updates a user
	Update(ctx context.Context, user *entity.User) error
	
	// Delete soft deletes a user
	Delete(ctx context.Context, id uuid.UUID) error
	
	// List returns a paginated list of users
	List(ctx context.Context, offset, limit int) ([]*entity.User, int64, error)
	
	// UpdatePassword updates the user's password
	UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error
	
	// UpdateLastLogin updates the user's last login timestamp
	UpdateLastLogin(ctx context.Context, id uuid.UUID) error
	
	// VerifyEmail marks the user's email as verified
	VerifyEmail(ctx context.Context, id uuid.UUID) error
	
	// EmailExists checks if an email already exists
	EmailExists(ctx context.Context, email string) (bool, error)
}

// RefreshTokenRepository defines the interface for refresh token data access
type RefreshTokenRepository interface {
	// Create creates a new refresh token
	Create(ctx context.Context, token *entity.RefreshToken) error
	
	// GetByTokenHash retrieves a token by its hash
	GetByTokenHash(ctx context.Context, tokenHash string) (*entity.RefreshToken, error)
	
	// GetByUserID retrieves all tokens for a user
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.RefreshToken, error)
	
	// Revoke revokes a token
	Revoke(ctx context.Context, id uuid.UUID) error
	
	// RevokeAllForUser revokes all tokens for a user
	RevokeAllForUser(ctx context.Context, userID uuid.UUID) error
	
	// DeleteExpired deletes all expired tokens
	DeleteExpired(ctx context.Context) error
}

// PasswordResetTokenRepository defines the interface for password reset token data access
type PasswordResetTokenRepository interface {
	// Create creates a new password reset token
	Create(ctx context.Context, token *entity.PasswordResetToken) error
	
	// GetByTokenHash retrieves a token by its hash
	GetByTokenHash(ctx context.Context, tokenHash string) (*entity.PasswordResetToken, error)
	
	// GetLatestByUserID retrieves the latest token for a user
	GetLatestByUserID(ctx context.Context, userID uuid.UUID) (*entity.PasswordResetToken, error)
	
	// MarkUsed marks a token as used
	MarkUsed(ctx context.Context, id uuid.UUID) error
	
	// DeleteExpired deletes all expired tokens
	DeleteExpired(ctx context.Context) error
	
	// InvalidateAllForUser invalidates all tokens for a user
	InvalidateAllForUser(ctx context.Context, userID uuid.UUID) error
}

// EmailVerificationTokenRepository defines the interface for email verification token data access
type EmailVerificationTokenRepository interface {
	// Create creates a new email verification token
	Create(ctx context.Context, token *entity.EmailVerificationToken) error
	
	// GetByTokenHash retrieves a token by its hash
	GetByTokenHash(ctx context.Context, tokenHash string) (*entity.EmailVerificationToken, error)
	
	// MarkUsed marks a token as used
	MarkUsed(ctx context.Context, id uuid.UUID) error
	
	// DeleteExpired deletes all expired tokens
	DeleteExpired(ctx context.Context) error
}
