<!--
.. title: Frontend Documentation (TypeScript)
.. slug: frontend-docs
.. date: 2025-08-18
.. tags: frontend, typescript, vue, documentation
.. category: 
.. link: 
.. description: TypeScript/Vue frontend documentation for the Intelligent Presenter
.. type: text
-->

# Frontend Documentation (TypeScript/Vue)

The Intelligent Presenter frontend is a modern Vue 3 application built with TypeScript, providing an intuitive interface for presentation generation and project management.

## Architecture Overview

The frontend follows Vue 3 composition API patterns with TypeScript for type safety:

```text
frontend/src/
├── components/          # Reusable Vue components
├── services/           # API clients and business logic
├── stores/             # Pinia state management
├── types/              # TypeScript type definitions
├── utils/              # Utility functions
├── views/              # Page components
├── router/             # Vue Router configuration
└── main.ts            # Application entry point
```

## Technology Stack

- **Framework**: Vue 3 with Composition API
- **Language**: TypeScript
- **State Management**: Pinia
- **Routing**: Vue Router 4
- **Build Tool**: Vite
- **HTTP Client**: Axios
- **UI Components**: Custom components
- **Styling**: CSS with modern features
- **Testing**: Vitest + Vue Test Utils

## Core Architecture

### Application Entry Point (`main.ts`)

The main application setup includes:
- Vue app initialization with TypeScript support
- Pinia store configuration
- Router setup with type-safe routing
- Global component registration
- Error handling configuration

```typescript
import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'

const app = createApp(App)
app.use(createPinia())
app.use(router)
app.mount('#app')
```

### Routing (`router/index.ts`)

**Route Configuration:**
- **Home** (`/`) - Welcome page and project selection
- **Login** (`/login`) - Authentication interface
- **Callback** (`/callback`) - OAuth callback handling
- **Project Selection** (`/projects`) - Project list and selection
- **Presentation** (`/presentation`) - Slide generation and viewing

**Router Features:**
- Type-safe route definitions
- Navigation guards for authentication
- Route-based code splitting
- Dynamic route parameters

### State Management (`stores/`)

#### Auth Store (`stores/auth.ts`)
Manages authentication state and user session:

```typescript
interface AuthState {
  user: UserInfo | null
  token: string | null
  isAuthenticated: boolean
  isLoading: boolean
}
```

**Key Actions:**
- `login()` - Initiate OAuth flow
- `handleCallback()` - Process OAuth callback
- `refreshToken()` - Refresh authentication token
- `logout()` - Clear session and redirect
- `getUserInfo()` - Fetch user profile

#### Slides Store (`stores/slides.ts`)
Manages presentation state and slide generation:

```typescript
interface SlidesState {
  currentSession: SlideSession | null
  slides: SlideContent[]
  isGenerating: boolean
  progress: GenerationProgress
  selectedProject: Project | null
}
```

**Key Actions:**
- `generateSlides()` - Start slide generation
- `connectWebSocket()` - Establish real-time connection
- `updateProgress()` - Track generation progress
- `addSlide()` - Add generated slide to collection
- `clearSession()` - Reset generation state

### API Services (`services/`)

#### API Client (`services/api.ts`)
Centralized HTTP client with TypeScript interfaces:

**Features:**
- Axios-based HTTP client with interceptors
- Automatic token injection
- Request/response type safety
- Error handling and retry logic
- WebSocket service integration

**API Modules:**
```typescript
export const authApi = {
  initiateOAuth: (redirectUrl: string) => Promise<OAuthInitResponse>
  handleCallback: (code: string, state: string) => Promise<AuthResponse>
  refreshToken: () => Promise<AuthResponse>
  logout: () => Promise<void>
  getUserInfo: () => Promise<UserInfo>
}

export const projectApi = {
  getProjects: () => Promise<Project[]>
  getProjectOverview: (projectId: ProjectID) => Promise<ProjectOverview>
  getProjectProgress: (projectId: ProjectID) => Promise<ProjectProgress>
  getProjectIssues: (projectId: ProjectID) => Promise<ProjectIssues>
}

export const slideApi = {
  generateSlides: (request: SlideGenerationRequest) => Promise<SlideGenerationResponse>
  getSlideStatus: (sessionId: string) => Promise<SlideStatus>
}
```

#### WebSocket Service
Real-time communication for slide generation:

```typescript
class WebSocketService {
  private socket: WebSocket | null = null
  private eventHandlers: Map<string, Function[]> = new Map()

  connect(sessionId: string, token: string): void
  disconnect(): void
  on(event: string, handler: Function): void
  off(event: string, handler: Function): void
  private handleMessage(event: MessageEvent): void
}
```

#### Slidev Service (`services/slidev.ts`)
Slide processing and compilation:

