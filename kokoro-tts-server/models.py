"""
Kokoro TTS Models
82M Parameter Open-Weight TTS Model definitions and utilities
"""

from typing import Dict, List, Optional
from pydantic import BaseModel

class VoiceModel(BaseModel):
    """Voice model configuration for Kokoro TTS"""
    id: str
    name: str
    language: str
    language_code: str
    sample_rate: int = 24000
    description: str = ""

class KokoroModels:
    """Available models and languages for Kokoro TTS"""
    
    # Language mappings based on Kokoro documentation
    LANGUAGE_CODES: Dict[str, str] = {
        'en': 'a',  # 🇺🇸 American English, 🇬🇧 British English  
        'ja': 'j',  # 🇯🇵 Japanese
        'es': 'e',  # 🇪🇸 Spanish
        'fr': 'f',  # 🇫🇷 French fr-fr
        'hi': 'h',  # 🇮🇳 Hindi
        'it': 'i',  # 🇮🇹 Italian
        'pt': 'p',  # 🇧🇷 Brazilian Portuguese pt-br
        'zh': 'z'   # 🇨🇳 Mandarin Chinese
    }
    
    MODELS: Dict[str, VoiceModel] = {
        "kokoro-en": VoiceModel(
            id="kokoro-en",
            name="Kokoro English (82M)",
            language="en",
            language_code="a",
            description="82M parameter open-weight TTS model for English"
        ),
        "kokoro-ja": VoiceModel(
            id="kokoro-ja", 
            name="Kokoro Japanese (82M)",
            language="ja",
            language_code="j",
            description="82M parameter open-weight TTS model for Japanese"
        ),
        "kokoro-es": VoiceModel(
            id="kokoro-es",
            name="Kokoro Spanish (82M)",
            language="es",
            language_code="e",
            description="82M parameter open-weight TTS model for Spanish"
        ),
        "kokoro-fr": VoiceModel(
            id="kokoro-fr",
            name="Kokoro French (82M)",
            language="fr",
            language_code="f",
            description="82M parameter open-weight TTS model for French"
        ),
        "kokoro-hi": VoiceModel(
            id="kokoro-hi",
            name="Kokoro Hindi (82M)",
            language="hi",
            language_code="h",
            description="82M parameter open-weight TTS model for Hindi"
        ),
        "kokoro-it": VoiceModel(
            id="kokoro-it",
            name="Kokoro Italian (82M)",
            language="it",
            language_code="i",
            description="82M parameter open-weight TTS model for Italian"
        ),
        "kokoro-pt": VoiceModel(
            id="kokoro-pt",
            name="Kokoro Portuguese (82M)",
            language="pt",
            language_code="p",
            description="82M parameter open-weight TTS model for Brazilian Portuguese"
        ),
        "kokoro-zh": VoiceModel(
            id="kokoro-zh",
            name="Kokoro Chinese (82M)",
            language="zh",
            language_code="z",
            description="82M parameter open-weight TTS model for Mandarin Chinese"
        )
    }
    
    @classmethod
    def get_model(cls, model_id: str) -> Optional[VoiceModel]:
        """Get model by ID"""
        return cls.MODELS.get(model_id)
    
    @classmethod
    def get_models_by_language(cls, language: str) -> List[VoiceModel]:
        """Get all models for a specific language"""
        return [model for model in cls.MODELS.values() if model.language == language]
    
    @classmethod
    def get_all_models(cls) -> List[VoiceModel]:
        """Get all available models"""
        return list(cls.MODELS.values())
    
    @classmethod
    def get_language_code(cls, language: str) -> Optional[str]:
        """Get Kokoro language code for given language"""
        return cls.LANGUAGE_CODES.get(language)
    
    @classmethod
    def get_supported_languages(cls) -> List[str]:
        """Get list of supported languages"""
        return list(cls.LANGUAGE_CODES.keys())
    
    @classmethod
    def is_language_supported(cls, language: str) -> bool:
        """Check if language is supported"""
        return language in cls.LANGUAGE_CODES