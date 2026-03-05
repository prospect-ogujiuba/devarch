import { useState } from 'react'
import { createFileRoute, Link, useNavigate } from '@tanstack/react-router'
import { z } from 'zod'
import { ArrowLeft, Loader2, Clock, RefreshCw, Cpu, MemoryStick, Network, Pencil, Trash2, Download, Maximize2, Minimize2, Terminal, Save } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Tabs, TabsContent } from '@/components/ui/tabs'
import { StatCard } from '@/components/ui/stat-card'
import { CopyButton } from '@/components/ui/copy-button'
import { ConfirmDialog } from '@/components/ui/confirm-dialog'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { ResponsiveTabsList } from '@/components/ui/responsive-tabs-list'
import { useServiceDetailController } from '@/features/services/useServiceDetailController'
import { usePersistToLibrary } from '@/features/services/queries'
import { ProxyConfigPanel } from '@/components/proxy/proxy-config-panel'
import { StatusBadge } from '@/components/services/status-badge'
import { ActionButton } from '@/components/services/action-button'
import { LogViewer } from '@/components/services/log-viewer'
import { EditablePorts } from '@/components/services/editable-ports'
import { EditableVolumes } from '@/components/services/editable-volumes'
import { EditableEnvVars } from '@/components/services/editable-env-vars'
import { EditableDependencies } from '@/components/services/editable-dependencies'
import { EditableEnvFiles } from '@/components/services/editable-env-files'
import { EditableNetworks } from '@/components/services/editable-networks'
import { EditableConfigMounts } from '@/components/services/editable-config-mounts'
import { EditableHealthcheck } from '@/components/services/editable-healthcheck'
import { EditableLabels } from '@/components/services/editable-labels'
import { EditableDomains } from '@/components/services/editable-domains'
import { categoryLabel } from '@/lib/utils'
import { TagPicker } from '@/components/services/tag-picker'
import { ConfigFilesPanel } from '@/components/services/config-files-panel'
import { CodeEditor } from '@/components/services/code-editor'
import { TerminalDialog } from '@/components/terminal/terminal-dialog'
import { formatUptime, formatBytes } from '@/lib/format'

export const Route = createFileRoute('/services/$name')({
  validateSearch: z.object({
    tab: z.enum(['info', 'env', 'logs', 'compose', 'files', 'proxy']).optional(),
  }),
  component: ServiceDetailPage,
})

const serviceTabs = ['info', 'env', 'logs', 'compose', 'files', 'proxy'] as const
type ServiceTab = (typeof serviceTabs)[number]

