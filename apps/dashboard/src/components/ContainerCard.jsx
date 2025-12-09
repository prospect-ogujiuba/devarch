import { useState } from 'react'
import { getContainerStatusBgClass, getCategoryBgClass, formatContainerStatus, formatCategory, getResourceBarColor } from '../utils/containers'
import { ContainerDetailModal } from './ContainerDetailModal'
import { LogsModal } from './LogsModal'
import { useContainerControl } from '../hooks/useContainerControl'

// Icon components
const PlayIcon = () => (
  <svg className="w-3 h-3" fill="currentColor" viewBox="0 0 20 20">
    <path d="M6.3 2.841A1.5 1.5 0 004 4.11V15.89a1.5 1.5 0 002.3 1.269l9.344-5.89a1.5 1.5 0 000-2.538L6.3 2.84z"/>
  </svg>
)

const StopIcon = () => (
  <svg className="w-3 h-3" fill="currentColor" viewBox="0 0 20 20">
    <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8 7a1 1 0 00-1 1v4a1 1 0 001 1h4a1 1 0 001-1V8a1 1 0 00-1-1H8z" clipRule="evenodd"/>
  </svg>
)

const RestartIcon = () => (
  <svg className="w-3 h-3" fill="currentColor" viewBox="0 0 20 20">
    <path fillRule="evenodd" d="M4 2a1 1 0 011 1v2.101a7.002 7.002 0 0111.601 2.566 1 1 0 11-1.885.666A5.002 5.002 0 005.999 7H9a1 1 0 010 2H4a1 1 0 01-1-1V3a1 1 0 011-1zm.008 9.057a1 1 0 011.276.61A5.002 5.002 0 0014.001 13H11a1 1 0 110-2h5a1 1 0 011 1v5a1 1 0 11-2 0v-2.101a7.002 7.002 0 01-11.601-2.566 1 1 0 01.61-1.276z" clipRule="evenodd"/>
  </svg>
)

const WrenchIcon = () => (
  <svg className="w-3 h-3" fill="currentColor" viewBox="0 0 20 20">
    <path fillRule="evenodd" d="M11.3 1.046A1 1 0 0112 2v5h4a1 1 0 01.82 1.573l-7 10A1 1 0 018 18v-5H4a1 1 0 01-.82-1.573l7-10a1 1 0 011.12-.38z" clipRule="evenodd"/>
  </svg>
)

const SpinnerIcon = () => (
  <svg className="w-3 h-3 animate-spin" fill="none" viewBox="0 0 24 24">
    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"/>
    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"/>
  </svg>
)

