import { createFileRoute, Link } from '@tanstack/react-router'
import { ArrowLeft, Loader2, Clock, RefreshCw, Heart, Cpu, MemoryStick } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { useService, useServiceCompose } from '@/features/services/queries'
import { StatusBadge } from '@/components/services/status-badge'
import { ActionButton } from '@/components/services/action-button'
import { LogViewer } from '@/components/services/log-viewer'

export const Route = createFileRoute('/services/$name')({
  component: ServiceDetailPage,
})

function ServiceDetailPage() {
  const { name } = Route.useParams()
  const { data: service, isLoading } = useService(name)
  const { data: composeYaml, isLoading: composeLoading } = useServiceCompose(name)

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="size-8 animate-spin text-muted-foreground" />
      </div>
    )
  }

  if (!service) {
    return (
      <div className="space-y-4">
        <Link to="/services" className="inline-flex items-center gap-2 text-muted-foreground hover:text-foreground">
          <ArrowLeft className="size-4" />
          Back to Services
        </Link>
        <div className="text-center py-12">
          <p className="text-muted-foreground">Service not found</p>
        </div>
      </div>
    )
  }

  const status = service.status?.status ?? 'stopped'
  const image = `${service.image_name}:${service.image_tag}`
  const healthStatus = service.status?.health_status ?? service.healthcheck ? 'configured' : 'none'

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Link to="/services" className="text-muted-foreground hover:text-foreground">
          <ArrowLeft className="size-5" />
        </Link>
        <div className="flex-1">
          <div className="flex items-center gap-3">
            <h1 className="text-2xl font-bold">{service.name}</h1>
            <StatusBadge status={status} />
          </div>
          <p className="text-muted-foreground">{image}</p>
        </div>
        <ActionButton name={service.name} status={status} showRestart />
      </div>

      <div className="grid gap-4 md:grid-cols-4">
        <Card className="py-4">
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground flex items-center gap-2">
              <Heart className="size-4" />
              Health
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-lg font-semibold capitalize">{healthStatus}</div>
          </CardContent>
        </Card>

        <Card className="py-4">
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground flex items-center gap-2">
              <Clock className="size-4" />
              Started
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-lg font-semibold">
              {service.status?.started_at ? new Date(service.status.started_at).toLocaleString() : '-'}
            </div>
          </CardContent>
        </Card>

        <Card className="py-4">
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground flex items-center gap-2">
              <RefreshCw className="size-4" />
              Restarts
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-lg font-semibold">{service.status?.restart_count ?? 0}</div>
          </CardContent>
        </Card>

        {service.metrics && (
          <Card className="py-4">
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground flex items-center gap-2">
                <Cpu className="size-4" />
                Resources
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-sm space-y-1">
                <div>CPU: {service.metrics.cpu_percentage.toFixed(1)}%</div>
                <div>Mem: {service.metrics.memory_used_mb.toFixed(0)}MB / {service.metrics.memory_limit_mb.toFixed(0)}MB</div>
              </div>
            </CardContent>
          </Card>
        )}
      </div>

      <Tabs defaultValue="info" className="space-y-4">
        <TabsList>
          <TabsTrigger value="info">Info</TabsTrigger>
          <TabsTrigger value="env">Environment</TabsTrigger>
          <TabsTrigger value="logs">Logs</TabsTrigger>
          <TabsTrigger value="compose">Compose</TabsTrigger>
        </TabsList>

        <TabsContent value="info" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Details</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid gap-2 text-sm">
                <div className="flex"><span className="text-muted-foreground w-40">Category:</span> {service.category?.name ?? '-'}</div>
                <div className="flex"><span className="text-muted-foreground w-40">Image:</span> {image}</div>
                <div className="flex"><span className="text-muted-foreground w-40">Restart Policy:</span> {service.restart_policy}</div>
                {service.command && <div className="flex"><span className="text-muted-foreground w-40">Command:</span> <code>{service.command}</code></div>}
                {service.user_spec && <div className="flex"><span className="text-muted-foreground w-40">User:</span> {service.user_spec}</div>}
                <div className="flex"><span className="text-muted-foreground w-40">Enabled:</span> {service.enabled ? 'Yes' : 'No'}</div>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className="text-base">Ports</CardTitle>
            </CardHeader>
            <CardContent>
              {service.ports && service.ports.length > 0 ? (
                <div className="flex flex-wrap gap-2">
                  {service.ports.map((port, i) => (
                    <Badge key={i} variant="outline">
                      {port.host_ip ? `${port.host_ip}:` : ''}{port.host_port}:{port.container_port}/{port.protocol}
                    </Badge>
                  ))}
                </div>
              ) : (
                <p className="text-muted-foreground">No ports exposed</p>
              )}
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className="text-base">Volumes</CardTitle>
            </CardHeader>
            <CardContent>
              {service.volumes && service.volumes.length > 0 ? (
                <div className="space-y-1">
                  {service.volumes.map((vol, i) => (
                    <div key={i} className="text-sm font-mono text-muted-foreground">
                      {vol.source}:{vol.target}{vol.read_only ? ' (ro)' : ''}
                    </div>
                  ))}
                </div>
              ) : (
                <p className="text-muted-foreground">No volumes mounted</p>
              )}
            </CardContent>
          </Card>

          {service.dependencies && service.dependencies.length > 0 && (
            <Card>
              <CardHeader>
                <CardTitle className="text-base">Dependencies</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="flex flex-wrap gap-2">
                  {service.dependencies.map((dep) => (
                    <Badge key={dep} variant="outline">
                      <Link to="/services/$name" params={{ name: dep }} className="hover:underline">
                        {dep}
                      </Link>
                    </Badge>
                  ))}
                </div>
              </CardContent>
            </Card>
          )}

          {service.healthcheck && (
            <Card>
              <CardHeader>
                <CardTitle className="text-base">Healthcheck</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="grid gap-2 text-sm">
                  <div className="flex"><span className="text-muted-foreground w-40">Test:</span> <code>{service.healthcheck.test}</code></div>
                  <div className="flex"><span className="text-muted-foreground w-40">Interval:</span> {service.healthcheck.interval_seconds}s</div>
                  <div className="flex"><span className="text-muted-foreground w-40">Timeout:</span> {service.healthcheck.timeout_seconds}s</div>
                  <div className="flex"><span className="text-muted-foreground w-40">Retries:</span> {service.healthcheck.retries}</div>
                </div>
              </CardContent>
            </Card>
          )}
        </TabsContent>

        <TabsContent value="env">
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Environment Variables</CardTitle>
            </CardHeader>
            <CardContent>
              {service.env_vars && service.env_vars.length > 0 ? (
                <div className="space-y-1">
                  {service.env_vars.map((env, i) => (
                    <div key={i} className="text-sm font-mono flex">
                      <span className="text-muted-foreground min-w-[200px]">{env.key}:</span>
                      <span>{env.is_secret ? '********' : env.value}</span>
                    </div>
                  ))}
                </div>
              ) : (
                <p className="text-muted-foreground">No environment variables</p>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="logs">
          <LogViewer serviceName={service.name} />
        </TabsContent>

        <TabsContent value="compose">
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Generated Compose YAML</CardTitle>
            </CardHeader>
            <CardContent>
              {composeLoading ? (
                <div className="flex items-center justify-center py-8">
                  <Loader2 className="size-6 animate-spin text-muted-foreground" />
                </div>
              ) : composeYaml ? (
                <pre className="text-sm font-mono bg-muted/50 p-4 rounded-lg overflow-auto max-h-[500px]">
                  {composeYaml}
                </pre>
              ) : (
                <p className="text-muted-foreground">Compose YAML not available</p>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  )
}
