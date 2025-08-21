# Intelligent Presenter

An AI-powered presentation generator that creates professional slides from Backlog project data using advanced language models and real-time visualization.

## Overview

The Intelligent Presenter system automatically generates comprehensive presentation slides by analyzing Backlog project data. It leverages AI technologies to create executive-ready slides with integrated charts, diagrams, and narration capabilities.

### Key Features

- **AI-Powered Content Generation**: Uses OpenAI GPT and AWS Bedrock for intelligent slide content creation
- **Multi-Theme Support**: 10 specialized themes covering all aspects of project management
- **Real-Time Generation**: WebSocket-based live updates during slide generation
- **Visual Integration**: Automatic Mermaid diagrams and Chart.js visualizations
- **Audio Narration**: Text-to-speech synthesis for presentation delivery
- **Multi-Language Support**: Japanese and English content generation
- **Executive Format**: Optimized for management reporting and stakeholder presentations

## Architecture

### System Components

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Frontend      │    │    Backend      │    │  External APIs  │
│   (Vue.js)      │◄──►│    (Go)         │◄──►│                 │
│                 │    │                 │    │ • Backlog API   │
│ • Slide Display │    │ • Slide Service │    │ • OpenAI API    │
│ • Navigation    │    │ • MCP Service   │    │ • AWS Bedrock   │
│ • Chart.js      │    │ • AI Integration│    │ • TTS Services  │
│ • Mermaid       │    │ • WebSocket     │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Technology Stack

**Frontend:**
- Vue.js 3 with Composition API
- Pinia for state management
- TypeScript for type safety
- Chart.js for data visualization
- Mermaid for diagram rendering
- Vitest for testing

**Backend:**
- Go 1.19+ with Gin framework
- WebSocket for real-time communication
- AWS Bedrock and OpenAI integration
- MCP (Model Context Protocol) for Backlog access
- Custom TTS integration

**Infrastructure:**
- Docker containerization
- RESTful API design
- Real-time WebSocket communication
- Multi-provider AI fallback system

## Slide Themes

The system supports 10 specialized themes:

1. **Project Overview** - Basic project information and objectives
2. **Project Progress** - Completion rates and milestone tracking
3. **Issue Management** - Issue tracking and resolution metrics
4. **Risk Analysis** - Risk identification and mitigation strategies
5. **Team Collaboration** - Team activities and communication patterns
6. **Document Management** - Documentation and knowledge sharing status
7. **Codebase Activity** - Development metrics and code quality indicators
8. **Notifications** - Communication efficiency and information flow
9. **Predictive Analysis** - Forecasts and trend analysis
10. **Summary & Planning** - Project summaries and future recommendations

## Getting Started

### Prerequisites

- Node.js 18+
- Go 1.19+
- Docker (optional)
- Backlog API access
- OpenAI API key or AWS credentials

### Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd intelligent-presenter
   ```

2. **Setup Backend**
   ```bash
   cd backend
   go mod download
   cp .env.example .env
   # Configure your API keys in .env
   go run main.go
   ```

3. **Setup Frontend**
   ```bash
   cd frontend
   npm install
   npm run dev
   ```

4. **Setup MCP Services**
   ```bash
   cd backlog-server
   npm install
   npm start
   
   cd ../speech-server
   npm install
   npm start
   ```

### Configuration

Create a `.env` file in the backend directory:

```env
# API Configuration
OPENAI_API_KEY=your_openai_key
AWS_ACCESS_KEY_ID=your_aws_key
AWS_SECRET_ACCESS_KEY=your_aws_secret
AWS_REGION=us-east-1

# Service URLs
MCP_BACKLOG_URL=http://localhost:8080
MCP_SPEECH_URL=http://localhost:8081

# Server Configuration
PORT=8000
CORS_ORIGINS=http://localhost:3000
```

## API Documentation

### REST Endpoints

#### Generate Slides
```http
POST /api/slides/generate
Content-Type: application/json

{
  "projectId": "PROJECT_123",
  "themes": ["project_overview", "project_progress"],
  "language": "ja"
}
```

Response:
```json
{
  "slideId": "uuid-string",
  "status": "generating",
  "websocketUrl": "ws://localhost:8000/ws/slides/uuid-string"
}
```

#### Get Slide Status
```http
GET /api/slides/{slideId}/status
```

### WebSocket Events

Connect to: `ws://localhost:8000/ws/slides/{slideId}`

**Incoming Messages:**
- `slide_content` - New slide generated
- `slide_narration` - Narration text generated
- `slide_audio` - Audio file generated
- `presentation_complete` - All slides completed
- `error` - Generation error occurred

## Testing

### Backend Tests (Go)
```bash
cd backend
go test ./tests/...
```

### Frontend Tests (Vitest)
```bash
cd frontend
npm run test
npm run test:coverage
```

### E2E Tests
```bash
npm run test:e2e
```

## Documentation Generation

### Generate API Documentation
```bash
# Go documentation
cd backend
godoc -http=:6060

# TypeScript documentation
cd frontend
npx typedoc --out docs src/
```

### Multi-Language Documentation
Documentation is available in three languages:
- English: `/docs/en/`
- Japanese: `/docs/ja/`
- Chinese: `/docs/zh/`

## Development

### Project Structure
```
intelligent-presenter/
├── backend/                 # Go backend service
│   ├── internal/           # Internal packages
│   ├── pkg/               # Public packages
│   ├── tests/             # Test files
│   └── main.go
├── frontend/               # Vue.js frontend
│   ├── src/               # Source code
│   ├── tests/             # Test files
│   └── package.json
├── backlog-server/         # Backlog MCP server
├── speech-server/          # Speech synthesis server
└── docs/                  # Documentation
```

### Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

### Code Style

- **Go**: Follow standard Go conventions and use `gofmt`
- **TypeScript**: Use ESLint and Prettier configurations
- **Commits**: Use conventional commit messages

## Deployment

### Docker Deployment
```bash
docker-compose up -d
```

### Production Considerations

- Set up proper environment variables
- Configure CORS origins
- Set up SSL/TLS certificates
- Monitor API rate limits
- Implement proper logging
- Set up health checks

## Troubleshooting

### Common Issues

1. **API Rate Limits**: The system implements automatic fallback between AI providers
2. **Mermaid Rendering**: Ensure proper code block syntax with triple backticks
3. **Chart.js Issues**: Verify JSON configuration format in code blocks
4. **WebSocket Disconnects**: Check network connectivity and authentication

### Debugging

- Enable debug logging in backend: `LOG_LEVEL=debug`
- Use browser developer tools for frontend debugging
- Check MCP server logs for data retrieval issues

## License

This project is licensed under the MIT License. See LICENSE file for details.

## Support

For technical support or questions:
- Check the troubleshooting guide
- Review API documentation
- Submit issues on GitHub
- Contact the development team