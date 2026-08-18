import { ref } from 'vue'

export const fetchMemoryAPI = <T>() => ({ data: ref<T>(), close: () => {} })
export const fetchTrafficAPI = <T>() => ({ data: ref<T>(), close: () => {} })
