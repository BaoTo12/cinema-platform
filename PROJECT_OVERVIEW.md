# 🎬 CinemaOS - Complete Project Overview

## What This Application Does

**CinemaOS** is a full-stack cinema management platform - like building your own **Fandango**, **BookMyShow**, or **AMC Theatres** online booking system.

### Core Features:

1. **🎥 Movie Management**
   - Display movies with posters, trailers, ratings
   - Filter by genre, format (2D, 3D, IMAX)
   - Search functionality
   - Admin can add/edit/remove movies

2. **📅 Intelligent Showtime Scheduling**
   - Automatically generate schedules for multiple screens
   - Avoid conflicts (same movie on multiple screens)
   - Optimize screen utilization
   - Consider movie duration, cleaning time

3. **💺 Real-Time Seat Selection**
   - Interactive seat map (like airline booking)
   - Live seat availability
   - Different seat types (Standard, Premium, VIP)
   - Visual seat status (Available, Selected, Booked)

4. **💰 Dynamic Pricing Engine**
   - Prices change based on:
     - **Time of day** (peak hours cost more)
     - **Day of week** (weekends more expensive)
     - **Demand** (90% full = surge pricing)
     - **Seat type** (VIP costs more)
   - Promo codes and discounts

5. **🔒 Concurrent Booking Prevention**
   - Multiple users can't book same seat
   - Redis distributed locks
   - 5-minute seat holds
   - Optimistic locking for database

6. **👤 User Authentication**
   - Secure registration & login
   - JWT-based authentication
   - Role-based access (Customer, Manager, Admin)
   - Refresh token mechanism

7. **💳 Payment Processing**
   - Stripe integration ready
   - Secure payment handling
   - Booking confirmation emails

8. **📊 Admin Dashboard**
   - Manage movies and showtimes
   - View booking statistics
   - Revenue analytics
   - User management

---

## 🛠️ Complete Technology Stack

### Backend Technologies

#### **1. Go (Golang)** 🐹

**What it is:**  
A statically-typed, compiled programming language created by Google in 2009.

**Why it's special:**
- **Fast**: Compiles to machine code (not interpreted)
- **Concurrent**: Built-in goroutines for parallel processing
- **Simple**: Clean syntax, easy to learn
- **Reliable**: Strong typing catches bugs early

**Real-world users:** Google, Uber, Dropbox, Docker, Kubernetes

**In this project:**
- All backend server code
- Business logic and services
- Database operations
- API endpoints

**Example from our code:**
```go
func (s *AuthService) Login(email, password string) (*User, error) {
    // Business logic goes here
    user, err := database.FindUserByEmail(email)
    if err != nil {
        return nil, err
    }
    return user, nil
}
```

---

#### **2. GORM** 🗄️

**What it is:**  
Object-Relational Mapping library for Go - translates between Go structs and database tables.

**Official site:** https://gorm.io

**What problem does it solve:**
Without GORM, you write raw SQL:
```sql
INSERT INTO users (id, email, name) VALUES ($1, $2, $3)
SELECT * FROM users WHERE email = $1
```

With GORM, you write Go code:
```go
user := User{Email: "test@example.com", Name: "John"}
db.Create(&user)  // INSERT
db.Where("email = ?", email).First(&user)  // SELECT
```

**Key features:**
- **Auto-migration**: Creates/updates tables automatically
- **Associations**: Handles relationships (user has many bookings)
- **Hooks**: Run code before/after database operations
- **Transactions**: All-or-nothing operations
- **Soft deletes**: Mark as deleted without removing

**In this project:**
- 12 database models (User, Movie, Booking, etc.)
- Complex relationships (Cinema → Screens → Seats → Bookings)
- Auto-creates all tables on startup

**Example:**
```go
type User struct {
    ID       uuid.UUID `gorm:"type:uuid;primaryKey"`
    Email    string    `gorm:"uniqueIndex;not null"`
    Bookings []Booking `gorm:"foreignKey:UserID"`
}
```

---

#### **3. PostgreSQL** 🐘

**What it is:**  
Advanced open-source relational database management system.

**Official site:** https://postgresql.org

**Why PostgreSQL (not MySQL or MongoDB):**
- **More features**: JSON support, arrays, full-text search
- **Better data integrity**: Strict ACID compliance
- **Advanced types**: UUID, JSONB, geometric types
- **Reliable**: Used by Instagram, Spotify, Reddit

