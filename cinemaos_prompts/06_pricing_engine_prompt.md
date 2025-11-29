# Pricing Engine Prompt

## Overview
The Pricing Engine calculates dynamic ticket prices based on multiple factors including time of day, day of week, seat type, demand, and promotional offers.

## Pricing Strategy

### Base Pricing Model

```typescript
interface PriceCalculationInput {
  showtimeId: string;
  seatIds: string[];
  promoCode?: string;
  userId?: string; // For member discounts
}

interface PriceBreakdown {
  seatPrices: SeatPrice[];
  subtotal: number;
  discounts: Discount[];
  totalDiscount: number;
  tax: number;
  finalAmount: number;
}

interface SeatPrice {
  seatId: string;
  basePrice: number;
  seatTypeModifier: number;
  timeModifier: number;
  dayModifier: number;
  demandModifier: number;
  finalPrice: number;
  breakdown: string[]; // Human-readable pricing steps
}
```

### Pricing Factors

#### 1. Base Price
- Set per cinema location
- Varies by movie format (2D, 3D, IMAX)
- Example: Standard 2D = $10, 3D = $14, IMAX = $18

#### 2. Seat Type Modifier
```typescript
const SEAT_TYPE_MODIFIERS = {
  STANDARD: 0,        // No change
  PREMIUM: 3.00,      // +$3
  VIP: 6.00,          // +$6
  WHEELCHAIR: 0,      // No change
  COUPLE: 5.00        // +$5 (for couple seats)
};
```

#### 3. Time-Based Pricing
```typescript
const TIME_MODIFIERS = {
  MATINEE: -2.00,      // Before 5 PM: -$2
  PRIME: 2.00,         // 6 PM - 9 PM: +$2
  LATE_NIGHT: 0,       // After 10 PM: no change
};

function getTimeModifier(startTime: string): number {
  const hour = parseInt(startTime.split(':')[0]);
  
  if (hour < 17) return TIME_MODIFIERS.MATINEE;
  if (hour >= 18 && hour <= 21) return TIME_MODIFIERS.PRIME;
  return TIME_MODIFIERS.LATE_NIGHT;
}
```

#### 4. Day of Week Modifier
```typescript
const DAY_MODIFIERS = {
  WEEKEND: 2.00,       // Friday-Sunday: +$2
  WEEKDAY: 0,          // Monday-Thursday: no change
  HOLIDAY: 3.00        // Public holidays: +$3
};

function getDayModifier(date: Date): number {
  const dayOfWeek = date.getDay(); // 0=Sunday, 6=Saturday
  
  // Check if it's a holiday
  if (isHoliday(date)) return DAY_MODIFIERS.HOLIDAY;
  
  // Weekend (Friday to Sunday)
  if (dayOfWeek === 0 || dayOfWeek === 5 || dayOfWeek === 6) {
    return DAY_MODIFIERS.WEEKEND;
  }
  
  return DAY_MODIFIERS.WEEKDAY;
}
```

#### 5. Demand-Based Pricing (Dynamic)
```typescript
async function getDemandModifier(showtimeId: string): Promise<number> {
  const showtime = await prisma.showtime.findUnique({
    where: { id: showtimeId },
    select: { totalSeats: true, availableSeats: true }
  });
  
  if (!showtime) throw new Error('Showtime not found');
  
  const occupancyRate = 
    (showtime.totalSeats - showtime.availableSeats) / showtime.totalSeats;
  
  // Progressive pricing based on occupancy
  if (occupancyRate >= 0.90) return 4.00;  // >90% full: +$4
  if (occupancyRate >= 0.75) return 2.00;  // >75% full: +$2
  if (occupancyRate >= 0.50) return 1.00;  // >50% full: +$1
  
  return 0; // Low demand: no change
}
```

## Price Calculation Flow

