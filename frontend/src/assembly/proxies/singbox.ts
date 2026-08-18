import { disconnectByIdAPI } from '@/assembly/connections'
import { getConnectionChains } from '@/helper'
import { activeConnections } from '@/store/connections'
import { automaticDisconnection, iconReflectList } from '@/store/settings'
import type { Proxy } from '@/types'
import { Backend } from '@/utils/backend'
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

export const resetProxies = () => {
  groups = new Map()
  outbounds = new Map()
}

export const fetchProxies = async () => {
  try {
    const gList = await Backend.getGroups()
    groups = new Map()
    for (const g of gList || []) {
      groups.set(g.tag, g)
    }
  } catch {
    // ignore
  }
  rebuild()
}

export const handlerProxySelect = async (proxyGroupName: string, proxyName: string) => {
  await Backend.selectOutbound(proxyGroupName, proxyName)
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

const runURLTest = async (outboundTag: string) => {
  await Backend.urlTest(outboundTag)
  await fetchProxies()
}

export const proxyLatencyTest = async (
  proxyName: string,
  _url?: string,
  _timeout?: number,
  _groupName?: string,
) => {
  await runURLTest(proxyName)
}

export const proxyGroupLatencyTest = async (proxyGroupName: string) => {
  await runURLTest(proxyGroupName)
}

export const allProxiesLatencyTest = async () => {
  await Promise.allSettled(Array.from(groups.keys()).map((tag) => runURLTest(tag)))
}
