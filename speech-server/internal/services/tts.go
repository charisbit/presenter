// Package services provides text-to-speech synthesis services for the Speech MCP Server.
// It integrates multiple TTS engines including VOICEVOX, Kokoro TTS, and MLX-Audio
// to provide high-quality speech synthesis with caching and multilingual support.
package services

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"speech-mcp-server/internal/models"
	"speech-mcp-server/pkg/config"
	"github.com/google/uuid"
)

// TTSService provides text-to-speech synthesis capabilities using multiple engines.
// It manages voice selection, audio caching, engine fallback, and supports both
// Japanese and multilingual speech synthesis with high-quality neural voices.
type TTSService struct {
	config *config.Config // Service configuration including TTS engine preferences
}

// NewTTSService creates a new TTS service instance with the provided configuration.
// It initializes the service with access to all configured TTS engines and
// sets up audio caching directories.
//
// Parameters:
//   - cfg: Application configuration containing TTS engine settings
//
// Returns a configured TTSService ready for speech synthesis operations.
func NewTTSService(cfg *config.Config) *TTSService {
	return &TTSService{
		config: cfg,
	}
}

// SynthesizeSpeech converts text to speech using the best available TTS engine.
// It implements intelligent caching, engine selection, and fallback strategies
// to provide reliable high-quality speech synthesis.
//
// The synthesis process:
//   1. Generates a cache key based on text, language, and voice parameters
//   2. Checks for existing cached audio to improve performance
//   3. Selects appropriate TTS engine based on language and configuration
//   4. Generates audio using the selected engine with fallback support
//   5. Returns audio URL, metadata, and performance information
//
// Parameters:
//   - req: Speech synthesis request containing text, language, and voice preferences
//
// Returns:
//   - *models.SpeechResponse: Complete response with audio URL and metadata
//   - error: Any error that occurred during synthesis
func (s *TTSService) SynthesizeSpeech(req models.SpeechRequest) (*models.SpeechResponse, error) {
	// Generate cache key based on text, language, and voice
	cacheKey := s.generateCacheKey(req.Text, req.Language, req.Voice)
	
	// Check if audio file already exists in cache
	audioFile := filepath.Join(s.config.CacheDir, cacheKey+"."+s.config.AudioFormat)
	
	var cacheHit bool
	if _, err := os.Stat(audioFile); err == nil {
		cacheHit = true
	} else {
		// Generate audio file
		if err := s.generateAudioFile(req, audioFile); err != nil {
			return nil, fmt.Errorf("failed to generate audio: %w", err)
		}
		cacheHit = false
	}
	
	// Generate audio URL
	audioURL := fmt.Sprintf("/cache/%s.%s", cacheKey, s.config.AudioFormat)
	
	return &models.SpeechResponse{
		AudioURL:  audioURL,
		Duration:  s.estimateDuration(req.Text),
		Language:  req.Language,
		Voice:     req.Voice,
		CacheHit:  cacheHit,
		RequestID: uuid.New().String(),
	}, nil
}

// generateCacheKey creates a unique cache key for the TTS request.
// It uses MD5 hashing of the text, language, and voice parameters
// to create a consistent identifier for audio file caching.
//
// Parameters:
//   - text: The text content to be synthesized
//   - language: The target language code
//   - voice: The voice identifier or preference
//
// Returns a unique hash string suitable for use as a filename.
func (s *TTSService) generateCacheKey(text, language, voice string) string {
	content := fmt.Sprintf("%s:%s:%s", text, language, voice)
	hash := md5.Sum([]byte(content))
	return fmt.Sprintf("%x", hash)
}

// generateAudioFile creates the actual audio file using Japanese TTS engines
func (s *TTSService) generateAudioFile(req models.SpeechRequest, outputPath string) error {
	// Ensure cache directory exists
	if err := os.MkdirAll(s.config.CacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}
	
	// Use M4-optimized TTS to generate high-quality audio
	return s.generateM4OptimizedAudio(req, outputPath)
}

