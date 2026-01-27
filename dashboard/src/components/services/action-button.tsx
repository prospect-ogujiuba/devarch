import { Play, Square, RotateCw, Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useStartService, useStopService, useRestartService } from '@/features/services/queries'

interface ActionButtonProps {
  name: string
  status: 'running' | 'stopped' | 'starting' | 'error'
  showRestart?: boolean
  size?: 'default' | 'sm' | 'icon' | 'icon-sm'
}

export function ActionButton({ name, status, showRestart = false, size = 'sm' }: ActionButtonProps) {
  const startMutation = useStartService()
  const stopMutation = useStopService()
  const restartMutation = useRestartService()

  const isLoading = startMutation.isPending || stopMutation.isPending || restartMutation.isPending
  const isRunning = status === 'running'

  if (size === 'icon' || size === 'icon-sm') {
    if (showRestart && isRunning) {
      return (
        <Button
          variant="ghost"
          size={size}
          onClick={() => restartMutation.mutate(name)}
          disabled={isLoading}
        >
          {restartMutation.isPending ? (
            <Loader2 className="size-4 animate-spin" />
          ) : (
            <RotateCw className="size-4" />
          )}
        </Button>
      )
    }

    return (
      <Button
        variant={isRunning ? 'ghost' : 'ghost'}
        size={size}
        onClick={() => isRunning ? stopMutation.mutate(name) : startMutation.mutate(name)}
        disabled={isLoading || status === 'starting'}
      >
        {isLoading ? (
          <Loader2 className="size-4 animate-spin" />
        ) : isRunning ? (
          <Square className="size-4" />
        ) : (
          <Play className="size-4" />
        )}
      </Button>
    )
  }

  return (
    <div className="flex items-center gap-1">
      {isRunning ? (
        <>
          <Button
            variant="outline"
            size={size}
            onClick={() => stopMutation.mutate(name)}
            disabled={isLoading}
          >
            {stopMutation.isPending ? (
              <Loader2 className="size-4 animate-spin" />
            ) : (
              <Square className="size-4" />
            )}
            Stop
          </Button>
          {showRestart && (
            <Button
              variant="outline"
              size={size}
              onClick={() => restartMutation.mutate(name)}
              disabled={isLoading}
            >
              {restartMutation.isPending ? (
                <Loader2 className="size-4 animate-spin" />
              ) : (
                <RotateCw className="size-4" />
              )}
              Restart
            </Button>
          )}
        </>
      ) : (
        <Button
          variant="default"
          size={size}
          onClick={() => startMutation.mutate(name)}
          disabled={isLoading || status === 'starting'}
        >
          {startMutation.isPending ? (
            <Loader2 className="size-4 animate-spin" />
          ) : (
            <Play className="size-4" />
          )}
          Start
        </Button>
      )}
    </div>
  )
}
