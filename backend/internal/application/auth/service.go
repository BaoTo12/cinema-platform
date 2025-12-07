package auth

import (
	"context"
	"time"

	"cinemaos-backend/internal/domain/entity"
	"cinemaos-backend/internal/domain/repository"
	"cinemaos-backend/internal/infrastructure/auth"
	apperrors "cinemaos-backend/internal/pkg/errors"
	"cinemaos-backend/internal/pkg/logger"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Service handles authentication business logic
type Service struct {
	userRepo       repository.UserRepository
	refreshRepo    repository.RefreshTokenRepository
	resetTokenRepo repository.PasswordResetTokenRepository
	jwtManager     *auth.JWTManager
	passwordMgr    *auth.PasswordManager
	logger         *logger.Logger
	frontendURL    string
}

// NewService creates a new auth service
func NewService(
	userRepo repository.UserRepository,
	refreshRepo repository.RefreshTokenRepository,
	resetTokenRepo repository.PasswordResetTokenRepository,
	jwtManager *auth.JWTManager,
	passwordMgr *auth.PasswordManager,
	logger *logger.Logger,
	frontendURL string,
) *Service {
	return &Service{
		userRepo:       userRepo,
		refreshRepo:    refreshRepo,
		resetTokenRepo: resetTokenRepo,
		jwtManager:     jwtManager,
		passwordMgr:    passwordMgr,
		logger:         logger,
		frontendURL:    frontendURL,
	}
}

// Register registers a new user
func (s *Service) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	log := s.logger.WithContext(ctx)

	// Check if email already exists
	exists, err := s.userRepo.EmailExists(ctx, req.Email)
	if err != nil {
		log.Error("failed to check email existence", zap.Error(err))
		return nil, err
	}
	if exists {
		return nil, apperrors.ErrEmailExists()
	}

	// Hash password
	passwordHash, err := s.passwordMgr.HashPassword(req.Password)
	if err != nil {
		log.Error("failed to hash password", zap.Error(err))
		return nil, err
	}

	// Create user
	var phone *string
	if req.Phone != "" {
		phone = &req.Phone
	}

	user := &entity.User{
		Email:        req.Email,
		PasswordHash: passwordHash,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Phone:        phone,
		Role:         entity.RoleCustomer,
		IsActive:     true,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		log.Error("failed to create user", zap.Error(err))
		return nil, err
	}

	log.Info("user registered successfully")

	// Generate tokens
	return s.generateAuthResponse(ctx, user)
}

// Login authenticates a user
func (s *Service) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	log := s.logger.WithContext(ctx)

	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if apperrors.Is(err, apperrors.CodeUserNotFound) {
			return nil, apperrors.ErrInvalidCredentials()
		}
		log.Error("failed to get user", zap.Error(err))
		return nil, err
	}

	// Check if account is active
	if !user.IsActive {
		return nil, apperrors.ErrAccountDisabled()
	}

	// Verify password
	if !s.passwordMgr.CheckPassword(req.Password, user.PasswordHash) {
		return nil, apperrors.ErrInvalidCredentials()
	}

	// Update last login
	if err := s.userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		log.Warn("failed to update last login")
	}

	log.Info("user logged in successfully")

	// Generate tokens
	return s.generateAuthResponse(ctx, user)
}

// RefreshToken refreshes an access token
func (s *Service) RefreshToken(ctx context.Context, req RefreshTokenRequest) (*TokenRefreshResponse, error) {
	log := s.logger.WithContext(ctx)

	// Validate refresh token
	claims, err := s.jwtManager.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, err
	}

	// Hash the token and check in database
	tokenHash := auth.HashToken(req.RefreshToken)
	storedToken, err := s.refreshRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, err
	}

	// Verify token is valid
	if !storedToken.IsValid() {
		return nil, apperrors.ErrTokenExpired()
	}

	// Get user
	userID, _ := uuid.Parse(claims.UserID)
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if !user.IsActive {
		return nil, apperrors.ErrAccountDisabled()
	}

	// Generate new access token
	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		log.Error("failed to generate access token", zap.Error(err))
		return nil, apperrors.ErrInternal("failed to generate token")
	}

	return &TokenRefreshResponse{
		AccessToken: accessToken,
		ExpiresIn:   int64(s.jwtManager.GetAccessTokenExpiry().Seconds()),
		TokenType:   "Bearer",
	}, nil
}

// Logout revokes a refresh token
func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	tokenHash := auth.HashToken(refreshToken)
	
	storedToken, err := s.refreshRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		// Token not found is ok for logout
		return nil
	}

	return s.refreshRepo.Revoke(ctx, storedToken.ID)
}

// LogoutAll revokes all refresh tokens for a user
func (s *Service) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	return s.refreshRepo.RevokeAllForUser(ctx, userID)
}

