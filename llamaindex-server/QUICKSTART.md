# LlamaIndex Server Quickstart

This guide will help you get the LlamaIndex RAG server running locally.

## Prerequisites

- Python 3.11+
- Google Gemini API key ([Get one here](https://makersuite.google.com/app/apikey))
- Documents in `rag-data/{project}/{version}/` directories

## Quick Start (Local)

1. **Set up Python environment:**

**Option A: Using uv (recommended - much faster):**
```bash
cd llamaindex-server
./setup.sh
```

**Option B: Using pip:**
```bash
cd llamaindex-server
python -m venv .venv
source .venv/bin/activate  # Windows: .venv\Scripts\activate
pip install -r requirements.txt
```

2. **Set your API key:**
```bash
export GEMINI_API_KEY="your-key-here"
```

3. **Build base indexes:**
```bash
python build_indexes.py --data ../rag-data --out ./local-storage
```

This will scan `../rag-data/` for `{project}/{version}/` directories and build vector indexes.

4. **Run the server:**
```bash
STORAGE_ROOT=./local-storage \
DELTA_ROOT=./local-delta \
INJECT_ROOT=./local-injected \
STATE_ROOT=./local-state \
FLASK_APP=app.py \
flask run
```

5. **Test the server:**
```bash
curl http://localhost:5000/health
```

## Quick Start (Docker)

1. **Build the image:**
```bash
export GEMINI_API_KEY="your-key-here"
docker build --build-arg GEMINI_API_KEY=$GEMINI_API_KEY -t llamaindex-server .
```

2. **Run the container:**
```bash
docker run -p 5000:5000 \
  -e GEMINI_API_KEY=$GEMINI_API_KEY \
  -v $(pwd)/local-delta:/app/storage-delta \
  -v $(pwd)/local-injected:/app/injected \
  -v $(pwd)/local-state:/app/state \
  llamaindex-server
```

## Adding Knowledge

1. Create directory: `../rag-data/{project}/{version}/`
2. Add markdown, text, or documentation files
3. Rebuild indexes: `python build_indexes.py --data ../rag-data --out ./local-storage`
4. Restart server

## Testing

Install test dependencies:
```bash
pip install -r requirements-test.txt
```

Run tests:
```bash
pytest test_server.py -v
```

## Example API Usage

**Answer a question:**
```bash
curl -X POST http://localhost:5000/v1/answer \
  -H "Content-Type: application/json" \
  -d '{
    "project": "sriov",
    "version": "4.16",
    "thread_slug": "test-thread-123",
    "message": "How do I configure SR-IOV?"
  }'
```

**Inject content:**
```bash
curl -X POST http://localhost:5000/v1/inject \
  -H "Content-Type: application/json" \
  -d '{
    "project": "sriov",
    "version": "4.16",
    "textContent": "Additional SR-IOV configuration notes..."
  }'
```

**Elaborate on something:**
```bash
curl -X POST http://localhost:5000/v1/elaborate \
  -H "Content-Type: application/json" \
  -d '{
    "thread_slug": "test-thread-123",
    "message": "Explain SR-IOV in more detail"
  }'
```

## Configuration

Environment variables (all optional except GEMINI_API_KEY):

```bash
# Required
export GEMINI_API_KEY="your-key"

# Abstention behavior (defaults shown)
export MIN_HITS=2                    # Min docs needed to answer
export SIMILARITY_CUTOFF=0.75        # Min similarity score
export CONFIDENCE_THRESHOLD=0.2      # Min confidence to answer
export TOP_K=5                       # Docs to retrieve
export TEMPERATURE=0.0               # LLM temperature

# Storage paths
export STORAGE_ROOT=./local-storage
export DELTA_ROOT=./local-delta
export INJECT_ROOT=./local-injected
export STATE_ROOT=./local-state
```

## Troubleshooting

**"No documents found"** - Make sure your `rag-data/{project}/{version}/` directories contain files.

**"GEMINI_API_KEY required"** - Set the environment variable before running.

**Server returns "I don't know" too often** - Lower `SIMILARITY_CUTOFF` or `CONFIDENCE_THRESHOLD`:
```bash
export SIMILARITY_CUTOFF=0.6
export CONFIDENCE_THRESHOLD=0.1
```

**Indexes not loading** - Check that `STORAGE_ROOT` points to where you built indexes.

## Next Steps

- See full API documentation in [README.md](README.md)
- Integrate with the Slack bot (see main project README)
- Tune abstention parameters for your use case


