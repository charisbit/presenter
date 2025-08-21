<!--
.. title: Backend Documentation (Go)
.. slug: backend-docs
.. date: 2025-08-18
.. tags: backend, go, documentation
.. category: 
.. link: 
.. description: Go backend documentation for the Intelligent Presenter
.. type: text
-->

# Backend Documentation (Go)

The Intelligent Presenter backend is built with Go and provides a robust, scalable API for slide generation and project management integration.

## Architecture Overview

The backend follows a clean architecture pattern with clear separation of concerns:

```text
backend/
├── cmd/
│   └── main.go              # Application entry point
├── internal/
│   ├── api/
│   │   ├── handlers/        # HTTP request handlers
│   │   └── routes.go        # Route configuration
│   ├── auth/
│   │   └── middleware.go    # Authentication middleware
│   ├── mcp/
│   │   └── client.go        # MCP client implementation
│   ├── models/              # Data models and types
│   └── services/            # Business logic services
├── pkg/
│   └── config/              # Configuration management
└── tests/                   # Test suites
```

## Core Packages

### Main Application (`cmd/main.go`)

The main application entry point that:
- Loads environment configuration
- Sets up HTTP server with Gin framework
- Configures CORS and middleware
- Initializes all API routes
- Implements graceful shutdown

**Key Features:**
- Environment variable configuration with `.env` support
- Graceful shutdown with signal handling
- CORS configuration for frontend integration
- Centralized error handling and logging

### Configuration (`pkg/config`)

**Config Structure:**
```go
type Config struct {
    Port                string
    Environment         string
    BacklogDomain       string
    BacklogClientID     string
    BacklogClientSecret string
    OAuthRedirectURL    string
    AIProvider          string
    OpenAIAPIKey        string
    AWSAccessKeyID      string
    AWSSecretAccessKey  string
    AWSRegion           string
    MCPBacklogURL       string
    MCPSpeechURL        string
    JWTSecretKey        string
    SpeechCacheDir      string
}
```

Configuration is loaded from environment variables with sensible defaults for development.

### Authentication (`internal/auth`)

**JWT Middleware:**
- Validates JWT tokens from Authorization header
- Extracts user information and Backlog tokens
- Provides WebSocket authentication support
- Implements token generation and validation

**Key Functions:**
- `RequireAuth()` - HTTP middleware for protected routes
- `RequireWSAuth()` - WebSocket authentication
- `GenerateToken()` - Create JWT tokens with user claims
- `ValidateToken()` - Token validation and claims extraction

### API Handlers (`internal/api/handlers`)

#### AuthHandler
Manages OAuth2 authentication flow with Backlog:
- OAuth2 state management with CSRF protection
- Token exchange and user information retrieval
- JWT token generation and refresh
- Session management

#### SlideHandler
Handles slide generation and WebSocket communication:
- Slide generation request processing
- Real-time progress updates via WebSocket
- Session management for concurrent generations
- Integration with AI services for content creation

#### MCPHandler
Provides project data access through MCP protocol:
- Project list retrieval
- Project details and analytics
- Integration with Backlog MCP server
- Data validation and error handling

### Business Logic Services (`internal/services`)

#### SlideService
Core slide generation service:
- AI provider abstraction (OpenAI, AWS Bedrock)
- Content generation from project data
- Theme-based slide templates
- Multi-language support

**Key Methods:**
```go
func (s *SlideService) GenerateSlides(projectData, themes, language) (*SlideContent, error)
func (s *SlideService) ProcessSlideContent(content) (string, error)
func (s *SlideService) GenerateSlideAudio(text, language) (*AudioFile, error)
```

#### MCPService
Model Context Protocol service for external data access:
- MCP client management and session handling
- Backlog project data retrieval
- Error handling and retry logic
- Data transformation and validation

#### BedrockService & BedrockSDKService
AWS Bedrock integration services:
- Multiple implementation approaches (HTTP and SDK)
- Request signing and authentication
- Model selection and parameter optimization
- Response processing and error handling

#### SpeechService
Text-to-speech functionality:
- Multiple TTS provider support
- Audio file caching and optimization
- Language and voice selection
- Streaming audio support

### MCP Client (`internal/mcp`)

**MCPClient Features:**
- JSON-RPC 2.0 protocol implementation
- Session management for stateful connections
- Tool invocation and resource access
- Error handling and timeout management

### Data Models (`internal/models`)

#### Slide Models
```go
type SlideTheme string
type SlideContent struct {
    Title       string      `json:"title"`
    Content     string      `json:"content"`
    Theme       SlideTheme  `json:"theme"`
    SlideIndex  int         `json:"slideIndex"`
    AudioURL    string      `json:"audioUrl,omitempty"`
    GeneratedAt time.Time   `json:"generatedAt"`
}
```

#### Authentication Models
```go
type TokenInfo struct {
    AccessToken  string    `json:"accessToken"`
    RefreshToken string    `json:"refreshToken"`
    TokenType    string    `json:"tokenType"`
    ExpiresAt    time.Time `json:"expiresAt"`
    Scope        string    `json:"scope"`
}

type UserInfo struct {
    ID       int    `json:"id"`
    Name     string `json:"name"`
    Email    string `json:"email"`
    Language string `json:"language"`
}
```

## Security Features

### Authentication & Authorization
- OAuth2 integration with Backlog
- JWT token-based authentication
- CSRF protection with state parameters
- Secure token storage and validation

### Request Security
- CORS configuration for frontend integration
- Request validation and sanitization
- Rate limiting implementation
- Error message sanitization

### Data Protection
- Secure credential handling
- Environment-based configuration
- Token encryption and secure storage
- Input validation and SQL injection prevention

## Performance Optimizations

### Caching
- Audio file caching for TTS
- Project data caching
- Token validation caching
- Static file serving optimization

### Concurrency
- Goroutine-based request handling
- Concurrent slide generation
- WebSocket connection pooling
- Async audio processing

### Resource Management
- Connection pooling for external services
- Memory-efficient data structures
- Graceful shutdown procedures
- Resource cleanup and garbage collection

## Testing

The backend includes comprehensive test suites:

### Unit Tests
- Service layer testing
- Model validation testing
- Authentication flow testing
- MCP client testing

### Integration Tests
- API endpoint testing
- Database integration testing
- External service mocking
- End-to-end workflow testing

**Running Tests:**
```bash
cd backend
go test ./...
go test -cover ./...
```

## Deployment Configuration

### Environment Variables
```bash
PORT=8080
ENVIRONMENT=production
BACKLOG_DOMAIN=yourspace.backlog.jp
BACKLOG_CLIENT_ID=your_client_id
BACKLOG_CLIENT_SECRET=your_client_secret
AI_PROVIDER=openai  # or bedrock
OPENAI_API_KEY=your_openai_key
AWS_ACCESS_KEY_ID=your_aws_key
AWS_SECRET_ACCESS_KEY=your_aws_secret
AWS_REGION=us-east-1
```

### Docker Support
The backend includes Docker configuration for containerized deployment:
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o main cmd/main.go

FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
CMD ["./main"]
```

## Go Documentation

<iframe src="/files/go-docs.txt" width="100%" height="600px" style="border: 1px solid #ccc; border-radius: 4px;"></iframe>

[View Full Go Documentation](/files/go-docs.txt)