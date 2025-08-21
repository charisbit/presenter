/**
 * @fileoverview API client and WebSocket service for the Intelligent Presenter application.
 * 
 * This module provides:
 * - HTTP client configuration with automatic authentication
 * - API endpoints for authentication, slides, projects, and speech
 * - WebSocket service for real-time communication
 * - Request/response interceptors for error handling
 * - Automatic token management and refresh
 * 
 * @author Nulab Technical Challenge
 * @version 1.0.0
 */

import axios, { type AxiosInstance } from 'axios'
import type { AuthResponse, OAuthInitResponse, UserInfo } from '@/types/auth'
import type { SlideGenerationRequest, SlideGenerationResponse } from '@/types/slides'
import type { Project } from '@/types'

/**
 * Create and configure the main Axios HTTP client instance.
 * 
 * Configuration includes:
 * - Base URL from environment variables
 * - 30-second timeout for requests
 * - JSON content type headers
 * - Automatic authentication token injection
 * - Response error handling
 */
const api: AxiosInstance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json'
  }
})

/**
 * Request interceptor to automatically inject authentication tokens.
 * 
 * This interceptor:
 * 1. Retrieves the stored JWT token from localStorage
 * 2. Adds the token to the Authorization header for all requests
 * 3. Ensures authenticated API calls work seamlessly
 * 
 * @param config - Axios request configuration object
 * @returns Modified request configuration with authentication header
 */
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('auth_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

/**
 * Response interceptor for global error handling and authentication management.
 * 
 * This interceptor:
 * 1. Passes through successful responses unchanged
 * 2. Handles 401 Unauthorized responses by clearing tokens and redirecting to login
 * 3. Provides centralized error handling for authentication failures
 * 
 * @param response - Successful HTTP response (passed through)
 * @param error - HTTP error response object
 * @returns Promise resolving to response or rejecting with error
 */
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // Token expired or invalid, redirect to login
      localStorage.removeItem('auth_token')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

/**
 * Authentication API endpoints for OAuth 2.0 flow with Backlog.
 * 
 * This object provides methods for:
 * - Initiating OAuth authentication
 * - Handling OAuth callbacks
 * - Managing user sessions
 * - Token refresh and logout
 * 
 * @namespace authApi
 */
export const authApi = {
  /**
   * Initiates OAuth 2.0 authentication flow with Backlog.
   * 
   * @returns {Promise<OAuthInitResponse>} Object containing authorization URL and state parameter
   * @throws {Error} If the authentication initiation fails
   * 
   * @example
   * ```typescript
   * const { authUrl, state } = await authApi.initiateOAuth()
   * window.location.href = authUrl // Redirect user to Backlog login
   * ```
   */
  async initiateOAuth(): Promise<OAuthInitResponse> {
    const response = await api.get('/api/v1/auth/login')
    return response.data
  },

  /**
   * Handles OAuth callback from Backlog and exchanges code for tokens.
   * 
   * @param {string} code - Authorization code from Backlog OAuth
   * @param {string} state - CSRF protection state parameter
   * @returns {Promise<AuthResponse>} Complete authentication response with token and user info
   * @throws {Error} If the callback handling fails or parameters are invalid
   * 
   * @example
   * ```typescript
   * const authResponse = await authApi.handleCallback('auth_code_123', 'state_token_456')
   * localStorage.setItem('auth_token', authResponse.token)
   * ```
   */
  async handleCallback(code: string, state: string): Promise<AuthResponse> {
    const response = await api.get(`/api/v1/auth/callback?code=${code}&state=${state}`)
    return response.data
  },

  /**
   * Retrieves current user information from the API.
   * 
   * @returns {Promise<UserInfo>} Complete user profile information
   * @throws {Error} If user is not authenticated or API call fails
   * 
   * @example
   * ```typescript
   * const user = await authApi.getUserInfo()
   * console.log(`Welcome, ${user.name}!`)
   * ```
   */
  async getUserInfo(): Promise<UserInfo> {
    const response = await api.get('/api/v1/auth/me')
    return response.data
  },

  /**
   * Refreshes the current authentication token.
   * 
   * @returns {Promise<AuthResponse>} New authentication response with refreshed token
   * @throws {Error} If token refresh fails or user needs to re-authenticate
   * 
   * @example
   * ```typescript
   * const newAuth = await authApi.refreshToken()
   * localStorage.setItem('auth_token', newAuth.token)
   * ```
   */
  async refreshToken(): Promise<AuthResponse> {
    const response = await api.post('/api/v1/auth/refresh')
    return response.data
  },

  /**
   * Logs out the current user and invalidates their session.
   * 
   * @returns {Promise<void>} Promise that resolves when logout is complete
   * @throws {Error} If logout API call fails
   * 
   * @example
   * ```typescript
   * await authApi.logout()
   * localStorage.removeItem('auth_token')
   * router.push('/login')
   * ```
   */
  async logout(): Promise<void> {
    await api.post('/api/v1/auth/logout')
  }
}

/**
 * Slide generation and management API endpoints.
 * 
 * This object provides methods for:
 * - Initiating slide generation from Backlog projects
 * - Monitoring generation progress and status
 * - Retrieving completed slide content
 * 
 * @namespace slideApi
 */
export const slideApi = {
  /**
   * Initiates slide generation for a Backlog project with specified themes.
   * 
   * @param {SlideGenerationRequest} request - Configuration for slide generation
   * @param {string} request.projectId - Backlog project identifier
   * @param {SlideTheme[]} request.themes - Array of slide themes to generate
   * @param {string} request.language - Target language code ('ja' or 'en')
   * @returns {Promise<SlideGenerationResponse>} Generation response with slideId and WebSocket URL
   * @throws {Error} If generation initiation fails or request is invalid
   * 
   * @example
   * ```typescript
   * const response = await slideApi.generateSlides({
   *   projectId: '216125',
   *   themes: ['project_overview', 'issue_management'],
   *   language: 'ja'
   * })
   * console.log(`Generation started: ${response.slideId}`)
   * ```
   */
  async generateSlides(request: SlideGenerationRequest): Promise<SlideGenerationResponse> {
    const response = await api.post('/api/v1/slides/generate', request)
    return response.data
  },

  /**
   * Retrieves the current status of a slide generation session.
   * 
   * @param {string} slideId - Unique identifier for the generation session
   * @returns {Promise<any>} Current status information and progress details
   * @throws {Error} If slideId is invalid or API call fails
   * 
   * @example
   * ```typescript
   * const status = await slideApi.getSlideStatus('slide_123456')
   * console.log(`Progress: ${status.progress}%`)
   * ```
   */
  async getSlideStatus(slideId: string): Promise<any> {
    const response = await api.get(`/api/v1/slides/${slideId}/status`)
    return response.data
  }
}

/**
 * Backlog project data API endpoints.
 * 
 * This object provides methods for:
 * - Retrieving available projects for the authenticated user
 * - Fetching detailed project information and analytics
 * - Getting project-specific data for slide generation
 * 
 * @namespace projectApi
 */
export const projectApi = {
  /**
   * Retrieves all Backlog projects accessible to the current user.
   * 
   * @returns {Promise<Project[]>} Array of project objects with basic information
   * @throws {Error} If user is not authenticated or API call fails
   * 
   * @example
   * ```typescript
   * const projects = await projectApi.getProjects()
   * projects.forEach(project => {
   *   console.log(`${project.name} (${project.key})`)
   * })
   * ```
   */
  async getProjects(): Promise<Project[]> {
    const response = await api.get('/api/v1/projects')
    return response.data
  },

  /**
   * Retrieves comprehensive overview information for a specific project.
   * 
   * @param {string} projectId - Unique project identifier
   * @returns {Promise<any>} Project overview data including summary statistics
   * @throws {Error} If projectId is invalid or user lacks access
   * 
   * @example
   * ```typescript
   * const overview = await projectApi.getProjectOverview('216125')
   * console.log(`Project has ${overview.totalIssues} issues`)
   * ```
   */
  async getProjectOverview(projectId: string): Promise<any> {
    const response = await api.get(`/api/v1/projects/${projectId}/overview`)
    return response.data
  },

  /**
   * Retrieves project progress and completion metrics.
   * 
   * @param {string} projectId - Unique project identifier
   * @returns {Promise<any>} Progress data including completion rates and milestones
   * @throws {Error} If projectId is invalid or user lacks access
   * 
   * @example
   * ```typescript
   * const progress = await projectApi.getProjectProgress('216125')
   * console.log(`Completion rate: ${progress.completionRate}%`)
   * ```
   */
  async getProjectProgress(projectId: string): Promise<any> {
    const response = await api.get(`/api/v1/projects/${projectId}/progress`)
    return response.data
  },

  /**
   * Retrieves project issue statistics and management data.
   * 
   * @param {string} projectId - Unique project identifier
   * @returns {Promise<any>} Issue data including counts, priorities, and status distributions
   * @throws {Error} If projectId is invalid or user lacks access
   * 
   * @example
   * ```typescript
   * const issues = await projectApi.getProjectIssues('216125')
   * console.log(`Open issues: ${issues.openCount}`)
   * ```
   */
  async getProjectIssues(projectId: string): Promise<any> {
    const response = await api.get(`/api/v1/projects/${projectId}/issues`)
    return response.data
  },

  /**
   * Retrieves project team information and collaboration metrics.
   * 
   * @param {string} projectId - Unique project identifier
   * @returns {Promise<any>} Team data including member lists and activity statistics
   * @throws {Error} If projectId is invalid or user lacks access
   * 
   * @example
   * ```typescript
   * const team = await projectApi.getProjectTeam('216125')
   * console.log(`Team size: ${team.members.length}`)
   * ```
   */
  async getProjectTeam(projectId: string): Promise<any> {
    const response = await api.get(`/api/v1/projects/${projectId}/team`)
    return response.data
  },

  /**
   * Retrieves project risk analysis and mitigation data.
   * 
   * @param {string} projectId - Unique project identifier
   * @returns {Promise<any>} Risk data including identified risks and mitigation strategies
   * @throws {Error} If projectId is invalid or user lacks access
   * 
   * @example
   * ```typescript
   * const risks = await projectApi.getProjectRisks('216125')
   * console.log(`High priority risks: ${risks.highPriorityCount}`)
   * ```
   */
  async getProjectRisks(projectId: string): Promise<any> {
    const response = await api.get(`/api/v1/projects/${projectId}/risks`)
    return response.data
  }
}

/**
 * Text-to-speech synthesis API endpoints.
 * 
 * This object provides methods for:
 * - Converting slide text content to speech audio
 * - Supporting multiple languages and voices
 * - Managing audio file generation and storage
 * 
 * @namespace speechApi
 */
export const speechApi = {
  /**
   * Converts text content to speech audio using TTS service.
   * 
   * @param {string} text - Text content to convert to speech
   * @param {string} language - Language code for speech synthesis (e.g., 'ja', 'en')
   * @param {string} [voice] - Optional specific voice to use for synthesis
   * @returns {Promise<{audioUrl: string}>} Object containing URL to generated audio file
   * @throws {Error} If synthesis fails or text is invalid
   * 
   * @example
   * ```typescript
   * const audio = await speechApi.synthesizeSpeech(
   *   'こんにちは、プレゼンテーションを開始します。',
   *   'ja',
   *   'female'
   * )
   * audioElement.src = audio.audioUrl
   * ```
   */
  async synthesizeSpeech(text: string, language: string, voice?: string): Promise<{ audioUrl: string }> {
    const response = await api.post('/api/v1/speech/synthesize', {
      text,
      language,
      voice,
      streaming: false
    })
    return response.data
  }
}

/**
 * WebSocket service for real-time communication with the backend.
 * 
 * This class provides:
 * - Automatic connection management and reconnection
 * - Event-driven message handling
 * - Exponential backoff for reconnection attempts
 * - Clean connection lifecycle management
 * 
 * @class WebSocketService
 * @example
 * ```typescript
 * const wsService = new WebSocketService()
 * wsService.connect('/ws/slides/123', {
 *   onMessage: (data) => console.log('Received:', data),
 *   onOpen: () => console.log('Connected'),
 *   onClose: () => console.log('Disconnected')
 * })
 * ```
 */
export class WebSocketService {
  /** Active WebSocket connection instance */
  private ws: WebSocket | null = null
  
  /** Current number of reconnection attempts */
  private reconnectAttempts = 0
  
  /** Maximum number of reconnection attempts before giving up */
  private maxReconnectAttempts = 5
  
  /** Base interval for reconnection attempts (ms) */
  private reconnectInterval = 1000

  /**
   * Establishes a WebSocket connection to the specified endpoint.
   * 
   * @param {string} endpoint - WebSocket endpoint path (e.g., '/ws/slides/123')
   * @param {Object} handlers - Event handlers for WebSocket lifecycle
   * @param {Function} [handlers.onOpen] - Called when connection is established
   * @param {Function} [handlers.onClose] - Called when connection is closed
   * @param {Function} [handlers.onMessage] - Called when message is received
   * @param {Function} [handlers.onError] - Called when error occurs
   * 
   * @example
   * ```typescript
   * websocketService.connect('/ws/slides/abc123', {
   *   onOpen: () => console.log('Connected to slide generation'),
   *   onMessage: (data) => {
   *     if (data.type === 'slide_content') {
   *       // Handle new slide content
   *     }
   *   },
   *   onError: (error) => console.error('WebSocket error:', error)
   * })
   * ```
   */
  connect(endpoint: string, handlers: {
    onOpen?: () => void
    onClose?: () => void
    onMessage?: (data: any) => void
    onError?: (error: Event) => void
  }) {
    // Construct WebSocket URL from environment or current location
    const wsUrl = (import.meta.env.VITE_WS_URL || (
      (window.location.protocol === 'https:' ? 'wss://' : 'ws://') + window.location.host
    )) + endpoint
    
    try {
      this.ws = new WebSocket(wsUrl)

      // Handle successful connection
      this.ws.onopen = () => {
        console.log('WebSocket connected')
        this.reconnectAttempts = 0
        handlers.onOpen?.()
      }

      // Handle connection closure
      this.ws.onclose = () => {
        console.log('WebSocket disconnected')
        handlers.onClose?.()
        this.attemptReconnect(endpoint, handlers)
      }

      // Handle incoming messages
      this.ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data)
          handlers.onMessage?.(data)
        } catch (error) {
          console.error('Failed to parse WebSocket message:', error)
        }
      }

      // Handle connection errors
      this.ws.onerror = (error) => {
        console.error('WebSocket error:', error)
        handlers.onError?.(error)
      }
    } catch (error) {
      console.error('Failed to create WebSocket connection:', error)
      handlers.onError?.(error as Event)
    }
  }

  /**
   * Attempts to reconnect to the WebSocket with exponential backoff.
   * 
   * @private
   * @param {string} endpoint - Original endpoint to reconnect to
   * @param {any} handlers - Original event handlers
   * 
   * Uses exponential backoff strategy:
   * - Attempt 1: 1 second delay
   * - Attempt 2: 2 second delay  
   * - Attempt 3: 4 second delay
   * - Attempt 4: 8 second delay
   * - Attempt 5: 16 second delay
   */
  private attemptReconnect(endpoint: string, handlers: any) {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      setTimeout(() => {
        console.log(`Attempting to reconnect WebSocket (${this.reconnectAttempts + 1}/${this.maxReconnectAttempts})`)
        this.reconnectAttempts++
        this.connect(endpoint, handlers)
      }, this.reconnectInterval * Math.pow(2, this.reconnectAttempts))
    }
  }

  /**
   * Sends data through the WebSocket connection.
   * 
   * @param {any} data - Data to send (will be JSON stringified)
   * 
   * @example
   * ```typescript
   * websocketService.send({
   *   type: 'request_status',
   *   slideId: 'abc123'
   * })
   * ```
   */
  send(data: any) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(data))
    }
  }

  /**
   * Closes the WebSocket connection and cleans up resources.
   * 
   * @example
   * ```typescript
   * // Clean disconnect when component unmounts
   * onUnmounted(() => {
   *   websocketService.disconnect()
   * })
   * ```
   */
  disconnect() {
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
  }
}

/**
 * Singleton WebSocket service instance for use throughout the application.
 * This instance manages all real-time communication with the backend.
 * 
 * @type {WebSocketService}
 */
export const websocketService = new WebSocketService()

/**
 * Default export of the configured Axios API client.
 * This client is used for all HTTP requests in the application.
 * 
 * @type {AxiosInstance}
 */
export default api