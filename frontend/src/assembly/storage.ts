export const getStorageAPI = async () => Promise.reject<{ data: Record<string, unknown> }>('unsupported')
export const setStorageAPI = async (_value: Record<string, string>) => undefined
export const deleteStorageAPI = async () => undefined
