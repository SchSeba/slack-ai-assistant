# Slack AI Assistant

A sophisticated Slack bot built with Go that integrates with LlamaIndex RAG to provide AI-powered assistance with your documentation. The bot uses Cobra CLI framework, Slack's Socket Mode for real-time communication, and includes a worker pool for concurrent event processing.

## Features

- ✅ **Socket Mode**: Real-time WebSocket communication with Slack
- ✅ **RAG-Powered AI**: LlamaIndex with Google Gemini for intelligent, documentation-grounded responses
- ✅ **Worker Pool**: Concurrent event processing with configurable worker count
- ✅ **Database Integration**: SQLite database for thread and conversation management
- ✅ **App Mentions**: Responds to bot mentions with AI-powered commands
- ✅ **Thread Management**: Maintains conversation context across Slack threads
- ✅ **Document Injection**: Ability to inject content into AI knowledge base
- ✅ **Content Elaboration**: AI-powered content expansion and explanation
- ✅ **Docker Compose**: Easy multi-container deployment
- ✅ **Graceful Shutdown**: Signal handling for clean application termination
- ✅ **Debug Mode**: Configurable logging and debugging

## Quick Start with Docker Compose

### Prerequisites

1. **Docker & Docker Compose** installed on your system
2. **Slack App** configured with Socket Mode (see detailed setup below)
3. **Gemini API Key** from [Google AI Studio](https://makersuite.google.com/app/apikey)

### 1. Clone and Configure

```bash
# Clone the repository
git clone <your-repo-url>
cd slack-ai-assistant

# Create .env file with your credentials
cat > .env << 'EOF'
# Slack Credentials (required)
SLACK_BOT_TOKEN=xoxb-your-bot-token
SLACK_APP_TOKEN=xapp-your-app-token

# Gemini API Key (required)
GEMINI_API_KEY=your-gemini-api-key

# Optional: LlamaIndex tuning parameters
MIN_HITS=1
SIMILARITY_CUTOFF=0.5
CONFIDENCE_THRESHOLD=0.1
TOP_K=5
TEMPERATURE=0.0
EOF
```

### 2. Add Your Documentation

Place your documentation files in the `rag-data` directory:

```bash
# Structure: rag-data/{project}/{version}/
mkdir -p rag-data/myproject/1.0
# Add your markdown, text, or PDF files
cp /path/to/docs/* rag-data/myproject/1.0/
```

### 3. Build and Start Services

```bash
# Build Docker images
docker-compose build

# Start all services (LlamaIndex server + Slack bot)
make docker-compose-up

# View logs
make docker-compose-logs

# First startup builds indexes (takes a few minutes)
# Subsequent startups are instant
```

### 4. Stop Services

```bash
# Stop all services
make docker-compose-down

# Stop and remove volumes (clears all data)
docker-compose down -v
```

## Detailed Slack App Setup

### 1. Create a Slack App

1. **Go to** [api.slack.com/apps](https://api.slack.com/apps)
2. **Click** "Create New App" → "From scratch"
3. **Choose** a name and workspace

### 2. Configure Bot Permissions

1. **Go to** "OAuth & Permissions"
2. **Add these Bot Token Scopes**:
   - `app_mentions:read` - To receive app mention events
   - `channels:history` - To read messages in channels
   - `chat:write` - To send messages
   - `commands` - For slash commands (if needed)

### 3. Enable Socket Mode

1. **Go to** "Socket Mode" in your app settings
2. **Enable** Socket Mode
3. **Generate** an App-Level Token with `connections:write` scope
4. **Copy** the token (starts with `xapp-`)

### 4. Enable Events

1. **Go to** "Event Subscriptions"
2. **Enable** Events
3. **Subscribe to Bot Events**:
   - `app_mention` - When someone mentions your bot

### 5. Install the App

1. **Go to** "Install App"
2. **Install** to your workspace
3. **Copy** the "Bot User OAuth Token" (starts with `xoxb-`)
4. **Copy** the "App-Level Token" (starts with `xapp-`)

## Docker Compose Services

The application consists of two services:

### 1. LlamaIndex Server (`llamaindex-server`)
- **Python-based RAG service** using LlamaIndex and Google Gemini
- **Port**: 5000
- **Volumes**:
  - `./local-delta` - Runtime document injections
  - `./local-injected` - Injected content storage
  - `./local-state` - Thread memory and conversation state
- **Auto-builds indexes** from `rag-data/` on first startup

### 2. Slack Bot (`slack-bot`)
- **Go-based Slack client** with Socket Mode
- **Depends on**: llamaindex-server
- **Volume**: `./slack-bot-data` - SQLite database and bot data

## Configuration

### Environment Variables (in .env file)

**Required:**
```bash
SLACK_BOT_TOKEN=xoxb-your-bot-token          # From Slack OAuth & Permissions
SLACK_APP_TOKEN=xapp-your-app-token          # From Slack Socket Mode
GEMINI_API_KEY=your-gemini-api-key           # From Google AI Studio
```

**Optional - LlamaIndex Tuning:**
```bash
MIN_HITS=1                    # Minimum retrieval hits before answering (default: 1)
SIMILARITY_CUTOFF=0.5         # Minimum similarity score (default: 0.5)
CONFIDENCE_THRESHOLD=0.1      # Minimum confidence to answer (default: 0.1)
TOP_K=5                       # Number of documents to retrieve (default: 5)
TEMPERATURE=0.0               # LLM temperature, 0.0 = deterministic (default: 0.0)
```

### Docker Compose Commands

**Common Make Commands:**
```bash
# Start services
make docker-compose-up

# View logs
make docker-compose-logs

# Stop services
make docker-compose-down

# Show all available make commands
make help
```

**Additional Docker Compose Commands:**
```bash
# Build images (rebuild after adding new documents to rag-data/)
docker-compose build

# View logs for specific service
docker-compose logs -f slack-bot          # Only bot logs
docker-compose logs -f llamaindex-server  # Only LlamaIndex logs

# Restart specific service
docker-compose restart slack-bot
docker-compose restart llamaindex-server

# Stop and remove volumes (full cleanup)
docker-compose down -v

# Rebuild and restart
docker-compose build && make docker-compose-up
```

## Project Structure

This is a multi-service project:

```
slack-ai-assistant/
├── docker-compose.yml          # Multi-service orchestration
├── .env                        # Environment variables (create this)
├── rag-data/                   # Documentation for RAG
│   └── {project}/{version}/    # Organized by project and version
├── slack-assistant/            # Go-based Slack bot service
│   ├── cmd/server/
│   │   └── main.go            # Main application entry point
│   ├── pkg/
│   │   ├── agent/             # Core agent logic
│   │   ├── database/          # Database interface
│   │   ├── llm/               # LLM clients (LlamaIndex, AnythingLLM)
│   │   └── slack-bot/         # Slack API handling
│   ├── Dockerfile             # Slack bot container
│   ├── go.mod
│   └── go.sum
├── llamaindex-server/          # Python-based LlamaIndex RAG service
│   ├── app.py                 # Flask server
│   ├── build_indexes.py       # Pre-build indexes locally
│   ├── Dockerfile             # LlamaIndex container
│   ├── requirements.txt
│   └── pyproject.toml
└── Makefile                   # Build automation
```

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

### Key Components

1. **Slack Bot (Go)**
   - Socket Mode WebSocket handler
   - Worker pool for concurrent event processing
   - SQLite database for thread/conversation management
   - Command parsing and routing

2. **LlamaIndex Server (Python)**
   - Flask REST API
   - Vector store and retrieval system
   - Per-thread conversation memory
   - Document injection and indexing

3. **Data Flow**
   ```
   User mentions bot in Slack
         ↓
   Socket Mode receives event
         ↓
   Worker pool queues event
         ↓
   Agent processes command
         ↓
   Query sent to LlamaIndex server
         ↓
   RAG retrieval + LLM generation
         ↓
   Response posted to Slack thread
   ```

### Database

The Slack bot uses SQLite (`slack-ai-assistant.db`) for:
- Thread mapping between Slack and LlamaIndex
- Conversation state management
- Auto-migration on startup

### Security Best Practices

✅ **Secrets Management**:
- GEMINI_API_KEY is **NEVER** in the Docker image
- All secrets passed at runtime via environment variables
- Safe to push images to public registries

✅ **Volume Persistence**:
- Indexes, state, and injected docs persist across restarts
- Conversation memory maintained per thread
- Easy backup via Docker volumes

## Adding New Documentation

To add documentation for RAG:

### 1. Organize by Project and Version

```bash
# Create directory structure
mkdir -p rag-data/myproject/1.0

# Add your files (markdown, text, PDF)
cp /path/to/docs/*.md rag-data/myproject/1.0/
```

### 2. Rebuild and Restart

```bash
# Rebuild LlamaIndex server to include new docs
docker-compose build llamaindex-server

# Restart services
make docker-compose-up

# First startup with new docs will rebuild indexes (takes a few minutes)
# Watch logs to monitor progress
docker-compose logs -f llamaindex-server
```

### 3. Use in Commands

```
@bot-name answer myproject 1.0
@bot-name answer-all myproject 1.0
@bot-name inject myproject 1.0
```

## Dependencies

### Go Dependencies (Slack Bot)
- `github.com/spf13/cobra` - CLI framework
- `github.com/slack-go/slack` - Official Slack Go SDK
- `github.com/google/uuid` - UUID generation
- SQLite database driver

### Python Dependencies (LlamaIndex Server)
- `Flask` - Web framework
- `llama-index-core` - RAG framework
- `llama-index-llms-google-genai` - Google Gemini LLM
- `llama-index-embeddings-google-genai` - Google embeddings
- `google-generativeai` - Google AI SDK

## Troubleshooting

### Docker Compose Issues

**Services won't start:**
```bash
# Check logs
make docker-compose-logs

# Verify .env file exists with required variables
cat .env

# Check if ports are available
lsof -i :5000  # LlamaIndex port
```

**Slack bot can't connect:**
```bash
# Verify tokens in .env
docker-compose logs slack-bot

# Check Socket Mode is enabled in Slack app
# Verify bot token starts with 'xoxb-'
# Verify app token starts with 'xapp-'
```

**LlamaIndex not responding:**
```bash
# Check server logs
docker-compose logs llamaindex-server

# Verify GEMINI_API_KEY is valid
# Check if indexes are building (first startup takes time)
# Test server directly:
curl http://localhost:5000/health
```

**No AI responses or "I don't know":**
```bash
# Verify docs exist in rag-data/{project}/{version}/
ls -la rag-data/

# Rebuild indexes
docker-compose build llamaindex-server
make docker-compose-up

# Check retrieval parameters (may be too strict)
# Lower thresholds in .env:
# MIN_HITS=1
# SIMILARITY_CUTOFF=0.3
# CONFIDENCE_THRESHOLD=0.05
```

**Database errors:**
```bash
# Check volume permissions
make docker-compose-down
sudo chown -R $USER:$USER ./slack-bot-data
make docker-compose-up
```

### Performance Tuning

**LlamaIndex Server:**
- `TOP_K`: More documents = better context but slower
- `TEMPERATURE`: 0.0 for deterministic, 0.7 for creative
- `MIN_HITS`: Lower = more permissive answering

**Slack Bot:**
- Default worker pool: 10 concurrent events
- To adjust, modify Dockerfile CMD or override in docker-compose.yml

### Debug Mode

View detailed logs:
```bash
# All logs
make docker-compose-logs

# Just Slack bot with debug output
docker-compose logs -f slack-bot

# Just LlamaIndex server
docker-compose logs -f llamaindex-server
```

## Alternative Deployment Methods

### Using AnythingLLM Instead of LlamaIndex

If you prefer AnythingLLM over LlamaIndex:

1. **Set up AnythingLLM instance** (separate installation)
2. **Get API credentials** from your AnythingLLM dashboard
3. **Modify docker-compose.yml** to remove llamaindex-server
4. **Update slack-bot environment**:
   ```yaml
   slack-bot:
     environment:
       - AI_BACKEND=anythingllm
       - ANYTHINGLLM_HOST=your-anythingllm-host
       - ANYTHINGLLM_API_KEY=your-api-key
   ```
5. **Rebuild and start**: `docker-compose build && make docker-compose-up`

### Local Development (without Docker)

**LlamaIndex Server:**
```bash
cd llamaindex-server

# Install uv package manager
./setup.sh

# Build indexes from rag-data
uv run build_indexes.py --data ../rag-data --out ./local-storage

# Run server
GEMINI_API_KEY=your-key \
STORAGE_ROOT=./local-storage \
DELTA_ROOT=./local-delta \
INJECT_ROOT=./local-injected \
STATE_ROOT=./local-state \
uv run app.py
```

**Slack Bot:**
```bash
cd slack-assistant

# Install dependencies
go mod tidy

# Build
go build -o slack-ai-assistant ./cmd/server

# Run
AI_BACKEND=llamaindex \
LLAMAINDEX_HOST=http://localhost:5000 \
./slack-ai-assistant \
  --bot-token "xoxb-your-token" \
  --app-token "xapp-your-token" \
  --workers 10 \
  --debug
```

## Development

### Adding New Bot Commands

1. **Edit** `slack-assistant/pkg/agent/agent.go`
2. **Add command parsing** in `handleAppMention()`
3. **Implement handler function** 
4. **Update this README** with command documentation

### Extending AI Features

1. **Define interface** in `slack-assistant/pkg/llm/types.go`
2. **Implement for LlamaIndex** in `llm/llamaindex.go`
3. **Implement for AnythingLLM** in `llm/llm.go`
4. **Call from agent** in `agent.go`

### LlamaIndex Server Details

**Retrieval Strategy:**
- Vector similarity search with configurable `TOP_K`
- Similarity cutoff filtering
- Minimum hits requirement for answering

**Abstention Logic** - "I don't know" responses when:
- Fewer than `MIN_HITS` documents retrieved
- Similarity scores below `SIMILARITY_CUTOFF`
- Confidence below `CONFIDENCE_THRESHOLD`

**Data Persistence:**
- **Base indexes**: Built from `rag-data/` on first startup
- **Delta indexes**: Runtime injections (JSONL + vector index)
- **Thread memory**: JSON conversation history per thread
- **Volumes**: All data persists across restarts

**Security:**
- ✅ GEMINI_API_KEY never in Docker image
- ✅ Runtime-only secrets via environment
- ✅ Safe for public registries

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
