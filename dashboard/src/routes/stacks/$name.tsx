import { useState } from 'react'
import { createFileRoute, Link, Outlet, useMatch, useNavigate } from '@tanstack/react-router'
import { ArrowLeft, Loader2, Layers, Power, PowerOff, MoreVertical, Copy, Edit, Trash2, FileEdit, Plus, Globe, Download, AlertTriangle, Maximize2, Minimize2, Play, Square, RotateCcw } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { useStack, useEnableStack, useDisableStack, useStopStack, useStartStack, useRestartStack, useStackNetwork, useStackCompose, useGeneratePlan, useApplyPlan } from '@/features/stacks/queries'
import { CodeEditor } from '@/components/services/code-editor'
import { useInstances, useUpdateInstance, useStopInstance, useStartInstance, useRestartInstance } from '@/features/instances/queries'
import { EditStackDialog } from '@/components/stacks/edit-stack-dialog'
import { DeleteStackDialog } from '@/components/stacks/delete-stack-dialog'
import { CloneStackDialog } from '@/components/stacks/clone-stack-dialog'
import { RenameStackDialog } from '@/components/stacks/rename-stack-dialog'
import { DisableStackDialog } from '@/components/stacks/disable-stack-dialog'
import { AddInstanceDialog } from '@/components/stacks/add-instance-dialog'
import {
  DeleteInstanceDialog,
  DuplicateInstanceDialog,
} from '@/components/instances/instance-actions'
import type { Instance, StackPlan } from '@/types/api'
import { cn } from '@/lib/utils'

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
  onDelete: (id: string) => void
  onDuplicate: (id: string) => void
}

function InstanceCard({ instance, stackName, onDelete, onDuplicate }: InstanceCardProps) {
  const updateInstance = useUpdateInstance(stackName, instance.instance_id)
  const stopInstance = useStopInstance(stackName, instance.instance_id)
  const startInstance = useStartInstance(stackName, instance.instance_id)
  const restartInstance = useRestartInstance(stackName, instance.instance_id)
  const statusColor = instance.enabled ? 'bg-green-500' : 'bg-muted-foreground'

  const handleToggle = (e: React.MouseEvent) => {
    e.preventDefault()
    updateInstance.mutate({ enabled: !instance.enabled })
  }

  return (
    <Card className="py-3 hover:border-primary/50 transition-colors h-full group relative">
      <Link to="/stacks/$name/instances/$instance" params={{ name: stackName, instance: instance.instance_id }} className="block">
        <CardHeader className="pb-2">
          <div className="flex items-start gap-2">
            <div className={cn('size-2 rounded-full mt-1.5 shrink-0', statusColor)} />
            <div className="flex-1 min-w-0">
              <CardTitle className="text-sm font-semibold truncate">
                {instance.instance_id}
              </CardTitle>
              <p className="text-xs text-muted-foreground truncate">
                {instance.template_name}
              </p>
            </div>
          </div>
        </CardHeader>
        <CardContent className="space-y-2">
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
          {instance.override_count > 0 && (
            <Badge variant="secondary" className="text-xs">
              {instance.override_count} {instance.override_count === 1 ? 'override' : 'overrides'}
            </Badge>
          )}
        </CardContent>
      </Link>
      <div className="absolute top-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity">
        <DropdownMenu>
          <DropdownMenuTrigger asChild onClick={(e) => e.preventDefault()}>
            <Button variant="ghost" size="sm" className="size-6 p-0">
              <MoreVertical className="size-3" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" onClick={(e) => e.stopPropagation()}>
            <DropdownMenuItem onClick={(e) => { e.preventDefault(); stopInstance.mutate() }}>
              <Square className="size-4" />
              Stop
            </DropdownMenuItem>
            <DropdownMenuItem onClick={(e) => { e.preventDefault(); startInstance.mutate() }}>
              <Play className="size-4" />
              Start
            </DropdownMenuItem>
            <DropdownMenuItem onClick={(e) => { e.preventDefault(); restartInstance.mutate() }}>
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
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </Card>
  )
}

export const Route = createFileRoute('/stacks/$name')({
  component: StackDetailLayout,
})

function StackDetailLayout() {
  const childMatch = useMatch({ from: '/stacks/$name/instances/$instance', shouldThrow: false })
  if (childMatch) return <Outlet />
  return <StackDetailPage />
}

