// sing-box 后端的 config 组装:通过 Wails v3 Go Backend (AppService) 管理 Clash Mode 状态
import type { Config } from '@/types'
import { configs, defaultConfig } from './index'

const fetchSingboxConfigs = async (): Promise<Config> => {
  const appService = (window as any).go?.app?.AppService
  if (appService?.GetClashModeStatus) {
    try {
      const status = await appService.GetClashModeStatus()
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
  }
  return { ...defaultConfig }
}

export const fetchConfigs = async () => {
  configs.value = await fetchSingboxConfigs()
}

export const updateConfigs = async (cfg: Record<string, string | boolean | object | number>) => {
  if (typeof cfg.mode === 'string') {
    const appService = (window as any).go?.app?.AppService
    if (appService?.SetClashMode) {
      await appService.SetClashMode(cfg.mode)
    }
  }
  fetchConfigs()
}
