import type { ReactNode } from 'react'
import { Plus, X, Check, Pencil, RotateCcw } from 'lucide-react'
import { Card, CardHeader, CardTitle, CardAction } from '@/components/ui/card'
import { Button } from '@/components/ui/button'

interface EditableCardProps {
  title: string
  editing: boolean
  onEdit: () => void
  onCancel: () => void
  onSave: () => void
  onAdd: () => void
  isPending?: boolean
  isDirty?: boolean
  onResetAll?: () => void
  showResetAll?: boolean
  editVariant?: 'icon' | 'text'
  children: ReactNode
}

export function EditableCard({
  title,
  editing,
  onEdit,
  onCancel,
  onSave,
  onAdd,
  isPending,
  isDirty,
  onResetAll,
  showResetAll,
  editVariant = 'icon',
  children,
}: EditableCardProps) {
  return (
    <Card>
      {editing ? (
        <CardHeader className="flex flex-row flex-wrap items-center justify-between gap-2">
          <CardTitle className="text-base">{title}</CardTitle>
          <div className="flex flex-wrap gap-1">
            <Button variant="outline" size="sm" onClick={onAdd}>
              <Plus className="size-4" /> Add
            </Button>
            {onResetAll && showResetAll && (
              <Button variant="outline" size="sm" onClick={onResetAll}>
                <RotateCcw className="size-4" /> Reset All
              </Button>
            )}
            <Button variant="ghost" size="icon-sm" onClick={onCancel}>
              <X className="size-4" />
            </Button>
            <Button
              variant="default"
              size="icon-sm"
              onClick={onSave}
              disabled={isPending || (isDirty !== undefined && !isDirty)}
            >
              <Check className="size-4" />
            </Button>
          </div>
        </CardHeader>
      ) : (
        <CardHeader>
          <CardTitle className="text-base">{title}</CardTitle>
          <CardAction>
            {editVariant === 'text' ? (
              <Button variant="outline" size="sm" onClick={onEdit}>
                Edit
              </Button>
            ) : (
              <Button variant="ghost" size="icon-sm" onClick={onEdit}>
                <Pencil className="size-4" />
              </Button>
            )}
          </CardAction>
        </CardHeader>
      )}
      {children}
    </Card>
  )
}
