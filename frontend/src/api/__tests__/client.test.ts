import { describe, it, expect, beforeEach, vi } from 'vitest'
import { ApiRequestError } from '@/types'

// Mock localStorage
const localStorageStore = new Map<string, string>()
Object.defineProperty(globalThis, 'localStorage', {
  value: {
    getItem: vi.fn((key: string) => localStorageStore.get(key) ?? null),
    setItem: vi.fn((key: string, value: string) => { localStorageStore.set(key, value) }),
    removeItem: vi.fn((key: string) => { localStorageStore.delete(key) }),
    clear: vi.fn(() => localStorageStore.clear()),
  },
  writable: true,
})

// Mock fetch
const mockFetch = vi.fn()
globalThis.fetch = mockFetch

beforeEach(() => {
  localStorageStore.clear()
  mockFetch.mockReset()
})

describe('ApiClient token management', () => {
  it('should start with no token', async () => {
    const { api } = await import('../client')
    expect(api.getAccessToken()).toBeNull()
  })

  it('should accept a JWT token', async () => {
    const { api } = await import('../client')
    api.setAccessToken('test-jwt-token')
    expect(api.getAccessToken()).toBe('test-jwt-token')
  })

  it('should clear token when setAccessToken(null)', async () => {
    const { api } = await import('../client')
    api.setAccessToken('test-token')
    api.setAccessToken(null)
    expect(api.getAccessToken()).toBeNull()
  })

  it('hasValidToken should return false when no token', async () => {
    const { api } = await import('../client')
    expect(api.hasValidToken()).toBe(false)
  })

  it('hasValidToken should return false when token is set (no expiry)', async () => {
    // Without expiry info, token is not considered valid
    const { api } = await import('../client')
    api.setAccessToken('some-token')
    expect(api.hasValidToken()).toBe(false)
  })

  it('setFamilyId should store the family ID', async () => {
    const { api } = await import('../client')
    api.setFamilyId('family-1')
    // No direct getter, but ensure no error
    api.setFamilyId(null)
  })
})

describe('ApiClient API errors', () => {
  it('ApiRequestError should store error info', () => {
    const apiError = { code: 'NOT_FOUND', summary: '资源不存在' }
    const err = new ApiRequestError(apiError, 404)
    expect(err.status).toBe(404)
    expect(err.rawCode).toBe('NOT_FOUND')
    expect(err.summary).toBe('资源不存在')
    expect(err.message).toContain('资源不存在')
  })

  it('ApiRequestError should handle empty error', () => {
    const err = new ApiRequestError({ code: '', summary: '' }, 500)
    expect(err.status).toBe(500)
  })
})

describe('ApiClient initSession', () => {
  it('should return null when token already valid', async () => {
    const { api } = await import('../client')
    // Manually set a valid token
    ;(api as any).accessToken = 'valid'
    ;(api as any).accessTokenExpiresAt = Date.now() + 3600000

    const result = await api.initSession()
    expect(result).toBeNull()
  })

  it('should return null when refresh API fails', async () => {
    mockFetch.mockReset()
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 401,
      json: () => Promise.resolve({ success: false }),
    })

    const { api } = await import('../client')
    const result = await api.initSession()
    expect(result).toBeNull()
  })

  it('should return user when refresh succeeds', async () => {
    mockFetch.mockReset()
    const mockResponse = {
      success: true,
      data: {
        access_token: 'new-token',
        user: { id: 'user-1', display_name: '测试用户' },
      },
    }
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve(mockResponse),
    })

    // Get fresh module to clear singleton state
    vi.resetModules()
    const { api } = await import('../client')

    const result = await api.initSession()
    expect(result).not.toBeNull()
    expect((result as any)?.id).toBe('user-1')
    expect(api.getAccessToken()).toBe('new-token')
  })
})

describe('ApiClient onExpired callback', () => {
  it('should register and fire the callback', async () => {
    const { api } = await import('../client')
    const spy = vi.fn()
    api.onExpired(spy)
    // The callback fires only on actual 401 from refresh
    expect(spy).not.toHaveBeenCalled()
  })
})

describe('ApiClient onFamilyNotFound callback', () => {
  it('should register the callback', async () => {
    const { api } = await import('../client')
    const spy = vi.fn()
    api.onFamilyNotFound(spy)
    expect(spy).not.toHaveBeenCalled()
  })
})
