# Makefile for Slack AI Assistant Project

# Variables
APP_NAME := slack-ai-assistant
GO_SERVICE_DIR := slack-assistant
BINARY_NAME := $(APP_NAME)
GO_VERSION := 1.24
MAIN_PATH := ./cmd/server
BUILD_DIR := bin
CONTAINER_REGISTRY := quay.io
CONTAINER_REPO := $(CONTAINER_REGISTRY)/schseba/$(APP_NAME)

# Container runtime (docker or podman)
CONTAINER_RUNTIME ?= docker

# Tool versions
GOLANGCI_LINT_VERSION ?= v2.3.0

# Get version from git
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT_HASH := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

# Go build flags
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT_HASH) -X main.buildTime=$(BUILD_TIME)"

# Container build platform
PLATFORM ?= linux/amd64

.PHONY: help
help: ## Display this help message
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: clean
clean: clean-go ## Clean all build artifacts
	@echo "Cleaning all build artifacts..."

.PHONY: clean-go
clean-go: ## Clean Go build artifacts
	@echo "Cleaning Go build artifacts..."
	cd $(GO_SERVICE_DIR) && rm -rf $(BUILD_DIR)
	cd $(GO_SERVICE_DIR) && rm -f $(BINARY_NAME)
	cd $(GO_SERVICE_DIR) && go clean -cache
	$(CONTAINER_RUNTIME) image prune -f --filter label=app=$(APP_NAME) || true

.PHONY: deps
deps: deps-go ## Download and tidy all dependencies

.PHONY: deps-go
deps-go: ## Download and tidy Go dependencies
	@echo "Downloading Go dependencies..."
	cd $(GO_SERVICE_DIR) && go mod download
	cd $(GO_SERVICE_DIR) && go mod tidy

.PHONY: fmt
fmt: fmt-go ## Format all code

.PHONY: fmt-go
fmt-go: ## Format Go code
	@echo "Formatting Go code..."
	cd $(GO_SERVICE_DIR) && go fmt ./...

.PHONY: vet
vet: vet-go ## Run all vet checks

.PHONY: vet-go
vet-go: ## Run go vet
	@echo "Running go vet..."
	cd $(GO_SERVICE_DIR) && go vet ./...

.PHONY: lint
lint: lint-go ## Run all linters

.PHONY: lint-go
lint-go: ## Run golangci-lint
	@echo "Running golangci-lint..."
	@if ! which golangci-lint > /dev/null; then \
		echo "Installing golangci-lint $(GOLANGCI_LINT_VERSION)..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin $(GOLANGCI_LINT_VERSION); \
	fi
	cd $(GO_SERVICE_DIR) && golangci-lint run

.PHONY: test
test: test-go ## Run all unit tests

.PHONY: test-go
test-go: ## Run Go unit tests
	@echo "Running Go unit tests..."
	cd $(GO_SERVICE_DIR) && go test -v ./...

.PHONY: test-race
test-race: test-race-go ## Run all unit tests with race detection

.PHONY: test-race-go
test-race-go: ## Run Go unit tests with race detection
	@echo "Running Go unit tests with race detection..."
	cd $(GO_SERVICE_DIR) && go test -v -race ./...

.PHONY: test-coverage
test-coverage: test-coverage-go ## Run all tests with coverage report

.PHONY: test-coverage-go
test-coverage-go: ## Run Go tests with coverage report
	@echo "Running Go tests with coverage..."
	cd $(GO_SERVICE_DIR) && go test -v -coverprofile=coverage.out ./...
	cd $(GO_SERVICE_DIR) && go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: $(GO_SERVICE_DIR)/coverage.html"

.PHONY: mock-generate
mock-generate: mock-generate-go ## Generate all mock files

.PHONY: mock-generate-go
mock-generate-go: ## Generate Go mock files using mockgen
	@echo "Generating Go mock files..."
	cd $(GO_SERVICE_DIR) && mkdir -p pkg/mocks/database pkg/mocks/slack-bot pkg/mocks/llm
	cd $(GO_SERVICE_DIR) && go run go.uber.org/mock/mockgen@v0.5.2 -source=pkg/database/database.go -destination=pkg/mocks/database/mock_database.go -package=database
	cd $(GO_SERVICE_DIR) && go run go.uber.org/mock/mockgen@v0.5.2 -source=pkg/slack-bot/slack-bot.go -destination=pkg/mocks/slack-bot/mock_slack_bot.go -package=slackbot
	cd $(GO_SERVICE_DIR) && go run go.uber.org/mock/mockgen@v0.5.2 -source=pkg/llm/types.go -destination=pkg/mocks/llm/mock_llm.go -package=llm
	@echo "Go mock files generated successfully!"

