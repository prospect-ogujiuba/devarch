import { useState, useRef } from 'react'
import { createFileRoute, Link, Outlet, useMatch, useNavigate } from '@tanstack/react-router'
import { z } from 'zod'
import { ArrowLeft, Loader2, Layers, Power, Copy, Edit, Trash2, FileEdit, Plus, Globe, Download, AlertTriangle, Maximize2, Minimize2, Play, Square, RotateCcw, Network, Unplug, Clock3, Upload, PowerOff } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Tabs, TabsContent } from '@/components/ui/tabs'
import { ResponsiveTabsList } from '@/components/ui/responsive-tabs-list'
import {
  DropdownMenuItem,
  DropdownMenuSeparator,
} from '@/components/ui/dropdown-menu'
import { LifecycleButtons, EnableToggle, MoreActionsMenu } from '@/components/ui/entity-actions'
import { useStackDetailController } from '@/features/stacks/useStackDetailController'
import { ProxyConfigPanel } from '@/components/proxy/proxy-config-panel'
import { CodeEditor } from '@/components/services/code-editor'
import { useUpdateInstance, useStopInstance, useStartInstance, useRestartInstance } from '@/features/instances/queries'
import { EditStackDialog } from '@/components/stacks/edit-stack-dialog'
import { DeleteStackDialog } from '@/components/stacks/delete-stack-dialog'
import { CloneStackDialog } from '@/components/stacks/clone-stack-dialog'
import { RenameStackDialog } from '@/components/stacks/rename-stack-dialog'
import { DisableStackDialog } from '@/components/stacks/disable-stack-dialog'
import { AddInstanceDialog } from '@/components/stacks/add-instance-dialog'
import { WiringTab } from '@/components/stacks/wiring-tab'
import {
  DeleteInstanceDialog,
  DuplicateInstanceDialog,
} from '@/components/instances/instance-actions'
import type { Instance, StackPlan } from '@/types/api'
import { cn } from '@/lib/utils'
import { StatCard } from '@/components/ui/stat-card'

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

interface InstanceCardProps {
  instance: Instance
  stackName: string
  runningContainerNames: Set<string>
  onDelete: (id: string) => void
  onDuplicate: (id: string) => void
}

