import { configs, defaultConfig } from './index'

export const fetchConfigs = async () => {
  configs.value = { ...defaultConfig }
}

export const updateConfigs = async (_cfg: Record<string, string | boolean | object | number>) => {
  fetchConfigs()
}
