# CinemaOS Architecture

## System Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                        Client Layer                              │
├─────────────────────────────────────────────────────────────────┤
│  Next.js Frontend  │  Admin Dashboard  │  Mobile App (Future)   │
└────────────┬────────────────────────────────────────────────────┘
             │
             │ HTTPS/WSS
             ▼
┌─────────────────────────────────────────────────────────────────┐
│                     API Gateway / Load Balancer                  │
└────────────┬────────────────────────────────────────────────────┘
             │
             ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Backend Services                            │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │   Auth       │  │   Movie      │  │   Booking    │          │
│  │   Service    │  │   Service    │  │   Service    │          │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘          │
│         │                  │                  │                   │
│  ┌──────┴───────┐  ┌──────┴───────┐  ┌──────┴───────┐          │
│  │   Payment    │  │   Scheduler  │  │   Pricing    │          │
│  │   Service    │  │   Service    │  │   Service    │          │
│  └──────────────┘  └──────────────┘  └──────────────┘          │
│                                                                   │
└────────────┬────────────────────────────────────────────────────┘
             │
             ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Data Layer                                  │
├─────────────────────────────────────────────────────────────────┤
│  PostgreSQL (Primary)  │  Redis (Cache)  │  S3 (Assets)         │
└─────────────────────────────────────────────────────────────────┘
```

## Service Breakdown

### 1. Authentication Service
**Responsibilities:**
- User registration and login
- JWT token generation and validation
- Password reset and email verification
- Role-based access control (RBAC)
- Session management

**Endpoints:**
- `POST /api/auth/register`
- `POST /api/auth/login`
- `POST /api/auth/refresh`
- `POST /api/auth/logout`
- `POST /api/auth/forgot-password`

### 2. Movie Service
**Responsibilities:**
- CRUD operations for movies
- Movie metadata management
- Genre and rating management
- Movie poster and trailer links
- Search and filtering

**Endpoints:**
- `GET /api/movies` - List all movies
- `GET /api/movies/:id` - Get movie details
- `POST /api/movies` - Create movie (admin)
- `PUT /api/movies/:id` - Update movie (admin)
- `DELETE /api/movies/:id` - Delete movie (admin)

### 3. Scheduler Service
**Responsibilities:**
- Automatic showtime generation
- Screen availability management
- Conflict detection and resolution
- Schedule optimization algorithms
- Manual override capabilities

**Endpoints:**
- `POST /api/schedule/generate` - Auto-generate schedule
- `GET /api/schedule/cinema/:cinemaId` - Get cinema schedule
- `POST /api/schedule/showtime` - Create manual showtime
- `PUT /api/schedule/showtime/:id` - Update showtime
- `DELETE /api/schedule/showtime/:id` - Cancel showtime

### 4. Booking Service
**Responsibilities:**
- Seat reservation and locking
- Booking creation and management
- Booking cancellation and refunds
- Seat map generation
- Concurrent booking handling with locks

**Endpoints:**
- `GET /api/showtimes/:id/seats` - Get seat availability
- `POST /api/bookings/hold` - Hold seats temporarily
- `POST /api/bookings` - Confirm booking
- `GET /api/bookings/:id` - Get booking details
- `DELETE /api/bookings/:id` - Cancel booking

### 5. Pricing Service
**Responsibilities:**
- Calculate ticket prices dynamically
- Apply discounts and promotions
- Time-based pricing rules
- Seat category pricing
- Revenue optimization

**Endpoints:**
- `POST /api/pricing/calculate` - Calculate price for seats
- `GET /api/pricing/rules` - Get pricing rules
- `POST /api/pricing/promocode/validate` - Validate promo code

### 6. Payment Service
**Responsibilities:**
- Payment gateway integration
- Payment processing
- Refund handling
- Invoice generation
- Payment verification

**Endpoints:**
- `POST /api/payments/process` - Process payment
- `POST /api/payments/refund` - Issue refund
- `GET /api/payments/:id/invoice` - Get invoice

## Database Schema Overview

### Core Tables
1. **users** - User accounts and authentication
2. **cinemas** - Cinema locations
3. **screens** - Individual screens within cinemas
4. **movies** - Movie information
5. **showtimes** - Scheduled movie screenings
6. **seats** - Seat configurations per screen
7. **bookings** - Customer reservations
8. **booking_seats** - Junction table for booked seats
9. **payments** - Payment transactions
10. **pricing_rules** - Dynamic pricing configuration

## Caching Strategy

### Redis Cache Keys
- `movie:{id}` - Movie details (TTL: 1 hour)
- `showtime:{id}:seats` - Seat availability (TTL: 30 seconds)
- `user:{id}:session` - User session data
- `schedule:cinema:{id}:date:{date}` - Daily schedule (TTL: 15 minutes)

### Cache Invalidation
- On movie update → Clear `movie:{id}`
- On booking → Clear `showtime:{id}:seats`
- On schedule change → Clear affected schedule caches

## Concurrency Handling

### Seat Locking Mechanism
1. User selects seats → 5-minute temporary lock via Redis
2. Lock stored as: `lock:showtime:{id}:seat:{seatId}` with user session ID
3. During checkout → Verify locks still held by user
4. On payment success → Convert locks to permanent bookings
5. On expiry/cancel → Automatically release locks

### Race Condition Prevention
- Optimistic locking with version fields in database
- Distributed locks using Redis for critical operations
- Transaction-based seat allocation

## Scalability Considerations

### Horizontal Scaling
- Stateless API servers behind load balancer
- Session data in Redis (shared state)
- Database read replicas for queries
- CDN for static assets

### Performance Optimization
- Database indexing on frequently queried fields
- Connection pooling
- Query optimization and proper JOIN usage
- Background jobs for heavy operations (email, reports)

## Security Architecture

### Authentication Flow
1. User login → Validate credentials
2. Generate access token (15 min expiry) + refresh token (7 days)
3. Access token in Authorization header
4. Refresh token in httpOnly cookie
5. Token refresh flow for extended sessions

### Authorization
- Role-based permissions (Customer, Staff, Manager, Admin)
- Resource-level permissions
- API endpoint protection with middleware

### Data Security
- Password hashing with bcrypt (salt rounds: 10)
- Encrypted sensitive data at rest
- HTTPS/TLS for data in transit
- SQL injection prevention via ORM
- XSS protection with input sanitization