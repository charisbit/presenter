# API Reference

## Overview

The Intelligent Presenter API provides endpoints for generating AI-powered presentation slides from Backlog project data. The API supports real-time generation updates via WebSocket and multiple AI providers with automatic fallback.

## Base URL

```
http://localhost:8000/api
```

## Authentication

The API uses token-based authentication. Include the authentication token in requests:

```http
Authorization: Bearer <token>
```

## Content Types

All requests and responses use JSON format:

```http
Content-Type: application/json
```

## Rate Limiting

- **Request Rate**: 10 requests per minute per IP
- **Generation Limit**: 1 concurrent generation per user
- **AI Provider Fallback**: Automatic switching on rate limit errors

## Endpoints

### 1. Generate Slides

Generate presentation slides for a Backlog project.

```http
POST /api/slides/generate
```

#### Request Body

```json
{
  "projectId": "string",
  "themes": ["string"],
  "language": "string"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `projectId` | string | Yes | Backlog project ID or key |
| `themes` | array[string] | Yes | List of slide themes to generate |
| `language` | string | Yes | Content language (`"ja"` or `"en"`) |

#### Available Themes

- `project_overview` - Project information and objectives
- `project_progress` - Completion rates and milestones
- `issue_management` - Issue tracking and resolution
- `risk_analysis` - Risk identification and mitigation
- `team_collaboration` - Team activities and communication
- `document_management` - Documentation and knowledge sharing
- `codebase_activity` - Development metrics and code quality
- `notifications` - Communication efficiency
- `predictive_analysis` - Forecasts and trends
- `summary_plan` - Project summary and planning

#### Response

```json
{
  "slideId": "uuid",
  "status": "generating",
  "websocketUrl": "ws://localhost:8000/ws/slides/{slideId}"
}
```

#### Status Codes

- `200` - Generation started successfully
- `400` - Invalid request parameters
- `401` - Authentication required
- `429` - Rate limit exceeded
- `500` - Internal server error

#### Example

```bash
curl -X POST http://localhost:8000/api/slides/generate \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "projectId": "PROJECT_123",
    "themes": ["project_overview", "project_progress"],
    "language": "ja"
  }'
```

### 2. Get Slide Status

Retrieve the current status of slide generation.

```http
GET /api/slides/{slideId}/status
```

#### Path Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `slideId` | string | Unique slide generation ID |

#### Response

```json
{
  "slideId": "uuid",
  "projectId": "string",
  "status": "string",
  "themes": ["string"]
}
```

#### Status Values

- `generating` - Slides are being generated
- `completed` - All slides generated successfully
- `error` - Generation failed
- `cancelled` - Generation was cancelled

#### Status Codes

- `200` - Status retrieved successfully
- `404` - Slide ID not found
- `401` - Authentication required

### 3. Health Check

Check API health and availability.

```http
GET /api/health
```

#### Response

```json
{
  "status": "healthy",
  "timestamp": "2024-08-17T10:30:00Z",
  "services": {
    "database": "healthy",
    "mcp_backlog": "healthy",
    "mcp_speech": "healthy",
    "ai_providers": {
      "openai": "healthy",
      "bedrock": "healthy"
    }
  }
}
```

## WebSocket API

### Connection

Connect to receive real-time slide generation updates:

```
ws://localhost:8000/ws/slides/{slideId}?token={authToken}
```

### Message Types

#### 1. Slide Content

Sent when a new slide is generated:

```json
{
  "type": "slide_content",
  "data": {
    "index": 1,
    "theme": "project_overview",
    "title": "Project Overview",
    "markdown": "# Project Overview\n\n...",
    "html": "<div><h1>Project Overview</h1>...</div>",
    "generatedAt": "2024-08-17T10:30:00Z"
  }
}
```

#### 2. Slide Narration

Sent when narration text is generated:

```json
{
  "type": "slide_narration",
  "data": {
    "slideIndex": 1,
    "text": "This slide presents the project overview...",
    "language": "en"
  }
}
```

#### 3. Slide Audio

Sent when audio file is generated:

```json
{
  "type": "slide_audio",
  "data": {
    "slideIndex": 1,
    "audioUrl": "/audio/slide_1_narration.mp3",
    "duration": 30
  }
}
```

#### 4. Presentation Complete

Sent when all slides are generated:

```json
{
  "type": "presentation_complete",
  "data": {
    "totalSlides": 10,
    "duration": "Generated successfully"
  }
}
```

#### 5. Error

Sent when generation errors occur:

```json
{
  "type": "error",
  "data": {
    "message": "Failed to generate slide",
    "code": "GENERATION_ERROR"
  }
}
```

## Error Handling

### Error Response Format

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable error message",
    "details": "Additional error details",
    "timestamp": "2024-08-17T10:30:00Z"
  }
}
```

### Common Error Codes