**Key features we use:**
- **UUID primary keys**: Globally unique identifiers
- **JSONB columns**: Store flexible JSON data efficiently
- **Indexes**: Fast lookups on email, booking reference
- **Foreign keys**: Enforce data relationships
- **Transactions**: Ensure booking consistency

**In this project:**
Stores everything:
- Users and authentication
- Movies and showtimes
- Bookings and payments
- Cinemas, screens, seats

**Database schema example:**
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY,
    email VARCHAR UNIQUE NOT NULL,
    password_hash VARCHAR NOT NULL,
    created_at TIMESTAMP
);

CREATE TABLE bookings (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    showtime_id UUID REFERENCES showtimes(id),
    total_amount DECIMAL(10,2)
);
```

---

#### **4. Redis** 🔴

**What it is:**  
In-memory data structure store - super fast key-value cache.

**Official site:** https://redis.io

**Why it's insanely fast:**
- Stores data in **RAM** (not disk)
- Simple data structures (strings, hashes, sets)
- Single-threaded (no lock overhead)

**Speed comparison:**
- Disk database: ~10ms per query
- Redis: ~0.1ms per query (100x faster!)

**Key features:**
- **Atomic operations**: Thread-safe by design
- **TTL (Time To Live)**: Data expires automatically
- **Pub/Sub**: Real-time messaging
- **Persistent**: Can save to disk

**Why we use it:**
**Problem**: Two users try to book the same seat simultaneously

**Without Redis:**
```
Time: 0ms - User A checks seat available ✓
Time: 1ms - User B checks seat available ✓
Time: 2ms - User A books seat
Time: 3ms - User B books seat
Result: DOUBLE BOOKING! ❌
```

**With Redis atomic locks:**
```
Time: 0ms - User A gets lock on seat ✓
Time: 1ms - User B tries lock, FAILS (already locked)
Time: 2ms - User A completes booking
Time: 5min - Lock expires (if user abandons)
Result: No double booking! ✓
```

**In this project:**
- Seat locking (5-minute holds)
- Session management
- Can cache frequently accessed data

**Code example:**
```go
// Try to lock seat atomically
locked := redis.SetNX("lock:seat:A5", userSessionID, 5*time.Minute)
if !locked {
    return "Seat already taken"
}
```

---

#### **5. JWT (JSON Web Tokens)** 🔐

**What it is:**  
A compact way to securely transmit information between parties.

**Official site:** https://jwt.io

**How traditional sessions work:**
```
User logs in → Server creates session in database
Every request → Server checks database
Problem: Database lookup on every request!
```

**How JWT works:**
```
User logs in → Server creates signed token
Every request → Server verifies signature (no database!)
Benefit: Stateless, fast, scalable
```

**Token structure:**
```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9  ← Header (algorithm)
.
eyJ1c2VyX2lkIjoiMTIzIiwiZW1haWwiOiJ0ZXN0QGV4YW1wbGUuY29tIn0  ← Payload (data)
.
SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c  ← Signature
```

**Security:**
- Cryptographically signed (can't be tampered)
- Contains expiration time
- Includes user info (no database lookup)

**In this project:**
- **Access token**: Short-lived (15 min), used for API calls
- **Refresh token**: Long-lived (7 days), to get new access tokens
- Contains: userID, email, role

**Code example:**
```go
// Create token
token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
    "user_id": userID,
    "email": email,
    "exp": time.Now().Add(15 * time.Minute).Unix(),
})
signedToken, _ := token.SignedString([]byte(secretKey))

// Verify token
claims, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
    return []byte(secretKey), nil
})
```

---

#### **6. bcrypt** 🔒

**What it is:**  
Password hashing function designed to be slow.

**Why "slow is good":**
- Attacker tries to crack password
- Must test millions of combinations
- Each attempt takes ~100ms
- Makes brute force impractical

**How it works:**
```go
password := "myPassword123"

// Hash (one-way, can't reverse)
hashed, _ := bcrypt.GenerateFromPassword([]byte(password), 10)
// Result: "$2a$10$N9qo8uLOcsvXXaLKr8hP0.NvKhP4fKUL1vQ9tJ3kBq..."

