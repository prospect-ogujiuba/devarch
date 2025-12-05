import { ThemeToggle } from './ThemeToggle'
import { formatTime } from '../utils/formatters'

export function Header({ activeTab, onTabChange, lastUpdated, onRefresh, isRefreshing }) {
  return (
    <header className="bg-white dark:bg-slate-800 border-b border-slate-200 dark:border-slate-700 shadow-sm">
      <div className="max-w-[1400px] mx-auto px-4 sm:px-6 lg:px-8 py-4 sm:py-6">
        <div className="flex items-center justify-between mb-4">
          <div>
            <h1 className="text-2xl sm:text-3xl font-bold text-slate-900 dark:text-slate-100">
              DevArch Dashboard
            </h1>
            <p className="text-sm text-slate-600 dark:text-slate-400 mt-1">
              Application runtime detection and container monitoring
            </p>
          </div>
          <div className="flex items-center gap-3">
            {lastUpdated && (
              <div className="hidden sm:block text-xs text-slate-500 dark:text-slate-400">
                Last updated: {formatTime(lastUpdated)}
              </div>
            )}
            <button
              onClick={onRefresh}
              disabled={isRefreshing}
              className="p-2 rounded-lg bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
              aria-label="Refresh data"
              title="Refresh application data"
            >
              <svg
                className={`w-5 h-5 text-slate-700 dark:text-slate-300 ${isRefreshing ? 'animate-spin' : ''}`}
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
                />
              </svg>
            </button>
            <ThemeToggle />
          </div>
        </div>

        {/* Tab navigation */}
        <nav className="flex gap-2">
          <button
            onClick={() => onTabChange('apps')}
            className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
              activeTab === 'apps'
                ? 'bg-slate-900 dark:bg-slate-100 text-white dark:text-slate-900'
                : 'bg-slate-100 dark:bg-slate-700 text-slate-700 dark:text-slate-300 hover:bg-slate-200 dark:hover:bg-slate-600'
            }`}
          >
            Applications
          </button>
          <button
            onClick={() => onTabChange('containers')}
            className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
              activeTab === 'containers'
                ? 'bg-slate-900 dark:bg-slate-100 text-white dark:text-slate-900'
                : 'bg-slate-100 dark:bg-slate-700 text-slate-700 dark:text-slate-300 hover:bg-slate-200 dark:hover:bg-slate-600'
            }`}
          >
            Containers
          </button>
        </nav>
      </div>
    </header>
  )
}
