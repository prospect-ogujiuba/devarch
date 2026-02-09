import { useState } from 'react'
import { createFileRoute, Link, useNavigate } from '@tanstack/react-router'
import { z } from 'zod'
import { ArrowLeft, Loader2, Clock, RefreshCw, Cpu, MemoryStick, Network, Pencil, Trash2, Download, Maximize2, Minimize2 } from 'lucide-react'
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
import { useService, useServiceCompose, useDeleteService, useUpdateService } from '@/features/services/queries'
import { StatusBadge } from '@/components/services/status-badge'
import { ActionButton } from '@/components/services/action-button'
import { LogViewer } from '@/components/services/log-viewer'
import { EditablePorts } from '@/components/services/editable-ports'
import { EditableVolumes } from '@/components/services/editable-volumes'
import { EditableEnvVars } from '@/components/services/editable-env-vars'
import { EditableDependencies } from '@/components/services/editable-dependencies'
import { EditableHealthcheck } from '@/components/services/editable-healthcheck'
import { EditableLabels } from '@/components/services/editable-labels'
import { EditableDomains } from '@/components/services/editable-domains'
import { TagPicker } from '@/components/services/tag-picker'
import { ConfigFilesPanel } from '@/components/services/config-files-panel'
import { CodeEditor } from '@/components/services/code-editor'
import { formatUptime, formatBytes, computeUptime } from '@/lib/format'

export const Route = createFileRoute('/services/$name')({
  validateSearch: z.object({
    tab: z.enum(['info', 'env', 'logs', 'compose', 'files']).optional(),
  }),
  component: ServiceDetailPage,
})

const serviceTabs = ['info', 'env', 'logs', 'compose', 'files'] as const
type ServiceTab = (typeof serviceTabs)[number]

