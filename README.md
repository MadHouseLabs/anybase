# AnyBase - Firebase-like Layer for AWS DocumentDB

AnyBase provides a Firebase-like abstraction layer on top of AWS DocumentDB (or MongoDB) with comprehensive user management, authentication, and governance capabilities built in Go.

## Features

### Core Capabilities
- **User Management**: Complete user registration, authentication, and profile management
- **JWT Authentication**: Secure token-based authentication with refresh tokens
- **RBAC System**: Role-Based Access Control with fine-grained permissions
- **AWS DocumentDB Integration**: Optimized for DocumentDB with MongoDB driver compatibility
- **RESTful API**: Clean API design with versioning support
- **Security First**: Password hashing, rate limiting, CORS, and audit logging

### Security Features
- Bcrypt password hashing
- JWT token authentication with refresh tokens
- Account lockout after failed attempts
- Email verification flow
- Password reset functionality
- Rate limiting per IP address
- CORS configuration
- Audit logging for compliance

## Tech Stack

- **Backend**: Go 1.23+
- **Framework**: Gin Web Framework
- **Database**: AWS DocumentDB / MongoDB
- **Authentication**: JWT (golang-jwt)
- **Configuration**: Viper
- **Task Runner**: Taskgo

## Getting Started

### Prerequisites

- Go 1.23 or higher
- MongoDB (for local development) or AWS DocumentDB connection
- Task (taskgo) installed

### Installation

1. Clone the repository:
```bash
git clone https://github.com/karthik/anybase.git
cd anybase
```

2. Install dependencies:
```bash
task install
```

3. Configure the application:
```bash
cp config.yaml.example config.yaml
# Edit config.yaml with your settings
```

4. Run the application:
```bash
task dev  # For development with hot reload
# or
task run  # For normal execution
```

### Configuration

The application uses a YAML configuration file. Key settings include:

- **Server**: Port, host, timeouts
- **Database**: DocumentDB/MongoDB connection string
- **Auth**: JWT secrets, token expiration, password policies
- **AWS**: Region and credentials (for DocumentDB)
- **Logging**: Level and format

## API Documentation

### Authentication Endpoints

#### Register User
```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "securepassword",
  "first_name": "John",
  "last_name": "Doe"
}
```

#### Login
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "securepassword"
}
```

#### Refresh Token
```http
POST /api/v1/auth/refresh
Content-Type: application/json

{
  "refresh_token": "your_refresh_token"
}
```

### Protected Endpoints

All protected endpoints require the Authorization header:
```http
Authorization: Bearer <access_token>
```

#### Get User Profile
```http
GET /api/v1/users/profile
```

#### Change Password
```http
POST /api/v1/auth/change-password
Content-Type: application/json

{
  "old_password": "currentpassword",
  "new_password": "newpassword"
}
```

### Admin Endpoints

Admin endpoints require admin role:

```http
GET /api/v1/admin/users         # List all users
GET /api/v1/admin/roles         # List all roles
POST /api/v1/admin/roles        # Create new role
GET /api/v1/admin/permissions   # List all permissions
```

## Project Structure

```
anybase/
├── cmd/
│   └── server/          # Application entrypoint
├── internal/
│   ├── auth/           # Authentication logic
│   ├── user/           # User management
│   ├── governance/     # RBAC and permissions
│   ├── database/       # Database connection
│   ├── middleware/     # HTTP middleware
│   └── config/         # Configuration management
├── pkg/
│   ├── models/         # Data models
│   ├── utils/          # Utilities
│   └── errors/         # Custom errors
├── api/
│   └── v1/             # API handlers
├── scripts/            # Build and deployment
├── config.yaml         # Configuration file
└── Taskfile.yml        # Task definitions
```

## Development

### Available Tasks

```bash
task install      # Install dependencies
task dev         # Run with hot reload
task build       # Build binary
task test        # Run tests
task lint        # Run linter
task fmt         # Format code
```

### Testing

```bash
task test                # Run all tests
task test:coverage       # Generate coverage report
```

### Building

```bash
task build              # Build binary
task docker:build       # Build Docker image
```

## Deployment

### Docker

```bash
task docker:build
task docker:run
```

### AWS DocumentDB

1. Update `config.yaml` with your DocumentDB connection string:
```yaml
database:
  uri: "mongodb://username:password@docdb-cluster.amazonaws.com:27017/?tls=true&tlsCAFile=rds-ca-2019-root.pem"
  retry_writes: false  # DocumentDB doesn't support retryable writes
```

2. Ensure proper IAM roles and security groups are configured

## Security Considerations

1. **Change default JWT secret** in production
2. Use **IAM roles** instead of explicit AWS credentials
3. Configure **CORS** with specific origins in production
4. Enable **TLS/SSL** for database connections
5. Implement **rate limiting** based on your requirements
6. Set up **audit logging** for compliance
7. Use **environment variables** for sensitive configuration

## Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## License

MIT License - see LICENSE file for details

## Support

For issues and questions, please create an issue on GitHub.

## Roadmap

- [ ] WebSocket support for real-time updates
- [ ] GraphQL API support
- [ ] Multi-tenancy support
- [ ] Data encryption at rest
- [ ] Advanced query builder
- [ ] Caching layer with Redis
- [ ] Metrics and monitoring
- [ ] API documentation with Swagger