# Testing Prompt

## Testing Strategy

### Testing Pyramid
```
          E2E Tests (10%)
       ┌─────────────────┐
       │  User Journeys  │
       └─────────────────┘
     
    Integration Tests (30%)
  ┌──────────────────────────┐
  │   API Endpoints          │
  │   Database Operations    │
  │   Service Integration    │
  └──────────────────────────┘

     Unit Tests (60%)
┌───────────────────────────────┐
│  Business Logic               │
│  Utility Functions            │
│  Component Logic              │
└───────────────────────────────┘
```

## Backend Testing

### Unit Tests

#### Test Framework
- **Framework**: Jest
- **Assertion**: Jest Matchers
- **Mocking**: Jest Mocks

#### Example: Pricing Engine Test
```typescript
// __tests__/services/pricing.test.ts
import { PricingEngine } from '@/services/pricing';
import { PrismaClient } from '@prisma/client';

jest.mock('@prisma/client');

describe('PricingEngine', () => {
  let pricingEngine: PricingEngine;
  let prisma: jest.Mocked<PrismaClient>;

  beforeEach(() => {
    prisma = new PrismaClient() as jest.Mocked<PrismaClient>;
    pricingEngine = new PricingEngine(prisma);
  });

  describe('calculatePrice', () => {
    it('should apply matinee discount for shows before 5 PM', async () => {
      const result = await pricingEngine.calculatePrice({
        showtimeId: 'showtime-1',
        seatIds: ['seat-1'],
        startTime: '14:00',
        basePrice: 10
      });

      expect(result.seatPrices[0].timeModifier).toBe(-2);
      expect(result.seatPrices[0].finalPrice).toBe(8);
    });

    it('should apply prime time surcharge for 6-9 PM shows', async () => {
      const result = await pricingEngine.calculatePrice({
        showtimeId: 'showtime-1',
        seatIds: ['seat-1'],
        startTime: '20:00',
        basePrice: 10
      });

      expect(result.seatPrices[0].timeModifier).toBe(2);
      expect(result.seatPrices[0].finalPrice).toBe(12);
    });

    it('should apply premium seat surcharge', async () => {
      const result = await pricingEngine.calculatePrice({
        showtimeId: 'showtime-1',
        seatIds: ['premium-seat-1'],
        seatType: 'PREMIUM',
        basePrice: 10
      });

      expect(result.seatPrices[0].seatTypeModifier).toBe(3);
      expect(result.seatPrices[0].finalPrice).toBe(13);
    });

    it('should apply demand-based pricing when >75% full', async () => {
      prisma.showtime.findUnique.mockResolvedValue({
        totalSeats: 100,
        availableSeats: 20 // 80% occupied
      });

      const result = await pricingEngine.calculatePrice({
        showtimeId: 'showtime-1',
        seatIds: ['seat-1'],
        basePrice: 10
      });

      expect(result.seatPrices[0].demandModifier).toBe(2);
    });
  });

  describe('validatePromoCode', () => {
    it('should apply percentage discount correctly', async () => {
      prisma.promocode.findFirst.mockResolvedValue({
        code: 'SAVE20',
        discountType: 'PERCENTAGE',
        discountValue: 20,
        maxDiscount: null
      });

      const discount = await pricingEngine.validatePromoCode('SAVE20', 100);

      expect(discount.amount).toBe(20);
    });

    it('should cap percentage discount at maxDiscount', async () => {
      prisma.promocode.findFirst.mockResolvedValue({
        code: 'SAVE20',
        discountType: 'PERCENTAGE',
        discountValue: 20,
        maxDiscount: 10
      });

      const discount = await pricingEngine.validatePromoCode('SAVE20', 100);

      expect(discount.amount).toBe(10); // Capped
    });

    it('should throw error for expired promo code', async () => {
      prisma.promocode.findFirst.mockResolvedValue(null);

      await expect(
        pricingEngine.validatePromoCode('EXPIRED', 100)
      ).rejects.toThrow('Invalid or expired promo code');
    });
  });
});
```

#### Example: Seat Locking Test
```typescript
// __tests__/services/booking.test.ts
import { BookingService } from '@/services/booking';
import Redis from 'ioredis-mock';

describe('BookingService', () => {
  let bookingService: BookingService;
  let redis: Redis;

  beforeEach(() => {
    redis = new Redis();
    bookingService = new BookingService(redis);
  });

  describe('lockSeats', () => {
    it('should successfully lock available seats', async () => {
      const result = await bookingService.lockSeats(
        'showtime-1',
        ['seat-1', 'seat-2'],
        'session-123'
      );

      expect(result).toBe(true);
    });

    it('should fail to lock already locked seats', async () => {
      // First lock
      await bookingService.lockSeats(
        'showtime-1',
        ['seat-1'],
        'session-123'
      );

      // Second lock attempt
      const result = await bookingService.lockSeats(
        'showtime-1',
        ['seat-1'],
        'session-456'
      );

      expect(result).toBe(false);
    });

    it('should release locks after expiry', async () => {
      jest.useFakeTimers();

      await bookingService.lockSeats(
        'showtime-1',
        ['seat-1'],
        'session-123',
        2 // 2 seconds
      );

      // Fast-forward 3 seconds
      jest.advanceTimersByTime(3000);

      // Should be able to lock again
      const result = await bookingService.lockSeats(
        'showtime-1',
        ['seat-1'],
        'session-456'
      );

      expect(result).toBe(true);

      jest.useRealTimers();
    });
  });
});
```

