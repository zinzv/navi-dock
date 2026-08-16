export type HealthInfo = {
  status: string
  service: string
  version?: string
}

export async function fetchAppVersion(): Promise<string> {
  try {
    const res = await fetch('/api/health')
    if (!res.ok) return 'unknown'
    const data = (await res.json()) as HealthInfo
    return data.version?.trim() || 'unknown'
  } catch {
    return 'unknown'
  }
}
