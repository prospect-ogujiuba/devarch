import { useEffect, useRef } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { RefreshCw, Download, Loader2 } from 'lucide-react'
import { useServiceLogs } from '@/features/services/queries'

interface LogViewerProps {
  serviceName: string
}

export function LogViewer({ serviceName }: LogViewerProps) {
  const { data, isLoading, refetch, isRefetching } = useServiceLogs(serviceName, 200)
  const containerRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (containerRef.current) {
      containerRef.current.scrollTop = containerRef.current.scrollHeight
    }
  }, [data])

  const handleDownload = () => {
    if (!data) return
    const blob = new Blob([data], { type: 'text/plain' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `${serviceName}-logs.txt`
    a.click()
    URL.revokeObjectURL(url)
  }

  return (
    <Card>
      <CardHeader className="flex-row items-center justify-between pb-4">
        <CardTitle className="text-base">Logs</CardTitle>
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => refetch()}
            disabled={isRefetching}
          >
            {isRefetching ? (
              <Loader2 className="size-4 animate-spin" />
            ) : (
              <RefreshCw className="size-4" />
            )}
            Refresh
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={handleDownload}
            disabled={!data}
          >
            <Download className="size-4" />
            Download
          </Button>
        </div>
      </CardHeader>
      <CardContent>
        <div
          ref={containerRef}
          className="log-viewer h-[400px] overflow-auto rounded-lg bg-muted/50 p-4"
        >
          {isLoading ? (
            <div className="flex items-center justify-center h-full">
              <Loader2 className="size-6 animate-spin text-muted-foreground" />
            </div>
          ) : data ? (
            <pre className="text-xs leading-relaxed">{data}</pre>
          ) : (
            <div className="flex items-center justify-center h-full text-muted-foreground">
              No logs available
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  )
}
