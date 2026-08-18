export interface ProcessInfo {
  processId: number
  userId: number
  userName: string
  processPath: string
  processName: string
  packageNames?: string[]
}

export interface Flow {
  id: string
  inbound: string
  inboundType: string
  ipVersion: number
  network: string
  source: string
  destination: string
  domain: string
  protocol: string
  user: string
  fromOutbound: string
  rule: string
  outbound: string
  outboundType: string
  chainList?: string[]
  process?: ProcessInfo
  createdAt: string
  closedAt?: string
  uploadTotal: number
  downloadTotal: number
  uploadDelta: number
  downloadDelta: number
  uploadRate: number
  downloadRate: number
  lastActiveAt: string
  isActive: boolean
}

export interface TimeSeriesPoint {
  timestamp: number
  uploadRate: number
  downloadRate: number
  activeFlows: number
}

export interface NamedAggregate {
  key: string
  name: string
  category?: string
  connectionCount: number
  activeCount: number
  uploadTotal: number
  downloadTotal: number
  totalBytes: number
  uploadRate: number
  downloadRate: number
  totalRate: number
  lastActiveAt?: number
}

export interface ProcessAggregate {
  processName: string
  processPath: string
  processId: number
  connectionCount: number
  activeCount: number
  uploadTotal: number
  downloadTotal: number
  totalBytes: number
  uploadRate: number
  downloadRate: number
  topDomains: NamedAggregate[]
  topDestinations: NamedAggregate[]
}

export interface OverviewSummary {
  uploadRate: number
  downloadRate: number
  sessionUpload: number
  sessionDownload: number
  activeTunFlows: number
  activeTotalFlows: number
  tcpCount: number
  udpCount: number
  topProcess?: NamedAggregate
  topDomain?: NamedAggregate
  topDestination?: NamedAggregate
  topOutbound?: NamedAggregate
  timeSeries: TimeSeriesPoint[]
}

export type ConnectionState =
  | 'Disconnected'
  | 'Connecting'
  | 'Connected'
  | 'Reconnecting'
  | 'Authentication failed'
  | 'API incompatible'
  | 'Error'

export interface ServerConnectionInfo {
  state: ConnectionState
  serverUrl: string
  singBoxVersion: string
  apiVersion: number
  errorMessage?: string
  connectedAt?: string
  lastEventAt?: string
}

export interface SystemStatus {
  memory: number
  goroutines: number
  connectionsIn: number
  connectionsOut: number
  trafficAvailable: boolean
  uplink: number
  downlink: number
  uplinkTotal: number
  downlinkTotal: number
  timestamp: string
}

export interface LogMessage {
  level: string
  message: string
  timestamp: string
}

export interface GroupItem {
  tag: string
  type: string
  urlTestTime: number
  urlTestDelay: number
}

export interface OutboundGroup {
  tag: string
  type: string
  selectable: boolean
  selected: string
  isExpand: boolean
  items: GroupItem[]
}

export interface OutboundInfo {
  tag: string
  type: string
  urlTestTime: number
  urlTestDelay: number
}

export interface RuleInfo {
  type: string
  payload: string
  proxy: string
  hitCount: number
  totalBytes: number
  lastHitAt: number
  uuid: string
  index: number
}

export interface BatchAnalysisResult {
  totalFlows: number
  activeFlows: number
  totalUploadBytes: number
  totalDownloadBytes: number
  totalUploadRate: number
  totalDownloadRate: number
  byProcess: ProcessAggregate[]
  byDomain: NamedAggregate[]
  byDestination: NamedAggregate[]
  byOutbound: NamedAggregate[]
  byRule: NamedAggregate[]
  byProtocol: NamedAggregate[]
  computeTimeUs: number
  engine: string
}
