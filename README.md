# AnyBase

<div align="center">
  <h3>Firebase-like API Layer for AWS DocumentDB</h3>
  <p>A high-performance, scalable backend service with real-time capabilities, MCP support, and enterprise-grade security</p>
  
  [![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://go.dev)
  [![Next.js](https://img.shields.io/badge/Next.js-15.5-black?style=flat&logo=next.js)](https://nextjs.org)
  [![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
  [![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat&logo=docker)](https://www.docker.com)
</div>

## ✨ Features

### Core Capabilities
- **🚀 RESTful API** - Fast, scalable REST API with JWT authentication
- **📊 Dynamic Collections** - Schema-flexible document storage with validation
- **🔍 Advanced Querying** - MongoDB-compatible query language with aggregation support
- **🎯 Views & Aggregations** - Pre-defined queries with parameterized execution
- **🔐 Row-Level Security** - Fine-grained access control with RBAC
- **🔑 Access Keys** - API key authentication with scoped permissions
- **🤖 MCP Integration** - Model Context Protocol support for AI applications

### Dashboard Features
- **📈 Real-time Analytics** - Monitor system health and usage metrics
- **👥 User Management** - Complete user administration interface
- **🗂️ Collection Explorer** - Visual database management tools
- **🔧 Settings Management** - System and user configuration
- **📝 Schema Editor** - Visual JSON schema builder

## 🚀 Quick Start

### Prerequisites
- Go 1.22 or higher
- Node.js 20+ and pnpm
- Database: MongoDB, AWS DocumentDB, or PostgreSQL
- Docker (optional)

### Installation

#### Using Docker Compose (Recommended)

```bash
# Clone the repository
git clone https://github.com/MadHouseLabs/anybase.git
cd anybase

# Start with Docker Compose
docker-compose up -d
```

#### Manual Installation

```bash
# Clone the repository
git clone https://github.com/MadHouseLabs/anybase.git
cd anybase

# Install backend dependencies
go mod download

# Install dashboard dependencies
cd dashboard
pnpm install

# Set up environment variables
cp .env.example .env.local
# Edit .env.local with your configuration

# Run the backend
go run cmd/server/main.go

# In another terminal, run the dashboard
cd dashboard
pnpm dev
```

## 📁 Project Structure

```
anybase/
├── api/                    # API handlers and routes
│   └── v1/                # Version 1 API implementation
├── cmd/                   # Application entrypoints
│   └── server/           # Main server application
├── internal/             # Private application code
│   ├── accesskey/       # Access key management
│   ├── auth/           # Authentication logic
│   ├── collection/     # Collection operations
│   ├── config/        # Configuration management
│   ├── database/      # Database connectivity
│   ├── governance/    # RBAC and permissions
│   ├── middleware/    # HTTP middleware
│   ├── models/        # Data models
│   ├── settings/      # Settings management
│   └── user/         # User management
├── dashboard/          # Next.js dashboard application
│   ├── app/          # App router pages
│   ├── components/   # React components
│   └── lib/         # Utility functions
└── docker/          # Docker configurations
```

## 🔧 Configuration

### Database Configuration

AnyBase supports multiple database backends. Choose one:

#### MongoDB Configuration

```env
# MongoDB/DocumentDB Connection
ANYBASE_DATABASE_TYPE=mongodb
ANYBASE_DATABASE_URI=mongodb://localhost:27017
ANYBASE_DATABASE_DATABASE=anybase
```

#### PostgreSQL Configuration

```env
# PostgreSQL Connection
ANYBASE_DATABASE_TYPE=postgres
ANYBASE_DATABASE_URI=postgres://username@localhost/anybase?sslmode=disable
ANYBASE_DATABASE_DATABASE=anybase
```

### Environment Variables

Create a `.env` file in the root directory:

```env
# Database Configuration (see above for database-specific settings)
ANYBASE_DATABASE_TYPE=mongodb  # or postgres
ANYBASE_DATABASE_URI=mongodb://localhost:27017
ANYBASE_DATABASE_DATABASE=anybase

# JWT Configuration
JWT_SECRET=your-secret-key-here
JWT_EXPIRY=24h
REFRESH_TOKEN_EXPIRY=168h

# Server Configuration
PORT=8080
HOST=0.0.0.0
MODE=development

# Initial Admin User (created on first startup)
INIT_ADMIN_EMAIL=admin@anybase.local
INIT_ADMIN_PASSWORD=admin123  # Change this immediately after first login!

# Dashboard Configuration
NEXT_PUBLIC_API_URL=http://localhost:8080
```

### Database Indexes

AnyBase automatically creates necessary indexes on startup:
- **MongoDB**: Native MongoDB indexes with compound and unique constraints
- **PostgreSQL**: B-tree and GIN indexes on JSONB fields for optimal query performance

## 📖 API Documentation

Complete API documentation is available in the [OpenAPI specification](./openapi.yaml). You can:
- Import it into Postman or Insomnia for testing
- Use it to generate client SDKs
- View it with Swagger UI or ReDoc

## 🔌 API Usage

### Authentication

```bash
# Register a new user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"securepass","name":"John Doe"}'

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"securepass"}'
```

### Collections

```bash
# Create a collection
curl -X POST http://localhost:8080/api/v1/collections \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "products",
    "description": "Product catalog",
    "schema": {
      "type": "object",
      "properties": {
        "name": {"type": "string"},
        "price": {"type": "number"},
        "category": {"type": "string"}
      },
      "required": ["name", "price"]
    }
  }'

# Insert a document
curl -X POST http://localhost:8080/api/v1/data/products \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Widget","price":29.99,"category":"Hardware"}'
```

### Access Keys

```bash
# Create an access key
curl -X POST http://localhost:8080/api/v1/access-keys \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Production API Key",
    "permissions": ["collection:products:read", "collection:products:write"]
  }'
```

## 🤖 MCP (Model Context Protocol) Integration

AnyBase supports MCP for AI model integration. Configure your MCP client to connect to:

```
http://localhost:8080/mcp
```

The MCP server provides collection-specific tools with full schema awareness for optimal AI interactions.

## 🐳 Docker Deployment

### Using Pre-built Images

```bash
# Pull the images
docker pull ghcr.io/madhouselabs/anybase-backend:latest
docker pull ghcr.io/madhouselabs/anybase-dashboard:latest

# Run with docker-compose
docker-compose up -d
```

### Building from Source

```bash
# Build backend
docker build -t anybase-backend .

# Build dashboard
docker build -t anybase-dashboard ./dashboard
```

## 🧪 Testing

```bash
# Run backend tests
go test ./...

# Run backend tests with coverage
go test -cover ./...

# Run dashboard tests
cd dashboard
pnpm test
```

## 📊 Performance

AnyBase is optimized for high performance:

- **Connection Pooling**: Efficient MongoDB connection management
- **Indexed Queries**: Automatic index creation for optimal query performance
- **Caching**: Built-in caching for frequently accessed data
- **Rate Limiting**: Per-IP rate limiting to prevent abuse
- **Concurrent Processing**: Go's goroutines for parallel request handling

## 🔒 Security

- **JWT Authentication**: Secure token-based authentication
- **RBAC**: Role-based access control with fine-grained permissions
- **Input Validation**: JSON Schema validation for all inputs
- **Rate Limiting**: Built-in rate limiting to prevent abuse
- **CORS**: Configurable CORS policies
- **Secure Headers**: Security headers for XSS and CSRF protection

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built with [Gin](https://github.com/gin-gonic/gin) web framework
- Dashboard powered by [Next.js](https://nextjs.org)
- Database support via [MongoDB Go Driver](https://github.com/mongodb/mongo-go-driver)
- UI components from [Radix UI](https://www.radix-ui.com)

## 📞 Support

- 🐛 Issues: [GitHub Issues](https://github.com/MadHouseLabs/anybase/issues)
- 💬 Discussions: [GitHub Discussions](https://github.com/MadHouseLabs/anybase/discussions)

---

<div align="center">
  Made with ❤️ by <a href="https://github.com/MadHouseLabs">MadHouse Labs</a>
</div>