**SlidevProcessor Features:**
- Markdown to HTML conversion
- Chart.js integration
- Mermaid diagram support
- Theme application
- Asset optimization

```typescript
class SlidevProcessor {
  processSlideContent(content: string): Promise<string>
  compileSlides(slides: SlideContent[]): Promise<CompiledSlides>
  processChartConfig(config: ChartConfig): ChartConfiguration
  processMermaidDiagram(diagram: string): string
}
```

### Type Definitions (`types/`)

#### Authentication Types (`types/auth.ts`)
```typescript
interface UserInfo {
  id: number
  name: string
  email: string
  language: string
}

interface AuthResponse {
  token: string
  user: UserInfo
}

interface OAuthInitResponse {
  authUrl: string
  state: string
}
```

#### Slide Types (`types/slides.ts`)
```typescript
type SlideTheme = 
  | 'project_overview'
  | 'project_progress'
  | 'issue_management'
  | 'risk_analysis'
  | 'team_overview'
  | 'document_management'

interface SlideContent {
  title: string
  content: string
  theme: SlideTheme
  slideIndex: number
  audioUrl?: string
  generatedAt: string
}

interface SlideGenerationRequest {
  projectId: ProjectID
  themes: SlideTheme[]
  language: string
  includeAudio: boolean
}
```

#### Project Types (`types/index.ts`)
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

type ProjectID = number | string
```

### Utility Functions (`utils/`)

#### Slide Utilities (`utils/slideUtils.ts`)
Helper functions for slide processing:

```typescript
// Theme label mapping
export function getThemeLabel(theme: SlideTheme): string

// Code block processing
export function processBareMessageCode(content: string): string

// Chart.js configuration processing
export function processChartJSConfigs(content: string): string
```

## Component Architecture

### Views (Page Components)

#### HomeView
- Welcome interface and navigation
- Project selection overview
- Authentication status display

#### LoginView
- OAuth authentication interface
- Error handling and feedback
- Redirect management

#### ProjectSelectionView
- Project list display
- Search and filtering
- Project details preview

#### PresentationView
- Real-time slide generation interface
- Progress tracking and status updates
- Generated slide display and navigation
- Audio playback controls

#### CallbackView
- OAuth callback processing
- Loading states and error handling
- Automatic redirection

### Components

#### ChartComponent
Reusable chart rendering component:
- Chart.js integration
- Dynamic data binding
- Responsive design
- Type-safe configuration

## State Flow

### Authentication Flow
1. User initiates login → `authStore.login()`
2. OAuth URL generated → Redirect to Backlog
3. Callback processing → `authStore.handleCallback()`
4. Token storage → Update auth state
5. User info retrieval → `authStore.getUserInfo()`

### Slide Generation Flow
1. Project selection → Update `slidesStore.selectedProject`
2. Theme selection → Configure generation parameters
3. Generation initiation → `slidesStore.generateSlides()`
4. WebSocket connection → Real-time progress updates
5. Slide reception → Add to slides collection
6. Completion handling → Update UI state

## Development Tools

### TypeScript Configuration
Strict TypeScript configuration with:
- Strict null checks
- No implicit any
- Unused locals detection
- Import/export validation

### Build Configuration (Vite)
- Fast HMR development server
- Optimized production builds
- Asset optimization
- Environment variable handling

### Testing Setup (Vitest)
- Unit tests for stores and utilities
- Component testing with Vue Test Utils
- Mocking for external dependencies
- Coverage reporting

**Running Tests:**
```bash
npm run test          # Run tests
npm run test:ui       # Visual test interface
npm run test:coverage # Generate coverage report
```

### Code Quality
- ESLint configuration with Vue and TypeScript rules
- Prettier for code formatting
- Husky for pre-commit hooks
- Type checking in CI/CD

## Performance Optimizations

### Code Splitting
- Route-based lazy loading
- Component-level code splitting
- Dynamic imports for large dependencies

### State Management
- Computed properties for derived state
- Reactive state updates
- Efficient re-rendering

### Asset Optimization
- Image optimization and lazy loading
- CSS tree shaking
- Bundle size optimization
- Caching strategies

## TypeScript Documentation

<iframe src="/files/typescript-docs/index.html" width="100%" height="600px" style="border: 1px solid #ccc; border-radius: 4px;"></iframe>

[View Full TypeScript Documentation](/files/typescript-docs/index.html)

## Development Guidelines

### Component Development
- Use Composition API with `<script setup>`
- Implement proper TypeScript interfaces
- Follow Vue 3 best practices
- Include comprehensive prop validation

### Store Development
- Use Pinia with TypeScript
- Implement proper action error handling
- Maintain immutable state updates
- Document store interfaces

### API Integration
- Use typed API client methods
- Implement proper error boundaries
- Handle loading and error states
- Maintain request/response type safety