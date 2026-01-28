import { Link } from '@tanstack/react-router'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { ResourceBar } from '@/components/ui/resource-bar'
import { Button } from '@/components/ui/button'
import { Play, Square, Loader2 } from 'lucide-react'
import { useStartCategory, useStopCategory } from '@/features/categories/queries'
import { titleCase } from '@/lib/utils'
import type { Category } from '@/types/api'

interface CategoryTableProps {
  categories: Category[]
}

export function CategoryTable({ categories }: CategoryTableProps) {
  return (
    <div className="rounded-lg border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Name</TableHead>
            <TableHead>Services</TableHead>
            <TableHead>Running</TableHead>
            <TableHead>Progress</TableHead>
            <TableHead>Order</TableHead>
            <TableHead className="text-right">Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {categories.length === 0 ? (
            <TableRow>
              <TableCell colSpan={6} className="text-center text-muted-foreground py-8">
                No categories found
              </TableCell>
            </TableRow>
          ) : (
            categories.map((cat) => (
              <CategoryTableRow key={cat.name} category={cat} />
            ))
          )}
        </TableBody>
      </Table>
    </div>
  )
}

function CategoryTableRow({ category }: { category: Category }) {
  const startMutation = useStartCategory()
  const stopMutation = useStopCategory()
  const isLoading = startMutation.isPending || stopMutation.isPending
  const running = category.runningCount ?? 0
  const total = category.service_count ?? 0
  const pct = total > 0 ? (running / total) * 100 : 0

  return (
    <TableRow>
      <TableCell>
        <Link
          to="/services"
          search={{ category: category.name }}
          className="font-medium hover:underline"
        >
          {titleCase(category.name)}
        </Link>
      </TableCell>
      <TableCell>{total}</TableCell>
      <TableCell>{running}</TableCell>
      <TableCell>
        <ResourceBar value={pct} className="w-24" />
      </TableCell>
      <TableCell className="text-muted-foreground">{category.startup_order}</TableCell>
      <TableCell className="text-right">
        <div className="flex items-center justify-end gap-1">
          {running < total && (
            <Button
              variant="ghost"
              size="icon-sm"
              onClick={() => startMutation.mutate(category.name)}
              disabled={isLoading}
            >
              {startMutation.isPending ? <Loader2 className="size-4 animate-spin" /> : <Play className="size-4" />}
            </Button>
          )}
          {running > 0 && (
            <Button
              variant="ghost"
              size="icon-sm"
              onClick={() => stopMutation.mutate(category.name)}
              disabled={isLoading}
            >
              {stopMutation.isPending ? <Loader2 className="size-4 animate-spin" /> : <Square className="size-4" />}
            </Button>
          )}
        </div>
      </TableCell>
    </TableRow>
  )
}
