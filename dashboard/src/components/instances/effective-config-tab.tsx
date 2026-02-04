import { useState } from 'react'
import { Copy, Check } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { useEffectiveConfig } from '@/features/instances/queries'
import { cn } from '@/lib/utils'
import YAML from 'yaml'

interface Props {
  stackName: string
  instanceId: string
}

export function EffectiveConfigTab({ stackName, instanceId }: Props) {
  const { data: raw, isLoading } = useEffectiveConfig(stackName, instanceId)
  const [format, setFormat] = useState<'yaml' | 'json'>('yaml')
  const [copied, setCopied] = useState(false)

  if (isLoading) {
    return <div className="text-muted-foreground">Loading...</div>
  }

  if (!raw) {
    return <div className="text-muted-foreground">No configuration available</div>
  }

  const config = {
    ...raw,
    ports: raw.ports ?? [],
    volumes: raw.volumes ?? [],
    env_vars: raw.env_vars ?? [],
    labels: raw.labels ?? [],
    domains: raw.domains ?? [],
    dependencies: raw.dependencies ?? [],
    config_files: raw.config_files ?? [],
  }

  const oa = config.overrides_applied ?? {}
  const o = {
    ports: !!oa.ports,
    volumes: !!oa.volumes,
    env_vars: !!oa.env_vars,
    labels: !!oa.labels,
    domains: !!oa.domains,
    healthcheck: !!oa.healthcheck,
    config_files: !!oa.config_files,
  }

  const handleCopy = () => {
    const content = format === 'yaml' ? toYAML(config) : JSON.stringify(config, null, 2)
    navigator.clipboard.writeText(content)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div className="text-sm text-muted-foreground">
          Merged configuration from template and overrides
        </div>
        <div className="flex items-center gap-2">
          <div className="flex border rounded-md">
            <Button
              variant={format === 'yaml' ? 'default' : 'ghost'}
              size="sm"
              onClick={() => setFormat('yaml')}
              className="h-7 rounded-r-none"
            >
              YAML
            </Button>
            <Button
              variant={format === 'json' ? 'default' : 'ghost'}
              size="sm"
              onClick={() => setFormat('json')}
              className="h-7 rounded-l-none"
            >
              JSON
            </Button>
          </div>
          <Button size="sm" variant="outline" onClick={handleCopy}>
            {copied ? <Check className="size-4" /> : <Copy className="size-4" />}
            {copied ? 'Copied' : 'Copy'}
          </Button>
        </div>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Image</CardTitle>
        </CardHeader>
        <CardContent className="space-y-2 text-sm">
          <div className="flex"><span className="text-muted-foreground w-32">Image:</span> <code>{config.image_name}:{config.image_tag}</code></div>
          <div className="flex"><span className="text-muted-foreground w-32">Restart:</span> {config.restart_policy}</div>
          {config.command && (
            <div className="flex"><span className="text-muted-foreground w-32">Command:</span> <code>{config.command}</code></div>
          )}
          <div className="flex"><span className="text-muted-foreground w-32">Container:</span> <code>{config.container_name}</code></div>
        </CardContent>
      </Card>

      <Card className={cn(o.ports && 'border-l-2 border-l-blue-500')}>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle className="text-base">Ports</CardTitle>
            {o.ports && <Badge variant="secondary">Overridden</Badge>}
          </div>
        </CardHeader>
        <CardContent>
          {config.ports.length === 0 ? (
            <p className="text-sm text-muted-foreground">No ports configured</p>
          ) : (
            <div className="overflow-auto">
              <table className="w-full text-sm">
                <thead className="border-b">
                  <tr className="text-left">
                    <th className="pb-2 font-medium">Host IP</th>
                    <th className="pb-2 font-medium">Host Port</th>
                    <th className="pb-2 font-medium">Container Port</th>
                    <th className="pb-2 font-medium">Protocol</th>
                  </tr>
                </thead>
                <tbody>
                  {config.ports.map((p, i) => (
                    <tr key={i} className="border-b last:border-b-0">
                      <td className="py-2 font-mono">{p.host_ip}</td>
                      <td className="py-2 font-mono">{p.host_port}</td>
                      <td className="py-2 font-mono">{p.container_port}</td>
                      <td className="py-2 font-mono">{p.protocol}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </CardContent>
      </Card>

      <Card className={cn(o.volumes && 'border-l-2 border-l-blue-500')}>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle className="text-base">Volumes</CardTitle>
            {o.volumes && <Badge variant="secondary">Overridden</Badge>}
          </div>
        </CardHeader>
        <CardContent>
          {config.volumes.length === 0 ? (
            <p className="text-sm text-muted-foreground">No volumes configured</p>
          ) : (
            <div className="overflow-auto">
              <table className="w-full text-sm">
                <thead className="border-b">
                  <tr className="text-left">
                    <th className="pb-2 font-medium">Source</th>
                    <th className="pb-2 font-medium">Target</th>
                    <th className="pb-2 font-medium">Type</th>
                    <th className="pb-2 font-medium">Read-Only</th>
                  </tr>
                </thead>
                <tbody>
                  {config.volumes.map((v, i) => (
                    <tr key={i} className="border-b last:border-b-0">
                      <td className="py-2 font-mono">{v.source}</td>
                      <td className="py-2 font-mono">{v.target}</td>
                      <td className="py-2">{v.volume_type}</td>
                      <td className="py-2">{v.read_only ? 'Yes' : 'No'}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </CardContent>
      </Card>

      <Card className={cn(o.env_vars && 'border-l-2 border-l-blue-500')}>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle className="text-base">Environment Variables</CardTitle>
            {o.env_vars && <Badge variant="secondary">Overridden</Badge>}
          </div>
        </CardHeader>
        <CardContent>
          {config.env_vars.length === 0 ? (
            <p className="text-sm text-muted-foreground">No environment variables</p>
          ) : (
            <div className="overflow-auto">
              <table className="w-full text-sm">
                <thead className="border-b">
                  <tr className="text-left">
                    <th className="pb-2 font-medium">Key</th>
                    <th className="pb-2 font-medium">Value</th>
                  </tr>
                </thead>
                <tbody>
                  {config.env_vars.map((e, i) => (
                    <tr key={i} className="border-b last:border-b-0">
                      <td className="py-2 font-mono">{e.key}</td>
                      <td className="py-2 font-mono">{e.is_secret ? '***' : e.value ?? ''}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </CardContent>
      </Card>

      <Card className={cn(o.labels && 'border-l-2 border-l-blue-500')}>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle className="text-base">Labels</CardTitle>
            {o.labels && <Badge variant="secondary">Overridden</Badge>}
          </div>
        </CardHeader>
        <CardContent>
          {config.labels.length === 0 ? (
            <p className="text-sm text-muted-foreground">No labels</p>
          ) : (
            <div className="overflow-auto">
              <table className="w-full text-sm">
                <thead className="border-b">
                  <tr className="text-left">
                    <th className="pb-2 font-medium">Key</th>
                    <th className="pb-2 font-medium">Value</th>
                  </tr>
                </thead>
                <tbody>
                  {config.labels.map((l, i) => (
                    <tr key={i} className="border-b last:border-b-0">
                      <td className="py-2 font-mono">{l.key}</td>
                      <td className="py-2 font-mono">{l.value}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </CardContent>
      </Card>

      <Card className={cn(o.domains && 'border-l-2 border-l-blue-500')}>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle className="text-base">Domains</CardTitle>
            {o.domains && <Badge variant="secondary">Overridden</Badge>}
          </div>
        </CardHeader>
        <CardContent>
          {config.domains.length === 0 ? (
            <p className="text-sm text-muted-foreground">No domains configured</p>
          ) : (
            <div className="overflow-auto">
              <table className="w-full text-sm">
                <thead className="border-b">
                  <tr className="text-left">
                    <th className="pb-2 font-medium">Domain</th>
                    <th className="pb-2 font-medium">Proxy Port</th>
                  </tr>
                </thead>
                <tbody>
                  {config.domains.map((d, i) => (
                    <tr key={i} className="border-b last:border-b-0">
                      <td className="py-2 font-mono">{d.domain}</td>
                      <td className="py-2 font-mono">{d.proxy_port ?? 'N/A'}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </CardContent>
      </Card>

      <Card className={cn(o.healthcheck && 'border-l-2 border-l-blue-500')}>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle className="text-base">Healthcheck</CardTitle>
            {o.healthcheck && <Badge variant="secondary">Overridden</Badge>}
          </div>
        </CardHeader>
        <CardContent>
          {!config.healthcheck ? (
            <p className="text-sm text-muted-foreground">No healthcheck configured</p>
          ) : (
            <div className="space-y-2 text-sm">
              <div className="flex"><span className="text-muted-foreground w-32">Test:</span> <code>{config.healthcheck.test}</code></div>
              <div className="flex"><span className="text-muted-foreground w-32">Interval:</span> {config.healthcheck.interval_seconds}s</div>
              <div className="flex"><span className="text-muted-foreground w-32">Timeout:</span> {config.healthcheck.timeout_seconds}s</div>
              <div className="flex"><span className="text-muted-foreground w-32">Retries:</span> {config.healthcheck.retries}</div>
              <div className="flex"><span className="text-muted-foreground w-32">Start Period:</span> {config.healthcheck.start_period_seconds}s</div>
            </div>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Dependencies</CardTitle>
        </CardHeader>
        <CardContent>
          {config.dependencies.length === 0 ? (
            <p className="text-sm text-muted-foreground">No dependencies</p>
          ) : (
            <div className="space-y-1">
              {config.dependencies.map((dep, i) => (
                <div key={i} className="flex items-center gap-2 text-sm">
                  <code className="font-mono">{dep}</code>
                  <Badge variant="outline" className="text-xs">from template</Badge>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      <Card className={cn(o.config_files && 'border-l-2 border-l-blue-500')}>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle className="text-base">Config Files</CardTitle>
            {o.config_files && <Badge variant="secondary">Overridden</Badge>}
          </div>
        </CardHeader>
        <CardContent>
          {config.config_files.length === 0 ? (
            <p className="text-sm text-muted-foreground">No config files</p>
          ) : (
            <div className="space-y-2">
              {config.config_files.map((f, i) => (
                <div key={i} className="p-3 border rounded-md">
                  <div className="flex items-center justify-between mb-2">
                    <code className="text-sm font-mono">{f.file_path}</code>
                    <Badge variant="outline" className="text-xs">{f.file_mode}</Badge>
                  </div>
                  <pre className="text-xs bg-muted p-2 rounded overflow-x-auto max-h-32">
                    {f.content}
                  </pre>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}

function toYAML(config: any): string {
  const doc: any = {
    version: '3.8',
    services: {
      [config.instance_id]: {
        image: `${config.image_name}:${config.image_tag}`,
        container_name: config.container_name,
        restart: config.restart_policy,
      },
    },
  }

  const svc = doc.services[config.instance_id]

  if (config.command) {
    svc.command = config.command
  }

  if (config.ports?.length > 0) {
    svc.ports = config.ports.map((p: any) => `${p.host_ip}:${p.host_port}:${p.container_port}/${p.protocol}`)
  }

  if (config.volumes?.length > 0) {
    svc.volumes = config.volumes.map((v: any) => {
      const vol = `${v.source}:${v.target}`
      return v.read_only ? `${vol}:ro` : vol
    })
  }

  if (config.env_vars?.length > 0) {
    svc.environment = config.env_vars.reduce((acc: any, e: any) => {
      acc[e.key] = e.is_secret ? '***' : (e.value ?? '')
      return acc
    }, {})
  }

  if (config.labels?.length > 0) {
    svc.labels = config.labels.reduce((acc: any, l: any) => {
      acc[l.key] = l.value
      return acc
    }, {})
  }

  if (config.dependencies?.length > 0) {
    svc.depends_on = config.dependencies
  }

  if (config.healthcheck) {
    svc.healthcheck = {
      test: config.healthcheck.test,
      interval: `${config.healthcheck.interval_seconds}s`,
      timeout: `${config.healthcheck.timeout_seconds}s`,
      retries: config.healthcheck.retries,
      start_period: `${config.healthcheck.start_period_seconds}s`,
    }
  }

  return YAML.stringify(doc)
}
