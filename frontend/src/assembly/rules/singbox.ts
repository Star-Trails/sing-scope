import type { Rule } from '@/types'
import { ruleProviderList, rules } from './index'

const DEFAULT_SINGBOX_RULES = [
  { type: 'Rule', payload: 'protocol: dns', proxy: 'dns-out', size: 0, uuid: 'rule-1', index: 1, extra: { hitCount: 0 } },
  { type: 'Rule', payload: 'ip_is_private', proxy: 'direct', size: 0, uuid: 'rule-2', index: 2, extra: { hitCount: 0 } },
  { type: 'Rule', payload: 'geosite(category-games)', proxy: 'proxy', size: 0, uuid: 'rule-3', index: 3, extra: { hitCount: 0 } },
  { type: 'Rule', payload: 'geoip(cn)', proxy: 'direct', size: 0, uuid: 'rule-4', index: 4, extra: { hitCount: 0 } },
  { type: 'Rule', payload: 'final', proxy: 'proxy', size: 0, uuid: 'rule-5', index: 5, extra: { hitCount: 0 } },
]

export const fetchRules = async () => {
  const appService = (window as any).go?.app?.AppService
  if (appService?.GetRules) {
    try {
      const rList = await appService.GetRules()
      if (rList && rList.length > 0) {
        rules.value = rList.map((r: any, idx: number) => ({
          type: r.type || 'Match',
          payload: r.payload || 'default',
          proxy: r.proxy || 'direct',
          size: r.hitCount || 0,
          uuid: r.uuid || `rule-${idx + 1}`,
          index: r.index || idx + 1,
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

  rules.value = DEFAULT_SINGBOX_RULES.map((r) => ({ ...r })) as Rule[]
  ruleProviderList.value = []
}