// estimateDuration estimates speech duration based on text length
func (s *TTSService) estimateDuration(text string) time.Duration {
	// Rough estimation: average speaking rate is about 150-160 words per minute
	// For Japanese, we'll estimate based on character count
	
	wordCount := len([]rune(text)) / 3 // Rough estimate for Japanese
	if wordCount < 1 {
		wordCount = 1
	}
	
	// Assume 150 words per minute
	minutes := float64(wordCount) / 150.0
	seconds := minutes * 60.0
	
	// Minimum duration of 1 second
	if seconds < 1.0 {
		seconds = 1.0
	}
	
	return time.Duration(seconds * float64(time.Second))
}

// GetAvailableVoices returns a comprehensive list of available voices from all TTS engines.
// It includes voices from VOICEVOX (Japanese high-quality), Kokoro TTS (multilingual),
// and MLX-Audio (Apple Silicon optimized) with detailed metadata for each voice.
//
// The returned voices support:
//   - Japanese: VOICEVOX (2 voices) + Kokoro (1 voice) + MLX-Audio (2 voices)
//   - Multilingual: Kokoro TTS supporting 8 languages with natural voices
//
// Returns a slice of VoiceInfo structs containing voice metadata for client selection.
func (s *TTSService) GetAvailableVoices() []models.VoiceInfo {
	return []models.VoiceInfo{
		// VOICEVOX voices (high-quality Japanese TTS)
		{
			ID:       "voicevox-ja-female",
			Name:     "Japanese High-Quality Female (VOICEVOX)",
			Language: "ja",
			Gender:   "female",
			Styles:   []string{"natural", "clear", "expressive"},
		},
		{
			ID:       "voicevox-ja-male",
			Name:     "Japanese High-Quality Male (VOICEVOX)",
			Language: "ja",
			Gender:   "male",
			Styles:   []string{"natural", "clear", "expressive"},
		},
		// Kokoro TTS voices (82M parameter multilingual model)
		{
			ID:       "kokoro-ja-heart",
			Name:     "Japanese Natural Voice (Kokoro 82M)",
			Language: "ja",
			Gender:   "female",
			Styles:   []string{"natural", "multilingual", "82m-params"},
		},
		{
			ID:       "kokoro-en-heart",
			Name:     "English Natural Voice (Kokoro 82M)",
			Language: "en",
			Gender:   "female",
			Styles:   []string{"natural", "multilingual", "82m-params"},
		},
		{
			ID:       "kokoro-es-heart",
			Name:     "Spanish Natural Voice (Kokoro 82M)",
			Language: "es",
			Gender:   "female",
			Styles:   []string{"natural", "multilingual", "82m-params"},
		},
		{
			ID:       "kokoro-fr-heart",
			Name:     "French Natural Voice (Kokoro 82M)",
			Language: "fr",
			Gender:   "female",
			Styles:   []string{"natural", "multilingual", "82m-params"},
		},
		{
			ID:       "kokoro-hi-heart",
			Name:     "Hindi Natural Voice (Kokoro 82M)",
			Language: "hi",
			Gender:   "female",
			Styles:   []string{"natural", "multilingual", "82m-params"},
		},
		{
			ID:       "kokoro-it-heart",
			Name:     "Italian Natural Voice (Kokoro 82M)",
			Language: "it",
			Gender:   "female",
			Styles:   []string{"natural", "multilingual", "82m-params"},
		},
		{
			ID:       "kokoro-pt-heart",
			Name:     "Portuguese Natural Voice (Kokoro 82M)",
			Language: "pt",
			Gender:   "female",
			Styles:   []string{"natural", "multilingual", "82m-params"},
		},
		{
			ID:       "kokoro-zh-heart",
			Name:     "Chinese Natural Voice (Kokoro 82M)",
			Language: "zh",
			Gender:   "female",
			Styles:   []string{"natural", "multilingual", "82m-params"},
		},
		// MLX-Audio voices (Apple Silicon optimized)
		{
			ID:       "mlx-ja-female",
			Name:     "Japanese High-Quality Female (MLX-Audio)",
			Language: "ja",
			Gender:   "female",
			Styles:   []string{"natural", "neural", "apple-silicon"},
		},
		{
			ID:       "mlx-ja-male",
			Name:     "Japanese High-Quality Male (MLX-Audio)",
			Language: "ja",
			Gender:   "male",
			Styles:   []string{"natural", "neural", "apple-silicon"},
		},
	}
}

