/**
 * Test suite for authentication store
 * Tests Pinia store functionality, OAuth flow, and token management
 */

import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useAuthStore } from '@/stores/auth'

// Mock the API service
vi.mock('@/services/api', () => ({
  authApi: {
    initiateOAuth: vi.fn(),
    handleCallback: vi.fn(),
    getUserInfo: vi.fn(),
    logout: vi.fn(),
    refreshToken: vi.fn(),
  },
}))

describe('Auth Store', () => {
  beforeEach(() => {
    // Clear localStorage before each test
    localStorage.clear()
    // Create fresh Pinia instance
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('should initialize with default state when no token in localStorage', () => {
    // Ensure localStorage is empty
    expect(localStorage.getItem('auth_token')).toBeNull()
    
    const authStore = useAuthStore()
    
    expect(authStore.isAuthenticated).toBe(false)
    expect(authStore.userInfo).toBeNull()
    expect(authStore.token).toBeNull()
    expect(authStore.isLoading).toBe(false)
  })

  it('should initialize with token from localStorage', () => {
    // Set token in localStorage BEFORE creating store
    localStorage.setItem('auth_token', 'stored-token')
    
    // Create store after setting localStorage
    const authStore = useAuthStore()
    
    expect(authStore.token).toBe('stored-token')
    expect(authStore.isAuthenticated).toBe(true)
  })

  it('should handle successful login', async () => {
    const authStore = useAuthStore()
    const mockAuthUrl = 'https://example.backlog.jp/OAuth2AccessRequest.action?...'
    
    // Mock the login API call
    const { authApi } = await import('@/services/api')
    vi.mocked(authApi.initiateOAuth).mockResolvedValue({ authUrl: mockAuthUrl })
    
    const result = await authStore.login()
    
    expect(authApi.initiateOAuth).toHaveBeenCalled()
    expect(result).toBe(mockAuthUrl)
    expect(authStore.isLoading).toBe(false)
  })

  it('should handle OAuth callback successfully', async () => {
    const authStore = useAuthStore()
    const mockResponse = {
      token: 'mock-jwt-token',
      userInfo: {
        id: 123,
        userId: 'testuser',
        name: 'Test User',
        roleType: 2,
        lang: 'ja',
        mailAddress: 'test@example.com',
        nulabAccount: {
          nulabId: 'nulab123',
          name: 'Test User',
          uniqueId: 'unique456'
        }
      },
    }
    
    const { authApi } = await import('@/services/api')
    vi.mocked(authApi.handleCallback).mockResolvedValue(mockResponse)
    
    await authStore.handleCallback('auth-code', 'state-token')
    
    expect(authApi.handleCallback).toHaveBeenCalledWith('auth-code', 'state-token')
    expect(authStore.isAuthenticated).toBe(true)
    expect(authStore.token).toBe(mockResponse.token)
    expect(authStore.userInfo).toEqual(mockResponse.userInfo)
  })

  it('should handle callback errors gracefully', async () => {
    const authStore = useAuthStore()
    
    const { authApi } = await import('@/services/api')
    vi.mocked(authApi.handleCallback).mockRejectedValue(new Error('OAuth error'))
    
    await expect(authStore.handleCallback('invalid-code', 'invalid-state')).rejects.toThrow('OAuth error')
    
    expect(authStore.isAuthenticated).toBe(false)
    expect(authStore.token).toBeNull()
  })

  it('should handle logout correctly', async () => {
    const authStore = useAuthStore()
    
    // Set up authenticated state using setToken method
    authStore.setToken('test-token')
    expect(authStore.isAuthenticated).toBe(true)
    
    const { authApi } = await import('@/services/api')
    vi.mocked(authApi.logout).mockResolvedValue(undefined)
    
    await authStore.logout()
    
    expect(authApi.logout).toHaveBeenCalled()
    expect(authStore.isAuthenticated).toBe(false)
    expect(authStore.token).toBeNull()
  })

  it('should restore authentication from localStorage', async () => {
    const mockToken = 'stored-token'
    const mockUserInfo = {
      id: 123,
      userId: 'storeduser',
      name: 'Stored User',
      roleType: 2,
      lang: 'ja',
      mailAddress: 'stored@example.com',
      nulabAccount: {
        nulabId: 'nulab456',
        name: 'Stored User',
        uniqueId: 'unique789'
      }
    }
    
    // Set token in localStorage BEFORE creating store
    localStorage.setItem('auth_token', mockToken)
    
    const { authApi } = await import('@/services/api')
    vi.mocked(authApi.getUserInfo).mockResolvedValue(mockUserInfo)
    
    // Create store after setting localStorage
    const authStore = useAuthStore()
    
    // Verify initial state from localStorage
    expect(authStore.token).toBe(mockToken)
    expect(authStore.isAuthenticated).toBe(true)
    
    // Call initializeAuth to fetch user info
    await authStore.initializeAuth()
    
    expect(authStore.userInfo).toEqual(mockUserInfo)
  })

  it('should handle loading states correctly', async () => {
    const authStore = useAuthStore()
    
    const { authApi } = await import('@/services/api')
    let resolveLogin: (value: any) => void
    const loginPromise = new Promise(resolve => {
      resolveLogin = resolve
    })
    vi.mocked(authApi.initiateOAuth).mockReturnValue(loginPromise)
    
    const loginCall = authStore.login()
    
    expect(authStore.isLoading).toBe(true)
    
    resolveLogin({ authUrl: 'test-url' })
    await loginCall
    
    expect(authStore.isLoading).toBe(false)
  })

  it('should clear invalid tokens during initialization', async () => {
    const invalidToken = 'invalid-token'
    
    // Set invalid token in localStorage BEFORE creating store
    localStorage.setItem('auth_token', invalidToken)
    
    const { authApi } = await import('@/services/api')
    vi.mocked(authApi.getUserInfo).mockRejectedValue(new Error('Invalid token'))
    
    // Create store after setting localStorage
    const authStore = useAuthStore()
    
    // Initially should have the token from localStorage
    expect(authStore.token).toBe(invalidToken)
    expect(authStore.isAuthenticated).toBe(true)
    
    // After initializeAuth, invalid token should be cleared
    await authStore.initializeAuth()
    
    expect(authStore.isAuthenticated).toBe(false)
    expect(authStore.token).toBeNull()
    expect(authStore.userInfo).toBeNull()
  })

  it('should handle setToken method', () => {
    const authStore = useAuthStore()
    const testToken = 'test-token-123'
    
    authStore.setToken(testToken)
    
    expect(authStore.token).toBe(testToken)
    expect(authStore.isAuthenticated).toBe(true)
    expect(localStorage.getItem('auth_token')).toBe(testToken)
  })

  it('should handle token refresh', async () => {
    const authStore = useAuthStore()
    const newAuthResponse = {
      token: 'new-refreshed-token',
      userInfo: {
        id: 123,
        userId: 'user',
        name: 'User',
        roleType: 2,
        lang: 'ja',
        mailAddress: 'user@example.com',
        nulabAccount: {
          nulabId: 'nulab123',
          name: 'User',
          uniqueId: 'unique123'
        }
      },
    }
    
    const { authApi } = await import('@/services/api')
    vi.mocked(authApi.refreshToken).mockResolvedValue(newAuthResponse)
    
    await authStore.refreshToken()
    
    expect(authApi.refreshToken).toHaveBeenCalled()
    expect(authStore.token).toBe(newAuthResponse.token)
    expect(authStore.userInfo).toEqual(newAuthResponse.userInfo)
  })

  it('should clear auth on refresh token failure', async () => {
    const authStore = useAuthStore()
    authStore.setToken('old-token')
    
    const { authApi } = await import('@/services/api')
    vi.mocked(authApi.refreshToken).mockRejectedValue(new Error('Token expired'))
    
    await expect(authStore.refreshToken()).rejects.toThrow('Token expired')
    
    expect(authStore.token).toBeNull()
    expect(authStore.isAuthenticated).toBe(false)
  })

  it('should handle computed properties correctly', () => {
    const authStore = useAuthStore()
    
    // Initially not authenticated
    expect(authStore.isAuthenticated).toBe(false)
    expect(authStore.token).toBeNull()
    expect(authStore.userInfo).toBeNull()
    expect(authStore.isLoading).toBe(false)
    
    // Set token
    authStore.setToken('test-token')
    expect(authStore.isAuthenticated).toBe(true)
    expect(authStore.token).toBe('test-token')
  })

  it('should handle logout without token', async () => {
    const authStore = useAuthStore()
    
    const { authApi } = await import('@/services/api')
    vi.mocked(authApi.logout).mockResolvedValue(undefined)
    
    // Logout without being authenticated
    await authStore.logout()
    
    // Should not call API if no token
    expect(authApi.logout).not.toHaveBeenCalled()
    expect(authStore.isAuthenticated).toBe(false)
  })

  it('should handle localStorage persistence correctly', async () => {
    const authStore = useAuthStore()
    const testToken = 'persistence-test-token'
    
    // Initially no token
    expect(localStorage.getItem('auth_token')).toBeNull()
    
    // Set token
    authStore.setToken(testToken)
    
    // Verify localStorage is updated
    expect(localStorage.getItem('auth_token')).toBe(testToken)
    expect(authStore.token).toBe(testToken)
    
    // Clear auth (logout is async)
    await authStore.logout()
    
    // Verify localStorage is cleared
    expect(localStorage.getItem('auth_token')).toBeNull()
    expect(authStore.token).toBeNull()
  })
})
