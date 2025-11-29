# Frontend Prompt

## Technology Stack
- **Framework**: Next.js 14 (App Router)
- **Language**: TypeScript
- **Styling**: Tailwind CSS
- **State Management**: Zustand
- **Forms**: React Hook Form + Zod
- **HTTP Client**: Axios with interceptors
- **UI Components**: Custom components with Headless UI
- **Icons**: Lucide React
- **Animations**: Framer Motion

## Project Structure
```
frontend/
├── src/
│   ├── app/
│   │   ├── (auth)/
│   │   │   ├── login/
│   │   │   └── register/
│   │   ├── (main)/
│   │   │   ├── movies/
│   │   │   ├── booking/
│   │   │   └── account/
│   │   ├── admin/
│   │   └── layout.tsx
│   ├── components/
│   │   ├── ui/          # Reusable UI components
│   │   ├── movies/      # Movie-related components
│   │   ├── booking/     # Booking flow components
│   │   └── layout/      # Layout components
│   ├── lib/
│   │   ├── api/         # API client functions
│   │   ├── hooks/       # Custom React hooks
│   │   ├── utils/       # Utility functions
│   │   └── validations/ # Zod schemas
│   ├── store/           # Zustand stores
│   ├── types/           # TypeScript types
│   └── styles/          # Global styles
├── public/
└── tailwind.config.ts
```

## Design System

### Color Palette
```typescript
// tailwind.config.ts
export default {
  theme: {
    extend: {
      colors: {
        primary: {
          50: '#fef2f2',
          100: '#fee2e2',
          500: '#ef4444',  // Main red
          600: '#dc2626',
          900: '#7f1d1d'
        },
        dark: {
          DEFAULT: '#0a0a0a',
          light: '#1a1a1a',
          lighter: '#2a2a2a'
        },
        accent: {
          gold: '#fbbf24',
          blue: '#3b82f6'
        }
      },
      fontFamily: {
        sans: ['Inter', 'sans-serif'],
        display: ['Outfit', 'sans-serif']
      }
    }
  }
};
```

### Typography
- **Headings**: Outfit font
- **Body**: Inter font
- **Size Scale**: text-sm, text-base, text-lg, text-xl, text-2xl, text-3xl, text-4xl

## Core Pages

### 1. Home Page (`/`)
**Features:**
- Hero section with featured movies
- Now showing movies grid
- Coming soon section
- Quick booking CTA

**Components:**
- `MovieCard` - Display movie poster, title, rating
- `MovieCarousel` - Horizontal scrolling movie list
- `QuickBookingBar` - Sticky bottom bar for quick access

### 2. Movie Details Page (`/movies/[movieId]`)
**Features:**
- Movie poster and backdrop
- Synopsis, cast, director info
- Trailer video player
- Available showtimes
- Reviews and ratings

**Components:**
- `MovieHero` - Full-width backdrop with overlay
- `MovieInfo` - Detailed movie information
- `ShowtimeSelector` - List of available showtimes
- `TrailerPlayer` - YouTube/video embed

### 3. Booking Flow

#### Step 1: Select Showtime (`/booking/movie/[movieId]`)
```tsx
// Date and cinema selector
// Showtime cards with pricing
<ShowtimeSelector
  movieId={movieId}
  onSelectShowtime={(showtimeId) => router.push(`/booking/seats/${showtimeId}`)}
/>
```

#### Step 2: Select Seats (`/booking/seats/[showtimeId]`)
```tsx
// Interactive seat map
<SeatMap
  showtimeId={showtimeId}
  onSeatsSelected={(seatIds) => setSelectedSeats(seatIds)}
/>

// Seat legend
<SeatLegend />

// Selected seats summary
<SeatSummary seats={selectedSeats} />

// Continue button
<Button onClick={proceedToCheckout}>
  Continue to Checkout
</Button>
```

#### Step 3: Checkout (`/booking/checkout/[holdId]`)
```tsx
// Booking summary
<BookingSummary />

// Promo code input
<PromoCodeInput />

// Payment method selection
<PaymentMethodSelector />

// Confirm and pay button
<Button onClick={confirmBooking}>
  Pay ${finalAmount}
</Button>
```

#### Step 4: Confirmation (`/booking/confirmation/[bookingId]`)
```tsx
// Success animation
<SuccessAnimation />

// Booking details
<BookingConfirmation booking={booking} />

// QR code for ticket
<QRCode value={bookingReference} />

// Download/Email options
<TicketActions />
```

### 4. User Account (`/account`)
**Tabs:**
- **My Bookings**: Past and upcoming bookings
- **Profile**: Edit personal information
- **Payment Methods**: Saved cards
- **Preferences**: Favorite cinemas, notification settings

### 5. Admin Dashboard (`/admin`)
**Sections:**
- **Dashboard**: Analytics and metrics
- **Movies**: Manage movie catalog
- **Schedule**: Create/edit showtimes
- **Bookings**: View all bookings
- **Reports**: Revenue and occupancy reports
- **Settings**: System configuration

## Key Components