.PHONY: build
build: build-go ## Build all services

.PHONY: build-go
build-go: deps-go fmt-go vet-go ## Build the Go application
	@echo "Building $(BINARY_NAME) version $(VERSION)..."
	cd $(GO_SERVICE_DIR) && mkdir -p $(BUILD_DIR)
	cd $(GO_SERVICE_DIR) && CGO_ENABLED=1 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Binary built: $(GO_SERVICE_DIR)/$(BUILD_DIR)/$(BINARY_NAME)"

.PHONY: build-local
build-local: build-local-go ## Build all services for local development

.PHONY: build-local-go
build-local-go: ## Build Go app for local development (no deps check)
	@echo "Building $(BINARY_NAME) for local development..."
	cd $(GO_SERVICE_DIR) && CGO_ENABLED=1 go build $(LDFLAGS) -o $(BINARY_NAME) $(MAIN_PATH)

.PHONY: run
run: run-go ## Run the application locally

.PHONY: run-go
run-go: build-local-go ## Build and run the Go application locally
	@echo "Running $(BINARY_NAME)..."
	@echo "Note: You need to provide --bot-token and --app-token flags"
	cd $(GO_SERVICE_DIR) && ./$(BINARY_NAME) --help

.PHONY: install-tools
install-tools: install-tools-go ## Install all required development tools

.PHONY: install-tools-go
install-tools-go: ## Install required Go development tools
	@echo "Installing Go development tools..."
	@if ! which golangci-lint > /dev/null; then \
		echo "Installing golangci-lint $(GOLANGCI_LINT_VERSION)..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin $(GOLANGCI_LINT_VERSION); \
	fi

# Container targets
.PHONY: container-build
container-build: container-build-go ## Build all container images

.PHONY: container-build-go
container-build-go: ## Build Go container image
	@echo "Building container image $(CONTAINER_REPO):$(VERSION) using $(CONTAINER_RUNTIME)..."
	$(CONTAINER_RUNTIME) build \
		--platform $(PLATFORM) \
		--build-arg GO_VERSION=$(GO_VERSION) \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT_HASH=$(COMMIT_HASH) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--label app=$(APP_NAME) \
		--label version=$(VERSION) \
		--label commit=$(COMMIT_HASH) \
		-t $(CONTAINER_REPO):$(VERSION) \
		-t $(CONTAINER_REPO):latest \
		-f $(GO_SERVICE_DIR)/Dockerfile \
		$(GO_SERVICE_DIR)

.PHONY: container-push
container-push: container-push-go ## Push all container images to registry

.PHONY: container-push-go
container-push-go: ## Push Go container image to registry
	@echo "Pushing container image to $(CONTAINER_REPO) using $(CONTAINER_RUNTIME)..."
	$(CONTAINER_RUNTIME) push $(CONTAINER_REPO):$(VERSION)
	$(CONTAINER_RUNTIME) push $(CONTAINER_REPO):latest

.PHONY: container-run
container-run: container-run-go ## Run container locally

.PHONY: container-run-go
container-run-go: ## Run Go container locally
	@echo "Running container locally using $(CONTAINER_RUNTIME)..."
	@echo "Note: You need to set SLACK_BOT_TOKEN, SLACK_APP_TOKEN, ANYTHINGLLM_HOST, and ANYTHINGLLM_API_KEY environment variables"
	$(CONTAINER_RUNTIME) run --rm -it \
		-e SLACK_BOT_TOKEN \
		-e SLACK_APP_TOKEN \
		-e ANYTHINGLLM_HOST \
		-e ANYTHINGLLM_API_KEY \
		$(CONTAINER_REPO):$(VERSION) \
		--bot-token "$$SLACK_BOT_TOKEN" \
		--app-token "$$SLACK_APP_TOKEN"

# CI/CD targets
.PHONY: ci
ci: deps fmt vet lint test ## Run all CI checks
	@echo "All CI checks completed successfully!"

