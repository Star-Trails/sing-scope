// sing-box 后端的连接组装:优先通过 Wails v3 Go Backend (AppService) 获取实时连接,
// 兜底回退到 gRPC SubscribeConnections 订阅。
import { getSingboxClient } from '@/api/singbox/client'
import { subscribeStream } from '@/api/singbox/subscriptions'
import {
  ConnectionEventType,
  type ConnectionEvents,
  type Connection as PbConnection,
} from '@/gen/daemon/started_service_pb'
import type { Connection } from '@/types'
import { shallowRef, type Ref } from 'vue'
import {
  createGetConnectionDisplayValue,
  createGetConnectionVisibleSearchValues,
  type ConnectionAccessor,
  type ConnectionsSnapshot,
} from './accessor'

const fetchSingboxConnections = (): {
  data: Ref<ConnectionsSnapshot | undefined>
  close: () => void
} => {
  const data = shallowRef<ConnectionsSnapshot>()

  const appService = (window as any).go?.app?.AppService
  if (appService?.GetFlows) {
    let pollTimer: number | null = null
    const poll = async () => {
      try {
        const res = await appService.GetFlows({ activeOnly: false, limit: 1000 })
        const active: Connection[] = []
        const closed: Connection[] = []

        for (const f of res.flows || []) {
          const c: any = {
            id: f.id,
            inbound: f.inbound,
            inboundType: f.inboundType,
            ipVersion: f.ipVersion,
            network: f.network,
            source: f.source,
            destination: f.destination,
            domain: f.domain,
            protocol: f.protocol,
            user: f.user,
            fromOutbound: f.fromOutbound,
            rule: f.rule,
            outbound: f.outbound,
            outboundType: f.outboundType,
            chainList: f.chainList || [f.outbound].filter(Boolean),
            processInfo: f.process
              ? {
                  processId: f.process.processId,
                  userId: f.process.userId,
                  userName: f.process.userName,
                  processPath: f.process.processPath,
                }
              : undefined,
            createdAt: BigInt(new Date(f.createdAt || Date.now()).getTime()),
            closedAt: f.closedAt ? BigInt(new Date(f.closedAt).getTime()) : 0n,
            uplinkTotal: BigInt(f.uploadTotal || 0),
            downlinkTotal: BigInt(f.downloadTotal || 0),
            uploadSpeed: f.uploadRate || 0,
            downloadSpeed: f.downloadRate || 0,
          }

          if (f.isActive) {
            active.push(c as Connection)
          } else {
            closed.push(c as Connection)
          }
        }

        data.value = { active, closed }
      } catch {
        // ignore
      }
    }

    poll()
    pollTimer = window.setInterval(poll, 1000)

    return {
      data,
      close: () => {
        if (pollTimer) clearInterval(pollTimer)
      },
    }
  }

  // 活跃连接表
  const conns = new Map<string, Connection>()
  let newlyClosed: Connection[] = []
  let timer: number | null = null

  const enrich = (c: PbConnection | Connection, down: number, up: number): Connection =>
    Object.assign({}, c, { downloadSpeed: down, uploadSpeed: up }) as Connection

  const close = (id: string, base?: PbConnection | Connection) => {
    const c = base ?? conns.get(id)
    conns.delete(id)
    if (c) newlyClosed.push(enrich(c, 0, 0))
  }

  const emit = () => {
    timer = null
    data.value = {
      active: Array.from(conns.values()),
      closed: newlyClosed,
    }
    newlyClosed = []
  }
  const scheduleEmit = () => {
    if (timer) return
    timer = window.setTimeout(emit, 100)
  }

  const handle = subscribeStream<ConnectionEvents>('connections', (msg) => {
    if (msg.reset) {
      conns.clear()
    }
    for (const event of msg.events) {
      const downDelta = Number(event.downlinkDelta)
      const upDelta = Number(event.uplinkDelta)

      switch (event.type) {
        case ConnectionEventType.CONNECTION_EVENT_NEW:
          if (event.connection) {
            if (event.connection.closedAt > 0n) close(event.id, event.connection)
            else conns.set(event.id, enrich(event.connection, 0, 0))
          }
          break
        case ConnectionEventType.CONNECTION_EVENT_UPDATE: {
          if (event.connection) {
            if (event.connection.closedAt > 0n) close(event.id, event.connection)
            else conns.set(event.id, enrich(event.connection, downDelta, upDelta))
          } else {
            const prev = conns.get(event.id)
            if (prev) {
              const s = asSingbox(prev)
              conns.set(
                event.id,
                enrich(
                  {
                    ...s,
                    uplinkTotal: s.uplinkTotal + event.uplinkDelta,
                    downlinkTotal: s.downlinkTotal + event.downlinkDelta,
                  },
                  downDelta,
                  upDelta,
                ),
              )
            }
          }
          break
        }
        case ConnectionEventType.CONNECTION_EVENT_CLOSED: {
          close(event.id, event.connection)
          break
        }
      }
    }
    scheduleEmit()
  })

  return {
    data,
    close: () => {
      if (timer) clearTimeout(timer)
      handle.close()
    },
  }
}

