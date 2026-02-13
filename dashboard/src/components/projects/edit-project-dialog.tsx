import { useState, useEffect } from 'react'
import { Loader2 } from 'lucide-react'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { useUpdateProject } from '@/features/projects/queries'
import type { Project } from '@/types/api'

interface EditProjectDialogProps {
  project: Project
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function EditProjectDialog({ project, open, onOpenChange }: EditProjectDialogProps) {
  const updateProject = useUpdateProject(project.name)
  const [projectType, setProjectType] = useState(project.project_type)
  const [language, setLanguage] = useState(project.language ?? '')
  const [framework, setFramework] = useState(project.framework ?? '')
  const [description, setDescription] = useState(project.description ?? '')
  const [domain, setDomain] = useState(project.domain ?? '')

  useEffect(() => {
    if (open) {
      // eslint-disable-next-line react-hooks/exhaustive-deps
      setProjectType(project.project_type)
      // eslint-disable-next-line react-hooks/exhaustive-deps
      setLanguage(project.language ?? '')
      // eslint-disable-next-line react-hooks/exhaustive-deps
      setFramework(project.framework ?? '')
      // eslint-disable-next-line react-hooks/exhaustive-deps
      setDescription(project.description ?? '')
      // eslint-disable-next-line react-hooks/exhaustive-deps
      setDomain(project.domain ?? '')
    }
  }, [open, project])

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    updateProject.mutate(
      { project_type: projectType, language, framework, description, domain },
      { onSuccess: () => onOpenChange(false) }
    )
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Edit {project.name}</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit}>
          <div className="grid gap-4 py-4">
            <div className="grid grid-cols-2 gap-4">
              <div className="grid gap-2">
                <label className="text-sm font-medium">Type</label>
                <Input value={projectType} onChange={(e) => setProjectType(e.target.value)} />
              </div>
              <div className="grid gap-2">
                <label className="text-sm font-medium">Language</label>
                <Input value={language} onChange={(e) => setLanguage(e.target.value)} />
              </div>
            </div>
            <div className="grid gap-2">
              <label className="text-sm font-medium">Framework</label>
              <Input value={framework} onChange={(e) => setFramework(e.target.value)} />
            </div>
            <div className="grid gap-2">
              <label className="text-sm font-medium">Domain</label>
              <Input value={domain} onChange={(e) => setDomain(e.target.value)} placeholder="myapp.test" />
            </div>
            <div className="grid gap-2">
              <label className="text-sm font-medium">Description</label>
              <Textarea value={description} onChange={(e) => setDescription(e.target.value)} />
            </div>
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>Cancel</Button>
            <Button type="submit" disabled={updateProject.isPending}>
              {updateProject.isPending ? <Loader2 className="size-4 animate-spin" /> : 'Save'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
