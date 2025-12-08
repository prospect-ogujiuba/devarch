export function LoadingSpinner({ size = 'md' }) {
  const sizeClasses = {
    sm: 'w-4 h-4 border-2',
    md: 'w-8 h-8 border-4',
    lg: 'w-12 h-12 border-4',
  }

  return (
    <div className="flex items-center justify-center">
      <div className={`${sizeClasses[size]} border-slate-300 dark:border-slate-600 border-t-slate-700 dark:border-t-slate-300 rounded-full animate-spin`} />
    </div>
  )
}
