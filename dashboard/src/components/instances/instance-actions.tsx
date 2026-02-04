import { useState } from 'react'
import { Loader2 } from 'lucide-react'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  AlertDialog,
  AlertDialogContent,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogCancel,
  AlertDialogAction,
} from '@/components/ui/alert-dialog'
import {
  useDeleteInstance,
  useDuplicateInstance,
  useRenameInstance,
  useInstanceDeletePreview,
} from '@/features/instances/queries'

interface DeleteInstanceDialogProps {
  stackName: string
  instanceId: string
  open: boolean
  onOpenChange: (open: boolean) => void
  onSuccess?: () => void
}

export function DeleteInstanceDialog({
  stackName,
  instanceId,
  open,
  onOpenChange,
  onSuccess,
}: DeleteInstanceDialogProps) {
  const deleteInstance = useDeleteInstance(stackName, instanceId)
  const { data: preview, isLoading: previewLoading } = useInstanceDeletePreview(stackName, instanceId)

  const handleConfirm = () => {
    deleteInstance.mutate(undefined, {
      onSuccess: () => {
        onOpenChange(false)
        onSuccess?.()
      },
    })
  }

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Delete {instanceId}?</AlertDialogTitle>
          <AlertDialogDescription asChild>
            <div className="space-y-2">
              <p>This will permanently delete the instance and all overrides.</p>
              {previewLoading ? (
                <div className="flex items-center gap-2 py-2">
                  <Loader2 className="size-4 animate-spin text-muted-foreground" />
                  <span className="text-sm text-muted-foreground">Loading preview...</span>
                </div>
              ) : preview ? (
                <div className="bg-muted p-3 rounded-md text-sm space-y-2">
                  <div>
                    <span className="font-medium">Instance:</span> {preview.instance_id}
                  </div>
                  <div>
                    <span className="font-medium">Template:</span> {preview.template_name}
                  </div>
                  <div>
                    <span className="font-medium">Overrides:</span> {preview.override_count}
                  </div>
                  {preview.container_name && (
                    <div>
                      <span className="font-medium">Container:</span> <code>{preview.container_name}</code>
                    </div>
                  )}
                </div>
              ) : null}
            </div>
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>Cancel</AlertDialogCancel>
          <AlertDialogAction
            onClick={handleConfirm}
            disabled={deleteInstance.isPending}
            className="bg-destructive text-white hover:bg-destructive/90"
          >
            {deleteInstance.isPending ? <Loader2 className="size-4 animate-spin" /> : 'Delete'}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}

interface DuplicateInstanceDialogProps {
  stackName: string
  instanceId: string
  open: boolean
  onOpenChange: (open: boolean) => void
  onSuccess?: () => void
}

export function DuplicateInstanceDialog({
  stackName,
  instanceId,
  open,
  onOpenChange,
  onSuccess,
}: DuplicateInstanceDialogProps) {
  const [newName, setNewName] = useState(`${instanceId}-copy`)
  const duplicateInstance = useDuplicateInstance(stackName, instanceId)

  const handleOpenChange = (isOpen: boolean) => {
    if (isOpen) {
      setNewName(`${instanceId}-copy`)
    }
    onOpenChange(isOpen)
  }

  const handleDuplicate = () => {
    duplicateInstance.mutate(newName, {
      onSuccess: () => {
        onOpenChange(false)
        onSuccess?.()
      },
    })
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Duplicate Instance</DialogTitle>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          <div className="grid gap-2">
            <label className="text-sm font-medium">New Instance Name</label>
            <Input
              value={newName}
              onChange={(e) => setNewName(e.target.value)}
              placeholder="Enter instance name"
            />
          </div>
          <p className="text-sm text-muted-foreground">
            This will create a copy of the instance with all overrides.
          </p>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button
            onClick={handleDuplicate}
            disabled={duplicateInstance.isPending || !newName}
          >
            {duplicateInstance.isPending ? <Loader2 className="size-4 animate-spin" /> : 'Duplicate'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

interface RenameInstanceDialogProps {
  stackName: string
  instanceId: string
  open: boolean
  onOpenChange: (open: boolean) => void
  onSuccess?: (newName: string) => void
}

export function RenameInstanceDialog({
  stackName,
  instanceId,
  open,
  onOpenChange,
  onSuccess,
}: RenameInstanceDialogProps) {
  const [newName, setNewName] = useState(instanceId)
  const renameInstance = useRenameInstance(stackName, instanceId)

  const handleOpenChange = (isOpen: boolean) => {
    if (isOpen) {
      setNewName(instanceId)
    }
    onOpenChange(isOpen)
  }

  const handleRename = () => {
    renameInstance.mutate({ instance_id: newName }, {
      onSuccess: () => {
        onOpenChange(false)
        onSuccess?.(newName)
      },
    })
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Rename Instance</DialogTitle>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          <div className="grid gap-2">
            <label className="text-sm font-medium">Instance Name</label>
            <Input
              value={newName}
              onChange={(e) => setNewName(e.target.value)}
              placeholder="Enter new name"
            />
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button
            onClick={handleRename}
            disabled={renameInstance.isPending || !newName || newName === instanceId}
          >
            {renameInstance.isPending ? <Loader2 className="size-4 animate-spin" /> : 'Rename'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