function InstanceCard({ instance, stackName, runningContainerNames, onDelete, onDuplicate }: InstanceCardProps) {
  const updateInstance = useUpdateInstance(stackName, instance.instance_id)
  const stopInstance = useStopInstance(stackName, instance.instance_id)
  const startInstance = useStartInstance(stackName, instance.instance_id)
  const restartInstance = useRestartInstance(stackName, instance.instance_id)
  const isRunning = Boolean(instance.container_name && runningContainerNames.has(instance.container_name))
  const canStart = instance.enabled && !isRunning && !startInstance.isPending
  const canStop = isRunning && !stopInstance.isPending
  const canRestart = isRunning && !restartInstance.isPending

  const handleToggle = (e: React.MouseEvent) => {
    e.preventDefault()
    updateInstance.mutate({ enabled: !instance.enabled })
  }

  return (
    <Card className="py-3 hover:border-primary/50 transition-colors h-full group relative">
      <Link to="/stacks/$name/instances/$instance" params={{ name: stackName, instance: instance.instance_id }} className="block">
        <CardHeader className="pb-2 px-4">
          <CardTitle className="text-sm font-semibold truncate">
            {instance.instance_id}
          </CardTitle>
          <p className="text-xs text-muted-foreground truncate">
            {instance.template_name}
          </p>
        </CardHeader>
        <CardContent className="space-y-2 px-4">
          {instance.container_name && (
            <div className="text-xs font-mono text-muted-foreground truncate">
              {instance.container_name}
            </div>
          )}
          {instance.description && (
            <p className="text-xs text-muted-foreground line-clamp-2">
              {instance.description}
            </p>
          )}
          <div className="flex items-center justify-between">
            {instance.override_count > 0 ? (
              <Badge variant="secondary" className="text-xs">
                {instance.override_count} {instance.override_count === 1 ? 'override' : 'overrides'}
              </Badge>
            ) : (
              <span />
            )}
            <div className="flex items-center gap-2">
              <Power className={cn('size-3', instance.enabled ? 'text-green-500' : 'text-muted-foreground/40')} />
              <Play className={cn('size-3', isRunning ? 'text-green-500' : 'text-muted-foreground/40')} />
            </div>
          </div>
        </CardContent>
      </Link>
      <div className="absolute right-2 top-2 opacity-100 transition-opacity sm:opacity-0 sm:group-hover:opacity-100">
        <MoreActionsMenu
          variant="ghost"
          triggerClassName="size-6 p-0"
          iconClassName="size-3"
          triggerProps={{ onClick: (e) => e.preventDefault() }}
          contentProps={{ onClick: (e) => e.stopPropagation() }}
        >
          <DropdownMenuItem disabled={!canStop} onClick={(e) => { e.preventDefault(); if (canStop) stopInstance.mutate() }}>
            <Square className="size-4" />
            Stop
          </DropdownMenuItem>
          <DropdownMenuItem disabled={!canStart} onClick={(e) => { e.preventDefault(); if (canStart) startInstance.mutate() }}>
            <Play className="size-4" />
            Start
          </DropdownMenuItem>
          <DropdownMenuItem disabled={!canRestart} onClick={(e) => { e.preventDefault(); if (canRestart) restartInstance.mutate() }}>
            <RotateCcw className="size-4" />
            Restart
          </DropdownMenuItem>
          <DropdownMenuSeparator />
          <DropdownMenuItem onClick={handleToggle}>
            {instance.enabled ? <PowerOff className="size-4" /> : <Power className="size-4" />}
            {instance.enabled ? 'Disable' : 'Enable'}
          </DropdownMenuItem>
          <DropdownMenuItem onClick={(e) => {
            e.preventDefault()
            onDuplicate(instance.instance_id)
          }}>
            <Copy className="size-4" />
            Duplicate
          </DropdownMenuItem>
          <DropdownMenuSeparator />
          <DropdownMenuItem variant="destructive" onClick={(e) => {
            e.preventDefault()
            onDelete(instance.instance_id)
          }}>
            <Trash2 className="size-4" />
            Delete
          </DropdownMenuItem>
        </MoreActionsMenu>
      </div>
    </Card>
  )
}

export const Route = createFileRoute('/stacks/$name')({
  validateSearch: z.object({
    tab: z.enum(['instances', 'compose', 'wiring', 'deploy', 'proxy']).optional(),
  }),
  component: StackDetailLayout,
})

const stackTabs = ['instances', 'compose', 'wiring', 'deploy', 'proxy'] as const
type StackTab = (typeof stackTabs)[number]

function StackDetailLayout() {
  const childMatch = useMatch({ from: '/stacks/$name/instances/$instance', shouldThrow: false })
  if (childMatch) return <Outlet />
  return <StackDetailPage />
}

