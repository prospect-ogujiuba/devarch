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
import { useDeleteProject } from '@/features/projects/queries'
import { useNavigate } from '@tanstack/react-router'

interface DeleteProjectDialogProps {
  projectName: string
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function DeleteProjectDialog({ projectName, open, onOpenChange }: DeleteProjectDialogProps) {
  const navigate = useNavigate()
  const deleteProject = useDeleteProject(projectName)

  const handleDelete = () => {
    deleteProject.mutate(undefined, {
      onSuccess: () => {
        onOpenChange(false)
        navigate({ to: '/projects' })
      },
    })
  }

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Delete {projectName}?</AlertDialogTitle>
          <AlertDialogDescription>
            The project will be deleted and its backing stack will be moved to trash.
            You can restore the stack later from the stacks trash.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>Cancel</AlertDialogCancel>
          <AlertDialogAction
            onClick={handleDelete}
            disabled={deleteProject.isPending}
            className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
          >
            {deleteProject.isPending ? <Loader2 className="size-4 animate-spin" /> : 'Delete'}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
