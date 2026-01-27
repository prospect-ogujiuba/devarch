import { createFileRoute } from '@tanstack/react-router'
import { Loader2 } from 'lucide-react'
import { useServices } from '@/features/services/queries'
import { ServiceTable } from '@/components/services/service-table'
import { z } from 'zod'

const servicesSearchSchema = z.object({
  search: z.string().optional(),
  category: z.string().optional(),
  status: z.string().optional(),
})

export const Route = createFileRoute('/services/')({
  component: ServicesPage,
  validateSearch: servicesSearchSchema,
})

function ServicesPage() {
  const { data, isLoading } = useServices()
  const search = Route.useSearch()

  const services = data?.services ?? []
  const total = data?.total ?? 0
  const categories = [...new Set(services.map((s) => s.category?.name).filter(Boolean))] as string[]

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="size-8 animate-spin text-muted-foreground" />
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Services</h1>
        <p className="text-muted-foreground">
          Manage all {total} services in your environment
        </p>
      </div>

      <ServiceTable
        services={services}
        categories={categories}
        searchQuery={search.search}
        categoryFilter={search.category}
        statusFilter={search.status}
      />
    </div>
  )
}
