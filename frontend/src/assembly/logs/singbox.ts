// sing-box 后端的日志订阅: 100% 通过 Wails v3 Go Backend (AppService) 获取
import { LOG_LEVEL } from '@/constant'
import type { Log } from '@/types'
import { Backend } from '@/utils/backend'
import type { LogsSubscription } from './types'

export const subscribeLogs = (
  _params: Record<string, string>,
  onBatch: (batch: Log[]) => void,
): LogsSubscription => {
  let timer: number | null = null
  const seenMessages = new Set<string>()

  const poll = async () => {
    try {
      const logs = await Backend.getLogs(200)
      if (logs && logs.length > 0) {
        const newLogs: Log[] = []
        for (const l of logs) {
          const key = `${l.timestamp || ''}_${l.message || ''}`
          if (!seenMessages.has(key)) {
            seenMessages.add(key)
            if (seenMessages.size > 2000) {
              const firstKey = seenMessages.values().next().value
              if (firstKey) seenMessages.delete(firstKey)
            }
            newLogs.push({
              type: l.level ? (l.level.toLowerCase() as LOG_LEVEL) : LOG_LEVEL.Info,
              payload: l.message || '',
            })
          }
        }
        if (newLogs.length > 0) {
          onBatch(newLogs)
        }
      }
    } catch {
      // ignore
    }
  }

  poll()
  timer = window.setInterval(poll, 1000)

  return {
    close: () => {
      clearInterval(timer!)
      seenMessages.clear()
    },
  }
}
