package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"intelligent-presenter-backend/pkg/config"
)

type BedrockService struct {
	config *config.Config
	client *http.Client
}

type BedrockRequest struct {
	Prompt            string  `json:"prompt"`
	MaxTokensToSample int     `json:"max_tokens_to_sample"`
	Temperature       float64 `json:"temperature"`
	TopP              float64 `json:"top_p"`
	TopK              int     `json:"top_k"`
	StopSequences     []string `json:"stop_sequences"`
}

type BedrockResponse struct {
	Completion string `json:"completion"`
	StopReason string `json:"stop_reason"`
}

type ClaudeMessageRequest struct {
	Model         string    `json:"model"`
	MaxTokens     int       `json:"max_tokens"`
	Temperature   float64   `json:"temperature"`
	Messages      []Message `json:"messages"`
	AnthropicVersion string `json:"anthropic_version"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ClaudeMessageResponse struct {
	Content []ContentBlock `json:"content"`
	ID      string         `json:"id"`
	Model   string         `json:"model"`
	Role    string         `json:"role"`
	Type    string         `json:"type"`
	Usage   Usage          `json:"usage"`
}

type ContentBlock struct {
	Text string `json:"text"`
	Type string `json:"type"`
}

type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

func NewBedrockService(cfg *config.Config) *BedrockService {
	return &BedrockService{
		config: cfg,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (s *BedrockService) GenerateText(prompt string) (string, error) {
	// Use Claude-3 Messages API format for newer models
	if s.isClaudeMessagesModel() {
		return s.generateWithMessages(prompt)
	}
	
	// Use legacy text completion for older models
	return s.generateWithCompletion(prompt)
}

func (s *BedrockService) isClaudeMessagesModel() bool {
	modelID := s.config.BedrockModelID
	return modelID == "anthropic.claude-3-haiku-20240307-v1:0" ||
		   modelID == "anthropic.claude-3-sonnet-20240229-v1:0" ||
		   modelID == "anthropic.claude-3-opus-20240229-v1:0" ||
		   modelID == "anthropic.claude-3-5-sonnet-20240620-v1:0"
}

func (s *BedrockService) generateWithMessages(prompt string) (string, error) {
	request := ClaudeMessageRequest{
		Model:       s.config.BedrockModelID,
		MaxTokens:   1500,
		Temperature: 0.7,
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		AnthropicVersion: "bedrock-2023-05-31",
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := s.callBedrock(jsonData)
	if err != nil {
		return "", err
	}

	var response ClaudeMessageResponse
	if err := json.Unmarshal(resp, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(response.Content) == 0 {
		return "", fmt.Errorf("no content in response")
	}

	return response.Content[0].Text, nil
}

func (s *BedrockService) generateWithCompletion(prompt string) (string, error) {
	// Format prompt for Claude completion models
	formattedPrompt := fmt.Sprintf("\n\nHuman: %s\n\nAssistant:", prompt)
	
	request := BedrockRequest{
		Prompt:            formattedPrompt,
		MaxTokensToSample: 1500,
		Temperature:       0.7,
		TopP:              0.9,
		TopK:              250,
		StopSequences:     []string{"\n\nHuman:"},
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := s.callBedrock(jsonData)
	if err != nil {
		return "", err
	}

	var response BedrockResponse
	if err := json.Unmarshal(resp, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return response.Completion, nil
}

func (s *BedrockService) callBedrock(jsonData []byte) ([]byte, error) {
	url := fmt.Sprintf("https://bedrock-runtime.%s.amazonaws.com/model/%s/invoke",
		s.config.AWSRegion, s.config.BedrockModelID)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add AWS Signature V4 headers
	if err := s.signRequest(req, jsonData); err != nil {
		return nil, fmt.Errorf("failed to sign request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	fmt.Printf("Making Bedrock API call to model: %s\n", s.config.BedrockModelID)
	
	resp, err := s.client.Do(req)
	if err != nil {
		fmt.Printf("Bedrock API call error: %v\n", err)
		return nil, fmt.Errorf("failed to call Bedrock API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("Bedrock API error - Status: %d\n", resp.StatusCode)
		var errorBytes bytes.Buffer
		errorBytes.ReadFrom(resp.Body)
		fmt.Printf("Bedrock error response: %s\n", errorBytes.String())
		return nil, fmt.Errorf("Bedrock API returned status %d", resp.StatusCode)
	}

	var responseBody bytes.Buffer
	responseBody.ReadFrom(resp.Body)
	
	fmt.Printf("Bedrock API call successful\n")
	return responseBody.Bytes(), nil
}

func (s *BedrockService) signRequest(req *http.Request, payload []byte) error {
	// For simplicity, we'll use AWS credentials directly
	// In production, consider using AWS SDK for proper signing
	accessKey := s.config.AWSAccessKeyID
	secretKey := s.config.AWSSecretAccessKey
	region := s.config.AWSRegion
	
	if accessKey == "" || secretKey == "" {
		return fmt.Errorf("AWS credentials not configured")
	}

	// Create AWS Signature V4
	signer := &AWSV4Signer{
		AccessKey: accessKey,
		SecretKey: secretKey,
		Region:    region,
		Service:   "bedrock",
	}

	return signer.SignRequest(req, payload)
}

