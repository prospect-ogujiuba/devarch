import { useEffect, useRef, useState } from 'react'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import { WebLinksAddon } from '@xterm/addon-web-links'
import '@xterm/xterm/css/xterm.css'
import { buildExecWsUrl } from '@/hooks/use-container-exec'

type ConnectionStatus = 'connecting' | 'connected' | 'disconnected'

interface ContainerTerminalProps {
  containerName: string
  onDisconnect?: () => void
}

export function ContainerTerminal({ containerName, onDisconnect }: ContainerTerminalProps) {
  const termRef = useRef<HTMLDivElement>(null)
  const [status, setStatus] = useState<ConnectionStatus>('connecting')

  useEffect(() => {
    if (!termRef.current) return

    const term = new Terminal({
      cursorBlink: true,
      fontSize: 14,
      fontFamily: 'Menlo, Monaco, "Courier New", monospace',
      theme: {
        background: '#282c34',
        foreground: '#abb2bf',
        cursor: '#528bff',
      },
    })

    const fitAddon = new FitAddon()
    const webLinksAddon = new WebLinksAddon()
    term.loadAddon(fitAddon)
    term.loadAddon(webLinksAddon)
    term.open(termRef.current)
    fitAddon.fit()

    let ws: WebSocket | null = null
    let resizeTimer: ReturnType<typeof setTimeout>
    let mounted = true

    async function connect() {
      const url = await buildExecWsUrl(containerName, term.cols, term.rows)
      if (!mounted) return
      ws = new WebSocket(url)

      ws.onopen = () => {
        setStatus('connected')
      }

      ws.onmessage = (event) => {
        term.write(event.data)
      }

      ws.onclose = () => {
        setStatus('disconnected')
        term.write('\r\n\x1b[31mDisconnected\x1b[0m\r\n')
        onDisconnect?.()
      }

      ws.onerror = () => {
        setStatus('disconnected')
      }

      term.onData((data) => {
        if (ws?.readyState === WebSocket.OPEN) {
          ws.send(data)
        }
      })

      term.onResize(({ cols, rows }) => {
        if (ws?.readyState === WebSocket.OPEN) {
          const msg = JSON.stringify({ type: 'resize', cols, rows })
          ws.send(new Blob([msg]))
        }
      })
    }

    const handleWindowResize = () => {
      clearTimeout(resizeTimer)
      resizeTimer = setTimeout(() => {
        fitAddon.fit()
      }, 250)
    }

    window.addEventListener('resize', handleWindowResize)
    connect()

    return () => {
      mounted = false
      window.removeEventListener('resize', handleWindowResize)
      clearTimeout(resizeTimer)
      ws?.close()
      term.dispose()
    }
  }, [containerName, onDisconnect])

  return (
    <div className="flex flex-col gap-2">
      <div className="flex items-center gap-2 text-sm">
        <div className={`h-2 w-2 rounded-full ${
          status === 'connected' ? 'bg-green-500' :
          status === 'connecting' ? 'bg-yellow-500 animate-pulse' :
          'bg-red-500'
        }`} />
        <span className="text-muted-foreground capitalize">{status}</span>
      </div>
      <div ref={termRef} className="h-[400px] w-full rounded-md overflow-hidden" />
    </div>
  )
}