function StackDetailPage() {
  const { name } = Route.useParams()
  const routeSearch = Route.useSearch()
  const routeNavigate = Route.useNavigate()
  const navigate = useNavigate()
  const ctrl = useStackDetailController(name)

  const fileInputRef = useRef<HTMLInputElement>(null)
  const [editOpen, setEditOpen] = useState(false)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [cloneOpen, setCloneOpen] = useState(false)
  const [renameOpen, setRenameOpen] = useState(false)
  const [disableDialogOpen, setDisableDialogOpen] = useState(false)
  const [addInstanceOpen, setAddInstanceOpen] = useState(false)
  const [instanceDeleteOpen, setInstanceDeleteOpen] = useState(false)
  const [instanceDuplicateOpen, setInstanceDuplicateOpen] = useState(false)
  const [composeExpanded, setComposeExpanded] = useState(false)
  const [selectedInstanceId, setSelectedInstanceId] = useState<string | null>(null)
  const [currentPlan, setCurrentPlan] = useState<StackPlan | null>(null)

  const handleToggleEnabled = () => {
    if (!ctrl.stack) return
    if (ctrl.stack.enabled) {
      setDisableDialogOpen(true)
    } else {
      ctrl.enableStack.mutate(ctrl.stack.name)
    }
  }

  const handleDeleteSuccess = () => {
    navigate({ to: '/stacks' })
  }

  const openInstanceDelete = (instanceId: string) => {
    setSelectedInstanceId(instanceId)
    setInstanceDeleteOpen(true)
  }

  const openInstanceDuplicate = (instanceId: string) => {
    setSelectedInstanceId(instanceId)
    setInstanceDuplicateOpen(true)
  }

  const handleDownload = () => {
    if (!ctrl.composeData?.yaml) return
    const blob = new Blob([ctrl.composeData.yaml], { type: 'text/yaml' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `docker-compose-${name}.yml`
    a.click()
    URL.revokeObjectURL(url)
  }

  const handleImport = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (file) {
      ctrl.importStack.mutate(file)
      if (fileInputRef.current) {
        fileInputRef.current.value = ''
      }
    }
  }

  if (ctrl.isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="size-8 animate-spin text-muted-foreground" />
      </div>
    )
  }

  if (!ctrl.stack) {
    return (
      <div className="space-y-4">
        <Link to="/stacks" className="inline-flex items-center gap-2 text-muted-foreground hover:text-foreground">
          <ArrowLeft className="size-4" />
          Back to Stacks
        </Link>
        <div className="text-center py-12">
          <p className="text-muted-foreground">Stack not found</p>
        </div>
      </div>
    )
  }

  const updatedAgo = timeAgo(ctrl.stack.updated_at)
  const activeTab: StackTab = routeSearch.tab ?? 'instances'
  const stackTabItems = [
    { value: 'instances', label: `Instances (${ctrl.instances.length})` },
    { value: 'compose', label: 'Compose' },
    { value: 'wiring', label: 'Wiring' },
    { value: 'deploy', label: 'Deploy' },
    { value: 'proxy', label: 'Proxy' },
  ]

  return (
    <div className="space-y-5 sm:space-y-6">
      <div className="space-y-3">
        <Link to="/stacks" className="inline-flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground">
          <ArrowLeft className="size-4" />
          Back to Stacks
        </Link>

        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between sm:gap-4">
          <div className="min-w-0">
            <div className="flex flex-wrap items-center gap-2 sm:gap-3">
              <h1 className="truncate text-xl font-bold sm:text-2xl">{ctrl.stack.name}</h1>
              <Badge variant={ctrl.stack.enabled ? 'success' : 'outline'}>
                {ctrl.stack.enabled ? 'Enabled' : 'Disabled'}
              </Badge>
            </div>
            <p className="mt-1 text-sm text-muted-foreground sm:text-base">
              {ctrl.stack.description || `Manage ${ctrl.stack.instance_count} instance${ctrl.stack.instance_count === 1 ? '' : 's'} in this stack`}
            </p>
          </div>
          <div className="grid w-full grid-cols-2 gap-2 sm:flex sm:w-auto sm:flex-wrap sm:items-center sm:justify-end">
          {ctrl.stack.enabled && (
            <LifecycleButtons
              isRunning={ctrl.stack.running_count > 0}
              onStart={() => ctrl.startStack.mutate(ctrl.stack!.name)}
              onStop={() => ctrl.stopStack.mutate(ctrl.stack!.name)}
              onRestart={() => ctrl.restartStack.mutate(ctrl.stack!.name)}
              isStartPending={ctrl.startStack.isPending}
              isStopPending={ctrl.stopStack.isPending}
              isRestartPending={ctrl.restartStack.isPending}
              showRestart
              className="col-span-2"
              buttonClassName="w-full sm:w-auto"
            />
          )}
          <EnableToggle
            enabled={ctrl.stack.enabled}
            onToggle={handleToggleEnabled}
            isPending={ctrl.enableStack.isPending || ctrl.disableStack.isPending}
            className="w-full sm:w-auto"
          />
          <MoreActionsMenu triggerClassName="w-full sm:w-auto" mobileLabel="Actions">
            <DropdownMenuItem onClick={() => setEditOpen(true)}>
              <Edit className="size-4" />
              Edit Description
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => setCloneOpen(true)}>
              <Copy className="size-4" />
              Clone
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => setRenameOpen(true)}>
              <FileEdit className="size-4" />
              Rename
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem onClick={() => ctrl.exportStack.mutate()} disabled={ctrl.exportStack.isPending}>
              <Download className="size-4" />
              Export
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => fileInputRef.current?.click()} disabled={ctrl.importStack.isPending}>
              <Upload className="size-4" />
              Import
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem variant="destructive" onClick={() => setDeleteOpen(true)}>
              <Trash2 className="size-4" />
              Delete
            </DropdownMenuItem>
          </MoreActionsMenu>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-2 gap-3 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard
          icon={Power}
          label="Status"
          value={ctrl.stack.enabled ? 'Enabled' : 'Disabled'}
          color={ctrl.stack.enabled ? 'text-green-500' : 'text-muted-foreground'}
        />
        <StatCard icon={Layers} label="Instances" value={ctrl.stack.instance_count} />
        <StatCard
          icon={Play}
          label="Running"
          value={`${ctrl.stack.running_count} / ${ctrl.stack.instance_count}`}
          color={ctrl.stack.running_count > 0 ? 'text-green-500' : 'text-muted-foreground'}
        />
        <StatCard icon={Clock3} label="Updated" value={updatedAgo} />
      </div>

      <Tabs
        value={activeTab}
        onValueChange={(tab) => {
          if (!stackTabs.includes(tab as StackTab)) return
          routeNavigate({ search: (prev) => ({ ...prev, tab: tab as StackTab }) })
        }}
        className="space-y-4"
      >
        <ResponsiveTabsList
          tabs={stackTabItems}
          value={activeTab}
          onValueChange={(tab) => {
            if (!stackTabs.includes(tab as StackTab)) return
            routeNavigate({ search: (prev) => ({ ...prev, tab: tab as StackTab }) })
          }}
        />

        <TabsContent value="instances" className="space-y-6">
          <Card>
            <CardHeader>
              <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                <CardTitle className="text-base">
                  Instances ({ctrl.instances.length})
                </CardTitle>
                <Button size="sm" className="w-full sm:w-auto" onClick={() => setAddInstanceOpen(true)}>
                  <Plus className="size-4" />
                  Add Instance
                </Button>
              </div>
            </CardHeader>
            <CardContent>
              {ctrl.instances.length === 0 ? (
                <div className="flex flex-col items-center justify-center py-12 text-center">
                  <div className="rounded-full bg-muted p-3 mb-4">
                    <Layers className="size-8 text-muted-foreground" />
                  </div>
                  <h3 className="font-medium mb-1">Add your first service</h3>
                  <p className="text-sm text-muted-foreground mb-4">
                    Choose from available templates to get started
                  </p>
                  <Button onClick={() => setAddInstanceOpen(true)}>
                    <Plus className="size-4" />
                    Add Instance
                  </Button>
                </div>
              ) : (
                <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
                  {ctrl.instances.map((instance) => (
                    <InstanceCard
                      key={instance.id}
                      instance={instance}
                      stackName={name}
                      runningContainerNames={ctrl.runningContainerNames}
                      onDelete={openInstanceDelete}
                      onDuplicate={openInstanceDuplicate}
                    />
                  ))}
                </div>
              )}
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                <div className="flex items-center gap-2">
                  <Globe className="size-4 text-blue-500" />
                  <CardTitle className="text-base">Network</CardTitle>
                </div>
                {ctrl.networkStatus?.status === 'not_created' && (
                  <Button
                    size="sm"
                    variant="outline"
                    className="w-full sm:w-auto"
                    onClick={() => ctrl.createNetwork.mutate(name)}
                    disabled={ctrl.createNetwork.isPending}
                  >
                    {ctrl.createNetwork.isPending ? <Loader2 className="size-4 animate-spin" /> : <Network className="size-4" />}
                    Create Network
                  </Button>
                )}
                {ctrl.networkStatus?.status === 'active' && (
                  <Button
                    size="sm"
                    variant="outline"
                    className="w-full sm:w-auto"
                    onClick={() => ctrl.removeNetwork.mutate(name)}
                    disabled={ctrl.removeNetwork.isPending}
                  >
                    {ctrl.removeNetwork.isPending ? <Loader2 className="size-4 animate-spin" /> : <Unplug className="size-4" />}
                    Remove Network
                  </Button>
                )}
              </div>
            </CardHeader>
            <CardContent className="space-y-3">
              {ctrl.networkStatus ? (
                <>
                  <div className="flex items-center gap-2">
                    <span className="text-sm text-muted-foreground">Name:</span>
                    <code className="bg-muted px-2 py-1 rounded text-sm font-mono">{ctrl.networkStatus.network_name}</code>
                  </div>
                  <div className="flex items-center gap-2">
                    <span className="text-sm text-muted-foreground">Status:</span>
                    <Badge variant={ctrl.networkStatus.status === 'active' ? 'success' : 'outline'}>
                      {ctrl.networkStatus.status === 'active' ? 'Active' : 'Not Created'}
                    </Badge>
                  </div>
                  {ctrl.networkStatus.status === 'active' && (
                    <>
                      <div className="flex items-center gap-2">
                        <span className="text-sm text-muted-foreground">Driver:</span>
                        <span className="text-sm">{ctrl.networkStatus.driver}</span>
                      </div>
                      {ctrl.connectedContainers.length > 0 && (
                        <div>
                          <div className="text-sm text-muted-foreground mb-1">Connected containers:</div>
                          <div className="flex flex-wrap gap-1">
                            {ctrl.connectedContainers.map((container) => (
                              <code key={container} className="bg-muted px-2 py-0.5 rounded text-xs font-mono">
                                {container}
                              </code>
                            ))}
                          </div>
                        </div>
                      )}
                      <div className="text-xs text-muted-foreground pt-1 border-t">
                        Containers resolve each other by instance name within this network
                      </div>
                    </>
                  )}
                </>
              ) : (
                <div className="flex items-center gap-2">
                  <Loader2 className="size-4 animate-spin text-muted-foreground" />
                  <span className="text-sm text-muted-foreground">Loading network status...</span>
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="compose">
          <Card>
            <CardHeader>
              <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                <CardTitle className="text-base">Generated Compose YAML</CardTitle>
                <div className="flex w-full items-center gap-2 sm:w-auto">
                  <Button size="sm" variant="outline" className="flex-1 sm:flex-none" onClick={() => setComposeExpanded(!composeExpanded)} disabled={!ctrl.composeData?.yaml}>
                    {composeExpanded ? <Minimize2 className="size-4" /> : <Maximize2 className="size-4" />}
                    {composeExpanded ? 'Collapse' : 'Expand'}
                  </Button>
                  <Button size="sm" variant="outline" className="flex-1 sm:flex-none" onClick={handleDownload} disabled={!ctrl.composeData?.yaml}>
                    <Download className="size-4" />
                    Download
                  </Button>
                </div>
              </div>
            </CardHeader>
            <CardContent className="space-y-4">
              {ctrl.composeLoading ? (
                <div className="flex items-center justify-center py-8">
                  <Loader2 className="size-6 animate-spin text-muted-foreground" />
                </div>
              ) : ctrl.composeData?.yaml ? (
                <CodeEditor value={ctrl.composeData.yaml} onChange={() => {}} language="yaml" readOnly autoHeight={composeExpanded} />
              ) : (
                <div className="text-center py-8 text-muted-foreground">
                  <p>No compose YAML available</p>
                  <p className="text-sm mt-1">Add instances to this stack to generate compose configuration</p>
                </div>
              )}

              {ctrl.composeData?.warnings && ctrl.composeData.warnings.length > 0 && (
                <div className="space-y-2">
                  <h4 className="text-sm font-medium flex items-center gap-2">
                    <AlertTriangle className="size-4 text-yellow-500" />
                    Warnings ({ctrl.composeData.warnings.length})
                  </h4>
                  <div className="space-y-1">
                    {ctrl.composeData.warnings.map((warning, i) => (
                      <div key={i} className="text-sm text-yellow-600 dark:text-yellow-400 bg-yellow-50 dark:bg-yellow-950/20 px-3 py-2 rounded-md">
                        {warning}
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="wiring">
          <WiringTab stackName={name} />
        </TabsContent>

        <TabsContent value="proxy">
          <ProxyConfigPanel scope="stack" name={name} generateMutation={ctrl.generateProxyConfig} />
        </TabsContent>

        <TabsContent value="deploy">
          <Card>
            <CardHeader>
              <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                <CardTitle className="text-base">Deploy Stack</CardTitle>
                <Button
                  size="sm"
                  className="w-full sm:w-auto"
                  onClick={() => ctrl.generatePlan.mutate(undefined, { onSuccess: (data) => setCurrentPlan(data) })}
                  disabled={ctrl.generatePlan.isPending || ctrl.applyPlan.isPending}
                >
                  {ctrl.generatePlan.isPending ? <Loader2 className="size-4 animate-spin" /> : <Play className="size-4" />}
                  Generate Plan
                </Button>
              </div>
            </CardHeader>
            <CardContent className="space-y-4">
              {!currentPlan ? (
                <div className="flex flex-col items-center justify-center py-12 text-center">
                  <div className="rounded-full bg-muted p-3 mb-4">
                    <Play className="size-8 text-muted-foreground" />
                  </div>
                  <h3 className="font-medium mb-1">Generate a plan to preview changes</h3>
                  <p className="text-sm text-muted-foreground">
                    Compare running state against desired configuration
                  </p>
                </div>
              ) : (currentPlan.changes ?? []).length === 0 ? (
                <div className="flex items-center gap-2 text-green-600 dark:text-green-400 bg-green-50 dark:bg-green-950/20 px-4 py-3 rounded-md">
                  <svg className="size-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                  </svg>
                  <span className="font-medium">Stack is up to date — no changes needed</span>
                </div>
              ) : (
                <>
                  <div className="text-sm font-medium">
                    {(currentPlan.changes ?? []).length} change(s): {(currentPlan.changes ?? []).filter(c => c.action === 'add').length} to add, {(currentPlan.changes ?? []).filter(c => c.action === 'modify').length} to modify, {(currentPlan.changes ?? []).filter(c => c.action === 'remove').length} to remove
                  </div>
                  <div className="space-y-2">
                    {(currentPlan.changes ?? []).map((change, i) => (
                      <div
                        key={i}
                        className={cn(
                          'border-l-4 px-4 py-3 rounded-r-md space-y-2',
                          change.action === 'add' && 'border-green-500 bg-green-50 dark:bg-green-950/20',
                          change.action === 'modify' && 'border-yellow-500 bg-yellow-50 dark:bg-yellow-950/20',
                          change.action === 'remove' && 'border-red-500 bg-red-50 dark:bg-red-950/20'
                        )}
                      >
                        <div className="flex items-start gap-2">
                          <span className={cn(
                            'text-lg font-bold mt-0.5',
                            change.action === 'add' && 'text-green-600 dark:text-green-400',
                            change.action === 'modify' && 'text-yellow-600 dark:text-yellow-400',
                            change.action === 'remove' && 'text-red-600 dark:text-red-400'
                          )}>
                            {change.action === 'add' ? '+' : change.action === 'modify' ? '~' : '-'}
                          </span>
                          <div className="flex-1 min-w-0">
                            <div className="font-medium">{change.instance_id}</div>
                            <div className="text-sm text-muted-foreground">
                              {change.template_name} {change.container_name && `→ ${change.container_name}`}
                            </div>
                            {change.action === 'modify' && change.fields && (
                              <div className="mt-2 space-y-1 text-sm">
                                {Object.entries(change.fields).map(([field, fieldChange]) => (
                                  <div key={field} className="font-mono text-xs">
                                    <span className="text-muted-foreground">{field}:</span>{' '}
                                    <span className="text-muted-foreground line-through">{String(fieldChange.old)}</span>
                                    {' → '}
                                    <span className="font-semibold">{String(fieldChange.new)}</span>
                                  </div>
                                ))}
                              </div>
                            )}
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                  <div className="flex items-center gap-2 pt-2">
                    <Button
                      onClick={() => ctrl.applyPlan.mutate({ token: currentPlan.token }, { onSuccess: () => setCurrentPlan(null) })}
                      disabled={ctrl.applyPlan.isPending}
                    >
                      {ctrl.applyPlan.isPending ? <Loader2 className="size-4 animate-spin" /> : null}
                      Apply Changes
                    </Button>
                    <Button
                      variant="outline"
                      onClick={() => ctrl.generatePlan.mutate(undefined, { onSuccess: (data) => setCurrentPlan(data) })}
                      disabled={ctrl.generatePlan.isPending || ctrl.applyPlan.isPending}
                    >
                      Regenerate Plan
                    </Button>
                  </div>
                </>
              )}

              {currentPlan?.warnings && currentPlan.warnings.length > 0 && (
                <div className="space-y-2">
                  <h4 className="text-sm font-medium flex items-center gap-2">
                    <AlertTriangle className="size-4 text-yellow-500" />
                    Warnings ({currentPlan.warnings.length})
                  </h4>
                  <div className="space-y-1">
                    {currentPlan.warnings.map((warning, i) => (
                      <div key={i} className="text-sm text-yellow-600 dark:text-yellow-400 bg-yellow-50 dark:bg-yellow-950/20 px-3 py-2 rounded-md">
                        {warning}
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      <input
        type="file"
        accept=".yml,.yaml"
        className="hidden"
        ref={fileInputRef}
        onChange={handleImport}
      />

      {ctrl.stack && (
        <>
          <EditStackDialog stack={ctrl.stack} open={editOpen} onOpenChange={setEditOpen} />
          <DeleteStackDialog stack={ctrl.stack} open={deleteOpen} onOpenChange={setDeleteOpen} onSuccess={handleDeleteSuccess} />
          <CloneStackDialog stack={ctrl.stack} open={cloneOpen} onOpenChange={setCloneOpen} />
          <RenameStackDialog stack={ctrl.stack} open={renameOpen} onOpenChange={setRenameOpen} />
          <DisableStackDialog stack={ctrl.stack} open={disableDialogOpen} onOpenChange={setDisableDialogOpen} />
          <AddInstanceDialog stackName={name} open={addInstanceOpen} onOpenChange={setAddInstanceOpen} />
        </>
      )}

      {selectedInstanceId && (
        <>
          <DeleteInstanceDialog
            stackName={name}
            instanceId={selectedInstanceId}
            open={instanceDeleteOpen}
            onOpenChange={setInstanceDeleteOpen}
          />
          <DuplicateInstanceDialog
            stackName={name}
            instanceId={selectedInstanceId}
            open={instanceDuplicateOpen}
            onOpenChange={setInstanceDuplicateOpen}
          />
        </>
      )}
    </div>
  )
}
