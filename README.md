# Homelab Dashboard

A personal dashboard application for displaying homelab infrastructure metrics and projects. This application serves as a centralized view of Kubernetes cluster status, service metrics, and other homelab-related information.

## Purpose

This dashboard is designed to be deployed at the root of a personal domain to showcase:

- **Kubernetes Metrics**: Real-time cluster health, node status, and pod information
- **Service Monitoring**: Traefik ingress metrics, pod uptime, and restart tracking
- **Infrastructure Overview**: Visual representation of homelab services and their current state

The application pulls metrics from Prometheus/Mimir and presents them through a clean, responsive web interface.

## Architecture

**Backend**: Go-based HTTP server with OIDC authentication
- Fetches and caches Prometheus metrics
- Handles user sessions and authentication
- Serves API endpoints for frontend consumption

**Frontend**: React application with TypeScript
- TanStack Router for navigation
- React Query for data management
- Recharts for metric visualization
- Tailwind CSS for styling

## key Features

- OIDC authentication for secure access to sensitive details
- Real-time metrics from Prometheus/Mimir
- Configurable dashboard cards and queries
- TTL-based caching for performance
- Mobile-responsive design
- Docker deployment ready

## Development

See `CLAUDE.md` for detailed development setup and commands.

## Deployment

The application is containerized and includes Helm charts for Kubernetes deployment.