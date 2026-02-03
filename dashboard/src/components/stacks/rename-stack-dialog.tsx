import { useState } from 'react'
import { Loader2 } from 'lucide-react'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { useRenameStack } from '@/features/stacks/queries'
import type { Stack } from '@/types/api'

interface RenameStackDialogProps {
  stack: Stack
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function RenameStackDialog({ stack, open, onOpenChange }: RenameStackDialogProps) {
  const renameStack = useRenameStack()
  const [newName, setNewName] = useState(stack.name)
  const [nameError, setNameError] = useState('')

  const validateName = (value: string) => {
    if (!value) {
      setNameError('Name is required')
      return false
    }
    if (!/^[a-z0-9-]+$/.test(value)) {
      setNameError('Name must contain only lowercase letters, numbers, and hyphens')
      return false
    }
    if (value.length > 63) {
      setNameError('Name must be 63 characters or less')
      return false
    }
    if (value.startsWith('-') || value.endsWith('-')) {
      setNameError('Name cannot start or end with a hyphen')
      return false
    }
    if (value === stack.name) {
      setNameError('New name must be different from the current name')
      return false
    }
    setNameError('')
    return true
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!validateName(newName)) return

    renameStack.mutate(
      { name: stack.name, newName },
      {
        onSuccess: () => {
          onOpenChange(false)
        },
      }
    )
  }

  const handleOpenChange = (newOpen: boolean) => {
    if (!newOpen) {
      setNewName(stack.name)
      setNameError('')
    }
    onOpenChange(newOpen)
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Rename Stack</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit}>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <label className="text-sm font-medium">Current Name</label>
              <div className="text-sm text-muted-foreground bg-muted px-3 py-2 rounded-md">
                {stack.name}
              </div>
            </div>
            <div className="grid gap-2">
              <label className="text-sm font-medium">New Name</label>
              <Input
                value={newName}
                onChange={(e) => {
                  setNewName(e.target.value)
                  setNameError('')
                }}
                onBlur={(e) => validateName(e.target.value)}
                placeholder="new-stack-name"
                autoFocus
              />
              {nameError && (
                <p className="text-sm text-destructive">{nameError}</p>
              )}
              <p className="text-xs text-muted-foreground">
                Rename creates a copy with the new name and archives the old stack. This operation is atomic.
              </p>
            </div>
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => handleOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={renameStack.isPending || newName === stack.name}>
              {renameStack.isPending ? <Loader2 className="size-4 animate-spin" /> : 'Rename'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
