<!--
.. title: Development Guide
.. slug: development
.. date: 2025-08-18
.. tags: development, setup, contributing
.. category: 
.. link: 
.. description: Development setup and contributing guide
.. type: text
-->

# Development Guide

This guide helps you set up a local development environment and contribute to the Intelligent Presenter project.

## Prerequisites

### Required Software
- **Go**: 1.21 or later
- **Node.js**: 18 or later
- **npm**: 9 or later
- **Git**: Latest version
- **Docker**: 20.10+ (optional but recommended)

### System Requirements
- **OS**: macOS, Linux, or Windows with WSL2
- **RAM**: 8 GB minimum, 16 GB recommended
- **Storage**: 10 GB free space

## Project Setup

### 1. Clone the Repository
```bash
git clone https://github.com/your-org/intelligent-presenter.git
cd intelligent-presenter
```

### 2. Environment Configuration
```bash
# Copy environment template
cp .env.example .env

# Edit configuration
nano .env
```

### 3. Required Environment Variables
```bash
# Backlog Configuration
BACKLOG_DOMAIN=yourspace.backlog.jp
BACKLOG_CLIENT_ID=your_client_id
BACKLOG_CLIENT_SECRET=your_client_secret

# AI Provider (choose one)
AI_PROVIDER=openai
OPENAI_API_KEY=your_openai_key

# OR for AWS Bedrock
AI_PROVIDER=bedrock
AWS_ACCESS_KEY_ID=your_aws_key
AWS_SECRET_ACCESS_KEY=your_aws_secret
AWS_REGION=us-east-1

# Development URLs
FRONTEND_URL=http://localhost:3000
OAUTH_REDIRECT_URL=http://localhost:8080/api/v1/auth/callback
```

### 4. Backend Setup
```bash
cd backend

# Install Go dependencies
go mod download

# Run tests to verify setup
go test ./...

# Start development server
go run cmd/main.go
```

The backend will start on `http://localhost:8080`

### 5. Frontend Setup
```bash
cd frontend

# Install Node.js dependencies
npm install

# Run tests
npm run test

# Start development server
npm run dev
```

The frontend will start on `http://localhost:3000`

### 6. Supporting Services

#### Backlog MCP Server
```bash
cd backlog-server
go mod download
go run main.go
```

#### Speech Server (Optional)
```bash
cd speech-server
go mod download
go run cmd/main.go
```

## Development Workflow

### Branch Strategy
- **main**: Production-ready code
- **develop**: Integration branch for features
- **feature/***: Feature development branches
- **fix/***: Bug fix branches
- **release/***: Release preparation branches

### Creating a Feature Branch
```bash
git checkout develop
git pull origin develop
git checkout -b feature/your-feature-name
```

### Making Changes
1. **Write Tests First**: Follow TDD practices
2. **Implement Feature**: Write clean, documented code
3. **Run Tests**: Ensure all tests pass
4. **Update Documentation**: Keep docs current

### Running Tests

#### Backend Tests
```bash
cd backend

# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test package
go test ./internal/services

# Run with verbose output
go test -v ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

#### Frontend Tests
```bash
cd frontend

# Run unit tests
npm run test

# Run tests in watch mode
npm run test:watch

# Run tests with coverage
npm run test:coverage

# Run tests with UI
npm run test:ui

# Type checking
npm run type-check

# Linting
npm run lint
```

#### End-to-End Tests
```bash
# Run E2E tests with Playwright
cd e2e-tests
npm install
npm run test:e2e
```

### Code Quality

#### Backend Code Style
- Follow Go conventions and `gofmt` formatting
- Use meaningful variable and function names
- Write comprehensive comments for exported functions
- Handle errors explicitly
- Use interfaces for abstraction

#### Frontend Code Style
- Follow TypeScript strict mode guidelines
- Use Vue 3 Composition API patterns
- Implement proper component prop validation
- Write JSDoc comments for complex functions
- Use semantic HTML and ARIA attributes

#### Linting and Formatting
```bash
# Backend formatting
cd backend
go fmt ./...
go vet ./...

# Frontend linting
cd frontend
npm run lint
npm run lint:fix
npm run format
```

## Architecture Guidelines

### Backend Architecture
- **Clean Architecture**: Separate concerns into layers
- **Dependency Injection**: Use interfaces for loose coupling
- **Error Handling**: Consistent error response patterns
- **Logging**: Structured logging with context
- **Testing**: Unit tests for business logic, integration tests for APIs

### Frontend Architecture
- **Component Structure**: Single responsibility principle
- **State Management**: Use Pinia for global state
- **Type Safety**: Comprehensive TypeScript coverage
- **Performance**: Lazy loading and code splitting
- **Accessibility**: WCAG 2.1 compliance

## API Development

### Adding New Endpoints

#### 1. Define Models
```go
// internal/models/your_model.go
type YourModel struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}
```

#### 2. Create Service
```go
// internal/services/your_service.go
type YourService struct {
    config *config.Config
}

func NewYourService(cfg *config.Config) *YourService {
    return &YourService{config: cfg}
}

func (s *YourService) GetYourData() (*YourModel, error) {
    // Implementation
}
```

#### 3. Implement Handler
```go
// internal/api/handlers/your_handler.go
type YourHandler struct {
    service *services.YourService
}

func NewYourHandler(cfg *config.Config) *YourHandler {
    return &YourHandler{
        service: services.NewYourService(cfg),
    }
}

func (h *YourHandler) GetYourData(c *gin.Context) {
    data, err := h.service.GetYourData()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, data)
}
```

#### 4. Add Routes
```go
// internal/api/routes.go
func SetupRoutes(router *gin.Engine, cfg *config.Config) {
    yourHandler := handlers.NewYourHandler(cfg)
    
    v1 := router.Group("/api/v1")
    v1.GET("/your-endpoint", yourHandler.GetYourData)
}
```

#### 5. Write Tests
```go
// tests/your_handler_test.go
func TestYourHandler_GetYourData(t *testing.T) {
    // Test implementation
}
```

### Frontend API Integration

#### 1. Define Types
```typescript
// src/types/your-types.ts
export interface YourModel {
  id: number
  name: string
}
```

#### 2. Create API Client
```typescript
// src/services/api.ts
export const yourApi = {
  getYourData: (): Promise<YourModel> => {
    return apiClient.get<YourModel>('/your-endpoint').then(response => response.data)
  }
}
```

#### 3. Add to Store
```typescript
// src/stores/your-store.ts
export const useYourStore = defineStore('your', () => {
  const data = ref<YourModel | null>(null)
  
  const fetchData = async () => {
    try {
      data.value = await yourApi.getYourData()
    } catch (error) {
      console.error('Failed to fetch data:', error)
    }
  }
  
  return { data, fetchData }
})
```

## Testing Guidelines

### Backend Testing

#### Unit Tests
```go
func TestYourService_Method(t *testing.T) {
    // Arrange
    cfg := &config.Config{/* test config */}
    service := NewYourService(cfg)
    
    // Act
    result, err := service.Method()
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, result)
}
```

#### Integration Tests
```go
func TestYourHandler_Integration(t *testing.T) {
    // Setup test server
    router := gin.New()
    SetupRoutes(router, testConfig)
    
    // Create test request
    req := httptest.NewRequest("GET", "/api/v1/endpoint", nil)
    w := httptest.NewRecorder()
    
    // Execute request
    router.ServeHTTP(w, req)
    
    // Verify response
    assert.Equal(t, http.StatusOK, w.Code)
}
```

### Frontend Testing

#### Component Tests
```typescript
// tests/components/YourComponent.test.ts
import { mount } from '@vue/test-utils'
import YourComponent from '@/components/YourComponent.vue'

