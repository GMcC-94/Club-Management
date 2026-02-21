.PHONY: help install-tools generate run build clean test docker-build docker-up docker-down

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

install-tools: ## Install required tools (templ)
	@echo "Installing templ..."
	@go install github.com/a-h/templ/cmd/templ@latest
	@echo "Done! Make sure $(go env GOPATH)/bin is in your PATH"

generate: ## Generate templ files
	@echo "Generating templ files..."
	@templ generate
	@echo "Done!"

watch: ## Watch for changes and regenerate templ files
	@echo "Watching for changes... (Ctrl+C to stop)"
	@templ generate --watch

run: generate ## Generate templates and run the server
	@echo "Starting server..."
	@go run cmd/server/main.go

build: generate ## Build the application binary
	@echo "Building application..."
	@go build -o bin/club-server cmd/server/main.go
	@echo "Binary created at: bin/club-server"

clean: ## Clean generated files and build artifacts
	@echo "Cleaning..."
	@find web/templates -name "*_templ.go" -delete
	@rm -rf bin/
	@echo "Done!"

test: ## Run tests
	@echo "Running tests..."
	@go test -v ./...

docker-build: ## Build Docker images
	@echo "Building Docker images..."
	@docker-compose build

docker-up: ## Start Docker containers
	@echo "Starting Docker containers..."
	@docker-compose up -d
	@echo "Application running at http://localhost:8080"

docker-down: ## Stop Docker containers
	@echo "Stopping Docker containers..."
	@docker-compose down

docker-logs: ## Show Docker logs
	@docker-compose logs -f

setup: install-tools ## First time setup
	@echo "Setting up development environment..."
	@cp .env.example .env
	@echo "Edit .env file with your database credentials"
	@echo "Then run: make generate && make run"
