# 🎬 CinemaOS - Cinema Management Platform

A modern, full-stack cinema management system with real-time seat booking, dynamic pricing, and automated scheduling.

## ✨ Features

- 🎟️ **Real-time Seat Booking** - Live seat availability with WebSocket updates
- 💰 **Dynamic Pricing** - Time-based, demand-based, and seat-type pricing
- 📅 **Auto Scheduler** - Intelligent movie scheduling across multiple screens
- 💳 **Payment Processing** - Stripe integration for secure payments
- 👥 **Multi-role Support** - Customer, Staff, Manager, and Admin roles
- 📱 **Responsive Design** - Beautiful UI works on all devices
- 🔐 **Secure Authentication** - JWT-based auth with refresh tokens

## 🛠️ Technology Stack

### Backend
- **Runtime**: Node.js 20 with TypeScript
- **API**: tRPC (type-safe RPC)
- **Database**: PostgreSQL 15 + Prisma ORM
- **Cache**: Redis 7
- **Real-time**: Socket.IO
- **Payment**: Stripe

### Frontend
- **Framework**: Next.js 14 (App Router)
- **Language**: TypeScript
- **Styling**: Tailwind CSS
- **State**: Zustand
- **API Client**: tRPC + React Query
- **Animations**: Framer Motion

### DevOps
- **Containers**: Docker & Docker Compose
- **CI/CD**: GitHub Actions
- **Web Server**: Nginx
- **Monitoring**: Winston + PM2

## 🚀 Quick Start

### Prerequisites
- Docker & Docker Compose
- Node.js 20+ (for local development)
- Git

### Installation

1. **Clone the repository**
```bash
git clone <repository-url>
cd cinema-platform
```

2. **Set up environment variables**
```bash
cp .env.example .env
# Edit .env with your configuration
```

3. **Start services with Docker**
```bash
docker-compose up -d
```

4. **Run database migrations**
```bash
docker-compose exec backend npx prisma migrate deploy
```

5. **Seed the database**
```bash
docker-compose exec backend npm run seed
```

6. **Access the application**
- Frontend: http://localhost:3000
- Backend API: http://localhost:5000
- Admin Panel: http://localhost:3000/admin

### Default Credentials
```
Email: admin@cinemaos.com
Password: Admin123!
```

## 📁 Project Structure

```
cinema-platform/
├── backend/                # tRPC API server
│   ├── prisma/            # Database schema & migrations
│   ├── src/
│   │   ├── routers/       # tRPC routers
│   │   ├── services/      # Business logic
│   │   ├── middleware/    # Auth & validation
│   │   └── utils/         # Helper functions
│   └── tests/             # Backend tests
├── frontend/              # Next.js application
│   ├── src/
│   │   ├── app/          # Next.js pages (App Router)
│   │   ├── components/   # React components
│   │   ├── lib/          # Utilities & tRPC client
│   │   └── store/        # Zustand stores
│   └── e2e/              # Playwright tests
├── nginx/                # Nginx configuration
└── docker-compose.yml    # Docker services
```

## 🧪 Testing

### Run All Tests
```bash
# Backend unit & integration tests
cd backend && npm test

# Frontend component tests
cd frontend && npm test

# E2E tests
cd frontend && npm run test:e2e
```

### Test Coverage
```bash
# Backend coverage
cd backend && npm run test:coverage

# Frontend coverage
cd frontend && npm run test:coverage
```

## 📖 Documentation

Detailed documentation available in `/cinemaos_prompts/`:
- [System Overview](cinemaos_prompts/01_system_overview.md)
- [Architecture](cinemaos_prompts/02_architecture.md)
- [Auto Scheduler](cinemaos_prompts/03_scheduler_prompt.md)
- [Database Schema](cinemaos_prompts/04_db_schema_prompt.md)
- [Booking Engine](cinemaos_prompts/05_booking_engine_prompt.md)
- [Pricing Engine](cinemaos_prompts/06_pricing_engine_prompt.md)
- [DevOps Guide](cinemaos_prompts/07_devops_prompt.md)
- [API Reference](cinemaos_prompts/08_api_spec_prompt.md)
- [Frontend Guide](cinemaos_prompts/09_frontend_prompt.md)
- [Testing Strategy](cinemaos_prompts/10_testing_prompt.md)

## 🔧 Development

### Backend Development
```bash
cd backend
npm install
npm run dev
```

### Frontend Development
```bash
cd frontend
npm install
npm run dev
```

### Database Commands
```bash
# Create new migration
npx prisma migrate dev --name migration_name

# Reset database
npx prisma migrate reset

# Open Prisma Studio
npx prisma studio
```

## 🌐 Deployment

### Production Build
```bash
docker-compose -f docker-compose.prod.yml up -d --build
```

### Environment Variables
Ensure all production environment variables are set in `.env` before deployment.

## 📝 License

MIT License - see LICENSE file for details

## 🤝 Contributing

Contributions welcome! Please read our contributing guidelines first.

## 📧 Support

For support, email support@cinemaos.com or open an issue on GitHub.
