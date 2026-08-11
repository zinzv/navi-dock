import { getAuthToken, clearAuthToken } from './auth'

export class ApiError extends Error {
  status: number
  constructor(message: string, status: number) {
    super(message)
    this.status = status
  }
}

async function parseError(res: Response): Promise<string> {
  try {
    const data = (await res.json()) as { message?: string }
    if (data?.message) return data.message
  } catch {
    /* ignore */
  }
  return `request failed: ${res.status}`
}

export async function apiRequest<T>(url: string, init?: RequestInit): Promise<T> {
  const token = getAuthToken()
  const headers = new Headers(init?.headers || {})
  if (!headers.has('Content-Type') && !(init?.body instanceof FormData)) {
    headers.set('Content-Type', 'application/json')
  }
  if (token) headers.set('Authorization', `Bearer ${token}`)

  const res = await fetch(url, { ...init, headers })
  if (res.status === 401) {
    clearAuthToken()
    if (!window.location.pathname.startsWith('/login') && !window.location.pathname.startsWith('/setup')) {
      window.location.assign('/login')
    }
    throw new ApiError(await parseError(res), 401)
  }
  if (!res.ok) {
    throw new ApiError(await parseError(res), res.status)
  }
  if (res.status === 204) {
    return undefined as T
  }
  return res.json() as Promise<T>
}

export async function apiUpload<T>(url: string, file: File): Promise<T> {
  const form = new FormData()
  form.append('file', file)
  return apiRequest<T>(url, { method: 'POST', body: form })
}