// Verify (compares hashes)
err := bcrypt.CompareHashAndPassword(hashed, []byte("myPassword123"))
// Returns: nil (match!) ✓

err := bcrypt.CompareHashAndPassword(hashed, []byte("wrongPassword"))
// Returns: error (no match!) ❌
```

**Key features:**
- **One-way**: Can't reverse engineer
- **Salted**: Same password → different hash each time
- **Adaptive**: Can increase difficulty over time

**In this project:**
- All passwords hashed before storing
- Never store plain text passwords
- Secure against rainbow table attacks

---

#### **7. Connect RPC** 🔌

**What it is:**  
Modern RPC (Remote Procedure Call) framework - like REST but better.

**Official site:** https://connectrpc.com

**REST vs Connect RPC:**

**REST (traditional):**
```typescript
// Frontend
fetch('/api/users/123', {method: 'GET'})
fetch('/api/users', {method: 'POST', body: JSON.stringify(userData)})

// Problems:
// - No type safety (typos in URLs)
// - Manual serialization
// - Manual error handling
```

**Connect RPC (modern):**
```typescript
// Frontend
const user = await userService.getUser({id: '123'})  // Type-safe!
await userService.createUser(userData)  // Auto-completes!

// Benefits:
// - Type safety (compiler catches errors)
// - Like calling functions
// - Auto-generated clients
```

**Why we use it:**
- **Type-safe** across frontend and backend
- **Protocol Buffers** define schema
- **HTTP/1.1 and HTTP/2** support
- **Simpler** than gRPC (no special setup)

**In this project:**
- 5 services defined: Auth, Movies, Showtimes, Bookings, Pricing
- `.proto` files define all APIs
- Auto-generates Go and TypeScript code

---

#### **8. Protocol Buffers** 📦

**What it is:**  
Google's language-agnostic data serialization format.

**Official site:** https://protobuf.dev

**JSON vs Protocol Buffers:**

**JSON (traditional):**
```json
{
  "id": "123",
  "email": "test@example.com",
  "age": 25
}
```
- Human-readable ✓
- Large size ❌
- No schema ❌
- Type errors at runtime ❌

**Protocol Buffers:**
```protobuf
message User {
  string id = 1;
  string email = 2;
  int32 age = 3;
}
```
- Smaller (binary format)
- Strict schema
- Type errors at compile time
- Auto-generates code

**In this project:**
All APIs defined in `.proto` files:
```protobuf
service AuthService {
  rpc Login(LoginRequest) returns (LoginResponse);
  rpc Register(RegisterRequest) returns (RegisterResponse);
}

message LoginRequest {
  string email = 1;
  string password = 2;
}
```

Generates type-safe code for both Go (backend) and TypeScript (frontend)!

---

#### **9. Docker** 🐳

**What it is:**  
Platform for developing, shipping, and running applications in containers.

**Official site:** https://docker.com

**Problem without Docker:**
```
Developer: "Works on my machine!" 💻
Production server: "Doesn't work here!" 🔥
- Different OS
- Different dependencies
- Different configurations
```

**Solution with Docker:**
```
Everything packaged together:
- Code
- Runtime (Go, Node.js)
- Dependencies
- System libraries
- Configuration

Runs EXACTLY the same everywhere! ✓
```

**Shipping container analogy:**
Just like shipping containers standardized global trade (works on any ship, truck, crane), Docker containers standardize software deployment (works on any computer).

**In this project:**
Each service in its own container:
- **postgres**: Database
- **redis**: Cache
- **backend**: Go API server
- **frontend**: Next.js UI

All built from `Dockerfile`:
```dockerfile
FROM golang:1.21 AS builder
WORKDIR /app
COPY . .
RUN go build -o server

FROM alpine:latest
COPY --from=builder /app/server /server
CMD ["/server"]
```

---

#### **10. Docker Compose** 📦

**What it is:**  
Tool for defining and running multi-container Docker applications.

**What it solves:**
```
Without Docker Compose:
docker run postgres ...
docker run redis ...
docker run backend ...
docker run frontend ...
(4 separate commands, manual networking, hard to manage)

With Docker Compose:
docker-compose up
(One command, everything connected!)
```

**Our `docker-compose.yml`:**
```yaml
services:
  postgres:
    image: postgres:15
    volumes:
      - postgres_data:/var/lib/postgresql/data
  
  redis:
    image: redis:7-alpine
  
  backend:
    build: ./backend
    depends_on:
      - postgres
      - redis
  
  frontend:
    build: ./frontend
    depends_on:
      - backend
