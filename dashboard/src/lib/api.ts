import axios, { isAxiosError } from 'axios'

function getObjectErrorMessage(data: unknown): string | null {
  if (!data || typeof data !== 'object') return null

  const message = (data as Record<string, unknown>).message
  if (typeof message === 'string' && message.trim()) return message

  const error = (data as Record<string, unknown>).error
  if (typeof error === 'string' && error.trim()) return error

  const detail = (data as Record<string, unknown>).detail
  if (typeof detail === 'string' && detail.trim()) return detail

  const errors = (data as Record<string, unknown>).errors
  if (Array.isArray(errors) && errors.length > 0 && typeof errors[0] === 'object' && errors[0] !== null) {
    const firstMessage = (errors[0] as Record<string, unknown>).message
    if (typeof firstMessage === 'string' && firstMessage.trim()) return firstMessage
  }

  return null
}

export function getErrorMessage(error: unknown, fallback: string): string {
  if (isAxiosError(error)) {
    if (typeof error.response?.data === 'string' && error.response.data.trim()) {
      return error.response.data
    }

    const objectMessage = getObjectErrorMessage(error.response?.data)
    if (objectMessage) {
      return objectMessage
    }

    if (typeof error.message === 'string' && error.message.trim()) {
      return error.message
    }

    return fallback
  }

  if (error instanceof Error && typeof error.message === 'string' && error.message.trim()) {
    return error.message
  }

  return fallback
}

const apiKey = localStorage.getItem('devarch-api-key') ?? ''

export const api = axios.create({
  baseURL: '/api/v1',
  headers: {
    'Accept': 'application/json',
    'Content-Type': 'application/json',
    ...(apiKey ? { 'X-API-Key': apiKey } : {}),
  },
})

export function setApiKey(key: string) {
  localStorage.setItem('devarch-api-key', key)
  api.defaults.headers.common['X-API-Key'] = key
}

export function getApiKey(): string {
  return localStorage.getItem('devarch-api-key') ?? ''
}

export function clearApiKey() {
  localStorage.removeItem('devarch-api-key')
  delete api.defaults.headers.common['X-API-Key']
}

api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (isAxiosError(error) && error.response?.status === 401) {
      clearApiKey()
      window.location.href = '/login'
    }
    return Promise.reject(error)
  },
)
