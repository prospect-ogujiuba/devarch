import { Loader2 } from 'lucide-react'
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
import { useDeleteStack, useDeletePreview } from '@/features/stacks/queries'
import type { Stack } from '@/types/api'

interface DeleteStackDialogProps {
  stack: Stack
  open: boolean
  onOpenChange: (open: boolean) => void
  onSuccess?: () => void
}

export function DeleteStackDialog({ stack, open, onOpenChange, onSuccess }: DeleteStackDialogProps) {
  const deleteStack = useDeleteStack()
  const { data: preview, isLoading: previewLoading } = useDeletePreview(stack.name)

  const handleConfirm = () => {
    deleteStack.mutate(stack.name, {
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
          <AlertDialogTitle>Delete {stack.name}?</AlertDialogTitle>
          <AlertDialogDescription asChild>
            <div className="space-y-2">
              <p>This will soft-delete the stack and stop all containers.</p>
              {previewLoading ? (
                <div className="flex items-center gap-2 py-2">
                  <Loader2 className="size-4 animate-spin text-muted-foreground" />
                  <span className="text-sm text-muted-foreground">Loading cascade preview...</span>
                </div>
              ) : preview ? (
                <div className="bg-muted p-3 rounded-md text-sm">
                  <p className="font-medium">
                    {preview.instance_count} {preview.instance_count === 1 ? 'instance' : 'instances'} will be deleted
                  </p>
                  {preview.container_names.length > 0 && (
                    <div className="mt-2">
                      <p className="text-muted-foreground mb-1">Containers to be stopped:</p>
                      <ul className="list-disc list-inside space-y-0.5">
                        {preview.container_names.map((name) => (
                          <li key={name} className="text-xs font-mono">
                            {name}
                          </li>
                        ))}
                      </ul>
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
            disabled={deleteStack.isPending}
            className="bg-destructive text-white hover:bg-destructive/90"
          >
            {deleteStack.isPending ? <Loader2 className="size-4 animate-spin" /> : 'Delete'}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
