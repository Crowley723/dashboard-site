# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Architecture

This is a full-stack homelab dashboard application consisting of:

### Backend (Go)
- **Location**: Root directory and `internal/` package structure
- **Framework**: HTTP server using Go's standard library with chi router
- **Architecture**: Clean architecture with separated concerns:
  - `internal/server/` - HTTP server setup and routing
  - `internal/handlers/` - HTTP request handlers
  - `internal/auth/` - OIDC authentication and session management
  - `internal/data/` - Prometheus/Mimir data fetching and caching
  - `internal/config/` - Configuration management
  - `internal/models/` - Data models
  - `internal/middlewares/` - HTTP middlewares and app context

### Frontend (React + TypeScript)
- **Location**: `web/` directory
- **Framework**: React 19 with Vite, TanStack Router, and React Query
- **Key Technologies**:
  - TanStack Router for file-based routing
  - React Query for server state management
  - Tailwind CSS v4 for styling
  - Radix UI for components
  - Recharts for data visualization

## Development Commands

### Full Stack Development
```bash
make dev                    # Start both frontend and backend servers
make dev-debug             # Start with Go debugger on port 2345
make install               # Install all dependencies
```

### Backend Only
```bash
make dev-backend           # Start Go server with hot reload (reflex)
make dev-backend-debug     # Start with debugger
go test ./...              # Run tests
make coverage              # Run tests with coverage
make coverage-html         # Generate HTML coverage report
make mocks                 # Generate mocks
```

### Frontend Only
```bash
cd web && pnpm run dev     # Start Vite dev server
cd web && pnpm run build   # Build for production
cd web && pnpm run lint    # Run ESLint
cd web && pnpm run format  # Format with Prettier
cd web && pnpm run format:check # Check formatting
```

### Docker
```bash
make build                 # Build Docker image
```

## Configuration

- **Backend Config**: `config.yaml` - Contains OIDC, Prometheus, caching, and server settings
- **Docker Config**: `config.docker.yaml` - Docker-specific configuration (not tracked in git)
  - Use `config.docker.yaml.template` as a starting point
  - Copy template: `cp config.docker.yaml.template config.docker.yaml`
  - Update with your specific values (OIDC credentials, Prometheus URL, etc.)
- **Frontend Config**: Uses environment variables and TanStack Router configuration

## Key Features

- **Authentication**: OIDC integration with session management
- **Data Sources**: Prometheus/Mimir metrics with configurable queries and TTL-based caching
- **Dashboard Components**: Pre-built cards for node status, pod uptime, Traefik metrics, and pod restarts
- **Real-time Updates**: Background data fetching with caching layer
- **Responsive Design**: Mobile-friendly UI with Tailwind CSS

## Testing

- **Backend**: Go standard testing with mocks generated via `go generate`
- **Frontend**: ESLint and Prettier for code quality

## Route Structure

- Frontend uses TanStack Router with file-based routing in `web/src/routes/`
- Backend API endpoints handled in `internal/handlers/` with chi router

## Deployment

- Dockerized application with Helm charts in `helm/` directory
- CI/CD configuration in `.github/` directory