export function ErrorMessage({ error, onRetry }) {
  return (
    <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-6 text-center">
      <div className="text-4xl mb-4">⚠️</div>
      <h3 className="text-lg font-semibold text-red-900 dark:text-red-200 mb-2">
        Error Loading Applications
      </h3>
      <p className="text-red-700 dark:text-red-300 mb-4">{error}</p>
      {onRetry && (
        <button
          onClick={onRetry}
          className="px-4 py-2 bg-red-600 hover:bg-red-700 text-white rounded-lg transition-colors"
        >
          Try Again
        </button>
      )}
    </div>
  )
}
