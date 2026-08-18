// 组装层 · overview 门面
import * as singbox from './singbox'

export const fetchMemoryAPI = <T>() => singbox.fetchMemoryAPI<T>()
export const fetchTrafficAPI = <T>() => singbox.fetchTrafficAPI<T>()
