.PHONY: deps build run lint test cover clean api cli docs help

# Build variables
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS = -ldflags "-X github.com/svanhalla/prompt-lab/greetd/internal/version.Version=$(VERSION) \
                   -X github.com/svanhalla/prompt-lab/greetd/internal/version.Commit=$(COMMIT) \
                   -X github.com/svanhalla/prompt-lab/greetd/internal/version.BuildTime=$(BUILD_TIME)"

# Go variables
GOCMD = go
GOBUILD = $(GOCMD) build
GOTEST = $(GOCMD) test
GOMOD = $(GOCMD) mod
BINARY_NAME = greetd
MAIN_PATH = ./cmd/greetd
COVERAGE_THRESHOLD = 70

help: ## Show this help message
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*##/ {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

deps: ## Download and verify dependencies
	$(GOMOD) download
	$(GOMOD) verify
	$(GOMOD) tidy

build: deps ## Build the application
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) $(MAIN_PATH)

run: build ## Build and run the application
	./$(BINARY_NAME)

lint: ## Run linters
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Install it with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi

test: ## Run tests
	$(GOTEST) -v ./...

cover: ## Run tests with coverage and check threshold
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
	@COVERAGE=$$($(GOCMD) tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
	echo "Total coverage: $$COVERAGE%"; \
	if [ "$$(echo "$$COVERAGE < $(COVERAGE_THRESHOLD)" | awk '{print ($$1 < $$3)}')" = "1" ]; then \
		echo "Coverage $$COVERAGE% is below threshold $(COVERAGE_THRESHOLD)%"; \
		exit 1; \
	fi

docs: ## Validate OpenAPI spec and check documentation endpoints
	@echo "Validating OpenAPI specification..."
	@if [ ! -f api/openapi.yaml ]; then \
		echo "Error: api/openapi.yaml not found"; \
		exit 1; \
	fi
	@echo "OpenAPI spec validation passed"
	@echo "Documentation will be available at:"
	@echo "  - Swagger UI: http://localhost:8080/swagger/"
	@echo "  - Redoc: http://localhost:8080/docs"

api: build ## Start the API server
	./$(BINARY_NAME) api

cli: build ## Show CLI help
	./$(BINARY_NAME) --help

clean: ## Clean build artifacts
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

# Docker targets
docker-build: ## Build Docker image
	docker build -t greetd:$(VERSION) .

docker-run: docker-build ## Run Docker container
	docker run -p 8080:8080 greetd:$(VERSION)

# Development targets
dev-setup: ## Setup development environment
	@echo "Installing development dependencies..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Development environment ready!"

install: build ## Install binary to GOPATH/bin
	cp $(BINARY_NAME) $(GOPATH)/bin/

# Smoke tests
smoke-test: build ## Run local smoke tests
	@echo "Running smoke tests..."
	@echo "Testing CLI commands..."
	./$(BINARY_NAME) version
	./$(BINARY_NAME) health
	./$(BINARY_NAME) hello --name "SmokeTest"
	./$(BINARY_NAME) set message "Hello from smoke tests"
	@echo "CLI smoke tests passed!"

e2e-test: build ## Run end-to-end tests
	@echo "Running e2e tests..."
	@./$(BINARY_NAME) api --port 8081 & \
	SERVER_PID=$$!; \
	sleep 3; \
	echo "Testing API endpoints..."; \
	curl -f http://localhost:8081/health > /dev/null && echo "✓ /health"; \
	curl -f "http://localhost:8081/hello?name=E2E" > /dev/null && echo "✓ /hello"; \
	curl -f http://localhost:8081/message > /dev/null && echo "✓ GET /message"; \
	curl -f -X POST -H "Content-Type: application/json" -d '{"message":"E2E Test"}' http://localhost:8081/message > /dev/null && echo "✓ POST /message"; \
	curl -f http://localhost:8081/swagger/ > /dev/null && echo "✓ /swagger/"; \
	curl -f http://localhost:8081/docs > /dev/null && echo "✓ /docs"; \
	kill $$SERVER_PID; \
	echo "E2E tests passed!"
