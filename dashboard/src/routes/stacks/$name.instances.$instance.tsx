import { useState } from 'react'
import { createFileRoute, Link, useNavigate } from '@tanstack/react-router'
import { ArrowLeft, Loader2, Power, PowerOff, Copy, Edit, Trash2, MoreVertical } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { ConfirmDialog } from '@/components/ui/confirm-dialog'
import {
  useInstance,
  useUpdateInstance,
  useDeleteInstance,
  useDuplicateInstance,
  useRenameInstance,
  useInstanceDeletePreview,
} from '@/features/instances/queries'
import { useService } from '@/features/services/queries'
import { OverridePorts } from '@/components/instances/override-ports'
import { OverrideVolumes } from '@/components/instances/override-volumes'
import { OverrideEnvVars } from '@/components/instances/override-env-vars'
import { OverrideLabels } from '@/components/instances/override-labels'
import { OverrideDomains } from '@/components/instances/override-domains'
import { OverrideHealthcheck } from '@/components/instances/override-healthcheck'
import { OverrideConfigFiles } from '@/components/instances/override-config-files'
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

export const Route = createFileRoute('/stacks/$name/instances/$instance')({
  component: InstanceDetailPage,
})

function InstanceDetailPage() {
  const { name: stackName, instance: instanceId } = Route.useParams()
  const navigate = useNavigate()
  const { data: instance, isLoading } = useInstance(stackName, instanceId)
  const { data: templateService } = useService(instance?.template_name ?? '')
  const updateInstance = useUpdateInstance(stackName, instanceId)
  const deleteInstance = useDeleteInstance(stackName, instanceId)
  const duplicateInstance = useDuplicateInstance(stackName, instanceId)
  const renameInstance = useRenameInstance(stackName, instanceId)
  const { data: deletePreview } = useInstanceDeletePreview(stackName, instanceId)

  const [editOpen, setEditOpen] = useState(false)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [duplicateOpen, setDuplicateOpen] = useState(false)
  const [renameOpen, setRenameOpen] = useState(false)
  const [editDescription, setEditDescription] = useState('')
  const [duplicateName, setDuplicateName] = useState('')
  const [renameName, setRenameName] = useState('')

  const openEdit = () => {
    if (!instance) return
    setEditDescription(instance.description ?? '')
    setEditOpen(true)
  }

  const openDuplicate = () => {
    if (!instance) return
    setDuplicateName(`${instanceId}-copy`)
    setDuplicateOpen(true)
  }

  const openRename = () => {
    if (!instance) return
    setRenameName(instanceId)
    setRenameOpen(true)
  }

  const handleToggleEnabled = () => {
    if (!instance) return
    updateInstance.mutate({ enabled: !instance.enabled })
  }

  const handleEditSave = () => {
    updateInstance.mutate({ description: editDescription }, {
      onSuccess: () => setEditOpen(false),
    })
  }

  const handleDuplicate = () => {
    duplicateInstance.mutate(duplicateName, {
      onSuccess: () => {
        setDuplicateOpen(false)
        navigate({ to: '/stacks/$name', params: { name: stackName } })
      },
    })
  }

  const handleRename = () => {
    renameInstance.mutate({ instance_id: renameName }, {
      onSuccess: () => {
        setRenameOpen(false)
        navigate({ to: '/stacks/$name/instances/$instance', params: { name: stackName, instance: renameName } })
      },
    })
  }

  const handleDelete = () => {
    deleteInstance.mutate(undefined, {
      onSuccess: () => navigate({ to: '/stacks/$name', params: { name: stackName } }),
    })
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="size-8 animate-spin text-muted-foreground" />
      </div>
    )
  }

  if (!instance) {
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

  const createdAgo = timeAgo(instance.created_at)
  const updatedAgo = timeAgo(instance.updated_at)

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Link to="/stacks/$name" params={{ name: stackName }} className="text-muted-foreground hover:text-foreground">
          <ArrowLeft className="size-5" />
        </Link>
        <div className="flex-1">
          <div className="text-xs text-muted-foreground mb-1">
            Stacks &gt; {stackName} &gt; {instanceId}
          </div>
          <div className="flex items-center gap-3">
            <h1 className="text-2xl font-bold">{instance.instance_id}</h1>
            <div className={cn('size-2 rounded-full', instance.enabled ? 'bg-green-500' : 'bg-muted-foreground')} />
          </div>
          <p className="text-muted-foreground">{instance.template_name}</p>
          {instance.container_name && (
            <p className="text-xs font-mono text-muted-foreground">{instance.container_name}</p>
          )}
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant={instance.enabled ? 'outline' : 'success'}
            size="sm"
            onClick={handleToggleEnabled}
            disabled={updateInstance.isPending}
          >
            {updateInstance.isPending ? (
              <Loader2 className="size-4 animate-spin" />
            ) : instance.enabled ? (
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
              <DropdownMenuItem onClick={openEdit}>
                <Edit className="size-4" />
                Edit Description
              </DropdownMenuItem>
              <DropdownMenuItem onClick={openDuplicate}>
                <Copy className="size-4" />
                Duplicate
              </DropdownMenuItem>
              <DropdownMenuItem onClick={openRename}>
                <Edit className="size-4" />
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

      <Tabs defaultValue="info" className="space-y-4">
        <TabsList>
          <TabsTrigger value="info">Info</TabsTrigger>
          <TabsTrigger value="ports">Ports</TabsTrigger>
          <TabsTrigger value="volumes">Volumes</TabsTrigger>
          <TabsTrigger value="environment">Environment</TabsTrigger>
          <TabsTrigger value="labels">Labels</TabsTrigger>
          <TabsTrigger value="domains">Domains</TabsTrigger>
          <TabsTrigger value="healthcheck">Healthcheck</TabsTrigger>
          <TabsTrigger value="files">Config Files</TabsTrigger>
          <TabsTrigger value="effective">Effective Config</TabsTrigger>
        </TabsList>

        <TabsContent value="info" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Details</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid gap-2 text-sm">
                <div className="flex"><span className="text-muted-foreground w-40">Instance Name:</span> {instance.instance_id}</div>
                <div className="flex"><span className="text-muted-foreground w-40">Template:</span> {instance.template_name}</div>
                <div className="flex"><span className="text-muted-foreground w-40">Container Name:</span> <code>{instance.container_name ?? 'not set'}</code></div>
                <div className="flex"><span className="text-muted-foreground w-40">Enabled:</span> {instance.enabled ? 'Yes' : 'No'}</div>
                <div className="flex"><span className="text-muted-foreground w-40">Overrides:</span> {instance.override_count}</div>
                <div className="flex"><span className="text-muted-foreground w-40">Created:</span> {createdAgo}</div>
                <div className="flex"><span className="text-muted-foreground w-40">Updated:</span> {updatedAgo}</div>
              </div>
            </CardContent>
          </Card>

          {instance.description && (
            <Card>
              <CardHeader>
                <CardTitle className="text-base">Description</CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-sm">{instance.description}</p>
              </CardContent>
            </Card>
          )}
        </TabsContent>

        <TabsContent value="ports">
          {templateService && (
            <OverridePorts
              instance={instance}
              templateData={templateService}
              stackName={stackName}
              instanceId={instanceId}
            />
          )}
        </TabsContent>

        <TabsContent value="volumes">
          {templateService && (
            <OverrideVolumes
              instance={instance}
              templateData={templateService}
              stackName={stackName}
              instanceId={instanceId}
            />
          )}
        </TabsContent>

        <TabsContent value="environment">
          {templateService && (
            <OverrideEnvVars
              instance={instance}
              templateData={templateService}
              stackName={stackName}
              instanceId={instanceId}
            />
          )}
        </TabsContent>

        <TabsContent value="labels">
          {templateService && (
            <OverrideLabels
              instance={instance}
              templateData={templateService}
              stackName={stackName}
              instanceId={instanceId}
            />
          )}
        </TabsContent>

        <TabsContent value="domains">
          {templateService && (
            <OverrideDomains
              instance={instance}
              templateData={templateService}
              stackName={stackName}
              instanceId={instanceId}
            />
          )}
        </TabsContent>

        <TabsContent value="healthcheck">
          {templateService && (
            <OverrideHealthcheck
              instance={instance}
              templateData={templateService}
              stackName={stackName}
              instanceId={instanceId}
            />
          )}
        </TabsContent>

        <TabsContent value="files">
          {templateService && (
            <OverrideConfigFiles
              instance={instance}
              templateData={templateService}
              stackName={stackName}
              instanceId={instanceId}
            />
          )}
        </TabsContent>

        <TabsContent value="effective">
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Effective Configuration</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-muted-foreground">Coming in Plan 05</p>
            </CardContent>
          </Card>
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
            <Button onClick={handleEditSave} disabled={updateInstance.isPending}>
              {updateInstance.isPending ? <Loader2 className="size-4 animate-spin" /> : 'Save'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={duplicateOpen} onOpenChange={setDuplicateOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Duplicate Instance</DialogTitle>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <label className="text-sm font-medium">New Instance Name</label>
              <Input value={duplicateName} onChange={(e) => setDuplicateName(e.target.value)} />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDuplicateOpen(false)}>Cancel</Button>
            <Button onClick={handleDuplicate} disabled={duplicateInstance.isPending || !duplicateName}>
              {duplicateInstance.isPending ? <Loader2 className="size-4 animate-spin" /> : 'Duplicate'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={renameOpen} onOpenChange={setRenameOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Rename Instance</DialogTitle>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <label className="text-sm font-medium">Instance Name</label>
              <Input value={renameName} onChange={(e) => setRenameName(e.target.value)} />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setRenameOpen(false)}>Cancel</Button>
            <Button onClick={handleRename} disabled={renameInstance.isPending || !renameName || renameName === instanceId}>
              {renameInstance.isPending ? <Loader2 className="size-4 animate-spin" /> : 'Rename'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <ConfirmDialog
        open={deleteOpen}
        onOpenChange={setDeleteOpen}
        title={`Delete ${instanceId}?`}
        description={deletePreview ? `This will delete the instance and ${deletePreview.override_count} override(s). Container: ${deletePreview.container_name ?? 'none'}` : 'This will permanently delete the instance and all overrides.'}
        confirmLabel="Delete"
        onConfirm={handleDelete}
        variant="destructive"
      />
    </div>
  )
}
