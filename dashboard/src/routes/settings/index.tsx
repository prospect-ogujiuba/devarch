import { useState } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { Loader2, Monitor, Plug, Key, CheckCircle, XCircle } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { useRuntimeStatus, useSwitchRuntime, useSocketStatus, useStartSocket } from '@/features/runtime/queries'
import { getApiKey, setApiKey } from '@/lib/api'

export const Route = createFileRoute('/settings/')({
  component: SettingsPage,
})

function SettingsPage() {
  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-bold">Settings</h1>
        <p className="text-muted-foreground">Configure runtime, socket, and API settings</p>
      </div>

      <RuntimeSection />
      <SocketSection />
      <ApiKeySection />
    </div>
  )
}

function RuntimeSection() {
  const { data: runtime, isLoading } = useRuntimeStatus()
  const switchMutation = useSwitchRuntime()

  if (isLoading) {
    return <SectionLoader title="Container Runtime" />
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Monitor className="size-5" />
          Container Runtime
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="flex items-center gap-2">
          <span className="text-sm text-muted-foreground">Current:</span>
          <Badge variant={runtime?.current !== 'none' ? 'success' : 'destructive'}>
            {runtime?.current ?? 'none'}
          </Badge>
        </div>

        <div className="grid gap-3 md:grid-cols-2">
          {['podman', 'docker'].map((rt) => {
            const info = runtime?.available[rt]
            const isCurrent = runtime?.current === rt
            return (
              <div key={rt} className="border rounded-lg p-4 space-y-2">
                <div className="flex items-center justify-between">
                  <span className="font-medium capitalize">{rt}</span>
                  {info?.responsive ? (
                    <CheckCircle className="size-4 text-success" />
                  ) : (
                    <XCircle className="size-4 text-muted-foreground" />
                  )}
                </div>
                <div className="text-sm text-muted-foreground space-y-1">
                  <div>Installed: {info?.installed ? 'Yes' : 'No'}</div>
                  {info?.version && <div>Version: {info.version}</div>}
                  <div>Containers: {runtime?.containers[rt] ?? 0}</div>
                </div>
                {!isCurrent && info?.installed && (
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={() => switchMutation.mutate({
                      runtime: rt,
                      options: { stop_services: true, preserve_data: true, update_config: true },
                    })}
                    disabled={switchMutation.isPending}
                  >
                    {switchMutation.isPending ? <Loader2 className="size-4 animate-spin" /> : 'Switch'}
                  </Button>
                )}
              </div>
            )
          })}
        </div>

        <div className="text-sm text-muted-foreground">
          Network: {runtime?.microservices.network} ({runtime?.microservices.network_exists ? 'exists' : 'missing'}) &middot; {runtime?.microservices.running ?? 0} services running
        </div>
      </CardContent>
    </Card>
  )
}

function SocketSection() {
  const { data: socket, isLoading } = useSocketStatus()
  const startMutation = useStartSocket()

  if (isLoading) {
    return <SectionLoader title="Podman Socket" />
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Plug className="size-5" />
          Podman Socket
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="flex items-center gap-2">
          <span className="text-sm text-muted-foreground">Active:</span>
          <Badge variant={socket?.active !== 'none' ? 'success' : 'destructive'}>
            {socket?.active ?? 'none'}
          </Badge>
        </div>

        <div className="grid gap-3 md:grid-cols-2">
          {['rootless', 'rootful'].map((type) => {
            const info = socket?.sockets[type]
            const isActive = socket?.active === type
            return (
              <div key={type} className="border rounded-lg p-4 space-y-2">
                <div className="flex items-center justify-between">
                  <span className="font-medium capitalize">{type}</span>
                  {isActive ? (
                    <CheckCircle className="size-4 text-success" />
                  ) : (
                    <XCircle className="size-4 text-muted-foreground" />
                  )}
                </div>
                <div className="text-sm text-muted-foreground space-y-1">
                  <div>Path: {info?.socket_path ?? '-'}</div>
                  <div>Status: {info?.connectivity ?? 'unknown'}</div>
                </div>
                {!isActive && (
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={() => startMutation.mutate({
                      type,
                      options: { enable_lingering: true, stop_conflicting: true },
                    })}
                    disabled={startMutation.isPending}
                  >
                    {startMutation.isPending ? <Loader2 className="size-4 animate-spin" /> : 'Start'}
                  </Button>
                )}
              </div>
            )
          })}
        </div>

        <div className="text-sm text-muted-foreground">
          User: {socket?.environment.user} (UID {socket?.environment.uid}) &middot; DOCKER_HOST: {socket?.environment.docker_host || 'not set'}
        </div>
      </CardContent>
    </Card>
  )
}

function ApiKeySection() {
  const [key, setKey] = useState(getApiKey())
  const [saved, setSaved] = useState(false)

  function handleSave() {
    setApiKey(key)
    setSaved(true)
    setTimeout(() => setSaved(false), 2000)
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Key className="size-5" />
          API Key
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <p className="text-sm text-muted-foreground">
          Set the API key for authenticating with the Go API backend. Leave empty if auth is disabled.
        </p>
        <div className="flex gap-2 max-w-md">
          <Input
            type="password"
            placeholder="Enter API key..."
            value={key}
            onChange={(e) => setKey(e.target.value)}
          />
          <Button onClick={handleSave} variant={saved ? 'success' : 'default'}>
            {saved ? 'Saved' : 'Save'}
          </Button>
        </div>
      </CardContent>
    </Card>
  )
}

function SectionLoader({ title }: { title: string }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>{title}</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="flex items-center gap-2">
          <Loader2 className="size-4 animate-spin" />
          <span className="text-muted-foreground">Loading...</span>
        </div>
      </CardContent>
    </Card>
  )
}
