/**
 * @fileoverview Type definitions for slide presentation functionality.
 * This module contains all TypeScript interfaces and types used throughout
 * the slide generation and presentation system.
 */

/**
 * Available slide themes that can be generated from project data.
 * Each theme focuses on different aspects of project management and analysis.
 * 
 * @example
 * ```typescript
 * const themes: SlideTheme[] = ['project_overview', 'team_collaboration']
 * ```
 */
export type SlideTheme = 
  | 'project_overview'      // Basic project information and objectives
  | 'project_progress'      // Completion rates and milestone tracking
  | 'issue_management'      // Issue tracking and resolution metrics
  | 'risk_analysis'         // Risk identification and mitigation
  | 'team_collaboration'    // Team activities and collaboration patterns
  | 'document_management'   // Documentation and knowledge sharing
  | 'codebase_activity'     // Development metrics and code quality
  | 'notifications'         // Communication efficiency and flow
  | 'predictive_analysis'   // Forecasts and trend analysis
  | 'summary_plan'          // Project summary and future planning

/**
 * Request payload for initiating slide generation.
 * 
 * @interface SlideGenerationRequest
 * @property projectId - Backlog project identifier (string or numeric ID)
 * @property themes - Array of slide themes to generate
 * @property language - Target language code ('ja' for Japanese, 'en' for English)
 * 
 * @example
 * ```typescript
 * const request: SlideGenerationRequest = {
 *   projectId: 'PROJECT_123',
 *   themes: ['project_overview', 'project_progress'],
 *   language: 'ja'
 * }
 * ```
 */
export interface SlideGenerationRequest {
  projectId: string
  themes: SlideTheme[]
  language: string
}

/**
 * Response from slide generation initiation.
 * 
 * @interface SlideGenerationResponse
 * @property slideId - Unique identifier for this generation session
 * @property status - Current generation status ('generating', 'completed', 'error')
 * @property websocketUrl - WebSocket endpoint for real-time generation updates
 */
export interface SlideGenerationResponse {
  slideId: string
  status: string
  websocketUrl: string
}

/**
 * Complete slide content with both source and rendered data.
 * 
 * @interface SlideContent
 * @property index - Slide position in presentation (1-based indexing)
 * @property theme - Theme that generated this slide
 * @property title - Human-readable slide title
 * @property markdown - Source markdown content from AI generation
 * @property html - Rendered HTML content (added by LLM-based compiler)
 * @property generatedAt - ISO timestamp when slide was created
 * 
 * @example
 * ```typescript
 * const slide: SlideContent = {
 *   index: 1,
 *   theme: 'project_overview',
 *   title: 'Project Overview',
 *   markdown: '# Project Overview\n\n...',
 *   html: '<div><h1>Project Overview</h1>...</div>',
 *   generatedAt: '2024-08-17T10:30:00Z'
 * }
 * ```
 */
export interface SlideContent {
  index: number
  theme: SlideTheme
  title: string
  markdown: string
  html?: string           // Optional HTML content from LLM compilation
  generatedAt: string
}

export interface SlideNarration {
  slideIndex: number
  text: string
  language: string
}

export interface SlideAudio {
  slideIndex: number
  audioUrl: string
  duration: number
}

export interface PresentationComplete {
  totalSlides: number
  duration: string
}

export interface WebSocketMessage {
  type: string
  data: any
}

export interface ErrorMessage {
  message: string
  code: string
}

export interface SlideThemeOption {
  value: SlideTheme
  label: string
  description: string
}