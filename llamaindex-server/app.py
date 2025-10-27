#!/usr/bin/env python3
"""
Minimal LlamaIndex Flask server for Slack AI Assistant.
Supports: /v1/answer, /v1/elaborate, /v1/inject
"""
import os
import sys
import json
import uuid
from datetime import datetime
from pathlib import Path
from typing import Dict, Optional, List

from flask import Flask, request, jsonify
from llama_index.core import (
    VectorStoreIndex,
    StorageContext,
    load_index_from_storage,
    Document,
    Settings
)
from llama_index.embeddings.google_genai import GoogleGenAIEmbedding
from llama_index.core.postprocessor import SimilarityPostprocessor
from llama_index.core.schema import NodeWithScore
import google.generativeai as genai

app = Flask(__name__)

# Global registry
indexes: Dict[str, VectorStoreIndex] = {}  # base indexes
delta_indexes: Dict[str, VectorStoreIndex] = {}  # injected content indexes
threads: Dict[str, List[Dict]] = {}  # thread_slug -> [messages]

# Configuration from environment
GEMINI_API_KEY = os.environ.get("GEMINI_API_KEY")
DATA_ROOT = os.environ.get("DATA_ROOT", "/app/data")
STORAGE_ROOT = os.environ.get("STORAGE_ROOT", "/app/storage")
DELTA_ROOT = os.environ.get("DELTA_ROOT", "/app/storage-delta")
STATE_ROOT = os.environ.get("STATE_ROOT", "/app/state")
INJECT_ROOT = os.environ.get("INJECT_ROOT", "/app/injected")

# Abstention configuration (lowered defaults for better recall)
MIN_HITS = int(os.environ.get("MIN_HITS", "1"))
SIMILARITY_CUTOFF = float(os.environ.get("SIMILARITY_CUTOFF", "0.5"))
CONFIDENCE_THRESHOLD = float(os.environ.get("CONFIDENCE_THRESHOLD", "0.1"))
TOP_K = int(os.environ.get("TOP_K", "5"))
TEMPERATURE = float(os.environ.get("TEMPERATURE", "0.0"))

# Gemini model for chat (initialize once)
gemini_model = None


def get_gemini_model():
    """Lazy initialize Gemini model to avoid startup validation issues."""
    global gemini_model
    if gemini_model is None:
        genai.configure(api_key=GEMINI_API_KEY)
        # Use gemini-2.5-pro - latest stable pro model
        gemini_model = genai.GenerativeModel('models/gemini-2.5-pro')
    return gemini_model


def normalize_version(version):
    """Replace dots with -dot- for slug naming."""
    return version.replace(".", "-dot-")


def get_slug(project, version):
    """Generate slug for project/version."""
    if version:
        return f"{project}-{normalize_version(version)}"
    return project


def load_thread_memory(thread_slug: str) -> List[Dict]:
    """Load thread conversation history from disk."""
    thread_path = Path(STATE_ROOT) / "threads" / f"{thread_slug}.json"
    if thread_path.exists():
        with open(thread_path, 'r') as f:
            return json.load(f)
    return []


def save_thread_memory(thread_slug: str, messages: List[Dict]):
    """Persist thread conversation history to disk."""
    thread_dir = Path(STATE_ROOT) / "threads"
    thread_dir.mkdir(parents=True, exist_ok=True)
    thread_path = thread_dir / f"{thread_slug}.json"
    with open(thread_path, 'w') as f:
        json.dump(messages, f, indent=2)


def append_to_jsonl(project: str, version: str, text: str, metadata: Optional[Dict] = None):
    """Append injection to JSONL file."""
    inject_dir = Path(INJECT_ROOT) / project / normalize_version(version)
    inject_dir.mkdir(parents=True, exist_ok=True)
    
    today = datetime.now().strftime("%Y-%m-%d")
    jsonl_path = inject_dir / f"{today}.jsonl"
    
    record = {
        "id": str(uuid.uuid4()),
        "timestamp": datetime.now().isoformat(),
        "text": text,
        "metadata": metadata or {}
    }
    
    with open(jsonl_path, 'a') as f:
        f.write(json.dumps(record) + '\n')


