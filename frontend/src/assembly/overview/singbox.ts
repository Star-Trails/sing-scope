// sing-box 后端的概览统计组装: 100% 通过 Wails v3 Go Backend (AppService) 获取实时状态
import { ref, watch, type Ref } from 'vue'

export interface SingboxStream<T> {
  data: Ref<T | undefined>
  close: () => void
}

type StatusListener = (status: any) => void

const statusListeners = new Set<StatusListener>()
let latestStatus: any = null
let pollTimer: number | null = null

const closeSharedStatusStream = () => {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
  latestStatus = null
}

const ensureSharedStatusStream = () => {
  if (pollTimer) return true

  const appService = (window as any).go?.app?.AppService
  const poll = async () => {
    try {
      if (appService?.GetSystemStatus) {
        const [sys, ov] = await Promise.all([
          appService.GetSystemStatus(),
          appService.GetOverviewSummary ? appService.GetOverviewSummary('') : null,
        ])
        latestStatus = {
          memory: BigInt(sys?.memory || 0),
          goroutines: sys?.goroutines || 0,
          downlink: BigInt(Math.round(ov?.downloadRate || sys?.downlink || 0)),
          uplink: BigInt(Math.round(ov?.uploadRate || sys?.uplink || 0)),
          downlinkTotal: BigInt(ov?.sessionDownload || sys?.downlinkTotal || 0),
          uplinkTotal: BigInt(ov?.sessionUpload || sys?.uplinkTotal || 0),
        }
        statusListeners.forEach((listener) => listener(latestStatus))
      }
    } catch {
      // ignore
    }
  }

  poll()
  pollTimer = window.setInterval(poll, 1000)
  return true
}

const subscribeSingboxStatus = <T>(map: (status: any) => T): SingboxStream<T> | null => {
  const data = ref<T>()
  const listener: StatusListener = (status) => {
    data.value = map(status)
  }

  statusListeners.add(listener)
  ensureSharedStatusStream()
  if (latestStatus) listener(latestStatus)

  return {
    data,
    close: () => {
      statusListeners.delete(listener)
      if (statusListeners.size === 0) closeSharedStatusStream()
    },
  }
}

const createSingboxStat = <T>(kind: 'memory' | 'traffic'): SingboxStream<T> => {
  const data = ref<T>()
  const sub =
    kind === 'memory'
      ? subscribeSingboxStatus((status) => ({
          inuse: Number(status?.memory || 0),
          goroutines: status?.goroutines || 0,
        }))
      : subscribeSingboxStatus((status) => ({
          down: Number(status?.downlink || 0),
          up: Number(status?.uplink || 0),
          downTotal: Number(status?.downlinkTotal || 0),
          upTotal: Number(status?.uplinkTotal || 0),
        }))

  if (!sub) return { data, close: () => {} }
  watch(sub.data, (value) => (data.value = value as T), { immediate: true })
  return { data, close: sub.close }
}

export const fetchMemoryAPI = <T>() => createSingboxStat<T>('memory')

export const fetchTrafficAPI = <T>() => createSingboxStat<T>('traffic')
