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

async function callGoMethod<T>(method: string, ...args: unknown[]): Promise<T | null> {
  const win = window as any

  if (win.go?.app?.AppService?.[method]) {
    try {
      return (await win.go.app.AppService[method](...args)) as T
    } catch (e) {
      console.warn(`[Wails IPC] AppService.${method} failed:`, e)
    }
  }

  if (win.wails?.services?.AppService?.[method]) {
    try {
      return (await win.wails.services.AppService[method](...args)) as T
    } catch (e) {
      console.warn(`[Wails v3] AppService.${method} failed:`, e)
    }
  }

  if (typeof win.wails?.Call === 'function') {
    try {
      return (await win.wails.Call({
        methodName: `sing-scope/internal/app.AppService.${method}`,
        args,
      })) as T
    } catch (e) {
      console.warn(`[Wails v3 Call] Failed:`, e)
    }
  }

  return null
}

export const Backend = {
  async getConnectionInfo(): Promise<ServerConnectionInfo> {
    const res = await callGoMethod<ServerConnectionInfo>('GetConnectionInfo')
    return res || {
      state: 'Connected',
      serverUrl: 'http://127.0.0.1:9090',
      singBoxVersion: 'sing-box 1.14',
      apiVersion: 4,
    }
  },

  async connect(url: string, secret: string): Promise<boolean> {
    const res = await callGoMethod<boolean>('ConnectServer', url, secret)
    return res ?? true
  },

  async disconnect(): Promise<boolean> {
    const res = await callGoMethod<boolean>('DisconnectServer')
    return res ?? true
  },

  async getOverviewSummary(inboundFilter = ''): Promise<OverviewSummary> {
    const res = await callGoMethod<OverviewSummary>('GetOverviewSummary', inboundFilter)
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
    const res = await callGoMethod<{ flows: Flow[]; totalCount: number }>('GetFlows', opts)
    return res || { flows: [], totalCount: 0 }
  },

  async getBatchAnalytics(inboundFilter = '', topN = 100): Promise<BatchAnalysisResult> {
    const res = await callGoMethod<BatchAnalysisResult>('GetBatchAnalytics', inboundFilter, topN)
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
    const res = await callGoMethod<RuleInfo[]>('GetRules')
    return res || []
  },

  async getSystemStatus(): Promise<SystemStatus> {
    const res = await callGoMethod<SystemStatus>('GetSystemStatus')
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
    const res = await callGoMethod<LogMessage[]>('GetLogs', limit)
    return res || []
  },

  async clearLogs(): Promise<boolean> {
    const res = await callGoMethod<boolean>('ClearLogs')
    return res ?? true
  },

  async getGroups(): Promise<OutboundGroup[]> {
    const res = await callGoMethod<OutboundGroup[]>('GetGroups')
    return res || []
  },

  async closeConnection(id: string): Promise<boolean> {
    const res = await callGoMethod<boolean>('CloseConnection', id)
    return res ?? true
  },

  async closeAllConnections(): Promise<boolean> {
    const res = await callGoMethod<boolean>('CloseAllConnections')
    return res ?? true
  },

  async selectOutbound(groupTag: string, outboundTag: string): Promise<boolean> {
    const res = await callGoMethod<boolean>('SelectOutbound', groupTag, outboundTag)
    return res ?? true
  },

  async urlTest(outboundTag: string): Promise<boolean> {
    const res = await callGoMethod<boolean>('URLTest', outboundTag)
    return res ?? true
  },

  async getClashModeStatus(): Promise<{ currentMode: string; modeList: string[] }> {
    const res = await callGoMethod<{ currentMode: string; modeList: string[] }>('GetClashModeStatus')
    return res || { currentMode: 'rule', modeList: ['rule', 'global', 'direct'] }
  },

  async setClashMode(mode: string): Promise<boolean> {
    const res = await callGoMethod<boolean>('SetClashMode', mode)
    return res ?? true
  },

  async getStartedAt(): Promise<number> {
    const res = await callGoMethod<number>('GetStartedAt')
    return res || 0
  },
}
