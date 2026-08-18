import type { LogsSubscription } from './types'

export const subscribeLogs = (): LogsSubscription => ({ close: () => {} })
