import type {
  OverviewSummary,
  ServerConnectionInfo,
  SystemStatus,
  LogMessage,
  OutboundGroup,
  BatchAnalysisResult,
  RuleInfo,
  Flow,
} from '@/types/domain'

export interface QueryOptions {
  search?: string
  process?: string
  inbound?: string
  outbound?: string
  protocol?: string
  network?: string
  activeOnly?: boolean
  tunOnly?: boolean
  limit?: number
  offset?: number
  sortBy?: string
  sortDesc?: boolean
}

async function apiPost<T>(endpoint: string, body?: unknown): Promise<T | null> {
  try {
    const res = await fetch(`/api/${endpoint}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: body !== undefined ? JSON.stringify(body) : undefined,
    })
    if (!res.ok) return null
    return (await res.json()) as T
  } catch {
    return null
  }
}

async function apiGet<T>(endpoint: string): Promise<T | null> {
  try {
    const res = await fetch(`/api/${endpoint}`, {
      method: 'GET',
    })
    if (!res.ok) return null
    return (await res.json()) as T
  } catch {
    return null
  }
}

export const Backend = {
  async getConnectionInfo(): Promise<ServerConnectionInfo> {
    const res = await apiGet<ServerConnectionInfo>('connection')
    return res || {
      state: 'Connected',
      serverUrl: 'http://127.0.0.1:9090',
      singBoxVersion: 'sing-box 1.14',
      apiVersion: 4,
    }
  },

  async connect(url: string, secret: string): Promise<boolean> {
    const res = await apiPost<{ ok: boolean }>('connect', { url, secret })
    return res?.ok ?? true
  },

  async disconnect(): Promise<boolean> {
    const res = await apiPost<{ ok: boolean }>('disconnect')
    return res?.ok ?? true
  },

  async getOverviewSummary(inboundFilter = ''): Promise<OverviewSummary> {
    const res = await apiGet<OverviewSummary>(`overview?filter=${encodeURIComponent(inboundFilter)}`)
    return res || {
      uploadRate: 0,
      downloadRate: 0,
      sessionUpload: 0,
      sessionDownload: 0,
      activeTunFlows: 0,
      activeTotalFlows: 0,
      tcpCount: 0,
      udpCount: 0,
      timeSeries: [],
    }
  },

  async getFlows(opts: QueryOptions): Promise<{ flows: Flow[]; totalCount: number }> {
    const res = await apiPost<{ flows: Flow[]; totalCount: number }>('flows', opts)
    return res || { flows: [], totalCount: 0 }
  },

  async getBatchAnalytics(inboundFilter = '', topN = 100): Promise<BatchAnalysisResult> {
    const res = await apiGet<BatchAnalysisResult>(`analytics?filter=${encodeURIComponent(inboundFilter)}&topN=${topN}`)
    return res || {
      totalFlows: 0,
      activeFlows: 0,
      totalUploadBytes: 0,
      totalDownloadBytes: 0,
      totalUploadRate: 0,
      totalDownloadRate: 0,
      byProcess: [],
      byDomain: [],
      byDestination: [],
      byOutbound: [],
      byRule: [],
      byProtocol: [],
      computeTimeUs: 0,
      engine: 'pure-go',
    }
  },

  async getRules(): Promise<RuleInfo[]> {
    const res = await apiGet<RuleInfo[]>('rules')
    return res || []
  },

  async getSystemStatus(): Promise<SystemStatus> {
    const res = await apiGet<SystemStatus>('status')
    return res || {
      memory: 0,
      goroutines: 0,
      connectionsIn: 0,
      connectionsOut: 0,
      trafficAvailable: true,
      uplink: 0,
      downlink: 0,
      uplinkTotal: 0,
      downlinkTotal: 0,
      timestamp: new Date().toISOString(),
    }
  },

  async getLogs(limit = 100): Promise<LogMessage[]> {
    const res = await apiGet<LogMessage[]>(`logs?limit=${limit}`)
    return res || []
  },

  async clearLogs(): Promise<boolean> {
    const res = await apiPost<{ ok: boolean }>('clear-logs')
    return res ? res.ok : true
  },

  async getGroups(): Promise<OutboundGroup[]> {
    const res = await apiGet<OutboundGroup[]>('groups')
    return res || []
  },

  async closeConnection(id: string): Promise<boolean> {
    const res = await apiPost<{ ok: boolean }>(`close?id=${encodeURIComponent(id)}`)
    return res ? res.ok : true
  },

  async closeAllConnections(): Promise<boolean> {
    const res = await apiPost<{ ok: boolean }>('close-all')
    return res ? res.ok : true
  },

  async selectOutbound(groupTag: string, outboundTag: string): Promise<boolean> {
    const res = await apiPost<{ ok: boolean }>('select-outbound', { groupTag, outboundTag })
    return res ? res.ok : true
  },

  async urlTest(outboundTag: string): Promise<boolean> {
    const res = await apiPost<{ ok: boolean }>(`url-test?tag=${encodeURIComponent(outboundTag)}`)
    return res ? res.ok : true
  },

  async getClashModeStatus(): Promise<{ currentMode: string; modeList: string[] }> {
    const res = await apiGet<{ currentMode: string; modeList: string[] }>('clash-mode')
    return res || { currentMode: 'rule', modeList: ['rule', 'global', 'direct'] }
  },

  async setClashMode(mode: string): Promise<boolean> {
    const res = await apiPost<{ ok: boolean }>('set-clash-mode', { mode })
    return res ? res.ok : true
  },

  async getStartedAt(): Promise<number> {
    const res = await apiGet<{ startedAt: number }>('started-at')
    return res?.startedAt || 0
  },
}
