# GoTask Management 🚀

A comprehensive and scalable task management system built with Go, featuring a RESTful API and multiple storage backends including PostgreSQL, MySQL, MongoDB, SQLite, and JSON file storage.

## ✨ Features

- ✅ **Full CRUD Operations**: Create, read, update, and delete tasks
- ✅ **Task Status Management**: Mark tasks as completed or pending
- ✅ **Due Date Support**: Set and track due dates for tasks
- ✅ **Advanced Filtering**: Filter tasks by status, due dates, and more
- ✅ **Multiple Storage Backends**: PostgreSQL, MySQL, MongoDB, SQLite, JSON
- ✅ **RESTful API**: Clean JSON API with comprehensive endpoints
- ✅ **Docker Support**: Full containerization with Docker Compose
- ✅ **Database Admin Tools**: pgAdmin, phpMyAdmin, Mongo Express
- ✅ **Health Monitoring**: Built-in health check endpoints
- ✅ **Comprehensive Testing**: Unit tests for all components
- ✅ **Production Ready**: Logging, error handling, and monitoring
- ✅ **UTF-8 Support**: Full Unicode and emoji support
- ✅ **High Performance**: Optimized queries and connection pooling

## 🚀 Quick Start

### Using Docker Compose (Recommended)

1. **Clone the repository:**
```bash
git clone <repository-url>
cd GoTask_Management
```

2. **Start all services:**
```bash
docker-compose up -d
```

3. **Access the services:**
- **GoTask API**: http://localhost:8080
- **pgAdmin** (PostgreSQL): http://localhost:5050
- **phpMyAdmin** (MySQL): http://localhost:8081
- **Mongo Express** (MongoDB): http://localhost:8082

### Manual Installation

1. **Prerequisites:**
   - Go 1.21+ installed
   - Database of choice (PostgreSQL, MySQL, MongoDB, or none for SQLite/JSON)

2. **Install dependencies:**
```bash
go mod download
```

3. **Configure storage** (see Configuration section)

4. **Run the application:**
```bash
go run cmd/server/main.go
```

## 🗄️ Storage Backends

GoTask supports multiple storage backends, each optimized for different use cases:

### 1. PostgreSQL (Recommended for Production)
- **Best for**: Production environments, complex queries, ACID compliance
- **Features**: Advanced indexing, stored procedures, views
- **Configuration**: See PostgreSQL Configuration section

### 2. MySQL
- **Best for**: Web applications, high concurrency
- **Features**: UTF-8 support, stored procedures, views
- **Configuration**: See MySQL Configuration section

### 3. MongoDB
- **Best for**: Document-based storage, flexible schemas
- **Features**: Aggregation pipelines, text search, horizontal scaling
- **Configuration**: See MongoDB Configuration section

### 4. SQLite
- **Best for**: Development, small deployments, embedded applications
- **Features**: Zero configuration, file-based, ACID compliance
- **Configuration**: Set `STORAGE_TYPE=sqlite`

### 5. JSON File
- **Best for**: Development, testing, simple deployments
- **Features**: Human-readable, version control friendly
- **Configuration**: Set `STORAGE_TYPE=json` (default)

## ⚙️ Configuration

Configure the application using environment variables:

### General Settings
```bash
STORAGE_TYPE=postgres          # Storage backend: json, sqlite, postgres, mysql, mongodb
PORT=8080                     # Server port
```

### PostgreSQL Configuration
```bash
STORAGE_TYPE=postgres
DB_HOST=localhost
DB_PORT=5432
DB_USER=gotask_user
DB_PASSWORD=gotask_password
DB_NAME=gotask
POSTGRES_SSL_MODE=disable
POSTGRES_TIMEZONE=UTC
```

### MySQL Configuration
```bash
STORAGE_TYPE=mysql
DB_HOST=localhost
DB_PORT=3306
DB_USER=gotask_user
DB_PASSWORD=gotask_password
DB_NAME=gotask
MYSQL_CHARSET=utf8mb4
MYSQL_PARSE_TIME=true
MYSQL_LOC=Local
```

### MongoDB Configuration
```bash
STORAGE_TYPE=mongodb
MONGODB_URI=mongodb://localhost:27017
DB_NAME=gotask
MONGODB_COLLECTION=tasks
MONGODB_CONNECT_TIMEOUT=10s
MONGODB_QUERY_TIMEOUT=5s
```

