import { MIN_PROXY_CARD_WIDTH, PROXY_CARD_SIZE } from '@/constant'
import type { Backend, BackendType } from '@/types'
import { useMediaQuery } from '@vueuse/core'
import dayjs from 'dayjs'
import prettyBytes, { type Options } from 'pretty-bytes'
import { computed } from 'vue'

export const isPreferredDark = useMediaQuery('(prefers-color-scheme: dark)')
export const isMiddleScreen = computed(() => false)
export const isPWA = (() => {
  return window.matchMedia('(display-mode: standalone)').matches || (navigator as any).standalone
})()

export const prettyBytesHelper = (bytes: number, opts?: Options) => {
  return prettyBytes(Number.isFinite(bytes) ? bytes : 0, {
    binary: false,
    ...opts,
  })
}

export const fromNow = (timestamp: string | number) => {
  return dayjs(timestamp).fromNow()
}

export const getDashboardSettingsFromStorage = () => {
  const settings: Record<string, string> = {}

  for (const key in localStorage) {
    if (key.startsWith('config/')) {
      settings[key] = localStorage.getItem(key) as string
    }
  }

  return settings
}

export const applyDashboardSettingsToStorage = (settings: Record<string, unknown>) => {
  for (const key in settings) {
    if (key.startsWith('config/')) {
      localStorage.setItem(key, settings[key] as string)
    }
  }
}

export const exportSettings = () => {
  const settings = getDashboardSettingsFromStorage()
  const blob = new Blob([JSON.stringify(settings, null, 2)], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')

  a.href = url
  a.download = `zashboard-settings-${dayjs().format('YYYY-MM-DD-HH-mm-ss')}.json`
  a.click()
  URL.revokeObjectURL(url)
}

export const resetSettings = () => {
  const keysToReset: string[] = []

  for (const key in localStorage) {
    if (key.startsWith('config/')) {
      keysToReset.push(key)
    }
  }

  keysToReset.forEach((key) => localStorage.removeItem(key))
  window.location.reload()
}

export const getUrlFromBackend = (end: {
  protocol: string
  host: string
  port: string
  secondaryPath?: string
}) => {
  return `${end.protocol}://${end.host}:${end.port}${end.secondaryPath || ''}`
}

export const getSingboxUrlFromBackend = (
  end: Pick<Backend, 'type' | 'protocol' | 'host' | 'port'>,
) => {
  if (!end || !end.host) return ''
  return `${end.protocol || 'http'}://${end.host}:${end.port || '9090'}`
}

export const getSingboxSecret = (end: Pick<Backend, 'type' | 'password'>) =>
  end ? end.password || '' : ''

export const getLabelFromBackend = (end: Omit<Backend, 'uuid'>) => {
  if (!end) return ''
  return end.label || `${end.host}:${end.port}`
}

export const getMinCardWidth = (size: PROXY_CARD_SIZE) => {
  return size === PROXY_CARD_SIZE.LARGE ? MIN_PROXY_CARD_WIDTH.LARGE : MIN_PROXY_CARD_WIDTH.SMALL
}

export const PROXIES_PARENT_CLASS = 'proxies-scrollable-parent'

export const findScrollableParent = (el: HTMLElement | null): HTMLElement | null => {
  const parent = el?.parentElement

  if (
    parent?.classList.contains(PROXIES_PARENT_CLASS) &&
    parent.scrollHeight > parent.clientHeight
  ) {
    return parent
  }

  return parent ? findScrollableParent(parent) : null
}

export const scrollIntoCenter = (el: HTMLElement) => {
  const scrollableParent = findScrollableParent(el)

  if (!scrollableParent) return

  const parentTop = scrollableParent.offsetTop
  const childTop = el.offsetTop

  const relativeTop = childTop - parentTop - scrollableParent.scrollTop

  if (relativeTop >= 0 && relativeTop + el.clientHeight <= scrollableParent.clientHeight) return

  const centerOffset =
    childTop - parentTop - scrollableParent.clientHeight / 2 + el.clientHeight / 2

  scrollableParent.scrollTo({
    top: centerOffset,
    behavior: 'smooth',
  })
}

export const getBackendFromUrl = () => {
  const query = new URLSearchParams(
    window.location.search || location.hash.match(/\?.*$/)?.[0]?.replace('?', ''),
  )

  if (query.has('hostname')) {
    return {
      type: (query.get('type') === 'singbox' ? 'singbox' : 'clash') as BackendType,
      protocol: query.get('http')
        ? 'http'
        : query.get('https')
          ? 'https'
          : window.location.protocol.replace(':', ''),
      secondaryPath: query.get('secondaryPath') || '',
      host: query.get('hostname') as string,
      port: query.get('port') as string,
      password: query.get('secret') || '',
      label: query.get('label') || '',
      disableUpgradeCore:
        query.get('disableUpgradeCore') === '1' || query.get('disableUpgradeCore') === 'core',
      disableTunMode: query.get('disableTunMode') === '1' || query.get('disableTunMode') === 'tun',
    }
  }
  return null
}
