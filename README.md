# SQL2API - Powerful SQL to REST API Server

[![Build Status](https://github.com/zaops/sql2api/workflows/Build%20and%20Release%20SQL2API/badge.svg)](https://github.com/zaops/sql2api/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

SQL2API is a high-performance API server that converts SQL operations into RESTful endpoints, supporting native SQL queries, structured queries, and batch operations with PostgreSQL and Oracle databases.

[中文文档](README_CN.md) | [English](README.md)

## ✨ Features

### 🔍 SQL Engine
- **Native SQL Support**: Execute raw SQL statements with parameterized queries
- **Structured Queries**: JSON-based queries automatically converted to SQL
- **Batch Operations**: Support for transactional and non-transactional batch SQL execution
- **Convenient Inserts**: Simplified insert operations with conflict handling
- **Pagination & Sorting**: Built-in pagination and sorting capabilities

### 🛡️ Security
- **API Key Authentication**: Secure authentication with API key management
- **Fine-grained Permissions**: Operation-level permission control (query, insert, update, delete, batch)
- **IP Whitelist**: IP address and CIDR-based access control
- **SQL Injection Protection**: Multi-layer security validation to prevent SQL injection attacks
- **Table Whitelist**: Configurable table and operation access control

### 📊 Database Support
- **PostgreSQL**: Full support with PostgreSQL-specific features
- **Oracle**: Complete Oracle database integration
- **Multi-dialect**: Automatic database dialect detection and adaptation

### ⚡ Performance & Monitoring
- **Performance Monitoring**: Query execution time tracking and slow query detection
- **Memory Optimization**: Result set size limits and memory usage optimization
- **Error Handling**: Detailed error code system with database-specific error mapping
- **Health Checks**: Built-in health check endpoints

### 🔧 Developer Experience
- **Swagger Documentation**: Complete API documentation with interactive UI
- **Configuration Driven**: Flexible configuration via YAML files and environment variables
- **Graceful Shutdown**: Support for graceful service shutdown
- **Comprehensive Examples**: Detailed usage examples and best practices

## 🚀 Quick Start

### Installation

#### Option 1: Download Pre-built Binaries

Download the latest release for your platform from the [Releases](https://github.com/zaops/sql2api/releases) page:

- **Linux (x64)**: `sql2api-linux-amd64.tar.gz`
- **Linux (ARM64)**: `sql2api-linux-arm64.tar.gz`
- **Windows (x64)**: `sql2api-windows-amd64.zip`
- **macOS (Intel)**: `sql2api-darwin-amd64.tar.gz`
- **macOS (Apple Silicon)**: `sql2api-darwin-arm64.tar.gz`

```bash
# Extract the archive
tar -xzf sql2api-linux-amd64.tar.gz  # Linux/macOS
# or unzip sql2api-windows-amd64.zip  # Windows

cd sql2api-linux-amd64/

# Make executable (Linux/macOS only)
chmod +x sql2api

# Run the server
./sql2api
```

#### Option 2: Fork & GitHub Actions Build

1. Fork the [SQL2API repository](https://github.com/zaops/sql2api) on GitHub
2. In your fork, navigate to the Actions tab
3. Enable GitHub Actions (if not already enabled)
4. Run the "Build and Release SQL2API" workflow
5. After the build completes, download the binaries from the workflow run artifacts

Alternatively, you can clone your fork and build locally:

```bash
# Clone your fork
git clone https://github.com/<your-username>/sql2api.git
cd sql2api

# Install dependencies
go mod tidy

# Generate Swagger documentation
go run cmd/server/main.go swagger

# Build for current platform
go build -o sql2api cmd/server/main.go

# Run the server
./build/sql2api
```

### Configuration

1. **Copy the configuration template**:
```bash
cp config.yaml config.yaml.local
```

2. **Edit the configuration file**:
```yaml
# Database configuration
database:
  type: "postgres"  # or "oracle"
  host: "localhost"
  port: 5432
  name: "your_database"
  username: "your_username"
  password: "your_password"

# SQL engine configuration
sql:
  enabled: true
  max_query_time: 30
  max_result_size: 1000
  allowed_tables: ["items", "categories", "orders"]
  allowed_actions: ["select", "insert", "update", "delete"]

# API Keys configuration
api_keys:
  enabled: true
  keys:
    - key: "your-api-key-here"
      name: "Admin Key"
      permissions: ["sql.*"]
      active: true
```

3. **Start the server**:
```bash
./sql2api
```

4. **Access the API**:
- **Swagger UI**: http://localhost:8080/swagger/index.html
- **Health Check**: http://localhost:8080/health

## 📖 API Usage

### Authentication

All SQL endpoints require API Key authentication. Include your API key in the request header:

```bash
curl -H "X-API-Key: your-api-key-here" \
     -H "Content-Type: application/json" \
     -X POST http://localhost:8080/api/v1/sql \
     -d '{"database_type": "postgres", "sql": "SELECT * FROM items LIMIT 10"}'
```

### Core Endpoints

#### 1. Execute SQL Query
```http
POST /api/v1/sql
```

**Native SQL Example**:
```json
{
  "database_type": "postgres",
  "sql": "SELECT id, name, category FROM items WHERE active = $1",
  "params": {"active": true},
  "pagination": {"page": 1, "page_size": 20},
  "sort": {"sort_by": "created_at", "sort_order": "desc"}
}
```

**Structured Query Example**:
```json
{
  "database_type": "postgres",
  "query": {
    "table": "items",
    "action": "select",
    "fields": ["id", "name", "category"],
    "where": {"active": true, "category": "electronics"},
    "order_by": [{"field": "created_at", "order": "desc"}],
    "limit": 50
  }
}
```

#### 2. Batch SQL Operations
```http
POST /api/v1/sql/batch
```

```json
{
  "database_type": "postgres",
  "transactional": true,
  "operations": [
    {
      "database_type": "postgres",
      "sql": "INSERT INTO items (name, category) VALUES ($1, $2)",
      "params": {"name": "New Item", "category": "electronics"}
    },
    {
      "database_type": "postgres",
      "sql": "UPDATE items SET active = $1 WHERE id = $2",
      "params": {"active": true, "id": 1}
    }
  ]
}
```

#### 3. Convenient Insert
```http
POST /api/v1/sql/insert
```

```json
{
  "database_type": "postgres",
  "table": "items",
  "data": {
    "name": "New Product",
    "category": "electronics",
    "price": 299.99,
    "active": true
  },
  "on_conflict": "ignore",
  "return_fields": ["id", "created_at"]
}
```

#### 4. Batch Insert
```http
POST /api/v1/sql/batch-insert
```

```json
{
  "database_type": "postgres",
  "table": "items",
  "data": [
    {"name": "Product 1", "category": "electronics", "price": 199.99},
    {"name": "Product 2", "category": "books", "price": 29.99},
    {"name": "Product 3", "category": "electronics", "price": 399.99}
  ],
  "on_conflict": "update"
}
```

## 🔐 Security & Permissions

### Permission System

SQL2API uses a fine-grained permission system:

- `sql.query`: SELECT operations
- `sql.insert`: INSERT operations
- `sql.update`: UPDATE operations
- `sql.delete`: DELETE operations
- `sql.batch`: Batch operations
- `sql.*`: All SQL operations

### Security Features

- **SQL Injection Protection**: Multi-layer validation and parameterized queries
- **Table Whitelist**: Only allow access to specified tables
- **Operation Control**: Restrict allowed SQL operations
- **Query Complexity Limits**: Prevent resource-intensive queries
- **Result Size Limits**: Control memory usage

## 📊 Monitoring & Observability

### Performance Metrics

- Query execution time tracking
- Slow query detection and logging
- Memory usage monitoring
- Error rate tracking

### Health Checks

```bash
curl http://localhost:8080/health
```

### Logging

SQL2API provides comprehensive logging:
- Query execution logs
- Performance metrics
- Security events
- Error tracking

## 🛠️ Development

### Building

```bash
# Install build dependencies
make deps

# Run tests
make test

# Run linting
make lint

# Build for current platform
make build

# Build for all platforms
make build-all

# Create release packages
make package

# Complete release build
make release
```

### Project Structure

```
sql2api/
├── cmd/server/          # Application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── handler/         # HTTP handlers
│   ├── middleware/      # HTTP middleware
│   ├── model/           # Data models
│   ├── repository/      # Data access layer
│   ├── service/         # Business logic
│   └── sql/             # SQL engine and security
├── docs/                # Swagger documentation
├── examples/            # Usage examples
└── config.yaml          # Configuration template
```

## 📚 Documentation

- **API Documentation**: Available at `/swagger/index.html` when running
- **Usage Examples**: See [examples/sql_examples.md](examples/sql_examples.md)
- **Configuration Guide**: See [config.yaml](config.yaml)
- **中文文档**: See [README_CN.md](README_CN.md)

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Gin](https://github.com/gin-gonic/gin) - HTTP web framework
- [GORM](https://gorm.io/) - ORM library
- [Swagger](https://swagger.io/) - API documentation
- [Viper](https://github.com/spf13/viper) - Configuration management

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/zaops/sql2api/issues)
- **Email**: [zhangzhiao@proton.me](mailto:zhangzhiao@proton.me)

---

**SQL2API** - Transform your database into a powerful REST API! 🚀
```
