/**
 * @fileoverview Pinia store for authentication state management.
 * 
 * This module provides:
 * - OAuth 2.0 authentication flow management
 * - Secure token storage and retrieval
 * - User session lifecycle management
 * - Automatic token refresh and validation
 * - Authentication state reactivity
 * 
 * The store integrates with Backlog's OAuth API to provide seamless
 * authentication for the Intelligent Presenter application.
 * 
 * @author Nulab Technical Challenge
 * @version 1.0.0
 */

import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi } from '@/services/api'
import type { UserInfo, AuthResponse } from '@/types/auth'

/**
 * Pinia store for authentication state management and operations.
 * 
 * This store provides:
 * - Reactive authentication state
 * - OAuth 2.0 flow handling
 * - Token management and refresh
 * - User information caching
 * - Session persistence across browser refreshes
 * 
 * @returns Authentication store with reactive state and methods
 * 
 * @example
 * ```typescript
 * const authStore = useAuthStore()
 * 
 * // Check authentication status
 * if (authStore.isAuthenticated) {
 *   console.log(`Welcome, ${authStore.userInfo?.name}!`)
 * }
 * 
 * // Initiate login
 * const authUrl = await authStore.login()
 * window.location.href = authUrl
 * 
 * // Handle OAuth callback
 * await authStore.handleCallback(code, state)
 * ```
 */
