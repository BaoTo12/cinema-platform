# Auto Scheduler Prompt

## Overview
The Auto Scheduler is an intelligent system that automatically generates optimal movie showtimes across multiple screens within a cinema, maximizing screen utilization while ensuring operational constraints are met.

## Core Requirements

### 1. Scheduling Constraints

#### Time Constraints
- **Operating Hours**: Cinema opens at 9:00 AM, closes at 1:00 AM
- **First Show**: No earlier than 9:30 AM
- **Last Show**: Must end before 1:00 AM
- **Buffer Time**: 30 minutes between shows (cleaning + ads)
- **Minimum Break**: 15 minutes for screen maintenance every 6 shows

#### Movie Constraints
- **Runtime**: Includes movie duration + 15 minutes for previews/ads
- **Format Requirements**: IMAX/3D movies require compatible screens
- **Concurrent Limits**: Same movie can't play on multiple screens at same time

#### Screen Constraints
- **Screen Capacity**: Each screen has different seating capacity
- **Format Support**: Screens support specific formats (Standard, 3D, IMAX, 4DX)
- **Maintenance Windows**: Scheduled maintenance blocks

### 2. Optimization Goals

#### Primary Goals
1. **Maximize Revenue**: Prioritize popular movies during peak hours
2. **Screen Utilization**: Aim for >75% screen time usage
3. **Audience Coverage**: Distribute showtimes to serve different audience segments

#### Secondary Goals
1. **Even Distribution**: Balance showtimes across the day
2. **Popular Time Slots**: More shows during 6 PM - 11 PM
3. **Variety**: Ensure diverse movie options at any given time

## Scheduling Algorithm

### Input Parameters
```typescript
interface ScheduleInput {
  date: Date;
  cinemaId: string;
  movies: Movie[];           // Movies to schedule
  screens: Screen[];         // Available screens
  preferences: {
    peakHours: TimeRange[];  // e.g., 18:00-23:00
    popularMovies: string[]; // Movie IDs to prioritize
    minShowsPerMovie: number; // Minimum shows per movie
    maxShowsPerMovie: number; // Maximum shows per movie
  };
}

interface Movie {
  id: string;
  title: string;
  duration: number;     // In minutes
  format: 'STANDARD' | '3D' | 'IMAX' | '4DX';
  popularity: number;   // 1-10 rating
  releaseDate: Date;
}

interface Screen {
  id: string;
  name: string;
  capacity: number;
  supportedFormats: string[];
  maintenanceWindows: TimeRange[];
}
```

### Algorithm Steps

#### Phase 1: Initialization
1. Calculate total available minutes per screen
2. Estimate total shows possible across all screens
3. Allocate show counts to movies based on popularity

#### Phase 2: Time Slot Generation
1. Generate all possible start times (30-minute intervals)
2. Filter valid slots based on operating hours
3. Tag slots as peak/off-peak

#### Phase 3: Assignment
```
FOR each movie (sorted by popularity DESC):
  allocated_shows = 0
  
  FOR each time slot (prioritize peak hours):
    FOR each compatible screen:
      IF no conflicts AND within constraints:
        Assign show to screen at time slot
        allocated_shows++
        
      IF allocated_shows >= min_required:
        BREAK
```

#### Phase 4: Gap Filling
1. Identify gaps in schedule (>3 hours with no show)
2. Attempt to insert additional shows
3. Use less popular movies to fill gaps

#### Phase 5: Validation
1. Verify all hard constraints met
2. Calculate utilization metrics
3. Generate conflict report if any

### Example Schedule Output

**Screen 1 (300 seats, Standard/3D)**
- 10:00 AM - 12:15 PM: Movie A (2hr 15min)
- 12:45 PM - 02:30 PM: Movie B (1hr 45min)
- 03:00 PM - 05:15 PM: Movie A (2hr 15min)
- 05:45 PM - 08:00 PM: Movie C (2hr 15min) [PEAK]
- 08:30 PM - 10:15 PM: Movie D (1hr 45min) [PEAK]
- 10:45 PM - 12:30 AM: Movie B (1hr 45min)

