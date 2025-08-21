#!/usr/bin/env python3
"""
Kokoro TTS Server
82M Parameter Open-Weight TTS Model with Japanese Support
Based on the Kokoro pipeline code provided
"""

import os
import logging
import hashlib
import asyncio
from pathlib import Path
from typing import Optional, Dict, Any, List, Tuple
import uvicorn
from fastapi import FastAPI, HTTPException, Response
from fastapi.responses import FileResponse
from pydantic import BaseModel
import soundfile as sf
import numpy as np

# Kokoro TTS imports
from kokoro import KPipeline
import torch

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Configuration
MODEL_CACHE_DIR = os.getenv("MODEL_CACHE_DIR", "/app/models")
AUDIO_CACHE_DIR = os.getenv("AUDIO_CACHE_DIR", "/app/cache")
PORT = int(os.getenv("PORT", "8882"))

# Create directories
Path(MODEL_CACHE_DIR).mkdir(parents=True, exist_ok=True)
Path(AUDIO_CACHE_DIR).mkdir(parents=True, exist_ok=True)

app = FastAPI(
    title="Kokoro TTS Server",
    description="82M Parameter Open-Weight TTS Model with Japanese Support",
    version="1.0.0"
)

class TTSRequest(BaseModel):
    text: str
    language: str = "en"  # Default to English
    voice: str = "af_heart"  # Default Kokoro voice
    format: str = "wav"
    speed: float = 1.0

class TTSResponse(BaseModel):
    audio_url: str
    duration: float
    cache_hit: bool
    model: str = "kokoro-82m"
    voice_used: str
    segments: int

class KokoroTTS:
    """Kokoro TTS Engine for Multi-language Support"""
    
    def __init__(self):
        self.pipeline = None
        self.model_loaded = False
        self.sample_rate = 24000  # Kokoro uses 24kHz
        self.supported_languages = {
            'en': 'a',  # American English
            'ja': 'j',  # Japanese
            'es': 'e',  # Spanish
            'fr': 'f',  # French
            'hi': 'h',  # Hindi
            'it': 'i',  # Italian
            'pt': 'p',  # Brazilian Portuguese
            'zh': 'z'   # Mandarin Chinese
        }
        
    async def load_model(self, language: str = 'en'):
        """Load Kokoro pipeline for specified language"""
        try:
            lang_code = self.supported_languages.get(language, 'a')  # Default to English
            logger.info(f"Loading Kokoro TTS model for language: {language} (code: {lang_code})")
            
            # Initialize Kokoro pipeline
            self.pipeline = KPipeline(lang_code=lang_code)
            self.model_loaded = True
            
            logger.info(f"Kokoro TTS model loaded successfully for {language}")
        except Exception as e:
            logger.error(f"Failed to load Kokoro TTS model: {e}")
            raise
    
    async def synthesize(self, text: str, voice: str = "af_heart", speed: float = 1.0, language: str = "en") -> Tuple[np.ndarray, int]:
        """Synthesize speech from text using Kokoro TTS"""
        # Always reload model for the requested language (no model caching)
        await self.load_model(language)
        
        try:
            logger.info(f"Generating audio for text: '{text[:50]}...' with voice: {voice}")
            
            # Generate audio using Kokoro pipeline
            # Based on the provided code pattern
            generator = self.pipeline(
                text, 
                voice=voice,
                speed=speed, 
                split_pattern=r'\n+'
            )
            
            # Combine all audio segments
            audio_segments = []
            segment_count = 0
            
            for i, (gs, ps, audio) in enumerate(generator):
                logger.info(f"Segment {i}: graphemes='{gs[:30]}...', phonemes='{ps[:30]}...'")
                audio_segments.append(audio)
                segment_count += 1
                
                # Log phonemes for debugging (this shows the actual speech processing)
                logger.debug(f"Phonemes for segment {i}: {ps}")
            
            if not audio_segments:
                raise Exception("No audio segments generated")
            
            # Concatenate all audio segments
            if len(audio_segments) == 1:
                final_audio = audio_segments[0]
            else:
                final_audio = np.concatenate(audio_segments)
            
            logger.info(f"Generated {len(final_audio)} samples across {segment_count} segments")
            return final_audio, segment_count
            
        except Exception as e:
            logger.error(f"Kokoro TTS synthesis failed: {e}")
            raise

