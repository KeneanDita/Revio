# Revio

A production-ready pull request management and analytics platform built with Go and Next.js.

## Overview

Revio integrates with GitHub to provide teams with a unified view of pull request activity, review performance, and engineering workflow insights.

**Key capabilities:**

- GitHub OAuth authentication and repository connection
- Pull request listing, filtering, and detail view
- Review tracking and comment management
- Analytics dashboard with merge time, review latency, and contributor metrics
- Time-range filtering and trend charts
- Light and dark theme support
- Webhook-based real-time updates

## Tech Stack

| Layer      | Technology                              |
|------------|-----------------------------------------|
| Frontend   | Next.js 14 (App Router), TypeScript     |
| Styling    | Tailwind CSS, shadcn/ui                 |
| State      | TanStack React Query                    |
| Charts     | Recharts                                |
| Backend    | Go 1.21+, Gin                           |
| Database   | PostgreSQL 16                           |
| ORM        | GORM                                    |
| Auth       | JWT + GitHub OAuth2                     |
| Container  | Docker, Docker Compose                  |

## Prerequisites

- [Docker](https://www.docker.com/) 24+ and Docker Compose v2
- [Go](https://go.dev/) 1.21+ (for local backend development)
- [Node.js](https://nodejs.org/) 20+ (for local frontend development)
- A GitHub OAuth App ([create one here](https://github.com/settings/developers))

## Quick Start (Docker)

### 1. Clone the repository

```bash
git clone https://github.com/keneandita/revio.git
cd revio
```

### 2. Configure environment

```bash
cp .env.example .env
```

Edit `.env` and fill in:

- `GITHUB_CLIENT_ID` — from your GitHub OAuth App
- `GITHUB_CLIENT_SECRET` — from your GitHub OAuth App
- `JWT_SECRET` — a random string of at least 32 characters

### 3. Start all services

```bash
docker compose up --build
```

Services will be available at:

- Frontend: http://localhost:3000
- Backend API: http://localhost:8080
- PostgreSQL: localhost:5432

## Local Development (without Docker)

### Backend

```bash
cd backend
cp ../.env.example .env
go mod download
go run cmd/server/main.go
```

### Frontend

```bash
cd frontend
cp .env.local.example .env.local
npm install
npm run dev
```

### Database

Run PostgreSQL locally and apply the migration:

```bash
psql -U revio -d revio -f backend/migrations/001_initial.sql
```

## Project Structure

```
revio/
├── backend/                  # Go API server
│   ├── cmd/server/           # Application entry point
│   ├── internal/
│   │   ├── auth/             # JWT, OAuth, password hashing
│   │   ├── config/           # Environment configuration
│   │   ├── database/         # GORM connection setup
│   │   ├── handlers/         # HTTP request handlers
│   │   ├── middleware/       # Auth, CORS, rate limiting
│   │   ├── models/           # GORM data models
│   │   ├── routes/           # Route registration
│   │   └── services/         # GitHub API client, sync worker
│   └── migrations/           # PostgreSQL schema files
├── frontend/                 # Next.js application
│   ├── app/                  # App Router pages and layouts
│   ├── components/           # Reusable UI components
│   ├── hooks/                # Custom React hooks
│   ├── lib/                  # API client, auth helpers, utilities
│   ├── providers/            # React Query and Theme providers
│   └── types/                # TypeScript type definitions
├── docker-compose.yml
├── .env.example
└── README.md
```

## API Reference

### Authentication

| Method | Path                        | Description                  |
|--------|-----------------------------|------------------------------|
| POST   | /api/auth/signup            | Register with email/password |
| POST   | /api/auth/login             | Login with email/password    |
| GET    | /api/auth/github            | Initiate GitHub OAuth flow   |
| GET    | /api/auth/github/callback   | GitHub OAuth callback        |
| POST   | /api/auth/logout            | Invalidate session           |
| GET    | /api/auth/me                | Get current user             |

### Repositories

| Method | Path                  | Description                   |
|--------|-----------------------|-------------------------------|
| GET    | /api/repos            | List connected repositories   |
| POST   | /api/repos/connect    | Connect a GitHub repository   |
| DELETE | /api/repos/:id        | Disconnect a repository       |
| POST   | /api/repos/:id/sync   | Trigger manual PR sync        |

### Pull Requests

| Method | Path              | Description                    |
|--------|-------------------|--------------------------------|
| GET    | /api/prs          | List PRs (filterable)          |
| GET    | /api/prs/:id      | Get PR details                 |
| POST   | /api/prs/:id/comment | Post a comment to a PR      |

### Analytics

| Method | Path            | Description                         |
|--------|-----------------|-------------------------------------|
| GET    | /api/analytics  | Aggregated metrics (filterable)     |

### Webhooks

| Method | Path                    | Description                    |
|--------|-------------------------|--------------------------------|
| POST   | /api/webhooks/github    | GitHub webhook receiver        |

## GitHub OAuth App Setup

1. Go to GitHub Settings > Developer settings > OAuth Apps > New OAuth App
2. Set **Homepage URL** to `http://localhost:3000`
3. Set **Authorization callback URL** to `http://localhost:8080/api/auth/github/callback`
4. Copy the Client ID and generate a Client Secret
5. Add both to your `.env` file

## Environment Variables

| Variable                | Required | Default                       | Description                       |
|-------------------------|----------|-------------------------------|-----------------------------------|
| `DB_USER`               | Yes      | revio                         | PostgreSQL user                   |
| `DB_PASSWORD`           | Yes      | revio_secret                  | PostgreSQL password               |
| `DB_NAME`               | Yes      | revio                         | PostgreSQL database name          |
| `DB_HOST`               | Yes      | localhost                     | PostgreSQL host                   |
| `DB_PORT`               | No       | 5432                          | PostgreSQL port                   |
| `JWT_SECRET`            | Yes      | —                             | Secret for JWT signing (min 32ch) |
| `GITHUB_CLIENT_ID`      | Yes      | —                             | GitHub OAuth App client ID        |
| `GITHUB_CLIENT_SECRET`  | Yes      | —                             | GitHub OAuth App secret           |
| `GITHUB_REDIRECT_URL`   | No       | http://localhost:8080/...     | OAuth callback URL                |
| `FRONTEND_URL`          | No       | http://localhost:3000         | Frontend origin for CORS          |
| `PORT`                  | No       | 8080                          | Backend HTTP port                 |
| `ENV`                   | No       | development                   | Runtime environment               |
| `NEXT_PUBLIC_API_URL`   | Yes      | http://localhost:8080         | Backend URL (used by frontend)    |

---

## License

MIT
