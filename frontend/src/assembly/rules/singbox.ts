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
  if (appService?.GetBatchAnalytics) {
    try {
      const res = await appService.GetBatchAnalytics('', 100)
      if (res && res.byRule && res.byRule.length > 0) {
        rules.value = res.byRule.map((r: any, idx: number) => ({
          type: 'Match',
          payload: r.name || 'default',
          proxy: r.category || 'direct',
          size: r.connectionCount || 0,
          uuid: `rule-${idx + 1}`,
          index: idx + 1,
          extra: {
            hitCount: r.connectionCount || 0,
            hitAt: r.lastActiveAt ? new Date(r.lastActiveAt).toISOString() : new Date().toISOString(),
          },
        })) as Rule[]
        ruleProviderList.value = []
        return
      }
    } catch {
      // fallback
    }
  }

  rules.value = DEFAULT_SINGBOX_RULES.map((r) => ({ ...r })) as Rule[]
  ruleProviderList.value = []
}
