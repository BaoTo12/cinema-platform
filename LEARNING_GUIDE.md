# 🎓 Learning Guide: CinemaOS Golang Backend

## 📚 How to Learn From This Codebase

Welcome! This guide will help you understand and learn from the CinemaOS backend architecture. Whether you're new to Go or want to understand modern backend patterns, follow this structured approach.

---

## 🎯 Learning Objectives

By studying this codebase, you'll learn:
- ✅ Golang project structure and organization
- ✅ Database modeling with GORM
- ✅ Redis for caching and distributed locking
- ✅ JWT authentication implementation
- ✅ RESTful API design patterns
- ✅ Concurrent programming (seat locking)
- ✅ Docker containerization
- ✅ Clean architecture principles

---

## 🎓 Prerequisites - What You Need to Know

Before diving into this codebase, you should understand:
- ✅ Basic Go syntax (variables, functions, structs)
- ✅ Go packages and imports
- ❓ **Web development in Go** ← We'll teach you this!
- ❓ **Databases in Go** ← We'll teach you this!
- ❓ **HTTP servers** ← We'll teach you this!

If you only know Go basics, **start with Level 0 below**!

---

## 📖 Learning Path (Start Here!)

### Level 0: Go Web Development Fundamentals (1-2 hours)
**Goal**: Understand how web applications work in Go

> ⚡ **START HERE if you've never built a web app in Go!**

#### 🌐 Concept 1: How HTTP Works in Go

**The Basics:**
```go
// Every web application needs:
// 1. A server (listens for requests)
// 2. Handlers (respond to requests)
// 3. A router (maps URLs to handlers)

package main

import (
    "fmt"
    "net/http"
)

func main() {
    // This is a handler - a function that responds to HTTP requests
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "Hello, World!")
    })
    
    // Start the server on port 8080
    http.ListenAndServe(":8080", nil)
}
```

**What's happening:**
1. `http.HandleFunc("/", handler)` → "When someone visits `/`, run this function"
2. `w http.ResponseWriter` → What you send back to the browser
3. `r *http.Request` → What the browser sent to you
4. `http.ListenAndServe(":8080", nil)` → Start listening on port 8080

**Try it yourself:**
1. Save this as `simple-server.go`
2. Run: `go run simple-server.go`
3. Open browser to: http://localhost:8080
4. You'll see "Hello, World!"

---

#### 📦 Concept 2: Go Packages & Project Structure

**In Go, code is organized in packages:**

```go
// File: main.go
package main  // ← Package name

import (
    "fmt"                    // ← Standard library
    "net/http"               // ← Standard library
    "github.com/user/repo"   // ← External package
    "myproject/internal/db"  // ← Your own package
)
```

**Our project structure:**
```
backend/
├── cmd/server/main.go       ← package main (entry point)
├── internal/
│   ├── models/user.go       ← package models
│   ├── database/db.go       ← package database
│   └── services/auth.go     ← package services
└── go.mod                   ← Dependency management
```

