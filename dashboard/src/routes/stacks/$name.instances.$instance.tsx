import { useState } from 'react'
import { createFileRoute, Link, useNavigate } from '@tanstack/react-router'
import { z } from 'zod'
import { ArrowLeft, Loader2, Edit, Terminal, Minimize2, Maximize2, AlertTriangle, X } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Tabs, TabsContent } from '@/components/ui/tabs'
import { ResponsiveTabsList } from '@/components/ui/responsive-tabs-list'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog'
import {
  DropdownMenuItem,
  DropdownMenuSeparator,
} from '@/components/ui/dropdown-menu'
import { EnableToggle, LifecycleButtons, MoreActionsMenu } from '@/components/ui/entity-actions'
import { useInstanceDetailController } from '@/features/instances/useInstanceDetailController'
import { OverridePorts } from '@/components/instances/override-ports'
import { OverrideVolumes } from '@/components/instances/override-volumes'
import { OverrideEnvFiles } from '@/components/instances/override-env-files'
import { OverrideEnvVars } from '@/components/instances/override-env-vars'
import { OverrideNetworks } from '@/components/instances/override-networks'
import { OverrideLabels } from '@/components/instances/override-labels'
import { OverrideDomains } from '@/components/instances/override-domains'
import { OverrideHealthcheck } from '@/components/instances/override-healthcheck'
import { OverrideDependencies } from '@/components/instances/override-dependencies'
import { OverrideConfigMounts } from '@/components/instances/override-config-mounts'
import { OverrideConfigFiles } from '@/components/instances/override-config-files'
import { EffectiveConfigTab } from '@/components/instances/effective-config-tab'
import { ResourceLimits } from '@/components/instances/resource-limits'
import {
  DeleteInstanceDialog,
  DuplicateInstanceDialog,
  RenameInstanceDialog,
} from '@/components/instances/instance-actions'
import { TerminalDialog } from '@/components/terminal/terminal-dialog'
import { InstanceLogViewer } from '@/components/instances/instance-log-viewer'
import { ProxyConfigPanel } from '@/components/proxy/proxy-config-panel'
import { CodeEditor } from '@/components/services/code-editor'
import { cn } from '@/lib/utils'
import { getErrorMessage } from '@/lib/api'

function timeAgo(dateStr: string): string {
  const seconds = Math.floor((Date.now() - new Date(dateStr).getTime()) / 1000)
  const intervals: [number, string][] = [
    [31536000, 'year'], [2592000, 'month'], [86400, 'day'],
    [3600, 'hour'], [60, 'minute'], [1, 'second'],
  ]
  for (const [secs, label] of intervals) {
    const count = Math.floor(seconds / secs)
    if (count >= 1) return `${count} ${label}${count > 1 ? 's' : ''} ago`
  }
  return 'just now'
}

export const Route = createFileRoute('/stacks/$name/instances/$instance')({
  validateSearch: z.object({
    instanceTab: z.enum(['info', 'ports', 'volumes', 'env-files', 'environment', 'networks', 'labels', 'domains', 'healthcheck', 'dependencies', 'config-mounts', 'files', 'resources', 'effective', 'logs', 'compose', 'proxy']).optional(),
  }),
  component: InstanceDetailPage,
})

const instanceTabs = ['info', 'ports', 'volumes', 'env-files', 'environment', 'networks', 'labels', 'domains', 'healthcheck', 'dependencies', 'config-mounts', 'files', 'resources', 'effective', 'logs', 'compose', 'proxy'] as const
type InstanceTab = (typeof instanceTabs)[number]

