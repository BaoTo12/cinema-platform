# 🎓 CinemaOS Backend - Complete Learning Guide

> **For Go Web Development Beginners**  
> This guide explains the project's architecture and code in detail to help you learn modern backend development with Go.

---

## 📚 Table of Contents

1. [Project Overview](#-project-overview)
2. [How to Run the Project](#-how-to-run-the-project)
3. [Understanding the Architecture](#-understanding-the-architecture)
4. [Layer-by-Layer Deep Dive](#-layer-by-layer-deep-dive)
5. [Request Flow: How a Request Travels](#-request-flow-how-a-request-travels)
6. [Key Libraries Used](#-key-libraries-used)
7. [Studying the Code (Recommended Order)](#-studying-the-code-recommended-order)
8. [Common Patterns Explained](#-common-patterns-explained)
9. [Exercises for Practice](#-exercises-for-practice)

---

## 🌟 Project Overview

**CinemaOS** is a cinema booking platform backend built with Go. It demonstrates:

- **Clean/Hexagonal Architecture**: Separating business logic from infrastructure
- **RESTful API Design**: Using Gin framework
- **Database Access**: PostgreSQL with GORM ORM
- **Authentication**: JWT-based auth with refresh tokens
- **Caching**: Redis for session management
- **Observability**: Structured logging (Zap) + distributed tracing (OpenTelemetry)

### What You'll Learn

| Concept | Files to Study |
|---------|---------------|
| Go project structure | `cmd/api/main.go`, `internal/` folder |
| Database modeling | `internal/domain/entity/*.go` |
| Repository pattern | `internal/domain/repository/*.go` |
| Service/Use Case layer | `internal/application/*/service.go` |
| HTTP handlers | `internal/interfaces/http/handler/*.go` |
| Middleware | `internal/interfaces/http/middleware/*.go` |
| Dependency Injection | `cmd/api/main.go` |

---

## 🚀 How to Run the Project

### Prerequisites

- **Go 1.21+** installed
- **Docker** (for PostgreSQL and Redis)

### Steps

1. **Clone and navigate:**
   ```bash
   cd backend
   ```

2. **Install dependencies:**
   ```bash
   go mod download
   ```

3. **Start infrastructure (Docker):**
   ```bash
   docker-compose up -d
   ```
   This starts PostgreSQL, Redis, and Jaeger (tracing).

4. **Configure environment:**
   ```bash
   cp .env.example .env
   # Edit .env with your database credentials if needed
   ```

5. **Run the application:**
   ```bash
   go run ./cmd/api
   ```

6. **Test the API:**
   ```bash
   curl http://localhost:8080/health
   ```

---

## 🏗️ Understanding the Architecture

This project uses **Clean Architecture** (also called Hexagonal Architecture). The key principle is:

> **Dependencies point INWARD** - outer layers depend on inner layers, never the reverse.

```
┌────────────────────────────────────────────────────────────────┐
│                      INTERFACES LAYER                          │
│   (HTTP handlers, middleware, gRPC, CLI - how external         │
│    systems communicate with our app)                           │
├────────────────────────────────────────────────────────────────┤
│                      APPLICATION LAYER                          │
│   (Services/Use Cases - business logic orchestration)          │
├────────────────────────────────────────────────────────────────┤
│                       DOMAIN LAYER                              │
│   (Entities + Repository Interfaces - the core business models)│
├────────────────────────────────────────────────────────────────┤
│                    INFRASTRUCTURE LAYER                         │
│   (Database, cache, external APIs - technical implementations) │
└────────────────────────────────────────────────────────────────┘
```

### Project Structure Mapped to Architecture

```
backend/
├── cmd/api/main.go                 ← Entry point & Dependency Injection
│
├── internal/
│   ├── domain/                     ← DOMAIN LAYER (innermost)
│   │   ├── entity/                 ← Business entities (User, Movie, etc.)
│   │   └── repository/             ← Repository interfaces (contracts)
│   │
│   ├── application/                ← APPLICATION LAYER
│   │   ├── auth/                   ← Auth service + DTOs
│   │   ├── movie/                  ← Movie service + DTOs
│   │   ├── cinema/                 ← Cinema service + DTOs
│   │   └── showtime/               ← Showtime service + DTOs
│   │
│   ├── infrastructure/             ← INFRASTRUCTURE LAYER
│   │   ├── persistence/postgres/   ← PostgreSQL repository implementations
│   │   ├── persistence/redis/      ← Redis client
│   │   └── auth/                   ← JWT & password utilities
│   │
│   ├── interfaces/                 ← INTERFACES LAYER (outermost)
│   │   └── http/                   ← HTTP-specific code
│   │       ├── handler/            ← HTTP request handlers
│   │       ├── middleware/         ← Auth, CORS, logging middleware
│   │       ├── router/             ← Route definitions
│   │       └── server.go           ← HTTP server configuration
│   │
│   └── pkg/                        ← Shared utilities (logger, validator, etc.)
│
└── config/                         ← Configuration loading
```

---

## 🔍 Layer-by-Layer Deep Dive

### 1. Domain Layer (`internal/domain/`)

This is the **core** of your application. It contains:

- **Entities**: Pure Go structs representing business objects
- **Repository Interfaces**: Contracts that define what data operations are needed

#### 📁 `internal/domain/entity/user.go`

```go
// User is a DOMAIN ENTITY - it represents a business concept
type User struct {
    ID           uuid.UUID `gorm:"type:uuid;primary_key;..."`
    Email        string    `gorm:"uniqueIndex;not null"`
    PasswordHash string    `gorm:"not null" json:"-"`  // json:"-" hides it from API responses
    FirstName    *string   // Pointer = optional field (can be NULL in database)
    Role         UserRole  `gorm:"type:varchar(20);default:'CUSTOMER'"`
    IsActive     bool      `gorm:"default:true"`
    CreatedAt    time.Time
    UpdatedAt    time.Time
    DeletedAt    gorm.DeletedAt `gorm:"index"` // Soft delete support
}
```

**Key Learnings:**
- `gorm:` tags tell GORM how to map fields to database columns
- `json:` tags control JSON serialization (what the API returns)
- Pointer types (`*string`) represent optional/nullable fields
- `gorm.DeletedAt` enables soft deletes (data isn't actually deleted)

#### 📁 `internal/domain/repository/user_repository.go`

```go
// UserRepository is an INTERFACE - a contract
// It defines WHAT operations are needed, not HOW they're implemented
type UserRepository interface {
    Create(ctx context.Context, user *entity.User) error
    GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
    GetByEmail(ctx context.Context, email string) (*entity.User, error)
    Update(ctx context.Context, user *entity.User) error
    Delete(ctx context.Context, id uuid.UUID) error
}
```

**Why interfaces?**
- The Application layer works with interfaces, not implementations
- You can swap implementations (e.g., PostgreSQL → MongoDB) without changing business logic
- Makes testing easy (you can create mock implementations)

---

### 2. Infrastructure Layer (`internal/infrastructure/`)

This layer **implements** the repository interfaces defined in Domain.

#### 📁 `internal/infrastructure/persistence/postgres/user_repository.go`

```go
// userRepository IMPLEMENTS repository.UserRepository
type userRepository struct {
    db *Database  // Database connection wrapper
}

// NewUserRepository creates a new user repository
// Returns the INTERFACE type, not the struct
func NewUserRepository(db *Database) repository.UserRepository {
    return &userRepository{db: db}
}

// GetByEmail implements repository.UserRepository.GetByEmail
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
    var user entity.User
    
    // GORM query: SELECT * FROM users WHERE email = ?
    err := r.db.WithContext(ctx).
        Where("email = ?", email).
        First(&user).Error
    
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, apperrors.New(apperrors.CodeNotFound, "user not found")
        }
        return nil, apperrors.Wrap(err, apperrors.CodeInternal, "failed to get user")
    }
    
    return &user, nil
}
```

**Key Learnings:**
- The function returns an interface (`repository.UserRepository`), but creates a concrete struct
- `db.WithContext(ctx)` passes the request context for timeout/cancellation support
- Custom error wrapping provides consistent error handling

---

### 3. Application Layer (`internal/application/`)

This layer contains **business logic**. Each module (auth, movie, cinema) has:
- `service.go`: The main logic
- `dto.go`: Data Transfer Objects (request/response structs)

#### 📁 `internal/application/auth/service.go`

```go
// Service holds dependencies needed for auth operations
type Service struct {
    userRepo        repository.UserRepository      // Interface, not concrete type!
    refreshTokenRepo repository.RefreshTokenRepository
    jwtManager      *infraauth.JWTManager
    passwordManager *infraauth.PasswordManager
    logger          *logger.Logger
}

// NewService creates a new auth service with injected dependencies
func NewService(
    userRepo repository.UserRepository,
    refreshTokenRepo repository.RefreshTokenRepository,
    jwtManager *infraauth.JWTManager,
    passwordManager *infraauth.PasswordManager,
    logger *logger.Logger,
    frontendURL string,
) *Service {
    return &Service{
        userRepo:         userRepo,
        refreshTokenRepo: refreshTokenRepo,
        jwtManager:       jwtManager,
        passwordManager:  passwordManager,
        logger:           logger,
    }
}

// Register creates a new user account
func (s *Service) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
    // 1. Check if user already exists
    existing, _ := s.userRepo.GetByEmail(ctx, req.Email)
    if existing != nil {
        return nil, apperrors.New(apperrors.CodeConflict, "email already registered")
    }
    
    // 2. Hash the password (never store plain text!)
    hashedPassword, err := s.passwordManager.HashPassword(req.Password)
    if err != nil {
        return nil, err
    }
    
    // 3. Create the user entity
    user := &entity.User{
        Email:        req.Email,
        PasswordHash: hashedPassword,
        FirstName:    &req.FirstName,
        LastName:     &req.LastName,
        Role:         entity.UserRoleCustomer,
    }
    
    // 4. Save to database via repository
    if err := s.userRepo.Create(ctx, user); err != nil {
        return nil, err
    }
    
    // 5. Generate JWT tokens
    accessToken, err := s.jwtManager.GenerateAccessToken(user)
    refreshToken, err := s.jwtManager.GenerateRefreshToken(user)
    
    // 6. Return response
    return &AuthResponse{
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
        User:         toUserResponse(user),
    }, nil
}
```

**Key Learnings:**
- Services depend on **interfaces**, not concrete implementations
- Dependencies are **injected** through the constructor
- Business rules are enforced here (e.g., "email must be unique")
- The service doesn't know about HTTP, databases, or frameworks

#### 📁 `internal/application/auth/dto.go`

```go
// DTOs (Data Transfer Objects) define the shape of API requests/responses

// RegisterRequest is what the client sends
type RegisterRequest struct {
    Email     string `json:"email" validate:"required,email"`
    Password  string `json:"password" validate:"required,min=8,password"`
    FirstName string `json:"first_name" validate:"required,min=2"`
    LastName  string `json:"last_name" validate:"required,min=2"`
}

// AuthResponse is what we send back
type AuthResponse struct {
    AccessToken  string       `json:"access_token"`
    RefreshToken string       `json:"refresh_token"`
    User         UserResponse `json:"user"`
}
```

**Why DTOs?**
- Separate internal entity structure from API contract
- Add validation rules via struct tags (`validate:"..."`)
- Control exactly what data is exposed to clients

---

### 4. Interfaces Layer (`internal/interfaces/`)

This layer handles **communication with the outside world** (HTTP in our case).

#### 📁 `internal/interfaces/http/handler/auth_handler.go`

```go
// AuthHandler handles HTTP requests for auth endpoints
type AuthHandler struct {
    service   *auth.Service    // Depends on APPLICATION layer
    validator *validator.Validator
}

// Register handles POST /auth/register
func (h *AuthHandler) Register(c *gin.Context) {
    // 1. Parse JSON request body into DTO
    var req auth.RegisterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.BadRequest(c, "Invalid request body")
        return
    }
    
    // 2. Validate the request
    if validationErrors := h.validator.Validate(req); validationErrors != nil {
        response.ValidationError(c, validationErrors)
        return
    }
    
    // 3. Call the service (business logic)
    result, err := h.service.Register(c.Request.Context(), req)
    if err != nil {
        response.Error(c, err)  // Error handler maps errors to HTTP status codes
        return
    }
    
    // 4. Send success response
    response.Created(c, result)
}
```

**Key Learnings:**
- Handlers are **thin** - they only:
  1. Parse requests
  2. Validate input
  3. Call the service
  4. Format the response
- All business logic is in the **service**, not the handler
- `c.Request.Context()` passes the request context for cancellation/timeout

#### 📁 `internal/interfaces/http/router/router.go`

```go
// Router holds all route dependencies
type Router struct {
    cfg            *config.Config
    authMiddleware *middleware.AuthMiddleware
    authHandler    *handler.AuthHandler
    movieHandler   *handler.MovieHandler
    cinemaHandler  *handler.CinemaHandler
    showtimeHandler *handler.ShowtimeHandler
}

// Setup configures all routes
func (r *Router) Setup() *gin.Engine {
    router := gin.New()
    
    // Global middleware (applied to all routes)
    router.Use(middleware.RecoveryMiddleware(r.logger))
    router.Use(middleware.RequestIDMiddleware())
    router.Use(middleware.LoggingMiddleware(r.logger))
    
    // API routes
    v1 := router.Group("/api/v1")
    {
        // Auth routes (public)
        auth := v1.Group("/auth")
        {
            auth.POST("/register", r.authHandler.Register)
            auth.POST("/login", r.authHandler.Login)
            
            // Protected routes (require authentication)
            auth.POST("/logout", r.authMiddleware.Authenticate(), r.authHandler.Logout)
            auth.GET("/me", r.authMiddleware.Authenticate(), r.authHandler.GetCurrentUser)
        }
        
        // Movie routes
        movies := v1.Group("/movies")
        {
            movies.GET("", r.movieHandler.List)  // Public: anyone can list movies
            
            // Admin only routes
            movies.POST("", r.authMiddleware.Authenticate(), r.authMiddleware.RequireAdmin(), r.movieHandler.Create)
        }
    }
    
    return router
}
```

**Key Learnings:**
- Routes are organized by resource (`/auth`, `/movies`, `/cinemas`)
- Middleware can be applied globally or to specific routes
- `authMiddleware.Authenticate()` protects routes that require login
- `authMiddleware.RequireAdmin()` adds role-based access control

---

### 5. Entry Point & Dependency Injection

#### 📁 `cmd/api/main.go`

This is where everything comes together:

```go
func main() {
    // 1. Load configuration
    cfg, err := config.Load(configPath)
    
    // 2. Initialize infrastructure
    db, err := postgres.New(cfg.Database, log)       // Database
    redisClient, err := redis.New(cfg.Redis, log)    // Cache
    
    // 3. Create repositories (INFRASTRUCTURE implements DOMAIN interfaces)
    userRepo := postgres.NewUserRepository(db)
    movieRepo := postgres.NewMovieRepository(db)
    cinemaRepo := postgres.NewCinemaRepository(db)
    showtimeRepo := postgres.NewShowtimeRepository(db)
    
    // 4. Create infrastructure services
    jwtManager := infraauth.NewJWTManager(cfg.JWT)
    passwordManager := infraauth.NewPasswordManager()
    
    // 5. Create application services (inject dependencies)
    authService := authapp.NewService(userRepo, refreshRepo, jwtManager, passwordManager, log)
    movieService := movieapp.NewService(movieRepo, log)
    cinemaService := cinemaapp.NewService(cinemaRepo, screenRepo, seatRepo, log)
    showtimeService := showtimeapp.NewService(showtimeRepo, movieRepo, cinemaRepo, screenRepo, log)
    
    // 6. Create handlers (inject services)
    authHandler := handler.NewAuthHandler(authService, requestValidator)
    movieHandler := handler.NewMovieHandler(movieService, requestValidator)
    
    // 7. Create middleware
    authMiddleware := middleware.NewAuthMiddleware(jwtManager, log)
    
    // 8. Create router (inject handlers and middleware)
    appRouter := router.NewRouter(cfg, log, authMiddleware, authHandler, movieHandler, ...)
    
    // 9. Start server
    srv := httpserver.NewServer(cfg.Server, appRouter.Setup(), log)
    srv.Start()
}
```

**This is Dependency Injection:**
- Each component receives its dependencies through its constructor
- The `main` function is the **composition root** - it wires everything together
- No component creates its own dependencies

---

## 🔄 Request Flow: How a Request Travels

Let's trace a `POST /api/v1/auth/register` request:

```
1. HTTP Request arrives
   ↓
2. Global Middleware runs:
   - RecoveryMiddleware (catches panics)
   - RequestIDMiddleware (adds unique ID for tracing)
   - LoggingMiddleware (logs the request)
   ↓
3. Router matches URL → auth.POST("/register")
   ↓
4. AuthHandler.Register() called
   - Parses JSON body into RegisterRequest
   - Validates input
   - Calls authService.Register()
   ↓
5. AuthService.Register() executes
   - Checks if email exists (userRepo.GetByEmail)
   - Hashes password (passwordManager.HashPassword)
   - Creates user (userRepo.Create)
   - Generates tokens (jwtManager.GenerateAccessToken)
   ↓
6. UserRepository.Create() runs
   - Executes INSERT SQL via GORM
   - Returns success or error
   ↓
7. Response flows back up
   - Service returns AuthResponse
   - Handler calls response.Created()
   - JSON response sent to client
```

---

## 📦 Key Libraries Used

| Library | Purpose | Where Used |
|---------|---------|------------|
| [Gin](https://github.com/gin-gonic/gin) | HTTP web framework | `interfaces/http/` |
| [GORM](https://gorm.io) | ORM for database access | `infrastructure/persistence/` |
| [Zap](https://github.com/uber-go/zap) | Structured logging | `pkg/logger/` |
| [Viper](https://github.com/spf13/viper) | Configuration management | `config/` |
| [go-playground/validator](https://github.com/go-playground/validator) | Input validation | `pkg/validator/` |
| [golang-jwt/jwt](https://github.com/golang-jwt/jwt) | JWT tokens | `infrastructure/auth/` |
| [google/uuid](https://github.com/google/uuid) | UUID generation | `domain/entity/` |

---

## 📖 Studying the Code (Recommended Order)

### Week 1: Foundation

1. **Start with entities** - Read all files in `internal/domain/entity/`
   - Understand how data is modeled
   - Notice relationships between entities

2. **Study repository interfaces** - `internal/domain/repository/`
   - See what operations are defined
   - Understand the contract pattern

3. **Read one repository implementation** - `internal/infrastructure/persistence/postgres/user_repository.go`
   - See how interfaces are implemented
   - Learn GORM query patterns

### Week 2: Business Logic

4. **Read auth service** - `internal/application/auth/`
   - Study `service.go` for business logic
   - Look at `dto.go` for request/response shapes

5. **Read movie service** - `internal/application/movie/`
   - Compare patterns with auth service

### Week 3: HTTP Layer

6. **Read handlers** - `internal/interfaces/http/handler/`
   - See how HTTP requests are processed
   - Notice the thin handler pattern

7. **Read middleware** - `internal/interfaces/http/middleware/`
   - Understand authentication flow
   - See how logging/recovery work

8. **Read router** - `internal/interfaces/http/router/router.go`
   - See how routes are organized
   - Understand middleware application

### Week 4: Infrastructure

9. **Read main.go** - `cmd/api/main.go`
   - Understand dependency injection
   - See how everything connects

10. **Read configuration** - `config/`
    - Learn about environment variables
    - Understand configuration patterns

---

## 🎯 Common Patterns Explained

### Pattern 1: Constructor Injection

```go
// Instead of creating dependencies inside the struct:
// ❌ BAD
type Service struct {
    repo *UserRepository  // concrete type
}
func (s *Service) DoSomething() {
    s.repo = NewUserRepository()  // creates its own dependency
}

// Inject dependencies through constructor:
// ✅ GOOD
type Service struct {
    repo repository.UserRepository  // interface type
}
func NewService(repo repository.UserRepository) *Service {
    return &Service{repo: repo}  // dependency injected from outside
}
```

### Pattern 2: Interface for Testability

```go
// Define interface in DOMAIN layer
type UserRepository interface {
    GetByEmail(ctx context.Context, email string) (*User, error)
}

// Implement in INFRASTRUCTURE layer
type postgresUserRepo struct { db *Database }
func (r *postgresUserRepo) GetByEmail(ctx context.Context, email string) (*User, error) {
    // Real database call
}

// For testing, create a mock
type mockUserRepo struct { users map[string]*User }
func (r *mockUserRepo) GetByEmail(ctx context.Context, email string) (*User, error) {
    if user, ok := r.users[email]; ok {
        return user, nil
    }
    return nil, errors.New("not found")
}
```

### Pattern 3: Error Wrapping

```go
// Create domain-specific errors
var ErrNotFound = errors.New("not found")
var ErrConflict = errors.New("conflict")

// Wrap with context
func (r *userRepo) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
    err := r.db.First(&user, id).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, apperrors.New(apperrors.CodeNotFound, "user not found")
    }
    return nil, apperrors.Wrap(err, apperrors.CodeInternal, "database error")
}
```

### Pattern 4: Context Propagation

```go
// Always pass context through the call chain
func (h *Handler) GetUser(c *gin.Context) {
    ctx := c.Request.Context()  // Get context from HTTP request
    user, err := h.service.GetByID(ctx, id)
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
    return s.repo.GetByID(ctx, id)  // Pass context to repository
}

func (r *Repo) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
    return r.db.WithContext(ctx).First(&user, id).Error  // Use context in query
}
```

---

## 🏋️ Exercises for Practice

### Exercise 1: Add a New Field

Add a `phone_number` field to the User entity.

1. Edit `internal/domain/entity/user.go`
2. Edit `internal/application/auth/dto.go` (add to RegisterRequest)
3. Edit `internal/application/auth/service.go` (map field in Register)
4. Run migrations: `go run ./cmd/api` (AutoMigrate handles it)

### Exercise 2: Add a New Endpoint

Add `GET /api/v1/movies/:id/showtimes` to list showtimes for a movie.

1. Add method to `ShowtimeRepository` interface
2. Implement in `postgres/showtime_repository.go`
3. Add method to `ShowtimeService`
4. Add handler in `ShowtimeHandler`
5. Add route in `router.go`

### Exercise 3: Add Validation

Add validation to ensure `base_price` in showtimes is between 5.00 and 100.00.

1. Edit `internal/application/showtime/dto.go`
2. Add `validate:"min=5,max=100"` tag

### Exercise 4: Create a New Module

Create a simple "Feedback" module where users can submit feedback.

Follow the pattern:
1. Create `internal/domain/entity/feedback.go`
2. Create interface in `internal/domain/repository/feedback_repository.go`
3. Implement in `internal/infrastructure/persistence/postgres/feedback_repository.go`
4. Create `internal/application/feedback/service.go` and `dto.go`
5. Create `internal/interfaces/http/handler/feedback_handler.go`
6. Wire in `cmd/api/main.go` and `router.go`

---

## 🎉 You Made It!

By understanding this architecture, you now know:

- ✅ How to structure a professional Go backend
- ✅ Why Clean Architecture matters
- ✅ How to separate concerns properly
- ✅ How to use dependency injection
- ✅ How to make code testable and maintainable

**Next Steps:**
- Implement the Booking module (the most complex one!)
- Add unit tests for services
- Deploy to a cloud provider

Happy coding! 🚀