function ServiceDetailPage() {
  const { name } = Route.useParams()
  const routeSearch = Route.useSearch()
  const routeNavigate = Route.useNavigate()
  const navigate = useNavigate()
  const { data: service, isLoading } = useService(name)
  const { data: composeYaml, isLoading: composeLoading } = useServiceCompose(name)
  const deleteService = useDeleteService()
  const updateService = useUpdateService()

  const [deleteOpen, setDeleteOpen] = useState(false)
  const [composeExpanded, setComposeExpanded] = useState(false)
  const [editOpen, setEditOpen] = useState(false)
  const [editForm, setEditForm] = useState({
    image_name: '',
    image_tag: '',
    restart_policy: '',
    command: '',
    user_spec: '',
  })

  const openEdit = () => {
    if (!service) return
    setEditForm({
      image_name: service.image_name,
      image_tag: service.image_tag,
      restart_policy: service.restart_policy,
      command: service.command ?? '',
      user_spec: service.user_spec ?? '',
    })
    setEditOpen(true)
  }

  const handleDelete = () => {
    deleteService.mutate(name, {
      onSuccess: () => navigate({ to: '/services' }),
    })
  }

  const handleEditSave = () => {
    updateService.mutate(
      { name, data: editForm },
      { onSuccess: () => setEditOpen(false) },
    )
  }

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
  const healthStatus = service.status?.health_status ?? (service.healthcheck ? 'configured' : 'none')
  const uptime = computeUptime(service.status?.started_at)
  const cpuPct = service.metrics?.cpu_percentage ?? 0
  const memUsed = service.metrics?.memory_used_mb ?? 0
  const memLimit = service.metrics?.memory_limit_mb ?? 0

  const rxBytes = service.metrics?.network_rx_bytes ?? 0
  const txBytes = service.metrics?.network_tx_bytes ?? 0
  const activeTab: ServiceTab = routeSearch.tab ?? 'info'
  const serviceTabItems = [
    { value: 'info', label: 'Info' },
    { value: 'env', label: 'Environment' },
    { value: 'logs', label: 'Logs' },
    { value: 'compose', label: 'Compose' },
    { value: 'files', label: 'Files' },
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
              <h1 className="truncate text-xl font-bold sm:text-2xl">{service.name}</h1>
              <StatusBadge status={status} />
            </div>
            <p className="mt-1 truncate text-sm text-muted-foreground sm:text-base">{image}</p>
          </div>
          <div className="grid w-full grid-cols-2 gap-2 sm:flex sm:w-auto sm:flex-wrap sm:items-center sm:justify-end">
            <ActionButton
              name={service.name}
              status={status}
              showRestart
              className="col-span-2"
              buttonClassName="w-full sm:w-auto"
            />
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
        <StatCard icon={Cpu} label="CPU" value={`${cpuPct.toFixed(1)}%`} />
        <StatCard
          icon={MemoryStick}
          label="Memory"
          value={`${memUsed.toFixed(0)} / ${memLimit.toFixed(0)} MB`}
        />
        <StatCard
          icon={Network}
          label="Network I/O"
          value={`${formatBytes(rxBytes)} / ${formatBytes(txBytes)}`}
        />
        <StatCard
          icon={Clock}
          label="Uptime"
          value={status === 'running' && uptime > 0 ? formatUptime(uptime) : '-'}
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
                <div className="flex"><span className="text-muted-foreground w-40">Category:</span> {service.category?.name ?? '-'}</div>
                <div className="flex"><span className="text-muted-foreground w-40">Image:</span> {image}</div>
                <div className="flex"><span className="text-muted-foreground w-40">Restart Policy:</span> {service.restart_policy}</div>
                {service.command && <div className="flex"><span className="text-muted-foreground w-40">Command:</span> <code>{service.command}</code></div>}
                {service.user_spec && <div className="flex"><span className="text-muted-foreground w-40">User:</span> {service.user_spec}</div>}
                <div className="flex"><span className="text-muted-foreground w-40">Enabled:</span> {service.enabled ? 'Yes' : 'No'}</div>
                <div className="flex items-center">
                  <span className="text-muted-foreground w-40">Health:</span>
                  <Badge variant={healthStatus === 'healthy' ? 'success' : healthStatus === 'unhealthy' ? 'destructive' : 'secondary'}>
                    {healthStatus}
                  </Badge>
                </div>
                {service.status?.container_id && (
                  <div className="flex items-center">
                    <span className="text-muted-foreground w-40">Container ID:</span>
                    <code className="font-mono text-xs">{service.status.container_id.slice(0, 12)}</code>
                    <CopyButton value={service.status.container_id} className="ml-1" />
                  </div>
                )}
                {service.status?.restart_count !== undefined && (
                  <div className="flex">
                    <span className="text-muted-foreground w-40">Restarts:</span>
                    <span className="flex items-center gap-1">
                      <RefreshCw className="size-3" /> {service.status.restart_count}
                    </span>
                  </div>
                )}
              </div>
            </CardContent>
          </Card>

          <EditableDomains name={service.name} domains={service.domains ?? []} />
          <EditablePorts name={service.name} ports={service.ports ?? []} />
          <EditableVolumes name={service.name} volumes={service.volumes ?? []} />
          <EditableDependencies name={service.name} dependencies={service.dependencies ?? []} />
          <EditableHealthcheck name={service.name} healthcheck={service.healthcheck ?? null} />
          <EditableLabels name={service.name} labels={service.labels ?? []} />
        </TabsContent>

        <TabsContent value="env">
          <EditableEnvVars name={service.name} envVars={service.env_vars ?? []} />
        </TabsContent>

        <TabsContent value="logs">
          <LogViewer serviceName={service.name} />
        </TabsContent>

        <TabsContent value="compose">
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <CardTitle className="text-base">Generated Compose YAML</CardTitle>
                <div className="flex items-center gap-2">
                  <Button size="sm" variant="outline" onClick={() => setComposeExpanded(!composeExpanded)} disabled={!composeYaml}>
                    {composeExpanded ? <Minimize2 className="size-4" /> : <Maximize2 className="size-4" />}
                    {composeExpanded ? 'Collapse' : 'Expand'}
                  </Button>
                  <Button size="sm" variant="outline" disabled={!composeYaml} onClick={() => {
                    if (!composeYaml) return
                    const blob = new Blob([composeYaml], { type: 'text/yaml' })
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
              {composeLoading ? (
                <div className="flex items-center justify-center py-8">
                  <Loader2 className="size-6 animate-spin text-muted-foreground" />
                </div>
              ) : composeYaml ? (
                <CodeEditor value={composeYaml} onChange={() => {}} language="yaml" readOnly autoHeight={composeExpanded} />
              ) : (
                <p className="text-muted-foreground">Compose YAML not available</p>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="files">
          <ConfigFilesPanel serviceName={service.name} />
        </TabsContent>
      </Tabs>

      <ConfirmDialog
        open={deleteOpen}
        onOpenChange={setDeleteOpen}
        title={`Delete ${service.name}?`}
        description="This will permanently delete the service and all its configuration. This action cannot be undone."
        confirmLabel="Delete"
        onConfirm={handleDelete}
        variant="destructive"
      />

      <Dialog open={editOpen} onOpenChange={setEditOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Edit {service.name}</DialogTitle>
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
            <Button onClick={handleEditSave} disabled={updateService.isPending}>
              {updateService.isPending ? <Loader2 className="size-4 animate-spin" /> : 'Save'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
