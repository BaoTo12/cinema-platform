package handler

import (
	"net/http"

	"cinemaos-backend/internal/app/auth"
	"cinemaos-backend/internal/middleware"
	"cinemaos-backend/internal/pkg/response"
	"cinemaos-backend/internal/pkg/validator"

	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication HTTP requests
type AuthHandler struct {
	authService *auth.Service
	validator   *validator.Validator
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *auth.Service, validator *validator.Validator) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		validator:   validator,
	}
}

// Register godoc
// @Summary Register a new user
// @Description Create a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body auth.RegisterRequest true "Registration details"
// @Success 201 {object} response.Response{data=auth.AuthResponse}
// @Failure 400 {object} response.Response
// @Failure 409 {object} response.Response
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req auth.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	// Validate request
	if errors := h.validator.Validate(req); errors != nil {
		response.ValidationError(c, errors)
		return
	}

	// Register user
	result, err := h.authService.Register(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}

	c.JSON(http.StatusCreated, response.Response{
		Success: true,
		Message: "Registration successful",
		Data:    result,
	})
}

// Login godoc
// @Summary Login user
// @Description Authenticate user and return tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body auth.LoginRequest true "Login credentials"
// @Success 200 {object} response.Response{data=auth.AuthResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req auth.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	if errors := h.validator.Validate(req); errors != nil {
		response.ValidationError(c, errors)
		return
	}

	result, err := h.authService.Login(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithMessage(c, "Login successful", result)
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Get a new access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body auth.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} response.Response{data=auth.TokenRefreshResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req auth.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	if errors := h.validator.Validate(req); errors != nil {
		response.ValidationError(c, errors)
		return
	}

	result, err := h.authService.RefreshToken(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, result)
}

// Logout godoc
// @Summary Logout user
// @Description Revoke refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body auth.RefreshTokenRequest true "Refresh token to revoke"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	var req auth.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	if err := h.authService.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithMessage(c, "Logged out successfully", nil)
}

// ForgotPassword godoc
// @Summary Request password reset
// @Description Send password reset email
// @Tags auth
// @Accept json
// @Produce json
// @Param request body auth.ForgotPasswordRequest true "Email address"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req auth.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	if errors := h.validator.Validate(req); errors != nil {
		response.ValidationError(c, errors)
		return
	}

	// Always return success to prevent email enumeration
	_ = h.authService.ForgotPassword(c.Request.Context(), req)

	response.SuccessWithMessage(c, "If the email exists, a password reset link has been sent", nil)
}

// ResetPassword godoc
// @Summary Reset password
// @Description Reset password using token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body auth.ResetPasswordRequest true "Reset token and new password"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req auth.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	if errors := h.validator.Validate(req); errors != nil {
		response.ValidationError(c, errors)
		return
	}

	if err := h.authService.ResetPassword(c.Request.Context(), req); err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithMessage(c, "Password has been reset successfully", nil)
}

// ChangePassword godoc
// @Summary Change password
// @Description Change password for authenticated user
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body auth.ChangePasswordRequest true "Current and new password"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /auth/change-password [post]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "Authentication required")
		return
	}

	var req auth.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	if errors := h.validator.Validate(req); errors != nil {
		response.ValidationError(c, errors)
		return
	}

	if err := h.authService.ChangePassword(c.Request.Context(), userID, req); err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithMessage(c, "Password changed successfully", nil)
}

// GetCurrentUser godoc
// @Summary Get current user
// @Description Get profile of authenticated user
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=auth.UserResponse}
// @Failure 401 {object} response.Response
// @Router /auth/me [get]
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "Authentication required")
		return
	}

	user, err := h.authService.GetCurrentUser(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, user)
}

// UpdateProfile godoc
// @Summary Update profile
// @Description Update authenticated user's profile
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body auth.UpdateProfileRequest true "Profile updates"
// @Success 200 {object} response.Response{data=auth.UserResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /auth/me [patch]
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "Authentication required")
		return
	}

	var req auth.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	if errors := h.validator.Validate(req); errors != nil {
		response.ValidationError(c, errors)
		return
	}

	user, err := h.authService.UpdateProfile(c.Request.Context(), userID, req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithMessage(c, "Profile updated successfully", user)
}
