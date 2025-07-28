# Slack AI Assistant

A sophisticated Slack bot built with Go that integrates with AnythingLLM to provide AI-powered assistance. The bot uses Cobra CLI framework, Slack's Socket Mode for real-time communication, and includes a worker pool for concurrent event processing.

## Features

- ✅ **Socket Mode**: Real-time WebSocket communication with Slack
- ✅ **AI Integration**: Powered by AnythingLLM for intelligent responses
- ✅ **Worker Pool**: Concurrent event processing with configurable worker count
- ✅ **Database Integration**: SQLite database for thread and conversation management
- ✅ **App Mentions**: Responds to bot mentions with AI-powered commands
- ✅ **Thread Management**: Maintains conversation context across Slack threads
- ✅ **Document Injection**: Ability to inject content into AI knowledge base
- ✅ **Content Elaboration**: AI-powered content expansion and explanation
- ✅ **Cobra CLI**: Professional command-line interface with comprehensive flags
- ✅ **Graceful Shutdown**: Signal handling for clean application termination
- ✅ **Debug Mode**: Configurable logging and debugging

## Prerequisites

### 1. Slack App Setup

1. **Create a Slack App**:
   - Go to [api.slack.com/apps](https://api.slack.com/apps)
   - Click "Create New App" → "From scratch"
   - Choose a name and workspace

2. **Configure Bot Permissions**:
   - Go to "OAuth & Permissions"
   - Add these Bot Token Scopes:
     - `app_mentions:read` - To receive app mention events
     - `channels:history` - To read messages in channels
     - `chat:write` - To send messages
     - `commands` - For slash commands (if needed)

3. **Enable Socket Mode**:
   - Go to "Socket Mode" in your app settings
   - Enable Socket Mode
   - Generate an App-Level Token with `connections:write` scope

4. **Enable Events**:
   - Go to "Event Subscriptions"
   - Enable Events
   - Subscribe to Bot Events:
     - `app_mention` - When someone mentions your bot

5. **Install the App**:
   - Go to "Install App" 
   - Install to your workspace
   - Copy the "Bot User OAuth Token" (starts with `xoxb-`)
   - Copy the "App-Level Token" (starts with `xapp-`)

### 2. AnythingLLM Setup

1. **Install AnythingLLM**: Set up your AnythingLLM instance
2. **Get API Credentials**: Obtain your AnythingLLM API key
3. **Set Environment Variables**:
   ```bash
   export ANYTHINGLLM_HOST="your-anythingllm-host"
   export ANYTHINGLLM_API_KEY="your-api-key"
   ```

## Installation

```bash
# Clone the repository
git clone <your-repo-url>
cd slack-ai-assistant

# Install dependencies
go mod tidy

# Build the application
go build -o slack-ai-assistant ./cmd/server
```

## Usage

### Environment Variables

Set the following environment variables:

```bash
export ANYTHINGLLM_HOST="your-anythingllm-host"
export ANYTHINGLLM_API_KEY="your-api-key"
```

### Running the Bot

```bash
# Basic usage
./slack-ai-assistant --bot-token "xoxb-your-bot-token" --app-token "xapp-your-app-token"

# With custom worker count and debug mode
./slack-ai-assistant \
  --bot-token "xoxb-your-bot-token" \
  --app-token "xapp-your-app-token" \
  --workers 20 \
  --debug

# Using environment variables for tokens
export SLACK_BOT_TOKEN="xoxb-your-bot-token"
export SLACK_APP_TOKEN="xapp-your-app-token"
./slack-ai-assistant --bot-token "$SLACK_BOT_TOKEN" --app-token "$SLACK_APP_TOKEN"
```

### Command Line Options

- `--bot-token`, `-b`: Slack Bot Token (required)
- `--app-token`, `-a`: Slack App Token (required)  
- `--workers`, `-w`: Number of workers for concurrent event processing (default: 10)
- `--debug`, `-d`: Enable debug mode for detailed logging
- `--help`, `-h`: Show help message

## Bot Commands

### AI Assistant Commands

Mention the bot (`@your-bot-name`) followed by one of these commands:

#### 1. Answer Questions
```
@bot-name answer <project> <version>
```
- Analyzes the last message in the thread and provides an AI-generated answer
- Uses the specified project and OpenShift version for context
- Example: `@bot-name answer sriov 4.16`

#### 2. Answer with Full Thread Context
```
@bot-name answer-all <project> <version>
```
- Similar to `answer` but uses the entire thread conversation for context
- Provides more comprehensive responses based on full conversation history
- Example: `@bot-name answer-all metallb 4.18`

#### 3. Inject Content
```
@bot-name inject <project> <version>
```
- Injects the user's recent messages into the AI knowledge base
- Helps improve future responses by adding domain-specific information
- Example: `@bot-name inject sriov 4.16`

#### 4. Elaborate Content
```
@bot-name elaborate
```
- Provides detailed explanation or expansion of the last message in the thread
- Uses a specialized "elaborate" workspace for enhanced explanations
- No project/version parameters needed

### Error Handling

If incorrect parameters are provided, the bot will respond with helpful usage instructions.

## Architecture

### Project Structure

```
slack-ai-assistant/
├── cmd/server/
│   ├── main.go              # Main application entry point
│   └── slack-ai-assistant.db # SQLite database (auto-created)
├── pkg/
│   ├── agent/
│   │   ├── agent.go         # Core agent logic and command handling
│   │   └── workerpool.go    # Worker pool for concurrent processing
│   ├── database/
│   │   └── database.go      # Database interface and operations
│   ├── llm/
│   │   ├── llm.go          # AnythingLLM client implementation
│   │   └── types.go        # LLM-related type definitions
│   └── slack-bot/
│       └── slack-bot.go    # Slack API and Socket Mode handling
├── go.mod                  # Go module dependencies
├── go.sum                  # Dependency checksums
└── README.md              # This file
```

### Key Components

1. **Agent**: Handles business logic and AI interactions
2. **Worker Pool**: Manages concurrent event processing
3. **Database**: Stores thread mappings and conversation state
4. **LLM Client**: Interfaces with AnythingLLM for AI responses
5. **Slack Bot**: Manages Slack API communication

### Event Flow

1. User mentions bot in Slack thread
2. Event received via Socket Mode
3. Event queued in worker pool
4. Worker processes command and parameters
5. Agent retrieves thread context from database
6. AI query sent to AnythingLLM
7. Response posted back to Slack thread

## Database

The bot automatically creates and manages a SQLite database (`slack-ai-assistant.db`) for:

- Thread mapping between Slack and AnythingLLM
- Conversation state management
- Database auto-migration on startup

## Dependencies

- `github.com/spf13/cobra` - CLI framework
- `github.com/slack-go/slack` - Official Slack Go SDK
- `github.com/SchSeba/anythingllm-go-sdk` - AnythingLLM Go SDK
- SQLite database driver and ORM functionality

## Troubleshooting

### Common Issues

1. **Connection Failed**: 
   - Verify bot token and app token are correct
   - Ensure Socket Mode is enabled in your Slack app
   - Check that your app has the required OAuth scopes

2. **AI Responses Not Working**:
   - Verify `ANYTHINGLLM_HOST` and `ANYTHINGLLM_API_KEY` environment variables
   - Check AnythingLLM instance is accessible
   - Ensure the specified project workspaces exist in AnythingLLM

3. **Database Errors**:
   - Check write permissions in the application directory
   - Ensure SQLite dependencies are properly installed

4. **Worker Pool Issues**:
   - Adjust worker count with `--workers` flag based on load
   - Monitor for memory usage with high worker counts

### Debug Mode

Run with `--debug` flag to see detailed logs:
- WebSocket connection status
- All incoming events and processing details
- AI API call information
- Database operations
- Worker pool activity

### Performance Tuning

- **Worker Count**: Adjust `--workers` based on expected load (default: 10)
- **Queue Size**: Worker pool uses a queue size of 200 events
- **Database**: SQLite provides good performance for typical Slack bot usage

## Development

### Adding New Commands

1. Add command parsing logic in `agent.go`
2. Implement command handler function
3. Update this README with command documentation

### Extending AI Features

1. Add new methods to the LLM interface in `llm/llm.go`
2. Implement AnythingLLM API calls
3. Update agent to use new LLM capabilities

### Custom Event Handling

1. Extend event types in the Socket Mode handler
2. Add new worker item types in `workerpool.go`
3. Implement processing logic in the agent

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
