# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Development Commands

Use the provided Makefile for all development tasks:

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

# Run individual checks
make fmt          # Format code
make vet          # Run go vet
make lint         # Run golangci-lint
make test         # Run unit tests
make test-race    # Run tests with race detection
make test-coverage # Run tests with coverage report

# Container operations (supports docker or podman)
make container-build    # Build container image
make container-push     # Push to quay.io registry
make container-run      # Run container locally

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

This is a Go-based Slack AI Assistant bot that integrates with AnythingLLM. The architecture follows a clean separation of concerns:

### Core Components

1. **Agent (`pkg/agent/`)**: Central orchestrator handling command parsing and business logic
   - `agent.go`: Main agent implementation with command handlers (`answer`, `answer-all`, `inject`, `elaborate`)
   - `workerpool.go`: Concurrent event processing with configurable worker pool (default: 10 workers, queue size: 200)

2. **Slack Bot (`pkg/slack-bot/`)**: Slack API integration using Socket Mode for real-time WebSocket communication

3. **LLM Client (`pkg/llm/`)**: AnythingLLM integration using custom Go SDK
   - Creates workspace threads with project-version naming (e.g., `sriov-4-dot-16`)
   - Handles chat interactions and document injection

4. **Database (`pkg/database/`)**: SQLite-based persistence using GORM
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

No custom test commands configured. Use standard Go testing:
- `go test ./...` for all packages
- `go test ./pkg/agent` for specific package testing
- No linting configuration found - consider adding golangci-lint

## Database

SQLite database (`slack-ai-assistant.db`) auto-created in working directory. Contains:
- `SlackThreadToSlug` table mapping Slack thread timestamps to AnythingLLM thread slugs
- Auto-migration runs on startup

## Common Issues

- Ensure Slack app has Socket Mode enabled with proper OAuth scopes
- Verify `ANYTHINGLLM_HOST` and `ANYTHINGLLM_API_KEY` environment variables are set
- Bot requires both `--bot-token` (xoxb-) and `--app-token` (xapp-) parameters
- Check workspace permissions in AnythingLLM for the specified projects/versions