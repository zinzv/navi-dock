import { apiRequest, apiUpload } from './http'

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

export function fetchSettings() {
  return apiRequest<AppSettings>('/api/settings')
}

export function updateSettings(body: AppSettings) {
  return apiRequest<AppSettings>('/api/settings', {
    method: 'PUT',
    body: JSON.stringify(body),
  })
}

export function uploadBackground(file: File) {
  return apiUpload<AppSettings>('/api/settings/background', file)
}

export function clearBackground() {
  return apiRequest<AppSettings>('/api/settings/background', {
    method: 'DELETE',
  })
}

export function uploadSiteIcon(file: File) {
  return apiUpload<AppSettings>('/api/settings/site-icon', file)
}

export function clearSiteIcon() {
  return apiRequest<AppSettings>('/api/settings/site-icon', {
    method: 'DELETE',
  })
}

export function fetchNavigation() {
  return apiRequest<{ groups: NavGroup[] }>('/api/navigation')
}

export function fetchGroups() {
  return apiRequest<{ groups: NavGroup[] }>('/api/groups')
}

export function createGroup(body: { name: string; icon?: string }) {
  return apiRequest<NavGroup>('/api/groups', {
    method: 'POST',
    body: JSON.stringify(body),
  })
}

export function updateGroup(id: string, body: { name: string; icon?: string }) {
  return apiRequest<NavGroup>(`/api/groups/${id}`, {
    method: 'PUT',
    body: JSON.stringify(body),
  })
}

export function deleteGroup(id: string) {
  return apiRequest<{ ok: boolean }>(`/api/groups/${id}`, {
    method: 'DELETE',
  })
}

export function sortGroups(ids: string[]) {
  return apiRequest<{ groups: NavGroup[] }>('/api/groups/sort', {
    method: 'PUT',
    body: JSON.stringify({ ids }),
  })
}

export function sortItems(groupId: string, ids: string[]) {
  return apiRequest<{ ok: boolean }>('/api/items/sort', {
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
  return apiRequest<NavItem>(`/api/items/${id}`, {
    method: 'PUT',
    body: JSON.stringify(body),
  })
}

export function createItem(body: UpsertItemPayload) {
  return apiRequest<NavItem>('/api/items', {
    method: 'POST',
    body: JSON.stringify(body),
  })
}

export function deleteItem(id: string) {
  return apiRequest<{ ok: boolean }>(`/api/items/${id}`, {
    method: 'DELETE',
  })
}

export type AssetKind = 'icons' | 'wallpapers'

export interface AssetItem {
  name: string
  url: string
}

export function listAssets(kind: AssetKind) {
  return apiRequest<{ kind: AssetKind; items: AssetItem[] }>(`/api/assets/${kind}`)
}

export function uploadAsset(kind: AssetKind, file: File) {
  return apiUpload<{ kind: AssetKind; name: string; url: string }>(`/api/assets/${kind}`, file)
}

export interface ImportResult {
  ok: boolean
  format?: string
  group_count: number
  item_count: number
}

export async function exportNavigation() {
  const data = await apiRequest<unknown>('/api/export')
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
  return apiRequest<ImportResult>('/api/import', {
    method: 'POST',
    body: JSON.stringify(body),
  })
}