// GetSupportedLanguages returns a list of all supported languages across TTS engines.
// It provides comprehensive language information including native names,
// voice counts, and support status for client language selection.
//
// Supported languages:
//   - Japanese (ja): 4 voices via VOICEVOX, Kokoro, and MLX-Audio
//   - English, Spanish, French, Hindi, Italian, Portuguese, Chinese: 1 voice each via Kokoro
//
// Returns a slice of LanguageInfo structs with language metadata.
func (s *TTSService) GetSupportedLanguages() []models.LanguageInfo {
	return []models.LanguageInfo{
		{
			Code:       "ja",
			Name:       "Japanese",
			NativeName: "日本語",
			Voices:     4, // VOICEVOX (2) + Kokoro (1) + MLX-Audio (2)
			Supported:  true,
		},
		{
			Code:       "en",
			Name:       "English",
			NativeName: "English",
			Voices:     1, // Kokoro (1)
			Supported:  true,
		},
		{
			Code:       "es",
			Name:       "Spanish",
			NativeName: "Español",
			Voices:     1, // Kokoro (1)
			Supported:  true,
		},
		{
			Code:       "fr",
			Name:       "French",
			NativeName: "Français",
			Voices:     1, // Kokoro (1)
			Supported:  true,
		},
		{
			Code:       "hi",
			Name:       "Hindi",
			NativeName: "हिन्दी",
			Voices:     1, // Kokoro (1)
			Supported:  true,
		},
		{
			Code:       "it",
			Name:       "Italian",
			NativeName: "Italiano",
			Voices:     1, // Kokoro (1)
			Supported:  true,
		},
		{
			Code:       "pt",
			Name:       "Portuguese",
			NativeName: "Português",
			Voices:     1, // Kokoro (1)
			Supported:  true,
		},
		{
			Code:       "zh",
			Name:       "Chinese",
			NativeName: "中文",
			Voices:     1, // Kokoro (1)
			Supported:  true,
		},
	}
}

// generateM4OptimizedAudio generates high-quality audio with multi-language support for Mac M4
func (s *TTSService) generateM4OptimizedAudio(req models.SpeechRequest, outputPath string) error {
	// Get preferred TTS engine from environment
	preferredEngine := os.Getenv("TTS_ENGINE")
	
	// Support multiple languages with engine-specific routing
	switch req.Language {
	case "ja":
		return s.generateJapaneseAudio(req, outputPath, preferredEngine)
	case "en", "es", "fr", "hi", "it", "pt", "zh":
		return s.generateMultilingualAudio(req, outputPath, preferredEngine)
	default:
		return fmt.Errorf("language '%s' is not supported. Supported languages: ja, en, es, fr, hi, it, pt, zh", req.Language)
	}
}