```

**Benefits:**
- Start everything: `docker-compose up -d`
- View logs: `docker-compose logs -f`
- Stop everything: `docker-compose down`
- Automatic networking between services

---

### Frontend Technologies

#### **11. Next.js 14** ⚛️

**What it is:**  
React framework for production applications.

**Official site:** https://nextjs.org

**React vs Next.js:**

**React alone:**
- Client-side only
- Manual routing setup
- Manual SEO optimization
- Manual performance optimization

**Next.js (React + extras):**
- Server-side rendering (faster first load)
- File-based routing (`/movies/page.tsx` = `/movies` route)
- Image optimization
- API routes
- Built-in CSS support

**Latest features (App Router - Next.js 14):**
- Server Components (default)
- Streaming
- Better layouts system

**In this project:**
- All UI pages and components
- Responsive dark cinema theme
- Real-time seat selection interface

**File structure:**
```
app/
├── page.tsx              → / (home)
├── movies/
│   ├── page.tsx         → /movies
│   └── [id]/page.tsx    → /movies/123
├── login/page.tsx       → /login
└── layout.tsx           → Shared layout
```

---

#### **12. TailwindCSS** 🎨

**What it is:**  
Utility-first CSS framework.

**Official site:** https://tailwindcss.com

**Traditional CSS:**
```css
/* styles.css */
.button {
  background-color: #3b82f6;
  color:white;
  padding: 0.5rem 1rem;
  border-radius: 0.5rem;
}
```
```html
<button class="button">Click me</button>
```

**TailwindCSS:**
```html
<button class="bg-blue-500 text-white px-4 py-2 rounded-lg">
  Click me
</button>
```

**Benefits:**
- No CSS files needed
- Faster development
- Smaller final bundle (purges unused classes)
- Consistent design system
- Responsive by default

**In this project:**
- Custom color palette (primary red, dark theme, gold accents)
- Reusable component classes (btn-primary, card, input)
- Responsive grid layouts
- Smooth animations

---

#### **13. TypeScript** 📘

**What it is:**  
JavaScript with static type checking.

**Official site:** https://typescriptlang.org

**JavaScript (dynamic typing):**
```javascript
function add(a, b) {
  return a + b
}

add(5, 3)      // 8 ✓
add(5, "3")    // "53" ❌ (bug, but no error!)
```

**TypeScript (static typing):**
```typescript
function add(a: number, b: number): number {
  return a + b
}

add(5, 3)      // 8 ✓
add(5, "3")    // ERROR at compile time! ✓
                 // Argument of type 'string' is not assignable to 'number'
```

**Benefits:**
- Catches bugs before running code
- Better autocomplete in editor
- Self-documenting code
- Safer refactoring

**In this project:**
- All frontend code in TypeScript
- Type-safe API calls (thanks to Protocol Buffers)
- Interface definitions for all data structures

---

#### **14. React Query** 🔄

**What it is:**  
Powerful data fetching and caching library for React.

**Official site:** https://tanstack.com/query

**Problem without React Query:**
```typescript
// Lots of boilerplate
const [data, setData] = useState(null)
const [loading, setLoading] = useState(true)
const [error, setError] = useState(null)

useEffect(() => {
  fetch('/api/movies')
    .then(res => res.json())
    .then(setData)
    .catch(setError)
    .finally(() => setLoading(false))
}, [])

// Need to handle caching manually
// Need to handle refetching manually
// Need to handle stale data manually
```

**With React Query:**
```typescript
const { data, isLoading, error } = useQuery({
  queryKey: ['movies'],
  queryFn: () => fetch('/api/movies').then(r => r.json())
})

// Auto-handles:
// - Caching
// - Background refetching
// - Stale data
// - Loading states
```

**Features:**
- Automatic caching
- Background refetching
- Optimistic updates
- Pagination support
- Infinite scrolling

**In this project:**
- Paired with Connect RPC for type-safe data fetching
- Caches movie data, showtime information
- Real-time seat availability updates

---

### Development Tools

#### **15. Air** 🌬️

**What it is:**  
Live reload tool for Go applications (like nodemon for Node.js).

**What it does:**
```
You save main.go
  ↓
