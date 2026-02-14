import { useState } from 'react'
import { Loader2 } from 'lucide-react'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { useCreateCategory } from '@/features/categories/queries'

interface CreateCategoryDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function CreateCategoryDialog({ open, onOpenChange }: CreateCategoryDialogProps) {
  const createCategory = useCreateCategory()
  const [name, setName] = useState('')
  const [displayName, setDisplayName] = useState('')
  const [color, setColor] = useState('#3b82f6')
  const [startupOrder, setStartupOrder] = useState('0')
  const [nameError, setNameError] = useState('')

  const validateName = (value: string) => {
    if (!value) {
      setNameError('Name is required')
      return false
    }
    if (!/^[a-z0-9][a-z0-9-]*[a-z0-9]$/.test(value) && !/^[a-z0-9]$/.test(value)) {
      setNameError('Must be lowercase alphanumeric with hyphens')
      return false
    }
    if (value.length > 64) {
      setNameError('Max 64 characters')
      return false
    }
    setNameError('')
    return true
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!validateName(name)) return

    createCategory.mutate(
      {
        name,
        display_name: displayName || undefined,
        color: color || undefined,
        startup_order: parseInt(startupOrder) || 0,
      },
      {
        onSuccess: () => {
          onOpenChange(false)
          resetForm()
        },
      }
    )
  }

  const resetForm = () => {
    setName('')
    setDisplayName('')
    setColor('#3b82f6')
    setStartupOrder('0')
    setNameError('')
  }

  const handleOpenChange = (newOpen: boolean) => {
    if (!newOpen) resetForm()
    onOpenChange(newOpen)
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Create Category</DialogTitle>
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
                onBlur={(e) => e.target.value && validateName(e.target.value)}
                placeholder="my-category"
                autoFocus
              />
              {nameError && <p className="text-sm text-destructive">{nameError}</p>}
            </div>
            <div className="grid gap-2">
              <label className="text-sm font-medium">Display Name</label>
              <Input
                value={displayName}
                onChange={(e) => setDisplayName(e.target.value)}
                placeholder="My Category"
              />
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div className="grid gap-2">
                <label className="text-sm font-medium">Color</label>
                <div className="flex items-center gap-2">
                  <input
                    type="color"
                    value={color}
                    onChange={(e) => setColor(e.target.value)}
                    className="size-8 rounded border cursor-pointer"
                  />
                  <Input
                    value={color}
                    onChange={(e) => setColor(e.target.value)}
                    placeholder="#3b82f6"
                    className="flex-1"
                  />
                </div>
              </div>
              <div className="grid gap-2">
                <label className="text-sm font-medium">Startup Order</label>
                <Input
                  type="number"
                  value={startupOrder}
                  onChange={(e) => setStartupOrder(e.target.value)}
                  min="0"
                />
              </div>
            </div>
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => handleOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={createCategory.isPending || !name}>
              {createCategory.isPending ? <Loader2 className="size-4 animate-spin" /> : 'Create'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