def load_injected_documents(project: str, version: str) -> List[Document]:
    """Load all injected documents from JSONL files for a project/version."""
    inject_dir = Path(INJECT_ROOT) / project / normalize_version(version)
    if not inject_dir.exists():
        return []
    
    documents = []
    for jsonl_file in inject_dir.glob("*.jsonl"):
        with open(jsonl_file, 'r') as f:
            for line in f:
                if line.strip():
                    record = json.loads(line)
                    doc = Document(
                        text=record["text"],
                        metadata=record.get("metadata", {}),
                        doc_id=record.get("id")
                    )
                    documents.append(doc)
    return documents


def load_or_create_delta_index(slug: str) -> Optional[VectorStoreIndex]:
    """Load or create delta index for a project/version slug."""
    delta_path = Path(DELTA_ROOT) / slug
    
    # Try loading existing persisted delta index
    if delta_path.exists():
        try:
            storage_context = StorageContext.from_defaults(persist_dir=str(delta_path))
            return load_index_from_storage(storage_context)
        except Exception as e:
            print(f"Warning: Could not load delta index from {delta_path}: {e}")
    
    # Try building from JSONL if available
    project, version_slug = slug.split("-", 1)
    version = version_slug.replace("-dot-", ".")
    documents = load_injected_documents(project, version)
    
    if documents:
        print(f"Building delta index for {slug} from {len(documents)} injected documents")
        index = VectorStoreIndex.from_documents(documents)
        delta_path.mkdir(parents=True, exist_ok=True)
        index.storage_context.persist(persist_dir=str(delta_path))
        return index
    
    return None


def retrieve_with_confidence(base_index: VectorStoreIndex, 
                              delta_index: Optional[VectorStoreIndex],
                              query: str) -> tuple[List[NodeWithScore], bool]:
    """
    Retrieve nodes and determine if we have sufficient confidence to answer.
    Returns (nodes, should_answer).
    """
    all_nodes = []
    
    # Retrieve from base index
    base_retriever = base_index.as_retriever(similarity_top_k=TOP_K)
    base_nodes = base_retriever.retrieve(query)
    all_nodes.extend(base_nodes)
    print(f"Retrieved {len(base_nodes)} nodes from base index")
    
    # Retrieve from delta index if available
    if delta_index:
        delta_retriever = delta_index.as_retriever(similarity_top_k=TOP_K)
        delta_nodes = delta_retriever.retrieve(query)
        all_nodes.extend(delta_nodes)
        print(f"Retrieved {len(delta_nodes)} nodes from delta index")
    
    # Debug: print scores before filtering
    if all_nodes:
        scores = [f"{node.score:.3f}" for node in all_nodes if node.score is not None]
        print(f"Scores before filtering: {scores}")
    
    # Apply similarity cutoff
    postprocessor = SimilarityPostprocessor(similarity_cutoff=SIMILARITY_CUTOFF)
    filtered_nodes = postprocessor.postprocess_nodes(all_nodes)
    print(f"After similarity cutoff ({SIMILARITY_CUTOFF}): {len(filtered_nodes)} nodes")
    
    # Check abstention conditions
    if len(filtered_nodes) < MIN_HITS:
        print(f"Abstaining: fewer than {MIN_HITS} hits")
        return filtered_nodes, False
    
    # Calculate aggregate confidence (mean of scores)
    if filtered_nodes:
        scores = [node.score for node in filtered_nodes if node.score is not None]
        if scores:
            mean_score = sum(scores) / len(scores)
            print(f"Mean confidence score: {mean_score:.3f} (threshold: {CONFIDENCE_THRESHOLD})")
            if mean_score < CONFIDENCE_THRESHOLD:
                print(f"Abstaining: confidence below threshold")
                return filtered_nodes, False
    
    print(f"Proceeding with answer using {len(filtered_nodes)} nodes")
    return filtered_nodes, True


