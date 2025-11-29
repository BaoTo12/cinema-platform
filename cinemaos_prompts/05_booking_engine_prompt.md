# Booking Engine Prompt

## Overview
The Booking Engine is the core component that handles seat reservations, managing concurrent bookings, preventing double bookings, and ensuring data consistency even under high load.

## Key Features

### 1. Seat Availability
- Real-time seat status display
- Visual seat map with color coding
- Seat type differentiation (Standard, Premium, VIP, Wheelchair)
- Temporary seat holds during checkout process

### 2. Booking Flow

#### Step 1: Browse & Select Showtime
```typescript
// User browses movies and showtimes
GET /api/movies/now-showing?cinema={cinemaId}&date={date}
GET /api/showtimes/movie/{movieId}?cinema={cinemaId}&date={date}
```

#### Step 2: View Seat Map
```typescript
// Fetch seat layout and availability
GET /api/showtimes/{showtimeId}/seats

Response:
{
  "showtimeId": "uuid",
  "totalSeats": 150,
  "availableSeats": 87,
  "layout": {
    "rows": [
      {
        "label": "A",
        "seats": [
          {
            "id": "seat-a1",
            "seatNumber": 1,
            "type": "STANDARD",
            "status": "AVAILABLE", // AVAILABLE, SELECTED, BOOKED, BLOCKED
            "price": 12.00
          }
        ]
      }
    ]
  }
}
```

#### Step 3: Select Seats
```typescript
// Lock seats temporarily (5 minutes)
POST /api/bookings/hold
{
  "showtimeId": "uuid",
  "seatIds": ["seat-a1", "seat-a2"],
  "sessionToken": "unique-session-id"
}

Response:
{
  "holdId": "hold-uuid",
  "expiresAt": "2025-12-01T15:35:00Z",
  "seats": [...],
  "totalAmount": 24.00
}
```

#### Step 4: Apply Promo Code (Optional)
```typescript
POST /api/pricing/calculate
{
  "holdId": "hold-uuid",
  "promoCode": "SAVE20"
}

Response:
{
  "subtotal": 24.00,
  "discount": 4.80,
  "finalAmount": 19.20
}
```

#### Step 5: Complete Booking
```typescript
POST /api/bookings/confirm
{
  "holdId": "hold-uuid",
  "userDetails": {
    "email": "user@example.com",
    "name": "John Doe",
    "phone": "+1234567890"
  },
  "paymentMethod": "CARD"
}

Response:
{
  "bookingId": "uuid",
  "bookingReference": "BK20251201ABCD",
  "status": "PENDING_PAYMENT",
  "paymentIntentId": "pi_stripe_id",
  "amount": 19.20
}
```

#### Step 6: Process Payment
```typescript
POST /api/payments/process
{
  "bookingId": "uuid",
  "paymentIntentId": "pi_stripe_id",
  "paymentMethod": {
    "type": "card",
    "token": "tok_stripe"
  }
}

Response:
{
  "success": true,
  "bookingReference": "BK20251201ABCD",
  "status": "CONFIRMED"
}
```

## Concurrency Control

### Seat Locking Mechanism

#### Redis Lock Implementation
```typescript
import Redis from 'ioredis';

const redis = new Redis(process.env.REDIS_URL);

async function lockSeats(
  showtimeId: string,
  seatIds: string[],
  sessionToken: string,
  expirySeconds: number = 300 // 5 minutes
): Promise<boolean> {
  const multi = redis.multi();
  
  for (const seatId of seatIds) {
    const lockKey = `lock:showtime:${showtimeId}:seat:${seatId}`;
    
    // Set lock only if it doesn't exist (NX)
    multi.set(lockKey, sessionToken, 'EX', expirySeconds, 'NX');
  }
  
  const results = await multi.exec();
  
  // Check if all locks were acquired
  const allLocked = results.every(([err, result]) => result === 'OK');
  
  if (!allLocked) {
    // Release any acquired locks
    await releaseSeats(showtimeId, seatIds, sessionToken);
    return false;
  }
  
  return true;
}

async function releaseSeats(
  showtimeId: string,
  seatIds: string[],
  sessionToken: string
): Promise<void> {
  const luaScript = `
    if redis.call("get", KEYS[1]) == ARGV[1] then
      return redis.call("del", KEYS[1])
    else
      return 0
    end
  `;
  
  for (const seatId of seatIds) {
    const lockKey = `lock:showtime:${showtimeId}:seat:${seatId}`;
    await redis.eval(luaScript, 1, lockKey, sessionToken);
  }
}
```

