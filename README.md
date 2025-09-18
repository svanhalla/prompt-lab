# Greetd

A production-quality Go CLI application that manages greetings and messages, providing both command-line interface and web API functionality.

## Why "Greetd"?

The name "greetd" was chosen for its simplicity and memorability. It follows Unix naming conventions for daemon-like applications while being short, descriptive, and easy to type. The "d" suffix suggests it can run as a service, which aligns with its API server capabilities.

## Features

- **CLI Interface**: Cobra-based commands for health checks, greetings, and message management
- **HTTP API**: RESTful endpoints with Echo framework
- **Web UI**: Clean, Tailwind CSS-styled interface for message management
- **Persistence**: JSON-based storage for configuration and messages
- **Logging**: Structured logging with logrus, file rotation support
- **Configuration**: Viper-based config with environment variables and file support
- **Production Ready**: Docker support, CI/CD, comprehensive testing

## Quick Start

### Prerequisites

- Go 1.25.1 or later
- Make (optional, for build automation)

### Installation

```bash
# Clone the repository
git clone <repository-url>
cd greetd

# Install dependencies and build
make deps
make build

# Or build manually
go build -o greetd ./cmd/greetd
```

### Basic Usage

```bash
# Show version information
./greetd version

# Check application health
./greetd health

# Print a greeting
./greetd hello
./greetd hello --name "Alice"

# Set a message
./greetd set message "Hello from Greetd!"

# Start the API server
./greetd api
# or with custom host/port
./greetd api --host 127.0.0.1 --port 9000
```

## CLI Commands

### Global Flags

- `--config`: Path to config file (default: `~/.greetd/config.json`)
- `--log-level`: Log level (debug, info, warn, error)
- `--log-format`: Log format (text, json)

### Commands

#### `greetd version`
Prints version, commit, build time, and Go version information.

#### `greetd health`
Returns JSON health information including status, version, and timestamp.

#### `greetd hello [--name NAME]`
Prints a friendly greeting. If no name is provided, defaults to "World".

#### `greetd set message <text>`
Stores a message to disk that will be served by the API and Web UI.

#### `greetd api [--host HOST] [--port PORT]`
Starts the HTTP API and Web server.

## API Endpoints

The API server provides the following endpoints:

- `GET /health` - Health check with version info
- `GET /hello?name=<name>` - Greeting endpoint
- `GET /message` - Get current stored message
- `POST /message` - Update stored message (JSON body: `{"message": "text"}`)
- `GET /ui` - Web interface for message management
- `GET /logs` - View recent application logs
- `GET /swagger/` - Swagger UI for API documentation
- `GET /docs` - Redoc API documentation
- `GET /swagger/openapi.yaml` - OpenAPI specification

### API Documentation

Interactive API documentation is available when the server is running:

- **Swagger UI**: http://localhost:8080/swagger/
- **Redoc**: http://localhost:8080/docs

Both interfaces are automatically generated from the OpenAPI 3.1 specification located at `api/openapi.yaml`.

### Example API Usage

```bash
# Health check
curl http://localhost:8080/health

# Get greeting
curl "http://localhost:8080/hello?name=Alice"

# Get current message
curl http://localhost:8080/message

# Update message
curl -X POST http://localhost:8080/message \
  -H "Content-Type: application/json" \
  -d '{"message": "Hello from API!"}'

# Access web UI
open http://localhost:8080/ui
```

## Configuration

### Configuration File

Default location: `~/.greetd/config.json`

```json
{
  "server": {
    "host": "0.0.0.0",
    "port": 8080
  },
  "logging": {
    "level": "info",
    "format": "text"
  },
  "data_path": "/home/user/.greetd"
}
```

### Environment Variables

All configuration can be overridden with environment variables using the `GREETD_` prefix:

- `GREETD_SERVER_HOST` - Server host (default: 0.0.0.0)
- `GREETD_SERVER_PORT` - Server port (default: 8080)
- `GREETD_LOGGING_LEVEL` - Log level (default: info)
- `GREETD_LOGGING_FORMAT` - Log format (default: text)
- `GREETD_DATA_PATH` - Data directory path

### Configuration Precedence

1. Command-line flags (highest priority)
2. Environment variables
3. Configuration file
4. Default values (lowest priority)

## Development

### Build Targets

```bash
make help          # Show available targets
make deps          # Download dependencies
make build         # Build the application
make run           # Build and run
make lint          # Run linters
make test          # Run tests
make cover         # Generate coverage report (70% threshold)
make docs          # Validate API documentation
make smoke-test    # Run local smoke tests
make runtime-verify # Run runtime verification tests
make e2e-test      # Run end-to-end tests
make clean         # Clean build artifacts
```

### Development Setup

```bash
# Install development tools
make dev-setup

# Run tests with coverage
make cover

# Run linters
make lint

# Start API server in development
make api
```

### Testing

The project includes comprehensive unit tests and table-driven test patterns:

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run runtime verification
make runtime-verify

# Run smoke tests
make smoke-test

# Run end-to-end tests
make e2e-test
```

#### Runtime Verification

The application includes comprehensive runtime verification that validates:

- **Server Startup**: Ensures the API and Web server start successfully without errors
- **Expected Responses**: Confirms all documented endpoints return correct responses:
  - `/health` → status JSON with `ok` and version info
  - `/hello?name=Test` → greeting JSON with `Hello, Test!`
  - `/message` → returns latest stored message
  - `POST /message` → updates and returns message JSON
  - `/ui` → HTML page renders current message and update form with Tailwind styling
  - `/logs` → HTML page renders recent logs in human-friendly format
  - `/swagger/` → Swagger UI loads and shows API documentation
  - `/docs` → Redoc page renders cleanly and matches OpenAPI spec
- **Core Endpoints**: Automated smoke/e2e tests validate critical functionality at runtime

### Docker

```bash
# Build Docker image
make docker-build

# Run in Docker
make docker-run

# Or manually
docker build -t greetd .
docker run -p 8080:8080 greetd
```

## Project Structure

```
greetd/
├── cmd/greetd/              # Main application entry point
├── internal/                # Internal packages
│   ├── api/                 # HTTP server and handlers
│   ├── cmd/                 # Cobra commands
│   ├── config/              # Configuration management
│   ├── logging/             # Logging setup
│   ├── storage/             # Data persistence
│   └── version/             # Version information
├── api/                     # OpenAPI specification
├── .github/workflows/       # CI/CD workflows
├── Dockerfile               # Container definition
├── Makefile                 # Build automation
├── .golangci.yml           # Linter configuration
├── .env.example            # Environment variables example
└── README.md               # This file
```

## API Documentation

Complete OpenAPI 3.1 specification is available at `api/openapi.yaml`. The spec includes:

- All endpoint definitions
- Request/response schemas
- Example payloads
- Error responses

## Architecture Decisions

### Technology Choices

- **Cobra + Viper**: Industry standard for Go CLI applications
- **Echo**: Lightweight, fast HTTP framework with good middleware support
- **Logrus**: Mature logging library with structured logging support
- **JSON Storage**: Simple, human-readable persistence suitable for the use case
- **Tailwind CSS**: Utility-first CSS framework for clean, modern UI

### Design Patterns

- **Dependency Injection**: Clean separation of concerns, testable code
- **Configuration Precedence**: Standard CLI pattern (flags > env > config > defaults)
- **Graceful Shutdown**: Proper signal handling for production deployment
- **Structured Logging**: JSON format for production, human-readable for development

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run `make lint` and `make test`
6. Submit a pull request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

For issues, questions, or contributions, please use the GitHub issue tracker.
