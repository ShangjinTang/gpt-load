# Default target
.DEFAULT_GOAL := help

# ==============================================================================
# Run & Development
# ==============================================================================
.PHONY: run
run: ## Run server
	@echo "--- Starting server... ---"
	go run ./main.go

.PHONY: dev
dev: ## Run in development mode (with race detection)
	@echo "🔧 Starting development mode..."
	go run -race ./main.go

# ==============================================================================
# Testing
# ==============================================================================
.PHONY: test
test: ## Run tests
	@echo "--- Running tests... ---"
	go test ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	@echo "--- Running tests with coverage... ---"
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: help
help: ## Display this help message
	@awk 'BEGIN {FS = ":.*?## "; printf "Usage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z0-9_-]+:.*?## / { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
