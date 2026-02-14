import { useState } from 'react'
import { Link } from '@tanstack/react-router'
import { Play, Square, Loader2, Pencil, Trash2 } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { ResourceBar } from '@/components/ui/resource-bar'
import { useStartCategory, useStopCategory } from '@/features/categories/queries'
import { EditCategoryDialog } from './edit-category-dialog'
import { DeleteCategoryDialog } from './delete-category-dialog'
import type { Category } from '@/types/api'

interface CategoryCardProps {
  category: Category
  compact?: boolean
}

export function CategoryCard({ category, compact }: CategoryCardProps) {
  const startMutation = useStartCategory()
  const stopMutation = useStopCategory()
  const isLoading = startMutation.isPending || stopMutation.isPending
  const [editOpen, setEditOpen] = useState(false)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const { name } = category
  const runningCount = category.runningCount ?? 0
  const serviceCount = category.service_count ?? 0
  const allRunning = runningCount === serviceCount && serviceCount > 0
  const allStopped = runningCount === 0
  const pct = serviceCount > 0 ? (runningCount / serviceCount) * 100 : 0

  if (compact) {
    return (
      <Card className="py-4">
        <CardHeader className="pb-2">
          <div className="flex items-center justify-between">
            <Link to="/services" search={{ category: name }} className="hover:underline">
              <CardTitle className="text-base capitalize">{name}</CardTitle>
            </Link>
            <span className="text-sm text-muted-foreground">
              {runningCount}/{serviceCount}
            </span>
          </div>
        </CardHeader>
        <CardContent>
          <div className="flex items-center gap-2">
            <ResourceBar value={pct} className="flex-1" />
            <div className="flex items-center gap-1">
              {!allRunning && (
                <Button variant="ghost" size="icon-sm" onClick={() => startMutation.mutate(name)} disabled={isLoading}>
                  {startMutation.isPending ? <Loader2 className="size-4 animate-spin" /> : <Play className="size-4" />}
                </Button>
              )}
              {!allStopped && (
                <Button variant="ghost" size="icon-sm" onClick={() => stopMutation.mutate(name)} disabled={isLoading}>
                  {stopMutation.isPending ? <Loader2 className="size-4 animate-spin" /> : <Square className="size-4" />}
                </Button>
              )}
            </div>
          </div>
        </CardContent>
      </Card>
    )
  }

  return (
    <>
    <Card className="py-4 hover:border-primary/50 transition-colors">
      <CardHeader className="pb-2">
        <div className="flex items-center justify-between">
          <Link to="/services" search={{ category: name }} className="hover:underline">
            <CardTitle className="text-base capitalize">{name}</CardTitle>
          </Link>
          <div className="flex items-center gap-1">
            <Button variant="ghost" size="icon-sm" onClick={() => setEditOpen(true)}>
              <Pencil className="size-3.5" />
            </Button>
            <Button variant="ghost" size="icon-sm" onClick={() => setDeleteOpen(true)}>
              <Trash2 className="size-3.5" />
            </Button>
          </div>
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="flex items-center gap-4">
          <div className="flex-1">
            <div className="text-2xl font-bold">
              {runningCount}
              <span className="text-muted-foreground font-normal">
                /{serviceCount}
              </span>
            </div>
            <div className="text-xs text-muted-foreground">services running</div>
          </div>
          <div className="h-2 flex-1 rounded-full bg-muted overflow-hidden">
            <div
              className="h-full bg-success transition-all"
              style={{ width: `${pct}%` }}
            />
          </div>
        </div>
        <div className="flex items-center gap-2">
          {!allRunning && (
            <Button
              variant="outline"
              size="sm"
              className="flex-1"
              onClick={() => startMutation.mutate(name)}
              disabled={isLoading}
            >
              {startMutation.isPending ? (
                <Loader2 className="size-4 animate-spin" />
              ) : (
                <Play className="size-4" />
              )}
              Start All
            </Button>
          )}
          {!allStopped && (
            <Button
              variant="outline"
              size="sm"
              className="flex-1"
              onClick={() => stopMutation.mutate(name)}
              disabled={isLoading}
            >
              {stopMutation.isPending ? (
                <Loader2 className="size-4 animate-spin" />
              ) : (
                <Square className="size-4" />
              )}
              Stop All
            </Button>
          )}
        </div>
      </CardContent>
    </Card>
    <EditCategoryDialog category={category} open={editOpen} onOpenChange={setEditOpen} />
    <DeleteCategoryDialog category={category} open={deleteOpen} onOpenChange={setDeleteOpen} />
    </>
  )
}
