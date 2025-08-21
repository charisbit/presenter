<!--
.. title: API Documentation
.. slug: api-documentation
.. date: 2025-08-18
.. tags: api, documentation
.. category: 
.. link: 
.. description: Complete API reference for the Intelligent Presenter
.. type: text
-->

# API Documentation

This page provides comprehensive API documentation for the Intelligent Presenter backend services.

## REST API Endpoints

### Authentication Endpoints

#### `GET /api/v1/auth/login`
Initiates OAuth2 authentication flow with Backlog.

**Parameters:**
- `redirectUrl` (query): Frontend URL to redirect after authentication

**Response:**
```json
{
  "authUrl": "https://yourspace.backlog.jp/OAuth2AccessRequest.action?...",
  "state": "random-state-string"
}
```

#### `GET /api/v1/auth/callback`
Handles OAuth2 callback from Backlog.

**Parameters:**
- `code` (query): Authorization code from Backlog
- `state` (query): State parameter for CSRF protection

**Response:**
```json
{
  "token": "jwt-token",
  "user": {
    "id": 12345,
    "name": "User Name",
    "email": "user@example.com"
  }
}
```

#### `POST /api/v1/auth/refresh`
Refreshes authentication token.

**Headers:**
- `Authorization: Bearer <token>`

**Response:**
```json
{
  "token": "new-jwt-token"
}
```

#### `POST /api/v1/auth/logout`
Logs out the current user.

**Headers:**
- `Authorization: Bearer <token>`

**Response:**
```json
{
  "message": "Logged out successfully"
}
```

### Project Endpoints

#### `GET /api/v1/projects`
Retrieves list of accessible projects.

**Headers:**
- `Authorization: Bearer <token>`

**Response:**
```json
{
  "projects": [
    {
      "id": 12345,
      "name": "Project Name",
      "key": "PROJ",
      "description": "Project description"
    }
  ]
}
```

#### `GET /api/v1/projects/{projectId}`
Retrieves detailed project information.

**Headers:**
- `Authorization: Bearer <token>`

**Response:**
```json
{
  "project": {
    "id": 12345,
    "name": "Project Name",
    "description": "Detailed description",
    "startDate": "2023-01-01",
    "dueDate": "2023-12-31",
    "progress": 65.5
  }
}
```

### Slide Generation Endpoints

#### `POST /api/v1/slides/generate`
Generates presentation slides for a project.

**Headers:**
- `Authorization: Bearer <token>`

**Request Body:**
```json
{
  "projectId": 12345,
  "themes": ["project_overview", "project_progress"],
  "language": "en",
  "includeAudio": true
}
```

**Response:**
```json
{
  "sessionId": "slide-session-uuid",
  "status": "processing",
  "message": "Slide generation started"
}
```

#### `GET /api/v1/slides/{sessionId}/status`
Checks slide generation status.

**Headers:**
- `Authorization: Bearer <token>`

**Response:**
```json
{
  "sessionId": "slide-session-uuid",
  "status": "completed",
  "progress": 100,
  "slidesGenerated": 8,
  "downloadUrl": "/api/v1/slides/session-uuid/download"
}
```

### Speech Synthesis Endpoints

#### `POST /api/v1/speech/synthesize`
Synthesizes speech from text.

**Headers:**
- `Authorization: Bearer <token>`

**Request Body:**
```json
{
  "text": "Text to synthesize",
  "language": "en",
  "voice": "default"
}
```

**Response:**
```json
{
  "audioUrl": "/cache/audio/filename.wav",
  "duration": 5.2,
  "format": "wav"
}
```

## WebSocket API

### Slide Generation WebSocket

#### Connection
```
wss://localhost:8080/ws/slides/{sessionId}
```

**Headers:**
- `Authorization: Bearer <token>`

#### Message Types

**Slide Generated:**
```json
{
  "type": "slide_generated",
  "data": {
    "slideIndex": 0,
    "title": "Project Overview",
    "content": "# Project Overview\n\nProject content...",
    "audioUrl": "/cache/audio/slide-0.wav"
  }
}
```

**Progress Update:**
```json
{
  "type": "progress_update",
  "data": {
    "currentSlide": 3,
    "totalSlides": 8,
    "percentage": 37.5,
    "status": "Generating slide content..."
  }
}
```

**Generation Complete:**
```json
{
  "type": "generation_complete",
  "data": {
    "sessionId": "session-uuid",
    "totalSlides": 8,
    "downloadUrl": "/api/v1/slides/session-uuid/download"
  }
}
```

**Error:**
```json
{
  "type": "error",
  "data": {
    "message": "Error description",
    "code": "ERROR_CODE"
  }
}
```

## Data Models

### Project
```typescript
interface Project {
  id: number
  name: string
  key: string
  description: string
  startDate?: string
  dueDate?: string
  progress?: number
}
```

### SlideContent
```typescript
interface SlideContent {
  title: string
  content: string
  theme: SlideTheme
  slideIndex: number
  audioUrl?: string
  generatedAt: string
}
```

### SlideTheme
```typescript
type SlideTheme = 
  | "project_overview"
  | "project_progress" 
  | "issue_management"
  | "risk_analysis"
  | "team_overview"
  | "document_management"
```

## Error Codes

| Code | Description | HTTP Status |
|------|-------------|-------------|
| `AUTH_REQUIRED` | Authentication required | 401 |
| `AUTH_INVALID` | Invalid authentication token | 401 |
| `AUTH_EXPIRED` | Authentication token expired | 401 |
| `FORBIDDEN` | Insufficient permissions | 403 |
| `PROJECT_NOT_FOUND` | Project not found | 404 |
| `INVALID_REQUEST` | Invalid request data | 400 |
| `GENERATION_FAILED` | Slide generation failed | 500 |
| `SERVICE_UNAVAILABLE` | External service unavailable | 503 |

## Rate Limiting

- **Authentication endpoints**: 10 requests per minute per IP
- **Project endpoints**: 100 requests per minute per user
- **Slide generation**: 5 concurrent sessions per user
- **Speech synthesis**: 50 requests per minute per user

## Response Headers

All API responses include:
- `Content-Type: application/json`
- `X-Request-ID: unique-request-identifier`
- `X-Rate-Limit-Remaining: number`
- `X-Rate-Limit-Reset: timestamp`