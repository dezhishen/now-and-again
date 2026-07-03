import type { APIResponse, User } from '@/types'
import { ApiRequestError } from '@/types'
import { requestToUTC, responseToLocal } from '@/composables/timezone'

const BASE_URL = '/api'
const TOKEN_KEY = 'na_access_token'
const SESSION_EXPIRED_CODE = 401

// ─── Token persistence ────────────────────────────────────────────
//
// Stored in localStorage so the access token survives tab close and
// browser restart. The refresh token is an httpOnly cookie (secure),
// and the access token is short-lived (15 min), so the XSS exposure
// window is limited.
//
// Format: JSON { token: string, expiresAt: number (epoch ms) }

interface StoredToken {
  token: string
  expiresAt: number // epoch milliseconds
}

function loadStoredToken(): StoredToken | null {
  try {
    const raw = localStorage.getItem(TOKEN_KEY)
    if (!raw) return null
    const parsed: StoredToken = JSON.parse(raw)
    if (!parsed.token || typeof parsed.expiresAt !== 'number') return null
    return parsed
  } catch { return null }
}

function saveStoredToken(t: StoredToken | null) {
  try {
    if (t) localStorage.setItem(TOKEN_KEY, JSON.stringify(t))
    else localStorage.removeItem(TOKEN_KEY)
  } catch { /* quota exceeded or private browsing */ }
}

// ─── JWT helpers (only used when receiving a NEW token) ──────────

/** Extract exp claim from a JWT. Returns epoch seconds, or 0 if unreadable. */
function getJWTExpiry(token: string): number {
  try {
    const payload = token.split('.')[1]
    // Convert base64url → base64 (Go's jwt lib uses RawURLEncoding)
    const base64 = payload.replace(/-/g, '+').replace(/_/g, '/')
    const claims = JSON.parse(atob(base64))
    return claims?.exp || 0
  } catch { return 0 }
}

// ─── API Client ──────────────────────────────────────────────────

class ApiClient {
  private accessToken: string | null = null
  private accessTokenExpiresAt: number = 0
  private refreshPromise: Promise<boolean> | null = null
  private onSessionExpired: (() => void) | null = null
  private sessionExpiredFired = false
  private familyId: string | null = null

  constructor() {
    const stored = loadStoredToken()
    if (stored && stored.expiresAt > Date.now()) {
      this.accessToken = stored.token
      this.accessTokenExpiresAt = stored.expiresAt
    }
  }

  private setToken(token: string | null, expiresAt: number = 0) {
    this.accessToken = token
    this.accessTokenExpiresAt = expiresAt
    if (token && expiresAt > 0) {
      saveStoredToken({ token, expiresAt })
    } else {
      saveStoredToken(null)
    }
    if (token) this.sessionExpiredFired = false
  }

  setAccessToken(token: string | null) {
    if (token) {
      const exp = getJWTExpiry(token)
      this.setToken(token, exp > 0 ? exp * 1000 : 0)
    } else {
      this.setToken(null)
    }
  }
  getAccessToken() { return this.accessToken }

  /** Set the active family ID, sent via X-Family-Id header on every request. */
  setFamilyId(id: string | null) { this.familyId = id }

  /** True if we hold a non-expired access token (no JWT decode needed). */
  hasValidToken(): boolean {
    return !!this.accessToken && this.accessTokenExpiresAt > Date.now()
  }

  /** True if the token will expire within `seconds`. */
  private isTokenExpiringSoon(seconds = 60): boolean {
    if (!this.accessToken || this.accessTokenExpiresAt <= 0) return false
    return Date.now() >= this.accessTokenExpiresAt - seconds * 1000
  }

  /** Register callback: fired exactly once when session is confirmed expired (refresh returned 401). */
  onExpired(fn: () => void) { this.onSessionExpired = fn }

  /**
   * Try to restore session from refresh token cookie.
   * Only call this when we don't have a valid token (not on every page load).
   * Returns user if successful, null if refresh also failed.
   */
  async initSession(): Promise<User | null> {
    // Already have a valid token — nothing to do
    if (this.hasValidToken()) return null

    try {
      const res = await fetch(`${BASE_URL}/auth/refresh`, {
        method: 'POST', credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
      })
      if (!res.ok) return null
      const json: APIResponse<{ access_token: string; user: User }> = await res.json()
      if (json.success && json.data?.access_token) {
        this.setAccessToken(json.data.access_token)
        return json.data.user
      }
      return null
    } catch { return null }
  }

  /**
   * Refresh the access token using the httpOnly refresh-token cookie.
   *
   * Returns:
   *   true  — new token obtained, stored in memory + sessionStorage
   *   false — refresh failed; if HTTP 401 the old token was already cleared
   *
   * Concurrent calls are deduplicated: all wait for the single in-flight refresh.
   */
  private async refreshAccessToken(): Promise<boolean> {
    // Dedupe: if a refresh is already in flight, wait for its result
    if (this.refreshPromise) return this.refreshPromise

    this.refreshPromise = (async () => {
      try {
        const res = await fetch(`${BASE_URL}/auth/refresh`, {
          method: 'POST', credentials: 'include',
          headers: { 'Content-Type': 'application/json' },
        })

        if (res.status === 401) {
          // Refresh token is also expired — session is truly over
          this.setToken(null)
          return false
        }

        if (!res.ok) {
          // Server error (5xx) or other — keep existing token, caller can retry
          return false
        }

        const json: APIResponse<{ access_token: string }> = await res.json()
        if (json.success && json.data?.access_token) {
          this.setAccessToken(json.data.access_token)
          return true
        }
        return false
      } catch {
        // Network error — keep existing token, caller can retry
        return false
      } finally {
        this.refreshPromise = null
      }
    })()

    return this.refreshPromise
  }

