import { useEffect, useState } from 'react'
import { getContainerStatusBgClass, getCategoryBgClass, formatContainerStatus, formatCategory, getHealthStatusBgClass, formatHealthStatus, getResourceBarColor } from '../utils/containers'

export function ContainerDetailModal({ container, onClose }) {
  const [copiedText, setCopiedText] = useState(null)

  const handleCopy = async (text) => {
    try {
      await navigator.clipboard.writeText(text)
      setCopiedText(text)
      setTimeout(() => setCopiedText(null), 2000)
    } catch (err) {
      console.error('Failed to copy:', err)
    }
  }
  // Close modal on escape key
  useEffect(() => {
    const handleEscape = (e) => {
      if (e.key === 'Escape') onClose()
    }
    window.addEventListener('keydown', handleEscape)
    return () => window.removeEventListener('keydown', handleEscape)
  }, [onClose])

  // Prevent body scroll when modal is open
  useEffect(() => {
    document.body.style.overflow = 'hidden'
    return () => {
      document.body.style.overflow = 'unset'
    }
  }, [])

  const isRunning = container.status === 'running'

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm"
      onClick={onClose}
    >
      <div
        className="bg-white dark:bg-slate-800 rounded-xl shadow-2xl max-w-2xl w-full max-h-[90vh] overflow-y-auto"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="sticky top-0 bg-white dark:bg-slate-800 border-b border-slate-200 dark:border-slate-700 px-6 py-4 flex items-center justify-between">
          <h2 className="text-2xl font-bold text-slate-900 dark:text-slate-100 font-mono">
            {container.name}
          </h2>
          <button
            onClick={onClose}
            className="p-2 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-700 transition-colors"
            aria-label="Close modal"
          >
            <svg className="w-6 h-6 text-slate-500 dark:text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <div className="p-6 space-y-6">
          {/* Container ID */}
          <div>
            <label className="text-sm font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">
              Container ID
            </label>
            <div className="mt-2">
              <code className="block px-4 py-3 bg-slate-100 dark:bg-slate-900 text-slate-900 dark:text-slate-100 rounded-lg text-sm font-mono break-all">
                {container.id}
              </code>
            </div>
          </div>

          {/* Status */}
          <div>
            <label className="text-sm font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">
              Status
            </label>
            <div className="mt-2 flex items-center gap-2">
              <div className={`w-4 h-4 rounded-full ${getContainerStatusBgClass(container.status)}`} />
              <span className="text-lg font-medium text-slate-900 dark:text-slate-100">
                {formatContainerStatus(container.status)}
              </span>
            </div>
          </div>

          {/* Uptime */}
          {container.uptime && (
            <div>
              <label className="text-sm font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">
                Uptime
              </label>
              <div className="mt-2">
                <span className="text-lg font-medium text-slate-900 dark:text-slate-100">
                  {container.uptime}
                </span>
              </div>
            </div>
          )}

          {/* Health Status */}
          <div>
            <label className="text-sm font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">
              Health Status
            </label>
            <div className="mt-2 flex items-center gap-2">
              <div className={`w-4 h-4 rounded-full ${getHealthStatusBgClass(container.healthStatus)}`} />
              <span className="text-lg font-medium text-slate-900 dark:text-slate-100">
                {formatHealthStatus(container.healthStatus)}
              </span>
            </div>
          </div>

          {/* Restart Count */}
          <div>
            <label className="text-sm font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">
              Restart Count
            </label>
            <div className="mt-2">
              <span className={`text-lg font-medium ${container.restartCount > 0 ? 'text-orange-600' : 'text-slate-900 dark:text-slate-100'}`}>
                {container.restartCount || 0}
              </span>
            </div>
          </div>

          {/* Image */}
          <div>
            <label className="text-sm font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">
              Image
            </label>
            <div className="mt-2 flex flex-wrap gap-2">
              <span className="inline-block px-4 py-2 rounded-lg text-sm font-semibold bg-slate-100 dark:bg-slate-700 text-slate-700 dark:text-slate-300">
                {container.image}:{container.version}
              </span>
            </div>
          </div>

          {/* Category */}
          <div>
            <label className="text-sm font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">
              Category
            </label>
            <div className="mt-2">
              <span className={`inline-block px-4 py-2 rounded-lg text-sm font-semibold text-white uppercase tracking-wide ${getCategoryBgClass(container.category)}`}>
                {formatCategory(container.category)}
              </span>
            </div>
          </div>

          {/* Resource Usage */}
          {isRunning && container.cpu !== 'N/A' && (
            <div>
              <label className="text-sm font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">
                Resource Usage
              </label>
              <div className="mt-2 space-y-3">
                {/* CPU with Progress Bar */}
                <div>
                  <div className="flex items-center justify-between mb-1">
                    <span className="text-sm text-slate-600 dark:text-slate-400">CPU</span>
                    <span className="font-mono text-sm font-medium text-slate-900 dark:text-slate-100">{container.cpu}</span>
                  </div>
                  <div className="w-full bg-slate-200 dark:bg-slate-700 rounded-full h-2">
                    <div
                      className={`${getResourceBarColor(container.cpuPercentage || 0)} h-2 rounded-full transition-all duration-300`}
                      style={{ width: `${Math.min(container.cpuPercentage || 0, 100)}%` }}
                    />
                  </div>
                </div>

                {/* Memory with Progress Bar */}
                <div>
                  <div className="flex items-center justify-between mb-1">
                    <span className="text-sm text-slate-600 dark:text-slate-400">Memory</span>
                    <span className="font-mono text-sm font-medium text-slate-900 dark:text-slate-100">
                      {container.memoryUsedMb}MB / {container.memoryLimitMb}MB ({container.memoryPercentage}%)
                    </span>
                  </div>
                  <div className="w-full bg-slate-200 dark:bg-slate-700 rounded-full h-2">
                    <div
                      className={`${getResourceBarColor(container.memoryPercentage || 0)} h-2 rounded-full transition-all duration-300`}
                      style={{ width: `${Math.min(container.memoryPercentage || 0, 100)}%` }}
                    />
                  </div>
                </div>

                {/* Network I/O */}
                {container.network !== 'N/A' && (
                  <div className="flex items-center justify-between p-3 bg-slate-100 dark:bg-slate-900 rounded-lg">
                    <span className="text-sm text-slate-600 dark:text-slate-400">Network I/O</span>
                    <span className="font-mono text-sm font-medium text-slate-900 dark:text-slate-100">{container.network}</span>
                  </div>
                )}
              </div>
            </div>
          )}

          {/* Port Mappings */}
          {container.ports.length > 0 && (
            <div>
              <label className="text-sm font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">
                Port Mappings
              </label>
              <div className="mt-2">
                <div className="bg-slate-100 dark:bg-slate-900 rounded-lg p-3">
                  {container.ports.map((port, index) => (
                    <div key={index} className="font-mono text-sm text-slate-900 dark:text-slate-100">
                      {port}
                    </div>
                  ))}
                </div>
              </div>
            </div>
          )}

          {/* .test Domains */}
          {container.testDomains?.length > 0 && (
            <div>
              <label className="text-sm font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">
                .test Domains
              </label>
              <div className="mt-2 space-y-2">
                {container.testDomains.map((domain, index) => (
                  <div key={index} className="flex items-center gap-2">
                    <a
                      href={`https://${domain}`}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="flex-1 px-4 py-2 bg-slate-100 dark:bg-slate-900 text-blue-600 dark:text-blue-400 hover:underline font-mono text-sm rounded-lg"
                    >
                      {domain}
                    </a>
                    <button
                      onClick={() => handleCopy(`https://${domain}`)}
                      className="px-3 py-2 bg-slate-200 dark:bg-slate-700 hover:bg-slate-300 dark:hover:bg-slate-600 rounded-lg transition-colors"
                      title="Copy URL"
                    >
                      {copiedText === `https://${domain}` ? (
                        <svg className="w-4 h-4 text-green-600 dark:text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                        </svg>
                      ) : (
                        <svg className="w-4 h-4 text-slate-600 dark:text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                        </svg>
                      )}
                    </button>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Localhost URLs */}
          {container.localhostUrls?.length > 0 && (
            <div>
              <label className="text-sm font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">
                Localhost URLs
              </label>
              <div className="mt-2 space-y-2">
                {container.localhostUrls.map((url, index) => (
                  <div key={index} className="flex items-center gap-2">
                    <a
                      href={url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="flex-1 px-4 py-2 bg-slate-100 dark:bg-slate-900 text-blue-600 dark:text-blue-400 hover:underline font-mono text-sm rounded-lg"
                    >
                      {url}
                    </a>
                    <button
                      onClick={() => handleCopy(url)}
                      className="px-3 py-2 bg-slate-200 dark:bg-slate-700 hover:bg-slate-300 dark:hover:bg-slate-600 rounded-lg transition-colors"
                      title="Copy URL"
                    >
                      {copiedText === url ? (
                        <svg className="w-4 h-4 text-green-600 dark:text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                        </svg>
                      ) : (
                        <svg className="w-4 h-4 text-slate-600 dark:text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                        </svg>
                      )}
                    </button>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Actions */}
          <div className="pt-4 border-t border-slate-200 dark:border-slate-700">
            <div className="flex gap-3">
              {(container.testDomains?.[0] || container.localhostUrls?.[0]) && (
                <a
                  href={container.testDomains?.[0] ? `https://${container.testDomains[0]}` : container.localhostUrls[0]}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="flex-1 text-center px-4 py-3 bg-slate-900 dark:bg-slate-100 text-white dark:text-slate-900 rounded-lg font-medium hover:bg-slate-800 dark:hover:bg-slate-200 transition-colors"
                >
                  Open Service
                </a>
              )}
              <button
                onClick={onClose}
                className="flex-1 text-center px-4 py-3 bg-slate-100 dark:bg-slate-700 text-slate-900 dark:text-slate-100 rounded-lg font-medium hover:bg-slate-200 dark:hover:bg-slate-600 transition-colors"
              >
                Close
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
