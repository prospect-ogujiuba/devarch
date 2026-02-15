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
import { useDeleteCategory } from '@/features/categories/queries'
import { categoryLabel } from '@/lib/utils'
import type { Category } from '@/types/api'

interface DeleteCategoryDialogProps {
  category: Category
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function DeleteCategoryDialog({ category, open, onOpenChange }: DeleteCategoryDialogProps) {
  const deleteCategory = useDeleteCategory()

  const handleConfirm = () => {
    deleteCategory.mutate(category.name, {
      onSuccess: () => onOpenChange(false),
    })
  }

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Delete {categoryLabel(category)}?</AlertDialogTitle>
          <AlertDialogDescription asChild>
            <div className="space-y-2">
              <p>This will permanently delete the category.</p>
              {(category.service_count ?? 0) > 0 && (
                <div className="bg-destructive/10 text-destructive p-3 rounded-md text-sm">
                  This category has {category.service_count} service(s). You must move or delete
                  them before deleting this category.
                </div>
              )}
            </div>
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>Cancel</AlertDialogCancel>
          <AlertDialogAction
            onClick={handleConfirm}
            disabled={deleteCategory.isPending || (category.service_count ?? 0) > 0}
            className="bg-destructive text-white hover:bg-destructive/90"
          >
            {deleteCategory.isPending ? <Loader2 className="size-4 animate-spin" /> : 'Delete'}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
