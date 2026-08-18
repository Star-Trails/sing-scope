// 组装层 · proxies 门面
import { NOT_CONNECTED, PROXY_TAB_TYPE, PROXY_TYPE, TEST_URL } from '@/constant'
import { notifyRequestError } from '@/helper/requestError'
import { groupTestUrls, independentLatencyTest, speedtestUrl } from '@/store/settings'
import type { Proxy, ProxyProvider } from '@/types'
import { useStorage } from '@vueuse/core'
import { last } from 'lodash'
import { computed, ref } from 'vue'
import * as singbox from './singbox'

export const proxiesFilter = ref('')
export const proxiesTabShow = ref(PROXY_TAB_TYPE.PROXIES)

export const proxyGroupList = ref<string[]>([])
export const proxyMap = ref<Record<string, Proxy>>({})
export const IPv6Map = useStorage<Record<string, boolean>>('cache/ipv6-map', {})
export const hiddenGroupMap = useStorage<Record<string, boolean>>('config/hidden-group-map', {})
export const proxyProviederList = ref<ProxyProvider[]>([])

export const speedtestUrlWithDefault = computed(() => {
  return speedtestUrl.value || TEST_URL
})

export const getTestUrl = (groupName?: string) => {
  if (!groupName || !independentLatencyTest.value) {
    return speedtestUrlWithDefault.value
  }

  const groupTestUrl = groupTestUrls.value.find((item) => item.name === groupName)

  if (groupTestUrl) {
    return groupTestUrl.url
  }

  const proxyNode =
    proxyMap.value[groupName] || proxyProviederList.value.find((p) => p.name === groupName)

  return proxyNode?.testUrl || speedtestUrlWithDefault.value
}

export const getLatencyFromHistory = (history: Proxy['history']) => {
  return last(history)?.delay ?? NOT_CONNECTED
}

export const getLatencyByName = (proxyName: string, groupName?: string) => {
  const history = getHistoryByName(proxyName, groupName)

  return getLatencyFromHistory(history)
}

export const getHistoryByName = (proxyName: string, _groupName?: string) => {
  const nowNode = proxyMap.value[getNowProxyNodeName(proxyName)]

  return nowNode?.history || []
}

export const getIPv6ByName = (proxyName: string) => {
  return IPv6Map.value[getNowProxyNodeName(proxyName)]
}

export const getNowProxyNodeName = (name: string) => {
  let node = proxyMap.value[name]

  if (!name || !node) {
    return name
  }

  while (node.now && node.now !== node.name) {
    const nextNode = proxyMap.value[node.now]

    if (!nextNode) {
      return node.name
    }

    node = nextNode
  }

  return node.name
}

export const getProxyGroupChains = (name: string) => {
  let proxyNode = proxyMap.value[name]

  if (!proxyNode) {
    return []
  }

  const result = [name]

  while (
    proxyNode.now &&
    proxyNode.now !== proxyNode.name &&
    proxyGroupList.value.includes(proxyNode.now)
  ) {
    result.push(proxyNode.now)
    proxyNode = proxyMap.value[proxyNode.now]
  }
  return result
}

export const hasSmartGroup = computed(() => {
  return Object.values(proxyMap.value).some(
    (proxy) => proxy.type.toLowerCase() === PROXY_TYPE.Smart,
  )
})

export const fetchProxies = () => singbox.fetchProxies()

export const handlerProxySelect = async (proxyGroupName: string, proxyName: string) => {
  try {
    return await singbox.handlerProxySelect(proxyGroupName, proxyName)
  } catch (e) {
    notifyRequestError(e)
  }
}

export const proxyLatencyTest = (
  proxyName: string,
  url?: string,
  timeout?: number,
  groupName?: string,
) => singbox.proxyLatencyTest(proxyName, url, timeout, groupName)

export const proxyGroupLatencyTest = (proxyGroupName: string) =>
  singbox.proxyGroupLatencyTest(proxyGroupName)

export const allProxiesLatencyTest = () => singbox.allProxiesLatencyTest()

export const resetProxies = () => singbox.resetProxies()

export const fetchSmartGroupWeightsAPI = async (_name?: unknown) => ({ data: {} as Record<string, number> })
export const fetchSmartWeightsAPI = async () => ({ data: {} as Record<string, number> })
export const flushSmartGroupWeightsAPI = async () => {}
export const proxyProviderHealthCheckAPI = async (_name?: unknown) => {}
export const updateProxyProviderAPI = async (_name?: unknown) => {}
