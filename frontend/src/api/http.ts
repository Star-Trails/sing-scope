// api 层 · axios 实例的全局拦截器。
// 这是 api 层唯一允许依赖 store/setup 的地方:请求需要从 activeBackend 取得
// 当前连接目标(baseURL / 鉴权)。其余 api 文件不得依赖上层。
import { ROUTE_NAME } from '@/constant'
import { showNotification } from '@/helper/notification'
import { getUrlFromBackend } from '@/helper/utils'
import { activeBackend, activeUuid } from '@/store/setup'
import axios, { AxiosError } from 'axios'
import { nextTick } from 'vue'

axios.interceptors.request.use((config) => {
  if (activeBackend.value) {
    config.baseURL = getUrlFromBackend(activeBackend.value)
    config.headers['Authorization'] = 'Bearer ' + activeBackend.value.password
  }
  return config
})

// 响应拦截器只做「401 → 退回 Setup 页」这一件事:它改变的是应用状态(清空当前后端
// 并跳转),任何请求打到 401 都必须如此,不是「要不要提示用户」的问题。
//
// 其余错误一律原样抛出,不在这里弹提示 —— 提示该由发起请求的业务层用 try-catch
// 决定(见 helper/requestError.ts):只有用户手动触发的动作才打扰用户,后台
// 自动拉取失败保持静默。以前靠 url 黑名单区分二者,加一个端点就得改一次名单,
// 而且拦截器根本不知道这次请求是谁发的、为什么发。
axios.interceptors.response.use(
  null,
  async (
    error: AxiosError<{
      message: string
    }>,
  ) => {
    if (error.status === 401 && activeUuid.value) {
      const { default: router } = await import('@/router')
      const currentBackendUuid = activeUuid.value
      activeUuid.value = null
      router.push({
        name: ROUTE_NAME.setup,
        query: { editBackend: currentBackendUuid },
      })
      nextTick(() => {
        showNotification({ content: 'unauthorizedTip' })
      })
    }

    return Promise.reject(error)
  },
)