.PHONY: release
release: ci build container-build container-push ## Full release pipeline
	@echo "Release $(VERSION) completed successfully!"

# Development targets
.PHONY: dev-setup
dev-setup: install-tools deps ## Set up development environment
	@echo "Development environment set up!"

.PHONY: check
check: fmt vet lint ## Run code quality checks
	@echo "Code quality checks completed!"

# Registry login helper
.PHONY: container-login
container-login: ## Login to container registry
	@echo "Logging into $(CONTAINER_REGISTRY) using $(CONTAINER_RUNTIME)..."
	@read -p "Username: " username; \
	$(CONTAINER_RUNTIME) login $(CONTAINER_REGISTRY) -u $$username

# Show build info
.PHONY: version
version: ## Show version information
	@echo "Application: $(APP_NAME)"
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT_HASH)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Container Runtime: $(CONTAINER_RUNTIME)"
	@echo "Container Repo: $(CONTAINER_REPO)"
	@echo "Golangci-lint Version: $(GOLANGCI_LINT_VERSION)"

# LlamaIndex Server targets
.PHONY: llamaindex-build-indexes
llamaindex-build-indexes: ## Build LlamaIndex base indexes from rag-data
	@echo "Building LlamaIndex indexes..."
	@if [ -z "$$GEMINI_API_KEY" ]; then \
		echo "Error: GEMINI_API_KEY environment variable is required"; \
		exit 1; \
	fi
	cd llamaindex-server && python build_indexes.py --data ../rag-data --out ./local-storage

.PHONY: llamaindex-server
llamaindex-server: ## Run LlamaIndex server locally
	@echo "Starting LlamaIndex server..."
	@if [ -z "$$GEMINI_API_KEY" ]; then \
		echo "Error: GEMINI_API_KEY environment variable is required"; \
		exit 1; \
	fi
	cd llamaindex-server && \
	STORAGE_ROOT=./local-storage \
	DELTA_ROOT=./local-delta \
	INJECT_ROOT=./local-injected \
	STATE_ROOT=./local-state \
	FLASK_APP=app.py \
	flask run

.PHONY: llamaindex-test
llamaindex-test: ## Run LlamaIndex server tests
	@echo "Running LlamaIndex server tests..."
	cd llamaindex-server && pytest test_server.py -v

.PHONY: llamaindex-container-build
llamaindex-container-build: ## Build LlamaIndex server container
	@echo "Building LlamaIndex container..."
	@echo "Note: GEMINI_API_KEY is passed at runtime, not build time (secure!)"
	$(CONTAINER_RUNTIME) build \
		-t $(CONTAINER_REPO)-llamaindex:$(VERSION) \
		-t $(CONTAINER_REPO)-llamaindex:latest \
		-f llamaindex-server/Dockerfile \
		.

.PHONY: llamaindex-setup
llamaindex-setup: ## Set up Python environment for LlamaIndex
	@echo "Setting up LlamaIndex Python environment..."
	cd llamaindex-server && python -m venv .venv
	@echo "Virtual environment created. Activate with:"
	@echo "  cd llamaindex-server && source .venv/bin/activate"
	@echo "Then install dependencies:"
	@echo "  pip install -r requirements.txt"

.PHONY: docker-compose-up
docker-compose-up: ## Start all services with docker-compose
	@echo "Starting services with docker-compose..."
	@if [ -z "$$GEMINI_API_KEY" ]; then \
		echo "Error: GEMINI_API_KEY environment variable is required"; \
		exit 1; \
	fi
	docker-compose up -d

.PHONY: docker-compose-down
docker-compose-down: ## Stop all services with docker-compose
	@echo "Stopping services with docker-compose..."
	docker-compose down

.PHONY: docker-compose-logs
docker-compose-logs: ## Show logs from docker-compose services
	docker-compose logs -f

.PHONY: podman-compose-up
podman-compose-up: ## Start all services with podman compose
	@echo "Starting services with podman compose..."
	@if [ -z "$$GEMINI_API_KEY" ]; then \
		echo "Error: GEMINI_API_KEY environment variable is required"; \
		exit 1; \
	fi
	podman compose up -d

.PHONY: podman-compose-down
podman-compose-down: ## Stop all services with podman compose
	@echo "Stopping services with podman compose..."
	podman compose down

.PHONY: podman-compose-logs
podman-compose-logs: ## Show logs from podman compose services
	podman compose logs -f
