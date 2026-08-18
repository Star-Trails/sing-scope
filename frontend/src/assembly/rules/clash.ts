import { ruleProviderList, rules } from './index'

export const fetchRules = async () => {
  rules.value = []
  ruleProviderList.value = []
}
