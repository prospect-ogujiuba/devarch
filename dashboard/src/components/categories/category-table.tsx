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
import type { CategoryItem } from '@/types/api'

interface CategoryTableProps {
  categories: CategoryItem[]
  compact?: boolean
}

export function CategoryTable({ categories, compact }: CategoryTableProps) {
  return (
    <div className="rounded-lg border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Name</TableHead>
            {!compact && <TableHead>Services</TableHead>}
            <TableHead>Running</TableHead>
            <TableHead>Progress</TableHead>
            {!compact && <TableHead>Order</TableHead>}
            <TableHead className="text-right">Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {categories.length === 0 ? (
            <TableRow>
              <TableCell colSpan={compact ? 4 : 6} className="text-center text-muted-foreground py-8">
                No categories found
              </TableCell>
            </TableRow>
          ) : (
            categories.map((cat) => (
              <CategoryTableRow key={cat.name} category={cat} compact={compact} />
            ))
          )}
        </TableBody>
      </Table>
    </div>
  )
}

function CategoryTableRow({ category, compact }: { category: CategoryItem; compact?: boolean }) {
  const startMutation = useStartCategory()
  const stopMutation = useStopCategory()
  const isLoading = startMutation.isPending || stopMutation.isPending
  const { name, runningCount, serviceCount, startupOrder } = category
  const pct = serviceCount > 0 ? (runningCount / serviceCount) * 100 : 0

  return (
    <TableRow>
      <TableCell>
        <Link to="/services" search={{ category: name }} className="font-medium hover:underline">
          {titleCase(name)}
        </Link>
      </TableCell>
      {!compact && <TableCell>{serviceCount}</TableCell>}
      <TableCell>{runningCount}/{serviceCount}</TableCell>
      <TableCell>
        <ResourceBar value={pct} className="w-24" />
      </TableCell>
      {!compact && <TableCell className="text-muted-foreground">{startupOrder}</TableCell>}
      <TableCell className="text-right">
        <div className="flex items-center justify-end gap-1">
          {runningCount < serviceCount && (
            <Button variant="ghost" size="icon-sm" onClick={() => startMutation.mutate(name)} disabled={isLoading}>
              {startMutation.isPending ? <Loader2 className="size-4 animate-spin" /> : <Play className="size-4" />}
            </Button>
          )}
          {runningCount > 0 && (
            <Button variant="ghost" size="icon-sm" onClick={() => stopMutation.mutate(name)} disabled={isLoading}>
              {stopMutation.isPending ? <Loader2 className="size-4 animate-spin" /> : <Square className="size-4" />}
            </Button>
          )}
        </div>
      </TableCell>
    </TableRow>
  )
}
