.PHONY: help build run test clean docker-build docker-up docker-down migrate lint fmt vet

# Variables
BINARY_NAME=movie-db-api
DOCKER_IMAGE=movie-db-api
DOCKER_TAG=latest
GO=go
GOTEST=$(GO) test
GOVET=$(GO) vet
GOFMT=gofmt

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

## Development
build: ## Build the application
	@echo "Building $(BINARY_NAME)..."
	$(GO) build -o $(BINARY_NAME) ./cmd/server/main.go

run: ## Run the application
	@echo "Running $(BINARY_NAME)..."
	$(GO) run ./cmd/server/main.go

dev: ## Run with hot reload (requires air)
	@which air > /dev/null || (echo "Installing air..." && go install github.com/cosmtrek/air@latest)
	air

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	rm -f main
	rm -f coverage.out
	rm -rf tmp/

## Testing
test: ## Run tests
	@echo "Running tests..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...

test-coverage: test ## Run tests with coverage report
	@echo "Generating coverage report..."
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

test-integration: ## Run integration tests
	@echo "Running integration tests..."
	$(GOTEST) -v -tags=integration ./...

bench: ## Run benchmarks
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...

## Code Quality
lint: ## Run linter
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run

fmt: ## Format code
	@echo "Formatting code..."
	$(GOFMT) -s -w .

vet: ## Run go vet
	@echo "Running go vet..."
	$(GOVET) ./...

check: fmt vet lint test ## Run all checks (fmt, vet, lint, test)

## Dependencies
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	$(GO) mod download

deps-update: ## Update dependencies
	@echo "Updating dependencies..."
	$(GO) get -u ./...
	$(GO) mod tidy

deps-verify: ## Verify dependencies
	@echo "Verifying dependencies..."
	$(GO) mod verify

## Docker
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

docker-build-prod: ## Build production Docker image
	@echo "Building production Docker image..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) -f Dockerfile .

docker-up: ## Start Docker services (development)
	@echo "Starting Docker services..."
	docker compose up -d

docker-down: ## Stop Docker services
	@echo "Stopping Docker services..."
	docker compose down

docker-logs: ## Show Docker logs
	docker compose logs -f

docker-clean: ## Clean Docker resources
	@echo "Cleaning Docker resources..."
	docker compose down -v
	docker image prune -f

docker-prod-up: ## Start production Docker services
	@echo "Starting production services..."
	docker compose -f docker-compose.prod.yml up -d

docker-prod-down: ## Stop production Docker services
	docker compose -f docker-compose.prod.yml down

## Database
db-create: ## Create database
	@echo "Creating database..."
	docker exec -it movie-db-postgres createdb -U postgres moviedb

db-drop: ## Drop database
	@echo "Dropping database..."
	docker exec -it movie-db-postgres dropdb -U postgres moviedb

db-reset: db-drop db-create ## Reset database
	@echo "Database reset complete"

db-migrate: ## Run database migrations
	@echo "Running migrations..."
	$(GO) run ./cmd/server/main.go

db-backup: ## Backup database
	@echo "Backing up database..."
	docker exec movie-db-postgres pg_dump -U postgres moviedb > backup_$$(date +%Y%m%d_%H%M%S).sql

db-restore: ## Restore database from backup (requires BACKUP_FILE variable)
	@echo "Restoring database from $(BACKUP_FILE)..."
	docker exec -i movie-db-postgres psql -U postgres -d moviedb < $(BACKUP_FILE)

db-shell: ## Connect to database shell
	docker exec -it movie-db-postgres psql -U postgres -d moviedb

## Swagger
swagger-gen: ## Generate Swagger documentation
	@which swag > /dev/null || (echo "Installing swag..." && go install github.com/swaggo/swag/cmd/swag@latest)
	swag init -g cmd/server/main.go

swagger-serve: swagger-gen run ## Generate Swagger docs and run server

## Production
prod-deploy: ## Deploy to production
	@echo "Deploying to production..."
	@echo "This will trigger the deployment workflow"
	git push origin main

prod-rollback: ## Rollback production deployment
	@echo "Rolling back production..."
	VERSION=previous docker compose -f docker-compose.prod.yml up -d

## Monitoring
logs: ## Show application logs
	@echo "Showing logs..."
	docker compose logs -f app

health: ## Check application health
	@echo "Checking health..."
	curl -f http://localhost:8080/health || echo "Service is down"

metrics: ## Show Prometheus metrics
	curl http://localhost:8080/metrics

## Security
security-scan: ## Run security scan
	@which gosec > /dev/null || (echo "Installing gosec..." && go install github.com/securego/gosec/v2/cmd/gosec@latest)
	gosec ./...

vuln-check: ## Check for vulnerabilities
	@which govulncheck > /dev/null || (echo "Installing govulncheck..." && go install golang.org/x/vuln/cmd/govulncheck@latest)
	govulncheck ./...

## CI/CD
ci-test: ## Run CI tests locally
	@echo "Running CI tests..."
	docker compose -f docker-compose.yml up -d postgres redis minio
	sleep 5
	make test
	docker compose down

ci-build: ## Build as in CI
	@echo "Building as in CI..."
	CGO_ENABLED=0 GOOS=linux $(GO) build -a -installsuffix cgo -o $(BINARY_NAME) ./cmd/server/main.go

## Initialize
init: ## Initialize project (first time setup)
	@echo "Initializing project..."
	cp .env.example .env
	@echo "Please update .env with your configuration"
	make deps
	make docker-up
	@echo "Waiting for services to start..."
	sleep 10
	make swagger-gen
	@echo "Setup complete! Run 'make run' to start the application"

## All-in-one commands
all: clean deps build test ## Clean, download deps, build, and test

install: build ## Build and install binary
	@echo "Installing $(BINARY_NAME)..."
	$(GO) install ./cmd/server

.DEFAULT_GOAL := help