**Key Rules:**
- `package main` → Can be run (`go run`)
- `package xyz` → Cannot be run directly, must be imported
- `internal/` → Private packages (can't be imported from outside)

---

#### 🗄️ Concept 3: What is a Database in Go?

**Without a library (raw SQL):**
```go
// Hard way - writing SQL manually
db.Query("SELECT * FROM users WHERE email = ?", email)
```

**With GORM (ORM - Object Relational Mapping):**
```go
// Easy way - using Go structs
type User struct {
    ID    uint
    Email string
    Name  string
}

// Query becomes:
db.Where("email = ?", email).First(&user)
```

**Why use GORM?**
- ✅ Write Go code, not SQL
- ✅ Type-safe (compiler catches errors)
- ✅ Auto-creates tables
- ✅ Handles relationships automatically

**What you'll see in this project:**
```go
// This Go struct:
type User struct {
    ID    uuid.UUID `gorm:"type:uuid;primaryKey"`
    Email string    `gorm:"uniqueIndex"`
}

// Becomes this SQL table:
CREATE TABLE users (
    id UUID PRIMARY KEY,
    email VARCHAR UNIQUE
);
```

---

#### 🔗 Concept 4: HTTP Methods (GET, POST, PUT, DELETE)

**In web apps, different actions use different methods:**

```go
func handler(w http.ResponseWriter, r *http.Request) {
    // Check what method was used
    switch r.Method {
    case "GET":
        // Read data (e.g., show a list of movies)
        fmt.Fprintf(w, "Getting data...")
        
    case "POST":
        // Create new data (e.g., register a user)
        fmt.Fprintf(w, "Creating data...")
        
    case "PUT":
        // Update existing data (e.g., update profile)
        fmt.Fprintf(w, "Updating data...")
        
    case "DELETE":
        // Delete data (e.g., cancel booking)
        fmt.Fprintf(w, "Deleting data...")
    }
}
```

**Common pattern:**
- `GET /movies` → List all movies
- `GET /movies/123` → Get movie with ID 123
- `POST /movies` → Create a new movie
- `PUT /movies/123` → Update movie 123
- `DELETE /movies/123` → Delete movie 123

---

#### 📝 Concept 5: JSON in Go

**Web APIs use JSON to send data:**

```go
// Define a struct
type Movie struct {
    ID    string `json:"id"`      // ← JSON tag
    Title string `json:"title"`
    Year  int    `json:"year"`
}

// Convert struct to JSON (encoding)
movie := Movie{ID: "1", Title: "Inception", Year: 2010}
jsonData, _ := json.Marshal(movie)
// Result: {"id":"1","title":"Inception","year":2010}

// Convert JSON to struct (decoding)
var movie Movie
json.Unmarshal(jsonData, &movie)
```

**In a web handler:**
```go
func getMovie(w http.ResponseWriter, r *http.Request) {
    movie := Movie{ID: "1", Title: "Inception", Year: 2010}
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(movie)
    // Browser receives: {"id":"1","title":"Inception","year":2010}
}
```

---

#### 🔐 Concept 6: Environment Variables

**Never hardcode secrets in code!**

```go
// ❌ BAD - Don't do this!
password := "my-secret-password"

// ✅ GOOD - Use environment variables
password := os.Getenv("DB_PASSWORD")
```

**In this project, we use `.env` file:**
```bash
# .env file
DB_PASSWORD=super-secret
JWT_SECRET=another-secret
```

**Then load it in Go:**
```go
import "github.com/joho/godotenv"

godotenv.Load()  // Reads .env file
password := os.Getenv("DB_PASSWORD")
```

---

#### 🧩 Concept 7: Struct Tags (The `backtick` things)

**You'll see this a lot:**
```go
type User struct {
    ID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
    Email string    `gorm:"uniqueIndex" json:"email"`
}
```

**What are those backticks?**
- They're called "struct tags"
- They're metadata that other packages read
- Different packages use different tags

**Common tags:**
- `json:"id"` → When converting to JSON, call this field "id"
- `gorm:"type:uuid"` → GORM: use UUID type in database
- `gorm:"uniqueIndex"` → GORM: make this unique in database

**Example:**
```go
type User struct {
    ID    int    `json:"user_id" gorm:"primaryKey"`
    Email string `json:"email" gorm:"uniqueIndex"`
}

// In JSON:    {"user_id": 1, "email": "..."}
// In Database: id (PRIMARY KEY), email (UNIQUE)
```

---

#### 🎯 Concept 8: Pointers in Web Development

**You'll see `*` and `&` a lot:**

```go
// Without pointer
func UpdateUser(user User) {
    user.Email = "new@email.com"  // Changes local copy only
}

// With pointer
func UpdateUser(user *User) {
    user.Email = "new@email.com"  // Changes the actual user
}

// Usage:
user := User{Email: "old@email.com"}
UpdateUser(&user)  // ← Pass address with &
fmt.Println(user.Email)  // "new@email.com"
```

**Common patterns:**
- `*User` → "Pointer to a User"
- `&user` → "Give me the address of user"
- Database queries often use pointers:
  ```go
  var user User
  db.First(&user)  // ← Need pointer to fill the struct
  ```

---

#### ⚡ Concept 9: Error Handling in Go

**Go doesn't have try/catch. Instead:**

```go
// Most functions return (result, error)
result, err := doSomething()

if err != nil {
    // Handle the error
    log.Fatal(err)
    return
}

// Use result (only if err was nil)
fmt.Println(result)
```

**Common pattern in this codebase:**
```go
func GetUser(id string) (*User, error) {
    var user User
    
    err := db.First(&user, id).Error
    if err != nil {
        return nil, err  // Return nil user and the error
    }
    
    return &user, nil  // Return user and nil error
}

// Usage:
user, err := GetUser("123")
if err != nil {
    // Handle error
}
```

---

#### 🔄 Concept 10: Middleware Pattern

**Middleware = Code that runs BEFORE your handler**

```go
// Without middleware
func handler(w http.ResponseWriter, r *http.Request) {
    // Check authentication
    token := r.Header.Get("Authorization")
    if token == "" {
        http.Error(w, "Unauthorized", 401)
        return
    }
    
    // Actual logic
    fmt.Fprintf(w, "Hello!")
}

// With middleware (cleaner!)
func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Check authentication
        token := r.Header.Get("Authorization")
        if token == "" {
            http.Error(w, "Unauthorized", 401)
            return
        }
        
        // Call next handler
        next(w, r)
    }
}

// Usage:
http.HandleFunc("/protected", authMiddleware(handler))
```

**In this project:**
- `middleware/auth.go` → Checks JWT tokens
- Applied to routes that need authentication

---

#### 📚 Concept 11: Go Modules (go.mod)

**`go.mod` is like `package.json` for Node.js:**

```go
module myproject

go 1.21

require (
    github.com/lib/pq v1.10.9        // PostgreSQL driver
    gorm.io/gorm v1.25.5             // ORM
)
```

**Commands:**
- `go mod init myproject` → Create new module
- `go mod tidy` → Add missing, remove unused dependencies
- `go get package@version` → Add specific package

---

#### 🎓 Quick Exercise: Build a Simple API

**Try this before reading the codebase:**

```go
package main

import (
    "encoding/json"
    "net/http"
)

type Movie struct {
    Title string `json:"title"`
    Year  int    `json:"year"`
}

func main() {
    // GET /movies - return list
    http.HandleFunc("/movies", func(w http.ResponseWriter, r *http.Request) {
        movies := []Movie{
            {Title: "Inception", Year: 2010},
            {Title: "Interstellar", Year: 2014},
        }
        
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(movies)
    })
    
    http.ListenAndServe(":8080", nil)
}
```

**Test it:**
1. Save as `movies.go`
2. Run: `go run movies.go`
3. Open: http://localhost:8080/movies
4. You should see JSON!

---

### ✅ Level 0 Checklist

Before moving to Level 1, make sure you understand:

- [ ] How to create a basic HTTP server
- [ ] What handlers do (w ResponseWriter, r Request)
- [ ] How packages work in Go
- [ ] What GORM does (ORM = struct ↔ database)
- [ ] HTTP methods (GET, POST, PUT, DELETE)
- [ ] JSON encoding/decoding
- [ ] Environment variables
- [ ] Struct tags (those backtick things)
- [ ] Pointers (`*` and `&`)
- [ ] Error handling (if err != nil)
- [ ] Middleware concept

**Once you understand these, you're ready for Level 1!**

---

### Level 1: Project Structure (30 mins)
**Goal**: Understand how the code is organized

#### Step 1: Explore the Directory Tree
```
backend/
├── cmd/server/          ← START HERE: Main application entry point
├── internal/            ← Core business logic (most important)
│   ├── models/         ← Database models (read second)
│   ├── database/       ← Database connection setup
│   ├── cache/          ← Redis client and seat locking
│   ├── services/       ← Business logic (complex part)
│   ├── middleware/     ← Auth and request processing
│   └── utils/          ← Helper functions
├── proto/              ← API definitions (Protocol Buffers)
└── go.mod              ← Dependencies
```

#### Start Reading Here (in order):
1. **`cmd/server/main.go`** (50 lines)
   - Application entry point
   - Shows how everything connects
   - Simple and easy to understand

2. **`internal/database/database.go`** (80 lines)
   - Database connection setup
   - Auto-migration logic
   - Connection pooling

3. **`internal/models/user.go`** (50 lines)
   - Simple model example
   - GORM basics
   - Relationships

---

### Level 2: Core Concepts (1-2 hours)

#### A. Database Models with GORM

**Read these files in order:**

1. **`internal/models/user.go`**
```go
// What to learn:
// - GORM struct tags (gorm:"type:uuid")
// - Relationships (HasMany, BelongsTo)
// - Table naming conventions
// - Indexes

type User struct {
    ID           uuid.UUID  `gorm:"type:uuid;primary_key"`
    Email        string     `gorm:"uniqueIndex;not null"`
    PasswordHash string     `gorm:"not null"`
    // ... more fields
    
    Bookings []Booking `gorm:"foreignKey:UserID"` // One-to-Many
}
```

**Key Concepts to Understand:**
- `gorm:"type:uuid"` → PostgreSQL UUID type
- `gorm:"uniqueIndex"` → Creates database index for fast lookups
- `gorm:"foreignKey:UserID"` → Defines relationship
- `@map("snake_case")` → Maps Go field to database column

2. **`internal/models/cinema.go`**
```go
// More complex relationships:
// - Multiple levels (Cinema → Screen → Seat)
// - JSONB fields for flexibility
// - Enums for type safety

type SeatType string
const (
    SeatStandard   SeatType = "STANDARD"
    SeatPremium    SeatType = "PREMIUM"
    SeatVIP        SeatType = "VIP"
)
```

**Exercise**: Try to understand:
- Why use enums instead of strings?
- How are Cinema, Screen, and Seat related?
- What's stored in JSON vs separate tables?

3. **`internal/models/booking.go`**
```go
// Advanced concepts:
// - Optimistic locking (Version field)
// - Soft deletes (DeletedAt)
// - Decimal types for money
// - Status state machines

type Booking struct {
    Version int `gorm:"default:0"` // Optimistic locking!
    FinalAmount float64 `gorm:"type:decimal(10,2)"`
}
```

---

#### B. Database Connection & Migrations

**Read: `internal/database/database.go`**

```go
// Key learning points:

// 1. Singleton Pattern
var DB *gorm.DB

// 2. Connection Pooling
sqlDB.SetMaxIdleConns(10)
sqlDB.SetMaxOpenConns(100)
sqlDB.SetConnMaxLifetime(time.Hour)

// 3. Auto-Migration
DB.AutoMigrate(&models.User{}, &models.Movie{}, ...)
```

**What to Learn:**
- Why use a singleton for database connection?
- What is connection pooling and why is it important?
- How does auto-migration work?
- When to use migrations vs manual SQL?

**Try This:**
1. Look at the connection string format
2. Understand each pooling parameter
3. See what happens during AutoMigrate()

---

#### C. Redis for Distributed Locking

**Read: `internal/cache/redis.go`**

This is **ADVANCED** but super important for concurrent booking!

```go
// The Problem: Two users trying to book the same seat simultaneously
// The Solution: Redis atomic locks

func LockSeat(showtimeID, seatID, sessionToken string, expiry time.Duration) (bool, error) {
    key := fmt.Sprintf("lock:showtime:%s:seat:%s", showtimeID, seatID)
    
    // SetNX = "Set if Not eXists" - ATOMIC operation!
    result, err := Client.SetNX(ctx, key, sessionToken, expiry).Result()
    return result, err
}
```

**Key Concepts:**
- **Atomic Operations**: Either succeeds completely or fails (no in-between)
- **Distributed Locking**: Lock works across multiple servers
- **TTL (Time To Live)**: Lock expires automatically (5 minutes)
- **Lua Scripts**: For atomicity across multiple operations

**Mental Model:**
```
User A tries to lock Seat 5 → SUCCESS (gets the lock)
User B tries to lock Seat 5 → FAILS (lock exists)
After 5 minutes → Lock expires automatically
```

**Exercise**:
- Read `LockMultipleSeats()` - how does it ensure all-or-nothing?
- Why use Lua script in `UnlockSeat()`?
- What happens if the user's browser crashes?

---

### Level 3: Business Logic (2-3 hours)

#### A. Authentication System

**Read: `internal/utils/jwt.go`**

```go
// JWT = JSON Web Token
// Three parts: Header.Payload.Signature

type Claims struct {
    UserID string `json:"user_id"`
    Email  string `json:"email"`
    Role   string `json:"role"`
    jwt.RegisteredClaims
}

func GenerateAccessToken(userID uuid.UUID, email, role string) (string, error) {
    claims := Claims{
        UserID: userID.String(),
        Email:  email,
        Role:   role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
        },
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(jwtSecret)
}
```

**Understand:**
- Why two tokens (access + refresh)?
- Why short expiry for access token (15 min)?
- How is token verified?
- What's the role of the secret key?

**Security Concepts:**
- Access token: Short-lived, used for API calls
- Refresh token: Long-lived, stored in database
- HMAC signing: Ensures token cannot be tampered with

---

#### B. Service Layer Architecture

**Read: `internal/services/auth_service.go`**

```go
// Service Pattern: Business logic separated from HTTP handlers

type AuthService struct{}

func (s *AuthService) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
    // 1. Validate input
    if req.Email == "" {
        return nil, errors.New("email required")
    }
    
    // 2. Check if user exists
    var existing User
    database.DB.Where("email = ?", req.Email).First(&existing)
    
    // 3. Hash password
    passwordHash, _ := utils.HashPassword(req.Password)
    
    // 4. Create user
    user := User{...}
    database.DB.Create(&user)
    
    // 5. Return response
    return &RegisterResponse{...}, nil
}
```

**Learn:**
- Input validation first
- Database queries with GORM
- Error handling patterns
- Transaction boundaries

**Exercise**: Read `Login()` method
- How is password verified?
- Why update `lastLoginAt`?
- Where are refresh tokens stored?

---

#### C. Pricing Engine (Dynamic Pricing!)

**Read: `internal/services/pricing_service.go`**

```go
// Super cool algorithm! Prices change based on multiple factors

func CalculatePrice(showtimeID, seatIDs) {
    price := basePrice  // Start with $10
    
    // 1. Seat Type Modifier
    if seat.Type == "PREMIUM" {
        price += 3.0
    }
    
    // 2. Time-Based Pricing
    if isPeakTime(startTime) {  // 6-9 PM
        price += 2.0
    }
    
    // 3. Day-of-Week Pricing
    if isWeekend(showDate) {
        price += 2.0
    }
    
    // 4. Demand-Based Pricing (Supply & Demand!)
    occupancyRate := (totalSeats - availableSeats) / totalSeats
    if occupancyRate > 0.90 {  // > 90% full
        price += 4.0
    }
    
    return price
}
```

**Concepts:**
- Dynamic pricing (like Uber surge pricing!)
- Multiple pricing factors combined
- Real-time demand calculation

**Exercise:**
- What if it's a premium seat, on Friday night, at 8 PM, and 95% full?
- Calculate the final price!
- How would you add holiday pricing?

---

### Level 4: Advanced Patterns (3-4 hours)

#### A. Concurrent Booking Flow

**Read: `internal/services/bookings_service.go`**

```go
// The Challenge: Prevent double-booking with high concurrency

func HoldSeats(showtimeID, seatIDs, sessionToken) {
    // Step 1: Try to lock in Redis (atomic!)
    locked := cache.LockMultipleSeats(showtimeID, seatIDs, sessionToken, 5*time.Minute)
    
    if !locked {
        return "Seats already taken"
    }
    
    // Step 2: Calculate pricing
    pricing := pricingService.CalculatePrice(...)
    
    // Step 3: Return hold confirmation
    return HoldResponse{
        HoldID: sessionToken,
        ExpiresAt: time.Now().Add(5 * time.Minute),
        Pricing: pricing,
    }
}

func ConfirmBooking(holdID) {
    // Use database transaction
    database.DB.Transaction(func(tx *gorm.DB) error {
        // 1. Create booking record
        booking := Booking{...}
        tx.Create(&booking)
        
        // 2. Decrement available seats (with optimistic locking!)
        result := tx.Model(&Showtime{}).
            Where("id = ? AND version = ?", id, version).
            Updates(map[string]interface{}{
                "available_seats": gorm.Expr("available_seats - ?", numSeats),
                "version": version + 1,  // Increment version!
            })
        
        if result.RowsAffected == 0 {
            return errors.New("Concurrent modification detected")
        }
        
        return nil
    })
}
```

**Advanced Concepts:**
- **Optimistic Locking**: Version field prevents race conditions
- **Database Transactions**: All-or-nothing operations
- **Two-Phase Locking**: Redis lock → Database confirm
- **Automatic Expiry**: Locks auto-release after 5 minutes

**Mental Model:**
```
Phase 1 (Hold):
- User clicks seats
- Redis locks created (5 min TTL)
- Pricing calculated
- User has 5 min to pay

Phase 2 (Confirm):
- User completes payment
- Database transaction starts
- Booking created
- Seats decremented (with version check)
- Redis locks released
- Transaction commits
```

---

## 🛠️ Hands-On Exercises

### Exercise 1: Trace a Request
Pick the registration flow and trace it through the code:
1. Start at `auth_service.go → Register()`
2. Follow the password hashing in `utils/helpers.go`
3. See the database insert with GORM
4. Understand the response format

### Exercise 2: Modify Pricing
Try adding a new pricing rule:
- Students get 20% discount
- Add a `StudentDiscount` field to User model
- Modify `pricing_service.go` to check it
- Calculate the new price

### Exercise 3: Add Logging
Pick any service and add logging:
```go
import "log"

log.Printf("User %s attempting login", email)
```

---

## 📊 Architecture Patterns Used

### 1. **Layered Architecture**
```
main.go → Services → Models → Database
         ↓
      Middleware
```

### 2. **Repository Pattern** (implicit with GORM)
```go
database.DB.Where("email = ?", email).First(&user)
// Instead of raw SQL
```

### 3. **Service Pattern**
Business logic in services, not controllers

### 4. **Singleton Pattern**
One database connection, one Redis client

---

## 🎯 Key Takeaways

**What Makes This Code Good:**
1. ✅ **Clear separation of concerns** (models, services, utils)
2. ✅ **Type safety** (Enums, strong typing)
3. ✅ **Error handling** (Always check errors)
4. ✅ **Concurrency safety** (Redis locks, optimistic locking)
5. ✅ **Security** (JWT, password hashing, CORS)

**Common Go Patterns You'll See:**
- Pointer receivers: `func (s *Service) Method()`
- Error returns: `func DoSomething() (result, error)`
- Struct embedding: `jwt.RegisteredClaims`
- Interface satisfaction: Implicit interfaces

---

## 📚 Further Learning

### Want to Understand More?

**Go Basics:**
- [Tour of Go](https://tour.golang.org/)
- [Effective Go](https://golang.org/doc/effective_go.html)

**GORM:**
- [GORM Guides](https://gorm.io/docs/)
- Focus on: Associations, Hooks, Transactions

**Redis:**
- [Redis University](https://university.redis.com/)
- Focus on: Atomic operations, Pub/Sub, TTL

**JWT:**
- [jwt.io](https://jwt.io/)
- Understanding: Claims, Signing, Verification

---

## 🎓 Learning Checklist

Track your progress:

- [ ] Understand project structure
- [ ] Read all model files
- [ ] Understand database connection
- [ ] Learn GORM basics
- [ ] Understand Redis locking
- [ ] Study JWT authentication
- [ ] Read auth service
- [ ] Read movies service
- [ ] Understand pricing algorithm
- [ ] Study booking flow
- [ ] Trace a complete request
- [ ] Try modifying code
- [ ] Add a new feature
- [ ] Write a test

---

## 💡 Tips for Learning

1. **Start Small**: Don't try to understand everything at once
2. **Use Debugger**: Add `log.Printf()` statements
3. **Draw Diagrams**: Sketch the data flow
4. **Ask Questions**: Why is it designed this way?
5. **Modify Code**: Change something small and see what breaks
6. **Read Tests**: (When added) Tests show how to use code

---

**Happy Learning! 🚀**

Questions? Check the inline comments in each file for more context.
