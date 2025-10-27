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

### 2. AI Backend Setup

You can choose between two AI backends:

#### Option A: AnythingLLM (Default)

1. **Install AnythingLLM**: Set up your AnythingLLM instance
2. **Get API Credentials**: Obtain your AnythingLLM API key
3. **Set Environment Variables**:
   ```bash
   export ANYTHINGLLM_HOST="your-anythingllm-host"
   export ANYTHINGLLM_API_KEY="your-api-key"
   ```

#### Option B: LlamaIndex with Gemini (RAG-based)

1. **Get Gemini API Key**: Obtain from [Google AI Studio](https://makersuite.google.com/app/apikey)
2. **Set Environment Variables**:
   ```bash
   export GEMINI_API_KEY="your-gemini-api-key"
   export AI_BACKEND="llamaindex"
   export LLAMAINDEX_HOST="http://localhost:5000"  # or your server URL
   ```
3. **Add Knowledge Base**: Place documentation in `rag-data/{project}/{version}/` directories
4. **Build Indexes** (for local development):
   ```bash
   cd llamaindex-server
   ./setup.sh  # Sets up uv and installs dependencies
   uv run build_indexes.py --data ../rag-data --out ./local-storage
   ```
5. **Run LlamaIndex Server** (local):
   ```bash
   cd llamaindex-server
   STORAGE_ROOT=./local-storage \
   DELTA_ROOT=./local-delta \
   INJECT_ROOT=./local-injected \
   STATE_ROOT=./local-state \
   uv run app.py
   ```

   Or using Docker:
   ```bash
   cd llamaindex-server
   # Build image (NO API key needed - never pass secrets at build time!)
   docker build -t llamaindex-server .
   
   # Run container (API key passed at runtime only)
   docker run -p 5000:5000 \
     -e GEMINI_API_KEY=$GEMINI_API_KEY \
     -v $(pwd)/local-delta:/app/storage-delta \
     -v $(pwd)/local-injected:/app/injected \
     -v $(pwd)/local-state:/app/state \
     llamaindex-server
   ```
   
   **Security Note**: The GEMINI_API_KEY is **NEVER** passed during build time or baked into the Docker image. It is only provided at runtime via environment variable. On first startup, the server will build indexes from the `rag-data` directory (takes a few minutes). Subsequent startups load pre-built indexes instantly.

## Project Structure

This is a multi-service project containing:

- **slack-assistant/** - Go-based Slack bot service (main application)
- **llamaindex-server/** - Python-based LlamaIndex RAG service (optional AI backend)
- **rag-data/** - Knowledge base for LlamaIndex RAG

## Installation

```bash
# Clone the repository
git clone <your-repo-url>
cd slack-ai-assistant

# Install dependencies for Go service
cd slack-assistant
go mod tidy

# Build the application
go build -o slack-ai-assistant ./cmd/server

# Or use make from the project root
cd ..
make build-go
```

## Usage

### Environment Variables

**For AnythingLLM (default):**
```bash
export ANYTHINGLLM_HOST="your-anythingllm-host"
export ANYTHINGLLM_API_KEY="your-api-key"
```

**For LlamaIndex:**
```bash
export AI_BACKEND="llamaindex"
export LLAMAINDEX_HOST="http://localhost:5000"
export GEMINI_API_KEY="your-gemini-api-key"
```

**LlamaIndex Server Configuration (optional tuning):**
```bash
export MIN_HITS=2                    # Minimum retrieval hits before answering
export SIMILARITY_CUTOFF=0.75        # Minimum similarity score for retrieved docs
export CONFIDENCE_THRESHOLD=0.2      # Minimum confidence to answer (vs "I don't know")
export TOP_K=5                       # Number of documents to retrieve
export TEMPERATURE=0.0               # LLM temperature (0.0 = deterministic)
```

### Running the Bot

```bash
# From the slack-assistant directory
cd slack-assistant

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

# Or use make from the project root
cd ..
make run-go
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
├── slack-assistant/        # Go-based Slack bot service
│   ├── cmd/server/
│   │   ├── main.go              # Main application entry point
│   │   └── slack-ai-assistant.db # SQLite database (auto-created)
│   ├── pkg/
│   │   ├── agent/
│   │   │   ├── agent.go         # Core agent logic and command handling
│   │   │   └── workerpool.go    # Worker pool for concurrent processing
│   │   ├── database/
│   │   │   └── database.go      # Database interface and operations
│   │   ├── llm/
│   │   │   ├── llm.go          # AnythingLLM client implementation
│   │   │   └── types.go        # LLM-related type definitions
│   │   ├── mocks/              # Generated mock files for testing
│   │   └── slack-bot/
│   │       └── slack-bot.go    # Slack API and Socket Mode handling
│   ├── go.mod                  # Go module dependencies
│   ├── go.sum                  # Dependency checksums
│   ├── Dockerfile              # Container image for Go service
│   └── .golangci.yml          # Go linter configuration
├── Makefile                   # Build automation for all services
├── README.md                  # Project documentation
└── .gitignore                # Git ignore patterns
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

### Go Dependencies
- `github.com/spf13/cobra` - CLI framework
- `github.com/slack-go/slack` - Official Slack Go SDK
- `github.com/SchSeba/anythingllm-go-sdk` - AnythingLLM Go SDK
- `github.com/google/uuid` - UUID generation for LlamaIndex threads
- SQLite database driver and ORM functionality

### Python Dependencies (LlamaIndex server)
- `Flask` - Web framework
- `llama-index-core` - RAG framework core
- `llama-index-llms-google-genai` - Google GenAI LLM (latest unified SDK)
- `llama-index-embeddings-google-genai` - Google GenAI embeddings (latest unified SDK)
- `google-generativeai` - Google AI SDK (installed as dependency)

## Troubleshooting

### Common Issues

1. **Connection Failed**: 
   - Verify bot token and app token are correct
   - Ensure Socket Mode is enabled in your Slack app
   - Check that your app has the required OAuth scopes

2. **AI Responses Not Working**:
   - **AnythingLLM**: Verify `ANYTHINGLLM_HOST` and `ANYTHINGLLM_API_KEY` environment variables; check AnythingLLM instance is accessible; ensure workspaces exist
   - **LlamaIndex**: Verify `GEMINI_API_KEY` is set; ensure LlamaIndex server is running on `LLAMAINDEX_HOST`; check that base indexes exist in storage directory; verify project/version docs exist in `rag-data/`

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

1. Add new methods to the LLM interface in `llm/types.go`
2. Implement in both `llm/llm.go` (AnythingLLM) and `llm/llamaindex.go` (LlamaIndex)
3. Update agent to use new LLM capabilities

### Adding Knowledge to LlamaIndex

1. Create directory structure: `rag-data/{project}/{version}/`
2. Add markdown, text, or other documents to the directory
3. **For local development**: Rebuild indexes:
   ```bash
   cd llamaindex-server
   uv run build_indexes.py --data ../rag-data --out ./local-storage
   ```
4. **For Docker**: Rebuild the image and restart:
   ```bash
   docker-compose down
   docker-compose build llamaindex-server  # Rebuilds with new docs
   docker-compose up
   ```
   The indexes will be built automatically on first startup (no API key in build!)
5. **For production**: Consider pre-building indexes locally and mounting as volume for faster startup

### LlamaIndex Architecture

**Security**: 
- ✅ **GEMINI_API_KEY is NEVER in the Docker image** - Only passed at runtime as environment variable
- ✅ Images can be safely pushed to public registries
- ✅ No secrets in build args or layers

**Abstention Logic**: The server uses strict "I don't know" behavior when:
- Fewer than `MIN_HITS` relevant documents are retrieved
- Document similarity scores are below `SIMILARITY_CUTOFF`
- Aggregate confidence is below `CONFIDENCE_THRESHOLD`

**Data Persistence**:
- Base indexes: Built on first container startup from `rag-data` directory, or prebuilt locally via `build_indexes.py`
- Delta indexes: Runtime injections stored as JSONL and indexed incrementally
- Thread memory: Per-thread conversation history in JSON files
- All persisted data survives server restarts

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
