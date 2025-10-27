# Slack AI Assistant - Go Service

This is the Go-based Slack bot service that integrates with AnythingLLM to provide AI-powered assistance in Slack.

## Quick Start

See the main [project README](../README.md) for complete setup and usage instructions.

## Development

### Build

```bash
# From this directory
go build -o slack-ai-assistant ./cmd/server

# Or from project root using make
cd ..
make build-go
```

### Test

```bash
# From this directory
go test -v ./...

# Or from project root using make
cd ..
make test-go
```

### Run

```bash
./slack-ai-assistant --bot-token "xoxb-your-bot-token" --app-token "xapp-your-app-token"
```

## Structure

- `cmd/server/` - Main application entry point
- `pkg/agent/` - Core agent logic and worker pool
- `pkg/database/` - Database interface and operations
- `pkg/llm/` - AnythingLLM client
- `pkg/slack-bot/` - Slack API integration
- `pkg/mocks/` - Generated mock files for testing

## Dependencies

- Go 1.24+
- SQLite (CGO enabled)
- See `go.mod` for complete dependency list

