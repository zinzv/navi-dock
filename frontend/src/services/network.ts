import type { NavItem, NetworkMode } from '../api/settings'

export type ReachableNetwork = 'internal' | 'external'
export type NetworkStatus = 'unknown' | 'checking' | ReachableNetwork

export type NetworkCache = {
  network: ReachableNetwork
  checkedAt: number
}

export type ResolvedNavigation = {
  url: string
  network: ReachableNetwork
  lanOnly: boolean
}

const CACHE_KEY = 'navidock:detected-network'
export const MEMORY_TTL_MS = 30_000
export const SESSION_TTL_MS = 5 * 60_000
export const LAN_PROBE_TIMEOUT_MS = 1_000
export const LAN_PROBE_PATH = '/api/network/ping'

/** Keep host only, e.g. lan.zeven.site */
export function normalizeProbeDomain(value: string): string {
  let input = value.trim()
  if (!input) return ''
  input = input.replace(/^https?:\/\//i, '').replace(/^\/\//, '')
  const slash = input.indexOf('/')
  if (slash >= 0) input = input.slice(0, slash)
  return input.replace(/^\.+|\.+$/g, '')
}

/** Build full probe URL from domain; path is fixed. */
export function buildProbeURL(domain: string): string {
  const host = normalizeProbeDomain(domain)
  if (!host) return ''
  return `${window.location.protocol}//${host}${LAN_PROBE_PATH}`
}

export function readNetworkCache(now = Date.now()): NetworkCache | null {
  try {
    const raw = sessionStorage.getItem(CACHE_KEY)
    if (!raw) return null
    const value = JSON.parse(raw) as Partial<NetworkCache>
    if (
      (value.network !== 'internal' && value.network !== 'external') ||
      typeof value.checkedAt !== 'number' ||
      now - value.checkedAt > SESSION_TTL_MS
    ) {
      sessionStorage.removeItem(CACHE_KEY)
      return null
    }
    return value as NetworkCache
  } catch {
    return null
  }
}

export function writeNetworkCache(value: NetworkCache) {
  try {
    sessionStorage.setItem(CACHE_KEY, JSON.stringify(value))
  } catch {
    // Storage can be unavailable in privacy-restricted contexts.
  }
}

export async function probeLAN(probeUrl: string, timeoutMs: number): Promise<boolean> {
  if (!probeUrl.trim()) return false
  const controller = new AbortController()
  const timer = window.setTimeout(() => controller.abort(), timeoutMs)
  try {
    const separator = probeUrl.includes('?') ? '&' : '?'
    const response = await fetch(`${probeUrl}${separator}t=${Date.now()}`, {
      method: 'GET',
      cache: 'no-store',
      credentials: 'omit',
      signal: controller.signal,
    })
    return response.ok
  } catch {
    return false
  } finally {
    window.clearTimeout(timer)
  }
}

function externalURL(item: NavItem) {
  return (item.external_url || item.url || '').trim()
}

function internalURL(item: NavItem) {
  return (item.internal_url || '').trim()
}

export function resolveNavigationURL(
  item: NavItem,
  mode: NetworkMode,
  detected: ReachableNetwork,
): ResolvedNavigation {
  const external = externalURL(item)
  const internal = internalURL(item)

  if (mode === 'internal') {
    return {
      url: internal || external,
      network: internal ? 'internal' : 'external',
      lanOnly: false,
    }
  }

  if (mode === 'external') {
    return {
      url: external || internal,
      network: external ? 'external' : 'internal',
      lanOnly: false,
    }
  }

  if (detected === 'internal') {
    return {
      url: internal || external,
      network: internal ? 'internal' : 'external',
      lanOnly: false,
    }
  }

  if (!external && internal) {
    return { url: '', network: 'internal', lanOnly: true }
  }

  return {
    url: external || internal,
    network: external ? 'external' : 'internal',
    lanOnly: false,
  }
}
