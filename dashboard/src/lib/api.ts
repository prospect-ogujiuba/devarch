import axios from 'axios'

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