def generate_with_gemini(prompt: str) -> str:
    """Generate response using Gemini model."""
    model = get_gemini_model()
    response = model.generate_content(
        prompt,
        generation_config=genai.types.GenerationConfig(
            temperature=TEMPERATURE,
        )
    )
    return response.text


def build_indexes_if_needed():
    """Build indexes from rag-data if they don't exist yet."""
    storage_path = Path(STORAGE_ROOT)
    data_path = Path(DATA_ROOT)
    
    # Check if we have any indexes already
    existing_indexes = list(storage_path.glob("*")) if storage_path.exists() else []
    
    if not existing_indexes and data_path.exists():
        print("No indexes found - building from rag-data on first startup...")
        print("This may take a few minutes depending on the amount of data.")
        
        # Import and run build_indexes
        import subprocess
        result = subprocess.run(
            ["python", "build_indexes.py", "--data", DATA_ROOT, "--out", STORAGE_ROOT],
            capture_output=True,
            text=True
        )
        
        if result.returncode != 0:
            print(f"Error building indexes: {result.stderr}")
        else:
            print("Indexes built successfully!")
            print(result.stdout)


def initialize_server():
    """Load base and delta indexes on startup."""
    if not GEMINI_API_KEY:
        print("Error: GEMINI_API_KEY environment variable is required")
        sys.exit(1)
    
    # Configure embeddings (we use this, not the LLM class)
    Settings.embed_model = GoogleGenAIEmbedding(
        model_name="models/text-embedding-004",
        api_key=GEMINI_API_KEY
    )
    
    # Create necessary directories
    for dir_path in [STORAGE_ROOT, DELTA_ROOT, STATE_ROOT, INJECT_ROOT]:
        Path(dir_path).mkdir(parents=True, exist_ok=True)
    
    # Build indexes if they don't exist (first startup)
    build_indexes_if_needed()
    
    # Load base indexes
    storage_path = Path(STORAGE_ROOT)
    if storage_path.exists():
        for index_dir in storage_path.iterdir():
            if index_dir.is_dir():
                slug = index_dir.name
                try:
                    storage_context = StorageContext.from_defaults(persist_dir=str(index_dir))
                    index = load_index_from_storage(storage_context)
                    indexes[slug] = index
                    print(f"Loaded base index: {slug}")
                except Exception as e:
                    print(f"Warning: Could not load index {slug}: {e}")
    
    # Load delta indexes
    delta_path = Path(DELTA_ROOT)
    if delta_path.exists():
        for index_dir in delta_path.iterdir():
            if index_dir.is_dir():
                slug = index_dir.name
                delta_index = load_or_create_delta_index(slug)
                if delta_index:
                    delta_indexes[slug] = delta_index
                    print(f"Loaded delta index: {slug}")
    
    print(f"Server initialized with {len(indexes)} base indexes and {len(delta_indexes)} delta indexes")


@app.route('/health', methods=['GET'])
def health():
    """Health check endpoint."""
    return jsonify({"status": "ok", "base_indexes": len(indexes), "delta_indexes": len(delta_indexes)})


@app.route('/v1/answer', methods=['POST'])
def answer():
    """
    Answer a question using RAG over base + delta indexes.
    Body: { project, version, thread_slug, message }
    Returns: { textResponse }
    """
    data = request.json
    project = data.get('project')
    version = data.get('version')
    thread_slug = data.get('thread_slug')
    message = data.get('message')
    
    if not all([project, version, thread_slug, message]):
        return jsonify({"error": "Missing required fields"}), 400
    
    slug = get_slug(project, version)
    
    # Get indexes
    base_index = indexes.get(slug)
    if not base_index:
        return jsonify({"error": f"No index found for {project}/{version}"}), 404
    
    delta_index = delta_indexes.get(slug)
    
    # Retrieve with confidence gating
    nodes, should_answer = retrieve_with_confidence(base_index, delta_index, message)
    
    if not should_answer:
        response_text = "I don't know."
    else:
        # Build context from nodes
        context = "\n\n".join([node.node.get_content() for node in nodes])
        
        # Generate response with Gemini directly
        prompt = f"""You are a helpful technical assistant with expertise in Kubernetes and cloud-native technologies.

Use the provided context as your PRIMARY source of information. When the user asks for examples or configurations:
1. Start with what's provided in the context
2. Use your knowledge to complete and enhance the example to make it fully functional
3. Ensure all parts of your example are consistent (matching labels, IPs, names, etc.)
4. Provide complete, working configurations that the user can directly use

If the context doesn't contain relevant information to answer the question at all, respond with: "I don't know."

Context:
{context}

Question: {message}

Answer:"""
        
        response_text = generate_with_gemini(prompt)
    
    # Update thread memory
    thread_messages = load_thread_memory(thread_slug)
    thread_messages.append({"role": "user", "content": message})
    thread_messages.append({"role": "assistant", "content": response_text})
    save_thread_memory(thread_slug, thread_messages)
    threads[thread_slug] = thread_messages
    
    return jsonify({"textResponse": response_text})


