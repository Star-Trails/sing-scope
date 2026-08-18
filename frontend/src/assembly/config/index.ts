import type { Config } from '@/types'
import { ref } from 'vue'
import { fetchConfigs as fetchSingboxConfigs, updateConfigs as updateSingboxConfigs } from './singbox'

export const defaultConfig: Config = {
  port: 0,
  'socks-port': 0,
  'redir-port': 0,
  'tproxy-port': 0,
  'mixed-port': 0,
  'allow-lan': false,
  'bind-address': '',
  mode: '',
  'mode-list': [],
  modes: [],
  'log-level': '',
  ipv6: false,
  tun: {
    enable: false,
  },
}

export const configs = ref<Config>({ ...defaultConfig })

export const fetchConfigs = fetchSingboxConfigs
export const updateConfigs = updateSingboxConfigs

export const flushDNSCacheAPI = async () => {}
export const flushFakeIPAPI = async () => {}
export const queryDNSAPI = async (_params?: unknown) => ({ data: { Answer: [] as any[] } })
export const reloadConfigsAPI = async () => {}
export const updateConfigsAPI = async (_cfg?: unknown, _force?: unknown) => {}
export const updateGeoDataAPI = async () => {}
export const upgradeCoreAPI = async (_type?: unknown) => {}
