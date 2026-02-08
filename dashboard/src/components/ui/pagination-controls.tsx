import { ChevronLeft, ChevronRight } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

interface PaginationControlsProps {
  page: number
  totalPages: number
  totalItems: number
  pageSize: number
  pageSizeOptions: number[]
  onPageChange: (page: number) => void
  onPageSizeChange: (pageSize: number) => void
  itemLabel?: string
}

export function PaginationControls({
  page,
  totalPages,
  totalItems,
  pageSize,
  pageSizeOptions,
  onPageChange,
  onPageSizeChange,
  itemLabel = 'items',
}: PaginationControlsProps) {
  const canPrev = page > 1
  const canNext = page < totalPages
  const start = totalItems === 0 ? 0 : (page - 1) * pageSize + 1
  const end = Math.min(page * pageSize, totalItems)

  return (
    <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
      <div className="text-sm text-muted-foreground">
        Showing {start}-{end} of {totalItems} {itemLabel}
      </div>

      <div className="flex items-center gap-2">
        <Select
          value={String(pageSize)}
          onValueChange={(value) => onPageSizeChange(Number(value))}
        >
          <SelectTrigger className="h-8 w-[130px]">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {pageSizeOptions.map((size) => (
              <SelectItem key={size} value={String(size)}>
                {size} / page
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        <Button
          variant="outline"
          size="sm"
          onClick={() => onPageChange(page - 1)}
          disabled={!canPrev}
        >
          <ChevronLeft className="size-4" />
          Prev
        </Button>
        <span className="min-w-20 text-center text-sm text-muted-foreground">
          {page} / {totalPages}
        </span>
        <Button
          variant="outline"
          size="sm"
          onClick={() => onPageChange(page + 1)}
          disabled={!canNext}
        >
          Next
          <ChevronRight className="size-4" />
        </Button>
      </div>
    </div>
  )
}
