# Database Schema Prompt

## Database Technology
- **Primary Database**: PostgreSQL 15+
- **Cache Layer**: Redis 7+
- **ORM**: Prisma (for type safety and migrations)

## Core Schema

### 1. Users & Authentication

```sql
CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email VARCHAR(255) UNIQUE NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  first_name VARCHAR(100) NOT NULL,
  last_name VARCHAR(100) NOT NULL,
  phone VARCHAR(20),
  role VARCHAR(20) NOT NULL DEFAULT 'CUSTOMER', -- CUSTOMER, STAFF, MANAGER, ADMIN
  email_verified BOOLEAN DEFAULT FALSE,
  is_active BOOLEAN DEFAULT TRUE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  last_login_at TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);

CREATE TABLE refresh_tokens (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash VARCHAR(255) NOT NULL,
  expires_at TIMESTAMP NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  revoked BOOLEAN DEFAULT FALSE
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);
```

### 2. Cinema & Screens

```sql
CREATE TABLE cinemas (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(200) NOT NULL,
  address TEXT NOT NULL,
  city VARCHAR(100) NOT NULL,
  state VARCHAR(100),
  postal_code VARCHAR(20),
  country VARCHAR(100) NOT NULL,
  phone VARCHAR(20),
  email VARCHAR(255),
  operating_hours JSONB, -- {"monday": {"open": "09:00", "close": "01:00"}, ...}
  is_active BOOLEAN DEFAULT TRUE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE screens (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  cinema_id UUID NOT NULL REFERENCES cinemas(id) ON DELETE CASCADE,
  name VARCHAR(100) NOT NULL, -- e.g., "Screen 1", "IMAX Hall"
  capacity INTEGER NOT NULL,
  screen_type VARCHAR(50) NOT NULL, -- STANDARD, IMAX, 4DX
  supported_formats TEXT[] DEFAULT ARRAY['STANDARD'], -- ['STANDARD', '3D', 'IMAX']
  seat_layout JSONB NOT NULL, -- Detailed seat configuration
  is_active BOOLEAN DEFAULT TRUE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(cinema_id, name)
);

CREATE INDEX idx_screens_cinema_id ON screens(cinema_id);
```

### 3. Seats

```sql
CREATE TABLE seats (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  screen_id UUID NOT NULL REFERENCES screens(id) ON DELETE CASCADE,
  row_label VARCHAR(10) NOT NULL, -- A, B, C, etc.
  seat_number INTEGER NOT NULL,
  seat_type VARCHAR(50) DEFAULT 'STANDARD', -- STANDARD, PREMIUM, VIP, WHEELCHAIR
  is_active BOOLEAN DEFAULT TRUE,
  x_position DECIMAL(5,2), -- For visual layout
  y_position DECIMAL(5,2),
  UNIQUE(screen_id, row_label, seat_number)
);

CREATE INDEX idx_seats_screen_id ON seats(screen_id);
CREATE INDEX idx_seats_type ON seats(seat_type);
```

### 4. Movies

```sql
CREATE TABLE movies (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tmdb_id INTEGER UNIQUE, -- The Movie Database API ID
  title VARCHAR(300) NOT NULL,
  original_title VARCHAR(300),
  description TEXT,
  duration INTEGER NOT NULL, -- In minutes
  release_date DATE NOT NULL,
  rating VARCHAR(10), -- G, PG, PG-13, R, etc.
  language VARCHAR(50),
  genres TEXT[], -- ['Action', 'Thriller']
  director VARCHAR(200),
  cast TEXT[], -- Array of actor names
  poster_url TEXT,
  backdrop_url TEXT,
  trailer_url TEXT,
  format VARCHAR(50) DEFAULT 'STANDARD', -- STANDARD, 3D, IMAX, 4DX
  is_active BOOLEAN DEFAULT TRUE,
  popularity_score DECIMAL(3,1) DEFAULT 5.0, -- 1-10 rating
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_movies_release_date ON movies(release_date);
CREATE INDEX idx_movies_is_active ON movies(is_active);
CREATE INDEX idx_movies_format ON movies(format);
```

### 5. Showtimes

