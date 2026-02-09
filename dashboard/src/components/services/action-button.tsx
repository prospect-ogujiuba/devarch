import { useStartService, useStopService, useRestartService } from '@/features/services/queries'
import { LifecycleButtons } from '@/components/ui/entity-actions'

interface ActionButtonProps {
  name: string
  status: string
  showRestart?: boolean
  size?: 'default' | 'sm' | 'icon' | 'icon-sm'
  className?: string
  buttonClassName?: string
}

export function ActionButton({
  name,
  status,
  showRestart = false,
  size = 'sm',
  className,
  buttonClassName,
}: ActionButtonProps) {
  const startMutation = useStartService()
  const stopMutation = useStopService()
  const restartMutation = useRestartService()

  const isRunning = status === 'running'

  return (
    <LifecycleButtons
      isRunning={isRunning}
      onStart={() => startMutation.mutate(name)}
      onStop={() => stopMutation.mutate(name)}
      onRestart={() => restartMutation.mutate(name)}
      isStartPending={startMutation.isPending}
      isStopPending={stopMutation.isPending}
      isRestartPending={restartMutation.isPending}
      startDisabled={status === 'starting'}
      showRestart={showRestart}
      size={size}
      className={className}
      buttonClassName={buttonClassName}
    />
  )
}