Air detects change
  ↓
Recompiles code
  ↓
Restarts server
  ↓
Instant feedback! (< 1 second)
```

**In this project:**
- `.air.toml` configuration
- Watches all `.go` files
- Excludes test files and vendor directory
- Makes development much faster

---

#### **16. Git** 📚

**What it is:**  
Version control system for tracking code changes.

**In this project:**
- 13+ commits with clear messages
- Tracks all changes
- Easy to revert if something breaks
- Collaboration ready

---

## 🔄 How Everything Works Together

### User Books a Ticket - Complete Flow:

**1. User lands on website**
```
Browser → Next.js (React)
  ↓
Fetch movies from API
  ↓
Display movie cards (TailwindCSS styling)
```

**2. User selects movie and showtime**
```
Click showtime button
  ↓
Navigate to /booking/[showtimeId]
  ↓
Frontend calls: showtimesService.getSeatMap()
  ↓
Connect RPC converts to HTTP request
  ↓
Go backend receives request
```

**3. Backend fetches seat map**
```
Go backend → GORM query
  ↓
SELECT * FROM seats WHERE screen_id = ?
  ↓
PostgreSQL returns seat data
  ↓
Check Redis for locked seats
  ↓
Combine data and return seat map
  ↓
Frontend displays interactive seat grid
```

**4. User selects seats**
```
User clicks seat A5, A6
  ↓
Frontend calls: bookingsService.holdSeats(['A5', 'A6'])
  ↓
Backend tries to lock in Redis:
  └─ redis.SetNX('lock:seat:A5', sessionID, 5min)
  └─ redis.SetNX('lock:seat:A6', sessionID, 5min)
  ↓
If successful, calculate price:
  └─ Base: $10/seat
  └─ Peak time: +$2
  └─ Weekend: +$2
  └─ Demand (85% full): +$2
  └─ Total: $16/seat × 2 = $32
  ↓
Return hold confirmation
  └─ Hold expires in 5 minutes
  └─ Total price: $32
```

**5. User completes payment**
```
Frontend: Stripe payment form
  ↓
User enters card details
  ↓
Stripe processes payment
  ↓
Frontend calls: bookingsService.confirmBooking(holdID, paymentID)
  ↓
Backend starts database transaction:
  ├─ Create booking record
  ├─ Create booking_seats records
  ├─ Decrement showtime.available_seats (with version check)
  ├─ Create payment record
  └─ If all succeed → COMMIT
      If any fails → ROLLBACK
  ↓
Release Redis locks
  ↓
Send confirmation email
  ↓
Return success + ticket details
```

**6. Display confirmation**
```
Frontend shows:
  ├─ Booking reference: #BK-ABCD1234
  ├─ Movie: The Dark Knight
  ├─ Time: 8:00 PM
  ├─ Seats: A5, A6
  ├─ Total: $32
  └─ QR code for entry
```

### Why Each Technology is Critical:

| Technology | Critical Role |
|------------|--------------|
| **Redis** | Prevents double-booking (distributed locks) |
| **PostgreSQL** | Stores all permanent data reliably |
| **GORM** | Makes database operations simple and safe |
| **JWT** | Stateless auth (scales to millions of users) |
| **Docker** | Same environment everywhere (dev = production) |
| **Next.js** | Fast, SEO-friendly frontend |
| **Protocol Buffers** | Type safety across frontend/backend |
| **Go** | Fast, handles concurrent bookings easily |

---

## 📚 Learn More

**Essential Tutorials:**
1. **Go Basics**: https://tour.golang.org (1-2 hours)
2. **GORM**: https://gorm.io/docs (30 min)
3. **Redis**: https://try.redis.io (30 min)
4. **Next.js**: https://nextjs.org/learn (2 hours)
5. **TailwindCSS**: https://tailwindcss.com/docs (30 min)

**Advanced Topics:**
- Connect RPC documentation
- Protocol Buffers guide
- Docker & Docker Compose tutorials
- PostgreSQL performance tuning

---

**Total Lines of Code**: ~8,000  
**Technologies Used**: 16  
**Services**: 4 (Containers)  
**Database Models**: 12  
**API Endpoints**: 25+

This is a production-ready, enterprise-grade cinema booking system! 🚀