// generateJapaneseAudio generates Japanese audio using VOICEVOX/Kokoro/MLX-Audio with new priority order
func (s *TTSService) generateJapaneseAudio(req models.SpeechRequest, outputPath string, preferredEngine string) error {
	// Japanese TTS priority: VOICEVOX (primary) -> Kokoro (secondary) -> MLX-Audio (fallback)
	switch preferredEngine {
	case "voicevox":
		if err := s.generateVoicevoxAudio(req, outputPath); err == nil {
			return nil
		} else {
			fmt.Printf("VOICEVOX TTS failed, trying Kokoro: %v\n", err)
		}
		// Fallback to Kokoro
		if err := s.generateKokoroAudio(req, outputPath); err == nil {
			return nil
		} else {
			fmt.Printf("Kokoro failed, trying MLX-Audio: %v\n", err)
		}
		// Final fallback to MLX-Audio
		return s.generateMLXAudio(req, outputPath)
	case "kokoro":
		if err := s.generateKokoroAudio(req, outputPath); err == nil {
			return nil
		} else {
			fmt.Printf("Kokoro TTS failed, trying VOICEVOX: %v\n", err)
		}
		// Fallback to VOICEVOX
		if err := s.generateVoicevoxAudio(req, outputPath); err == nil {
			return nil
		} else {
			fmt.Printf("VOICEVOX failed, trying MLX-Audio: %v\n", err)
		}
		// Final fallback to MLX-Audio
		return s.generateMLXAudio(req, outputPath)
	case "mlx-audio":
		if err := s.generateMLXAudio(req, outputPath); err == nil {
			return nil
		} else {
			fmt.Printf("MLX-Audio failed, trying VOICEVOX: %v\n", err)
		}
		// Fallback to VOICEVOX
		if err := s.generateVoicevoxAudio(req, outputPath); err == nil {
			return nil
		}
		// Final fallback to Kokoro
		return s.generateKokoroAudio(req, outputPath)
	default:
		// Default order for Japanese: VOICEVOX -> Kokoro -> MLX-Audio
		if err := s.generateVoicevoxAudio(req, outputPath); err == nil {
			return nil
		}
		if err := s.generateKokoroAudio(req, outputPath); err == nil {
			return nil
		}
		return s.generateMLXAudio(req, outputPath)
	}
}

// generateMultilingualAudio generates non-Japanese audio using Kokoro TTS
func (s *TTSService) generateMultilingualAudio(req models.SpeechRequest, outputPath string, preferredEngine string) error {
	// For non-Japanese languages, use Kokoro TTS as primary engine
	fmt.Printf("Using Kokoro TTS for %s language text: %s\n", req.Language, req.Text[:min(50, len(req.Text))])
	return s.generateKokoroAudio(req, outputPath)
}