### File-based Storage
```bash
STORAGE_TYPE=json              # or sqlite
STORAGE_FILE_PATH=tasks.json   # or tasks.db for SQLite
```

## 📡 API Endpoints

### Tasks

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/tasks` | Get all tasks |
| `GET` | `/api/v1/tasks?status=done` | Get completed tasks |
| `GET` | `/api/v1/tasks?status=undone` | Get pending tasks |
| `POST` | `/api/v1/tasks` | Create a new task |
| `GET` | `/api/v1/tasks/{id}` | Get a specific task |
| `PUT` | `/api/v1/tasks/{id}` | Update a task |
| `DELETE` | `/api/v1/tasks/{id}` | Delete a task |
| `GET` | `/api/v1/tasks/due` | Get tasks due in the next 7 days |
| `GET` | `/api/v1/tasks/due?days=3` | Get tasks due in the next 3 days |

### Health Check

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/health` | Health check endpoint |

### Example API Usage

#### Create a Task
```bash
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Complete project documentation",
    "due_date": "2024-01-20T15:00:00Z"
  }'
```

#### Get All Tasks
```bash
curl -X GET http://localhost:8080/api/v1/tasks
```

#### Update a Task
```bash
curl -X PUT http://localhost:8080/api/v1/tasks/{task-id} \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Updated task title",
    "done": true
  }'
```

#### Delete a Task
```bash
curl -X DELETE http://localhost:8080/api/v1/tasks/{task-id}
```

## 🐳 Docker Deployment

### Full Stack with Docker Compose

The `docker-compose.yml` includes:
- GoTask application
- PostgreSQL with pgAdmin
- MySQL with phpMyAdmin
- MongoDB with Mongo Express
- Persistent volumes for data
- Health checks for all services

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f gotask

# Stop all services
docker-compose down

# Stop and remove volumes (⚠️ This will delete all data)
docker-compose down -v
```

### Individual Service Management

```bash
# Start only PostgreSQL
docker-compose up -d postgres pgadmin

# Start only MySQL
docker-compose up -d mysql phpmyadmin

# Start only MongoDB
docker-compose up -d mongodb mongo-express
```

### Environment-specific Configurations

Create environment-specific compose files:

**docker-compose.prod.yml** (Production):
```yaml
version: '3.8'
services:
  gotask:
    environment:
      - STORAGE_TYPE=postgres
      - DB_HOST=postgres
      - DB_USER=gotask_user
      - DB_PASSWORD=secure_password
```

**docker-compose.dev.yml** (Development):
```yaml
version: '3.8'
services:
  gotask:
    environment:
      - STORAGE_TYPE=json
      - STORAGE_FILE_PATH=/app/data/tasks.json
```

Run with specific configuration:
```bash
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d
```

## 🧪 Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests for a specific package
go test ./internal/task

# Run storage tests (requires databases)
POSTGRES_TEST_DSN="postgres://user:pass@localhost/test" go test ./internal/storage

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### Storage Integration Tests

To run integration tests for different storage backends:

```bash
# PostgreSQL tests
export POSTGRES_TEST_HOST=localhost
export POSTGRES_TEST_USER=postgres
export POSTGRES_TEST_PASSWORD=password
export POSTGRES_TEST_DB=gotask_test
go test ./internal/storage -run TestPostgreSQL

# MySQL tests
export MYSQL_TEST_HOST=localhost
export MYSQL_TEST_USER=root
export MYSQL_TEST_PASSWORD=password
export MYSQL_TEST_DB=gotask_test
go test ./internal/storage -run TestMySQL

