import { PROXY_TAB_TYPE, RULE_TAB_TYPE } from '@/constant'
import { initConnections, stopConnections } from '@/store/connections'
import { initSatistic, stopSatistic } from '@/store/overview'
import { activeBackend } from '@/store/setup'
import { Backend } from '@/utils/backend'
import { watch } from 'vue'
import { fetchConfigs } from './config'
import { initLogs, stopLogs } from './logs'
import { fetchProxies, proxiesTabShow, resetProxies } from './proxies'
import { fetchRules, rulesTabShow } from './rules'
import { probeActiveBackend } from './version'

let generation = 0

export const startBackendSession = async () => {
  const current = ++generation

  const backend = activeBackend.value
  if (backend) {
    const url = `${backend.protocol || 'http'}://${backend.host}:${backend.port || '9090'}`
    await Backend.connect(url, backend.password || '')
  }

  probeActiveBackend()
  stopConnections()
  stopLogs()
  stopSatistic()
  resetProxies()

  if (current !== generation) return
  if (!activeBackend.value) return

  rulesTabShow.value = RULE_TAB_TYPE.RULES
  proxiesTabShow.value = PROXY_TAB_TYPE.PROXIES
  fetchConfigs()
  fetchProxies()
  fetchRules()
  initConnections()
  initLogs()
  initSatistic()
}

watch(activeBackend, startBackendSession, { immediate: true })