// generateVoicevoxAudio generates high-quality Japanese audio using VOICEVOX Engine
func (s *TTSService) generateVoicevoxAudio(req models.SpeechRequest, outputPath string) error {
	// Get VOICEVOX Engine URL from environment or use default
	voicevoxURL := os.Getenv("VOICEVOX_ENGINE_URL")
	if voicevoxURL == "" {
		voicevoxURL = "http://localhost:50021"
	}
	
	fmt.Printf("Using VOICEVOX Engine for Japanese text: %s\n", req.Text[:min(50, len(req.Text))])
	
	// Check if VOICEVOX Engine is available
	client := &http.Client{Timeout: 5 * time.Second}
	if _, err := client.Get(voicevoxURL + "/docs"); err != nil {
		return fmt.Errorf("VOICEVOX Engine not available: %w", err)
	}
	
	// Use speaker ID "3" (ずんだもん ノーマル) as default
	speakerID := "3"
	if strings.Contains(strings.ToLower(req.Voice), "male") {
		speakerID = "2" // Alternative male voice option
	}
	
	// Step 1: Create audio query
	// POST /audio_query?text=<encoded_text>&speaker=<speaker_id>
	encodedText := url.QueryEscape(req.Text)
	queryURL := fmt.Sprintf("%s/audio_query?text=%s&speaker=%s", 
		voicevoxURL, 
		encodedText, 
		speakerID)
	
	queryResp, err := client.Post(queryURL, "application/json", nil)
	if err != nil {
		return fmt.Errorf("VOICEVOX audio_query failed: %w", err)
	}
	defer queryResp.Body.Close()
	
	if queryResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(queryResp.Body)
		return fmt.Errorf("VOICEVOX audio_query returned status %d: %s", queryResp.StatusCode, string(body))
	}
	
	// Read the query response (this is the JSON query object)
	queryData, err := io.ReadAll(queryResp.Body)
	if err != nil {
		return fmt.Errorf("failed to read query response: %w", err)
	}
	
	// Validate that we received valid JSON
	var queryJSON map[string]interface{}
	if err := json.Unmarshal(queryData, &queryJSON); err != nil {
		return fmt.Errorf("audio_query response is not valid JSON: %w", err)
	}
	
	// Step 2: Synthesize audio
	// POST /synthesis?speaker=<speaker_id> with the query JSON as body
	synthURL := fmt.Sprintf("%s/synthesis?speaker=%s", voicevoxURL, speakerID)
	synthReq, err := http.NewRequest("POST", synthURL, bytes.NewReader(queryData))
	if err != nil {
		return fmt.Errorf("failed to create synthesis request: %w", err)
	}
	
	synthReq.Header.Set("Content-Type", "application/json")
	synthReq.Header.Set("Accept", "audio/wav")
	
	client = &http.Client{Timeout: 30 * time.Second}
	synthResp, err := client.Do(synthReq)
	if err != nil {
		return fmt.Errorf("VOICEVOX synthesis failed: %w", err)
	}
	defer synthResp.Body.Close()
	
	if synthResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(synthResp.Body)
		return fmt.Errorf("VOICEVOX synthesis returned status %d: %s", synthResp.StatusCode, string(body))
	}
	
	// Create output file
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()
	
	// Copy audio data to file
	_, err = io.Copy(file, synthResp.Body)
	if err != nil {
		return fmt.Errorf("failed to write audio data: %w", err)
	}
	
	// Verify the output file was created and has content
	fileStat, err := os.Stat(outputPath)
	if err != nil {
		return fmt.Errorf("audio file was not created: %w", err)
	}
	if fileStat.Size() == 0 {
		return fmt.Errorf("audio file is empty")
	}
	
	fmt.Printf("Successfully generated audio using VOICEVOX: %s (%d bytes)\n", outputPath, fileStat.Size())
	return nil
}

// generateMLXAudio generates high-quality Japanese audio using MLX-Audio TTS
func (s *TTSService) generateMLXAudio(req models.SpeechRequest, outputPath string) error {
	// Get MLX-Audio URL from environment or use default
	mlxURL := os.Getenv("MLX_AUDIO_URL")
	if mlxURL == "" {
		mlxURL = "http://localhost:8881"
	}
	
	fmt.Printf("Using MLX-Audio for Japanese text: %s\n", req.Text[:min(50, len(req.Text))])
	
	// Check if MLX-Audio server is available
	client := &http.Client{Timeout: 5 * time.Second}
	if _, err := client.Get(mlxURL + "/health"); err != nil {
		return fmt.Errorf("MLX-Audio server not available: %w", err)
	}
	
	// Map voice requests to MLX-Audio voice parameters
	voice := "female"
	if strings.Contains(strings.ToLower(req.Voice), "male") {
		voice = "male"
	}
	
	// Prepare request payload for MLX-Audio API
	payload := map[string]interface{}{
		"text":     req.Text,
		"language": req.Language,
		"voice":    voice,
		"format":   "wav",
		"speed":    1.0,
	}
	
	// Convert payload to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal request payload: %w", err)
	}
	
	// Create HTTP request to MLX-Audio API
	url := mlxURL + "/api/tts"
	req_http, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}
	
	// Set headers
	req_http.Header.Set("Content-Type", "application/json")
	req_http.Header.Set("Accept", "audio/wav")
	
	// Send request
	client = &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req_http)
	if err != nil {
		return fmt.Errorf("MLX-Audio request failed: %w", err)
	}
	defer resp.Body.Close()
	
	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("MLX-Audio returned status %d: %s", resp.StatusCode, string(body))
	}
	
	// Create output file
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()
	
	// Copy audio data to file
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write audio data: %w", err)
	}
	
	// Verify the output file was created and has content
	fileStat, err := os.Stat(outputPath)
	if err != nil {
		return fmt.Errorf("audio file was not created: %w", err)
	}
	if fileStat.Size() == 0 {
		return fmt.Errorf("audio file is empty")
	}
	
	fmt.Printf("Successfully generated audio using MLX-Audio: %s (%d bytes)\n", outputPath, fileStat.Size())
	return nil
}