# Global TTS engine instance
tts_engine = KokoroTTS()

def generate_cache_key(text: str, voice: str, speed: float, language: str) -> str:
    """Generate cache key for audio file"""
    content = f"kokoro:{text}:{voice}:{speed}:{language}"
    return hashlib.md5(content.encode()).hexdigest()

@app.on_event("startup")
async def startup_event():
    """Initialize TTS engine on startup"""
    logger.info("Starting Kokoro TTS Server...")
    try:
        # Pre-load English model by default
        await tts_engine.load_model('en')
        logger.info("Kokoro TTS Server ready")
    except Exception as e:
        logger.error(f"Failed to initialize TTS engine: {e}")

@app.get("/health")
async def health_check():
    """Health check endpoint"""
    return {
        "status": "healthy",
        "service": "kokoro-tts",
        "model_loaded": tts_engine.model_loaded,
        "default_language": "en",
        "platform": "arm64",
        "supported_languages": list(tts_engine.supported_languages.keys()),
        "sample_rate": tts_engine.sample_rate
    }

@app.get("/models")
async def list_models():
    """List available TTS models and voices"""
    return {
        "model": "kokoro-82m",
        "parameters": "82 million",
        "license": "Apache 2.0",
        "default_language": "en",
        "supported_languages": tts_engine.supported_languages,
        "available_voices": [
            "af_heart",  # Default voice
            # Additional voices can be added based on Kokoro's voice library
        ]
    }

@app.post("/api/tts", response_model=TTSResponse)
async def synthesize_speech(request: TTSRequest):
    """Synthesize speech from text using Kokoro TTS"""
    try:
        # Validate language support
        if request.language not in tts_engine.supported_languages:
            raise HTTPException(
                status_code=400, 
                detail=f"Language '{request.language}' not supported. Supported: {list(tts_engine.supported_languages.keys())}"
            )
        
        # Generate cache key
        cache_key = generate_cache_key(request.text, request.voice, request.speed, request.language)
        cache_file = Path(AUDIO_CACHE_DIR) / f"{cache_key}.{request.format}"
        
        cache_hit = cache_file.exists()
        segments = 0
        
        if not cache_hit:
            # Generate audio
            logger.info(f"Generating audio for: '{request.text[:50]}...'")
            audio_data, segments = await tts_engine.synthesize(
                text=request.text,
                voice=request.voice,
                speed=request.speed,
                language=request.language
            )
            
            # Save to cache
            sf.write(str(cache_file), audio_data, tts_engine.sample_rate)
            logger.info(f"Audio saved to cache: {cache_file}")
        else:
            # For cached files, we don't know the segment count, so set to 1
            segments = 1
        
        # Calculate duration
        info = sf.info(str(cache_file))
        duration = info.duration
        
        return TTSResponse(
            audio_url=f"/audio/{cache_key}.{request.format}",
            duration=duration,
            cache_hit=cache_hit,
            voice_used=request.voice,
            segments=segments
        )
        
    except Exception as e:
        logger.error(f"TTS synthesis failed: {e}")
        raise HTTPException(status_code=500, detail=f"TTS synthesis failed: {str(e)}")

@app.get("/audio/{filename}")
async def get_audio_file(filename: str):
    """Serve audio files"""
    file_path = Path(AUDIO_CACHE_DIR) / filename
    
    if not file_path.exists():
        raise HTTPException(status_code=404, detail="Audio file not found")
    
    return FileResponse(
        path=str(file_path),
        media_type="audio/wav",
        headers={"Cache-Control": "public, max-age=3600"}
    )

@app.get("/")
async def root():
    """Root endpoint with service information"""
    return {
        "service": "Kokoro TTS Server",
        "model": "kokoro-82m",
        "version": "1.0.0",
        "platform": "Apple Silicon (ARM64)",
        "description": "82M Parameter Open-Weight TTS Model",
        "license": "Apache 2.0",
        "endpoints": {
            "health": "/health",
            "models": "/models", 
            "synthesize": "/api/tts",
            "audio": "/audio/{filename}"
        }
    }

if __name__ == "__main__":
    # Run the server
    logger.info(f"Starting Kokoro TTS Server on port {PORT}")
    uvicorn.run(
        "app:app",
        host="0.0.0.0",
        port=PORT,
        log_level="info",
        access_log=True
    )