@app.route('/v1/elaborate', methods=['POST'])
def elaborate():
    """
    Elaborate/summarize a message in a more readable fashion using pure chat (no retrieval/RAG).
    This is used when a human provides detailed information and wants it reformatted.
    Body: { thread_slug, message }
    Returns: { textResponse }
    """
    data = request.json
    thread_slug = data.get('thread_slug')
    message = data.get('message')
    
    if not all([thread_slug, message]):
        return jsonify({"error": "Missing required fields"}), 400
    
    # Load thread memory
    thread_messages = load_thread_memory(thread_slug)
    
    # Build conversation context
    conversation = "\n".join([
        f"{msg['role']}: {msg['content']}" for msg in thread_messages
    ])
    
    # Generate with Gemini - pure chat, no retrieval
    if conversation:
        prompt = (
            f"Previous conversation:\n{conversation}\n\n"
            f"Please take the following content and reformat it in a clear, readable, and well-organized way. "
            f"Summarize key points, improve structure, and make it easier to understand:\n\n"
            f"{message}\n\n"
            f"Reformatted version:"
        )
    else:
        prompt = (
            f"Please take the following content and reformat it in a clear, readable, and well-organized way. "
            f"Summarize key points, improve structure, and make it easier to understand:\n\n"
            f"{message}\n\n"
            f"Reformatted version:"
        )
    
    response_text = generate_with_gemini(prompt)
    
    # Update thread memory
    thread_messages.append({"role": "user", "content": message})
    thread_messages.append({"role": "assistant", "content": response_text})
    save_thread_memory(thread_slug, thread_messages)
    threads[thread_slug] = thread_messages
    
    return jsonify({"textResponse": response_text})


@app.route('/v1/inject', methods=['POST'])
def inject():
    """
    Inject content into delta index for a project/version.
    Body: { project, version, textContent, metadata? }
    Returns: 200 OK
    """
    data = request.json
    project = data.get('project')
    version = data.get('version')
    text_content = data.get('textContent')
    metadata = data.get('metadata', {})
    
    if not all([project, version, text_content]):
        return jsonify({"error": "Missing required fields"}), 400
    
    slug = get_slug(project, version)
    
    # Append to JSONL
    append_to_jsonl(project, version, text_content, metadata)
    
    # Create document
    doc = Document(
        text=text_content,
        metadata=metadata,
        doc_id=str(uuid.uuid4())
    )
    
    # Update or create delta index
    delta_index = delta_indexes.get(slug)
    if delta_index:
        # Insert into existing index
        delta_index.insert(doc)
    else:
        # Create new delta index
        delta_index = VectorStoreIndex.from_documents([doc])
        delta_indexes[slug] = delta_index
    
    # Persist delta index
    delta_path = Path(DELTA_ROOT) / slug
    delta_path.mkdir(parents=True, exist_ok=True)
    delta_index.storage_context.persist(persist_dir=str(delta_path))
    
    return jsonify({"status": "ok"}), 200


if __name__ == '__main__':
    initialize_server()
    app.run(host='0.0.0.0', port=5000, debug=False)
