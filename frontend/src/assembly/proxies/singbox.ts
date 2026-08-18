// sing-box API(gRPC daemon.StartedService)后端的代理「组装逻辑」。
// 优先通过 Wails v3 Go Backend 获取,兜底使用 gRPC-Web 流。
import { getSingboxClient } from '@/api/singbox/client'
import type { StreamHandle } from '@/api/singbox/streams'
import { subscribeStream } from '@/api/singbox/subscriptions'
import { disconnectByIdAPI } from '@/assembly/connections'
import type { Groups, OutboundList } from '@/gen/daemon/started_service_pb'
import { getConnectionChains } from '@/helper'
import { activeConnections } from '@/store/connections'
import { automaticDisconnection, iconReflectList, speedtestTimeout } from '@/store/settings'
import { activeBackend } from '@/store/setup'
import type { Proxy } from '@/types'
import { proxyGroupList, proxyMap, proxyProviederList } from './index'

const getHistoryFromItem = (item: any): Proxy['history'] =>
  item.urlTestDelay > 0
    ? [
        {
          time: new Date(Number(item.urlTestTime || Date.now() / 1000) * 1000).toISOString(),
          delay: item.urlTestDelay,
        },
      ]
    : []

const nodeToProxy = (item: any): Proxy => {
  return {
    name: item.tag,
    type: item.type,
    now: '',
    history: getHistoryFromItem(item),
    extra: {},
    icon: '',
  }
}

let groups = new Map<string, any>()
let outbounds = new Map<string, any>()
let handles: StreamHandle[] = []
let sessionKey = ''
let ready: Promise<void> | null = null

type URLTestWaiter = {
  resolve: () => void
  reject: (reason: Error) => void
  timer: number
}

const urlTestWaiters = new Set<URLTestWaiter>()

const resolveURLTestWaiters = () => {
  for (const waiter of urlTestWaiters) {
    clearTimeout(waiter.timer)
    waiter.resolve()
  }
  urlTestWaiters.clear()
}

const rejectURLTestWaiters = (reason: Error) => {
  for (const waiter of urlTestWaiters) {
    clearTimeout(waiter.timer)
    waiter.reject(reason)
  }
  urlTestWaiters.clear()
}

const waitForURLTestResult = (timeout: number) => {
  let waiter!: URLTestWaiter
  const promise = new Promise<void>((resolve, reject) => {
    const timer = window.setTimeout(
      () => {
        urlTestWaiters.delete(waiter)
        reject(new Error('sing-box URL test result timeout'))
      },
      Math.max(5000, timeout) + 1000,
    )

    waiter = { resolve, reject, timer }
    urlTestWaiters.add(waiter)
  })

  return {
    promise,
    cancel: () => {
      clearTimeout(waiter.timer)
      urlTestWaiters.delete(waiter)
    },
  }
}

const rebuild = () => {
  const proxies: Record<string, Proxy> = {}

  for (const item of outbounds.values()) {
    proxies[item.tag] = nodeToProxy(item)
  }
  for (const group of groups.values()) {
    for (const item of group.items || []) {
      if (!proxies[item.tag]) proxies[item.tag] = nodeToProxy(item)
    }
  }
  for (const group of groups.values()) {
    proxies[group.tag] = {
      name: group.tag,
      type: group.type,
      now: group.selected,
      all: (group.items || []).map((i: any) => i.tag),
      selectable: group.selectable,
      history: [],
      extra: {},
      icon: '',
    }
  }
  for (const group of groups.values()) {
    for (const item of group.items || []) {
      const node = proxies[item.tag]
      if (node && !node.all?.length && item.urlTestDelay > 0) {
        node.history = getHistoryFromItem(item)
      }
    }
  }
  for (const iconReflect of iconReflectList.value) {
    const node = proxies[iconReflect.name]
    if (node) node.icon = iconReflect.icon
  }

  proxyMap.value = proxies
  proxyGroupList.value = Array.from(groups.values())
    .filter((g) => g.items && g.items.length)
    .map((g) => g.tag)
  proxyProviederList.value = []
}

const closeStreams = () => {
  handles.forEach((h) => h.close())
  handles = []
  rejectURLTestWaiters(new Error('sing-box proxy stream closed'))
  sessionKey = ''
  ready = null
}

const stop = () => {
  closeStreams()
  groups = new Map()
  outbounds = new Map()
}

const ensureSession = () => {
  const backend = activeBackend.value
  const client = getSingboxClient()?.client
  if (!backend || backend.type !== 'singbox' || !client) {
    stop()
    return
  }
  if (sessionKey === backend.uuid && handles.length) return

  stop()
  sessionKey = backend.uuid

  let resolveReady!: () => void
  let resolved = false
  ready = new Promise<void>((r) => (resolveReady = r))

  handles = [
    subscribeStream<Groups>('groups', (msg) => {
      groups = new Map()
      for (const g of msg.group) groups.set(g.tag, g)
      rebuild()
      if (!resolved) {
        resolved = true
        resolveReady()
      } else {
        resolveURLTestWaiters()
      }
    }),
    subscribeStream<OutboundList>('outbounds', (msg) => {
      outbounds = new Map()
      for (const o of msg.outbounds) outbounds.set(o.tag, o)
      rebuild()
    }),
  ]
}

export const resetProxies = () => stop()

export const fetchProxies = async () => {
  const appService = (window as any).go?.app?.AppService
  if (appService?.GetGroups) {
    try {
      const gList = await appService.GetGroups()
      groups = new Map()
      for (const g of gList || []) {
        groups.set(g.tag, g)
      }
      rebuild()
      return
    } catch {
      // fallback
    }
  }
  ensureSession()
  if (ready) await ready
  rebuild()
}

export const handlerProxySelect = async (proxyGroupName: string, proxyName: string) => {
  const appService = (window as any).go?.app?.AppService
  if (appService?.SelectOutbound) {
    await appService.SelectOutbound(proxyGroupName, proxyName)
    const group = groups.get(proxyGroupName)
    if (group) {
      group.selected = proxyName
      rebuild()
    }
    return
  }

  const client = getSingboxClient()?.client
  const proxyGroup = proxyMap.value[proxyGroupName]
  if (!client || proxyGroup?.selectable === false) return

  await client.selectOutbound({ groupTag: proxyGroupName, outboundTag: proxyName })

  const group = groups.get(proxyGroupName)
  if (group) {
    group.selected = proxyName
    rebuild()
  }

  if (automaticDisconnection.value) {
    activeConnections.value
      .filter((c) => getConnectionChains(c).includes(proxyGroupName))
      .forEach((c) => disconnectByIdAPI(c.id).catch(() => {}))
  }
}

const runURLTest = async (outboundTag: string, timeout = speedtestTimeout.value) => {
  const appService = (window as any).go?.app?.AppService
  if (appService?.URLTest) {
    await appService.URLTest(outboundTag)
    await fetchProxies()
    return
  }

  ensureSession()
  if (ready) await ready

  const client = getSingboxClient()?.client
  if (!client) return

  const result = waitForURLTestResult(timeout)
  try {
    await Promise.all([client.uRLTest({ outboundTag }), result.promise])
  } finally {
    result.cancel()
  }
}

export const proxyLatencyTest = async (
  proxyName: string,
  _url?: string,
  timeout = speedtestTimeout.value,
) => {
  await runURLTest(proxyName, timeout)
}

export const proxyGroupLatencyTest = async (proxyGroupName: string) => {
  await runURLTest(proxyGroupName)
}

export const allProxiesLatencyTest = async () => {
  await Promise.allSettled(Array.from(groups.keys()).map((tag) => runURLTest(tag)))
}
