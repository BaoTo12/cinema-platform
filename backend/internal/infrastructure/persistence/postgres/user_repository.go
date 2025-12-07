package postgres

import (
	"context"
	"errors"
	"time"

	"cinemaos-backend/internal/domain/entity"
	"cinemaos-backend/internal/domain/repository"
	apperrors "cinemaos-backend/internal/pkg/errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// userRepository implements repository.UserRepository
type userRepository struct {
	db *Database
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *Database) repository.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return apperrors.Wrap(err, apperrors.CodeInternal, "failed to create user")
	}
	return nil
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.CodeUserNotFound, "user not found")
		}
		return nil, apperrors.Wrap(err, apperrors.CodeInternal, "failed to get user")
	}
	return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).First(&user, "email = ?", email).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New(apperrors.CodeUserNotFound, "user not found")
		}
		return nil, apperrors.Wrap(err, apperrors.CodeInternal, "failed to get user")
	}
	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *entity.User) error {
	if err := r.db.WithContext(ctx).Save(user).Error; err != nil {
		return apperrors.Wrap(err, apperrors.CodeInternal, "failed to update user")
	}
	return nil
}

func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&entity.User{}, "id = ?", id).Error; err != nil {
		return apperrors.Wrap(err, apperrors.CodeInternal, "failed to delete user")
	}
	return nil
}

func (r *userRepository) List(ctx context.Context, offset, limit int) ([]*entity.User, int64, error) {
	var users []*entity.User
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.User{})
	
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, apperrors.CodeInternal, "failed to count users")
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, apperrors.CodeInternal, "failed to list users")
	}

	return users, total, nil
}

func (r *userRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	result := r.db.WithContext(ctx).Model(&entity.User{}).
		Where("id = ?", id).
		Update("password_hash", passwordHash)
	
	if result.Error != nil {
		return apperrors.Wrap(result.Error, apperrors.CodeInternal, "failed to update password")
	}
	if result.RowsAffected == 0 {
		return apperrors.New(apperrors.CodeUserNotFound, "user not found")
	}
	return nil
}

func (r *userRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&entity.User{}).
		Where("id = ?", id).
		Update("last_login_at", now)
	
	if result.Error != nil {
		return apperrors.Wrap(result.Error, apperrors.CodeInternal, "failed to update last login")
	}
	return nil
}

func (r *userRepository) VerifyEmail(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Model(&entity.User{}).
		Where("id = ?", id).
		Update("email_verified", true)
	
	if result.Error != nil {
		return apperrors.Wrap(result.Error, apperrors.CodeInternal, "failed to verify email")
	}
	if result.RowsAffected == 0 {
		return apperrors.New(apperrors.CodeUserNotFound, "user not found")
	}
	return nil
}

func (r *userRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entity.User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return false, apperrors.Wrap(err, apperrors.CodeInternal, "failed to check email")
	}
	return count > 0, nil
}

// refreshTokenRepository implements repository.RefreshTokenRepository
type refreshTokenRepository struct {
	db *Database
}

// NewRefreshTokenRepository creates a new refresh token repository
func NewRefreshTokenRepository(db *Database) repository.RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) Create(ctx context.Context, token *entity.RefreshToken) error {
	if err := r.db.WithContext(ctx).Create(token).Error; err != nil {
		return apperrors.Wrap(err, apperrors.CodeInternal, "failed to create refresh token")
	}
	return nil
}

func (r *refreshTokenRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*entity.RefreshToken, error) {
	var token entity.RefreshToken
	err := r.db.WithContext(ctx).First(&token, "token_hash = ?", tokenHash).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrTokenInvalid()
		}
		return nil, apperrors.Wrap(err, apperrors.CodeInternal, "failed to get refresh token")
	}
	return &token, nil
}

func (r *refreshTokenRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.RefreshToken, error) {
	var tokens []*entity.RefreshToken
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&tokens).Error; err != nil {
		return nil, apperrors.Wrap(err, apperrors.CodeInternal, "failed to get refresh tokens")
	}
	return tokens, nil
}

func (r *refreshTokenRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&entity.RefreshToken{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"revoked":    true,
			"revoked_at": now,
		})
	
	if result.Error != nil {
		return apperrors.Wrap(result.Error, apperrors.CodeInternal, "failed to revoke token")
	}
	return nil
}

func (r *refreshTokenRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	if err := r.db.WithContext(ctx).Model(&entity.RefreshToken{}).
		Where("user_id = ? AND revoked = ?", userID, false).
		Updates(map[string]interface{}{
			"revoked":    true,
			"revoked_at": now,
		}).Error; err != nil {
		return apperrors.Wrap(err, apperrors.CodeInternal, "failed to revoke tokens")
	}
	return nil
}

func (r *refreshTokenRepository) DeleteExpired(ctx context.Context) error {
	if err := r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&entity.RefreshToken{}).Error; err != nil {
		return apperrors.Wrap(err, apperrors.CodeInternal, "failed to delete expired tokens")
	}
	return nil
}

// passwordResetTokenRepository implements repository.PasswordResetTokenRepository
type passwordResetTokenRepository struct {
	db *Database
}

// NewPasswordResetTokenRepository creates a new password reset token repository
func NewPasswordResetTokenRepository(db *Database) repository.PasswordResetTokenRepository {
	return &passwordResetTokenRepository{db: db}
}

func (r *passwordResetTokenRepository) Create(ctx context.Context, token *entity.PasswordResetToken) error {
	if err := r.db.WithContext(ctx).Create(token).Error; err != nil {
		return apperrors.Wrap(err, apperrors.CodeInternal, "failed to create password reset token")
	}
	return nil
}

func (r *passwordResetTokenRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*entity.PasswordResetToken, error) {
	var token entity.PasswordResetToken
	err := r.db.WithContext(ctx).First(&token, "token_hash = ?", tokenHash).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrTokenInvalid()
		}
		return nil, apperrors.Wrap(err, apperrors.CodeInternal, "failed to get password reset token")
	}
	return &token, nil
}

func (r *passwordResetTokenRepository) GetLatestByUserID(ctx context.Context, userID uuid.UUID) (*entity.PasswordResetToken, error) {
	var token entity.PasswordResetToken
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND used = ?", userID, false).
		Order("created_at DESC").
		First(&token).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, apperrors.Wrap(err, apperrors.CodeInternal, "failed to get password reset token")
	}
	return &token, nil
}

func (r *passwordResetTokenRepository) MarkUsed(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&entity.PasswordResetToken{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"used":    true,
			"used_at": now,
		})
	
	if result.Error != nil {
		return apperrors.Wrap(result.Error, apperrors.CodeInternal, "failed to mark token as used")
	}
	return nil
}

func (r *passwordResetTokenRepository) DeleteExpired(ctx context.Context) error {
	if err := r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&entity.PasswordResetToken{}).Error; err != nil {
		return apperrors.Wrap(err, apperrors.CodeInternal, "failed to delete expired tokens")
	}
	return nil
}

func (r *passwordResetTokenRepository) InvalidateAllForUser(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	if err := r.db.WithContext(ctx).Model(&entity.PasswordResetToken{}).
		Where("user_id = ? AND used = ?", userID, false).
		Updates(map[string]interface{}{
			"used":    true,
			"used_at": now,
		}).Error; err != nil {
		return apperrors.Wrap(err, apperrors.CodeInternal, "failed to invalidate tokens")
	}
	return nil
}
