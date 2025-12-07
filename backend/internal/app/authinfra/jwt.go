package authinfra

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"time"

	"cinemaos-backend/internal/config"
	apperrors "cinemaos-backend/internal/pkg/errors"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// TokenType represents the type of token
type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
	TokenTypeReset   TokenType = "reset"
	TokenTypeVerify  TokenType = "verify"
)

// Claims represents JWT claims
type Claims struct {
	UserID string    `json:"user_id"`
	Email  string    `json:"email"`
	Role   string    `json:"role"`
	Type   TokenType `json:"type"`
	jwt.RegisteredClaims
}

// JWTManager handles JWT operations
type JWTManager struct {
	accessSecret       []byte
	refreshSecret      []byte
	accessTokenExpiry  time.Duration
	refreshTokenExpiry time.Duration
	resetTokenExpiry   time.Duration
	verifyTokenExpiry  time.Duration
	issuer             string
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(cfg config.JWTConfig) *JWTManager {
	return &JWTManager{
		accessSecret:       []byte(cfg.AccessSecret),
		refreshSecret:      []byte(cfg.RefreshSecret),
		accessTokenExpiry:  cfg.AccessTokenExpiry,
		refreshTokenExpiry: cfg.RefreshTokenExpiry,
		resetTokenExpiry:   cfg.ResetTokenExpiry,
		verifyTokenExpiry:  cfg.VerifyTokenExpiry,
		issuer:             cfg.Issuer,
	}
}

// GenerateAccessToken generates an access token
func (m *JWTManager) GenerateAccessToken(userID uuid.UUID, email, role string) (string, error) {
	claims := Claims{
		UserID: userID.String(),
		Email:  email,
		Role:   role,
		Type:   TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.accessTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    m.issuer,
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.accessSecret)
}

// GenerateRefreshToken generates a refresh token
func (m *JWTManager) GenerateRefreshToken(userID uuid.UUID, email, role string) (string, error) {
	claims := Claims{
		UserID: userID.String(),
		Email:  email,
		Role:   role,
		Type:   TokenTypeRefresh,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.refreshTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    m.issuer,
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.refreshSecret)
}

// ValidateAccessToken validates an access token
func (m *JWTManager) ValidateAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return m.accessSecret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, apperrors.ErrTokenExpired()
		}
		return nil, apperrors.ErrTokenInvalid()
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, apperrors.ErrTokenInvalid()
	}

	if claims.Type != TokenTypeAccess {
		return nil, apperrors.ErrTokenInvalid()
	}

	return claims, nil
}

// ValidateRefreshToken validates a refresh token
func (m *JWTManager) ValidateRefreshToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return m.refreshSecret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, apperrors.ErrTokenExpired()
		}
		return nil, apperrors.ErrTokenInvalid()
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, apperrors.ErrTokenInvalid()
	}

	if claims.Type != TokenTypeRefresh {
		return nil, apperrors.ErrTokenInvalid()
	}

	return claims, nil
}

// GetAccessTokenExpiry returns access token expiry duration
func (m *JWTManager) GetAccessTokenExpiry() time.Duration {
	return m.accessTokenExpiry
}

// GetRefreshTokenExpiry returns refresh token expiry duration
func (m *JWTManager) GetRefreshTokenExpiry() time.Duration {
	return m.refreshTokenExpiry
}

// GetResetTokenExpiry returns reset token expiry duration
func (m *JWTManager) GetResetTokenExpiry() time.Duration {
	return m.resetTokenExpiry
}

// PasswordManager handles password operations
type PasswordManager struct {
	cost int
}

// NewPasswordManager creates a new password manager
func NewPasswordManager() *PasswordManager {
	return &PasswordManager{cost: bcrypt.DefaultCost}
}

// HashPassword hashes a password
func (m *PasswordManager) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), m.cost)
	if err != nil {
		return "", apperrors.Wrap(err, apperrors.CodeInternal, "failed to hash password")
	}
	return string(bytes), nil
}

// CheckPassword compares a password with a hash
func (m *PasswordManager) CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateRandomToken generates a cryptographically secure random token
func GenerateRandomToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// HashToken hashes a token for storage
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// GenerateBookingReference generates a unique booking reference
func GenerateBookingReference() string {
	date := time.Now().Format("20060102")
	random, _ := GenerateRandomToken(4)
	return "BK" + date + random[:6]
}

// GeneratePaymentReference generates a unique payment reference
func GeneratePaymentReference() string {
	timestamp := time.Now().Unix()
	random, _ := GenerateRandomToken(4)
	return "PAY" + string(rune(timestamp)) + random[:6]
}
