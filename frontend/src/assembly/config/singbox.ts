import type { Config } from '@/types'
import { Backend } from '@/utils/backend'
import { configs, defaultConfig } from './index'

const fetchSingboxConfigs = async (): Promise<Config> => {
  try {
    const status = await Backend.getClashModeStatus()
    if (status) {
      return {
        ...defaultConfig,
        mode: status.currentMode || 'rule',
        'mode-list': status.modeList || ['rule', 'global', 'direct'],
        modes: status.modeList || ['rule', 'global', 'direct'],
      }
    }
  } catch {
    // fallback
  }
  return { ...defaultConfig }
}

export const fetchConfigs = async () => {
  configs.value = await fetchSingboxConfigs()
}

export const updateConfigs = async (cfg: Record<string, string | boolean | object | number>) => {
  if (typeof cfg.mode === 'string') {
    await Backend.setClashMode(cfg.mode)
  }
  fetchConfigs()
}
