# 🎓 CinemaOS Backend - Complete Learning Guide

> **A comprehensive guide for Go web development beginners**  
> Learn modern backend development by exploring a real production-grade cinema booking system.

---

## 📋 Table of Contents

1. [What You'll Learn](#-what-youll-learn)
2. [Project Structure (golang-standards)](#-project-structure)
3. [Understanding Each Directory](#-understanding-each-directory)
4. [How the Code Works](#-how-the-code-works)
5. [Request Lifecycle](#-request-lifecycle)
6. [Key Concepts Explained](#-key-concepts-explained)
7. [Code Walkthrough](#-code-walkthrough)
8. [How to Run](#-how-to-run)
9. [Exercises](#-exercises)

---

## 🎯 What You'll Learn

By studying this codebase, you'll master:

| Skill | Where to Find It |
|-------|------------------|
| Go project structure | Root folder organization |
| HTTP routing with Gin | `internal/router/` |
| Middleware patterns | `internal/middleware/` |
| Database with GORM | `internal/app/postgres/` |
| JWT Authentication | `internal/app/authinfra/` |
| Clean Architecture | `internal/app/` organization |
| Configuration management | `internal/config/` |
| Dependency injection | `cmd/api/main.go` |

---

## 📁 Project Structure

This project follows the [golang-standards/project-layout](https://github.com/golang-standards/project-layout):

```
backend/
├── cmd/                    # Application entry points
│   └── api/
│       └── main.go        # 👈 START HERE - wires everything together
│
├── internal/               # Private application code
│   ├── app/               # Business logic (services, entities, repos)
│   │   ├── entity/        # Domain models (User, Movie, Booking)
│   │   ├── repository/    # Repository interfaces
│   │   ├── auth/          # Auth service + DTOs
│   │   ├── authinfra/     # JWT & password utilities
│   │   ├── movie/         # Movie service
│   │   ├── cinema/        # Cinema service
│   │   ├── showtime/      # Showtime service
│   │   ├── postgres/      # Database implementations
│   │   └── redis/         # Cache client
│   │
│   ├── config/            # Configuration loader (Go code)
│   ├── handler/           # HTTP request handlers
│   ├── middleware/        # Auth, CORS, logging middleware
│   ├── router/            # Route definitions
│   ├── server/            # HTTP server setup
│   └── pkg/               # Shared utilities
│       ├── logger/        # Structured logging (Zap)
│       ├── validator/     # Input validation
│       ├── response/      # Standardized API responses
│       ├── errors/        # Custom error types
│       └── tracer/        # Distributed tracing
│
├── api/                    # API definitions
│   └── proto/             # Protocol Buffer files (gRPC)
│
├── configs/                # Configuration FILES (yaml, json)
│   └── config.yaml        # Default configuration
│
├── build/                  # Build & packaging
│   └── package/
│       └── Dockerfile     # Container image
│
├── deployments/            # Deployment configs
│   └── docker-compose.yml # Local development stack
│
├── scripts/                # Build/deploy scripts
├── test/                   # Integration tests
├── go.mod                  # Go module definition
└── Makefile               # Common commands
```

---

## 📂 Understanding Each Directory

### `/cmd` - Entry Points

**What is it?** The `cmd` directory contains your application's `main.go` files.

**Rule:** Each subdirectory becomes a separate executable. The directory name becomes the binary name.

```
cmd/
└── api/
    └── main.go    → builds to `api.exe` or `api`
```

**What `main.go` does:**
1. Loads configuration
2. Creates database connections
3. Creates all repositories, services, and handlers
4. Wires everything together (Dependency Injection)
5. Starts the HTTP server

---

### `/internal` - Private Code

**What is it?** Code that ONLY this project can import. Go enforces this!

**Why?** Prevents external packages from depending on your internal implementation.

```go
// ❌ This import would FAIL from another project:
import "cinemaos-backend/internal/app/auth"

// ✅ This would work (if we put it in /pkg):
import "cinemaos-backend/pkg/utils"
```

---

### `/internal/app` - Business Logic

This is where your **domain logic** lives, organized by feature:

```
app/
├── entity/        # Data structures (what things ARE)
├── repository/    # Interfaces (what operations exist)
├── auth/          # Auth feature (service + DTOs)
├── movie/         # Movie feature
├── cinema/        # Cinema feature
├── showtime/      # Showtime feature
├── postgres/      # HOW we store data (implementation)
└── redis/         # HOW we cache data
```

**Key principle:** Services depend on INTERFACES, not implementations:

```go
// ✅ GOOD - depends on interface
type MovieService struct {
    repo repository.MovieRepository  // Interface!
}

// ❌ BAD - depends on concrete type
type MovieService struct {
    repo *postgres.MovieRepositoryImpl  // Ties you to PostgreSQL!
}
```

---

### `/configs` - Configuration Files

**What is it?** Template configuration files (YAML, JSON, TOML).

**vs `/internal/config`:** 
- `/configs/config.yaml` = The actual config FILE
- `/internal/config/config.go` = Go CODE that loads and parses it

```yaml
# configs/config.yaml
app:
  name: CinemaOS
  environment: development

server:
  port: 8080

database:
  host: localhost
  port: 5432
```

```go
// internal/config/config.go
type Config struct {
    App    AppConfig    `mapstructure:"app"`
    Server ServerConfig `mapstructure:"server"`
}
```

---

## 🔧 How the Code Works

### 1. Configuration (`internal/config/config.go`)

Configuration is loaded using **Viper**, a popular Go config library:

```go
// What this does:
// 1. Sets default values
// 2. Reads config.yaml file
// 3. Overrides with environment variables
// 4. Returns a typed Config struct

func Load(configPath string) (*Config, error) {
    v := viper.New()
    
    // Set defaults
    v.SetDefault("server.port", 8080)
    
    // Read file
    v.SetConfigFile(configPath)
    v.ReadInConfig()
    
    // Environment variables override file
    v.SetEnvPrefix("CINEMAOS")
    v.AutomaticEnv()
    
    // Parse into struct
    var cfg Config
    v.Unmarshal(&cfg)
    return &cfg, nil
}
```

**Key learning:** Environment variables win over config files (important for Docker/Kubernetes).

---

### 2. Entities (`internal/app/entity/`)

Entities are your **data models** - Go structs that map to database tables:

```go
// internal/app/entity/user.go
type User struct {
    ID           uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    Email        string         `gorm:"uniqueIndex;not null"`
    PasswordHash string         `gorm:"not null" json:"-"`  // json:"-" hides from API
    FirstName    string         `gorm:"not null"`
    LastName     string         `gorm:"not null"`
    Role         UserRole       `gorm:"type:varchar(20);default:'CUSTOMER'"`
    IsActive     bool           `gorm:"default:true"`
    CreatedAt    time.Time
    UpdatedAt    time.Time
    DeletedAt    gorm.DeletedAt `gorm:"index"`  // Soft delete
}

// This method is called by your code
func (u *User) FullName() string {
    return u.FirstName + " " + u.LastName
}
```

**Struct tags explained:**
- `gorm:"..."` - Instructions for the database ORM
- `json:"..."` - Instructions for JSON encoding
- `validate:"..."` - Instructions for validation

---

### 3. Repository Interface (`internal/app/repository/`)

Repositories define **WHAT** data operations exist (not HOW):

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

**Why interfaces?**
1. **Testability:** Mock the interface in tests
2. **Flexibility:** Swap PostgreSQL for MongoDB without changing services
3. **Clean code:** Services don't know about SQL

---

### 4. Repository Implementation (`internal/app/postgres/`)

This is **HOW** we actually store data:

```go
// internal/app/postgres/user_repository.go
type userRepository struct {
    db *Database
}

// Constructor returns INTERFACE type, not struct
func NewUserRepository(db *Database) repository.UserRepository {
    return &userRepository{db: db}
}

// GetByEmail implements the interface method
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
    var user entity.User
    
    // GORM query - translates to: SELECT * FROM users WHERE email = ?
    err := r.db.WithContext(ctx).
        Where("email = ?", email).
        First(&user).Error
    
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, apperrors.ErrUserNotFound()
        }
        return nil, err
    }
    
    return &user, nil
}
```

---

### 5. Service/Application Layer (`internal/app/auth/`)

Services contain **business logic** - the rules of your application:

```go
// internal/app/auth/service.go
type Service struct {
    userRepo       repository.UserRepository      // Interfaces!
    refreshRepo    repository.RefreshTokenRepository
    jwtManager     *authinfra.JWTManager
    passwordMgr    *authinfra.PasswordManager
    logger         *logger.Logger
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
    // 1. Business rule: email must be unique
    exists, _ := s.userRepo.EmailExists(ctx, req.Email)
    if exists {
        return nil, apperrors.ErrEmailExists()
    }
    
    // 2. Security: hash the password
    hash, err := s.passwordMgr.HashPassword(req.Password)
    if err != nil {
        return nil, err
    }
    
    // 3. Create user entity
    user := &entity.User{
        Email:        req.Email,
        PasswordHash: hash,
        FirstName:    req.FirstName,
        LastName:     req.LastName,
    }
    
    // 4. Persist to database
    if err := s.userRepo.Create(ctx, user); err != nil {
        return nil, err
    }
    
    // 5. Generate JWT tokens
    accessToken, _ := s.jwtManager.GenerateAccessToken(user.ID, user.Email, string(user.Role))
    
    return &AuthResponse{
        AccessToken: accessToken,
        User:        toUserResponse(user),
    }, nil
}
```

**Key patterns:**
- Services depend on interfaces (injected via constructor)
- Services validate business rules
- Services orchestrate operations across repositories

---

### 6. DTOs (`internal/app/auth/dto.go`)

DTOs (Data Transfer Objects) define API request/response shapes:

```go
// What the client SENDS
type RegisterRequest struct {
    Email     string `json:"email" validate:"required,email"`
    Password  string `json:"password" validate:"required,min=8"`
    FirstName string `json:"first_name" validate:"required"`
    LastName  string `json:"last_name" validate:"required"`
}

// What we RETURN
type AuthResponse struct {
    AccessToken  string       `json:"access_token"`
    RefreshToken string       `json:"refresh_token"`
    ExpiresIn    int64        `json:"expires_in"`
    User         UserResponse `json:"user"`
}
```

**Why DTOs?**
- Separate API contract from internal entities
- Add validation rules
- Control what data is exposed

---

### 7. HTTP Handlers (`internal/handler/`)

Handlers translate HTTP requests to service calls:

```go
// internal/handler/auth_handler.go
func (h *AuthHandler) Register(c *gin.Context) {
    // 1. Parse JSON body
    var req auth.RegisterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.BadRequest(c, "Invalid request body")
        return
    }
    
    // 2. Validate
    if errors := h.validator.Validate(req); errors != nil {
        response.ValidationError(c, errors)
        return
    }
    
    // 3. Call service
    result, err := h.authService.Register(c.Request.Context(), req)
    if err != nil {
        response.Error(c, err)
        return
    }
    
    // 4. Return response
    response.Created(c, result)
}
```

**Handlers are THIN:** They only:
1. Parse request
2. Validate input
3. Call service
4. Format response

---

### 8. Middleware (`internal/middleware/`)

Middleware runs **before** or **after** handlers:

```go
// Authentication middleware
func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Get token from header
        token := c.GetHeader("Authorization")
        token = strings.TrimPrefix(token, "Bearer ")
        
        // Validate JWT
        claims, err := m.jwtManager.ValidateAccessToken(token)
        if err != nil {
            response.Unauthorized(c, "Invalid token")
            c.Abort()  // Stop the chain!
            return
        }
        
        // Store user info in context
        c.Set("user_id", claims.UserID)
        c.Set("user_role", claims.Role)
        
        c.Next()  // Continue to handler
    }
}
```

**Common middleware:**
- **Auth:** Verify JWT tokens
- **CORS:** Allow cross-origin requests
- **Logging:** Log all requests
- **Recovery:** Catch panics

---

### 9. Router (`internal/router/`)

The router maps URLs to handlers:

```go
func (r *Router) Setup() *gin.Engine {
    router := gin.New()
    
    // Global middleware (runs on ALL requests)
    router.Use(middleware.RecoveryMiddleware(r.logger))
    router.Use(middleware.LoggingMiddleware(r.logger))
    
    // API v1 routes
    v1 := router.Group("/api/v1")
    {
        // Public routes (no auth required)
        auth := v1.Group("/auth")
        {
            auth.POST("/register", r.authHandler.Register)
            auth.POST("/login", r.authHandler.Login)
        }
        
        // Protected routes (auth required)
        movies := v1.Group("/movies")
        {
            movies.GET("", r.movieHandler.List)  // Public
            
            // Admin only
            movies.POST("", 
                r.authMiddleware.Authenticate(),
                r.authMiddleware.RequireAdmin(),
                r.movieHandler.Create,
            )
        }
    }
    
    return router
}
```

---

## 🔄 Request Lifecycle

Here's how a request flows through the system:

```
Client: POST /api/v1/auth/register
        {"email": "john@example.com", "password": "secret123"}
            │
            ▼
┌─────────────────────────────────────────────────────────────┐
│  1. MIDDLEWARE CHAIN                                         │
│     RecoveryMiddleware → LoggingMiddleware → CORSMiddleware │
└─────────────────────────────────────────────────────────────┘
            │
            ▼
┌─────────────────────────────────────────────────────────────┐
│  2. ROUTER                                                   │
│     Matches POST /api/v1/auth/register → authHandler.Register│
└─────────────────────────────────────────────────────────────┘
            │
            ▼
┌─────────────────────────────────────────────────────────────┐
│  3. HANDLER (internal/handler/auth_handler.go)              │
│     - Parses JSON body into RegisterRequest                  │
│     - Validates fields                                       │
│     - Calls authService.Register(req)                        │
└─────────────────────────────────────────────────────────────┘
            │
            ▼
┌─────────────────────────────────────────────────────────────┐
│  4. SERVICE (internal/app/auth/service.go)                  │
│     - Checks if email exists (userRepo.EmailExists)          │
│     - Hashes password (passwordMgr.HashPassword)             │
│     - Creates user (userRepo.Create)                         │
│     - Generates JWT (jwtManager.GenerateAccessToken)         │
│     - Returns AuthResponse                                   │
└─────────────────────────────────────────────────────────────┘
            │
            ▼
┌─────────────────────────────────────────────────────────────┐
│  5. REPOSITORY (internal/app/postgres/user_repository.go)   │
│     - Executes: INSERT INTO users (...)                      │
│     - Returns created user                                   │
└─────────────────────────────────────────────────────────────┘
            │
            ▼
CLIENT RECEIVES:
{
    "success": true,
    "data": {
        "access_token": "eyJhbGciOiJIUzI1NiIs...",
        "user": { "id": "...", "email": "john@example.com" }
    }
}
```

---

## 🧠 Key Concepts Explained

### Dependency Injection

Instead of creating dependencies inside a struct, we **inject** them:

```go
// ❌ BAD - Hard to test, tightly coupled
type UserService struct {}

func (s *UserService) GetUser(id string) *User {
    db := database.Connect()  // Creates its own dependency!
    return db.FindUser(id)
}

// ✅ GOOD - Testable, loosely coupled
type UserService struct {
    repo repository.UserRepository  // Injected!
}

func NewUserService(repo repository.UserRepository) *UserService {
    return &UserService{repo: repo}
}
```

**In `main.go`, we wire everything:**
```go
db := postgres.New(cfg.Database)
userRepo := postgres.NewUserRepository(db)
userService := auth.NewService(userRepo, ...)
userHandler := handler.NewAuthHandler(userService, ...)
```

---

### Context (`context.Context`)

Context carries request-scoped values and cancellation signals:

```go
func (s *Service) GetUser(ctx context.Context, id uuid.UUID) (*User, error) {
    // If the HTTP request is cancelled, ctx.Done() fires
    // GORM uses ctx for timeouts
    return s.repo.GetByID(ctx, id)
}
```

**Always pass context as the first parameter.**

---

### Error Handling

Go doesn't have exceptions. Functions return errors:

```go
user, err := s.userRepo.GetByEmail(ctx, email)
if err != nil {
    // Handle error
    return nil, err
}
// Use user safely
```

This project uses custom error types for better API responses:

```go
// internal/pkg/errors/errors.go
func ErrUserNotFound() error {
    return &AppError{
        Code:    CodeNotFound,
        Message: "User not found",
    }
}
```

---

## 🚀 How to Run

### Prerequisites
- Go 1.21+
- Docker & Docker Compose

### Steps

```bash
# 1. Start database & Redis
cd backend
docker-compose -f deployments/docker-compose.yml up -d

# 2. Copy and configure environment
cp .env.example .env

# 3. Run the application
go run ./cmd/api

# 4. Test the API
curl http://localhost:8080/health
```

---

## 🏋️ Exercises

### Exercise 1: Add a Field
Add `phone_number` to the User entity and registration API.

**Files to modify:**
1. `internal/app/entity/user.go` - Add field
2. `internal/app/auth/dto.go` - Add to RegisterRequest
3. `internal/app/auth/service.go` - Map the field

### Exercise 2: Add an Endpoint
Create `GET /api/v1/movies/:id/showtimes`.

**Files to modify:**
1. `internal/app/repository/` - Add interface method
2. `internal/app/postgres/` - Implement it
3. `internal/app/showtime/service.go` - Add service method
4. `internal/handler/` - Add handler method
5. `internal/router/router.go` - Add route

### Exercise 3: Add Middleware
Create a middleware that logs the response time of each request.

**Hint:** Use `time.Since(start)` after `c.Next()`.

---

## 📚 Further Reading

- [Gin Web Framework](https://gin-gonic.com/)
- [GORM Documentation](https://gorm.io/)
- [golang-standards/project-layout](https://github.com/golang-standards/project-layout)
- [Effective Go](https://golang.org/doc/effective_go)

---

**Happy coding! 🎬**