### SeatMap Component
```tsx
interface SeatMapProps {
  showtimeId: string;
  onSeatsSelected: (seatIds: string[]) => void;
}

export const SeatMap: FC<SeatMapProps> = ({ showtimeId, onSeatsSelected }) => {
  const [seats, setSeats] = useState<Seat[]>([]);
  const [selectedSeats, setSelectedSeats] = useState<string[]>([]);

  useEffect(() => {
    // Fetch seat layout
    fetchSeats(showtimeId).then(setSeats);
    
    // WebSocket for real-time updates
    const socket = io(API_URL);
    socket.emit('watch:showtime', showtimeId);
    
    socket.on('seats:updated', (data) => {
      // Update seat availability
      refreshSeats(data.seatIds);
    });
    
    return () => socket.disconnect();
  }, [showtimeId]);

  const toggleSeat = (seatId: string) => {
    setSelectedSeats(prev => 
      prev.includes(seatId)
        ? prev.filter(id => id !== seatId)
        : [...prev, seatId]
    );
  };

  return (
    <div className="seat-map">
      <div className="screen-indicator">SCREEN</div>
      
      {seats.map(row => (
        <div key={row.label} className="seat-row">
          <span className="row-label">{row.label}</span>
          
          {row.seats.map(seat => (
            <button
              key={seat.id}
              className={`seat seat-${seat.status.toLowerCase()}`}
              disabled={seat.status !== 'AVAILABLE'}
              onClick={() => toggleSeat(seat.id)}
            >
              {seat.seatNumber}
            </button>
          ))}
        </div>
      ))}
    </div>
  );
};
```

### MovieCard Component
```tsx
interface MovieCardProps {
  movie: Movie;
  onClick?: () => void;
}

export const MovieCard: FC<MovieCardProps> = ({ movie, onClick }) => {
  return (
    <motion.div
      whileHover={{ scale: 1.05 }}
      whileTap={{ scale: 0.98 }}
      className="movie-card group cursor-pointer"
      onClick={onClick}
    >
      <div className="relative aspect-[2/3] overflow-hidden rounded-lg">
        <Image
          src={movie.posterUrl}
          alt={movie.title}
          fill
          className="object-cover transition-transform group-hover:scale-110"
        />
        
        {/* Overlay on hover */}
        <div className="absolute inset-0 bg-gradient-to-t from-black/80 to-transparent opacity-0 group-hover:opacity-100 transition-opacity">
          <div className="absolute bottom-0 p-4">
            <h3 className="text-white font-bold">{movie.title}</h3>
            <p className="text-white/80 text-sm">{movie.genres.join(', ')}</p>
            <div className="flex items-center gap-2 mt-2">
              <span className="text-accent-gold">★ {movie.rating}</span>
              <span className="text-white/60">{movie.duration} min</span>
            </div>
          </div>
        </div>
      </div>
    </motion.div>
  );
};
```

## State Management

### Zustand Store
```typescript
// store/bookingStore.ts
import { create } from 'zustand';

interface BookingState {
  selectedShowtime: Showtime | null;
  selectedSeats: Seat[];
  holdId: string | null;
  pricing: PriceBreakdown | null;
  
  setShowtime: (showtime: Showtime) => void;
  setSeats: (seats: Seat[]) => void;
  setHoldId: (id: string) => void;
  setPricing: (pricing: PriceBreakdown) => void;
  clearBooking: () => void;
}

export const useBookingStore = create<BookingState>((set) => ({
  selectedShowtime: null,
  selectedSeats: [],
  holdId: null,
  pricing: null,
  
  setShowtime: (showtime) => set({ selectedShowtime: showtime }),
  setSeats: (seats) => set({ selectedSeats: seats }),
  setHoldId: (id) => set({ holdId: id }),
  setPricing: (pricing) => set({ pricing }),
  clearBooking: () => set({
    selectedShowtime: null,
    selectedSeats: [],
    holdId: null,
    pricing: null
  })
}));
```

## API Integration

### API Client
```typescript
// lib/api/client.ts
import axios from 'axios';

const apiClient = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
});

// Request interceptor - Add auth token
apiClient.interceptors.request.use((config) => {
  const token = localStorage.getItem('accessToken');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Response interceptor - Handle token refresh
apiClient.interceptors.response.use(
  (response) => response,
  async (error) => {
    if (error.response?.status === 401) {
      // Try to refresh token
      const refreshed = await refreshAccessToken();
      if (refreshed) {
        return apiClient.request(error.config);
      }
      // Redirect to login
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

export default apiClient;
```

## Responsive Design

### Breakpoints
- Mobile: 0-640px
- Tablet: 641-1024px
- Desktop: 1025px+

### Mobile-First Approach
```tsx
// Example responsive grid
<div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-4">
  {movies.map(movie => (
    <MovieCard key={movie.id} movie={movie} />
  ))}
</div>
```

## Performance Optimizations

1. **Image Optimization**: Use Next.js `<Image>` component
2. **Code Splitting**: Dynamic imports for heavy components
3. **Lazy Loading**: Load images and components on demand
4. **Memoization**: Use `useMemo` and `useCallback` for expensive operations
5. **Virtual Scrolling**: For long lists (movie catalog)

## Accessibility

- Semantic HTML
- ARIA labels for interactive elements
- Keyboard navigation support
- Screen reader friendly
- Color contrast compliance (WCAG AA)

## Animations

```tsx
// Framer Motion examples

// Page transitions
<motion.div
  initial={{ opacity: 0, y: 20 }}
  animate={{ opacity: 1, y: 0 }}
  exit={{ opacity: 0, y: -20 }}
  transition={{ duration: 0.3 }}
>
  {children}
</motion.div>

// Loading states
<motion.div
  animate={{ rotate: 360 }}
  transition={{ repeat: Infinity, duration: 1, ease: "linear" }}
>
  <LoadingIcon />
</motion.div>
```

## Environment Variables
```bash
# .env.local
NEXT_PUBLIC_API_URL=http://localhost:5000/v1
NEXT_PUBLIC_STRIPE_PUBLIC_KEY=pk_test_...
NEXT_PUBLIC_GOOGLE_MAPS_API_KEY=AIza...
```