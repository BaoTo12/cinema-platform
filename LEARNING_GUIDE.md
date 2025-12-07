# 🎓 CinemaOS Backend - Complete Go Web Development Guide

> **A comprehensive, beginner-friendly guide to Go web development**  
> Learn modern backend development by exploring a real production-grade cinema booking system.

---

## 📋 Table of Contents

1. [Prerequisites & Go Fundamentals](#-prerequisites--go-fundamentals)
2. [Project Structure (golang-standards)](#-project-structure)
3. [Deep Dive: How Each Layer Works](#-deep-dive-how-each-layer-works)
4. [Difficult Concepts Explained](#-difficult-concepts-explained)
5. [Request Lifecycle (Step-by-Step)](#-request-lifecycle-step-by-step)
6. [Code Patterns You'll See Everywhere](#-code-patterns-youll-see-everywhere)
7. [Common Mistakes & How to Avoid Them](#-common-mistakes--how-to-avoid-them)
8. [How to Run](#-how-to-run)
9. [Exercises with Solutions](#-exercises-with-solutions)
10. [Next Steps](#-next-steps)

---

## 🧱 Prerequisites & Go Fundamentals

### What You Should Know First

Before diving into this codebase, understand these Go basics:

```go
// 1. VARIABLES & TYPES
var name string = "John"    // Explicit type
age := 25                   // Type inference (shorthand)
var price float64           // Zero value: 0.0

// 2. POINTERS - Very important in Go!
x := 10
ptr := &x    // ptr holds the ADDRESS of x
fmt.Println(*ptr)  // *ptr = value at that address = 10

// WHY POINTERS? To avoid copying large data or to modify the original
func updateName(user *User) {  // Receives pointer, modifies original
    user.Name = "Updated"
}

// 3. STRUCTS - Like classes but simpler
type User struct {
    ID        int
    Email     string
    CreatedAt time.Time
}

// 4. METHODS - Functions attached to structs
func (u *User) FullName() string {  // Method on *User
    return u.FirstName + " " + u.LastName
}

// 5. INTERFACES - Define behavior contracts
type Repository interface {
    GetByID(id int) (*User, error)
    Create(user *User) error
}
// ANY struct with these methods satisfies the interface

// 6. ERROR HANDLING - No exceptions in Go!
result, err := doSomething()
if err != nil {
    return nil, err  // Always check and handle errors
}

// 7. GOROUTINES & CHANNELS (for async) - Optional for this project
go func() { /* runs concurrently */ }()
```

### Why Go for Web Development?

| Feature | Benefit |
|---------|---------|
| **Fast compilation** | Instant feedback while developing |
| **Static typing** | Catches errors before runtime |
| **Built-in concurrency** | Handle thousands of requests easily |
| **Single binary** | Deploy anywhere without dependencies |
| **Garbage collected** | Memory managed automatically |

---

## 📁 Project Structure

This project follows [golang-standards/project-layout](https://github.com/golang-standards/project-layout):

```
backend/
├── cmd/api/main.go        # 👈 START HERE - Entry point
│
├── internal/              # Private code (Go enforces this!)
│   ├── app/              # Business logic
│   │   ├── entity/       # Data models (User, Movie, Booking)
│   │   ├── repository/   # Interfaces for data access
│   │   ├── auth/         # Auth service + DTOs
│   │   ├── authinfra/    # JWT & password utilities
│   │   ├── movie/        # Movie service
│   │   ├── cinema/       # Cinema service
│   │   ├── showtime/     # Showtime service
│   │   ├── postgres/     # Database implementations
│   │   └── redis/        # Cache client
│   │
│   ├── config/           # Configuration loading CODE
│   ├── handler/          # HTTP request handlers
│   ├── middleware/       # Auth, CORS, logging, response time
│   ├── router/           # Route definitions
│   ├── server/           # HTTP server setup
│   └── pkg/              # Shared utilities
│
├── configs/              # Configuration FILES (yaml)
├── build/package/        # Dockerfile
├── deployments/          # docker-compose.yml
└── go.mod                # Module definition
```

### Why This Structure?

| Directory | Purpose | Can Others Import? |
|-----------|---------|-------------------|
| `cmd/` | Main applications | No (entry points) |
| `internal/` | Private code | ❌ Go enforces this! |
| `pkg/` | Public libraries | ✅ Anyone can import |
| `configs/` | Config files | N/A (not code) |

---

## 🔬 Deep Dive: How Each Layer Works

### Layer 1: Entities (`internal/app/entity/`)

**What are entities?** Go structs that represent your data. They map directly to database tables.

```go
// internal/app/entity/user.go

type User struct {
    // Column definitions using GORM tags
    ID           uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
    Email        string         `gorm:"uniqueIndex;not null"`
    PasswordHash string         `gorm:"not null" json:"-"`  // json:"-" = NEVER expose in API
    FirstName    string         `gorm:"not null"`
    Role         Role           `gorm:"type:varchar(20);default:'CUSTOMER'"`
    IsActive     bool           `gorm:"default:true"`
    CreatedAt    time.Time
    UpdatedAt    time.Time
    DeletedAt    gorm.DeletedAt `gorm:"index"`  // Soft delete (see explanation below)
}
```

**Understanding the Tags:**

| Tag | Meaning |
|-----|---------|
| `gorm:"primary_key"` | This is the primary key |
| `gorm:"uniqueIndex"` | Create a unique index (no duplicates) |
| `gorm:"not null"` | Required field |
| `gorm:"default:'value'"` | Default value in database |
| `json:"-"` | **Never include in JSON output** (security!) |
| `json:"email"` | JSON key name when serializing |

**What is Soft Delete?**
```go
DeletedAt gorm.DeletedAt `gorm:"index"`
```
Instead of permanently deleting records, we set `DeletedAt` to a timestamp. GORM automatically excludes these records from queries. Useful for:
- Audit trails
- Recovering accidentally deleted data
- Compliance requirements

---

### Layer 2: Repository Interfaces (`internal/app/repository/`)

**Why interfaces?** They define WHAT operations exist without HOW they're implemented.

```go
// internal/app/repository/user_repository.go

type UserRepository interface {
    Create(ctx context.Context, user *entity.User) error
    GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
    GetByEmail(ctx context.Context, email string) (*entity.User, error)
    Update(ctx context.Context, user *entity.User) error
    Delete(ctx context.Context, id uuid.UUID) error
    EmailExists(ctx context.Context, email string) (bool, error)
}
```

**Why Context in Every Method?**
```go
ctx context.Context  // First parameter, always
```
Context carries:
- **Request cancellation** - If user closes browser, we can stop the query
- **Timeouts** - Prevent queries from hanging forever
- **Request-scoped values** - User ID, trace ID, etc.

**The Power of Interfaces:**
```go
// Your service doesn't know or care about PostgreSQL
type AuthService struct {
    userRepo repository.UserRepository  // Just an interface!
}

// In tests, you can inject a mock:
type MockUserRepo struct{}
func (m *MockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
    return &entity.User{Email: "test@test.com"}, nil  // Fake data
}
```

---

### Layer 3: Repository Implementation (`internal/app/postgres/`)

**This is WHERE we actually talk to the database:**

```go
// internal/app/postgres/user_repository.go

type userRepository struct {
    db *Database  // Holds GORM connection
}

// Constructor - returns the INTERFACE type, not the struct!
func NewUserRepository(db *Database) repository.UserRepository {
    return &userRepository{db: db}
}

// GetByEmail implements repository.UserRepository
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
    var user entity.User
    
    // GORM query chain:
    err := r.db.WithContext(ctx).       // 1. Use request context
        Where("email = ?", email).       // 2. WHERE clause (parameterized!)
        First(&user).                    // 3. Get first matching row
        Error                            // 4. Get any error
    
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            // Convert GORM error to our custom error
            return nil, apperrors.ErrUserNotFound()
        }
        return nil, err
    }
    
    return &user, nil
}
```

**GORM Query Patterns:**
```go
// SELECT * FROM users WHERE id = ?
db.First(&user, id)

// SELECT * FROM users WHERE email = ?
db.Where("email = ?", email).First(&user)

// SELECT * FROM users WHERE role = ? AND is_active = true
db.Where("role = ? AND is_active = ?", "ADMIN", true).Find(&users)

// INSERT INTO users (...)
db.Create(&user)

// UPDATE users SET ... WHERE id = ?
db.Save(&user)

// Soft DELETE (sets deleted_at)
db.Delete(&user, id)

// Preload relationships (JOIN)
db.Preload("Movie").Preload("Cinema").First(&showtime, id)
```

---

### Layer 4: Services (`internal/app/auth/`, `internal/app/movie/`, etc.)

**Services contain BUSINESS LOGIC** - the rules of your application.

```go
// internal/app/auth/service.go

type Service struct {
    userRepo       repository.UserRepository       // Dependencies
    refreshRepo    repository.RefreshTokenRepository
    jwtManager     *authinfra.JWTManager
    passwordMgr    *authinfra.PasswordManager
    logger         *logger.Logger
    frontendURL    string
}

// Constructor with Dependency Injection
func NewService(
    userRepo repository.UserRepository,
    refreshRepo repository.RefreshTokenRepository,
    jwtManager *authinfra.JWTManager,
    // ... more dependencies
) *Service {
    return &Service{
        userRepo:    userRepo,
        refreshRepo: refreshRepo,
        jwtManager:  jwtManager,
        // ...
    }
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
    log := s.logger.WithContext(ctx)  // Include request ID in logs
    
    // BUSINESS RULE 1: Email must be unique
    exists, err := s.userRepo.EmailExists(ctx, req.Email)
    if err != nil {
        log.Error("failed to check email", zap.Error(err))
        return nil, err
    }
    if exists {
        return nil, apperrors.ErrEmailExists()
    }
    
    // BUSINESS RULE 2: Password must be hashed
    passwordHash, err := s.passwordMgr.HashPassword(req.Password)
    if err != nil {
        log.Error("failed to hash password", zap.Error(err))
        return nil, err
    }
    
    // Create domain entity
    user := &entity.User{
        Email:        req.Email,
        PasswordHash: passwordHash,
        FirstName:    req.FirstName,
        LastName:     req.LastName,
        Role:         entity.RoleCustomer,  // Default role
        IsActive:     true,
    }
    
    // BUSINESS RULE 3: Persist to database
    if err := s.userRepo.Create(ctx, user); err != nil {
        log.Error("failed to create user", zap.Error(err))
        return nil, err
    }
    
    log.Info("user registered successfully")
    
    // BUSINESS RULE 4: Generate JWT tokens
    return s.generateAuthResponse(ctx, user)
}
```

**Key Observations:**
1. Services don't know about HTTP, SQL, or any infrastructure
2. All dependencies are interfaces - easy to test
3. Error handling at every step
4. Logging for debugging

---

### Layer 5: DTOs (Data Transfer Objects)

**DTOs separate your API contract from internal entities:**

```go
// internal/app/auth/dto.go

// Request DTO - what clients SEND
type RegisterRequest struct {
    Email     string `json:"email" validate:"required,email"`
    Password  string `json:"password" validate:"required,password"`
    FirstName string `json:"first_name" validate:"required,min=2,max=50"`
    LastName  string `json:"last_name" validate:"required,min=2,max=50"`
    Phone     string `json:"phone,omitempty" validate:"omitempty,phone"`
}

// Response DTO - what we RETURN
type UserResponse struct {
    ID            string     `json:"id"`
    Email         string     `json:"email"`
    FirstName     string     `json:"first_name"`
    LastName      string     `json:"last_name"`
    FullName      string     `json:"full_name"`
    Phone         *string    `json:"phone,omitempty"`  // Pointer = optional
    Role          string     `json:"role"`
    EmailVerified bool       `json:"email_verified"`
    // Note: PasswordHash is NOT here - security!
}
```

**Validation Tags Explained:**
| Tag | Meaning |
|-----|---------|
| `required` | Field must be present |
| `email` | Must be valid email format |
| `min=2,max=50` | String length between 2-50 |
| `omitempty` | Skip if empty |
| `password` | Custom validator (see below) |

---

### Layer 6: HTTP Handlers (`internal/handler/`)

**Handlers translate HTTP ↔ Service calls:**

```go
// internal/handler/auth_handler.go

type AuthHandler struct {
    authService *auth.Service
    validator   *validator.Validator
}

func (h *AuthHandler) Register(c *gin.Context) {
    // STEP 1: Parse JSON body into DTO
    var req auth.RegisterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.BadRequest(c, "Invalid request body")
        return
    }
    
    // STEP 2: Validate the request
    if errors := h.validator.Validate(req); errors != nil {
        response.ValidationError(c, errors)  // Returns 400 with details
        return
    }
    
    // STEP 3: Call the service (business logic)
    result, err := h.authService.Register(c.Request.Context(), req)
    if err != nil {
        response.Error(c, err)  // Converts error to proper HTTP response
        return
    }
    
    // STEP 4: Return success response
    response.Created(c, result)  // 201 Created
}
```

**Handler Principles:**
- Handlers are **thin** - no business logic here
- They only: Parse → Validate → Call Service → Respond
- **Always pass context**: `c.Request.Context()` carries cancellation and values

---

### Layer 7: Middleware (`internal/middleware/`)

**Middleware runs BEFORE and/or AFTER handlers:**

```go
// Authentication Middleware
func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
    return func(c *gin.Context) {
        // === BEFORE HANDLER ===
        
        // 1. Extract token from header
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            response.Unauthorized(c, "Missing authorization header")
            c.Abort()  // STOP the chain - handler won't run
            return
        }
        
        // 2. Parse "Bearer <token>"
        token := strings.TrimPrefix(authHeader, "Bearer ")
        
        // 3. Validate JWT
        claims, err := m.jwtManager.ValidateAccessToken(token)
        if err != nil {
            response.Unauthorized(c, "Invalid token")
            c.Abort()
            return
        }
        
        // 4. Store user info in request context
        c.Set("user_id", claims.UserID)
        c.Set("user_email", claims.Email)
        c.Set("user_role", claims.Role)
        
        c.Next()  // Continue to handler
        
        // === AFTER HANDLER ===
        // (can log response, cleanup, etc.)
    }
}

// Response Time Middleware (Exercise 3)
func ResponseTimeMiddleware(log *logger.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()  // Record start time
        
        c.Next()  // Run handler
        
        duration := time.Since(start)  // Calculate duration
        
        log.Info("request completed",
            zap.String("method", c.Request.Method),
            zap.String("path", c.Request.URL.Path),
            zap.Int("status", c.Writer.Status()),
            zap.Duration("duration", duration),
        )
    }
}
```

**Middleware Execution Order:**
```
Request → Recovery → RequestID → Logging → ResponseTime → CORS → Auth → HANDLER
                                                                         ↓
Response ← Recovery ← RequestID ← Logging ← ResponseTime ← CORS ← Auth ← HANDLER
```

---

## 🧩 Difficult Concepts Explained

### 1. Why Pointers (`*`) Everywhere?

```go
// WITHOUT pointer - creates a COPY
func UpdateUser(user User) {
    user.Name = "New Name"  // ONLY modifies the copy!
}

// WITH pointer - modifies ORIGINAL
func UpdateUser(user *User) {
    user.Name = "New Name"  // Modifies the original
}

// GORM requires pointers for struct methods
func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
    var user entity.User
    err := r.db.First(&user, id).Error  // &user = "put result HERE"
    return &user, err  // Return pointer to avoid copying
}
```

**Rules:**
- Use `*Type` for large structs (User, Movie, etc.)
- Use value types for small types (int, string, bool)
- Repository/Service methods usually return pointers

### 2. What is `context.Context`?

```go
// Context carries request-scoped data
func (s *Service) GetUser(ctx context.Context, id uuid.UUID) (*User, error) {
    // ctx contains:
    // - Cancellation signal (user closed browser)
    // - Timeout (request should complete within 30s)
    // - Values (user_id, request_id, etc.)
    
    // Pass to all downstream calls
    return s.repo.GetByID(ctx, id)
}

// In handler, get context from HTTP request
func (h *Handler) GetUser(c *gin.Context) {
    ctx := c.Request.Context()  // HTTP request context
    // ...
}
```

### 3. Error Handling Pattern

Go doesn't have exceptions. Every function that can fail returns an error:

```go
// The pattern you'll see 1000 times:
result, err := doSomething()
if err != nil {
    // Handle the error
    return nil, err          // Propagate up
    // OR
    return nil, fmt.Errorf("context: %w", err)  // Add context
}
// Use result safely

// Custom errors for HTTP responses
func ErrUserNotFound() error {
    return &AppError{
        Code:    CodeNotFound,      // Maps to 404
        Message: "User not found",
    }
}

// In handler, errors become HTTP responses
response.Error(c, err)  // Checks error type, returns proper status code
```

### 4. Dependency Injection

Instead of creating dependencies inside, we INJECT them:

```go
// ❌ BAD - Tightly coupled, hard to test
type UserService struct {}

func (s *UserService) GetUser(id int) *User {
    db := database.Connect()  // Creates own dependency
    return db.Find(id)
}

// ✅ GOOD - Loosely coupled, testable
type UserService struct {
    repo repository.UserRepository  // Injected dependency
}

func NewUserService(repo repository.UserRepository) *UserService {
    return &UserService{repo: repo}
}

// In main.go, wire everything:
db := postgres.New(cfg.Database)
userRepo := postgres.NewUserRepository(db)
userService := NewUserService(userRepo)  // Inject!
```

### 5. Interface Satisfaction (Implicit)

Go interfaces are satisfied IMPLICITLY - no `implements` keyword:

```go
// Interface definition
type Repository interface {
    GetByID(ctx context.Context, id uuid.UUID) (*User, error)
}

// This struct satisfies Repository (has all required methods)
type PostgresRepo struct { db *gorm.DB }

func (r *PostgresRepo) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
    // implementation
}

// Compiler checks at assignment time
var repo Repository = &PostgresRepo{}  // Works!
```

---

## 🔄 Request Lifecycle (Step-by-Step)

Let's trace `POST /api/v1/auth/register`:

```
┌─────────────────────────────────────────────────────────────────────┐
│  CLIENT sends:                                                       │
│  POST /api/v1/auth/register                                         │
│  Content-Type: application/json                                      │
│  {"email": "john@example.com", "password": "Secret123!", ...}       │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│  1. MIDDLEWARE CHAIN EXECUTES (in order)                            │
│                                                                      │
│  RecoveryMiddleware: Sets up panic recovery                          │
│  RequestIDMiddleware: Generates X-Request-ID header                  │
│  LoggingMiddleware: Logs request start                               │
│  ResponseTimeMiddleware: Starts timer                               │
│  CORSMiddleware: Adds CORS headers                                   │
│  RateLimiter: Checks rate limits                                     │
│                                                                      │
│  (No AuthMiddleware - this is a public route)                       │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│  2. ROUTER matches: POST /api/v1/auth/register                      │
│     → Calls: authHandler.Register(c *gin.Context)                   │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│  3. HANDLER: auth_handler.go                                        │
│                                                                      │
│  a) Parse JSON body:                                                 │
│     c.ShouldBindJSON(&req) → RegisterRequest struct                  │
│                                                                      │
│  b) Validate:                                                        │
│     validator.Validate(req) → Check email format, password rules    │
│                                                                      │
│  c) Call service:                                                    │
│     authService.Register(ctx, req)                                   │
│                                                                      │
│  d) Return response:                                                 │
│     response.Created(c, result)                                      │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│  4. SERVICE: auth/service.go                                        │
│                                                                      │
│  a) Check email exists:                                              │
│     userRepo.EmailExists(ctx, email) → false                        │
│                                                                      │
│  b) Hash password:                                                   │
│     passwordMgr.HashPassword(password) → "$2a$10$..."               │
│                                                                      │
│  c) Create user:                                                     │
│     userRepo.Create(ctx, user) → INSERT INTO users                  │
│                                                                      │
│  d) Generate tokens:                                                 │
│     jwtManager.GenerateAccessToken(...)                              │
│     jwtManager.GenerateRefreshToken(...)                             │
│                                                                      │
│  e) Store refresh token:                                             │
│     refreshRepo.Create(ctx, token)                                   │
│                                                                      │
│  f) Return AuthResponse                                              │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│  5. REPOSITORY: postgres/user_repository.go                         │
│                                                                      │
│  EmailExists: SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)    │
│  Create:      INSERT INTO users (...) VALUES (...)                  │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│  6. MIDDLEWARE CHAIN COMPLETES (reverse order)                      │
│                                                                      │
│  ResponseTimeMiddleware: Logs "request completed, duration: 45ms"   │
│  LoggingMiddleware: Logs response status                            │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│  CLIENT receives:                                                    │
│  HTTP 201 Created                                                    │
│  X-Request-ID: uuid                                                  │
│  X-Response-Time: 45.123ms                                          │
│  {                                                                   │
│    "success": true,                                                  │
│    "data": {                                                         │
│      "access_token": "eyJhbGciOiJIUzI1NiIs...",                      │
│      "refresh_token": "eyJhbGciOiJIUzI1NiIs...",                     │
│      "expires_in": 900,                                              │
│      "user": { "id": "...", "email": "john@example.com", ... }      │
│    }                                                                 │
│  }                                                                   │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 🔧 Code Patterns You'll See Everywhere

### 1. Constructor Pattern

```go
// Always create instances through constructors
func NewService(repo Repository, logger *Logger) *Service {
    return &Service{
        repo:   repo,
        logger: logger,
    }
}

// Never: service := &Service{} (missing dependencies)
```

### 2. Options Pattern (for optional configs)

```go
type ServerOption func(*Server)

func WithTimeout(d time.Duration) ServerOption {
    return func(s *Server) {
        s.timeout = d
    }
}

// Usage: NewServer(addr, WithTimeout(30*time.Second))
```

### 3. Error Wrapping

```go
// Add context to errors
if err != nil {
    return fmt.Errorf("failed to create user: %w", err)
}

// Later, check original error
if errors.Is(err, gorm.ErrRecordNotFound) { ... }
```

---

## ⚠️ Common Mistakes & How to Avoid Them

| Mistake | Fix |
|---------|-----|
| Forgetting to check `err` | Always handle: `if err != nil { return err }` |
| Not using `ctx` | Pass context through all layers |
| Exposing passwords in JSON | Use `json:"-"` tag |
| SQL injection | Use parameterized queries: `Where("email = ?", email)` |
| Hardcoding config | Use environment variables + Viper |
| Creating dependencies inside | Inject them through constructors |

---

## 🚀 How to Run

```bash
# 1. Start database and Redis
cd backend
docker-compose -f deployments/docker-compose.yml up -d

# 2. Configure environment
cp .env.example .env
# Edit .env with your settings

# 3. Run the application
go run ./cmd/api

# 4. Test the API
curl http://localhost:8080/health
```

---

## 🏋️ Exercises with Solutions

### Exercise 1: Phone Field ✅
Already implemented! See: `internal/app/entity/user.go` line 27

### Exercise 2: Get Movie Showtimes ✅
Implemented as `GET /api/v1/movies/:id/showtimes`
- Repository: `internal/app/repository/booking_repository.go` (GetByMovieID)
- Implementation: `internal/app/postgres/showtime_repository.go`
- Service: `internal/app/showtime/service.go` (GetShowtimesByMovieID)
- Handler: `internal/handler/movie_handler.go` (GetShowtimes)
- Route: `internal/router/router.go`

### Exercise 3: Response Time Middleware ✅
Implemented in `internal/middleware/responsetime.go`

---

## 📚 Next Steps

1. **Build the Booking Module** - Create seat selection, booking, and payment
2. **Add Tests** - Unit tests for services, integration tests for handlers
3. **Add Swagger** - API documentation with `swag init`
4. **Deploy** - Use the Dockerfile to deploy to AWS/GCP/Digital Ocean

---

**Happy Coding! 🎬**
