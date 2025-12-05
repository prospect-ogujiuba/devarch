import { useState } from 'react'
import { getContainerStatusBgClass, getCategoryBgClass, formatContainerStatus, formatCategory, getResourceBarColor } from '../utils/containers'
import { ContainerDetailModal } from './ContainerDetailModal'

export function ContainerCard({ container }) {
  const [showDetails, setShowDetails] = useState(false)

  const isRunning = container.status === 'running'

  return (
    <>
      <div className="bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg p-6 shadow-sm hover:shadow-lg transition-all duration-200 hover:-translate-y-1">
        <div className="flex items-start justify-between mb-4">
          <div className="flex-1 mr-3">
            <h3 className="text-xl font-semibold text-slate-900 dark:text-slate-100 font-mono">
              {container.name}
            </h3>
            {container.restartCount > 0 && (
              <span className="inline-block mt-1 px-2 py-0.5 text-xs font-medium bg-orange-100 text-orange-800 dark:bg-orange-900 dark:text-orange-200 rounded">
                {container.restartCount} {container.restartCount === 1 ? 'restart' : 'restarts'}
              </span>
            )}
          </div>
          <div
            className={`w-3 h-3 rounded-full flex-shrink-0 ${getContainerStatusBgClass(container.status)} ring-2 ring-white dark:ring-slate-800`}
            title={`Status: ${formatContainerStatus(container.status)}`}
          />
        </div>

        <div className="flex flex-wrap gap-2 mb-4">
          <span className="inline-block px-3 py-1 rounded text-xs font-semibold bg-slate-100 dark:bg-slate-700 text-slate-700 dark:text-slate-300">
            {container.image}
          </span>
          <span className="inline-block px-2 py-1 rounded text-xs font-semibold bg-slate-200 dark:bg-slate-600 text-slate-600 dark:text-slate-400">
            {container.version}
          </span>
          <span
            className={`inline-block px-3 py-1 rounded text-xs font-semibold text-white uppercase tracking-wide ${getCategoryBgClass(container.category)}`}
          >
            {formatCategory(container.category)}
          </span>
        </div>

        {isRunning && container.cpu !== 'N/A' && (
          <div className="mb-4 space-y-3">
            {/* CPU Progress Bar */}
            <div>
              <div className="flex items-center justify-between text-xs mb-1">
                <span className="text-slate-600 dark:text-slate-400">CPU</span>
                <span className="font-mono text-slate-900 dark:text-slate-100">{container.cpu}</span>
              </div>
              <div className="w-full bg-slate-200 dark:bg-slate-700 rounded-full h-1.5">
                <div
                  className={`${getResourceBarColor(container.cpuPercentage || 0)} h-1.5 rounded-full transition-all duration-300`}
                  style={{ width: `${Math.min(container.cpuPercentage || 0, 100)}%` }}
                />
              </div>
            </div>

            {/* Memory Progress Bar */}
            <div>
              <div className="flex items-center justify-between text-xs mb-1">
                <span className="text-slate-600 dark:text-slate-400">Memory</span>
                <span className="font-mono text-slate-900 dark:text-slate-100">
                  {container.memoryUsedMb}MB / {container.memoryLimitMb}MB
                </span>
              </div>
              <div className="w-full bg-slate-200 dark:bg-slate-700 rounded-full h-1.5">
                <div
                  className={`${getResourceBarColor(container.memoryPercentage || 0)} h-1.5 rounded-full transition-all duration-300`}
                  style={{ width: `${Math.min(container.memoryPercentage || 0, 100)}%` }}
                />
              </div>
            </div>
          </div>
        )}

        {container.ports.length > 0 && (
          <div className="mb-4">
            <div className="text-xs text-slate-600 dark:text-slate-400 mb-1">Ports</div>
            <div className="text-xs font-mono text-slate-900 dark:text-slate-100">
              {container.ports.slice(0, 3).join(', ')}
              {container.ports.length > 3 && ` +${container.ports.length - 3} more`}
            </div>
          </div>
        )}

        <div className="flex gap-2">
          {container.testDomains?.[0] && (
            <a
              href={`https://${container.testDomains[0]}`}
              target="_blank"
              rel="noopener noreferrer"
              className="flex-1 text-center px-4 py-2 bg-slate-900 dark:bg-slate-100 text-white dark:text-slate-900 rounded-lg text-sm font-medium hover:bg-slate-800 dark:hover:bg-slate-200 transition-colors"
            >
              Open .test
            </a>
          )}
          {container.localhostUrls?.[0] && (
            <a
              href={container.localhostUrls[0]}
              target="_blank"
              rel="noopener noreferrer"
              className="flex-1 text-center px-4 py-2 bg-blue-600 dark:bg-blue-500 text-white rounded-lg text-sm font-medium hover:bg-blue-700 dark:hover:bg-blue-600 transition-colors"
            >
              Localhost
            </a>
          )}
          <button
            onClick={() => setShowDetails(true)}
            className="flex-1 text-center px-4 py-2 bg-slate-100 dark:bg-slate-700 text-slate-900 dark:text-slate-100 rounded-lg text-sm font-medium hover:bg-slate-200 dark:hover:bg-slate-600 transition-colors"
          >
            Details
          </button>
        </div>
      </div>

      {showDetails && (
        <ContainerDetailModal
          container={container}
          onClose={() => setShowDetails(false)}
        />
      )}
    </>
  )
}
