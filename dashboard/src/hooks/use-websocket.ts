import { useEffect, useRef } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import type { WebSocketMessage } from '@/types/api'

export function useWebSocket() {
  const queryClient = useQueryClient()
  const wsRef = useRef<WebSocket | null>(null)
  const reconnectTimeoutRef = useRef<number | null>(null)

  useEffect(() => {
    let mounted = true

    function connect() {
      if (!mounted) return

      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      const wsUrl = `${protocol}//${window.location.host}/api/v1/ws/status`

      try {
        const ws = new WebSocket(wsUrl)
        wsRef.current = ws

        ws.onopen = () => {
          console.log('WebSocket connected')
        }

        ws.onmessage = (event) => {
          try {
            const message: WebSocketMessage = JSON.parse(event.data)
            if (message.type === 'status') {
              queryClient.invalidateQueries({ queryKey: ['services'] })
              queryClient.invalidateQueries({ queryKey: ['status'] })
              queryClient.invalidateQueries({ queryKey: ['categories'] })
            }
          } catch (e) {
            console.error('Failed to parse WebSocket message:', e)
          }
        }

        ws.onclose = () => {
          console.log('WebSocket disconnected, reconnecting...')
          if (mounted) {
            reconnectTimeoutRef.current = window.setTimeout(connect, 3000)
          }
        }

        ws.onerror = (error) => {
          console.error('WebSocket error:', error)
          ws.close()
        }
      } catch (error) {
        console.error('Failed to create WebSocket:', error)
        if (mounted) {
          reconnectTimeoutRef.current = window.setTimeout(connect, 3000)
        }
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
