import type { APIResponse, User } from '@/types'

const BASE_URL = '/api'
const TOKEN_KEY = 'na_access_token'
const SESSION_EXPIRED_CODE = 401

// ─── Token persistence (sessionStorage: survives refresh, cleared on tab close) ─

function loadToken(): string | null {
  try { return sessionStorage.getItem(TOKEN_KEY) } catch { return null }
}
function saveToken(token: string | null) {
  try { if (token) sessionStorage.setItem(TOKEN_KEY, token); else sessionStorage.removeItem(TOKEN_KEY) } catch { /* */ }
}

// ─── JWT helpers ─────────────────────────────────────────────────

function decodeJWT(token: string): { exp?: number } | null {
  try {
    const payload = token.split('.')[1]
    return JSON.parse(atob(payload))
  } catch { return null }
}

function isTokenExpired(token: string): boolean {
  const claims = decodeJWT(token)
  if (!claims?.exp) return false
  return Date.now() >= claims.exp * 1000
}

function isTokenExpiringSoon(token: string, seconds = 60): boolean {
  const claims = decodeJWT(token)
  if (!claims?.exp) return false
  return Date.now() >= (claims.exp - seconds) * 1000
}

// ─── API Client ──────────────────────────────────────────────────

class ApiClient {
  private accessToken: string | null = null
  private refreshPromise: Promise<boolean> | null = null
  private onSessionExpired: (() => void) | null = null
  /** Prevent concurrent expired-session redirects. */
  private sessionExpiredFired = false

  constructor() {
    // Restore token from sessionStorage on page refresh
    const saved = loadToken()
    if (saved && !isTokenExpired(saved)) {
      this.accessToken = saved
    }
  }

  private setToken(token: string | null) {
    this.accessToken = token
    saveToken(token)
    if (token) this.sessionExpiredFired = false // reset on new token
  }

  setAccessToken(token: string | null) { this.setToken(token) }
  getAccessToken() { return this.accessToken }

  /** True if we have a non-expired access token (memory or sessionStorage). */
  hasValidToken(): boolean {
    return !!this.accessToken && !isTokenExpired(this.accessToken)
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
        this.setToken(json.data.access_token)
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
          this.setToken(json.data.access_token)
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
    // Proactively refresh if token is about to expire.
    // If refresh fails (network error), keep the old token and let the actual
    // API call decide — it might still be valid, or it will return 401.
    if (this.accessToken && isTokenExpiringSoon(this.accessToken, 120)) {
      await this.refreshAccessToken()
      // Note: we do NOT check the result here. If refresh cleared the token
      // (401), the next API call will get 401 and the handler below will
      // trigger the logout flow. If refresh failed due to network, the old
      // token might still work.
    }

    const doFetch = async () => {
      const headers: Record<string, string> = {
        'Content-Type': 'application/json', Accept: 'application/json',
      }
      if (this.accessToken) headers['Authorization'] = `Bearer ${this.accessToken}`

      return fetch(`${BASE_URL}${path}`, {
        method, headers, credentials: 'include',
        body: body ? JSON.stringify(body) : undefined,
      })
    }

    let res = await doFetch()

    // ── 401: token expired → try refresh once ─────────────────
    // Guards:
    //  - Only on 401 (not 403, 5xx, etc.)
    //  - Must have a token to refresh (otherwise we'd loop on public endpoints)
    //  - Never refresh the refresh endpoint itself
    if (res.status === SESSION_EXPIRED_CODE && this.accessToken && path !== '/auth/refresh') {
      const refreshed = await this.refreshAccessToken()

      if (refreshed) {
        // New token obtained → retry the original request once
        res = await doFetch()
      } else if (!this.accessToken) {
        // Token was cleared by refreshAccessToken (received 401 from /auth/refresh).
        // Session is confirmed expired — redirect to login exactly once.
        if (!this.sessionExpiredFired) {
          this.sessionExpiredFired = true
          this.onSessionExpired?.()
        }
        throw new Error('Session expired')
      }
      // else: refresh failed due to network/server error, token kept.
      // Fall through to let the original 401 response propagate naturally.
    }

    const json: APIResponse<T> = await res.json()
    if (!json.success) throw new Error(json.error || 'Unknown error')
    return json.data
  }

  get<T>(path: string) { return this.request<T>('GET', path) }
  post<T>(path: string, body?: unknown) { return this.request<T>('POST', path, body) }
  patch<T>(path: string, body?: unknown) { return this.request<T>('PATCH', path, body) }
  put<T>(path: string, body?: unknown) { return this.request<T>('PUT', path, body) }
  delete<T>(path: string) { return this.request<T>('DELETE', path) }
}

export const api = new ApiClient()
