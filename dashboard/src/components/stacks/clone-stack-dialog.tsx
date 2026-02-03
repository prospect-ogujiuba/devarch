import { useState } from 'react'
import { Loader2 } from 'lucide-react'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { useCloneStack } from '@/features/stacks/queries'
import type { Stack } from '@/types/api'

interface CloneStackDialogProps {
  stack: Stack
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function CloneStackDialog({ stack, open, onOpenChange }: CloneStackDialogProps) {
  const cloneStack = useCloneStack()
  const [newName, setNewName] = useState('')
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
      setNameError('New name must be different from the original')
      return false
    }
    setNameError('')
    return true
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!validateName(newName)) return

    cloneStack.mutate(
      { name: stack.name, newName },
      {
        onSuccess: () => {
          onOpenChange(false)
          setNewName('')
        },
      }
    )
  }

  const handleOpenChange = (newOpen: boolean) => {
    if (!newOpen) {
      setNewName('')
      setNameError('')
    }
    onOpenChange(newOpen)
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Clone Stack</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit}>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <label className="text-sm font-medium">Source Stack</label>
              <div className="text-sm text-muted-foreground bg-muted px-3 py-2 rounded-md">
                {stack.name}
              </div>
            </div>
            <div className="grid gap-2">
              <label className="text-sm font-medium">New Stack Name</label>
              <Input
                value={newName}
                onChange={(e) => {
                  setNewName(e.target.value)
                  setNameError('')
                }}
                onBlur={(e) => validateName(e.target.value)}
                placeholder="my-new-stack"
                autoFocus
              />
              {nameError && (
                <p className="text-sm text-destructive">{nameError}</p>
              )}
              <p className="text-xs text-muted-foreground">
                Cloning creates a copy with all instances and overrides. No containers are started.
              </p>
            </div>
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => handleOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={cloneStack.isPending || !newName}>
              {cloneStack.isPending ? <Loader2 className="size-4 animate-spin" /> : 'Clone'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
