package services

import (
	"context"
	"errors"
	"time"

	"connectrpc.com/connect"
	"github.com/cinemaos/backend/internal/database"
	"github.com/cinemaos/backend/internal/middleware"
	"github.com/cinemaos/backend/internal/models"
	"github.com/cinemaos/backend/internal/utils"
	cinemav1 "github.com/cinemaos/backend/proto/cinema/v1"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuthService struct{}

func NewAuthService() *AuthService {
	return &AuthService{}
}

func (s *AuthService) Register(
	ctx context.Context,
	req *connect.Request[cinemav1.RegisterRequest],
) (*connect.Response[cinemav1.RegisterResponse], error) {
	// Validate input
	if req.Msg.Email == "" || req.Msg.Password == "" ||
		req.Msg.FirstName == "" || req.Msg.LastName == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("missing required fields"))
	}

	// Check if user already exists
	var existingUser models.User
	result := database.DB.Where("email = ?", req.Msg.Email).First(&existingUser)
	if result.Error == nil {
		return nil, connect.NewError(connect.CodeAlreadyExists, errors.New("user with this email already exists"))
	}

	// Hash password
	passwordHash, err := utils.HashPassword(req.Msg.Password)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Create user
	user := models.User{
		Email:        req.Msg.Email,
		PasswordHash: passwordHash,
		FirstName:    req.Msg.FirstName,
		LastName:     req.Msg.LastName,
		Phone:        req.Msg.Phone,
		Role:         models.RoleCustomer,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Prepare response
	resp := &cinemav1.RegisterResponse{
		Success: true,
		Message: "Registration successful",
		User: &cinemav1.User{
			Id:            user.ID.String(),
			Email:         user.Email,
			FirstName:     user.FirstName,
			LastName:      user.LastName,
			Phone:         user.Phone,
			Role:          string(user.Role),
			EmailVerified: user.EmailVerified,
			CreatedAt:     user.CreatedAt.Format(time.RFC3339),
		},
	}

	return connect.NewResponse(resp), nil
}

func (s *AuthService) Login(
	ctx context.Context,
	req *connect.Request[cinemav1.LoginRequest],
) (*connect.Response[cinemav1.LoginResponse], error) {
	// Find user by email
	var user models.User
	if err := database.DB.Where("email = ?", req.Msg.Email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("invalid email or password"))
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Check if user is active
	if !user.IsActive {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("account is deactivated"))
	}

	// Verify password
	if !utils.CheckPassword(req.Msg.Password, user.PasswordHash) {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("invalid email or password"))
	}

	// Generate tokens
	accessToken, err := utils.GenerateAccessToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Store refresh token in database
	token := models.RefreshToken{
		UserID:    user.ID,
		TokenHash: refreshToken,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	if err := database.DB.Create(&token).Error; err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Update last login
	database.DB.Model(&user).Update("last_login_at", time.Now())

	// Prepare response
	resp := &cinemav1.LoginResponse{
		Success:      true,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    900, // 15 minutes in seconds
		User: &cinemav1.User{
			Id:        user.ID.String(),
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Phone:     user.Phone,
			Role:      string(user.Role),
		},
	}

	return connect.NewResponse(resp), nil
}

func (s *AuthService) RefreshToken(
	ctx context.Context,
	req *connect.Request[cinemav1.RefreshTokenRequest],
) (*connect.Response[cinemav1.RefreshTokenResponse], error) {
	// Validate refresh token
	claims, err := utils.ValidateRefreshToken(req.Msg.RefreshToken)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("invalid or expired refresh token"))
	}

	// Check if token exists in database and is not revoked
	var storedToken models.RefreshToken
	result := database.DB.Where("token_hash = ? AND user_id = ? AND revoked = ? AND expires_at > ?",
		req.Msg.RefreshToken, claims.UserID, false, time.Now()).First(&storedToken)

	if result.Error != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("invalid or expired refresh token"))
	}

	// Get user
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("user not found"))
	}

	if !user.IsActive {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("account is deactivated"))
	}

	// Generate new access token
	accessToken, err := utils.GenerateAccessToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	resp := &cinemav1.RefreshTokenResponse{
		Success:     true,
		AccessToken: accessToken,
		ExpiresIn:   900,
	}

	return connect.NewResponse(resp), nil
}

func (s *AuthService) Logout(
	ctx context.Context,
	req *connect.Request[cinemav1.LogoutRequest],
) (*connect.Response[cinemav1.LogoutResponse], error) {
	// Get user from context
	userCtx, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, nil)
	}

	// Revoke refresh token
	database.DB.Model(&models.RefreshToken{}).
		Where("token_hash = ? AND user_id = ?", req.Msg.RefreshToken, userCtx.UserID).
		Update("revoked", true)

	resp := &cinemav1.LogoutResponse{
		Success: true,
		Message: "Logged out successfully",
	}

	return connect.NewResponse(resp), nil
}

func (s *AuthService) GetCurrentUser(
	ctx context.Context,
	req *connect.Request[cinemav1.GetCurrentUserRequest],
) (*connect.Response[cinemav1.GetCurrentUserResponse], error) {
	// Get user from context
	userCtx, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, nil)
	}

	// Get full user details
	userID, err := uuid.Parse(userCtx.UserID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("user not found"))
	}

	resp := &cinemav1.GetCurrentUserResponse{
		User: &cinemav1.User{
			Id:            user.ID.String(),
			Email:         user.Email,
			FirstName:     user.FirstName,
			LastName:      user.LastName,
			Phone:         user.Phone,
			Role:          string(user.Role),
			EmailVerified: user.EmailVerified,
			CreatedAt:     user.CreatedAt.Format(time.RFC3339),
		},
	}

	return connect.NewResponse(resp), nil
}
