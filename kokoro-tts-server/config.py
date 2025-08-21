"""
Kokoro TTS Server Configuration
82M Parameter Open-Weight TTS Model with multi-language support
"""

import os
from pathlib import Path

class Config:
    """Configuration settings for Kokoro TTS Server"""
    
    # Server settings
    HOST = "0.0.0.0"
    PORT = int(os.getenv("PORT", "8882"))
    
    # Model settings
    MODEL_CACHE_DIR = Path(os.getenv("MODEL_CACHE_DIR", "/app/models"))
    AUDIO_CACHE_DIR = Path(os.getenv("AUDIO_CACHE_DIR", "/app/cache"))
    
    # Audio settings
    SAMPLE_RATE = 24000  # Kokoro uses 24kHz
    AUDIO_FORMAT = "wav"
    
    # Kokoro TTS specific settings
    MODEL_NAME = "kokoro-82m"
    MODEL_PARAMETERS = "82 million"
    LICENSE = "Apache 2.0"
    
    # Language support mapping
    SUPPORTED_LANGUAGES = {
        'en': 'a',  # 🇺🇸 American English, 🇬🇧 British English
        'ja': 'j',  # 🇯🇵 Japanese
        'es': 'e',  # 🇪🇸 Spanish
        'fr': 'f',  # 🇫🇷 French
        'hi': 'h',  # 🇮🇳 Hindi
        'it': 'i',  # 🇮🇹 Italian
        'pt': 'p',  # 🇧🇷 Brazilian Portuguese
        'zh': 'z'   # 🇨🇳 Mandarin Chinese
    }
    
    # Available voices
    AVAILABLE_VOICES = [
        "af_heart",  # Default voice
    ]
    
    # Cache settings
    CACHE_MAX_SIZE_GB = 3.0
    CACHE_CLEANUP_INTERVAL = 3600  # 1 hour
    
    # Logging
    LOG_LEVEL = os.getenv("LOG_LEVEL", "INFO")
    
    def __init__(self):
        # Create directories
        self.MODEL_CACHE_DIR.mkdir(parents=True, exist_ok=True)
        self.AUDIO_CACHE_DIR.mkdir(parents=True, exist_ok=True)