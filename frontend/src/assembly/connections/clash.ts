import { shallowRef } from 'vue'
import {
  createGetConnectionDisplayValue,
  createGetConnectionVisibleSearchValues,
  type ConnectionAccessor,
  type ConnectionsSnapshot,
} from './accessor'

export const disconnectByIdAPI = async (_id: string) => {}
export const disconnectAllAPI = async () => {}
export const fetchConnectionsAPI = () => ({
  data: shallowRef<ConnectionsSnapshot>({ active: [], closed: [] }),
  close: () => {},
})

export const connectionAccessor: ConnectionAccessor = {
  chains: () => [],
  download: () => 0,
  upload: () => 0,
  start: () => 0,
  rule: () => '',
  rulePayload: () => '',
  sourceIP: () => '',
  sourcePort: () => '',
  network: () => '',
  networkType: () => '',
  hostname: () => '',
  host: () => '',
  process: () => '',
  destination: () => '',
  inboundUser: () => '',
  sniffHost: () => '',
  remoteAddress: () => '',
  protocol: () => '',
  outboundType: () => '',
  fromOutbound: () => '',
  smartBlock: () => undefined,
}

export const getConnectionDisplayValue = createGetConnectionDisplayValue(connectionAccessor)
export const getConnectionVisibleSearchValues = createGetConnectionVisibleSearchValues(connectionAccessor)