export const useAuthStore = defineStore('auth', () => {
  /** JWT authentication token stored in localStorage */
  const token = ref<string | null>(localStorage.getItem('auth_token'))
  
  /** Current user information from Backlog API */
  const userInfo = ref<UserInfo | null>(null)
  
  /** Loading state for authentication operations */
  const isLoading = ref(false)

  /** Computed property indicating whether user is authenticated */
  const isAuthenticated = computed(() => !!token.value)

  /**
   * Initializes authentication state on application startup.
   * 
   * This method:
   * 1. Checks for existing tokens in localStorage
   * 2. Validates tokens by fetching user information
   * 3. Clears invalid tokens and resets state
   * 4. Populates user information for valid sessions
   * 
   * Called automatically when the application starts to restore
   * authentication state from previous sessions.
   * 
   * @async
   * @function initializeAuth
   * @returns {Promise<void>} Promise that resolves when initialization is complete
   * 
   * @example
   * ```typescript
   * // In main application startup
   * onMounted(() => {
   *   authStore.initializeAuth()
   * })
   * ```
   */
  const initializeAuth = async () => {
    if (token.value) {
      try {
        const user = await authApi.getUserInfo()
        userInfo.value = user
      } catch (error) {
        // Token is invalid, clear it
        clearAuth()
      }
    }
  }

  /**
   * Sets authentication state from a complete auth response.
   * 
   * This method:
   * 1. Stores the JWT token in reactive state and localStorage
   * 2. Caches user information in the store
   * 3. Ensures persistence across browser sessions
   * 
   * @param {AuthResponse} authResponse - Complete authentication response
   * @param {string} authResponse.token - JWT authentication token
   * @param {UserInfo} authResponse.userInfo - User profile information
   * @param {string} authResponse.expiresAt - Token expiration timestamp
   * 
   * @example
   * ```typescript
   * const authResponse = await authApi.handleCallback(code, state)
   * authStore.setAuth(authResponse)
   * ```
   */
  const setAuth = (authResponse: AuthResponse) => {
    token.value = authResponse.token
    userInfo.value = authResponse.userInfo
    localStorage.setItem('auth_token', authResponse.token)
  }

  /**
   * Sets authentication token directly (used for URL-based token flow).
   * 
   * This method sets only the token without user information,
   * typically used when receiving tokens via URL parameters.
   * User information should be fetched separately.
   * 
   * @param {string} tokenValue - JWT authentication token
   * 
   * @example
   * ```typescript
   * // From URL parameters: ?token=eyJhbGciOiJIUzI1NiIs...
   * const token = route.query.token as string
   * authStore.setToken(token)
   * await authStore.initializeAuth() // Fetch user info
   * ```
   */
  const setToken = (tokenValue: string) => {
    token.value = tokenValue
    localStorage.setItem('auth_token', tokenValue)
  }

  /**
   * Clears all authentication state and storage.
   * 
   * This method:
   * 1. Removes tokens from reactive state
   * 2. Clears user information cache
   * 3. Removes persistent storage
   * 4. Resets the store to unauthenticated state
   * 
   * Called when tokens expire, logout occurs, or authentication fails.
   * 
   * @example
   * ```typescript
   * // On 401 error or explicit logout
   * authStore.clearAuth()
   * router.push('/login')
   * ```
   */
  const clearAuth = () => {
    token.value = null
    userInfo.value = null
    localStorage.removeItem('auth_token')
  }

  /**
   * Initiates the OAuth 2.0 login flow with Backlog.
   * 
   * This method:
   * 1. Sets loading state for UI feedback
   * 2. Calls the authentication API to get OAuth URL
   * 3. Returns the authorization URL for redirection
   * 4. Resets loading state when complete
   * 
   * @async
   * @function login
   * @returns {Promise<string>} Authorization URL for user redirection
   * @throws {Error} If OAuth initiation fails
   * 
   * @example
   * ```typescript
   * const handleLogin = async () => {
   *   try {
   *     const authUrl = await authStore.login()
   *     window.location.href = authUrl
   *   } catch (error) {
   *     console.error('Login failed:', error)
   *   }
   * }
   * ```
   */
  const login = async (): Promise<string> => {
    isLoading.value = true
    try {
      const response = await authApi.initiateOAuth()
      return response.authUrl
    } finally {
      isLoading.value = false
    }
  }

  /**
   * Handles OAuth callback from Backlog and completes authentication.
   * 
   * This method:
   * 1. Sets loading state for UI feedback
   * 2. Exchanges authorization code for access token
   * 3. Stores authentication data in the store
   * 4. Completes the authentication flow
   * 
   * @async
   * @function handleCallback
   * @param {string} code - Authorization code from Backlog OAuth
   * @param {string} state - CSRF protection state parameter
   * @returns {Promise<void>} Promise that resolves when callback is processed
   * @throws {Error} If callback processing fails
   * 
   * @example
   * ```typescript
   * // In callback route component
   * const code = route.query.code as string
   * const state = route.query.state as string
   * 
   * await authStore.handleCallback(code, state)
   * router.push('/') // Redirect to home page
   * ```
   */
  const handleCallback = async (code: string, state: string): Promise<void> => {
    isLoading.value = true
    try {
      const authResponse = await authApi.handleCallback(code, state)
      setAuth(authResponse)
    } finally {
      isLoading.value = false
    }
  }

  /**
   * Logs out the current user and cleans up authentication state.
   * 
   * This method:
   * 1. Sets loading state for UI feedback
   * 2. Calls logout API to invalidate server session
   * 3. Clears all local authentication state
   * 4. Handles logout errors gracefully
   * 
   * @async
   * @function logout
   * @returns {Promise<void>} Promise that resolves when logout is complete
   * 
   * @example
   * ```typescript
   * const handleLogout = async () => {
   *   await authStore.logout()
   *   router.push('/login')
   * }
   * ```
   */
  const logout = async () => {
    isLoading.value = true
    try {
      if (token.value) {
        await authApi.logout()
      }
    } catch (error) {
      console.error('Logout error:', error)
    } finally {
      clearAuth()
      isLoading.value = false
    }
  }

  /**
   * Refreshes the current authentication token.
   * 
   * This method:
   * 1. Calls the refresh token API endpoint
   * 2. Updates stored authentication data with new token
   * 3. Clears auth state if refresh fails (token expired)
   * 4. Throws error for calling code to handle
   * 
   * @async
   * @function refreshToken
   * @returns {Promise<void>} Promise that resolves when refresh is complete
   * @throws {Error} If token refresh fails (user needs to re-authenticate)
   * 
   * @example
   * ```typescript
   * try {
   *   await authStore.refreshToken()
   *   // Continue with authenticated operation
   * } catch (error) {
   *   // Token expired, redirect to login
   *   router.push('/login')
   * }
   * ```
   */
  const refreshToken = async () => {
    try {
      const authResponse = await authApi.refreshToken()
      setAuth(authResponse)
    } catch (error) {
      clearAuth()
      throw error
    }
  }

  /**
   * Return the store's public API with reactive state and methods.
   *
   * The returned object provides:
   * - Computed properties for reactive state access
   * - Authentication methods for login/logout operations
   * - Token management utilities
   * - Session lifecycle management
   */
  return {
    /** Computed property for current authentication token */
    token: computed(() => token.value),

    /** Computed property for current user information */
    userInfo: computed(() => userInfo.value),

    /** Computed property for authentication status */
    isAuthenticated,

    /** Computed property for loading state */
    isLoading: computed(() => isLoading.value),

    /** Initialize authentication state on app startup */
    initializeAuth,

    /** Set authentication token directly */
    setToken,

    /** Initiate OAuth login flow */
    login,

    /** Handle OAuth callback */
    handleCallback,

    /** Logout and clear authentication */
    logout,

    /** Refresh authentication token */
    refreshToken
  }
})