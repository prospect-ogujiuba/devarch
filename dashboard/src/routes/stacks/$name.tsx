import { useState } from 'react'
import { createFileRoute, Link, useNavigate } from '@tanstack/react-router'
import { ArrowLeft, Loader2, Layers, Power, PowerOff, MoreVertical, Copy, Edit, Trash2, FileEdit } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { useStack, useEnableStack, useDisableStack } from '@/features/stacks/queries'
import { CreateStackDialog } from '@/components/stacks/create-stack-dialog'
import { EditStackDialog } from '@/components/stacks/edit-stack-dialog'
import { DeleteStackDialog } from '@/components/stacks/delete-stack-dialog'
import { CloneStackDialog } from '@/components/stacks/clone-stack-dialog'
import { RenameStackDialog } from '@/components/stacks/rename-stack-dialog'
import { DisableStackDialog } from '@/components/stacks/disable-stack-dialog'
import { formatDistanceToNow } from 'date-fns'

export const Route = createFileRoute('/stacks/$name')({
  component: StackDetailPage,
})

function StackDetailPage() {
  const { name } = Route.useParams()
  const navigate = useNavigate()
  const { data: stack, isLoading } = useStack(name)
  const enableStack = useEnableStack()
  const disableStack = useDisableStack()

  const [editOpen, setEditOpen] = useState(false)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [cloneOpen, setCloneOpen] = useState(false)
  const [renameOpen, setRenameOpen] = useState(false)
  const [disableDialogOpen, setDisableDialogOpen] = useState(false)

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

  const createdAgo = formatDistanceToNow(new Date(stack.created_at), { addSuffix: true })
  const updatedAgo = formatDistanceToNow(new Date(stack.updated_at), { addSuffix: true })

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

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Instances</CardTitle>
        </CardHeader>
        <CardContent>
          {!stack.instances || stack.instances.length === 0 ? (
            <div className="text-center py-8">
              <Layers className="size-12 mx-auto text-muted-foreground/50 mb-2" />
              <p className="text-muted-foreground">No instances yet</p>
              <p className="text-sm text-muted-foreground">Add services in Phase 3</p>
            </div>
          ) : (
            <div className="border rounded-lg overflow-hidden">
              <table className="w-full text-sm">
                <thead className="bg-muted/50">
                  <tr>
                    <th className="text-left px-4 py-2 font-medium">Instance ID</th>
                    <th className="text-left px-4 py-2 font-medium">Template Service</th>
                    <th className="text-left px-4 py-2 font-medium">Container Name</th>
                    <th className="text-left px-4 py-2 font-medium">Status</th>
                  </tr>
                </thead>
                <tbody>
                  {stack.instances.map((instance) => (
                    <tr key={instance.id} className="border-t">
                      <td className="px-4 py-2 font-mono text-xs">{instance.instance_id}</td>
                      <td className="px-4 py-2">
                        {instance.template_service_id ? `Service #${instance.template_service_id}` : '-'}
                      </td>
                      <td className="px-4 py-2 font-mono text-xs">
                        {instance.container_name || '-'}
                      </td>
                      <td className="px-4 py-2">
                        <Badge variant={instance.enabled ? 'default' : 'outline'} className="text-xs">
                          {instance.enabled ? 'Enabled' : 'Disabled'}
                        </Badge>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Network</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="text-sm">
            {stack.network_name ? (
              <div className="flex items-center gap-2">
                <span className="text-muted-foreground">Network:</span>
                <code className="bg-muted px-2 py-1 rounded">{stack.network_name}</code>
              </div>
            ) : (
              <p className="text-muted-foreground">Not configured</p>
            )}
          </div>
        </CardContent>
      </Card>

      {stack && (
        <>
          <EditStackDialog stack={stack} open={editOpen} onOpenChange={setEditOpen} />
          <DeleteStackDialog stack={stack} open={deleteOpen} onOpenChange={setDeleteOpen} onSuccess={handleDeleteSuccess} />
          <CloneStackDialog stack={stack} open={cloneOpen} onOpenChange={setCloneOpen} />
          <RenameStackDialog stack={stack} open={renameOpen} onOpenChange={setRenameOpen} />
          <DisableStackDialog stack={stack} open={disableDialogOpen} onOpenChange={setDisableDialogOpen} />
        </>
      )}
    </div>
  )
}
