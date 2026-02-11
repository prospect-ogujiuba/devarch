import { useState } from 'react'
import { Loader2 } from 'lucide-react'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { useCreateProject } from '@/features/projects/queries'
import { useNavigate } from '@tanstack/react-router'

interface CreateProjectDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function CreateProjectDialog({ open, onOpenChange }: CreateProjectDialogProps) {
  const navigate = useNavigate()
  const createProject = useCreateProject()
  const [name, setName] = useState('')
  const [path, setPath] = useState('')
  const [projectType, setProjectType] = useState('')
  const [language, setLanguage] = useState('')
  const [framework, setFramework] = useState('')
  const [description, setDescription] = useState('')
  const [nameError, setNameError] = useState('')

  const validateName = (value: string) => {
    if (!value) {
      setNameError('Name is required')
      return false
    }
    if (!/^[a-z0-9-]+$/.test(value)) {
      setNameError('Lowercase letters, numbers, and hyphens only')
      return false
    }
    if (value.length > 63) {
      setNameError('63 characters max')
      return false
    }
    if (value.startsWith('-') || value.endsWith('-')) {
      setNameError('Cannot start or end with a hyphen')
      return false
    }
    setNameError('')
    return true
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!validateName(name)) return

    createProject.mutate(
      { name, path: path || '/unknown', project_type: projectType || 'unknown', language, framework, description },
      {
        onSuccess: () => {
          onOpenChange(false)
          resetForm()
          navigate({ to: '/projects/$name', params: { name } })
        },
      }
    )
  }

  const resetForm = () => {
    setName('')
    setPath('')
    setProjectType('')
    setLanguage('')
    setFramework('')
    setDescription('')
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
          <DialogTitle>Create Project</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit}>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <label className="text-sm font-medium">Name</label>
              <Input
                value={name}
                onChange={(e) => { setName(e.target.value); setNameError('') }}
                onBlur={(e) => validateName(e.target.value)}
                placeholder="my-app"
                autoFocus
              />
              {nameError && <p className="text-sm text-destructive">{nameError}</p>}
            </div>
            <div className="grid gap-2">
              <label className="text-sm font-medium">Path</label>
              <Input
                value={path}
                onChange={(e) => setPath(e.target.value)}
                placeholder="/path/to/project"
              />
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div className="grid gap-2">
                <label className="text-sm font-medium">Type</label>
                <Input
                  value={projectType}
                  onChange={(e) => setProjectType(e.target.value)}
                  placeholder="laravel, node, go..."
                />
              </div>
              <div className="grid gap-2">
                <label className="text-sm font-medium">Language</label>
                <Input
                  value={language}
                  onChange={(e) => setLanguage(e.target.value)}
                  placeholder="php, typescript..."
                />
              </div>
            </div>
            <div className="grid gap-2">
              <label className="text-sm font-medium">Framework</label>
              <Input
                value={framework}
                onChange={(e) => setFramework(e.target.value)}
                placeholder="Laravel 11, Next.js..."
              />
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
            <Button type="submit" disabled={createProject.isPending || !name}>
              {createProject.isPending ? <Loader2 className="size-4 animate-spin" /> : 'Create'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
