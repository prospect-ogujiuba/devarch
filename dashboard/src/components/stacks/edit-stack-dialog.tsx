import { useState, useEffect } from 'react'
import { Loader2 } from 'lucide-react'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import { useUpdateStack } from '@/features/stacks/queries'
import type { Stack } from '@/types/api'

interface EditStackDialogProps {
  stack: Stack
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function EditStackDialog({ stack, open, onOpenChange }: EditStackDialogProps) {
  const updateStack = useUpdateStack()
  const [description, setDescription] = useState(stack.description)

  useEffect(() => {
    setDescription(stack.description)
  }, [stack.description, open])

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    updateStack.mutate(
      { name: stack.name, data: { description } },
      {
        onSuccess: () => {
          onOpenChange(false)
        },
      }
    )
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Edit {stack.name}</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit}>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <label className="text-sm font-medium">Name</label>
              <div className="text-sm text-muted-foreground bg-muted px-3 py-2 rounded-md">
                {stack.name}
              </div>
              <p className="text-xs text-muted-foreground">Stack name cannot be changed (use rename instead)</p>
            </div>
            <div className="grid gap-2">
              <label className="text-sm font-medium">Description</label>
              <Textarea
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="Optional description"
              />
            </div>
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={updateStack.isPending}>
              {updateStack.isPending ? <Loader2 className="size-4 animate-spin" /> : 'Save'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
