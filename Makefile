.PHONY: deps build run lint test cover clean api cli help

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

cover: ## Run tests with coverage
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

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
