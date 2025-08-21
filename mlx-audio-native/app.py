#!/usr/bin/env python3
"""
Native MLX-Audio TTS Server for macOS
High-quality Text-to-Speech using MLX-Audio framework with official Kokoro support
"""

import os
import logging
import hashlib
import asyncio
from pathlib import Path
from typing import Optional, Dict, Any, Tuple
from contextlib import asynccontextmanager
import uvicorn
from fastapi import FastAPI, HTTPException, Response
from fastapi.responses import FileResponse
from pydantic import BaseModel
import soundfile as sf
import numpy as np

# Configure logging first
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Global initialization state
MLX_AUDIO_AVAILABLE = False
MLX_AUDIO_KOKORO_AVAILABLE = False
MLX_AUDIO_FULL_AVAILABLE = False

# Auto-download UniDic for Japanese text processing
async def ensure_unidic_installed():
    """Ensure UniDic dictionary is installed for Japanese text processing"""
    try:
        import unidic
        # Check if UniDic is already downloaded by trying to access DICDIR
        dicdir = unidic.DICDIR
        if os.path.exists(dicdir):
            logger.info("✅ UniDic dictionary already available")
            return True
        else:
            logger.info("📥 Downloading UniDic dictionary for Japanese support...")
            from unidic.download import download_version
            # Run in thread to avoid blocking the event loop
            import asyncio
            await asyncio.get_event_loop().run_in_executor(None, download_version)
            logger.info("✅ UniDic dictionary downloaded successfully")
            return True
    except ImportError:
        logger.warning("⚠️ UniDic not available, Japanese text processing may be limited")
        return False
    except Exception as e:
        logger.error(f"❌ Failed to download UniDic: {e}")
        return False

async def initialize_mlx_audio():
    """Initialize MLX-Audio imports and check availability"""
    global MLX_AUDIO_AVAILABLE, MLX_AUDIO_KOKORO_AVAILABLE, MLX_AUDIO_FULL_AVAILABLE
    
    try:
        # Test core Kokoro functionality first
        from mlx_audio.tts.models.kokoro import KokoroPipeline
        MLX_AUDIO_KOKORO_AVAILABLE = True
        logger.info("✅ MLX-Audio Kokoro available")
        
        # Try additional imports but don't fail if they're not available
        try:
            from mlx_audio.tts.utils import load_model
            from mlx_audio.tts.generate import generate_audio
            MLX_AUDIO_FULL_AVAILABLE = True
            logger.info("✅ MLX-Audio full library loaded successfully")
        except ImportError as e:
            MLX_AUDIO_FULL_AVAILABLE = False
            logger.warning(f"⚠️ Some MLX-Audio functions not available: {e}")
        
        MLX_AUDIO_AVAILABLE = MLX_AUDIO_KOKORO_AVAILABLE
        return True
        
    except ImportError as e:
        MLX_AUDIO_AVAILABLE = False
        MLX_AUDIO_KOKORO_AVAILABLE = False
        MLX_AUDIO_FULL_AVAILABLE = False
        logger.error(f"❌ MLX-Audio not available: {e}")
        logger.error("Please install mlx-audio: pip install mlx-audio")
        return False

# Configuration
MODEL_CACHE_DIR = os.getenv("MODEL_CACHE_DIR", "./models")
AUDIO_CACHE_DIR = os.getenv("AUDIO_CACHE_DIR", "./cache")
PORT = int(os.getenv("PORT", "8881"))

# Create directories
Path(MODEL_CACHE_DIR).mkdir(parents=True, exist_ok=True)
Path(AUDIO_CACHE_DIR).mkdir(parents=True, exist_ok=True)

# Global TTS engine instance
tts_engine = None

@asynccontextmanager
async def lifespan(app: FastAPI):
    """Lifespan management for FastAPI"""
    global tts_engine
    
    # Startup
    logger.info("Starting Native MLX-Audio TTS Server...")
    
    # Initialize UniDic dictionary for Japanese text processing
    logger.info("Initializing Japanese text processing...")
    unidic_success = await ensure_unidic_installed()
    
    # Initialize MLX-Audio imports
    logger.info("Initializing MLX-Audio...")
    mlx_success = await initialize_mlx_audio()
    
    if mlx_success and MLX_AUDIO_AVAILABLE:
        logger.info("Creating TTS engine...")
        tts_engine = NativeMLXAudioTTS()
        if tts_engine.kokoro_model:
            logger.info("✅ MLX-Audio TTS Server ready with Kokoro support")
            if unidic_success:
                logger.info("✅ Full Japanese text processing support enabled")
            else:
                logger.warning("⚠️ Limited Japanese text processing (UniDic not available)")
        else:
            logger.error("❌ Failed to initialize Kokoro model")
            raise RuntimeError("Failed to initialize Kokoro model")
    else:
        logger.error("❌ MLX-Audio not available - cannot start server")
        raise RuntimeError("MLX-Audio not available")
    
    yield
    
    # Shutdown
    logger.info("Shutting down Native MLX-Audio TTS Server...")

