/// <reference types="vite/client" />

declare const __APP_VERSION__: string
declare const __COMMIT_ID__: string
declare const __FONT__: string

interface Window {
  ksu?: object
  wails?: any
  go?: any
}

declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<object, object, unknown>
  export default component
}
