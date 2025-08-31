# Anybase Codebase Memory

## Project Overview
Anybase is a flexible database management system that provides a unified API for working with different database backends (MongoDB and PostgreSQL). It offers collection management, document CRUD operations, user authentication, and access control.

## Key Files and Structure

### API Documentation
- **`openapi.yaml`** - Complete OpenAPI 3.0.3 specification documenting all REST API endpoints, request/response schemas, authentication methods, and error handling. Use this for generating client SDKs or API documentation.

### Backend (Go)
- **`cmd/server/main.go`** - Main server entry point, initializes database, sets up routes, handles graceful shutdown
- **`internal/config/`** - Configuration management
- **`internal/database/`** - Database abstraction layer
  - `adapters/postgres/` - PostgreSQL adapter implementation (fully functional)
  - `adapters/mongodb/` - MongoDB adapter implementation
- **`internal/collection/`** - Collection service logic
  - `service.go` - MongoDB-based collection service
  - `service_adapter.go` - Adapter-based service for PostgreSQL/other databases
- **`api/v1/`** - HTTP handlers for all API endpoints
- **`internal/auth/`** - Authentication and JWT token management
- **`internal/governance/`** - RBAC and permissions
- **`pkg/models/`** - Data models and structures

### Frontend (Next.js/React)
- **`dashboard/`** - Next.js dashboard application
  - `app/` - App router pages and layouts
  - `components/` - Reusable React components
  - `lib/` - Utilities and API client

### Database Schema

#### PostgreSQL Tables (auto-created on startup)
- `users` - User accounts with JSONB data field
- `sessions` - User sessions
- `access_keys` - API access keys
- `audit_logs` - Audit trail
- `collections` - User-created collections metadata
- `settings` - System and user settings
- `_collections` - Collection schema metadata
- `_views` - View definitions
- `data_*` - Dynamic tables for user collections (e.g., `data_products`, `data_inventory`)

All tables use:
- UUID primary keys (`_id`)
- JSONB for flexible data storage
- System fields: `_created_by` (TEXT), `_updated_by` (TEXT), `_created_at`, `_updated_at`, `_version`, `_deleted_at`
- Automatic triggers for updating timestamps and version

## Key Features

### Database Support
- **MongoDB** - Native support with full feature set
- **PostgreSQL** - Complete adapter implementation with:
  - JSONB storage for flexible schemas
  - B-tree indexes for unique constraints
  - GIN indexes for full-text search
  - Proper ObjectID generation for compatibility
  - Automatic table creation for collections

### Authentication & Authorization
- JWT-based authentication
- Refresh token support
- Role-based access control (admin, developer, viewer)
- API key authentication for programmatic access

### Collection Management
- Dynamic collection creation with schema validation
- Index management (unique, compound, sparse)
- Soft delete support
- Versioning
- Auditing capabilities

### API Endpoints
All endpoints documented in `openapi.yaml`:
- `/api/v1/auth/*` - Authentication
- `/api/v1/collections/*` - Collection management
- `/api/v1/data/*` - Document CRUD
- `/api/v1/views/*` - View management
- `/api/v1/users/*` - User management
- `/api/v1/access-keys/*` - API key management
- `/api/v1/settings/*` - Settings

## Environment Variables
- `ANYBASE_DATABASE_TYPE` - Database type (mongodb/postgres)
- `ANYBASE_DATABASE_URI` - Database connection string
- `ANYBASE_DATABASE_DATABASE` - Database name
- `INIT_ADMIN_EMAIL` - Initial admin email (default: admin@anybase.local)
- `INIT_ADMIN_PASSWORD` - Initial admin password (default: admin123)

## Recent Improvements
- Fixed PostgreSQL adapter for complete functionality
- Proper ID generation (ObjectID-compatible hex strings)
- Automatic database initialization on startup
- Fixed CRUD operations (INSERT, QUERY, UPDATE, DELETE)
- Index creation with proper field ordering
- Duplicate collection prevention
- System tables auto-creation

## Testing
The PostgreSQL adapter has been thoroughly tested with:
- Fresh database initialization
- All CRUD operations
- Index creation and constraints
- Unique constraint enforcement
- Collection management
- ID generation and consistency

## Known Issues
None currently - PostgreSQL adapter is fully functional.

## Development Commands
```bash
# Start server with PostgreSQL
ANYBASE_DATABASE_TYPE=postgres ANYBASE_DATABASE_URI='postgres://user@localhost/anybase?sslmode=disable' go run cmd/server/main.go

# Start server with MongoDB
go run cmd/server/main.go

# Start frontend dashboard
cd dashboard && pnpm dev

# Build server
go build -o server cmd/server/main.go
```