**Utilization**: 14.5 hours / 16 hours = 90.6%

## API Specification

### Generate Schedule
```http
POST /api/schedule/generate
Authorization: Bearer {token}
Content-Type: application/json

{
  "date": "2025-12-01",
  "cinemaId": "cinema-123",
  "movieIds": ["movie-1", "movie-2", "movie-3"],
  "preferences": {
    "peakHours": [
      {"start": "18:00", "end": "23:00"}
    ],
    "minShowsPerMovie": 2,
    "maxShowsPerMovie": 6
  }
}
```

**Response:**
```json
{
  "success": true,
  "schedule": {
    "date": "2025-12-01",
    "cinemaId": "cinema-123",
    "screens": [
      {
        "screenId": "screen-1",
        "showtimes": [
          {
            "id": "show-1",
            "movieId": "movie-1",
            "startTime": "10:00",
            "endTime": "12:15",
            "isPeak": false
          }
        ]
      }
    ],
    "metrics": {
      "totalShows": 24,
      "avgUtilization": 87.5,
      "peakShows": 12
    },
    "conflicts": []
  }
}
```

## Manual Override

### Create Manual Showtime
```http
POST /api/schedule/showtime
Authorization: Bearer {token}

{
  "screenId": "screen-1",
  "movieId": "movie-1",
  "date": "2025-12-01",
  "startTime": "20:00"
}
```

**Validation Checks:**
1. Screen is available at requested time
2. Movie format compatible with screen
3. No overlapping shows
4. Within operating hours
5. Sufficient buffer time from adjacent shows

## Conflict Resolution

### Conflict Types
1. **Time Overlap**: Two shows on same screen overlap
2. **Format Mismatch**: Movie format incompatible with screen
3. **Maintenance Conflict**: Show during maintenance window
4. **Operating Hours**: Show outside cinema operating hours

### Resolution Strategies
1. **Auto-adjust**: Shift show time by ±30 minutes
2. **Screen Swap**: Move to different compatible screen
3. **Manual Review**: Flag for manager approval
4. **Cancel**: Remove conflicting show if low priority

## Reporting

### Schedule Report
```typescript
interface ScheduleReport {
  date: Date;
  totalShows: number;
  showsByMovie: Map<string, number>;
  utilizationByScreen: Map<string, number>;
  peakHourCoverage: number;
  estimatedRevenue: number;
  recommendations: string[];
}
```

### Recommendations Logic
- ✅ "Good utilization (>85%)" if avg utilization > 85%
- ⚠️ "Low utilization on Screen 2 (65%)" if any screen < 70%
- 💡 "Add more shows for Movie X during peak hours" if popular movie has <3 peak shows
- 🔧 "Schedule maintenance for Screen 1" if no maintenance window in schedule

## Cron Jobs

### Daily Auto-Schedule
Run at 2:00 AM daily to generate schedules for 7 days ahead:
```bash
0 2 * * * node dist/jobs/generate-schedules.js
```

### Schedule Publication
Publish upcoming schedules to frontend cache:
```bash
*/15 * * * * node dist/jobs/publish-schedules.js
```

## Database Operations

### Store Generated Schedule
```sql
BEGIN TRANSACTION;

-- Insert showtimes
INSERT INTO showtimes (screen_id, movie_id, start_time, end_time, price_tier)
VALUES (...);

-- Update screen availability
UPDATE screen_availability 
SET is_available = false
WHERE screen_id = ? AND time_range OVERLAPS (?);

COMMIT;
```

### Conflict Detection Query
```sql
SELECT * FROM showtimes
WHERE screen_id = ?
  AND date = ?
  AND (
    (start_time, end_time) OVERLAPS (?, ?)
  );
```

## Frontend Integration

### Display Schedule
- Calendar view with all showtimes
- Filter by movie, screen, time range
- Color-coded by utilization (green: >80%, yellow: 60-80%, red: <60%)
- Visual timeline for each screen

### Manager Actions
- Approve/reject auto-generated schedule
- Add/edit/delete individual showtimes
- Trigger re-generation with different parameters
- Export schedule to PDF/Excel