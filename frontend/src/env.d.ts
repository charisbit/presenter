/**
 * @fileoverview Environment type declarations for the Intelligent Presenter frontend.
 * 
 * This module provides TypeScript type definitions for:
 * - Vite client-side environment configuration
 * - Vue Single File Component (SFC) module declarations
 * - Environment variables used throughout the application
 * - Import meta interface extensions for type safety
 * 
 * These declarations ensure type safety when working with environment variables
 * and importing Vue components in the TypeScript codebase.
 * 
 * @author Nulab Technical Challenge
 * @version 1.0.0
 */

/// <reference types="vite/client" />

/**
 * Module declaration for Vue Single File Components.
 * This allows TypeScript to properly recognize .vue file imports
 * and provides type information for Vue components.
 * 
 * @example
 * ```typescript
 * import MyComponent from './MyComponent.vue'
 * // TypeScript now knows MyComponent is a Vue component
 * ```
 */
declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  
  /**
   * Generic Vue component type definition.
   * - Props: {} (empty object for simplicity, can be extended)
   * - Returns: {} (empty object for return value)
   * - Any: any additional component properties
   */
  const component: DefineComponent<{}, {}, any>
  export default component
}

/**
 * Interface defining the structure of environment variables available at import.meta.env.
 * These variables are set during the build process and provide configuration
 * for different deployment environments (development, staging, production).
 * 
 * @interface ImportMetaEnv
 * @extends {ViteClientEnv} Inherits from Vite's base environment interface
 */
interface ImportMetaEnv {
  /**
   * Base URL for the backend API server.
   * Used for all HTTP requests to the Intelligent Presenter backend.
   * 
   * @example
   * - Development: 'http://localhost:8080'
   * - Production: 'https://api.presenter.example.com'
   */
  readonly VITE_API_BASE_URL: string
  
  /**
   * WebSocket URL for real-time communication.
   * Used for receiving slide generation updates and progress notifications.
   * 
   * @example
   * - Development: 'ws://localhost:8080'
   * - Production: 'wss://api.presenter.example.com'
   */
  readonly VITE_WS_URL: string
}

/**
 * Extension of the ImportMeta interface to include our custom environment variables.
 * This provides type safety when accessing environment variables via import.meta.env.
 * 
 * @interface ImportMeta
 * @property {ImportMetaEnv} env - Typed environment variables object
 * 
 * @example
 * ```typescript
 * const apiUrl = import.meta.env.VITE_API_BASE_URL // TypeScript knows this is a string
 * const wsUrl = import.meta.env.VITE_WS_URL        // TypeScript knows this is a string
 * ```
 */
interface ImportMeta {
  readonly env: ImportMetaEnv
}