### Integration Tests

#### API Endpoint Tests
```typescript
// __tests__/api/bookings.integration.test.ts
import request from 'supertest';
import { app } from '@/app';
import { PrismaClient } from '@prisma/client';

const prisma = new PrismaClient();

describe('Bookings API', () => {
  let authToken: string;
  let showtimeId: string;

  beforeAll(async () => {
    // Setup test database
    await prisma.$executeRaw`TRUNCATE TABLE bookings CASCADE`;
    
    // Create test user and get auth token
    const response = await request(app)
      .post('/api/v1/auth/login')
      .send({
        email: 'test@example.com',
        password: 'TestPass123!'
      });
    
    authToken = response.body.accessToken;
    
    // Create test showtime
    const showtime = await prisma.showtime.create({
      data: {
        movieId: 'movie-test-1',
        screenId: 'screen-test-1',
        cinemaId: 'cinema-test-1',
        showDate: new Date(),
        startTime: '18:00',
        endTime: '20:00',
        totalSeats: 100,
        availableSeats: 100
      }
    });
    
    showtimeId = showtime.id;
  });

  afterAll(async () => {
    await prisma.$disconnect();
  });

  describe('POST /api/v1/bookings/hold', () => {
    it('should successfully hold seats', async () => {
      const response = await request(app)
        .post('/api/v1/bookings/hold')
        .set('Authorization', `Bearer ${authToken}`)
        .send({
          showtimeId,
          seatIds: ['seat-a1', 'seat-a2']
        });

      expect(response.status).toBe(200);
      expect(response.body.success).toBe(true);
      expect(response.body.data.holdId).toBeDefined();
      expect(response.body.data.expiresAt).toBeDefined();
    });

    it('should fail without authentication', async () => {
      const response = await request(app)
        .post('/api/v1/bookings/hold')
        .send({
          showtimeId,
          seatIds: ['seat-a1']
        });

      expect(response.status).toBe(401);
    });

    it('should prevent double booking', async () => {
      // First booking
      await request(app)
        .post('/api/v1/bookings/hold')
        .set('Authorization', `Bearer ${authToken}`)
        .send({
          showtimeId,
          seatIds: ['seat-b1']
        });

      // Second booking (should fail)
      const response = await request(app)
        .post('/api/v1/bookings/hold')
        .set('Authorization', `Bearer ${authToken}`)
        .send({
          showtimeId,
          seatIds: ['seat-b1']
        });

      expect(response.status).toBe(409);
      expect(response.body.error.code).toBe('SEATS_UNAVAILABLE');
    });
  });

  describe('POST /api/v1/bookings/confirm', () => {
    it('should confirm booking and decrement available seats', async () => {
      // Hold seats first
      const holdResponse = await request(app)
        .post('/api/v1/bookings/hold')
        .set('Authorization', `Bearer ${authToken}`)
        .send({
          showtimeId,
          seatIds: ['seat-c1', 'seat-c2']
        });

      const holdId = holdResponse.body.data.holdId;

      // Confirm booking
      const confirmResponse = await request(app)
        .post('/api/v1/bookings/confirm')
        .set('Authorization', `Bearer ${authToken}`)
        .send({
          holdId,
          paymentMethod: 'CARD'
        });

      expect(confirmResponse.status).toBe(201);
      expect(confirmResponse.body.data.status).toBe('CONFIRMED');

      // Verify showtime available seats decremented
      const showtime = await prisma.showtime.findUnique({
        where: { id: showtimeId }
      });

      expect(showtime.availableSeats).toBe(98); // Started with 100, booked 2
    });
  });
});
```

### E2E Tests

#### Playwright Configuration
```typescript
// playwright.config.ts
import { defineConfig } from '@playwright/test';

export default defineConfig({
  testDir: './e2e',
  use: {
    baseURL: 'http://localhost:3000',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure'
  },
  projects: [
    { name: 'chromium', use: { browserName: 'chromium' } },
    { name: 'firefox', use: { browserName: 'firefox' } },
    { name: 'webkit', use: { browserName: 'webkit' } }
  ]
});
```