export function ContainerCard({ container, onRefresh }) {
  const [showDetails, setShowDetails] = useState(false)
  const [showLogs, setShowLogs] = useState(false)
  const [actionLoading, setActionLoading] = useState(null)
  const { controlContainer } = useContainerControl()

  const isRunning = container.status === 'running'

  const handleControl = async (action) => {
    setActionLoading(action)
    try {
      await controlContainer(container.name, action)
      if (onRefresh) {
        setTimeout(() => onRefresh(), 500)
      }
    } catch (err) {
      console.error(`Failed to ${action} container:`, err)
    } finally {
      setActionLoading(null)
    }
  }

  return (
    <>
      <div className="bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg p-6 shadow-sm hover:shadow-lg transition-all duration-200 hover:-translate-y-1 flex flex-col h-full">
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

        {/* .test Domains */}
        {container.testDomains && container.testDomains.length > 0 && (
          <div className="mb-3">
            <div className="text-xs font-semibold text-slate-600 dark:text-slate-400 mb-1.5">
              Domains:
            </div>
            <div className="flex flex-wrap gap-1.5">
              {container.testDomains.map((domain, idx) => (
                <a
                  key={idx}
                  href={`https://${domain}`}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-xs px-2 py-1 bg-slate-200 dark:bg-slate-700 text-slate-700 dark:text-slate-300 rounded hover:bg-slate-300 dark:hover:bg-slate-600 transition-colors"
                >
                  {domain}
                </a>
              ))}
            </div>
          </div>
        )}

        {/* Localhost URLs */}
        {container.localhostUrls && container.localhostUrls.length > 0 && (
          <div className="mb-4">
            <div className="text-xs font-semibold text-slate-600 dark:text-slate-400 mb-1.5">
              Localhost:
            </div>
            <div className="flex flex-wrap gap-1.5">
              {container.localhostUrls.map((url, idx) => (
                <a
                  key={idx}
                  href={url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-xs px-2 py-1 bg-green-100 dark:bg-green-900 text-green-700 dark:text-green-300 rounded hover:bg-green-200 dark:hover:bg-green-800 transition-colors"
                >
                  {url.replace('http://127.0.0.1:', ':')}
                </a>
              ))}
            </div>
          </div>
        )}

        {/* Action Buttons */}
        <div className="mt-auto">
          <div className="flex gap-2">
            <button
              onClick={() => setShowDetails(true)}
              className="flex-1 text-center px-4 py-2 bg-slate-100 dark:bg-slate-700 text-slate-900 dark:text-slate-100 rounded-lg text-sm font-medium hover:bg-slate-200 dark:hover:bg-slate-600 transition-colors"
            >
              Details
            </button>
            <button
              onClick={() => setShowLogs(true)}
              className="flex-1 text-center px-4 py-2 bg-slate-100 dark:bg-slate-700 text-slate-900 dark:text-slate-100 rounded-lg text-sm font-medium hover:bg-slate-200 dark:hover:bg-slate-600 transition-colors"
            >
              Logs
            </button>
          </div>

          {/* Container Controls */}
          <div className="mt-3 flex gap-2">
            {isRunning ? (
              <>
                <button
                  onClick={() => handleControl('stop')}
                  disabled={actionLoading !== null}
                  className="flex-1 flex items-center justify-center gap-1.5 px-3 py-1.5 text-xs font-medium bg-red-500 text-white rounded hover:bg-red-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                >
                  {actionLoading === 'stop' ? <SpinnerIcon /> : <StopIcon />}
                  <span>{actionLoading === 'stop' ? 'Stopping' : 'Stop'}</span>
                </button>
                <button
                  onClick={() => handleControl('restart')}
                  disabled={actionLoading !== null}
                  className="flex-1 flex items-center justify-center gap-1.5 px-3 py-1.5 text-xs font-medium bg-yellow-500 text-white rounded hover:bg-yellow-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                >
                  {actionLoading === 'restart' ? <SpinnerIcon /> : <RestartIcon />}
                  <span>{actionLoading === 'restart' ? 'Restarting' : 'Restart'}</span>
                </button>
                <button
                  onClick={() => handleControl('rebuild')}
                  disabled={actionLoading !== null}
                  className="flex-1 flex items-center justify-center gap-1.5 px-3 py-1.5 text-xs font-medium bg-slate-500 dark:bg-slate-400 text-white rounded hover:bg-slate-600 dark:hover:bg-slate-300 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                >
                  {actionLoading === 'rebuild' ? <SpinnerIcon /> : <WrenchIcon />}
                  <span>{actionLoading === 'rebuild' ? 'Rebuilding' : 'Rebuild'}</span>
                </button>
              </>
            ) : (
              <button
                onClick={() => handleControl('start')}
                disabled={actionLoading !== null}
                className="flex-1 flex items-center justify-center gap-1.5 px-3 py-1.5 text-xs font-medium bg-green-500 text-white rounded hover:bg-green-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
              >
                {actionLoading === 'start' ? <SpinnerIcon /> : <PlayIcon />}
                <span>{actionLoading === 'start' ? 'Starting' : 'Start'}</span>
              </button>
            )}
          </div>
        </div>
      </div>

      {showDetails && (
        <ContainerDetailModal
          container={container}
          onClose={() => setShowDetails(false)}
        />
      )}

      {showLogs && (
        <LogsModal
          container={container}
          onClose={() => setShowLogs(false)}
        />
      )}
    </>
  )
}
