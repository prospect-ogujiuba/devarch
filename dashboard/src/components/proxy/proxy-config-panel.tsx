import { useState } from 'react'
import { Loader2, Download, Copy, Check, Shield } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Badge } from '@/components/ui/badge'
import { CodeEditor } from '@/components/services/code-editor'
import { useProxyTypes } from '@/features/proxy/queries'
import type { ProxyConfigResult } from '@/types/api'
import type { UseMutationResult } from '@tanstack/react-query'

interface ProxyConfigPanelProps {
  scope: string
  name: string
  generateMutation: UseMutationResult<ProxyConfigResult, Error, string>
}

const fileExtensions: Record<string, string> = {
  nginx: 'conf',
  caddy: 'Caddyfile',
  traefik: 'yml',
  haproxy: 'cfg',
  apache: 'conf',
}

const editorLanguages: Record<string, string> = {
  nginx: 'nginx',
  caddy: 'plaintext',
  traefik: 'yaml',
  haproxy: 'plaintext',
  apache: 'plaintext',
}

export function ProxyConfigPanel({ scope, name, generateMutation }: ProxyConfigPanelProps) {
  const { data: proxyTypes = [] } = useProxyTypes()
  const [selectedType, setSelectedType] = useState('')
  const [result, setResult] = useState<ProxyConfigResult | null>(null)
  const [copied, setCopied] = useState(false)

  const handleGenerate = () => {
    if (!selectedType) return
    generateMutation.mutate(selectedType, {
      onSuccess: (data) => setResult(data),
    })
  }

  const handleCopy = async () => {
    if (!result?.config) return
    await navigator.clipboard.writeText(result.config)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  const handleDownload = () => {
    if (!result?.config) return
    const ext = fileExtensions[result.proxy_type] ?? 'conf'
    const filename = `${name}.${result.proxy_type}.${ext}`
    const blob = new Blob([result.config], { type: 'text/plain' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = filename
    a.click()
    URL.revokeObjectURL(url)
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <CardTitle className="text-base flex items-center gap-2">
            <Shield className="size-4" />
            Reverse Proxy Config
          </CardTitle>
          <div className="flex w-full items-center gap-2 sm:w-auto">
            <Select value={selectedType} onValueChange={setSelectedType}>
              <SelectTrigger className="w-full sm:w-[160px]">
                <SelectValue placeholder="Select proxy..." />
              </SelectTrigger>
              <SelectContent>
                {proxyTypes.map((pt) => (
                  <SelectItem key={pt.id} value={pt.id}>
                    {pt.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Button
              size="sm"
              onClick={handleGenerate}
              disabled={!selectedType || generateMutation.isPending}
            >
              {generateMutation.isPending ? (
                <Loader2 className="size-4 animate-spin" />
              ) : null}
              Generate
            </Button>
          </div>
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        {result ? (
          <>
            <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
              <div className="flex flex-wrap items-center gap-2">
                <Badge variant="outline" className="capitalize">{result.proxy_type}</Badge>
                <Badge variant="secondary">{result.scope}: {result.name}</Badge>
                <Badge variant="secondary">{result.targets.length} target{result.targets.length !== 1 ? 's' : ''}</Badge>
              </div>
              <div className="flex items-center gap-2">
                <Button size="sm" variant="outline" onClick={handleCopy}>
                  {copied ? <Check className="size-4" /> : <Copy className="size-4" />}
                  {copied ? 'Copied' : 'Copy'}
                </Button>
                <Button size="sm" variant="outline" onClick={handleDownload}>
                  <Download className="size-4" />
                  Download
                </Button>
              </div>
            </div>
            <CodeEditor
              value={result.config}
              onChange={() => {}}
              language={editorLanguages[result.proxy_type] ?? 'plaintext'}
              readOnly
            />
          </>
        ) : (
          <div className="flex flex-col items-center justify-center py-8 text-center">
            <div className="rounded-full bg-muted p-3 mb-4">
              <Shield className="size-8 text-muted-foreground" />
            </div>
            <h3 className="font-medium mb-1">Generate reverse proxy config</h3>
            <p className="text-sm text-muted-foreground">
              Select a proxy type and click Generate to produce configuration for{' '}
              {scope === 'service' ? 'this service' : scope === 'stack' ? 'all instances in this stack' : 'this project'}
            </p>
            <p className="text-xs text-muted-foreground mt-2">
              Supports Nginx, Caddy, Traefik, HAProxy, and Apache
            </p>
          </div>
        )}
      </CardContent>
    </Card>
  )
}