const closeSingboxConnection = async (id: string) => {
  const appService = (window as any).go?.app?.AppService
  if (appService?.CloseConnection) {
    await appService.CloseConnection(id)
    return
  }
  const client = getSingboxClient()?.client
  if (!client) return
  await client.closeConnection({ id })
}

const closeAllSingboxConnections = async () => {
  const appService = (window as any).go?.app?.AppService
  if (appService?.CloseAllConnections) {
    await appService.CloseAllConnections()
    return
  }
  const client = getSingboxClient()?.client
  if (!client) return
  await client.closeAllConnections({})
}

export const disconnectByIdAPI = closeSingboxConnection

export const disconnectAllAPI = closeAllSingboxConnections

export const fetchConnectionsAPI = fetchSingboxConnections

// 拆分 "ip:port" / "[ipv6]:port"
const splitHostPort = (value: string): [string, string] => {
  if (!value) return ['', '']
  const idx = value.lastIndexOf(':')
  if (idx === -1) return [value, '']

  let host = value.slice(0, idx)
  const port = value.slice(idx + 1)

  if (host.startsWith('[') && host.endsWith(']')) {
    host = host.slice(1, -1)
  }

  return [host, port]
}

const asSingbox = (connection: Connection) => connection as PbConnection

const getNetwork = (c: PbConnection) => {
  const [, destinationPort] = splitHostPort(c.destination)

  if ((destinationPort === '443' || c.domain) && c.network === 'udp') {
    return 'quic'
  }

  return c.network
}

const getHostname = (c: PbConnection) => c.domain || splitHostPort(c.destination)[0]

export const connectionAccessor: ConnectionAccessor = {
  chains: (connection) => {
    const c = asSingbox(connection)

    return c.chainList && c.chainList.length ? c.chainList : [c.outbound].filter(Boolean)
  },
  download: (connection) => Number(asSingbox(connection).downlinkTotal),
  upload: (connection) => Number(asSingbox(connection).uplinkTotal),
  start: (connection) => Number(asSingbox(connection).createdAt),
  rule: (connection) => asSingbox(connection).rule,
  rulePayload: () => '',
  sourceIP: (connection) => {
    const c = asSingbox(connection)
    const proc = c.processInfo?.processPath ? c.processInfo.processPath.replace(/^.*[/\\](.*)$/, '$1') : ''
    if (proc) return proc
    if (c.user) return c.user
    return splitHostPort(c.source)[0] || c.inbound || '-'
  },
  sourcePort: (connection) => splitHostPort(asSingbox(connection).source)[1],
  network: (connection) => getNetwork(asSingbox(connection)),
  networkType: (connection) => {
    const c = asSingbox(connection)

    return `${c.inboundType} | ${getNetwork(c)}`
  },
  hostname: (connection) => getHostname(asSingbox(connection)),
  host: (connection) => {
    const c = asSingbox(connection)
    const [, destinationPort] = splitHostPort(c.destination)
    const host = getHostname(c)

    if (host.includes(':')) {
      return `[${host}]:${destinationPort}`
    }
    return `${host}:${destinationPort}`
  },
  process: (connection) => {
    const processPath = asSingbox(connection).processInfo?.processPath ?? ''

    return processPath.replace(/^.*[/\\](.*)$/, '$1') || '-'
  },
  destination: (connection) => {
    const c = asSingbox(connection)

    return splitHostPort(c.destination)[0] || c.domain
  },
  inboundUser: (connection) => {
    const c = asSingbox(connection)

    return c.user || c.inbound || '-'
  },
  sniffHost: (connection) => asSingbox(connection).domain,
  remoteAddress: (connection) => asSingbox(connection).destination,
  protocol: (connection) => asSingbox(connection).protocol,
  outboundType: (connection) => asSingbox(connection).outboundType,
  fromOutbound: (connection) => asSingbox(connection).fromOutbound,
  smartBlock: () => undefined,
}

export const getConnectionDisplayValue = createGetConnectionDisplayValue(connectionAccessor)

export const getConnectionVisibleSearchValues =
  createGetConnectionVisibleSearchValues(connectionAccessor)
