/**
 * @fileoverview Authentication type definitions for the Intelligent Presenter application.
 * 
 * This module contains TypeScript interfaces for authentication-related data structures
 * used throughout the OAuth 2.0 authentication flow with Backlog. These types ensure
 * type safety when handling user information, authentication tokens, and OAuth responses.
 * 
 * @author Nulab Technical Challenge
 * @version 1.0.0
 */

/**
 * Represents detailed user information retrieved from the Backlog API.
 * This interface contains all user properties provided by Backlog's user endpoints
 * and is used throughout the application for user identification and display.
 * 
 * @interface UserInfo
 * @property {number} id - Numeric user identifier from Backlog
 * @property {string} userId - String-based user ID (typically username)
 * @property {string} name - Display name of the user
 * @property {number} roleType - User role type identifier (1=Admin, 2=User, etc.)
 * @property {string} lang - User's preferred language code (e.g., 'ja', 'en')
 * @property {string} mailAddress - User's email address
 * @property {Object} nulabAccount - Nulab account information
 * @property {string} nulabAccount.nulabId - Unique Nulab account identifier
 * @property {string} nulabAccount.name - Nulab account display name
 * @property {string} nulabAccount.uniqueId - Unique Nulab account ID
 * 
 * @example
 * ```typescript
 * const user: UserInfo = {
 *   id: 12345,
 *   userId: "developer",
 *   name: "John Developer",
 *   roleType: 2,
 *   lang: "ja",
 *   mailAddress: "developer@example.com",
 *   nulabAccount: {
 *     nulabId: "nulab123",
 *     name: "John Developer",
 *     uniqueId: "unique456"
 *   }
 * }
 * ```
 */
export interface UserInfo {
  id: number
  userId: string
  name: string
  roleType: number
  lang: string
  mailAddress: string
  nulabAccount: {
    nulabId: string
    name: string
    uniqueId: string
  }
}

/**
 * Represents the response from successful authentication operations.
 * This interface is returned from both OAuth callback handling and token refresh operations.
 * 
 * @interface AuthResponse
 * @property {string} token - JWT access token for API authentication
 * @property {UserInfo} userInfo - Complete user information from Backlog
 * @property {string} expiresAt - ISO string timestamp when the token expires
 * 
 * @example
 * ```typescript
 * const authResponse: AuthResponse = {
 *   token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
 *   userInfo: {
 *     id: 12345,
 *     userId: "developer",
 *     name: "John Developer",
 *     // ... other user properties
 *   },
 *   expiresAt: "2024-08-18T10:30:00Z"
 * }
 * ```
 */
export interface AuthResponse {
  token: string
  userInfo: UserInfo
  expiresAt: string
}

/**
 * Represents the response from OAuth 2.0 initialization.
 * This interface contains the necessary information to redirect users
 * to Backlog's OAuth authorization endpoint.
 * 
 * @interface OAuthInitResponse
 * @property {string} authUrl - Complete authorization URL for user redirection
 * @property {string} state - CSRF protection state parameter for OAuth flow
 * 
 * @example
 * ```typescript
 * const oauthInit: OAuthInitResponse = {
 *   authUrl: "https://demo.backlog.com/OAuth2AccessRequest.action?client_id=...",
 *   state: "random-csrf-token-123456"
 * }
 * 
 * // Redirect user to authUrl for authentication
 * window.location.href = oauthInit.authUrl
 * ```
 */
export interface OAuthInitResponse {
  authUrl: string
  state: string
}