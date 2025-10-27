#!/usr/bin/env python3
"""
Quick test to verify LlamaIndex dependencies are installed correctly.
Run this after: pip install -r requirements.txt
"""

def test_imports():
    """Test that all required packages can be imported."""
    print("Testing imports...")
    
    try:
        import flask
        print("✓ Flask imported")
    except ImportError as e:
        print(f"✗ Flask import failed: {e}")
        return False
    
    try:
        from llama_index.core import VectorStoreIndex, Settings
        print("✓ llama-index-core imported")
    except ImportError as e:
        print(f"✗ llama-index-core import failed: {e}")
        return False
    
    try:
        from llama_index.llms.google_genai import GoogleGenAI
        print("✓ llama-index-llms-google-genai imported")
    except ImportError as e:
        print(f"✗ llama-index-llms-google-genai import failed: {e}")
        return False
    
    try:
        from llama_index.embeddings.google_genai import GoogleGenAIEmbedding
        print("✓ llama-index-embeddings-google-genai imported")
    except ImportError as e:
        print(f"✗ llama-index-embeddings-google-genai import failed: {e}")
        return False
    
    try:
        import google.generativeai
        print("✓ google-generativeai imported")
    except ImportError as e:
        print(f"✗ google-generativeai import failed: {e}")
        return False
    
    return True


if __name__ == "__main__":
    print("\nVerifying LlamaIndex server dependencies...\n")
    
    if test_imports():
        print("\n✓ All dependencies installed correctly!")
        print("\nNext steps:")
        print("1. export GEMINI_API_KEY=your-key")
        print("2. python build_indexes.py --data ../rag-data --out ./local-storage")
        print("3. STORAGE_ROOT=./local-storage FLASK_APP=app.py flask run")
    else:
        print("\n✗ Some dependencies failed to import")
        print("Try: pip install -r requirements.txt")

