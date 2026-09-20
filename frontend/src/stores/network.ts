import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { fetchNetworkConfig } from '../api/settings'
import {
  LAN_PROBE_TIMEOUT_MS,
  MEMORY_TTL_MS,
  buildProbeURL,
  probeLAN,
  readNetworkCache,
  writeNetworkCache,
  type NetworkStatus,
  type ReachableNetwork,
} from '../services/network'

const EVENT_DEBOUNCE_MS = 2_000
const FAILURE_RETRY_MS = 2_000

export const useNetworkStore = defineStore('network', () => {
  const status = ref<NetworkStatus>('unknown')
  const lastKnown = ref<ReachableNetwork | null>(null)
  const lastCheckedAt = ref(0)
  const probeDomain = ref('')
  const probeUrl = ref('')
  const configured = ref(false)

  let started = false
  let detectPromise: Promise<ReachableNetwork> | null = null
  let retryTimer = 0
  let eventTimer = 0

  const effectiveNetwork = computed<ReachableNetwork>(
    () =>
      (status.value === 'internal' || status.value === 'external'
        ? status.value
        : lastKnown.value) || 'external',
  )

  const canDetect = computed(() => Boolean(probeUrl.value))

  function restoreCache() {
    const cached = readNetworkCache()
    if (!cached) return
    lastKnown.value = cached.network
    status.value = cached.network
    lastCheckedAt.value = cached.checkedAt
  }

  function commitResult(network: ReachableNetwork) {
    const checkedAt = Date.now()
    status.value = network
    lastKnown.value = network
    lastCheckedAt.value = checkedAt
    writeNetworkCache({ network, checkedAt })
  }

  function scheduleFailureRetry() {
    window.clearTimeout(retryTimer)
    retryTimer = window.setTimeout(() => {
      void detect(true, false)
    }, FAILURE_RETRY_MS)
  }

  async function detect(force = false, allowRetry = true): Promise<ReachableNetwork> {
    if (!probeUrl.value) {
      status.value = lastKnown.value || 'unknown'
      return effectiveNetwork.value
    }

    if (!force && Date.now() - lastCheckedAt.value < MEMORY_TTL_MS) {
      return effectiveNetwork.value
    }
    if (detectPromise) return detectPromise

    status.value = 'checking'
    detectPromise = (async () => {
      const reachable = await probeLAN(probeUrl.value, LAN_PROBE_TIMEOUT_MS)
      const network: ReachableNetwork = reachable ? 'internal' : 'external'
      commitResult(network)
      if (!reachable && allowRetry) scheduleFailureRetry()
      return network
    })()

    try {
      return await detectPromise
    } finally {
      detectPromise = null
    }
  }

  function scheduleDetection(force = false) {
    window.clearTimeout(eventTimer)
    eventTimer = window.setTimeout(() => {
      void detect(force)
    }, EVENT_DEBOUNCE_MS)
  }

  async function loadConfig() {
    try {
      const config = await fetchNetworkConfig()
      probeDomain.value = (config.probe_domain || '').trim()
      probeUrl.value = buildProbeURL(probeDomain.value)
    } catch {
      probeDomain.value = ''
      probeUrl.value = ''
    } finally {
      configured.value = true
    }
  }

  async function start() {
    if (started) return
    started = true
    restoreCache()
    await loadConfig()
    void detect()

    window.addEventListener('focus', () => scheduleDetection())
    window.addEventListener('online', () => scheduleDetection(true))
    window.addEventListener('offline', () => {
      status.value = 'unknown'
    })
    document.addEventListener('visibilitychange', () => {
      if (document.visibilityState === 'visible') scheduleDetection()
    })
  }

  async function refresh() {
    window.clearTimeout(retryTimer)
    return detect(true)
  }

  return {
    status,
    lastKnown,
    lastCheckedAt,
    probeDomain,
    probeUrl,
    configured,
    canDetect,
    effectiveNetwork,
    start,
    refresh,
  }
})