```typescript
class PricingEngine {
  async calculatePrice(input: PriceCalculationInput): Promise<PriceBreakdown> {
    // 1. Get showtime details
    const showtime = await this.getShowtimeDetails(input.showtimeId);
    
    // 2. Get seat details
    const seats = await this.getSeats(input.seatIds);
    
    // 3. Get base price for movie format
    const basePrice = await this.getBasePrice(
      showtime.cinemaId, 
      showtime.movie.format
    );
    
    // 4. Calculate price for each seat
    const seatPrices = await Promise.all(
      seats.map(async (seat) => {
        const breakdown: string[] = [];
        let price = basePrice;
        breakdown.push(`Base price: $${price.toFixed(2)}`);
        
        // Apply seat type modifier
        const seatModifier = SEAT_TYPE_MODIFIERS[seat.type] || 0;
        price += seatModifier;
        if (seatModifier !== 0) {
          breakdown.push(`${seat.type} seat: +$${seatModifier.toFixed(2)}`);
        }
        
        // Apply time modifier
        const timeModifier = getTimeModifier(showtime.startTime);
        price += timeModifier;
        if (timeModifier < 0) {
          breakdown.push(`Matinee discount: $${timeModifier.toFixed(2)}`);
        } else if (timeModifier > 0) {
          breakdown.push(`Prime time: +$${timeModifier.toFixed(2)}`);
        }
        
        // Apply day modifier
        const dayModifier = getDayModifier(showtime.showDate);
        price += dayModifier;
        if (dayModifier > 0) {
          breakdown.push(`Weekend/Holiday: +$${dayModifier.toFixed(2)}`);
        }
        
        // Apply demand modifier
        const demandModifier = await getDemandModifier(input.showtimeId);
        price += demandModifier;
        if (demandModifier > 0) {
          breakdown.push(`High demand: +$${demandModifier.toFixed(2)}`);
        }
        
        return {
          seatId: seat.id,
          basePrice,
          seatTypeModifier: seatModifier,
          timeModifier,
          dayModifier,
          demandModifier,
          finalPrice: Math.max(price, 0), // Never negative
          breakdown
        };
      })
    );
    
    // 5. Calculate subtotal
    const subtotal = seatPrices.reduce((sum, sp) => sum + sp.finalPrice, 0);
    
    // 6. Apply discounts
    const discounts = await this.applyDiscounts(
      subtotal,
      input.promoCode,
      input.userId,
      showtime
    );
    
    const totalDiscount = discounts.reduce((sum, d) => sum + d.amount, 0);
    
    // 7. Calculate tax (e.g., 8%)
    const taxableAmount = subtotal - totalDiscount;
    const tax = taxableAmount * 0.08;
    
    // 8. Final amount
    const finalAmount = taxableAmount + tax;
    
    return {
      seatPrices,
      subtotal: parseFloat(subtotal.toFixed(2)),
      discounts,
      totalDiscount: parseFloat(totalDiscount.toFixed(2)),
      tax: parseFloat(tax.toFixed(2)),
      finalAmount: parseFloat(finalAmount.toFixed(2))
    };
  }
}
```

## Discount System

### Promo Code Validation
```typescript
async function validatePromoCode(
  code: string,
  subtotal: number,
  showtimeId: string
): Promise<Discount | null> {
  const promo = await prisma.promocode.findFirst({
    where: {
      code: code.toUpperCase(),
      isActive: true,
      validFrom: { lte: new Date() },
      validUntil: { gte: new Date() },
      OR: [
        { usageLimit: null },
        { usageCount: { lt: prisma.promocode.fields.usageLimit } }
      ]
    }
  });
  
  if (!promo) return null;
  
  // Check minimum purchase requirement
  if (promo.minPurchase && subtotal < promo.minPurchase) {
    throw new Error(`Minimum purchase of $${promo.minPurchase} required`);
  }
  
  // Check cinema applicability
  const showtime = await prisma.showtime.findUnique({
    where: { id: showtimeId },
    select: { cinemaId: true, movieId: true }
  });
  
  if (promo.applicableCinemas?.length > 0) {
    if (!promo.applicableCinemas.includes(showtime.cinemaId)) {
      throw new Error('Promo code not valid for this cinema');
    }
  }
  
  if (promo.applicableMovies?.length > 0) {
    if (!promo.applicableMovies.includes(showtime.movieId)) {
      throw new Error('Promo code not valid for this movie');
    }
  }
  
  // Calculate discount
  let discountAmount = 0;
  
  if (promo.discountType === 'PERCENTAGE') {
    discountAmount = subtotal * (promo.discountValue / 100);
    
    // Apply max discount cap if specified
    if (promo.maxDiscount && discountAmount > promo.maxDiscount) {
      discountAmount = promo.maxDiscount;
    }
  } else {
    // FIXED discount
    discountAmount = promo.discountValue;
  }
  
  // Ensure discount doesn't exceed subtotal
  discountAmount = Math.min(discountAmount, subtotal);
  
  return {
    type: 'PROMO_CODE',
    code: promo.code,
    description: promo.description || `Promo: ${promo.code}`,
    amount: parseFloat(discountAmount.toFixed(2))
  };
}
```

### Member Discounts
```typescript
const MEMBER_DISCOUNTS = {
  BRONZE: 0.05,  // 5% off
  SILVER: 0.10,  // 10% off
  GOLD: 0.15,    // 15% off
  PLATINUM: 0.20 // 20% off
};

async function getMemberDiscount(
  userId: string,
  subtotal: number
): Promise<Discount | null> {
  const user = await prisma.user.findUnique({
    where: { id: userId },
    include: { membership: true }
  });
  
  if (!user?.membership || !user.membership.isActive) return null;
  
  const discountRate = MEMBER_DISCOUNTS[user.membership.tier] || 0;
  const amount = subtotal * discountRate;
  
  return {
    type: 'MEMBERSHIP',
    code: user.membership.tier,
    description: `${user.membership.tier} Member Discount`,
    amount: parseFloat(amount.toFixed(2))
  };
}
```

