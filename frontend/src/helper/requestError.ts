import { showNotification } from './notification'

export const getRequestErrorMessage = (error: unknown): string => {
  if (error instanceof Error) {
    return error.message
  }
  return String(error || 'Unknown error')
}

export const notifyRequestError = (error: unknown) => {
  const message = getRequestErrorMessage(error)
  showNotification({
    key: message,
    content: message,
    type: 'alert-error',
  })
}
