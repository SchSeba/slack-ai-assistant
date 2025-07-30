# Makefile for Slack AI Assistant

# Variables
APP_NAME := slack-ai-assistant
BINARY_NAME := $(APP_NAME)
GO_VERSION := 1.24
MAIN_PATH := ./cmd/server
BUILD_DIR := bin
CONTAINER_REGISTRY := quay.io
CONTAINER_REPO := $(CONTAINER_REGISTRY)/schseba/$(APP_NAME)

# Container runtime (docker or podman)
CONTAINER_RUNTIME ?= docker

# Tool versions
GOLANGCI_LINT_VERSION ?= v1.61.0

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
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -f $(BINARY_NAME)
	go clean -cache
	$(CONTAINER_RUNTIME) image prune -f --filter label=app=$(APP_NAME) || true

.PHONY: deps
deps: ## Download and tidy dependencies
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

.PHONY: fmt
fmt: ## Format Go code
	@echo "Formatting code..."
	go fmt ./...

.PHONY: vet
vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

.PHONY: lint
lint: ## Run golangci-lint
	@echo "Running golangci-lint..."
	@if ! which golangci-lint > /dev/null; then \
		echo "Installing golangci-lint $(GOLANGCI_LINT_VERSION)..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin $(GOLANGCI_LINT_VERSION); \
	fi
	golangci-lint run

.PHONY: test
test: ## Run unit tests
	@echo "Running unit tests..."
	go test -v ./...

.PHONY: test-race
test-race: ## Run unit tests with race detection
	@echo "Running unit tests with race detection..."
	go test -v -race ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: mock-generate
mock-generate: ## Generate mock files using mockgen
	@echo "Generating mock files..."
	@mkdir -p pkg/mocks/database pkg/mocks/slack-bot
	go run go.uber.org/mock/mockgen -source=pkg/database/database.go -destination=pkg/mocks/database/mock_database.go -package=database
	go run go.uber.org/mock/mockgen -source=pkg/slack-bot/slack-bot.go -destination=pkg/mocks/slack-bot/mock_slack_bot.go -package=slackbot
	@echo "Mock files generated successfully!"

.PHONY: build
build: deps fmt vet ## Build the application
	@echo "Building $(BINARY_NAME) version $(VERSION)..."
	mkdir -p $(BUILD_DIR)
	CGO_ENABLED=1 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Binary built: $(BUILD_DIR)/$(BINARY_NAME)"

.PHONY: build-local
build-local: ## Build for local development (no deps check)
	@echo "Building $(BINARY_NAME) for local development..."
	CGO_ENABLED=1 go build $(LDFLAGS) -o $(BINARY_NAME) $(MAIN_PATH)

.PHONY: run
run: build-local ## Build and run the application locally
	@echo "Running $(BINARY_NAME)..."
	@echo "Note: You need to provide --bot-token and --app-token flags"
	./$(BINARY_NAME) --help

.PHONY: install-tools
install-tools: ## Install required development tools
	@echo "Installing development tools..."
	@if ! which golangci-lint > /dev/null; then \
		echo "Installing golangci-lint $(GOLANGCI_LINT_VERSION)..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin $(GOLANGCI_LINT_VERSION); \
	fi

# Container targets
.PHONY: container-build
container-build: ## Build container image
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
		.

.PHONY: container-push
container-push: ## Push container image to registry
	@echo "Pushing container image to $(CONTAINER_REPO) using $(CONTAINER_RUNTIME)..."
	$(CONTAINER_RUNTIME) push $(CONTAINER_REPO):$(VERSION)
	$(CONTAINER_RUNTIME) push $(CONTAINER_REPO):latest

.PHONY: container-run
container-run: ## Run container locally
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
