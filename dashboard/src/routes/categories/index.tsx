import { createFileRoute } from '@tanstack/react-router'
import { Loader2 } from 'lucide-react'
import { useCategories } from '@/features/categories/queries'
import { CategoryCard } from '@/components/categories/category-card'

export const Route = createFileRoute('/categories/')({
  component: CategoriesPage,
})

function CategoriesPage() {
  const { data: categories, isLoading } = useCategories()

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="size-8 animate-spin text-muted-foreground" />
      </div>
    )
  }

  const totalServices = categories?.reduce((acc, cat) => acc + (cat.service_count ?? 0), 0) ?? 0
  const totalRunning = categories?.reduce((acc, cat) => acc + (cat.runningCount ?? 0), 0) ?? 0

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Categories</h1>
        <p className="text-muted-foreground">
          {totalRunning} of {totalServices} services running across {categories?.length ?? 0} categories
        </p>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {categories?.map((category) => (
          <CategoryCard key={category.name} category={category} />
        ))}
      </div>
    </div>
  )
}
