import type { APIResponse, User } from '@/types'

const BASE_URL = '/api'

class ApiClient {
  private accessToken: string | null = null
  private refreshPromise: Promise<boolean> | null = null
  private onSessionExpired: (() => void) | null = null

  setAccessToken(token: string | null) { this.accessToken = token }
  getAccessToken() { return this.accessToken }

  /** Register callback: called when refresh token is also expired. */
  onExpired(fn: () => void) { this.onSessionExpired = fn }

  /** Try to restore session from refresh token cookie. Returns user if successful. */
  async initSession(): Promise<User | null> {
    try {
      const res = await fetch(`${BASE_URL}/auth/refresh`, {
        method: 'POST', credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
      })
      if (!res.ok) return null
      const json: APIResponse<{ access_token: string; user: User }> = await res.json()
      if (json.success && json.data?.access_token) {
        this.accessToken = json.data.access_token
        return json.data.user
      }
      return null
    } catch { return null }
  }

  /** Refresh access token using the httpOnly cookie. Returns true if succeeded. */
  async refreshAccessToken(): Promise<boolean> {
    if (this.refreshPromise) return this.refreshPromise
    this.refreshPromise = (async () => {
      try {
        const res = await fetch(`${BASE_URL}/auth/refresh`, {
          method: 'POST', credentials: 'include',
          headers: { 'Content-Type': 'application/json' },
        })
        if (!res.ok) { this.accessToken = null; return false }
        const json: APIResponse<{ access_token: string }> = await res.json()
        if (json.success && json.data?.access_token) {
          this.accessToken = json.data.access_token
          return true
        }
        return false
      } catch { this.accessToken = null; return false
      } finally { this.refreshPromise = null }
    })()
    return this.refreshPromise
  }

  private async request<T>(method: string, path: string, body?: unknown): Promise<T> {
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

    // Auto-refresh on 401
    if (res.status === 401 && this.accessToken) {
      const refreshed = await this.refreshAccessToken()
      if (refreshed) {
        res = await doFetch()
      } else {
        this.accessToken = null
        this.onSessionExpired?.()
        throw new Error('Session expired')
      }
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