| Code | Description | HTTP Status |
|------|-------------|-------------|
| `INVALID_REQUEST` | Invalid request parameters | 400 |
| `AUTHENTICATION_REQUIRED` | Missing or invalid token | 401 |
| `PROJECT_NOT_FOUND` | Backlog project not found | 404 |
| `RATE_LIMIT_EXCEEDED` | Too many requests | 429 |
| `AI_SERVICE_UNAVAILABLE` | All AI providers unavailable | 503 |
| `GENERATION_ERROR` | Slide generation failed | 500 |

## Data Models

### SlideContent

```typescript
interface SlideContent {
  index: number           // Slide position (1-based)
  theme: SlideTheme      // Theme used for generation
  title: string          // Slide title
  markdown: string       // Source markdown content
  html?: string          // Rendered HTML (LLM-generated)
  generatedAt: string    // ISO timestamp
}
```

### SlideNarration

```typescript
interface SlideNarration {
  slideIndex: number     // Target slide index
  text: string           // Narration text
  language: string       // Language code
}
```

### SlideAudio

```typescript
interface SlideAudio {
  slideIndex: number     // Target slide index
  audioUrl: string       // Audio file URL
  duration: number       // Duration in seconds
}
```

## AI Provider Integration

### OpenAI Integration

- **Model**: GPT-3.5-turbo
- **Max Tokens**: 800 per request
- **Temperature**: 0.7
- **Automatic Fallback**: On rate limits or errors

### AWS Bedrock Integration

- **Model**: Claude-3 Sonnet
- **Region**: Configurable (default: us-east-1)
- **Fallback**: To OpenAI on failure

### Backlog MCP Integration

Available MCP tools for data retrieval:

- `get_project` - Project details
- `get_space` - Space information
- `get_users` - User list
- `get_issues` - Issue list with filters
- `count_issues` - Issue count statistics
- `get_issue_types` - Available issue types
- `get_priorities` - Priority levels
- `get_project_list` - Project list

## SDKs and Examples

### JavaScript/TypeScript

```typescript
import { SlideAPI } from '@intelligent-presenter/sdk'

const api = new SlideAPI({
  baseUrl: 'http://localhost:8000/api',
  token: 'your-auth-token'
})

// Generate slides
const response = await api.generateSlides({
  projectId: 'PROJECT_123',
  themes: ['project_overview', 'project_progress'],
  language: 'ja'
})

// Connect to WebSocket for updates
api.connectWebSocket(response.slideId, {
  onSlideContent: (slide) => console.log('New slide:', slide),
  onComplete: () => console.log('Generation complete'),
  onError: (error) => console.error('Error:', error)
})
```

### Go

```go
package main

import (
    "bytes"
    "encoding/json"
    "net/http"
)

type GenerateRequest struct {
    ProjectID string   `json:"projectId"`
    Themes    []string `json:"themes"`
    Language  string   `json:"language"`
}

func generateSlides(req GenerateRequest, token string) error {
    data, _ := json.Marshal(req)
    
    httpReq, _ := http.NewRequest("POST", 
        "http://localhost:8000/api/slides/generate",
        bytes.NewBuffer(data))
    
    httpReq.Header.Set("Content-Type", "application/json")
    httpReq.Header.Set("Authorization", "Bearer "+token)
    
    client := &http.Client{}
    resp, err := client.Do(httpReq)
    // Handle response...
    
    return err
}
```

### Python

```python
import requests
import websocket
import json

class IntelligentPresenterAPI:
    def __init__(self, base_url, token):
        self.base_url = base_url
        self.token = token
        self.headers = {
            'Content-Type': 'application/json',
            'Authorization': f'Bearer {token}'
        }
    
    def generate_slides(self, project_id, themes, language):
        data = {
            'projectId': project_id,
            'themes': themes,
            'language': language
        }
        
        response = requests.post(
            f'{self.base_url}/slides/generate',
            headers=self.headers,
            json=data
        )
        
        return response.json()
    
    def connect_websocket(self, slide_id, callbacks):
        ws_url = f'ws://localhost:8000/ws/slides/{slide_id}?token={self.token}'
        
        def on_message(ws, message):
            data = json.loads(message)
            if data['type'] in callbacks:
                callbacks[data['type']](data['data'])
        
        ws = websocket.WebSocketApp(ws_url, on_message=on_message)
        ws.run_forever()
```

## Testing

### Unit Tests

```bash
# Backend API tests
go test ./tests/api/...

# Frontend integration tests
npm run test:api
```

### Integration Tests

```bash
# Full API integration test
npm run test:integration
```

### Load Testing

```bash
# Test API performance
npm run test:load
```

## Changelog

### v1.2.0 (2024-08-17)
- Added LLM-based HTML compilation
- Enhanced AI provider fallback system
- Improved rate limiting and error handling

### v1.1.0 (2024-08-15)
- Added WebSocket real-time updates
- Implemented audio narration generation
- Added multi-language support

### v1.0.0 (2024-08-01)
- Initial API release
- Basic slide generation functionality
- Backlog integration via MCP