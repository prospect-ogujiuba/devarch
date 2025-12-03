import { useEffect } from 'react'
import { getRuntimeBgClass, getStatusBgClass } from '../utils/colors'
import { formatStatus, formatRuntime } from '../utils/formatters'

export function AppDetailModal({ app, onClose }) {
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
            {app.name}
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
          {/* Status */}
          <div>
            <label className="text-sm font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">
              Status
            </label>
            <div className="mt-2 flex items-center gap-2">
              <div className={`w-4 h-4 rounded-full ${getStatusBgClass(app.status)}`} />
              <span className="text-lg font-medium text-slate-900 dark:text-slate-100">
                {formatStatus(app.status)}
              </span>
            </div>
          </div>

          {/* Runtime */}
          <div>
            <label className="text-sm font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">
              Runtime
            </label>
            <div className="mt-2">
              <span className={`inline-block px-4 py-2 rounded-lg text-sm font-semibold text-white uppercase tracking-wide ${getRuntimeBgClass(app.runtime)}`}>
                {formatRuntime(app.runtime)}
              </span>
            </div>
          </div>

          {/* Framework */}
          <div>
            <label className="text-sm font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">
              Framework
            </label>
            <div className="mt-2">
              <span className="inline-block px-4 py-2 rounded-lg text-sm font-semibold bg-slate-100 dark:bg-slate-700 text-slate-700 dark:text-slate-300">
                {app.framework}
              </span>
            </div>
          </div>

          {/* URL */}
          <div>
            <label className="text-sm font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">
              Application URL
            </label>
            <div className="mt-2">
              <a
                href={app.url}
                target="_blank"
                rel="noopener noreferrer"
                className="inline-flex items-center gap-2 text-blue-600 dark:text-blue-400 hover:underline font-mono text-sm"
              >
                {app.url}
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
                </svg>
              </a>
            </div>
          </div>

          {/* Path */}
          <div>
            <label className="text-sm font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">
              File Path
            </label>
            <div className="mt-2">
              <code className="block px-4 py-3 bg-slate-100 dark:bg-slate-900 text-slate-900 dark:text-slate-100 rounded-lg text-sm font-mono break-all">
                {app.path}
              </code>
            </div>
          </div>

          {/* Actions */}
          <div className="pt-4 border-t border-slate-200 dark:border-slate-700">
            <div className="flex gap-3">
              <a
                href={app.url}
                target="_blank"
                rel="noopener noreferrer"
                className="flex-1 text-center px-4 py-3 bg-slate-900 dark:bg-slate-100 text-white dark:text-slate-900 rounded-lg font-medium hover:bg-slate-800 dark:hover:bg-slate-200 transition-colors"
              >
                Open Application
              </a>
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
