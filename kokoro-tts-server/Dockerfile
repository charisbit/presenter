# Kokoro TTS Server - 82M Parameter Open-Weight TTS Model
# Optimized for Apple Silicon (ARM64) with Japanese support

# Use Python 3.11 base image for ARM64 (Apple Silicon)
FROM --platform=linux/arm64 python:3.11-slim

# Set working directory
WORKDIR /app

# Install system dependencies for Kokoro TTS and audio processing
RUN apt-get update && apt-get install -y \
    build-essential \
    cmake \
    git \
    wget \
    curl \
    ffmpeg \
    libsndfile1 \
    libportaudio2 \
    portaudio19-dev \
    espeak-ng \
    && rm -rf /var/lib/apt/lists/*

# Install Python dependencies first
RUN pip install --no-cache-dir --upgrade pip

# Install core dependencies
RUN pip install --no-cache-dir \
    numpy \
    torch \
    torchaudio \
    transformers \
    fastapi \
    uvicorn \
    pydantic \
    soundfile \
    scipy \
    requests \
    aiofiles

# Install Kokoro TTS and required Japanese language support
RUN pip install --no-cache-dir \
    kokoro>=0.9.4 \
    'misaki[ja]'

# Download Japanese language model for Kokoro
RUN python -m unidic download

# Create necessary directories
RUN mkdir -p /app/models /app/cache /app/logs

# Copy application files
COPY requirements.txt* ./
COPY app.py ./
COPY models.py ./
COPY config.py ./

# Install additional requirements if present
RUN if [ -f requirements.txt ]; then pip install --no-cache-dir -r requirements.txt; fi

# Expose port for TTS API
EXPOSE 8882

# Health check endpoint
HEALTHCHECK --interval=30s --timeout=15s --start-period=120s --retries=3 \
  CMD curl -f http://localhost:8882/health || exit 1

# Set environment variables
ENV PYTHONPATH=/app
ENV MODEL_CACHE_DIR=/app/models
ENV AUDIO_CACHE_DIR=/app/cache
ENV LANG=C.UTF-8
ENV LC_ALL=C.UTF-8

# Run the Kokoro TTS server
CMD ["python", "app.py"]