function StackDetailPage() {
  const { name } = Route.useParams()
  const navigate = useNavigate()
  const { data: stack, isLoading } = useStack(name)
  const { data: instances = [] } = useInstances(name)
  const { data: networkStatus } = useStackNetwork(name)
  const { data: composeData, isLoading: composeLoading } = useStackCompose(name)
  const enableStack = useEnableStack()
  const disableStack = useDisableStack()
  const stopStack = useStopStack()
  const startStack = useStartStack()
  const restartStack = useRestartStack()
  const generatePlan = useGeneratePlan(name)
  const applyPlan = useApplyPlan(name)

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
    if (!stack) return
    if (stack.enabled) {
      setDisableDialogOpen(true)
    } else {
      enableStack.mutate(stack.name)
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
    if (!composeData?.yaml) return
    const blob = new Blob([composeData.yaml], { type: 'text/yaml' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `docker-compose-${name}.yml`
    a.click()
    URL.revokeObjectURL(url)
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="size-8 animate-spin text-muted-foreground" />
      </div>
    )
  }

  if (!stack) {
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

  const createdAgo = timeAgo(stack.created_at)
  const updatedAgo = timeAgo(stack.updated_at)

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Link to="/stacks" className="text-muted-foreground hover:text-foreground">
          <ArrowLeft className="size-5" />
        </Link>
        <div className="flex-1">
          <div className="flex items-center gap-3">
            <h1 className="text-2xl font-bold">{stack.name}</h1>
            <Badge variant={stack.enabled ? 'success' : 'outline'}>
              {stack.enabled ? 'Enabled' : 'Disabled'}
            </Badge>
          </div>
          {stack.description && (
            <p className="text-muted-foreground mt-1">{stack.description}</p>
          )}
        </div>
        <div className="flex items-center gap-2">
          {stack.enabled && stack.running_count > 0 && (
            <Button
              variant="outline"
              size="sm"
              onClick={() => stopStack.mutate(stack.name)}
              disabled={stopStack.isPending}
            >
              {stopStack.isPending ? <Loader2 className="size-4 animate-spin" /> : <Square className="size-4" />}
              Stop
            </Button>
          )}
          {stack.enabled && stack.running_count === 0 && (
            <Button
              variant="outline"
              size="sm"
              onClick={() => startStack.mutate(stack.name)}
              disabled={startStack.isPending}
            >
              {startStack.isPending ? <Loader2 className="size-4 animate-spin" /> : <Play className="size-4" />}
              Start
            </Button>
          )}
          {stack.enabled && stack.running_count > 0 && (
            <Button
              variant="outline"
              size="sm"
              onClick={() => restartStack.mutate(stack.name)}
              disabled={restartStack.isPending}
            >
              {restartStack.isPending ? <Loader2 className="size-4 animate-spin" /> : <RotateCcw className="size-4" />}
              Restart
            </Button>
          )}
          <Button
            variant={stack.enabled ? 'outline' : 'success'}
            size="sm"
            onClick={handleToggleEnabled}
            disabled={enableStack.isPending || disableStack.isPending}
          >
            {enableStack.isPending || disableStack.isPending ? (
              <Loader2 className="size-4 animate-spin" />
            ) : stack.enabled ? (
              <>
                <PowerOff className="size-4" />
                Disable
              </>
            ) : (
              <>
                <Power className="size-4" />
                Enable
              </>
            )}
          </Button>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="outline" size="sm">
                <MoreVertical className="size-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
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
              <DropdownMenuItem variant="destructive" onClick={() => setDeleteOpen(true)}>
                <Trash2 className="size-4" />
                Delete
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>

      <div className="grid gap-4 md:grid-cols-4">
        <Card className="py-4">
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Status</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-lg font-semibold">
              {stack.enabled ? 'Enabled' : 'Disabled'}
            </div>
          </CardContent>
        </Card>

        <Card className="py-4">
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Instances</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-lg font-semibold">{stack.instance_count}</div>
          </CardContent>
        </Card>

        <Card className="py-4">
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Running</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-lg font-semibold">
              {stack.running_count} / {stack.instance_count}
            </div>
          </CardContent>
        </Card>

        <Card className="py-4">
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Created</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-sm">{createdAgo}</div>
            <div className="text-xs text-muted-foreground">Updated {updatedAgo}</div>
          </CardContent>
        </Card>
      </div>

      <Tabs defaultValue="instances" className="space-y-4">
        <TabsList>
          <TabsTrigger value="instances">Instances ({instances.length})</TabsTrigger>
          <TabsTrigger value="compose">Compose</TabsTrigger>
          <TabsTrigger value="deploy">Deploy</TabsTrigger>
        </TabsList>

        <TabsContent value="instances" className="space-y-6">
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <CardTitle className="text-base">
                  Instances ({instances.length})
                </CardTitle>
                <Button size="sm" onClick={() => setAddInstanceOpen(true)}>
                  <Plus className="size-4" />
                  Add Instance
                </Button>
              </div>
            </CardHeader>
            <CardContent>
              {instances.length === 0 ? (
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
                  {instances.map((instance) => (
                    <InstanceCard
                      key={instance.id}
                      instance={instance}
                      stackName={name}
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
              <div className="flex items-center gap-2">
                <Globe className="size-4 text-blue-500" />
                <CardTitle className="text-base">Network</CardTitle>
              </div>
            </CardHeader>
            <CardContent className="space-y-3">
              {networkStatus ? (
                <>
                  <div className="flex items-center gap-2">
                    <span className="text-sm text-muted-foreground">Name:</span>
                    <code className="bg-muted px-2 py-1 rounded text-sm font-mono">{networkStatus.network_name}</code>
                  </div>
                  <div className="flex items-center gap-2">
                    <span className="text-sm text-muted-foreground">Status:</span>
                    <Badge variant={networkStatus.status === 'active' ? 'success' : 'outline'}>
                      {networkStatus.status === 'active' ? 'Active' : 'Not Created'}
                    </Badge>
                  </div>
                  {networkStatus.status === 'active' && (
                    <>
                      <div className="flex items-center gap-2">
                        <span className="text-sm text-muted-foreground">Driver:</span>
                        <span className="text-sm">{networkStatus.driver}</span>
                      </div>
                      {networkStatus.containers.length > 0 && (
                        <div>
                          <div className="text-sm text-muted-foreground mb-1">Connected containers:</div>
                          <div className="flex flex-wrap gap-1">
                            {networkStatus.containers.map((container) => (
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
              <div className="flex items-center justify-between">
                <CardTitle className="text-base">Generated Compose YAML</CardTitle>
                <div className="flex items-center gap-2">
                  <Button size="sm" variant="outline" onClick={() => setComposeExpanded(!composeExpanded)} disabled={!composeData?.yaml}>
                    {composeExpanded ? <Minimize2 className="size-4" /> : <Maximize2 className="size-4" />}
                    {composeExpanded ? 'Collapse' : 'Expand'}
                  </Button>
                  <Button size="sm" variant="outline" onClick={handleDownload} disabled={!composeData?.yaml}>
                    <Download className="size-4" />
                    Download
                  </Button>
                </div>
              </div>
            </CardHeader>
            <CardContent className="space-y-4">
              {composeLoading ? (
                <div className="flex items-center justify-center py-8">
                  <Loader2 className="size-6 animate-spin text-muted-foreground" />
                </div>
              ) : composeData?.yaml ? (
                <CodeEditor value={composeData.yaml} onChange={() => {}} language="yaml" readOnly autoHeight={composeExpanded} />
              ) : (
                <div className="text-center py-8 text-muted-foreground">
                  <p>No compose YAML available</p>
                  <p className="text-sm mt-1">Add instances to this stack to generate compose configuration</p>
                </div>
              )}

              {composeData?.warnings && composeData.warnings.length > 0 && (
                <div className="space-y-2">
                  <h4 className="text-sm font-medium flex items-center gap-2">
                    <AlertTriangle className="size-4 text-yellow-500" />
                    Warnings ({composeData.warnings.length})
                  </h4>
                  <div className="space-y-1">
                    {composeData.warnings.map((warning, i) => (
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

        <TabsContent value="deploy">
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <CardTitle className="text-base">Deploy Stack</CardTitle>
                <Button
                  size="sm"
                  onClick={() => generatePlan.mutate(undefined, { onSuccess: (data) => setCurrentPlan(data.data) })}
                  disabled={generatePlan.isPending || applyPlan.isPending}
                >
                  {generatePlan.isPending ? <Loader2 className="size-4 animate-spin" /> : <Play className="size-4" />}
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
              ) : currentPlan.changes.length === 0 ? (
                <div className="flex items-center gap-2 text-green-600 dark:text-green-400 bg-green-50 dark:bg-green-950/20 px-4 py-3 rounded-md">
                  <svg className="size-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                  </svg>
                  <span className="font-medium">Stack is up to date — no changes needed</span>
                </div>
              ) : (
                <>
                  <div className="text-sm font-medium">
                    {currentPlan.changes.length} change(s): {currentPlan.changes.filter(c => c.action === 'add').length} to add, {currentPlan.changes.filter(c => c.action === 'modify').length} to modify, {currentPlan.changes.filter(c => c.action === 'remove').length} to remove
                  </div>
                  <div className="space-y-2">
                    {currentPlan.changes.map((change, i) => (
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
                      onClick={() => applyPlan.mutate({ token: currentPlan.token }, { onSuccess: () => setCurrentPlan(null) })}
                      disabled={applyPlan.isPending}
                    >
                      {applyPlan.isPending ? <Loader2 className="size-4 animate-spin" /> : null}
                      Apply Changes
                    </Button>
                    <Button
                      variant="outline"
                      onClick={() => generatePlan.mutate(undefined, { onSuccess: (data) => setCurrentPlan(data.data) })}
                      disabled={generatePlan.isPending || applyPlan.isPending}
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

      {stack && (
        <>
          <EditStackDialog stack={stack} open={editOpen} onOpenChange={setEditOpen} />
          <DeleteStackDialog stack={stack} open={deleteOpen} onOpenChange={setDeleteOpen} onSuccess={handleDeleteSuccess} />
          <CloneStackDialog stack={stack} open={cloneOpen} onOpenChange={setCloneOpen} />
          <RenameStackDialog stack={stack} open={renameOpen} onOpenChange={setRenameOpen} />
          <DisableStackDialog stack={stack} open={disableDialogOpen} onOpenChange={setDisableDialogOpen} />
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