```sql
CREATE TABLE showtimes (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  screen_id UUID NOT NULL REFERENCES screens(id) ON DELETE CASCADE,
  movie_id UUID NOT NULL REFERENCES movies(id) ON DELETE CASCADE,
  cinema_id UUID NOT NULL REFERENCES cinemas(id) ON DELETE CASCADE,
  show_date DATE NOT NULL,
  start_time TIME NOT NULL,
  end_time TIME NOT NULL,
  price_tier VARCHAR(50) DEFAULT 'STANDARD', -- STANDARD, PEAK, MATINEE
  total_seats INTEGER NOT NULL,
  available_seats INTEGER NOT NULL,
  is_auto_generated BOOLEAN DEFAULT FALSE,
  status VARCHAR(50) DEFAULT 'SCHEDULED', -- SCHEDULED, ONGOING, COMPLETED, CANCELLED
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  version INTEGER DEFAULT 0 -- For optimistic locking
);

CREATE INDEX idx_showtimes_screen_date ON showtimes(screen_id, show_date);
CREATE INDEX idx_showtimes_movie_date ON showtimes(movie_id, show_date);
CREATE INDEX idx_showtimes_cinema_date ON showtimes(cinema_id, show_date);
CREATE INDEX idx_showtimes_status ON showtimes(status);

-- Prevent overlapping showtimes on same screen
CREATE UNIQUE INDEX idx_showtimes_no_overlap ON showtimes(screen_id, show_date, start_time)
WHERE status != 'CANCELLED';
```

### 6. Bookings

```sql
CREATE TABLE bookings (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  booking_reference VARCHAR(20) UNIQUE NOT NULL, -- e.g., "BK20251201ABCD"
  user_id UUID REFERENCES users(id) ON DELETE SET NULL, -- Nullable for guest bookings
  showtime_id UUID NOT NULL REFERENCES showtimes(id) ON DELETE CASCADE,
  guest_email VARCHAR(255), -- For guest bookings
  guest_name VARCHAR(200),
  guest_phone VARCHAR(20),
  num_tickets INTEGER NOT NULL,
  total_amount DECIMAL(10,2) NOT NULL,
  discount_amount DECIMAL(10,2) DEFAULT 0,
  final_amount DECIMAL(10,2) NOT NULL,
  promo_code VARCHAR(50),
  booking_status VARCHAR(50) DEFAULT 'PENDING', -- PENDING, CONFIRMED, COMPLETED, CANCELLED, REFUNDED
  payment_status VARCHAR(50) DEFAULT 'PENDING', -- PENDING, PAID, FAILED, REFUNDED
  payment_method VARCHAR(50), -- CARD, UPI, WALLET, CASH
  booked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  expires_at TIMESTAMP, -- For temporary holds
  confirmed_at TIMESTAMP,
  cancelled_at TIMESTAMP,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_bookings_user_id ON bookings(user_id);
CREATE INDEX idx_bookings_showtime_id ON bookings(showtime_id);
CREATE INDEX idx_bookings_reference ON bookings(booking_reference);
CREATE INDEX idx_bookings_status ON bookings(booking_status);
CREATE INDEX idx_bookings_expires_at ON bookings(expires_at) WHERE booking_status = 'PENDING';
```

### 7. Booking Seats (Junction Table)

```sql
CREATE TABLE booking_seats (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  booking_id UUID NOT NULL REFERENCES bookings(id) ON DELETE CASCADE,
  seat_id UUID NOT NULL REFERENCES seats(id) ON DELETE CASCADE,
  showtime_id UUID NOT NULL REFERENCES showtimes(id) ON DELETE CASCADE,
  price DECIMAL(10,2) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(showtime_id, seat_id) -- Prevent double booking
);

CREATE INDEX idx_booking_seats_booking_id ON booking_seats(booking_id);
CREATE INDEX idx_booking_seats_showtime_seat ON booking_seats(showtime_id, seat_id);
```

### 8. Payments

```sql
CREATE TABLE payments (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  booking_id UUID NOT NULL REFERENCES bookings(id) ON DELETE CASCADE,
  payment_reference VARCHAR(100) UNIQUE NOT NULL,
  payment_gateway VARCHAR(50), -- STRIPE, RAZORPAY, PAYPAL
  gateway_transaction_id VARCHAR(255),
  amount DECIMAL(10,2) NOT NULL,
  currency VARCHAR(10) DEFAULT 'USD',
  payment_status VARCHAR(50) DEFAULT 'PENDING', -- PENDING, SUCCESS, FAILED, REFUNDED
  payment_method VARCHAR(50),
  card_last_four VARCHAR(4),
  failure_reason TEXT,
  gateway_response JSONB, -- Raw response from payment gateway
  paid_at TIMESTAMP,
  refunded_at TIMESTAMP,
  refund_amount DECIMAL(10,2),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_payments_booking_id ON payments(booking_id);
CREATE INDEX idx_payments_status ON payments(payment_status);
CREATE INDEX idx_payments_gateway_tx_id ON payments(gateway_transaction_id);
```

### 9. Pricing Rules

