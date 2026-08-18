// 组装层 · 工具面板接口定义（已收口至本地）
export const getSingboxClient = () => null
export const serverStream = (_method: unknown, _req: unknown, _signal: unknown) => ({})
export const runStream = (_factory: unknown, _onMessage?: unknown, _onEnd?: unknown) => ({ close: () => {} })

export class GrpcWebSocketStream<Req = unknown> {
  close() {}
  send(_msg: Req) {}
}

export type StreamHandle = { close: () => void }
export type GrpcStatus = { code: number; message: string }