app = FastAPI(
    title="Native MLX-Audio TTS Server",
    description="High-quality TTS using MLX-Audio framework with Kokoro support on macOS",
    version="1.0.0",
    lifespan=lifespan
)

class TTSRequest(BaseModel):
    text: str
    language: str = "ja"
    voice: str = "female"
    format: str = "wav"
    speed: float = 1.0

class TTSResponse(BaseModel):
    audio_url: str
    duration: float
    cache_hit: bool
    model: str = "mlx-audio-kokoro"
    voice_used: str

class NativeMLXAudioTTS:
    """Native MLX-Audio TTS Engine using official Kokoro support"""
    
    def __init__(self):
        self.sample_rate = 24000
        self.model_name = "prince-canuma/Kokoro-82M"
        self.kokoro_pipeline = None
        self.kokoro_model = None
        
        # Language mapping for Kokoro
        self.supported_languages = {
            'en': 'a',  # American English
            'ja': 'j',  # Japanese  
            'zh': 'z',  # Mandarin Chinese
        }
        
        # Voice mapping per language
        self.voice_mapping = {
            'en': {  # English voices
                "female": "af_heart",
                "male": "af_alloy",
                "af_heart": "af_heart",
                "af_alloy": "af_alloy",
                "af_bella": "af_bella",
                "af_nova": "af_nova"
            },
            'ja': {  # Japanese voices (use compatible ones)
                "female": "af_bella",  # Use more neutral voice for Japanese
                "male": "af_alloy",
                "af_heart": "af_bella",  # Map af_heart to compatible voice
                "af_bella": "af_bella",
                "af_alloy": "af_alloy"
            },
            'zh': {  # Chinese voices
                "female": "af_nicole",
                "male": "af_alloy",
                "af_heart": "af_nicole",
                "af_nicole": "af_nicole",
                "af_alloy": "af_alloy"
            }
        }
        
        if MLX_AUDIO_KOKORO_AVAILABLE:
            self._initialize_kokoro()
        else:
            logger.error("MLX-Audio Kokoro not available - cannot initialize")
    
    def _initialize_kokoro(self):
        """Initialize MLX-Audio Kokoro TTS model with proper configuration"""
        try:
            logger.info(f"Loading MLX-Audio Kokoro model: {self.model_name}")
            
            # Import required classes
            from mlx_audio.tts.models.kokoro import Model, ModelConfig
            from huggingface_hub import hf_hub_download
            import json
            import mlx.core as mx
            
            # Download and load the config from Hugging Face
            logger.info("Downloading config from Hugging Face...")
            config_path = hf_hub_download(self.model_name, 'config.json')
            
            with open(config_path, 'r') as f:
                config_dict = json.load(f)
            
            logger.info("✅ Config loaded successfully")
            
            # Create ModelConfig from the loaded dictionary
            config = ModelConfig(**config_dict)
            logger.info("✅ ModelConfig created successfully")
            
            # Create the Kokoro model with the proper config
            self.kokoro_model = Model(config, repo_id=self.model_name)
            
            # Ensure model parameters are properly loaded and evaluated
            logger.info("📥 Loading and evaluating model parameters...")
            mx.eval(self.kokoro_model.parameters())
            logger.info("✅ Model parameters evaluated successfully")
            
            # Set model to evaluation mode
            self.kokoro_model.eval()
            logger.info("✅ Kokoro model created successfully")
            
            # We don't need the pipeline, just use the model directly
            self.kokoro_pipeline = None
            
            logger.info("✅ MLX-Audio Kokoro model initialized successfully")
            
        except Exception as e:
            logger.error(f"❌ Failed to initialize MLX-Audio Kokoro model: {e}")
            logger.error(f"Error details: {str(e)}")
            self.kokoro_model = None
            self.kokoro_pipeline = None
    
    def get_voice_for_language(self, language: str, voice_preference: str) -> str:
        """Map voice preference to Kokoro voice ID based on language"""
        lang_voices = self.voice_mapping.get(language, self.voice_mapping['en'])
        return lang_voices.get(voice_preference, lang_voices.get("female", "af_bella"))
    
    def get_lang_code(self, language: str) -> str:
        """Get Kokoro language code"""
        return self.supported_languages.get(language, 'j')  # Default to Japanese
    
    async def synthesize(self, text: str, voice: str = "female", speed: float = 1.0, language: str = "ja") -> Tuple[np.ndarray, float]:
        """Synthesize speech using MLX-Audio Kokoro model directly"""
        if not MLX_AUDIO_AVAILABLE or not self.kokoro_model:
            raise Exception("MLX-Audio not available or not initialized")
        
        try:
            logger.info(f"Generating MLX-Audio Kokoro for text: '{text[:50]}...' in {language}")
            
            # Get language code and voice
            lang_code = self.get_lang_code(language)
            kokoro_voice = self.get_voice_for_language(language, voice)
            
            # Use the model's generate method directly
            logger.info(f"Using voice: {kokoro_voice}, lang_code: {lang_code}, speed: {speed}")
            
            # Try a different approach - use the KokoroPipeline directly
            try:
                from mlx_audio.tts.models.kokoro import KokoroPipeline
                import mlx.core as mx
                
                # Create pipeline for the specified language
                pipeline = KokoroPipeline(lang_code)
                
                # Generate audio using pipeline
                logger.info("🔄 Using KokoroPipeline directly for better compatibility")
                audio_output = pipeline.generate(
                    text=text,
                    voice=kokoro_voice,
                    speed=speed
                )
                
                # If that fails, fall back to the model's generate method
                if audio_output is None:
                    raise Exception("KokoroPipeline returned None")
                    
            except Exception as pipeline_error:
                logger.warning(f"⚠️ KokoroPipeline failed ({pipeline_error}), falling back to model.generate")
                
                # Fallback to model's generate method
                audio_output = self.kokoro_model.generate(
                    text=text,
                    voice=kokoro_voice,
                    speed=speed,
                    lang_code=lang_code,
                    split_pattern=r'\n+'
                )
            
            # Extract audio from the generator output
            audio_segments = []
            segment_count = 0
            
            for segment in audio_output:
                segment_audio = None
                if hasattr(segment, 'audio'):
                    # If it's an output object with audio attribute
                    segment_audio = segment.audio
                elif isinstance(segment, np.ndarray):
                    # If it's directly numpy array
                    segment_audio = segment
                else:
                    # Try to convert to numpy array
                    segment_audio = np.array(segment)
                
                # Convert MLX array to numpy if needed
                if hasattr(segment_audio, 'numpy'):
                    segment_audio = segment_audio.numpy()
                elif hasattr(segment_audio, '__array__'):
                    segment_audio = np.asarray(segment_audio)
                
                # Check for NaN values and fix them
                if np.any(np.isnan(segment_audio)):
                    logger.warning(f"⚠️ Found NaN values in segment {segment_count}, replacing with zeros")
                    segment_audio = np.nan_to_num(segment_audio, nan=0.0)
                
                # Check for infinite values and fix them
                if np.any(np.isinf(segment_audio)):
                    logger.warning(f"⚠️ Found infinite values in segment {segment_count}, clamping")
                    segment_audio = np.nan_to_num(segment_audio, posinf=1.0, neginf=-1.0)
                
                # Normalize audio to prevent clipping
                if np.max(np.abs(segment_audio)) > 1.0:
                    logger.info(f"📊 Normalizing segment {segment_count} to prevent clipping")
                    segment_audio = segment_audio / np.max(np.abs(segment_audio))
                
                # Ensure float32 dtype
                segment_audio = segment_audio.astype(np.float32)
                
                audio_segments.append(segment_audio)
                segment_count += 1
                
                logger.info(f"📊 Segment {segment_count}: shape={segment_audio.shape}, min={np.min(segment_audio):.6f}, max={np.max(segment_audio):.6f}, mean={np.mean(segment_audio):.6f}")
            
            if not audio_segments:
                raise Exception("No audio segments generated")
            
            # Concatenate all audio segments
            if len(audio_segments) == 1:
                final_audio = audio_segments[0]
            else:
                final_audio = np.concatenate(audio_segments)
            
            # Final validation and cleanup
            if np.any(np.isnan(final_audio)):
                logger.error("❌ Final audio still contains NaN values, replacing with silence")
                final_audio = np.nan_to_num(final_audio, nan=0.0)
            
            if np.any(np.isinf(final_audio)):
                logger.error("❌ Final audio contains infinite values, clamping")
                final_audio = np.nan_to_num(final_audio, posinf=1.0, neginf=-1.0)
            
            # Ensure proper audio range [-1, 1]
            if np.max(np.abs(final_audio)) > 1.0:
                logger.info("📊 Normalizing final audio to [-1, 1] range")
                final_audio = final_audio / np.max(np.abs(final_audio))
            
            duration = len(final_audio) / self.sample_rate
            
            logger.info(f"✅ Successfully generated {duration:.2f}s of MLX-Audio Kokoro speech with {segment_count} segments")
            logger.info(f"📊 Final audio: shape={final_audio.shape}, min={np.min(final_audio):.6f}, max={np.max(final_audio):.6f}, mean={np.mean(final_audio):.6f}")
            
            return final_audio, duration
            
        except Exception as e:
            logger.error(f"❌ MLX-Audio Kokoro synthesis failed: {e}")
            raise Exception(f"MLX-Audio synthesis failed: {str(e)}")

