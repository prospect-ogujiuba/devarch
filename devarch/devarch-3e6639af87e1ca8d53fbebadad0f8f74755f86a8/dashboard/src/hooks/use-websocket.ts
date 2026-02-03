import { useEffect, useRef } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import type { WebSocketMessage } from '@/types/api'

const BASE_DELAY = 3000
const MAX_DELAY = 30000

export function useWebSocket() {
  const queryClient = useQueryClient()
  const wsRef = useRef<WebSocket | null>(null)
  const reconnectTimeoutRef = useRef<number | null>(null)
  const attemptRef = useRef(0)

  useEffect(() => {
    let mounted = true

    function getDelay() {
      const delay = Math.min(BASE_DELAY * 2 ** attemptRef.current, MAX_DELAY)
      attemptRef.current++
      return delay
    }

    function scheduleReconnect() {
      if (mounted) {
        reconnectTimeoutRef.current = window.setTimeout(connect, getDelay())
      }
    }

    function connect() {
      if (!mounted) return

      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      const wsUrl = `${protocol}//${window.location.host}/api/v1/ws/status`

      try {
        const ws = new WebSocket(wsUrl)
        wsRef.current = ws

        ws.onopen = () => {
          attemptRef.current = 0
        }

        ws.onmessage = (event) => {
          try {
            const message: WebSocketMessage = JSON.parse(event.data)
            if (message.type === 'status') {
              queryClient.invalidateQueries({ queryKey: ['services'] })
              queryClient.invalidateQueries({ queryKey: ['status'] })
              queryClient.invalidateQueries({ queryKey: ['categories'] })
              queryClient.invalidateQueries({ predicate: (q) => {
                const key = q.queryKey
                return Array.isArray(key) && key.length >= 3 && key[0] === 'services' && key[2] === 'metrics'
              }})
            }
          } catch {
            // ignore malformed messages
          }
        }

        ws.onclose = () => {
          scheduleReconnect()
        }

        ws.onerror = () => {
          ws.close()
        }
      } catch {
        scheduleReconnect()
      }
    }

    connect()

    return () => {
      mounted = false
      if (wsRef.current) {
        wsRef.current.close()
      }
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current)
      }
    }
  }, [queryClient])
}
