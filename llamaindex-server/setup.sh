#!/bin/bash
# Setup script using uv for faster, better dependency resolution

set -e

echo "🚀 Setting up LlamaIndex server with uv..."

# Check if uv is installed
if ! command -v uv &> /dev/null; then
    echo "📦 Installing uv..."
    curl -LsSf https://astral.sh/uv/install.sh | sh
    export PATH="$HOME/.cargo/bin:$PATH"
fi

echo "✓ uv is installed"

# Sync dependencies with uv (creates venv and installs)
echo "📥 Syncing dependencies with uv..."
uv sync

echo ""
echo "✅ Setup complete!"
echo ""
echo "Next steps:"
echo "1. Set your API key: export GEMINI_API_KEY=your-key"
echo "2. Build indexes: uv run build_indexes.py --data ../rag-data --out ./local-storage"
echo "3. Run server: uv run --env-file .env app.py"
echo ""
echo "Or verify install: uv run verify_install.py"

