import { ref } from 'vue'

export const smartWeightsMap = ref<Record<string, Record<string, string>>>({})
export const smartOrderMap = ref<Record<string, Record<string, number>>>({})

export const initSmartWeights = async (_smartGroups: string[]) => {
  smartWeightsMap.value = {}
  smartOrderMap.value = {}
}
