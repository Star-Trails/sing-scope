// 组装层 · 版本与状态。
// 100% 通过 Wails v3 Go Backend (AppService) 同步获取。
import HonkLogo from '@/assets/images/honk.svg'
import MetacubexLogo from '@/assets/images/metacubex.jpg'
import SingBoxLogo from '@/assets/images/sing-box.svg'
import { MIHOMO, MIHOMO_CHANNEL } from '@/constant'
import { activeBackend } from '@/store/setup'
import type { Backend } from '@/types'
import { computed, nextTick, ref } from 'vue'
import { apiVersion, can, core, Core, resetCore } from './backend'

export const version = ref('sing-box 1.14')
export const isCoreUpdateAvailable = ref(false)
export const zashboardVersion = ref(__APP_VERSION__)

export type BackendProbe = {
  uuid: string
  status: 'probing' | 'connected' | 'failed'
  latency: number
  message: string
}

export const backendProbe = ref<BackendProbe | undefined>()
export const startedAt = ref(0)

const detectCore = (versionString: string): Core => {
  if (!versionString) return Core.Singbox
  if (versionString.includes('sing-box')) return Core.Singbox
  if (/\bhonk\b/i.test(versionString)) return Core.Honk
  return Core.Mihomo
}

export const coreBrand = computed(() => {
  switch (core.value) {
    case Core.Singbox:
      return { logo: SingBoxLogo, url: 'https://github.com/sagernet/sing-box' }
    case Core.Honk:
      return { logo: HonkLogo, url: 'https://github.com/Glassyiris/honk' }
    default:
      return {
        logo: MetacubexLogo,
        url: MIHOMO_CHANNEL[mihomo.value?.[0] ?? MIHOMO.Meta].url,
      }
  }
})

export const mihomo = computed<[MIHOMO, string] | undefined>(() => {
  if (core.value !== Core.Mihomo) return undefined

  const match = /(alpha-smart|alpha|beta|meta)-?(\w+)/.exec(version.value)
  switch (match?.[1]) {
    case 'alpha':
      return [MIHOMO.Alpha, match[2] ?? version.value]
    case 'alpha-smart':
      return [MIHOMO.Smart, match[2] ?? version.value]
    case 'meta':
      return [MIHOMO.Meta, match[2] ?? version.value]
    default:
      return [MIHOMO.Meta, version.value]
  }
})

const fetchSingboxVersion = async () => {
  const appService = (window as any).go?.app?.AppService
  if (appService?.GetConnectionInfo) {
    try {
      const info = await appService.GetConnectionInfo()
      apiVersion.value = info.apiVersion || 4
      const v = info.singBoxVersion ? `sing-box ${info.singBoxVersion}` : 'sing-box 1.14'
      return { data: { version: v } }
    } catch {
      // fallback
    }
  }
  apiVersion.value = 4
  return { data: { version: 'sing-box 1.14' } }
}

export const fetchVersionAPI = () => fetchSingboxVersion()

const fetchSingboxStartedAt = async (): Promise<number> => {
  const appService = (window as any).go?.app?.AppService
  if (appService?.GetStartedAt) {
    try {
      const ts = await appService.GetStartedAt()
      return Number(ts || 0)
    } catch {
      return 0
    }
  }
  return 0
}

const probeBackend = async (backend: Backend) => {
  const startAt = Date.now()
  const { data } = await fetchVersionAPI()

  if (activeBackend.value?.uuid !== backend.uuid) return

  version.value = data?.version || 'sing-box 1.14'
  core.value = detectCore(version.value)
  backendProbe.value = {
    uuid: backend.uuid,
    status: 'connected',
    latency: Date.now() - startAt,
    message: '',
  }
  startedAt.value = can('startedAt') ? await fetchSingboxStartedAt() : 0
}

let probe: Promise<void> = Promise.resolve()

export const coreReady = async () => {
  await nextTick()
  await probe
}

export const probeActiveBackend = () => {
  const backend = activeBackend.value
  resetCore()
  version.value = ''
  startedAt.value = 0
  isCoreUpdateAvailable.value = false
  backendProbe.value = backend
    ? { uuid: backend.uuid, status: 'probing', latency: 0, message: '' }
    : undefined

  probe = backend ? probeBackend(backend).catch(() => {}) : Promise.resolve()
  return probe
}

export const fetchIsUIUpdateAvailable = async () => false
export const fetchBackendUpdateAvailableAPI = async () => false
export const isUIUpdateAvailable = ref(false)
export const checkUIUpdate = async () => {}

export const restartCoreAPI = async () => {}
export const upgradeCoreAPI = async (_type?: any) => {}
export const upgradeUIAPI = async () => {}
