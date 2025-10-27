# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Structure

This is a multi-service project:

- **slack-assistant/** - Go-based Slack bot service (main application)
- Future services will be added as separate directories

All services can be built and managed from the root Makefile.

## Build and Development Commands

Use the provided Makefile from the project root for all development tasks:

```bash
# Set up development environment
make dev-setup

# Build the application
make build

# Build for local development
make build-local

# Run the application locally
make run

# Run all CI checks (format, vet, lint, test)
make ci

# Run individual checks (all services)
make fmt          # Format all code
make vet          # Run all vet checks
make lint         # Run all linters
make test         # Run all unit tests
make test-race    # Run all tests with race detection
make test-coverage # Run all tests with coverage report

# Go-specific targets
make fmt-go          # Format Go code
make vet-go          # Run go vet
make lint-go         # Run golangci-lint
make test-go         # Run Go unit tests
make test-race-go    # Run Go tests with race detection
make test-coverage-go # Run Go tests with coverage report
make mock-generate-go # Generate Go mock files
make build-go        # Build Go service
make build-local-go  # Build Go service for local development
make run-go          # Build and run Go service

# Container operations (supports docker or podman)
make container-build    # Build all container images
make container-build-go # Build Go container image
make container-push     # Push all images to quay.io registry
make container-push-go  # Push Go image to registry
make container-run      # Run container locally
make container-run-go   # Run Go container locally

# Use podman instead of docker
CONTAINER_RUNTIME=podman make container-build

# Clean build artifacts
make clean

# Show version information
make version

# Show all available targets
make help
```

### Makefile Variables

- `CONTAINER_RUNTIME`: Set to `docker` (default) or `podman`
- `GOLANGCI_LINT_VERSION`: Version of golangci-lint to install (default: v1.61.0)
- `VERSION`: Build version (auto-detected from git or set manually)
- `PLATFORM`: Container build platform (default: linux/amd64)

## Required Environment Variables

- `ANYTHINGLLM_HOST`: Host URL for AnythingLLM instance
- `ANYTHINGLLM_API_KEY`: API key for AnythingLLM authentication

## Architecture Overview

This is a Go-based Slack AI Assistant bot that integrates with AnythingLLM. The architecture follows a clean separation of concerns.

### Directory Structure

All Go code is located in the `slack-assistant/` directory:

- `slack-assistant/cmd/server/` - Main application entry point
- `slack-assistant/pkg/` - Core packages
- `slack-assistant/go.mod` - Go module dependencies
- `slack-assistant/Dockerfile` - Container image definition

### Core Components

1. **Agent (`slack-assistant/pkg/agent/`)**: Central orchestrator handling command parsing and business logic
   - `agent.go`: Main agent implementation with command handlers (`answer`, `answer-all`, `inject`, `elaborate`)
   - `workerpool.go`: Concurrent event processing with configurable worker pool (default: 10 workers, queue size: 200)

2. **Slack Bot (`slack-assistant/pkg/slack-bot/`)**: Slack API integration using Socket Mode for real-time WebSocket communication

3. **LLM Client (`slack-assistant/pkg/llm/`)**: AnythingLLM integration using custom Go SDK
   - Creates workspace threads with project-version naming (e.g., `sriov-4-dot-16`)
   - Handles chat interactions and document injection

4. **Database (`slack-assistant/pkg/database/`)**: SQLite-based persistence using GORM
   - Maps Slack thread timestamps to AnythingLLM thread slugs
   - Auto-migration on startup

### Event Flow

1. User mentions bot in Slack (`@bot-name command project version`)
2. Event received via Socket Mode WebSocket
3. Event queued in worker pool for concurrent processing
4. Agent parses command and retrieves/creates thread mapping
5. LLM query sent to AnythingLLM with workspace context
6. Response posted back to Slack thread

### Key Features

- **Thread Management**: Maintains conversation context across Slack threads
- **Concurrent Processing**: Worker pool handles multiple events simultaneously
- **Graceful Shutdown**: Signal handling (SIGINT, SIGTERM) for clean termination
- **Debug Mode**: Detailed logging of WebSocket, API calls, and worker activity

## Bot Commands

When mentioned (`@bot-name`), the bot accepts these commands:

- `answer <project> <version>`: Analyzes last message in thread for AI response
- `answer-all <project> <version>`: Uses entire thread conversation for context
- `inject <project> <version>`: Injects user messages into AI knowledge base
- `elaborate`: Expands/explains last message using specialized workspace

Examples:
- `@bot-name answer sriov 4.16`
- `@bot-name answer-all metallb 4.18`
- `@bot-name inject sriov 4.16`
- `@bot-name elaborate`

## Dependencies

- `github.com/spf13/cobra`: CLI framework
- `github.com/slack-go/slack`: Slack API SDK with Socket Mode
- `github.com/SchSeba/anythingllm-go-sdk`: Custom AnythingLLM Go SDK
- `gorm.io/gorm` + `gorm.io/driver/sqlite`: ORM and SQLite driver

## Testing

The project uses Ginkgo and Gomega for testing with mock generation via mockgen.

Run tests using:
- `make test-go` - Run all Go tests from project root
- `cd slack-assistant && go test ./...` - Run tests from Go service directory
- `cd slack-assistant && go test ./pkg/agent` - Test specific package
- `make test-coverage-go` - Generate coverage report
- `make mock-generate-go` - Regenerate mock files

Linting is configured with golangci-lint (`.golangci.yml` in slack-assistant directory)

## Database

SQLite database (`slack-ai-assistant.db`) auto-created in the working directory (typically `slack-assistant/cmd/server/`). Contains:
- `SlackThreadToSlug` table mapping Slack thread timestamps to AnythingLLM thread slugs
- Auto-migration runs on startup
- Database file is .gitignored

## Common Issues

- Ensure Slack app has Socket Mode enabled with proper OAuth scopes
- Verify `ANYTHINGLLM_HOST` and `ANYTHINGLLM_API_KEY` environment variables are set
- Bot requires both `--bot-token` (xoxb-) and `--app-token` (xapp-) parameters
- Check workspace permissions in AnythingLLM for the specified projects/versions