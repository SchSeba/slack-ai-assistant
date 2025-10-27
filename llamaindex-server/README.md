# LlamaIndex RAG Server

Minimal Flask server providing RAG capabilities using LlamaIndex and Google Gemini.

## Features

- **Gemini-powered RAG**: Uses `gemini-2.5-pro` for chat and `text-embedding-004` for embeddings
- **Prebuilt base indexes**: Indexes built at Docker build time from `rag-data/`
- **Runtime injections**: Dynamic content injection with JSONL persistence and delta indexing
- **Thread memory**: Conversation history persisted per thread
- **Strict abstention**: Returns "I don't know" when evidence is insufficient

## API Endpoints

### POST /v1/answer
Answer questions using RAG over project/version knowledge base.

**Request:**
```json
{
  "project": "sriov",
  "version": "4.16",
  "thread_slug": "uuid-here",
  "message": "How do I configure SR-IOV?"
}
```

**Response:**
```json
{
  "textResponse": "To configure SR-IOV on OpenShift 4.16..."
}
```

### POST /v1/elaborate
Elaborate on content using pure chat (no retrieval).

**Request:**
```json
{
  "thread_slug": "uuid-here",
  "message": "Explain this in more detail"
}
```

**Response:**
```json
{
  "textResponse": "Let me elaborate..."
}
```

### POST /v1/inject
Inject content into the knowledge base for a project/version.

**Request:**
```json
{
  "project": "sriov",
  "version": "4.16",
  "textContent": "Additional configuration notes...",
  "metadata": {"source": "slack"}
}
```

**Response:**
```json
{
  "status": "ok"
}
```

### GET /health
Health check endpoint.

**Response:**
```json
{
  "status": "ok",
  "base_indexes": 2,
  "delta_indexes": 1
}
```

## Local Development

### Prerequisites
- Python 3.11+
- Google Gemini API key

### Setup

1. Create virtual environment:
```bash
python -m venv .venv
source .venv/bin/activate  # Windows: .venv\Scripts\activate
```

2. Install dependencies:
```bash
pip install -r requirements.txt
```

3. Set environment variables:
```bash
export GEMINI_API_KEY="your-key-here"
```

4. Build base indexes:
```bash
python build_indexes.py --data ../rag-data --out ./local-storage
```

5. Run server:
```bash
STORAGE_ROOT=./local-storage \
DELTA_ROOT=./local-delta \
INJECT_ROOT=./local-injected \
STATE_ROOT=./local-state \
FLASK_APP=app.py \
flask run
```

## Docker Deployment

### Build
```bash
docker build --build-arg GEMINI_API_KEY=$GEMINI_API_KEY -t llamaindex-server .
```

### Run
```bash
docker run -p 5000:5000 \
  -e GEMINI_API_KEY=$GEMINI_API_KEY \
  -v $(pwd)/delta:/app/storage-delta \
  -v $(pwd)/injected:/app/injected \
  -v $(pwd)/state:/app/state \
  llamaindex-server
```

## Configuration

Environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `GEMINI_API_KEY` | (required) | Google Gemini API key |
| `STORAGE_ROOT` | `/app/storage` | Base indexes directory |
| `DELTA_ROOT` | `/app/storage-delta` | Delta indexes directory |
| `STATE_ROOT` | `/app/state` | Thread memory directory |
| `INJECT_ROOT` | `/app/injected` | JSONL injection logs |
| `MIN_HITS` | `2` | Min retrieval hits to answer |
| `SIMILARITY_CUTOFF` | `0.75` | Min similarity for docs |
| `CONFIDENCE_THRESHOLD` | `0.2` | Min confidence to answer |
| `TOP_K` | `5` | Number of docs to retrieve |
| `TEMPERATURE` | `0.0` | LLM temperature |

## Data Layout

```
/app/
├── storage/              # Prebuilt base indexes
│   ├── sriov-4-dot-16/
│   └── metallb-4-dot-18/
├── storage-delta/        # Runtime delta indexes
│   └── sriov-4-dot-16/
├── injected/             # JSONL injection logs
│   └── sriov/
│       └── 4-dot-16/
│           └── 2025-10-27.jsonl
└── state/                # Thread memory
    └── threads/
        └── {thread-uuid}.json
```

## Testing

Run tests:
```bash
pytest test_server.py -v
```

## Abstention Behavior

The server implements strict "I don't know" responses when:

1. Fewer than `MIN_HITS` documents match the query above `SIMILARITY_CUTOFF`
2. Mean similarity score is below `CONFIDENCE_THRESHOLD`
3. LLM is instructed via system prompt to refuse answering when context is insufficient

This prevents hallucinations and ensures answers are grounded in the knowledge base.


