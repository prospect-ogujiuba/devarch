import { useState } from 'react'
import { getRuntimeBgClass, getStatusBgClass } from '../utils/colors'
import { formatStatus } from '../utils/formatters'
import { AppDetailModal } from './AppDetailModal'

export function AppCard({ app }) {
  const [showDetails, setShowDetails] = useState(false)

  return (
    <>
      <div className="bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg p-6 shadow-sm hover:shadow-lg transition-all duration-200 hover:-translate-y-1">
        <div className="flex items-start justify-between mb-4">
          <h3 className="text-xl font-semibold text-slate-900 dark:text-slate-100 font-mono">
            {app.name}
          </h3>
          <div
            className={`w-3 h-3 rounded-full ${getStatusBgClass(app.status)} ring-2 ring-white dark:ring-slate-800`}
            title={`Status: ${formatStatus(app.status)}`}
          />
        </div>

        <div className="flex flex-wrap gap-2 mb-4">
          <span
            className={`inline-block px-3 py-1 rounded text-xs font-semibold text-white uppercase tracking-wide ${getRuntimeBgClass(app.runtime)}`}
          >
            {app.runtime}
          </span>
          <span className="inline-block px-3 py-1 rounded text-xs font-semibold bg-slate-100 dark:bg-slate-700 text-slate-700 dark:text-slate-300">
            {app.framework}
          </span>
        </div>

        <div className="mb-4">
          <p className="text-sm text-slate-600 dark:text-slate-400 font-mono break-all">
            {app.url}
          </p>
        </div>

        <div className="flex gap-2">
          <a
            href={app.url}
            target="_blank"
            rel="noopener noreferrer"
            className="flex-1 text-center px-4 py-2 bg-slate-900 dark:bg-slate-100 text-white dark:text-slate-900 rounded-lg text-sm font-medium hover:bg-slate-800 dark:hover:bg-slate-200 transition-colors"
          >
            Open
          </a>
          <button
            onClick={() => setShowDetails(true)}
            className="flex-1 text-center px-4 py-2 bg-slate-100 dark:bg-slate-700 text-slate-900 dark:text-slate-100 rounded-lg text-sm font-medium hover:bg-slate-200 dark:hover:bg-slate-600 transition-colors"
          >
            Details
          </button>
        </div>
      </div>

      {showDetails && (
        <AppDetailModal app={app} onClose={() => setShowDetails(false)} />
      )}
    </>
  )
}
