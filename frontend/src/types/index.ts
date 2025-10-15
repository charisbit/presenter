/**
 * @fileoverview Base type definitions for the Intelligent Presenter application.
 * 
 * This module contains core TypeScript interfaces and types that are used
 * throughout the application. It serves as a central hub for common data
 * structures and re-exports specialized types from other modules.
 * 
 * @author Technical Challenge
 * @version 1.0.0
 */

/**
 * Represents a Backlog project within the Intelligent Presenter system.
 * This interface defines the essential properties of a project that are
 * used for slide generation and presentation creation.
 * 
 * @interface Project
 * @property {string} id - Unique identifier for the project (from Backlog API)
 * @property {string} name - Human-readable project name
 * @property {string} [key] - Optional project key/code (e.g., "PROJ", "DEV")
 * @property {string} [description] - Optional project description
 * 
 * @example
 * ```typescript
 * const project: Project = {
 *   id: "216125",
 *   name: "Intelligent Presenter",
 *   key: "IP",
 *   description: "AI-powered presentation generation from Backlog data"
 * }
 * ```
 */
export interface Project {
  id: string
  name: string
  key?: string
  description?: string
}

/**
 * Union type defining the available slide themes for presentation generation.
 * Each theme focuses on a specific aspect of project management and analysis,
 * generating tailored content based on Backlog project data.
 * 
 * @typedef {string} SlideTheme
 * 
 * Available themes:
 * - `project_overview`: Basic project information, goals, and team structure
 * - `project_progress`: Completion rates, milestone tracking, and timeline analysis
 * - `issue_management`: Issue statistics, priority distribution, and resolution metrics
 * - `risk_analysis`: Risk identification, mitigation strategies, and impact assessment
 * - `team_collaboration`: Team activities, communication patterns, and productivity metrics
 * - `document_management`: Wiki usage, documentation coverage, and knowledge sharing
 * - `codebase_activity`: Development metrics, code quality, and repository statistics
 * - `notifications`: Communication efficiency, alert patterns, and engagement metrics
 * - `predictive_analysis`: Forecasts, trend analysis, and future planning insights
 * - `summary_plan`: Project summary, lessons learned, and next steps
 * 
 * @example
 * ```typescript
 * const selectedThemes: SlideTheme[] = [
 *   'project_overview',
 *   'issue_management',
 *   'team_collaboration'
 * ]
 * ```
 */
export type SlideTheme = 
  | 'project_overview'
  | 'project_progress' 
  | 'issue_management'
  | 'risk_analysis'
  | 'team_collaboration'
  | 'document_management'
  | 'codebase_activity'
  | 'notifications'
  | 'predictive_analysis'
  | 'summary_plan'

/**
 * Re-export all authentication-related types from the auth module.
 * This includes user information, authentication responses, and OAuth interfaces.
 * 
 * @see {@link ./auth} for detailed authentication type definitions
 */
export * from './auth'

/**
 * Re-export all slide-related types from the slides module.
 * This includes slide content, generation requests, WebSocket messages, and more.
 * 
 * @see {@link ./slides} for detailed slide type definitions
 */
export * from './slides'