### Database Optimistic Locking

```typescript
async function confirmBooking(holdId: string): Promise<Booking> {
  const hold = await getHold(holdId);
  
  return await prisma.$transaction(async (tx) => {
    // Check showtime version (optimistic lock)
    const showtime = await tx.showtime.findUnique({
      where: { id: hold.showtimeId },
      select: { version: true, availableSeats: true }
    });
    
    if (!showtime) throw new Error('Showtime not found');
    if (showtime.availableSeats < hold.seatIds.length) {
      throw new Error('Not enough seats available');
    }
    
    // Verify seats are still locked by this session
    const locksValid = await verifyLocks(
      hold.showtimeId,
      hold.seatIds,
      hold.sessionToken
    );
    
    if (!locksValid) {
      throw new Error('Seat locks expired or invalid');
    }
    
    // Create booking
    const booking = await tx.booking.create({
      data: {
        bookingReference: generateBookingReference(),
        showtimeId: hold.showtimeId,
        userId: hold.userId,
        numTickets: hold.seatIds.length,
        totalAmount: hold.totalAmount,
        bookingStatus: 'PENDING'
      }
    });
    
    // Create booking_seats records
    await tx.bookingSeat.createMany({
      data: hold.seatIds.map((seatId, idx) => ({
        bookingId: booking.id,
        seatId,
        showtimeId: hold.showtimeId,
        price: hold.seatPrices[idx]
      }))
    });
    
    // Update showtime with version increment (optimistic lock)
    const updated = await tx.showtime.updateMany({
      where: {
        id: hold.showtimeId,
        version: showtime.version // Only update if version matches
      },
      data: {
        availableSeats: { decrement: hold.seatIds.length },
        version: { increment: 1 }
      }
    });
    
    if (updated.count === 0) {
      throw new Error('Showtime was modified, please retry');
    }
    
    return booking;
  });
}
```

## Seat Status Management

### Status Enum
```typescript
enum SeatStatus {
  AVAILABLE = 'AVAILABLE',     // Seat is free
  LOCKED = 'LOCKED',           // Temporarily held (Redis lock)
  BOOKED = 'BOOKED',           // Confirmed booking
  BLOCKED = 'BLOCKED'          // Maintenance/broken seat
}
```

### Real-time Status Calculation
```typescript
async function getSeatStatus(
  showtimeId: string,
  seatId: string
): Promise<SeatStatus> {
  // Check if permanently booked in database
  const bookedSeat = await prisma.bookingSeat.findFirst({
    where: {
      showtimeId,
      seatId,
      booking: {
        bookingStatus: { in: ['CONFIRMED', 'PENDING'] }
      }
    }
  });
  
  if (bookedSeat) return SeatStatus.BOOKED;
  
  // Check if temporarily locked in Redis
  const lockKey = `lock:showtime:${showtimeId}:seat:${seatId}`;
  const lock = await redis.get(lockKey);
  
  if (lock) return SeatStatus.LOCKED;
  
  // Check if seat is blocked
  const seat = await prisma.seat.findUnique({
    where: { id: seatId },
    select: { isActive: true }
  });
  
  if (!seat?.isActive) return SeatStatus.BLOCKED;
  
  return SeatStatus.AVAILABLE;
}
```

## Booking Cancellation & Refunds

