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

  const poll = async () => {
    try {
      const logs = await Backend.getLogs(50)
      if (logs && logs.length) {
        const batch: Log[] = logs.map((l) => ({
          type: l.level ? (l.level.toLowerCase() as LOG_LEVEL) : LOG_LEVEL.Info,
          payload: l.message || '',
        }))
        onBatch(batch)
      }
    } catch {
      // ignore
    }
  }

  poll()
  timer = window.setInterval(poll, 1500)

  return {
    close: () => {
      if (timer) clearInterval(timer)
    },
  }
}