### Group Booking Discount
```typescript
function getGroupDiscount(numTickets: number, subtotal: number): Discount | null {
  if (numTickets >= 20) {
    // 20+ tickets: 20% off
    return {
      type: 'GROUP',
      code: 'GROUP_20',
      description: 'Group booking discount (20+ tickets)',
      amount: parseFloat((subtotal * 0.20).toFixed(2))
    };
  } else if (numTickets >= 10) {
    // 10-19 tickets: 10% off
    return {
      type: 'GROUP',
      code: 'GROUP_10',
      description: 'Group booking discount (10+ tickets)',
      amount: parseFloat((subtotal * 0.10).toFixed(2))
    };
  }
  
  return null;
}
```

## API Endpoints

### Calculate Price
```http
POST /api/pricing/calculate
Content-Type: application/json

{
  "showtimeId": "uuid",
  "seatIds": ["seat-1", "seat-2"],
  "promoCode": "SAVE20",
  "userId": "user-uuid"
}

Response:
{
  "seatPrices": [
    {
      "seatId": "seat-1",
      "basePrice": 10.00,
      "finalPrice": 14.00,
      "breakdown": [
        "Base price: $10.00",
        "PREMIUM seat: +$3.00",
        "Prime time: +$2.00",
        "Weekend: +$2.00",
        "High demand: +$1.00"
      ]
    }
  ],
  "subtotal": 28.00,
  "discounts": [
    {
      "type": "PROMO_CODE",
      "code": "SAVE20",
      "description": "20% off discount",
      "amount": 5.60
    }
  ],
  "totalDiscount": 5.60,
  "tax": 1.79,
  "finalAmount": 24.19
}
```

### Get Pricing Rules
```http
GET /api/pricing/rules?cinemaId={id}

Response:
{
  "rules": [
    {
      "id": "rule-1",
      "name": "Weekend Surcharge",
      "type": "DAY_OF_WEEK",
      "dayOfWeek": [5, 6, 0],
      "priceModifier": 2.00,
      "modifierType": "FIXED"
    }
  ]
}
```

## Admin Configuration

### Create Pricing Rule
```typescript
POST /api/admin/pricing/rules
Authorization: Bearer {adminToken}

{
  "cinemaId": "cinema-uuid",
  "name": "Holiday Special Pricing",
  "ruleType": "DAY_OF_WEEK",
  "dayOfWeek": [0], // Sundays
  "timeRange": {
    "start": "12:00",
    "end": "18:00"
  },
  "priceModifier": -3.00,
  "modifierType": "FIXED",
  "priority": 10,
  "validFrom": "2025-12-01",
  "validUntil": "2025-12-31"
}
```

### Create Promo Code
```typescript
POST /api/admin/promocodes
Authorization: Bearer {adminToken}

{
  "code": "NEWYEAR25",
  "description": "New Year Special - 25% off",
  "discountType": "PERCENTAGE",
  "discountValue": 25,
  "maxDiscount": 10.00,
  "minPurchase": 20.00,
  "usageLimit": 1000,
  "validFrom": "2025-12-31T00:00:00Z",
  "validUntil": "2026-01-02T23:59:59Z",
  "applicableCinemas": null, // All cinemas
  "applicableMovies": null    // All movies
}
```

## Revenue Optimization

### A/B Testing Prices
- Test different pricing strategies
- Track conversion rates
- Measure revenue impact
- Automatically adjust pricing based on results

### Price Elasticity Analysis
```sql
-- Analyze how price changes affect demand
SELECT 
  price_tier,
  AVG(final_amount / num_tickets) as avg_ticket_price,
  COUNT(*) as bookings,
  SUM(num_tickets) as tickets_sold,
  SUM(final_amount) as revenue,
  (SUM(final_amount) / COUNT(*)) as revenue_per_booking
FROM bookings
WHERE booking_status = 'CONFIRMED'
  AND booked_at >= NOW() - INTERVAL '30 days'
GROUP BY price_tier;
```

## Performance Considerations

1. **Cache Pricing Rules**: Store in Redis, invalidate on update
2. **Pre-calculate Demand**: Update demand modifier every 5 minutes
3. **Batch Calculations**: Calculate all seats in one operation
4. **Round Strategically**: Always round to 2 decimal places to avoid floating point errors