package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"intelligent-presenter-backend/pkg/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

type BedrockSDKService struct {
	config *config.Config
	client *bedrockruntime.Client
}

// Types are shared with bedrock.go - no need to redeclare them here

func NewBedrockSDKService(cfg *config.Config) (*BedrockSDKService, error) {
	// Create AWS config with explicit credentials
	awsCfg, err := awsconfig.LoadDefaultConfig(context.TODO(),
		awsconfig.WithRegion(cfg.AWSRegion),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AWSAccessKeyID,
			cfg.AWSSecretAccessKey,
			"", // no session token
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create Bedrock Runtime client
	client := bedrockruntime.NewFromConfig(awsCfg)

	return &BedrockSDKService{
		config: cfg,
		client: client,
	}, nil
}

func (s *BedrockSDKService) GenerateText(prompt string) (string, error) {
	// Use Claude-3 Messages API format for Bedrock (without model field)
	request := map[string]interface{}{
		"max_tokens":         1500,
		"temperature":        0.7,
		"messages": []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		"anthropic_version": "bedrock-2023-05-31",
	}

	// Marshal the request
	requestBody, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	fmt.Printf("Making Bedrock API call using AWS SDK to model: %s\n", s.config.BedrockModelID)

	// Call Bedrock using AWS SDK
	output, err := s.client.InvokeModel(context.TODO(), &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(s.config.BedrockModelID),
		ContentType: aws.String("application/json"),
		Accept:      aws.String("application/json"),
		Body:        requestBody,
	})

	if err != nil {
		fmt.Printf("Bedrock SDK API call error: %v\n", err)
		return "", fmt.Errorf("failed to call Bedrock API: %w", err)
	}

	// Parse the response
	var response ClaudeMessageResponse
	if err := json.Unmarshal(output.Body, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(response.Content) == 0 {
		return "", fmt.Errorf("no content in response")
	}

	fmt.Printf("Bedrock SDK API call successful\n")
	return response.Content[0].Text, nil
}

func (s *BedrockSDKService) isClaudeMessagesModel() bool {
	modelID := s.config.BedrockModelID
	return strings.Contains(modelID, "claude-3")
}