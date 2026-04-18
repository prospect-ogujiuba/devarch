interface ErrorPanelProps {
  title?: string
  message: string
}

export function ErrorPanel({ title = 'Something went wrong', message }: ErrorPanelProps) {
  return (
    <div className="panel panel--error" role="alert">
      <strong>{title}</strong>
      <p>{message}</p>
    </div>
  )
}
