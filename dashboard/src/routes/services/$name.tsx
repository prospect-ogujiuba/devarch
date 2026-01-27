import { createFileRoute, Link } from '@tanstack/react-router'
import { ArrowLeft, Loader2, Clock, RefreshCw, Heart } from 'lucide-react'
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

  const maskedEnv = Object.entries(service.environment).map(([key, value]) => {
    const shouldMask = key.toLowerCase().includes('password') ||
      key.toLowerCase().includes('secret') ||
      key.toLowerCase().includes('key') ||
      key.toLowerCase().includes('token')
    return [key, shouldMask ? '••••••••' : value]
  })

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Link to="/services" className="text-muted-foreground hover:text-foreground">
          <ArrowLeft className="size-5" />
        </Link>
        <div className="flex-1">
          <div className="flex items-center gap-3">
            <h1 className="text-2xl font-bold">{service.name}</h1>
            <StatusBadge status={service.status} />
          </div>
          <p className="text-muted-foreground">{service.image}</p>
        </div>
        <ActionButton name={service.name} status={service.status} showRestart />
      </div>

      <div className="grid gap-4 md:grid-cols-3">
        <Card className="py-4">
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground flex items-center gap-2">
              <Heart className="size-4" />
              Health
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-lg font-semibold capitalize">
              {service.health ?? 'Unknown'}
            </div>
          </CardContent>
        </Card>

        <Card className="py-4">
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground flex items-center gap-2">
              <Clock className="size-4" />
              Uptime
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-lg font-semibold">
              {service.uptime ?? '-'}
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
            <div className="text-lg font-semibold">
              {service.restartCount ?? 0}
            </div>
          </CardContent>
        </Card>
      </div>

      <Tabs defaultValue="info" className="space-y-4">
        <TabsList>
          <TabsTrigger value="info">Info</TabsTrigger>
          <TabsTrigger value="logs">Logs</TabsTrigger>
          <TabsTrigger value="compose">Compose</TabsTrigger>
        </TabsList>

        <TabsContent value="info" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Ports</CardTitle>
            </CardHeader>
            <CardContent>
              {service.ports.length > 0 ? (
                <div className="flex flex-wrap gap-2">
                  {service.ports.map((port, i) => (
                    <Badge key={i} variant="outline">
                      {port.host}:{port.container}
                      {port.protocol && `/${port.protocol}`}
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
              {service.volumes.length > 0 ? (
                <div className="space-y-1">
                  {service.volumes.map((vol, i) => (
                    <div key={i} className="text-sm font-mono text-muted-foreground">
                      {vol}
                    </div>
                  ))}
                </div>
              ) : (
                <p className="text-muted-foreground">No volumes mounted</p>
              )}
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className="text-base">Environment Variables</CardTitle>
            </CardHeader>
            <CardContent>
              {maskedEnv.length > 0 ? (
                <div className="space-y-1">
                  {maskedEnv.map(([key, value], i) => (
                    <div key={i} className="text-sm font-mono flex">
                      <span className="text-muted-foreground min-w-[200px]">{key}:</span>
                      <span>{value}</span>
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