function InstanceDetailPage() {
  const { name: stackName, instance: instanceId } = Route.useParams()
  const routeSearch = Route.useSearch()
  const routeNavigate = Route.useNavigate()
  const navigate = useNavigate()
  const ctrl = useInstanceDetailController(stackName, instanceId)
  const isRunning = ctrl.instance?.running ?? false

  const [editOpen, setEditOpen] = useState(false)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [duplicateOpen, setDuplicateOpen] = useState(false)
  const [renameOpen, setRenameOpen] = useState(false)
  const [terminalOpen, setTerminalOpen] = useState(false)
  const [editDescription, setEditDescription] = useState('')
  const [composeExpanded, setComposeExpanded] = useState(false)

  const openEdit = () => {
    if (!ctrl.instance) return
    setEditDescription(ctrl.instance.description ?? '')
    setEditOpen(true)
  }

  const handleToggleEnabled = () => {
    if (!ctrl.instance) return
    ctrl.updateInstance.mutate({ enabled: !ctrl.instance.enabled })
  }

  const handleEditSave = () => {
    ctrl.updateInstance.mutate({ description: editDescription }, {
      onSuccess: () => setEditOpen(false),
    })
  }

  const handleDeleteSuccess = () => {
    navigate({ to: '/stacks/$name', params: { name: stackName } })
  }

  const handleDuplicateSuccess = () => {
    navigate({ to: '/stacks/$name', params: { name: stackName } })
  }

  const handleRenameSuccess = (newName: string) => {
    navigate({ to: '/stacks/$name/instances/$instance', params: { name: stackName, instance: newName } })
  }

  if (ctrl.isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="size-8 animate-spin text-muted-foreground" />
      </div>
    )
  }

  if (!ctrl.instance) {
    return (
      <div className="space-y-4">
        <Link to="/stacks/$name" params={{ name: stackName }} className="inline-flex items-center gap-2 text-muted-foreground hover:text-foreground">
          <ArrowLeft className="size-4" />
          Back to Stack
        </Link>
        <div className="text-center py-12">
          <p className="text-muted-foreground">Instance not found</p>
        </div>
      </div>
    )
  }

  const createdAgo = timeAgo(ctrl.instance.created_at)
  const updatedAgo = timeAgo(ctrl.instance.updated_at)
  const activeTab: InstanceTab = routeSearch.instanceTab && instanceTabs.includes(routeSearch.instanceTab as InstanceTab)
    ? (routeSearch.instanceTab as InstanceTab)
    : 'info'
  const instanceTabItems = [
    { value: 'info', label: 'Info' },
    { value: 'ports', label: 'Ports' },
    { value: 'volumes', label: 'Volumes' },
    { value: 'env-files', label: 'Env Files' },
    { value: 'environment', label: 'Environment' },
    { value: 'networks', label: 'Networks' },
    { value: 'labels', label: 'Labels' },
    { value: 'domains', label: 'Domains' },
    { value: 'healthcheck', label: 'Healthcheck' },
    { value: 'dependencies', label: 'Dependencies' },
    { value: 'config-mounts', label: 'Config Mounts' },
    { value: 'files', label: 'Config Files' },
    { value: 'resources', label: 'Resources' },
    { value: 'effective', label: 'Effective Config' },
    { value: 'logs', label: 'Logs' },
    { value: 'compose', label: 'Compose' },
    { value: 'proxy', label: 'Proxy' },
  ]

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:gap-4">
        <Link to="/stacks/$name" params={{ name: stackName }} className="text-muted-foreground hover:text-foreground">
          <ArrowLeft className="size-5" />
        </Link>
        <div className="flex-1">
          <div className="text-xs text-muted-foreground mb-1">
            Stacks &gt; {stackName} &gt; {instanceId}
          </div>
          <div className="flex items-center gap-3">
            <h1 className="text-2xl font-bold">{ctrl.instance.instance_id}</h1>
            <div className={cn('size-2 rounded-full', isRunning ? 'bg-green-500' : ctrl.instance.enabled ? 'bg-yellow-500' : 'bg-muted-foreground')} />
          </div>
          <p className="text-muted-foreground">{ctrl.instance.template_name}</p>
          {ctrl.instance.container_name && (
            <p className="text-xs font-mono text-muted-foreground">{ctrl.instance.container_name}</p>
          )}
        </div>
        <div className="grid w-full grid-cols-2 gap-2 sm:flex sm:w-auto sm:items-center">
          {ctrl.instance.enabled && (
            <LifecycleButtons
              isRunning={isRunning}
              onStart={() => ctrl.startInstance.mutate()}
              onStop={() => ctrl.stopInstance.mutate()}
              onRestart={() => ctrl.restartInstance.mutate()}
              isStartPending={ctrl.startInstance.isPending}
              isStopPending={ctrl.stopInstance.isPending}
              isRestartPending={ctrl.restartInstance.isPending}
              showRestart
              className="col-span-2"
              buttonClassName="w-full sm:w-auto"
            />
          )}
          <EnableToggle
            enabled={ctrl.instance.enabled}
            onToggle={handleToggleEnabled}
            isPending={ctrl.updateInstance.isPending}
            className="w-full sm:w-auto"
          />
          <MoreActionsMenu triggerClassName="w-full sm:w-auto" mobileLabel="Actions">
            <DropdownMenuItem onClick={openEdit}>
              <Edit className="size-4" />
              Edit Description
            </DropdownMenuItem>
            {isRunning && ctrl.instance.container_name && (
              <DropdownMenuItem onClick={() => setTerminalOpen(true)}>
                <Terminal className="size-4" />
                Terminal
              </DropdownMenuItem>
            )}
            <DropdownMenuItem onClick={() => setDuplicateOpen(true)}>
              Duplicate
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => setRenameOpen(true)}>
              Rename
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem variant="destructive" onClick={() => setDeleteOpen(true)}>
              Delete
            </DropdownMenuItem>
          </MoreActionsMenu>
        </div>
      </div>

      {(() => {
        const errors = [
          ctrl.startInstance.isError && { action: 'Start', error: ctrl.startInstance.error, reset: ctrl.startInstance.reset },
          ctrl.stopInstance.isError && { action: 'Stop', error: ctrl.stopInstance.error, reset: ctrl.stopInstance.reset },
          ctrl.restartInstance.isError && { action: 'Restart', error: ctrl.restartInstance.error, reset: ctrl.restartInstance.reset },
          ctrl.updateInstance.isError && { action: 'Update', error: ctrl.updateInstance.error, reset: ctrl.updateInstance.reset },
        ].filter((e): e is { action: string; error: unknown; reset: () => void } => !!e)
        if (errors.length === 0) return null
        return (
          <div className="space-y-2">
            {errors.map((item) => (
              <div key={item.action} className="flex items-start gap-2 bg-red-50 dark:bg-red-950/20 text-red-600 dark:text-red-400 px-4 py-3 rounded-md">
                <AlertTriangle className="size-4 mt-0.5 shrink-0" />
                <div className="flex-1 text-sm">
                  <span className="font-medium">{item.action} failed</span>
                  <span className="opacity-80"> — {getErrorMessage(item.error, 'Unknown error')}</span>
                </div>
                <button onClick={() => item.reset()} className="shrink-0 opacity-60 hover:opacity-100 transition-opacity">
                  <X className="size-4" />
                </button>
              </div>
            ))}
          </div>
        )
      })()}

      <Tabs
        value={activeTab}
        onValueChange={(tab) => {
          if (!instanceTabs.includes(tab as InstanceTab)) return
          routeNavigate({ search: (prev) => ({ ...prev, instanceTab: tab as InstanceTab }) })
        }}
        className="space-y-4"
      >
        <ResponsiveTabsList
          tabs={instanceTabItems}
          value={activeTab}
          onValueChange={(tab) => {
            if (!instanceTabs.includes(tab as InstanceTab)) return
            routeNavigate({ search: (prev) => ({ ...prev, instanceTab: tab as InstanceTab }) })
          }}
        />

        <TabsContent value="info" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Details</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid gap-2 text-sm">
                <div className="flex"><span className="text-muted-foreground w-40">Instance Name:</span> {ctrl.instance.instance_id}</div>
                <div className="flex"><span className="text-muted-foreground w-40">Template:</span> {ctrl.instance.template_name}</div>
                <div className="flex"><span className="text-muted-foreground w-40">Container Name:</span> <code>{ctrl.instance.container_name ?? 'not set'}</code></div>
                <div className="flex"><span className="text-muted-foreground w-40">Enabled:</span> {ctrl.instance.enabled ? 'Yes' : 'No'}</div>
                <div className="flex"><span className="text-muted-foreground w-40">Overrides:</span> {ctrl.instance.override_count}</div>
                <div className="flex"><span className="text-muted-foreground w-40">Created:</span> {createdAgo}</div>
                <div className="flex"><span className="text-muted-foreground w-40">Updated:</span> {updatedAgo}</div>
              </div>
            </CardContent>
          </Card>

          {ctrl.instance.description && (
            <Card>
              <CardHeader>
                <CardTitle className="text-base">Description</CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-sm">{ctrl.instance.description}</p>
              </CardContent>
            </Card>
          )}
        </TabsContent>

        <TabsContent value="ports">
          {ctrl.templateService && (
            <OverridePorts
              instance={ctrl.instance}
              templateData={ctrl.templateService}
              stackName={stackName}
              instanceId={instanceId}
            />
          )}
        </TabsContent>

        <TabsContent value="volumes">
          {ctrl.templateService && (
            <OverrideVolumes
              instance={ctrl.instance}
              templateData={ctrl.templateService}
              stackName={stackName}
              instanceId={instanceId}
            />
          )}
        </TabsContent>

        <TabsContent value="env-files">
          {ctrl.templateService && (
            <OverrideEnvFiles
              instance={ctrl.instance}
              templateData={ctrl.templateService}
              stackName={stackName}
              instanceId={instanceId}
            />
          )}
        </TabsContent>

        <TabsContent value="environment">
          {ctrl.templateService && (
            <OverrideEnvVars
              instance={ctrl.instance}
              templateData={ctrl.templateService}
              stackName={stackName}
              instanceId={instanceId}
            />
          )}
        </TabsContent>

        <TabsContent value="networks">
          {ctrl.templateService && (
            <OverrideNetworks
              instance={ctrl.instance}
              templateData={ctrl.templateService}
              stackName={stackName}
              instanceId={instanceId}
            />
          )}
        </TabsContent>

        <TabsContent value="labels">
          {ctrl.templateService && (
            <OverrideLabels
              instance={ctrl.instance}
              templateData={ctrl.templateService}
              stackName={stackName}
              instanceId={instanceId}
            />
          )}
        </TabsContent>

        <TabsContent value="domains">
          {ctrl.templateService && (
            <OverrideDomains
              instance={ctrl.instance}
              templateData={ctrl.templateService}
              stackName={stackName}
              instanceId={instanceId}
            />
          )}
        </TabsContent>

        <TabsContent value="healthcheck">
          {ctrl.templateService && (
            <OverrideHealthcheck
              instance={ctrl.instance}
              templateData={ctrl.templateService}
              stackName={stackName}
              instanceId={instanceId}
            />
          )}
        </TabsContent>

        <TabsContent value="dependencies">
          {ctrl.templateService && (
            <OverrideDependencies
              instance={ctrl.instance}
              templateData={ctrl.templateService}
              stackName={stackName}
              instanceId={instanceId}
            />
          )}
        </TabsContent>

        <TabsContent value="config-mounts">
          {ctrl.templateService && (
            <OverrideConfigMounts
              instance={ctrl.instance}
              templateData={ctrl.templateService}
              stackName={stackName}
              instanceId={instanceId}
            />
          )}
        </TabsContent>

        <TabsContent value="files">
          {ctrl.templateService && (
            <OverrideConfigFiles
              instance={ctrl.instance}
              templateData={ctrl.templateService}
              stackName={stackName}
              instanceId={instanceId}
            />
          )}
        </TabsContent>

        <TabsContent value="resources">
          <ResourceLimits stackName={stackName} instanceId={instanceId} />
        </TabsContent>

        <TabsContent value="effective">
          <EffectiveConfigTab stackName={stackName} instanceId={instanceId} />
        </TabsContent>

        <TabsContent value="logs">
          <InstanceLogViewer stackName={stackName} instanceId={instanceId} />
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
                    a.download = `${instanceId}-compose.yaml`
                    a.click()
                    URL.revokeObjectURL(url)
                  }}>
                    Download
                  </Button>
                </div>
              </div>
            </CardHeader>
            <CardContent>
              {!ctrl.composeYaml ? (
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

        <TabsContent value="proxy">
          <ProxyConfigPanel scope="instance" name={instanceId} generateMutation={ctrl.generateProxyConfig} />
        </TabsContent>
      </Tabs>

      <Dialog open={editOpen} onOpenChange={setEditOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Edit Description</DialogTitle>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <label className="text-sm font-medium">Description</label>
              <Input value={editDescription} onChange={(e) => setEditDescription(e.target.value)} placeholder="Optional description" />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setEditOpen(false)}>Cancel</Button>
            <Button onClick={handleEditSave} disabled={ctrl.updateInstance.isPending}>
              {ctrl.updateInstance.isPending ? <Loader2 className="size-4 animate-spin" /> : 'Save'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <DuplicateInstanceDialog
        stackName={stackName}
        instanceId={instanceId}
        open={duplicateOpen}
        onOpenChange={setDuplicateOpen}
        onSuccess={handleDuplicateSuccess}
      />

      <RenameInstanceDialog
        stackName={stackName}
        instanceId={instanceId}
        open={renameOpen}
        onOpenChange={setRenameOpen}
        onSuccess={handleRenameSuccess}
      />

      <DeleteInstanceDialog
        stackName={stackName}
        instanceId={instanceId}
        open={deleteOpen}
        onOpenChange={setDeleteOpen}
        onSuccess={handleDeleteSuccess}
      />

      {ctrl.instance.container_name && (
        <TerminalDialog containerName={ctrl.instance.container_name} open={terminalOpen} onOpenChange={setTerminalOpen} />
      )}
    </div>
  )
}
