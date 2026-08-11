const TOKEN_KEY = 'navi-dock-auth-token'

export interface UserPublic {
  id: string
  username: string
  role: string
  created_at: string
}

export interface AuthStatus {
  setup_required: boolean
  user?: UserPublic
}

export interface LoginResult {
  token: string
  user: UserPublic
}

export function getAuthToken(): string {
  return localStorage.getItem(TOKEN_KEY) || ''
}

export function setAuthToken(token: string) {
  if (token) localStorage.setItem(TOKEN_KEY, token)
  else localStorage.removeItem(TOKEN_KEY)
}

export function clearAuthToken() {
  localStorage.removeItem(TOKEN_KEY)
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

async function request<T>(url: string, init?: RequestInit): Promise<T> {
  const token = getAuthToken()
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(init?.headers as Record<string, string> | undefined),
  }
  if (token) headers.Authorization = `Bearer ${token}`

  const res = await fetch(url, { ...init, headers })
  if (!res.ok) {
    throw new Error(await parseError(res))
  }
  return res.json() as Promise<T>
}

export function fetchAuthStatus() {
  return request<AuthStatus>('/api/auth/status')
}

export function login(username: string, password: string) {
  return request<LoginResult>('/api/auth/login', {
    method: 'POST',
    body: JSON.stringify({ username, password }),
  })
}

export function changePassword(oldPassword: string, newPassword: string) {
  return request<{ ok: boolean }>('/api/auth/password', {
    method: 'PUT',
    body: JSON.stringify({ old_password: oldPassword, new_password: newPassword }),
  })
}

export function fetchUsers() {
  return request<{ users: UserPublic[] }>('/api/users')
}

export function createUser(username: string, password: string, role = 'admin') {
  return request<UserPublic>('/api/users', {
    method: 'POST',
    body: JSON.stringify({ username, password, role }),
  })
}

export function deleteUser(id: string) {
  return request<{ ok: boolean }>(`/api/users/${id}`, {
    method: 'DELETE',
  })
}
