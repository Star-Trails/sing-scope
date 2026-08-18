import { displayAllFeatures } from '@/store/settings'
import { activeBackend } from '@/store/setup'
import type { Backend } from '@/types'
import { computed, ref } from 'vue'

export enum Channel {
  Clash = 'clash',
  Singbox = 'singbox',
}

export enum Core {
  Mihomo = 'mihomo',
  Singbox = 'singbox',
  Honk = 'honk',
  Unknown = 'unknown',
}

export const channel = computed<Channel>(() => Channel.Singbox)

export const core = ref<Core>(Core.Singbox)
export const apiVersion = ref(4)

export const resetCore = () => {
  core.value = Core.Singbox
  apiVersion.value = 4
}

const isNonMihomoClashCore = computed(
  () =>
    channel.value === Channel.Clash && (core.value === Core.Singbox || core.value === Core.Honk),
)


export const showDisplayAllFeatures = computed(
  () => !!activeBackend.value && isNonMihomoClashCore.value,
)

const hard = computed(() => {
  return {
    rules: true,
    dnsQuery: false,
    dnsFlush: false,
    fakeIPFlush: false,
    coreActions: false,
    dashboardUpgrade: false,

    tools: false,
    goroutines: true,
    startedAt: true,
    usbip: false,
    openvpn: false,
  }
})

const soft = computed(() => {
  return {
    coreUpgrade: false,
    coreRestart: false,
    reloadConfigs: false,
    updateConfigs: false,
    updateGeoDatabase: false,
    syncSettings: false,
    independentLatency: false,
    coreUpdateCheck: false,
    configPatch: false,

    customGlobalNode: true,
    logTypeFilter: true,
    logConnectionDetail: true,
    disconnectOnModeChange: true,

    traceLogLevel: true,
    extraLogLevels: true,
    silentLogLevel: true,
  }
})

type HardCaps = typeof hard.value
type SoftCaps = typeof soft.value

export type HardCap = keyof HardCaps
export type SoftCap = keyof SoftCaps
export type Cap = HardCap | SoftCap

export const can = (cap: Cap): boolean => {
  if (!activeBackend.value) return false
  const hardCaps = hard.value
  if (cap in hardCaps) return hardCaps[cap as HardCap]
  return soft.value[cap as SoftCap]
}

export const isSingboxChannelAvailable = async (backend: Backend, _timeout: number = 10000) => {
  const appService = (window as any).go?.app?.AppService
  if (appService?.ConnectServer) {
    const url = `${backend.protocol || 'http'}://${backend.host}:${backend.port || '9090'}`
    return await appService.ConnectServer(url, backend.password || '')
  }
  return true
}

export const isBackendAvailable = (backend: Backend, timeout: number = 10000) =>
  isSingboxChannelAvailable(backend, timeout)
