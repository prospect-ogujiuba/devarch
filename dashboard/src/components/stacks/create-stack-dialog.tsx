import { useState } from 'react'
import { Loader2 } from 'lucide-react'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { useCreateStack } from '@/features/stacks/queries'
import { useNavigate } from '@tanstack/react-router'

interface CreateStackDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function CreateStackDialog({ open, onOpenChange }: CreateStackDialogProps) {
  const navigate = useNavigate()
  const createStack = useCreateStack()
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
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
    setNameError('')
    return true
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!validateName(name)) return

    createStack.mutate(
      { name, description },
      {
        onSuccess: () => {
          onOpenChange(false)
          setName('')
          setDescription('')
          navigate({ to: '/stacks/$name', params: { name } })
        },
      }
    )
  }

  const handleOpenChange = (newOpen: boolean) => {
    if (!newOpen) {
      setName('')
      setDescription('')
      setNameError('')
    }
    onOpenChange(newOpen)
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Create Stack</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit}>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <label className="text-sm font-medium">Name</label>
              <Input
                value={name}
                onChange={(e) => {
                  setName(e.target.value)
                  setNameError('')
                }}
                onBlur={(e) => validateName(e.target.value)}
                placeholder="my-laravel-stack"
                autoFocus
              />
              {nameError && (
                <p className="text-sm text-destructive">{nameError}</p>
              )}
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
            <Button type="button" variant="outline" onClick={() => handleOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={createStack.isPending || !name}>
              {createStack.isPending ? <Loader2 className="size-4 animate-spin" /> : 'Create'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