# MongoDB tests
export MONGODB_TEST_URI=mongodb://localhost:27017
export MONGODB_TEST_DB=gotask_test
go test ./internal/storage -run TestMongoDB
```

## 🏗️ Development

### Project Structure

```
GoTask_Management/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── api/                     # HTTP API layer
│   │   ├── handlers.go          # HTTP handlers
│   │   ├── middleware.go        # HTTP middleware
│   │   ├── server.go           # HTTP server setup
│   │   └── *_test.go           # API tests
│   ├── models/                  # Data models
│   │   └── task.go             # Task model
│   ├── storage/                 # Storage layer
│   │   ├── storage.go          # Storage interface
│   │   ├── json_storage.go     # JSON file storage
│   │   ├── sqlite_storage.go   # SQLite storage
│   │   ├── postgres_storage.go # PostgreSQL storage
│   │   ├── mysql_storage.go    # MySQL storage
│   │   ├── mongodb_storage.go  # MongoDB storage
│   │   ├── factory.go          # Storage factory
│   │   └── *_test.go           # Storage tests
│   └── task/                    # Business logic
│       ├── service.go          # Task service
│       └── service_test.go     # Service tests
├── scripts/                     # Database initialization
│   ├── postgres-init.sql
│   ├── mysql-init.sql
│   └── mongo-init.js
├── docker-compose.yml           # Docker Compose configuration
├── Dockerfile                   # Docker image definition
├── go.mod                       # Go module definition
└── README.md                    # This file
```

### Building

```bash
# Build for current platform
go build -o gotask cmd/server/main.go

# Build for Linux
GOOS=linux GOARCH=amd64 go build -o gotask-linux cmd/server/main.go

# Build for Windows
GOOS=windows GOARCH=amd64 go build -o gotask.exe cmd/server/main.go

# Build Docker image
docker build -t gotask:latest .
```

### Adding New Storage Backends

1. Implement the `Storage` interface in `internal/storage/`
2. Add configuration options in `factory.go`
3. Add tests following the pattern in existing `*_test.go` files
4. Update documentation

## 📊 Monitoring and Observability

### Health Checks

The application provides health check endpoints:

```bash
# Application health
curl http://localhost:8080/health

# Database-specific health (when using database storage)
curl http://localhost:8080/health/storage
```

### Logging

The application uses structured logging with different levels:

- **INFO**: General application flow
- **WARN**: Potentially harmful situations
- **ERROR**: Error events that might still allow the application to continue
- **DEBUG**: Detailed information for debugging (enable with `LOG_LEVEL=debug`)

### Metrics

Basic metrics are logged for:
- Request duration
- Request count by endpoint
- Storage operation performance
- Error rates

## 🔒 Security Considerations

### Database Security

- Use strong passwords for database users
- Enable SSL/TLS for database connections in production
- Regularly update database software
- Implement proper firewall rules

### API Security

- Consider implementing authentication/authorization
- Use HTTPS in production
- Implement rate limiting
- Validate all input data

### Docker Security

- Use non-root users in containers
- Regularly update base images
- Scan images for vulnerabilities
- Use secrets management for sensitive data

## 🚀 Production Deployment

### Environment Variables for Production

```bash
# Application
STORAGE_TYPE=postgres
PORT=8080
LOG_LEVEL=info

# Database (use secure values)
DB_HOST=your-db-host
DB_PORT=5432
DB_USER=gotask_user
DB_PASSWORD=secure_random_password
DB_NAME=gotask
POSTGRES_SSL_MODE=require
```

### Recommended Production Setup

1. **Use PostgreSQL or MySQL** for production workloads
2. **Enable SSL/TLS** for database connections
3. **Set up monitoring** and alerting
4. **Implement backups** for your database
5. **Use a reverse proxy** (nginx, Traefik) for HTTPS termination
6. **Set up log aggregation** (ELK stack, Fluentd)
7. **Implement health checks** in your orchestrator

### Kubernetes Deployment

Example Kubernetes manifests:

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: gotask
spec:
  replicas: 3
  selector:
    matchLabels:
      app: gotask
  template:
    metadata:
      labels:
        app: gotask
    spec:
      containers:
      - name: gotask
        image: gotask:latest
        ports:
        - containerPort: 8080
        env:
        - name: STORAGE_TYPE
          value: "postgres"
        - name: DB_HOST
          value: "postgres-service"
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: gotask-secrets
              key: db-password
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass (`go test ./...`)
6. Commit your changes (`git commit -m 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

### Development Guidelines

- Follow Go best practices and conventions
- Write tests for new functionality
- Update documentation for API changes
- Use meaningful commit messages
- Ensure code is properly formatted (`go fmt`)
- Run linters (`golangci-lint run`)

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Gorilla Mux](https://github.com/gorilla/mux) for HTTP routing
- [GORM](https://gorm.io/) for ORM functionality
- [MongoDB Go Driver](https://github.com/mongodb/mongo-go-driver) for MongoDB support
- [Docker](https://www.docker.com/) for containerization
- The Go community for excellent tooling and libraries

---

**Happy Task Managing! 🎯**
