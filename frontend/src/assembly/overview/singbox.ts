// sing-box 后端的概览统计组装:优先通过 Wails v3 Go Backend (AppService) 获取实时状态,
// 兜底回退到 gRPC SubscribeStatus 订阅。
import { subscribeStream } from '@/api/singbox/subscriptions'
import type { Status } from '@/gen/daemon/started_service_pb'
import { ref, watch, type Ref } from 'vue'

export interface SingboxStream<T> {
  data: Ref<T | undefined>
  close: () => void
}

type StatusListener = (status: any) => void

const statusListeners = new Set<StatusListener>()
let statusHandle: { close: () => void } | null = null
let latestStatus: any = null
let pollTimer: number | null = null

const closeSharedStatusStream = () => {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
  statusHandle?.close()
  statusHandle = null
  latestStatus = null
}

const ensureSharedStatusStream = () => {
  if (statusHandle || pollTimer) return true

  // Check if running inside Wails v3
  const appService = (window as any).go?.app?.AppService
  if (appService?.GetSystemStatus) {
    const poll = async () => {
      try {
        const [sys, ov] = await Promise.all([
          appService.GetSystemStatus(),
          appService.GetOverviewSummary ? appService.GetOverviewSummary('') : null,
        ])
        latestStatus = {
          memory: BigInt(sys.memory || 0),
          goroutines: sys.goroutines || 0,
          downlink: BigInt(Math.round(ov?.downloadRate || sys.downlink || 0)),
          uplink: BigInt(Math.round(ov?.uploadRate || sys.uplink || 0)),
          downlinkTotal: BigInt(ov?.sessionDownload || sys.downlinkTotal || 0),
          uplinkTotal: BigInt(ov?.sessionUpload || sys.uplinkTotal || 0),
        }
        statusListeners.forEach((listener) => listener(latestStatus))
      } catch {
        // ignore
      }
    }
    poll()
    pollTimer = window.setInterval(poll, 1000)
    return true
  }

  statusHandle = subscribeStream<Status>('status', (status) => {
    latestStatus = status
    statusListeners.forEach((listener) => listener(status))
  })

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
          inuse: Number(status.memory),
          goroutines: status.goroutines,
        }))
      : subscribeSingboxStatus((status) => ({
          down: Number(status.downlink),
          up: Number(status.uplink),
          downTotal: Number(status.downlinkTotal),
          upTotal: Number(status.uplinkTotal),
        }))

  if (!sub) return { data, close: () => {} }
  watch(sub.data, (value) => (data.value = value as T), { immediate: true })
  return { data, close: sub.close }
}

export const fetchMemoryAPI = <T>() => createSingboxStat<T>('memory')

export const fetchTrafficAPI = <T>() => createSingboxStat<T>('traffic')
