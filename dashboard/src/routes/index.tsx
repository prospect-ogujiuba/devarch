import { createFileRoute, Link } from '@tanstack/react-router'
import { Server, Play, Square, Activity, Loader2 } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { useStatusOverview } from '@/features/status/queries'
import { useStartCategory, useStopCategory } from '@/features/categories/queries'
import type { CategoryOverview } from '@/types/api'

export const Route = createFileRoute('/')({
  component: OverviewPage,
})

function OverviewPage() {
  const { data: status, isLoading } = useStatusOverview()

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="size-8 animate-spin text-muted-foreground" />
      </div>
    )
  }

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-bold">Overview</h1>
        <p className="text-muted-foreground">Monitor and manage your development services</p>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <StatCard
          title="Total Services"
          value={status?.total_services ?? 0}
          icon={<Server className="size-4" />}
        />
        <StatCard
          title="Running"
          value={status?.running_services ?? 0}
          icon={<Activity className="size-4" />}
          className="border-success/50"
        />
        <StatCard
          title="Stopped"
          value={status?.stopped_services ?? 0}
          icon={<Square className="size-4" />}
        />
        <StatCard
          title="Categories"
          value={status?.categories?.length ?? 0}
          icon={<Server className="size-4" />}
        />
      </div>

      <div>
        <h2 className="text-lg font-semibold mb-4">Categories</h2>
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {status?.categories?.map((category: CategoryOverview) => (
            <CategoryQuickCard key={category.name} category={category} />
          ))}
        </div>
      </div>
    </div>
  )
}

function StatCard({
  title,
  value,
  icon,
  className,
}: {
  title: string
  value: number
  icon: React.ReactNode
  className?: string
}) {
  return (
    <Card className={className}>
      <CardHeader className="flex flex-row items-center justify-between pb-2">
        <CardTitle className="text-sm font-medium text-muted-foreground">
          {title}
        </CardTitle>
        {icon}
      </CardHeader>
      <CardContent>
        <div className="text-3xl font-bold">{value}</div>
      </CardContent>
    </Card>
  )
}

function CategoryQuickCard({
  category,
}: {
  category: CategoryOverview
}) {
  const startMutation = useStartCategory()
  const stopMutation = useStopCategory()
  const isLoading = startMutation.isPending || stopMutation.isPending

  const serviceCount = category.total_services ?? 0
  const runningCount = category.running_services ?? 0
  const allRunning = runningCount === serviceCount && serviceCount > 0
  const allStopped = runningCount === 0

  return (
    <Card className="py-4">
      <CardHeader className="pb-2">
        <div className="flex items-center justify-between">
          <Link
            to="/services"
            search={{ category: category.name }}
            className="hover:underline"
          >
            <CardTitle className="text-base capitalize">{category.name}</CardTitle>
          </Link>
          <span className="text-sm text-muted-foreground">
            {runningCount}/{serviceCount}
          </span>
        </div>
      </CardHeader>
      <CardContent>
        <div className="flex items-center gap-2">
          <div className="h-2 flex-1 rounded-full bg-muted overflow-hidden">
            <div
              className="h-full bg-success transition-all"
              style={{
                width: serviceCount > 0 ? `${(runningCount / serviceCount) * 100}%` : '0%',
              }}
            />
          </div>
          <div className="flex items-center gap-1">
            {!allRunning && (
              <Button
                variant="ghost"
                size="icon-sm"
                onClick={() => startMutation.mutate(category.name)}
                disabled={isLoading}
              >
                {startMutation.isPending ? (
                  <Loader2 className="size-4 animate-spin" />
                ) : (
                  <Play className="size-4" />
                )}
              </Button>
            )}
            {!allStopped && (
              <Button
                variant="ghost"
                size="icon-sm"
                onClick={() => stopMutation.mutate(category.name)}
                disabled={isLoading}
              >
                {stopMutation.isPending ? (
                  <Loader2 className="size-4 animate-spin" />
                ) : (
                  <Square className="size-4" />
                )}
              </Button>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
