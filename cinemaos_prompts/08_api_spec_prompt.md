# API Specification Prompt

## API Design Principles
- RESTful architecture
- JSON request/response format
- JWT-based authentication
- Consistent error handling
- API versioning (v1)
- OpenAPI/Swagger documentation

## Base URL
```
Production: https://api.cinemaos.com/v1
Development: http://localhost:5000/v1
```

## Authentication

### Register
```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "SecurePass123!",
  "firstName": "John",
  "lastName": "Doe",
  "phone": "+1234567890"
}

Response 201:
{
  "success": true,
  "message": "Registration successful. Please verify your email.",
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "firstName": "John",
    "lastName": "Doe"
  }
}
```

### Login
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "SecurePass123!"
}

Response 200:
{
  "success": true,
  "accessToken": "eyJhbGciOiJIUzI1NiIs...",
  "refreshToken": "eyJhbGciOiJIUzI1NiIs...",
  "expiresIn": 900,
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "firstName": "John",
    "lastName": "Doe",
    "role": "CUSTOMER"
  }
}
```

## Movies

### Get Now Showing Movies
```http
GET /api/v1/movies/now-showing?cinemaId={id}&page=1&limit=20

Response 200:
{
  "success": true,
  "data": [
    {
      "id": "movie-uuid",
      "title": "Inception",
      "duration": 148,
      "rating": "PG-13",
      "genres": ["Action", "Sci-Fi", "Thriller"],
      "posterUrl": "https://...",
      "releaseDate": "2010-07-16"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 15,
    "totalPages": 1
  }
}
```

### Get Movie Details
```http
GET /api/v1/movies/{movieId}

Response 200:
{
  "success": true,
  "data": {
    "id": "movie-uuid",
    "title": "Inception",
    "description": "A thief who steals corporate secrets...",
    "duration": 148,
    "releaseDate": "2010-07-16",
    "rating": "PG-13",
    "language": "English",
    "genres": ["Action", "Sci-Fi", "Thriller"],
    "director": "Christopher Nolan",
    "cast": ["Leonardo DiCaprio", "Joseph Gordon-Levitt"],
    "posterUrl": "https://...",
    "backdropUrl": "https://...",
    "trailerUrl": "https://youtube.com/...",
    "format": "STANDARD"
  }
}
```

## Showtimes

### Get Showtimes for Movie
```http
GET /api/v1/showtimes/movie/{movieId}?cinemaId={id}&date=2025-12-01

Response 200:
{
  "success": true,
  "data": [
    {
      "id": "showtime-uuid",
      "movieId": "movie-uuid",
      "screenId": "screen-uuid",
      "cinemaId": "cinema-uuid",
      "showDate": "2025-12-01",
      "startTime": "14:30",
      "endTime": "17:00",
      "totalSeats": 150,
      "availableSeats": 87,
      "priceTier": "MATINEE",
      "screen": {
        "name": "Screen 1",
        "screenType": "STANDARD"
      }
    }
  ]
}
```

### Get Seat Map
```http
GET /api/v1/showtimes/{showtimeId}/seats

Response 200:
{
  "success": true,
  "data": {
    "showtimeId": "showtime-uuid",
    "totalSeats": 150,
    "available Seats": 87,
    "layout": {
      "rows": [
        {
          "label": "A",
          "seats": [
            {
              "id": "seat-uuid",
              "seatNumber": 1,
              "type": "STANDARD",
              "status": "AVAILABLE",
              "price": 12.00,
              "position": {"x": 0, "y": 0}
            }
          ]
        }
      ]
    }
  }
}
```

## Bookings

### Create Seat Hold
```http
POST /api/v1/bookings/hold
Authorization: Bearer {token}
Content-Type: application/json

{
  "showtimeId": "showtime-uuid",
  "seatIds": ["seat-1", "seat-2"]
}

Response 200:
{
  "success": true,
  "data": {
    "holdId": "hold-uuid",
    "expiresAt": "2025-12-01T15:35:00Z",
    "seats": [...],
    "pricing": {
      "subtotal": 24.00,
      "tax": 1.92,
      "total": 25.92
    }
  }
}
```

### Confirm Booking
```http
POST /api/v1/bookings/confirm
Authorization: Bearer {token}
Content-Type: application/json

{
  "holdId": "hold-uuid",
  "promoCode": "SAVE20",
  "paymentMethod": "CARD"
}

Response 201:
{
  "success": true,
  "data": {
    "bookingId": "booking-uuid",
    "bookingReference": "BK20251201ABCD",
    "status": "CONFIRMED",
    "finalAmount": 20.73,
    "paymentUrl": "https://checkout.stripe.com/..."
  }
}
```

### Get Booking Details
```http
GET /api/v1/bookings/{bookingId}
Authorization: Bearer {token}

Response 200:
{
  "success": true,
  "data": {
    "id": "booking-uuid",
    "bookingReference": "BK20251201ABCD",
    "status": "CONFIRMED",
    "showtime": {
      "movie": {
        "title": "Inception",
        "posterUrl": "..."
      },
      "showDate": "2025-12-01",
      "startTime": "14:30",
      "cinema": {
        "name": "CinemaOS Downtown",
        "address": "123 Main St"
      },
      "screen": {
        "name": "Screen 1"
      }
    },
    "seats": [
      {
        "row": "A",
        "seatNumber": 5,
        "type": "STANDARD"
      }
    ],
    "pricing": {
      "subtotal": 24.00,
      "discount": 4.80,
      "tax": 1.54,
      "total": 20.74
    },
    "bookedAt": "2025-11-29T10:30:00Z"
  }
}
```

### Cancel Booking
```http
DELETE /api/v1/bookings/{bookingId}
Authorization: Bearer {token}

{
  "reason": "Unable to attend"
}

Response 200:
{
  "success": true,
  "message": "Booking cancelled successfully",
  "refundAmount": 20.74,
  "refundStatus": "PENDING"
}
```

## Payments

### Process Payment
```http
POST /api/v1/payments/process
Authorization: Bearer {token}
Content-Type: application/json

{
  "bookingId": "booking-uuid",
  "paymentMethodId": "pm_stripe_id"
}

Response 200:
{
  "success": true,
  "data": {
    "paymentId": "payment-uuid",
    "status": "SUCCESS",
    "amount": 20.74,
    "transactionId": "txn_123456"
  }
}
```

## Admin APIs

### Create Movie
```http
POST /api/v1/admin/movies
Authorization: Bearer {adminToken}
Content-Type: application/json

{
  "title": "New Movie",
  "description": "Description...",
  "duration": 120,
  "releaseDate": "2025-12-25",
  "rating": "PG-13",
  "genres": ["Action"],
  "director": "Director Name",
  "format": "STANDARD"
}

Response 201:
{
  "success": true,
  "data": {
    "id": "movie-uuid",
    "title": "New Movie",
    ...
  }
}
```

### Generate Schedule
```http
POST /api/v1/admin/schedule/generate
Authorization: Bearer {adminToken}
Content-Type: application/json

{
  "cinemaId": "cinema-uuid",
  "date": "2025-12-01",
  "movieIds": ["movie-1", "movie-2"],
  "preferences": {
    "minShowsPerMovie": 3,
    "maxShowsPerMovie": 6
  }
}

Response 200:
{
  "success": true,
  "data": {
    "totalShows": 24,
    "metrics": {
      "avgUtilization": 87.5
    }
  }
}
```

## Error Responses

### Standard Error Format
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid input data",
    "details": [
      {
        "field": "email",
        "message": "Invalid email format"
      }
    ]
  }
}
```

### HTTP Status Codes
- `200 OK` - Successful request
- `201 Created` - Resource created
- `400 Bad Request` - Invalid input
- `401 Unauthorized` - Missing/invalid token
- `403 Forbidden` - Insufficient permissions
- `404 Not Found` - Resource not found
- `409 Conflict` - Resource conflict (e.g., seat already booked)
- `422 Unprocessable Entity` - Validation error
- `429 Too Many Requests` - Rate limit exceeded
- `500 Internal Server Error` - Server error

## Rate Limiting
- General APIs: 100 requests/minute
- Booking APIs: 10 requests/minute
- Admin APIs: 200 requests/minute

## Pagination
```
?page=1&limit=20&sortBy=createdAt&order=desc
```

## WebSocket Events
```javascript
// Connect
socket.on('connect', () => {
  socket.emit('watch:showtime', 'showtime-uuid');
});

// Listen for seat updates
socket.on('seats:updated', (data) => {
  console.log('Seats updated:', data.seatIds);
});
```