// ForgotPassword initiates password reset
func (s *Service) ForgotPassword(ctx context.Context, req ForgotPasswordRequest) error {
	log := s.logger.WithContext(ctx)

	// Get user by email (don't reveal if email exists)
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if apperrors.Is(err, apperrors.CodeUserNotFound) {
			// Don't reveal that email doesn't exist
			log.Info("forgot password requested for non-existent email")
			return nil
		}
		return err
	}

	// Invalidate existing reset tokens
	if err := s.resetTokenRepo.InvalidateAllForUser(ctx, user.ID); err != nil {
		log.Warn("failed to invalidate existing reset tokens")
	}

	// Generate reset token
	token, err := auth.GenerateRandomToken(32)
	if err != nil {
		log.Error("failed to generate reset token", zap.Error(err))
		return apperrors.ErrInternal("failed to generate token")
	}

	// Store hashed token
	resetToken := &entity.PasswordResetToken{
		UserID:    user.ID,
		TokenHash: auth.HashToken(token),
		ExpiresAt: time.Now().Add(s.jwtManager.GetResetTokenExpiry()),
	}

	if err := s.resetTokenRepo.Create(ctx, resetToken); err != nil {
		log.Error("failed to store reset token", zap.Error(err))
		return err
	}

	// TODO: Send email with reset link
	// resetLink := fmt.Sprintf("%s/reset-password?token=%s", s.frontendURL, token)
	// emailService.SendPasswordResetEmail(user.Email, user.FirstName, resetLink)

	log.Info("password reset token generated")
	return nil
}

// ResetPassword resets user password with token
func (s *Service) ResetPassword(ctx context.Context, req ResetPasswordRequest) error {
	log := s.logger.WithContext(ctx)

	// Get token by hash
	tokenHash := auth.HashToken(req.Token)
	resetToken, err := s.resetTokenRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return err
	}

	// Verify token is valid
	if !resetToken.IsValid() {
		return apperrors.ErrTokenExpired()
	}

	// Hash new password
	passwordHash, err := s.passwordMgr.HashPassword(req.NewPassword)
	if err != nil {
		log.Error("failed to hash password", zap.Error(err))
		return err
	}

	// Update password
	if err := s.userRepo.UpdatePassword(ctx, resetToken.UserID, passwordHash); err != nil {
		return err
	}

	// Mark token as used
	if err := s.resetTokenRepo.MarkUsed(ctx, resetToken.ID); err != nil {
		log.Warn("failed to mark reset token as used")
	}

	// Revoke all refresh tokens for security
	if err := s.refreshRepo.RevokeAllForUser(ctx, resetToken.UserID); err != nil {
		log.Warn("failed to revoke refresh tokens")
	}

	log.Info("password reset successfully")
	return nil
}

// ChangePassword changes password for authenticated user
func (s *Service) ChangePassword(ctx context.Context, userID uuid.UUID, req ChangePasswordRequest) error {
	log := s.logger.WithContext(ctx)

	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	// Verify current password
	if !s.passwordMgr.CheckPassword(req.CurrentPassword, user.PasswordHash) {
		return apperrors.New(apperrors.CodeBadRequest, "current password is incorrect")
	}

	// Hash new password
	passwordHash, err := s.passwordMgr.HashPassword(req.NewPassword)
	if err != nil {
		log.Error("failed to hash password", zap.Error(err))
		return err
	}

	// Update password
	if err := s.userRepo.UpdatePassword(ctx, userID, passwordHash); err != nil {
		return err
	}

	// Revoke all refresh tokens for security
	if err := s.refreshRepo.RevokeAllForUser(ctx, userID); err != nil {
		log.Warn("failed to revoke refresh tokens")
	}

	log.Info("password changed successfully")
	return nil
}

// GetCurrentUser returns the current user
func (s *Service) GetCurrentUser(ctx context.Context, userID uuid.UUID) (*UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return toUserResponse(user), nil
}

// UpdateProfile updates user profile
func (s *Service) UpdateProfile(ctx context.Context, userID uuid.UUID, req UpdateProfileRequest) (*UserResponse, error) {
	log := s.logger.WithContext(ctx)

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.FirstName != "" {
		user.FirstName = req.FirstName
	}
	if req.LastName != "" {
		user.LastName = req.LastName
	}
	if req.Phone != "" {
		user.Phone = &req.Phone
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		log.Error("failed to update user", zap.Error(err))
		return nil, err
	}

	log.Info("profile updated successfully")
	return toUserResponse(user), nil
}

// generateAuthResponse generates auth response with tokens
func (s *Service) generateAuthResponse(ctx context.Context, user *entity.User) (*AuthResponse, error) {
	// Generate access token
	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		return nil, apperrors.ErrInternal("failed to generate access token")
	}

	// Generate refresh token
	refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		return nil, apperrors.ErrInternal("failed to generate refresh token")
	}

	// Store refresh token
	tokenEntity := &entity.RefreshToken{
		UserID:    user.ID,
		TokenHash: auth.HashToken(refreshToken),
		ExpiresAt: time.Now().Add(s.jwtManager.GetRefreshTokenExpiry()),
	}

	if err := s.refreshRepo.Create(ctx, tokenEntity); err != nil {
		return nil, apperrors.ErrInternal("failed to store refresh token")
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.jwtManager.GetAccessTokenExpiry().Seconds()),
		TokenType:    "Bearer",
		User:         *toUserResponse(user),
	}, nil
}

// toUserResponse converts entity to response DTO
func toUserResponse(user *entity.User) *UserResponse {
	return &UserResponse{
		ID:            user.ID.String(),
		Email:         user.Email,
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		FullName:      user.FullName(),
		Phone:         user.Phone,
		Role:          string(user.Role),
		EmailVerified: user.EmailVerified,
		CreatedAt:     user.CreatedAt,
		LastLoginAt:   user.LastLoginAt,
	}
}