def generate_cache_key(text: str, voice: str, speed: float, language: str) -> str:
    """Generate cache key for audio file"""
    content = f"mlx-native:{text}:{voice}:{speed}:{language}"
    return hashlib.md5(content.encode()).hexdigest()

@app.get("/health")
async def health_check():
    """Health check endpoint"""
    return {
        "status": "healthy",
        "service": "native-mlx-audio",
        "mlx_available": MLX_AUDIO_AVAILABLE,
        "kokoro_ready": tts_engine.kokoro_model is not None if tts_engine else False,
        "platform": "macOS Apple Silicon",
        "sample_rate": tts_engine.sample_rate if tts_engine else 24000,
        "supported_languages": list(tts_engine.supported_languages.keys()) if tts_engine else []
    }

@app.get("/models")
async def list_models():
    """List available TTS models"""
    return {
        "model": "kokoro-82m",
        "model_path": tts_engine.model_name if tts_engine else None,
        "supported_languages": tts_engine.supported_languages if tts_engine else {},
        "available_voices": list(tts_engine.voice_mapping.keys()) if tts_engine else [],
        "framework": "MLX-Audio (Native macOS)",
        "kokoro_initialized": tts_engine.kokoro_model is not None if tts_engine else False
    }

@app.post("/api/tts", response_model=TTSResponse)
async def synthesize_speech(request: TTSRequest):
    """Synthesize speech using native MLX-Audio Kokoro"""
    if not tts_engine:
        raise HTTPException(status_code=503, detail="TTS engine not initialized")
    
    try:
        # Generate cache key
        cache_key = generate_cache_key(request.text, request.voice, request.speed, request.language)
        cache_file = Path(AUDIO_CACHE_DIR) / f"{cache_key}.{request.format}"
        
        cache_hit = cache_file.exists()
        duration = 0.0
        
        if not cache_hit:
            # Generate audio
            logger.info(f"Generating audio for: '{request.text[:50]}...'")
            audio_data, duration = await tts_engine.synthesize(
                text=request.text,
                voice=request.voice,
                speed=request.speed,
                language=request.language
            )
            
            # Save to cache
            sf.write(str(cache_file), audio_data, tts_engine.sample_rate)
            logger.info(f"Audio saved to cache: {cache_file}")
        else:
            # For cached files, calculate duration from file
            info = sf.info(str(cache_file))
            duration = info.duration
        
        return TTSResponse(
            audio_url=f"/audio/{cache_key}.{request.format}",
            duration=duration,
            cache_hit=cache_hit,
            voice_used=request.voice
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
        "service": "Native MLX-Audio TTS Server",
        "version": "1.0.0",
        "mlx_available": MLX_AUDIO_AVAILABLE,
        "kokoro_ready": tts_engine.kokoro_model is not None if tts_engine else False,
        "platform": "macOS Apple Silicon (MLX optimized)",
        "model": tts_engine.model_name if tts_engine else None,
        "framework": "MLX-Audio",
        "endpoints": {
            "health": "/health",
            "models": "/models",
            "synthesize": "/api/tts",
            "audio": "/audio/{filename}"
        }
    }

if __name__ == "__main__":
    # Run the server
    logger.info(f"Starting Native MLX-Audio TTS Server on port {PORT}")
    uvicorn.run(
        "app:app",
        host="0.0.0.0",
        port=PORT,
        log_level="info",
        access_log=True,
        reload=False
    )