function ServiceDetailPage() {
  const { name } = Route.useParams()
  const routeSearch = Route.useSearch()
  const routeNavigate = Route.useNavigate()
  const navigate = useNavigate()
  const ctrl = useServiceDetailController(name)

  const persistToLibrary = usePersistToLibrary()
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [composeExpanded, setComposeExpanded] = useState(false)
  const [editOpen, setEditOpen] = useState(false)
  const [terminalOpen, setTerminalOpen] = useState(false)
  const [editForm, setEditForm] = useState({
    image_name: '',
    image_tag: '',
    restart_policy: '',
    command: '',
    user_spec: '',
  })

  const openEdit = () => {
    if (!ctrl.service) return
    setEditForm({
      image_name: ctrl.service.image_name,
      image_tag: ctrl.service.image_tag,
      restart_policy: ctrl.service.restart_policy,
      command: ctrl.service.command ?? '',
      user_spec: ctrl.service.user_spec ?? '',
    })
    setEditOpen(true)
  }

  const handleDelete = () => {
    ctrl.deleteService.mutate(name, {
      onSuccess: () => navigate({ to: '/services' }),
    })
  }

  const handleEditSave = () => {
    ctrl.updateService.mutate(
      { name, data: editForm },
      { onSuccess: () => setEditOpen(false) },
    )
  }

  if (ctrl.isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="size-8 animate-spin text-muted-foreground" />
      </div>
    )
  }

  if (!ctrl.service) {
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

  const activeTab: ServiceTab = routeSearch.tab ?? 'info'
  const serviceTabItems = [
    { value: 'info', label: 'Info' },
    { value: 'env', label: 'Environment' },
    { value: 'logs', label: 'Logs' },
    { value: 'compose', label: 'Compose' },
    { value: 'files', label: 'Files' },
    { value: 'proxy', label: 'Proxy' },
  ]

  const handleTabChange = (tab: string) => {
    if (!serviceTabs.includes(tab as ServiceTab)) return
    routeNavigate({ search: (prev) => ({ ...prev, tab: tab as ServiceTab }) })
  }

  return (
    <div className="space-y-6">
      <div className="space-y-3">
        <Link to="/services" className="inline-flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground">
          <ArrowLeft className="size-4" />
          Back to Services
        </Link>

        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between sm:gap-4">
          <div className="min-w-0">
            <div className="flex flex-wrap items-center gap-2 sm:gap-3">
              <h1 className="truncate text-xl font-bold sm:text-2xl">{ctrl.service.name}</h1>
              <StatusBadge status={ctrl.status} />
            </div>
            <p className="mt-1 truncate text-sm text-muted-foreground sm:text-base">{ctrl.image}</p>
          </div>
          <div className="grid w-full grid-cols-2 gap-2 sm:flex sm:w-auto sm:flex-wrap sm:items-center sm:justify-end">
            <ActionButton
              name={ctrl.service.name}
              status={ctrl.status}
              showRestart
              className="col-span-2"
              buttonClassName="w-full sm:w-auto"
            />
            {ctrl.status === 'running' && (
              <Button variant="outline" size="sm" onClick={() => setTerminalOpen(true)} className="w-full sm:w-auto">
                <Terminal className="size-4" />
                Terminal
              </Button>
            )}
            <Button
              variant="outline"
              size="sm"
              className="w-full sm:w-auto"
              disabled={persistToLibrary.isPending}
              onClick={() => persistToLibrary.mutate(name)}
            >
              {persistToLibrary.isPending ? <Loader2 className="size-4 animate-spin" /> : <Save className="size-4" />}
              Save to Library
            </Button>
            <Button variant="outline" size="sm" onClick={openEdit} className="w-full sm:w-auto">
              <Pencil className="size-4" />
              Edit
            </Button>
            <Button variant="destructive" size="sm" onClick={() => setDeleteOpen(true)} className="w-full sm:w-auto">
              <Trash2 className="size-4" />
              Delete
            </Button>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-2 gap-3 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard icon={Cpu} label="CPU" value={`${ctrl.cpuPct.toFixed(1)}%`} />
        <StatCard
          icon={MemoryStick}
          label="Memory"
          value={`${ctrl.memUsed.toFixed(0)} / ${ctrl.memLimit.toFixed(0)} MB`}
        />
        <StatCard
          icon={Network}
          label="Network I/O"
          value={`${formatBytes(ctrl.rxBytes)} / ${formatBytes(ctrl.txBytes)}`}
        />
        <StatCard
          icon={Clock}
          label="Uptime"
          value={ctrl.status === 'running' && ctrl.uptime > 0 ? formatUptime(ctrl.uptime) : '-'}
        />
      </div>

      <Tabs value={activeTab} onValueChange={handleTabChange} className="space-y-4">
        <ResponsiveTabsList tabs={serviceTabItems} value={activeTab} onValueChange={handleTabChange} />

        <TabsContent value="info" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Details</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid gap-2 text-sm">
                <div className="flex"><span className="text-muted-foreground w-40">Category:</span> {ctrl.service.category ? categoryLabel(ctrl.service.category) : '-'}</div>
                <div className="flex"><span className="text-muted-foreground w-40">Image:</span> {ctrl.image}</div>
                <div className="flex"><span className="text-muted-foreground w-40">Restart Policy:</span> {ctrl.service.restart_policy}</div>
                {ctrl.service.command && <div className="flex"><span className="text-muted-foreground w-40">Command:</span> <code>{ctrl.service.command}</code></div>}
                {ctrl.service.user_spec && <div className="flex"><span className="text-muted-foreground w-40">User:</span> {ctrl.service.user_spec}</div>}
                <div className="flex"><span className="text-muted-foreground w-40">Enabled:</span> {ctrl.service.enabled ? 'Yes' : 'No'}</div>
                <div className="flex items-center">
                  <span className="text-muted-foreground w-40">Health:</span>
                  <Badge variant={ctrl.healthStatus === 'healthy' ? 'success' : ctrl.healthStatus === 'unhealthy' ? 'destructive' : 'secondary'}>
                    {ctrl.healthStatus}
                  </Badge>
                </div>
                {ctrl.service.status?.container_id && (
                  <div className="flex items-center">
                    <span className="text-muted-foreground w-40">Container ID:</span>
                    <code className="font-mono text-xs">{ctrl.service.status.container_id.slice(0, 12)}</code>
                    <CopyButton value={ctrl.service.status.container_id} className="ml-1" />
                  </div>
                )}
                {ctrl.service.status?.restart_count !== undefined && (
                  <div className="flex">
                    <span className="text-muted-foreground w-40">Restarts:</span>
                    <span className="flex items-center gap-1">
                      <RefreshCw className="size-3" /> {ctrl.service.status.restart_count}
                    </span>
                  </div>
                )}
              </div>
            </CardContent>
          </Card>

          <EditableDomains name={ctrl.service.name} domains={ctrl.service.domains ?? []} />
          <EditablePorts name={ctrl.service.name} ports={ctrl.service.ports ?? []} />
          <EditableVolumes name={ctrl.service.name} volumes={ctrl.service.volumes ?? []} />
          <EditableEnvFiles name={ctrl.service.name} envFiles={ctrl.service.env_files ?? []} />
          <EditableNetworks name={ctrl.service.name} networks={ctrl.service.networks ?? []} />
          <EditableDependencies name={ctrl.service.name} dependencies={ctrl.service.dependencies ?? []} />
          <EditableConfigMounts name={ctrl.service.name} configMounts={ctrl.service.config_mounts ?? []} />
          <EditableHealthcheck name={ctrl.service.name} healthcheck={ctrl.service.healthcheck ?? null} />
          <EditableLabels name={ctrl.service.name} labels={ctrl.service.labels ?? []} />
        </TabsContent>

        <TabsContent value="env">
          <EditableEnvVars name={ctrl.service.name} envVars={ctrl.service.env_vars ?? []} />
        </TabsContent>

        <TabsContent value="logs">
          <LogViewer serviceName={ctrl.service.name} />
        </TabsContent>

        <TabsContent value="compose">
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <CardTitle className="text-base">Generated Compose YAML</CardTitle>
                <div className="flex items-center gap-2">
                  <Button size="sm" variant="outline" onClick={() => setComposeExpanded(!composeExpanded)} disabled={!ctrl.composeYaml}>
                    {composeExpanded ? <Minimize2 className="size-4" /> : <Maximize2 className="size-4" />}
                    {composeExpanded ? 'Collapse' : 'Expand'}
                  </Button>
                  <Button size="sm" variant="outline" disabled={!ctrl.composeYaml} onClick={() => {
                    if (!ctrl.composeYaml) return
                    const blob = new Blob([ctrl.composeYaml], { type: 'text/yaml' })
                    const url = URL.createObjectURL(blob)
                    const a = document.createElement('a')
                    a.href = url
                    a.download = `docker-compose-${name}.yml`
                    a.click()
                    URL.revokeObjectURL(url)
                  }}>
                    <Download className="size-4" />
                    Download
                  </Button>
                </div>
              </div>
            </CardHeader>
            <CardContent>
              {ctrl.composeLoading ? (
                <div className="flex items-center justify-center py-8">
                  <Loader2 className="size-6 animate-spin text-muted-foreground" />
                </div>
              ) : ctrl.composeYaml ? (
                <CodeEditor value={ctrl.composeYaml} onChange={() => {}} language="yaml" readOnly autoHeight={composeExpanded} />
              ) : (
                <p className="text-muted-foreground">Compose YAML not available</p>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="files">
          <ConfigFilesPanel serviceName={ctrl.service.name} />
        </TabsContent>

        <TabsContent value="proxy">
          <ProxyConfigPanel scope="service" name={ctrl.service.name} generateMutation={ctrl.generateProxyConfig} />
        </TabsContent>
      </Tabs>

      <ConfirmDialog
        open={deleteOpen}
        onOpenChange={setDeleteOpen}
        title={`Delete ${ctrl.service.name}?`}
        description="This will permanently delete the service and all its configuration. This action cannot be undone."
        confirmLabel="Delete"
        onConfirm={handleDelete}
        variant="destructive"
      />

      <Dialog open={editOpen} onOpenChange={setEditOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Edit {ctrl.service.name}</DialogTitle>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <label className="text-sm font-medium">Image Name</label>
              <Input value={editForm.image_name} onChange={(e) => setEditForm((f) => ({ ...f, image_name: e.target.value }))} />
            </div>
            <div className="grid gap-2">
              <label className="text-sm font-medium">Image Tag</label>
              <TagPicker imageName={editForm.image_name} value={editForm.image_tag} onValueChange={(v) => setEditForm((f) => ({ ...f, image_tag: v }))} />
            </div>
            <div className="grid gap-2">
              <label className="text-sm font-medium">Restart Policy</label>
              <Select value={editForm.restart_policy} onValueChange={(v) => setEditForm((f) => ({ ...f, restart_policy: v }))}>
                <SelectTrigger className="w-full">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="no">no</SelectItem>
                  <SelectItem value="always">always</SelectItem>
                  <SelectItem value="unless-stopped">unless-stopped</SelectItem>
                  <SelectItem value="on-failure">on-failure</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="grid gap-2">
              <label className="text-sm font-medium">Command</label>
              <Input value={editForm.command} onChange={(e) => setEditForm((f) => ({ ...f, command: e.target.value }))} placeholder="Optional" />
            </div>
            <div className="grid gap-2">
              <label className="text-sm font-medium">User</label>
              <Input value={editForm.user_spec} onChange={(e) => setEditForm((f) => ({ ...f, user_spec: e.target.value }))} placeholder="Optional" />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setEditOpen(false)}>Cancel</Button>
            <Button onClick={handleEditSave} disabled={ctrl.updateService.isPending}>
              {ctrl.updateService.isPending ? <Loader2 className="size-4 animate-spin" /> : 'Save'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <TerminalDialog containerName={ctrl.service.container_name_template || ctrl.service.name} open={terminalOpen} onOpenChange={setTerminalOpen} />
    </div>
  )
}
