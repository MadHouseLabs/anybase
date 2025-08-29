# Docker Setup for AnyBase

AnyBase provides Docker images and docker-compose configurations for easy deployment.

## Available Docker Images

The project publishes two Docker images to GitHub Container Registry:

- `ghcr.io/madhouselabs/anybase-backend`: Go backend API server
- `ghcr.io/madhouselabs/anybase-dashboard`: Next.js dashboard UI

## Quick Start with Docker Compose

### Production Environment

1. Clone the repository:
```bash
git clone https://github.com/madhouselabs/anybase.git
cd anybase
```

2. Create a `.env` file with your configuration:
```bash
MONGO_PASSWORD=your-secure-password
JWT_SECRET=your-secure-jwt-secret
```

3. Start the services:
```bash
docker compose up -d
```

This will start:
- MongoDB on port 27017
- Backend API on port 8080
- Dashboard UI on port 3000

### Development Environment

For development with hot reload:

```bash
docker compose -f docker-compose.dev.yml up
```

This configuration:
- Mounts source code as volumes
- Enables hot reload for both backend (using Air) and frontend (Next.js dev server)
- Uses development configurations

## Building Images Locally

### Backend Image
```bash
docker build -t anybase-backend .
```

### Dashboard Image
```bash
docker build -t anybase-dashboard ./dashboard
```

## Environment Variables

### Backend Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_CONNECTION_STRING` | MongoDB connection string | `mongodb://localhost:27017/anybase` |
| `DB_NAME` | Database name | `anybase` |
| `SERVER_HOST` | Server host | `0.0.0.0` |
| `SERVER_PORT` | Server port | `8080` |
| `SERVER_MODE` | Server mode (development/production) | `production` |
| `JWT_SECRET` | JWT signing secret | Required |
| `JWT_EXPIRY` | JWT token expiry | `24h` |
| `REFRESH_TOKEN_EXPIRY` | Refresh token expiry | `168h` |
| `CORS_ALLOWED_ORIGINS` | CORS allowed origins | `*` |

### Dashboard Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `NEXT_PUBLIC_API_URL` | Backend API URL | `http://localhost:8080/api/v1` |
| `NODE_ENV` | Node environment | `production` |

## Docker Compose Profiles

### Default Profile
Includes MongoDB, backend, and dashboard services.

### Production Profile
Adds Nginx reverse proxy:
```bash
docker compose --profile production up -d
```

## Health Checks

All services include health checks:
- **MongoDB**: Ping command
- **Backend**: HTTP GET /health
- **Dashboard**: HTTP GET /api/health

## Volumes

### Production
- `mongodb_data`: MongoDB data persistence

### Development
- `mongodb_dev_data`: MongoDB data for development
- Source code mounted for hot reload

## Networking

All services communicate through the `anybase-network` bridge network.

## Multi-Architecture Support

Images support both `linux/amd64` and `linux/arm64` architectures.

## CI/CD with GitHub Actions

The repository includes GitHub Actions workflow that:
1. Builds Docker images on push to main branch
2. Publishes images to GitHub Container Registry
3. Tags images with version numbers from git tags
4. Runs integration tests on pull requests

## Pulling Images from Registry

```bash
# Pull backend image
docker pull ghcr.io/madhouselabs/anybase-backend:latest

# Pull dashboard image
docker pull ghcr.io/madhouselabs/anybase-dashboard:latest
```

## Security Considerations

1. Change default passwords in production
2. Use strong JWT secrets
3. Configure CORS appropriately
4. Use HTTPS in production (via Nginx or load balancer)
5. Run containers as non-root users (already configured)

## Troubleshooting

### Containers not starting
Check logs:
```bash
docker compose logs backend
docker compose logs dashboard
docker compose logs mongodb
```

### Permission issues
Ensure proper file permissions:
```bash
chmod -R 755 .
```

### Port conflicts
Change ports in docker-compose.yml if default ports are in use:
```yaml
ports:
  - "3001:3000"  # Change host port to 3001
```

### MongoDB connection issues
Verify MongoDB is healthy:
```bash
docker compose exec mongodb mongosh --eval "db.adminCommand('ping')"
```