describe('YourComponent', () => {
  it('renders correctly', () => {
    const wrapper = mount(YourComponent, {
      props: { title: 'Test Title' }
    })
    
    expect(wrapper.text()).toContain('Test Title')
  })
})
```

#### Store Tests
```typescript
// tests/stores/your-store.test.ts
import { setActivePinia, createPinia } from 'pinia'
import { useYourStore } from '@/stores/your-store'

describe('Your Store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })
  
  it('fetches data correctly', async () => {
    const store = useYourStore()
    await store.fetchData()
    expect(store.data).toBeDefined()
  })
})
```

## Debugging

### Backend Debugging

#### Using Delve Debugger
```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Start debugging session
dlv debug cmd/main.go
```

#### Logging
```go
import "log/slog"

func YourFunction() {
    slog.Info("Processing request", "id", requestID)
    slog.Error("Failed to process", "error", err)
}
```

### Frontend Debugging

#### Browser DevTools
- Use Vue DevTools extension
- Monitor network requests
- Check console for errors
- Inspect component state

#### VS Code Debugging
```json
// .vscode/launch.json
{
  "type": "node",
  "request": "launch",
  "name": "Debug Frontend",
  "program": "${workspaceFolder}/frontend/src/main.ts",
  "outFiles": ["${workspaceFolder}/frontend/dist/**/*.js"]
}
```

## Contributing

### Pull Request Process

1. **Create Feature Branch**
```bash
git checkout -b feature/your-feature
```

2. **Make Changes and Test**
```bash
# Backend tests
cd backend && go test ./...

# Frontend tests
cd frontend && npm run test && npm run type-check
```

3. **Commit Changes**
```bash
git add .
git commit -m "feat: add your feature description"
```

4. **Push and Create PR**
```bash
git push origin feature/your-feature
# Create PR on GitHub
```

### Commit Message Convention
```
type(scope): description

feat: add new feature
fix: fix bug
docs: update documentation
style: formatting changes
refactor: code refactoring
test: add tests
chore: maintenance tasks
```

### Code Review Checklist
- [ ] Code follows project conventions
- [ ] Tests are included and passing
- [ ] Documentation is updated
- [ ] No security vulnerabilities
- [ ] Performance considerations addressed
- [ ] Error handling is comprehensive

## Documentation

### Updating Documentation
```bash
# Generate Go documentation
cd backend && go doc -all > ../docs/go-docs.txt

# Generate TypeScript documentation
cd frontend && npm run docs

# Build Nikola documentation site
cd docs-site && nikola build
```

### Writing Documentation
- Use clear, concise language
- Include code examples
- Document edge cases and limitations
- Keep README files current
- Add inline comments for complex logic

## Troubleshooting

### Common Development Issues

#### Backend Issues
```bash
# Module not found
go mod tidy

# Port already in use
lsof -ti:8080 | xargs kill -9

# Permission denied
chmod +x scripts/setup.sh
```

#### Frontend Issues
```bash
# Node modules issues
rm -rf node_modules package-lock.json
npm install

# Type errors
npm run type-check

# Build failures
npm run build --verbose
```

### Getting Help
- Check existing GitHub issues
- Review documentation thoroughly
- Ask questions in team channels
- Create detailed bug reports with:
  - Steps to reproduce
  - Expected vs actual behavior
  - Environment details
  - Error messages and logs