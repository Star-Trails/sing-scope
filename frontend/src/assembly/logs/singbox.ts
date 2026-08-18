// sing-box 后端的日志订阅: 100% 通过 Wails v3 Go Backend (AppService) 获取
import { LOG_LEVEL } from '@/constant'
import type { Log } from '@/types'
import type { LogsSubscription } from './types'

export const subscribeLogs = (
  _params: Record<string, string>,
  onBatch: (batch: Log[]) => void,
): LogsSubscription => {
  const appService = (window as any).go?.app?.AppService
  let timer: number | null = null

  const poll = async () => {
    try {
      if (appService?.GetLogs) {
        const logs = await appService.GetLogs(50)
        if (logs && logs.length) {
          const batch: Log[] = logs.map((l: any) => ({
            type: l.level ? (l.level.toLowerCase() as LOG_LEVEL) : LOG_LEVEL.Info,
            payload: l.message || '',
          }))
          onBatch(batch)
        }
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