  private async request<T>(method: string, path: string, body?: unknown): Promise<T> {
    // ── Proactive refresh ────────────────────────────────────
    // Save token before refresh so we can detect if it was cleared.
    const hadToken = !!this.accessToken
    if (this.accessToken && this.isTokenExpiringSoon(120)) {
      await this.refreshAccessToken()
    }
    // If proactive refresh cleared our token (refresh token expired),
    // don't even attempt the API call — session is over.
    if (hadToken && !this.accessToken) {
      if (!this.sessionExpiredFired) {
        this.sessionExpiredFired = true
        this.onSessionExpired?.()
      }
      throw new Error('Session expired')
    }

    const doFetch = async () => {
      const headers: Record<string, string> = {
        'Content-Type': 'application/json', Accept: 'application/json',
      }
      if (this.accessToken) headers['Authorization'] = `Bearer ${this.accessToken}`
      if (this.familyId) headers['X-Family-Id'] = this.familyId

      return fetch(`${BASE_URL}${path}`, {
        method, headers, credentials: 'include',
        body: body ? JSON.stringify(requestToUTC(body)) : undefined,
      })
    }

    let res = await doFetch()

    // ── 401: try refresh once, then retry ────────────────────
    if (res.status === SESSION_EXPIRED_CODE && this.accessToken && path !== '/auth/refresh') {
      const refreshed = await this.refreshAccessToken()

      if (refreshed) {
        // New token obtained → retry the original request
        res = await doFetch()

        // If the retry STILL returns 401 (e.g. user was deleted, token is
        // structurally invalid), session is confirmed dead.
        if (res.status === SESSION_EXPIRED_CODE) {
          this.setToken(null)
          if (!this.sessionExpiredFired) {
            this.sessionExpiredFired = true
            this.onSessionExpired?.()
          }
          throw new Error('Session expired')
        }
      } else if (!this.accessToken) {
        // refreshAccessToken cleared the token (received 401 itself)
        if (!this.sessionExpiredFired) {
          this.sessionExpiredFired = true
          this.onSessionExpired?.()
        }
        throw new Error('Session expired')
      }
      // else: refresh failed due to network — keep old token, fall through
      // to let the original 401 propagate.
    }

    const json: APIResponse<T> = await res.json()
    if (!json.success) {
      if (json.error) {
        throw new ApiRequestError(json.error)
      }
      throw new ApiRequestError({ code: 'INTERNAL_ERROR', summary: 'Unknown error' })
    }
    return responseToLocal(json.data) as T
  }

  get<T>(path: string) { return this.request<T>('GET', path) }
  post<T>(path: string, body?: unknown) { return this.request<T>('POST', path, body) }
  patch<T>(path: string, body?: unknown) { return this.request<T>('PATCH', path, body) }
  put<T>(path: string, body?: unknown) { return this.request<T>('PUT', path, body) }
  delete<T>(path: string) { return this.request<T>('DELETE', path) }

  /** Same as request() but skips requestToUTC / responseToLocal timezone conversion.
   *  Use for configuration data (templates, settings) that is not actual datetimes. */
  private async rawRequest<T>(method: string, path: string, body?: unknown): Promise<T> {
    const hadToken = !!this.accessToken
    if (this.accessToken && this.isTokenExpiringSoon(120)) {
      await this.refreshAccessToken()
    }
    if (hadToken && !this.accessToken) {
      if (!this.sessionExpiredFired) {
        this.sessionExpiredFired = true
        this.onSessionExpired?.()
      }
      throw new Error('Session expired')
    }

    const headers: Record<string, string> = {
      'Content-Type': 'application/json', Accept: 'application/json',
    }
    if (this.accessToken) headers['Authorization'] = `Bearer ${this.accessToken}`
    if (this.familyId) headers['X-Family-Id'] = this.familyId

    let res = await fetch(`${BASE_URL}${path}`, {
      method, headers, credentials: 'include',
      body: body ? JSON.stringify(body) : undefined,
    })

    if (res.status === SESSION_EXPIRED_CODE && this.accessToken && path !== '/auth/refresh') {
      const refreshed = await this.refreshAccessToken()
      if (refreshed) {
        res = await fetch(`${BASE_URL}${path}`, { method, headers, credentials: 'include', body: body ? JSON.stringify(body) : undefined })
        if (res.status === SESSION_EXPIRED_CODE) {
          this.setToken(null)
          if (!this.sessionExpiredFired) { this.sessionExpiredFired = true; this.onSessionExpired?.() }
          throw new Error('Session expired')
        }
      } else if (!this.accessToken) {
        if (!this.sessionExpiredFired) { this.sessionExpiredFired = true; this.onSessionExpired?.() }
        throw new Error('Session expired')
      }
    }

    const json: APIResponse<T> = await res.json()
    if (!json.success) {
      if (json.error) throw new ApiRequestError(json.error)
      throw new ApiRequestError({ code: 'INTERNAL_ERROR', summary: 'Unknown error' })
    }
    return json.data as T
  }

  getRaw<T>(path: string) { return this.rawRequest<T>('GET', path) }
  postRaw<T>(path: string, body?: unknown) { return this.rawRequest<T>('POST', path, body) }
  putRaw<T>(path: string, body?: unknown) { return this.rawRequest<T>('PUT', path, body) }
  deleteRaw<T>(path: string) { return this.rawRequest<T>('DELETE', path) }
}

export const api = new ApiClient()
