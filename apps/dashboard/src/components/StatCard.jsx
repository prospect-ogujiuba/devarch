import { getRuntimeTextClass } from '../utils/colors'

export function StatCard({ label, value, runtime }) {
  const colorClass = runtime ? getRuntimeTextClass(runtime) : 'text-slate-900 dark:text-slate-100'

  return (
    <div className="text-center">
      <div className={`text-3xl sm:text-4xl font-bold ${colorClass}`}>
        {value}
      </div>
      <div className="text-xs sm:text-sm text-slate-600 dark:text-slate-400 uppercase tracking-wider mt-1">
        {label}
      </div>
    </div>
  )
}

export function StatsGrid({ stats }) {
  return (
    <div className="bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg shadow-sm p-4 sm:p-6">
      <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-4 sm:gap-6">
        <StatCard label="Total Apps" value={stats.total} />
        <StatCard label="PHP" value={stats.php} runtime="php" />
        <StatCard label="Node.js" value={stats.node} runtime="node" />
        <StatCard label="Python" value={stats.python} runtime="python" />
        <StatCard label="Go" value={stats.go} runtime="go" />
        <StatCard label=".NET" value={stats.dotnet} runtime="dotnet" />
      </div>
    </div>
  )
}
