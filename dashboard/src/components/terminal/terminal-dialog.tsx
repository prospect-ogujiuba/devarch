import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { ContainerTerminal } from './container-terminal'

interface TerminalDialogProps {
  containerName: string
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function TerminalDialog({ containerName, open, onOpenChange }: TerminalDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-4xl">
        <DialogHeader>
          <DialogTitle>Terminal — {containerName}</DialogTitle>
        </DialogHeader>
        {open && <ContainerTerminal containerName={containerName} />}
      </DialogContent>
    </Dialog>
  )
}