#### E2E Test Example
```typescript
// e2e/booking-flow.spec.ts
import { test, expect } from '@playwright/test';

test.describe('Complete Booking Flow', () => {
  test('user can book tickets end-to-end', async ({ page }) => {
    // 1. Navigate to home page
    await page.goto('/');
    await expect(page).toHaveTitle(/CinemaOS/);

    // 2. Select a movie
    await page.click('text=Inception');
    await expect(page).toHaveURL(/\/movies\/\w+/);

    // 3. Select showtime
    await page.click('[data-testid="showtime-18:00"]');
    await expect(page).toHaveURL(/\/booking\/seats\//);

    // 4. Select seats
    await page.click('[data-seat-id="seat-a5"]');
    await page.click('[data-seat-id="seat-a6"]');
    
    // Verify 2 seats selected
    await expect(page.locator('[data-testid="selected-count"]')).toHaveText('2');

    // 5. Proceed to checkout
    await page.click('text=Continue to Checkout');
    await expect(page).toHaveURL(/\/booking\/checkout\//);

    // 6. Apply promo code
    await page.fill('[data-testid="promo-code-input"]', 'TEST20');
    await page.click('[data-testid="apply-promo"]');
    await expect(page.locator('[data-testid="discount-amount"]')).toBeVisible();

    // 7. Fill payment details (using test mode)
    await page.fill('[data-testid="card-number"]', '4242424242424242');
    await page.fill('[data-testid="card-expiry"]', '12/25');
    await page.fill('[data-testid="card-cvc"]', '123');

    // 8. Confirm booking
    await page.click('text=Confirm & Pay');

    // 9. Verify confirmation page
    await expect(page).toHaveURL(/\/booking\/confirmation\//);
    await expect(page.locator('[data-testid="booking-reference"]')).toBeVisible();
    await expect(page.locator('text=Booking Confirmed')).toBeVisible();

    // 10. Verify QR code displayed
    await expect(page.locator('[data-testid="ticket-qr-code"]')).toBeVisible();
  });

  test('should prevent booking same seats concurrently', async ({ browser }) => {
    // Create two browser contexts (simulate two users)
    const context1 = await browser.newContext();
    const context2 = await browser.newContext();
    
    const page1 = await context1.newPage();
    const page2 = await context2.newPage();

    // Both users navigate to same showtime
    await page1.goto('/booking/seats/showtime-123');
    await page2.goto('/booking/seats/showtime-123');

    // User 1 selects seats
    await page1.click('[data-seat-id="seat-a1"]');
    await page1.click('[data-seat-id="seat-a2"]');
    await page1.click('text=Continue');

    // User 2 tries to select same seats
    await page2.click('[data-seat-id="seat-a1"]');
    
    // Should see seat is locked/unavailable
    await expect(page2.locator('[data-seat-id="seat-a1"]')).toHaveClass(/locked/);
  });
});
```

## Frontend Testing

### Component Tests (React Testing Library)
```typescript
// components/__tests__/MovieCard.test.tsx
import { render, screen, fireEvent } from '@testing-library/react';
import { MovieCard } from '@/components/movies/MovieCard';

const mockMovie = {
  id: '1',
  title: 'Test Movie',
  posterUrl: '/poster.jpg',
  rating: 8.5,
  duration: 120,
  genres: ['Action', 'Thriller']
};

describe('MovieCard', () => {
  it('renders movie information correctly', () => {
    render(<MovieCard movie={mockMovie} />);

    expect(screen.getByText('Test Movie')).toBeInTheDocument();
    expect(screen.getByText(/Action, Thriller/)).toBeInTheDocument();
    expect(screen.getByText('120 min')).toBeInTheDocument();
  });

  it('calls onClick when clicked', () => {
    const handleClick = jest.fn();
    render(<MovieCard movie={mockMovie} onClick={handleClick} />);

    fireEvent.click(screen.getByRole('button'));

    expect(handleClick).toHaveBeenCalledTimes(1);
  });

  it('displays poster image with correct alt text', () => {
    render(<MovieCard movie={mockMovie} />);

    const image = screen.getByAltText('Test Movie');
    expect(image).toHaveAttribute('src', '/poster.jpg');
  });
});
```

## Test Coverage

### Coverage Goals
- **Unit Tests**: >80% code coverage
- **Integration Tests**: All API endpoints covered
- **E2E Tests**: Critical user journeys covered

### Run Coverage
```bash
# Backend
cd backend
npm run test:coverage

# Frontend
cd frontend
npm run test:coverage
```

### Coverage Report
```bash
# Generate HTML coverage report
npm run test:coverage -- --coverage-reporters=html

# Open in browser
open coverage/index.html
```

## Continuous Integration

### GitHub Actions Test Workflow
```yaml
name: Run Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    
    steps:
      - uses: actions/checkout@v3
      
      - name: Run Unit Tests
        run: npm test
      
      - name: Run Integration Tests
        run: npm run test:integration
      
      - name: Run E2E Tests
        run: npm run test:e2e
      
      - name: Upload Coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage/lcov.info
```

## Test Commands

```json
{
  "scripts": {
    "test": "jest",
    "test:watch": "jest --watch",
    "test:coverage": "jest --coverage",
    "test:integration": "jest --config jest.integration.config.js",
    "test:e2e": "playwright test",
    "test:e2e:ui": "playwright test --ui"
  }
}
```