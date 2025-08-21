package services

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"intelligent-presenter-backend/pkg/config"
)

type SpeechService struct {
	config    *config.Config
	cacheDir  string
	client    *http.Client
}

type SpeechRequest struct {
	Text      string `json:"text"`
	Language  string `json:"language"`
	Voice     string `json:"voice"`
	Streaming bool   `json:"streaming"`
}

type SpeechResponse struct {
	AudioURL  string        `json:"audioUrl"`
	Duration  time.Duration `json:"duration"`
	Language  string        `json:"language"`
	Voice     string        `json:"voice"`
	CacheHit  bool          `json:"cacheHit"`
	RequestID string        `json:"requestId"`
}

func NewSpeechService(cfg *config.Config) *SpeechService {
	cacheDir := "./cache/audio"
	os.MkdirAll(cacheDir, 0755)
	
	return &SpeechService{
		config:   cfg,
		cacheDir: cacheDir,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *SpeechService) SynthesizeSpeech(text, language, voice string) (string, error) {
	// Generate cache key
	cacheKey := s.generateCacheKey(text, language, voice)
	audioFile := filepath.Join(s.cacheDir, cacheKey+".wav")
	
	// Check if audio file already exists in cache
	if _, err := os.Stat(audioFile); err == nil {
		// Return cached file URL
		return fmt.Sprintf("/api/v1/speech/audio/%s.wav", cacheKey), nil
	}
	
	// Check if we have a separate speech server running
	if s.config.MCPSpeechURL != "" {
		return s.callSpeechServer(text, language, voice, cacheKey)
	}
	
	// Fall back to simple TTS implementation
	return s.generateSimpleTTS(text, language, voice, audioFile, cacheKey)
}

func (s *SpeechService) callSpeechServer(text, language, voice, cacheKey string) (string, error) {
	request := SpeechRequest{
		Text:      text,
		Language:  language,
		Voice:     voice,
		Streaming: false,
	}
	
	requestBody, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}
	
	resp, err := s.client.Post(
		s.config.MCPSpeechURL+"/api/v1/synthesize",
		"application/json",
		bytes.NewBuffer(requestBody),
	)
	if err != nil {
		return "", fmt.Errorf("failed to call speech server: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("speech server returned status %d", resp.StatusCode)
	}
	
	var speechResponse SpeechResponse
	if err := json.NewDecoder(resp.Body).Decode(&speechResponse); err != nil {
		return "", fmt.Errorf("failed to decode speech response: %w", err)
	}
	
	return speechResponse.AudioURL, nil
}

func (s *SpeechService) generateSimpleTTS(text, language, voice, audioFile, cacheKey string) (string, error) {
	// Create a simple placeholder audio file
	// In production, this would use a real TTS engine
	
	duration := s.estimateDuration(text)
	sampleRate := 16000
	bitsPerSample := 16
	channels := 1
	
	// Calculate file size
	audioDataSize := int(duration.Seconds()) * sampleRate * bitsPerSample / 8 * channels
	fileSize := 36 + audioDataSize
	
	// Create WAV header
	header := make([]byte, 44)
	
	// RIFF header
	copy(header[0:4], "RIFF")
	header[4] = byte(fileSize & 0xff)
	header[5] = byte((fileSize >> 8) & 0xff)
	header[6] = byte((fileSize >> 16) & 0xff)
	header[7] = byte((fileSize >> 24) & 0xff)
	copy(header[8:12], "WAVE")
	
	// fmt subchunk
	copy(header[12:16], "fmt ")
	header[16] = 16 // Subchunk1Size for PCM
	header[20] = 1  // AudioFormat (PCM)
	header[22] = byte(channels)
	header[24] = byte(sampleRate & 0xff)
	header[25] = byte((sampleRate >> 8) & 0xff)
	
	// data subchunk
	copy(header[36:40], "data")
	header[40] = byte(audioDataSize & 0xff)
	header[41] = byte((audioDataSize >> 8) & 0xff)
	header[42] = byte((audioDataSize >> 16) & 0xff)
	header[43] = byte((audioDataSize >> 24) & 0xff)
	
	// Write to file
	file, err := os.Create(audioFile)
	if err != nil {
		return "", fmt.Errorf("failed to create audio file: %w", err)
	}
	defer file.Close()
	
	// Write header
	if _, err := file.Write(header); err != nil {
		return "", fmt.Errorf("failed to write WAV header: %w", err)
	}
	
	// Write silence (zeros) as placeholder audio data
	silenceData := make([]byte, audioDataSize)
	if _, err := file.Write(silenceData); err != nil {
		return "", fmt.Errorf("failed to write audio data: %w", err)
	}
	
	return fmt.Sprintf("/api/v1/speech/audio/%s.wav", cacheKey), nil
}

func (s *SpeechService) generateCacheKey(text, language, voice string) string {
	content := fmt.Sprintf("%s:%s:%s", text, language, voice)
	hash := md5.Sum([]byte(content))
	return fmt.Sprintf("%x", hash)
}

func (s *SpeechService) estimateDuration(text string) time.Duration {
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

func (s *SpeechService) ServeAudioFile(filename string) (string, error) {
	audioPath := filepath.Join(s.cacheDir, filename)
	
	if _, err := os.Stat(audioPath); os.IsNotExist(err) {
		return "", fmt.Errorf("audio file not found: %s", filename)
	}
	
	return audioPath, nil
}