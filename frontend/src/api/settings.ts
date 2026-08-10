export type ThemeMode = 'light' | 'dark' | 'system'
export type NetworkMode = 'auto' | 'internal' | 'external'

export interface AppSettings {
  site_title: string
  site_icon: string
  language: 'zh' | 'en'
  theme: ThemeMode
  network_mode: NetworkMode
  background_image: string
  background_opacity: number
  clear_background?: boolean
  clear_site_icon?: boolean
}

export interface NavItem {
  id: string
  name: string
  description?: string
  icon: string
  icon_type?: 'iconify' | 'text' | 'image' | string
  /** @deprecated prefer external_url / internal_url */
  url?: string
  external_url?: string
  internal_url?: string
  open_type?: '_blank' | '_self' | string
  status?: string
  group_id?: string
  sort?: number
}

export interface NavGroup {
  id: string
  name: string
  icon?: string
  sort?: number
  items?: NavItem[]
  item_count?: number
}

async function request<T>(url: string, init?: RequestInit): Promise<T> {
  const res = await fetch(url, {
    headers: { 'Content-Type': 'application/json', ...(init?.headers || {}) },
    ...init,
  })
  if (!res.ok) {
    throw new Error(`request failed: ${res.status}`)
  }
  return res.json() as Promise<T>
}

export function fetchSettings() {
  return request<AppSettings>('/api/settings')
}

export function updateSettings(body: AppSettings) {
  return request<AppSettings>('/api/settings', {
    method: 'PUT',
    body: JSON.stringify(body),
  })
}

export async function uploadBackground(file: File) {
  const form = new FormData()
  form.append('file', file)
  const res = await fetch('/api/settings/background', {
    method: 'POST',
    body: form,
  })
  if (!res.ok) {
    throw new Error(`upload failed: ${res.status}`)
  }
  return res.json() as Promise<AppSettings>
}

export function clearBackground() {
  return request<AppSettings>('/api/settings/background', {
    method: 'DELETE',
  })
}

export function clearSiteIcon() {
  return request<AppSettings>('/api/settings/site-icon', {
    method: 'DELETE',
  })
}

export function fetchNavigation() {
  return request<{ groups: NavGroup[] }>('/api/navigation')
}

export function fetchGroups() {
  return request<{ groups: NavGroup[] }>('/api/groups')
}

export function createGroup(body: { name: string; icon?: string }) {
  return request<NavGroup>('/api/groups', {
    method: 'POST',
    body: JSON.stringify(body),
  })
}

export function updateGroup(id: string, body: { name: string; icon?: string }) {
  return request<NavGroup>(`/api/groups/${id}`, {
    method: 'PUT',
    body: JSON.stringify(body),
  })
}

export function deleteGroup(id: string) {
  return request<{ ok: boolean }>(`/api/groups/${id}`, {
    method: 'DELETE',
  })
}

export function sortGroups(ids: string[]) {
  return request<{ groups: NavGroup[] }>('/api/groups/sort', {
    method: 'PUT',
    body: JSON.stringify({ ids }),
  })
}

export function sortItems(groupId: string, ids: string[]) {
  return request<{ ok: boolean }>('/api/items/sort', {
    method: 'PUT',
    body: JSON.stringify({ group_id: groupId, ids }),
  })
}

export type UpsertItemPayload = {
  group_id?: string
  name: string
  description?: string
  icon_type?: string
  icon?: string
  internal_url?: string
  external_url?: string
  open_type?: string
}

export function updateItem(id: string, body: UpsertItemPayload) {
  return request<NavItem>(`/api/items/${id}`, {
    method: 'PUT',
    body: JSON.stringify(body),
  })
}

export function createItem(body: UpsertItemPayload) {
  return request<NavItem>('/api/items', {
    method: 'POST',
    body: JSON.stringify(body),
  })
}

export function deleteItem(id: string) {
  return request<{ ok: boolean }>(`/api/items/${id}`, {
    method: 'DELETE',
  })
}

export type AssetKind = 'icons' | 'wallpapers'

export interface AssetItem {
  name: string
  url: string
}

export function listAssets(kind: AssetKind) {
  return request<{ kind: AssetKind; items: AssetItem[] }>(`/api/assets/${kind}`)
}

export async function uploadAsset(kind: AssetKind, file: File) {
  const form = new FormData()
  form.append('file', file)
  const res = await fetch(`/api/assets/${kind}`, {
    method: 'POST',
    body: form,
  })
  if (!res.ok) {
    throw new Error(`upload failed: ${res.status}`)
  }
  return res.json() as Promise<{ kind: AssetKind; name: string; url: string }>
}

export interface ImportResult {
  ok: boolean
  format?: string
  group_count: number
  item_count: number
}

export async function exportNavigation() {
  const res = await fetch('/api/export')
  if (!res.ok) {
    throw new Error(`export failed: ${res.status}`)
  }
  const data = await res.json()
  const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  const stamp = new Date().toISOString().slice(0, 19).replace(/[:T]/g, '-')
  a.href = url
  a.download = `navi-dock-export-${stamp}.json`
  document.body.appendChild(a)
  a.click()
  a.remove()
  URL.revokeObjectURL(url)
  return data
}

export async function importNavigation(file: File) {
  const text = await file.text()
  let body: unknown
  try {
    body = JSON.parse(text)
  } catch {
    throw new Error('invalid json')
  }
  const res = await fetch('/api/import', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
  if (!res.ok) {
    let message = `import failed: ${res.status}`
    try {
      const err = (await res.json()) as { message?: string }
      if (err.message) message = err.message
    } catch {
      /* ignore */
    }
    throw new Error(message)
  }
  return res.json() as Promise<ImportResult>
}