### Cancel Booking
```typescript
POST /api/bookings/{bookingId}/cancel
{
  "reason": "User requested cancellation",
  "refundAmount": 19.20 // Optional partial refund
}

async function cancelBooking(
  bookingId: string,
  reason: string,
  refundAmount?: number
): Promise<void> {
  await prisma.$transaction(async (tx) => {
    const booking = await tx.booking.findUnique({
      where: { id: bookingId },
      include: { bookingSeats: true, showtime: true }
    });
    
    if (!booking) throw new Error('Booking not found');
    if (booking.bookingStatus === 'CANCELLED') {
      throw new Error('Already cancelled');
    }
    
    // Check cancellation policy (e.g., 2 hours before showtime)
    const showDateTime = new Date(
      `${booking.showtime.showDate} ${booking.showtime.startTime}`
    );
    const now = new Date();
    const hoursUntilShow = (showDateTime.getTime() - now.getTime()) / (1000 * 60 * 60);
    
    if (hoursUntilShow < 2) {
      throw new Error('Cannot cancel within 2 hours of showtime');
    }
    
    // Update booking status
    await tx.booking.update({
      where: { id: bookingId },
      data: {
        bookingStatus: 'CANCELLED',
        cancelledAt: new Date()
      }
    });
    
    // Release seats back to inventory
    await tx.showtime.update({
      where: { id: booking.showtimeId },
      data: {
        availableSeats: { increment: booking.numTickets }
      }
    });
    
    // Process refund if payment was made
    if (booking.paymentStatus === 'PAID' && refundAmount) {
      await processRefund(bookingId, refundAmount);
    }
  });
}
```

## Booking Notifications

### Email Templates
- **Booking Confirmation**: Sent after successful payment
- **Booking Reminder**: Sent 3 hours before showtime
- **Cancellation Confirmation**: Sent after cancellation
- **Refund Notification**: Sent after refund processing

### Queue System
```typescript
import Bull from 'bull';

const emailQueue = new Bull('emails', process.env.REDIS_URL);

emailQueue.process(async (job) => {
  const { type, bookingId, email } = job.data;
  
  const booking = await getBookingDetails(bookingId);
  
  switch (type) {
    case 'CONFIRMATION':
      await sendConfirmationEmail(email, booking);
      break;
    case 'REMINDER':
      await sendReminderEmail(email, booking);
      break;
    // ... other types
  }
});

// Add job to queue
await emailQueue.add({
  type: 'CONFIRMATION',
  bookingId: 'uuid',
  email: 'user@example.com'
}, {
  attempts: 3,
  backoff: {
    type: 'exponential',
    delay: 2000
  }
});
```

## Reporting & Analytics

### Booking Metrics
```typescript
// Daily booking report
SELECT 
  DATE(booked_at) as date,
  COUNT(*) as total_bookings,
  SUM(num_tickets) as total_tickets,
  SUM(final_amount) as total_revenue,
  AVG(final_amount / num_tickets) as avg_ticket_price
FROM bookings
WHERE booking_status = 'CONFIRMED'
  AND booked_at >= NOW() - INTERVAL '30 days'
GROUP BY DATE(booked_at)
ORDER BY date DESC;

// Popular movies
SELECT 
  m.title,
  COUNT(b.id) as bookings,
  SUM(b.num_tickets) as tickets_sold,
  SUM(b.final_amount) as revenue
FROM bookings b
JOIN showtimes st ON b.showtime_id = st.id
JOIN movies m ON st.movie_id = m.id
WHERE b.booking_status = 'CONFIRMED'
  AND b.booked_at >= NOW() - INTERVAL '7 days'
GROUP BY m.id, m.title
ORDER BY revenue DESC
LIMIT 10;
```

## WebSocket for Real-time Updates

```typescript
import { Server } from 'socket.io';

io.on('connection', (socket) => {
  socket.on('watch:showtime', (showtimeId) => {
    socket.join(`showtime:${showtimeId}`);
  });
  
  socket.on('unwatch:showtime', (showtimeId) => {
    socket.leave(`showtime:${showtimeId}`);
  });
});

// Emit seat update when booking changes
function notifySeatUpdate(showtimeId: string, seatIds: string[]) {
  io.to(`showtime:${showtimeId}`).emit('seats:updated', {
    showtimeId,
    seatIds,
    timestamp: new Date()
  });
}
```

## Performance Optimizations

1. **Database Connection Pooling**: Max 20 connections
2. **Redis Caching**: Cache seat layouts for 1 hour
3. **Rate Limiting**: 100 requests/min per IP for booking endpoints
4. **Seat Map Compression**: Gzip response for large seat layouts
5. **Lazy Loading**: Load seat prices on demand, not in initial map fetch