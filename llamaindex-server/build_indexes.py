#!/usr/bin/env python3
"""
Build base vector indexes from rag-data directory structure.
Reads from DATA_ROOT/{project}/{version}/ and writes indexes to STORAGE_ROOT/{project}-{version}/
"""
import os
import sys
import argparse
from pathlib import Path
from llama_index.core import VectorStoreIndex, SimpleDirectoryReader, StorageContext
from llama_index.embeddings.google_genai import GoogleGenAIEmbedding
from llama_index.core import Settings


def normalize_version(version):
    """Replace dots with -dot- for slug naming."""
    return version.replace(".", "-dot-")


def build_indexes(data_root, storage_root):
    """Build indexes for all project/version combinations found in data_root."""
    data_path = Path(data_root)
    storage_path = Path(storage_root)
    storage_path.mkdir(parents=True, exist_ok=True)
    
    # Configure Gemini embeddings
    gemini_api_key = os.environ.get("GEMINI_API_KEY")
    if not gemini_api_key:
        print("Error: GEMINI_API_KEY environment variable is required")
        sys.exit(1)
    
    embed_model = GoogleGenAIEmbedding(
        model_name="models/text-embedding-004",
        api_key=gemini_api_key
    )
    Settings.embed_model = embed_model
    
    # Walk data_root to find project/version directories
    if not data_path.exists():
        print(f"Warning: Data root {data_root} does not exist. No indexes built.")
        return
    
    projects = [p for p in data_path.iterdir() if p.is_dir()]
    for project_dir in projects:
        project_name = project_dir.name
        versions = [v for v in project_dir.iterdir() if v.is_dir()]
        
        for version_dir in versions:
            version_name = version_dir.name
            slug = f"{project_name}-{normalize_version(version_name)}"
            
            print(f"Building index for {project_name}/{version_name} → {slug}...")
            
            # Load documents
            try:
                documents = SimpleDirectoryReader(str(version_dir)).load_data()
                if not documents:
                    print(f"  Warning: No documents found in {version_dir}")
                    continue
                
                print(f"  Loaded {len(documents)} documents")
                
                # Create index
                index = VectorStoreIndex.from_documents(documents)
                
                # Persist index
                index_path = storage_path / slug
                index.storage_context.persist(persist_dir=str(index_path))
                print(f"  Index saved to {index_path}")
                
            except Exception as e:
                print(f"  Error building index for {slug}: {e}")
                continue


def main():
    parser = argparse.ArgumentParser(description="Build LlamaIndex vector indexes from rag-data")
    parser.add_argument("--data", default=os.environ.get("DATA_ROOT", "./rag-data"),
                        help="Root directory containing project/version subdirectories")
    parser.add_argument("--out", default=os.environ.get("STORAGE_ROOT", "./storage"),
                        help="Output directory for built indexes")
    
    args = parser.parse_args()
    
    print(f"Building indexes from {args.data} → {args.out}")
    build_indexes(args.data, args.out)
    print("Index building complete.")


if __name__ == "__main__":
    main()