```sql
CREATE TABLE pricing_rules (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  cinema_id UUID REFERENCES cinemas(id) ON DELETE CASCADE,
  name VARCHAR(200) NOT NULL,
  rule_type VARCHAR(50) NOT NULL, -- BASE, TIME_BASED, SEAT_TYPE, DAY_OF_WEEK
  seat_type VARCHAR(50), -- STANDARD, PREMIUM, VIP
  day_of_week INTEGER[], -- [1,2,3,4,5,6,7] for Mon-Sun
  time_range JSONB, -- {"start": "18:00", "end": "23:00"}
  price_modifier DECIMAL(10,2) NOT NULL, -- Fixed price or percentage
  modifier_type VARCHAR(20) DEFAULT 'FIXED', -- FIXED, PERCENTAGE
  priority INTEGER DEFAULT 0, -- Higher priority rules applied first
  is_active BOOLEAN DEFAULT TRUE,
  valid_from DATE,
  valid_until DATE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_pricing_rules_cinema_id ON pricing_rules(cinema_id);
CREATE INDEX idx_pricing_rules_active ON pricing_rules(is_active);
```

### 10. Promocodes

```sql
CREATE TABLE promocodes (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  code VARCHAR(50) UNIQUE NOT NULL,
  description TEXT,
  discount_type VARCHAR(20) NOT NULL, -- PERCENTAGE, FIXED
  discount_value DECIMAL(10,2) NOT NULL,
  max_discount DECIMAL(10,2), -- Max discount for percentage type
  min_purchase DECIMAL(10,2), -- Minimum booking amount
  usage_limit INTEGER, -- NULL for unlimited
  usage_count INTEGER DEFAULT 0,
  valid_from TIMESTAMP NOT NULL,
  valid_until TIMESTAMP NOT NULL,
  is_active BOOLEAN DEFAULT TRUE,
  applicable_cinemas UUID[], -- Array of cinema IDs, NULL for all
  applicable_movies UUID[], -- Array of movie IDs, NULL for all
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_promocodes_code ON promocodes(code);
CREATE INDEX idx_promocodes_active ON promocodes(is_active, valid_from, valid_until);
```

## Prisma Schema

```prisma
// prisma/schema.prisma
generator client {
  provider = "prisma-client-js"
}

datasource db {
  provider = "postgresql"
  url      = env("DATABASE_URL")
}

model User {
  id            String   @id @default(dbgenerated("gen_random_uuid()")) @db.Uuid
  email         String   @unique
  passwordHash  String   @map("password_hash")
  firstName     String   @map("first_name")
  lastName      String   @map("last_name")
  phone         String?
  role          Role     @default(CUSTOMER)
  emailVerified Boolean  @default(false) @map("email_verified")
  isActive      Boolean  @default(true) @map("is_active")
  createdAt     DateTime @default(now()) @map("created_at")
  updatedAt     DateTime @updatedAt @map("updated_at")
  lastLoginAt   DateTime? @map("last_login_at")
  
  bookings      Booking[]
  refreshTokens RefreshToken[]
  
  @@map("users")
  @@index([email])
  @@index([role])
}

enum Role {
  CUSTOMER
  STAFF
  MANAGER
  ADMIN
}

// ... additional models following the SQL schema above
```

## Migrations Strategy

### Initial Migration
```bash
npx prisma migrate dev --name init
```

### Seed Data
```typescript
// prisma/seed.ts
import { PrismaClient } from '@prisma/client';

const prisma = new PrismaClient();

async function main() {
  // Create admin user
  await prisma.user.create({
    data: {
      email: 'admin@cinemaos.com',
      passwordHash: '...', // Hashed password
      firstName: 'Admin',
      lastName: 'User',
      role: 'ADMIN'
    }
  });
  
  // Create sample cinema
  const cinema = await prisma.cinema.create({
    data: {
      name: 'CinemaOS Downtown',
      address: '123 Main St',
      city: 'New York',
      country: 'USA',
      operatingHours: {
        monday: { open: '09:00', close: '01:00' }
        // ... other days
      }
    }
  });
  
  // Create screens with seats
  // ... seed logic
}

main()
  .catch(console.error)
  .finally(() => prisma.$disconnect());
```

## Database Indexes Strategy

### Query Optimization Indexes
- Composite indexes for common query patterns
- Partial indexes for filtered queries
- GIN indexes for JSON/array columns if needed

### Example Queries

```sql
-- Get available seats for a showtime
SELECT s.* FROM seats s
WHERE s.screen_id = $1
  AND s.is_active = true
  AND s.id NOT IN (
    SELECT seat_id FROM booking_seats
    WHERE showtime_id = $2
      AND booking_id IN (
        SELECT id FROM bookings
        WHERE booking_status IN ('CONFIRMED', 'PENDING')
      )
  );

-- Get user booking history
SELECT b.*, st.start_time, st.show_date, m.title
FROM bookings b
JOIN showtimes st ON b.showtime_id = st.id
JOIN movies m ON st.movie_id = m.id
WHERE b.user_id = $1
ORDER BY b.created_at DESC
LIMIT 10;
```

## Data Retention Policy

- **Bookings**: Keep for 2 years
- **Payments**: Keep indefinitely for audit
- **Showtimes**: Archive after 30 days of completion
- **Refresh Tokens**: Auto-delete expired tokens daily