import type { Rule } from '@/types'
import { ruleProviderList, rules } from './index'

export const fetchRules = async () => {
  const appService = (window as any).go?.app?.AppService
  if (appService?.GetRules) {
    try {
      const rList = await appService.GetRules()
      if (rList && rList.length > 0) {
        rules.value = rList.map((r: any) => ({
          type: r.type || 'Rule',
          payload: r.payload || 'default',
          proxy: r.proxy || 'direct',
          size: r.hitCount || 0,
          uuid: r.uuid,
          index: r.index,
          extra: {
            hitCount: r.hitCount || 0,
            hitAt: r.lastHitAt > 0 ? new Date(r.lastHitAt).toISOString() : new Date().toISOString(),
          },
        })) as Rule[]
        ruleProviderList.value = []
        return
      }
    } catch {
      // ignore
    }
  }

  rules.value = []
  ruleProviderList.value = []
}
