// 组装层 · connections 门面
import { CONNECTIONS_TABLE_ACCESSOR_KEY } from '@/constant'
import type { Connection } from '@/types'
import type { ConnectionDisplayOptions, ConnectionsSnapshot } from './accessor'
import * as singbox from './singbox'

export type { ConnectionsSnapshot }

export const disconnectByIdAPI = (id: string) => singbox.disconnectByIdAPI(id)
export const disconnectAllAPI = () => singbox.disconnectAllAPI()
export const fetchConnectionsAPI = () => singbox.fetchConnectionsAPI()

export const connectionAccessor = () => singbox.connectionAccessor

export const getConnectionDisplayValue = (
  connection: Connection,
  key: CONNECTIONS_TABLE_ACCESSOR_KEY,
  options: ConnectionDisplayOptions,
) => singbox.getConnectionDisplayValue(connection, key, options)

export const getConnectionVisibleSearchValues = (
  connection: Connection,
  keys: CONNECTIONS_TABLE_ACCESSOR_KEY[],
  options: ConnectionDisplayOptions,
) => singbox.getConnectionVisibleSearchValues(connection, keys, options)

export const blockConnectionByIdAPI = async (_id?: unknown) => {}
