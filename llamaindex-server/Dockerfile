# Dockerfile for LlamaIndex server
# NOTE: Indexes should be pre-built locally and copied in, OR built on first container startup
# DO NOT pass GEMINI_API_KEY as a build arg - it should only be provided at runtime!

FROM python:3.11-slim

WORKDIR /app

# Install dependencies
COPY llamaindex-server/requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy application code and build script
COPY llamaindex-server/app.py .
COPY llamaindex-server/build_indexes.py .

# Copy pre-built indexes if they exist (optional - can build on startup)
# These should be built locally with: python build_indexes.py --data ./rag-data --out ./storage
COPY rag-data /app/data

# Create runtime directories (indexes will be built here on first startup if needed)
RUN mkdir -p /app/storage /app/state /app/storage-delta /app/injected

# Expose Flask port
EXPOSE 5000

# Set environment defaults
ENV STORAGE_ROOT=/app/storage \
    DELTA_ROOT=/app/storage-delta \
    STATE_ROOT=/app/state \
    INJECT_ROOT=/app/injected \
    DATA_ROOT=/app/data

# IMPORTANT: GEMINI_API_KEY must be provided at runtime via -e flag
# Example: docker run -e GEMINI_API_KEY=your-key ...

# Run Flask app
CMD ["python", "app.py"]