// generateKokoroAudio generates high-quality multilingual audio using Kokoro TTS (82M parameter model)
func (s *TTSService) generateKokoroAudio(req models.SpeechRequest, outputPath string) error {
	// Get Kokoro TTS URL from environment or use default
	kokoroURL := os.Getenv("KOKORO_TTS_URL")
	if kokoroURL == "" {
		kokoroURL = "http://localhost:8882"
	}
	
	fmt.Printf("Using Kokoro TTS for %s text: %s\n", req.Language, req.Text[:min(50, len(req.Text))])
	
	// Check if Kokoro TTS server is available
	client := &http.Client{Timeout: 5 * time.Second}
	if _, err := client.Get(kokoroURL + "/health"); err != nil {
		return fmt.Errorf("Kokoro TTS server not available: %w", err)
	}
	
	// Map voice requests to Kokoro voice parameters
	voice := "af_heart" // Default Kokoro voice
	
	// Prepare request payload for Kokoro TTS API
	payload := map[string]interface{}{
		"text":     req.Text,
		"language": req.Language,
		"voice":    voice,
		"format":   "wav",
		"speed":    1.0,
	}
	
	// Convert payload to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal request payload: %w", err)
	}
	
	// Create HTTP request to Kokoro TTS API
	url := kokoroURL + "/api/tts"
	req_http, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}
	
	// Set headers
	req_http.Header.Set("Content-Type", "application/json")
	req_http.Header.Set("Accept", "application/json")
	
	// Send request for TTS metadata
	client = &http.Client{Timeout: 600 * time.Second}
	resp, err := client.Do(req_http)
	if err != nil {
		return fmt.Errorf("Kokoro TTS request failed: %w", err)
	}
	defer resp.Body.Close()
	
	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Kokoro TTS returned status %d: %s", resp.StatusCode, string(body))
	}
	
	// Parse the response to get audio URL
	var ttsResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&ttsResponse); err != nil {
		return fmt.Errorf("failed to parse TTS response: %w", err)
	}
	
	audioURL, ok := ttsResponse["audio_url"].(string)
	if !ok {
		return fmt.Errorf("audio_url not found in TTS response")
	}
	
	// Download the audio file
	audioResp, err := client.Get(kokoroURL + audioURL)
	if err != nil {
		return fmt.Errorf("failed to download audio file: %w", err)
	}
	defer audioResp.Body.Close()
	
	if audioResp.StatusCode != http.StatusOK {
		return fmt.Errorf("audio download returned status %d", audioResp.StatusCode)
	}
	
	// Create output file
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()
	
	// Copy audio data to file
	_, err = io.Copy(file, audioResp.Body)
	if err != nil {
		return fmt.Errorf("failed to write audio data: %w", err)
	}
	
	// Verify the output file was created and has content
	fileStat, err := os.Stat(outputPath)
	if err != nil {
		return fmt.Errorf("audio file was not created: %w", err)
	}
	if fileStat.Size() == 0 {
		return fmt.Errorf("audio file is empty")
	}
	
	fmt.Printf("Successfully generated audio using Kokoro TTS: %s (%d bytes)\n", outputPath, fileStat.Size())
	return nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}