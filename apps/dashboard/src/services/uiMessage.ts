import type { MessageInstance } from 'antd/es/message/interface'

let globalMessage: MessageInstance | null = null

export function setGlobalMessage(instance: MessageInstance) {
  globalMessage = instance
}

type MessageMethod = 'success' | 'error' | 'warning' | 'info' | 'loading' | 'open'

function call(method: MessageMethod, ...args: unknown[]) {
  const api = globalMessage as unknown as Record<string, (...inner: unknown[]) => unknown> | null
  if (!api || typeof api[method] !== 'function') {
    return
  }
  api[method](...args)
}

export const uiMessage = {
  success: (...args: unknown[]) => call('success', ...args),
  error: (...args: unknown[]) => call('error', ...args),
  warning: (...args: unknown[]) => call('warning', ...args),
  info: (...args: unknown[]) => call('info', ...args),
  loading: (...args: unknown[]) => call('loading', ...args),
  open: (...args: unknown[]) => call('open', ...args),
}
