import { AppCard } from './AppCard'

export function AppsGrid({ apps }) {
  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      {apps.map((app) => (
        <AppCard key={app.name} app={app} />
      ))}
    </div>
  )
}
