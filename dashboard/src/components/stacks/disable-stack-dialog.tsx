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
import { useDisableStack } from '@/features/stacks/queries'
import type { Stack } from '@/types/api'

interface DisableStackDialogProps {
  stack: Stack
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function DisableStackDialog({ stack, open, onOpenChange }: DisableStackDialogProps) {
  const disableStack = useDisableStack()

  const handleConfirm = () => {
    disableStack.mutate(stack.name, {
      onSuccess: () => {
        onOpenChange(false)
      },
    })
  }

  const containerNames = stack.instances?.map(inst => inst.container_name).filter(Boolean) || []
  const containerCount = containerNames.length

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Disable {stack.name}?</AlertDialogTitle>
          <AlertDialogDescription asChild>
            <div className="space-y-2">
              <p>This will disable the stack and stop all running containers.</p>
              {containerCount > 0 ? (
                <div className="bg-muted p-3 rounded-md text-sm">
                  <p className="font-medium">
                    {containerCount} {containerCount === 1 ? 'container' : 'containers'} will be stopped:
                  </p>
                  <ul className="list-disc list-inside space-y-0.5 mt-2">
                    {containerNames.map((name) => (
                      <li key={name} className="text-xs font-mono">
                        {name}
                      </li>
                    ))}
                  </ul>
                </div>
              ) : (
                <p className="text-sm text-muted-foreground">No containers to stop.</p>
              )}
            </div>
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>Cancel</AlertDialogCancel>
          <AlertDialogAction
            onClick={handleConfirm}
            disabled={disableStack.isPending}
          >
            {disableStack.isPending ? <Loader2 className="size-4 animate-spin" /> : 'Disable'}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
