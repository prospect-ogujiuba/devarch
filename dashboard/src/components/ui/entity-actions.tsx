import { type ReactNode } from 'react'
import { Play, Square, RotateCw, Loader2, Power, PowerOff, MoreVertical } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'

type ButtonSize = 'default' | 'sm' | 'icon' | 'icon-sm'

interface LifecycleButtonsProps {
  isRunning: boolean
  onStart: () => void
  onStop: () => void
  onRestart?: () => void
  isPending?: boolean
  isStartPending?: boolean
  isStopPending?: boolean
  isRestartPending?: boolean
  startDisabled?: boolean
  showRestart?: boolean
  showAll?: boolean
  size?: ButtonSize
}

export function LifecycleButtons({
  isRunning,
  onStart,
  onStop,
  onRestart,
  isPending,
  isStartPending,
  isStopPending,
  isRestartPending,
  startDisabled,
  showRestart = false,
  showAll = false,
  size = 'sm',
}: LifecycleButtonsProps) {
  const anyPending = isPending ?? (isStartPending || isStopPending || isRestartPending) ?? false
  const isIcon = size === 'icon' || size === 'icon-sm'

  if (isIcon) {
    if (showRestart && isRunning && onRestart) {
      return (
        <Button
          variant="ghost"
          size={size}
          onClick={onRestart}
          disabled={anyPending}
        >
          {(isPending ?? isRestartPending) ? (
            <Loader2 className="size-4 animate-spin" />
          ) : (
            <RotateCw className="size-4" />
          )}
        </Button>
      )
    }

    return (
      <Button
        variant="ghost"
        size={size}
        onClick={() => isRunning ? onStop() : onStart()}
        disabled={anyPending || startDisabled}
      >
        {anyPending ? (
          <Loader2 className="size-4 animate-spin" />
        ) : isRunning ? (
          <Square className="size-4" />
        ) : (
          <Play className="size-4" />
        )}
      </Button>
    )
  }

  if (showAll) {
    return (
      <div className="flex items-center gap-1">
        <Button
          variant="outline"
          size={size}
          onClick={onStart}
          disabled={anyPending || startDisabled}
        >
          {(isPending ?? isStartPending) ? (
            <Loader2 className="size-4 animate-spin" />
          ) : (
            <Play className="size-4" />
          )}
          Start
        </Button>
        <Button
          variant="outline"
          size={size}
          onClick={onStop}
          disabled={anyPending}
        >
          {(isPending ?? isStopPending) ? (
            <Loader2 className="size-4 animate-spin" />
          ) : (
            <Square className="size-4" />
          )}
          Stop
        </Button>
        {onRestart && (
          <Button
            variant="outline"
            size={size}
            onClick={onRestart}
            disabled={anyPending}
          >
            {(isPending ?? isRestartPending) ? (
              <Loader2 className="size-4 animate-spin" />
            ) : (
              <RotateCw className="size-4" />
            )}
            Restart
          </Button>
        )}
      </div>
    )
  }

  return (
    <div className="flex items-center gap-1">
      {isRunning ? (
        <>
          <Button
            variant="outline"
            size={size}
            onClick={onStop}
            disabled={anyPending}
          >
            {(isPending ?? isStopPending) ? (
              <Loader2 className="size-4 animate-spin" />
            ) : (
              <Square className="size-4" />
            )}
            Stop
          </Button>
          {showRestart && onRestart && (
            <Button
              variant="outline"
              size={size}
              onClick={onRestart}
              disabled={anyPending}
            >
              {(isPending ?? isRestartPending) ? (
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
          onClick={onStart}
          disabled={anyPending || startDisabled}
        >
          {(isPending ?? isStartPending) ? (
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

interface EnableToggleProps {
  enabled: boolean
  onToggle: () => void
  isPending: boolean
  size?: ButtonSize
}

export function EnableToggle({ enabled, onToggle, isPending, size = 'sm' }: EnableToggleProps) {
  return (
    <Button
      variant={enabled ? 'outline' : 'success'}
      size={size}
      onClick={onToggle}
      disabled={isPending}
    >
      {isPending ? (
        <Loader2 className="size-4 animate-spin" />
      ) : enabled ? (
        <>
          <PowerOff className="size-4" />
          Disable
        </>
      ) : (
        <>
          <Power className="size-4" />
          Enable
        </>
      )}
    </Button>
  )
}

interface MoreActionsMenuProps {
  children: ReactNode
  align?: 'start' | 'end' | 'center'
  size?: ButtonSize
  variant?: 'outline' | 'ghost'
  iconClassName?: string
  triggerClassName?: string
  triggerProps?: { onClick?: (e: React.MouseEvent) => void }
  contentProps?: { onClick?: (e: React.MouseEvent) => void }
}

export function MoreActionsMenu({
  children,
  align = 'end',
  size = 'sm',
  variant = 'outline',
  iconClassName = 'size-4',
  triggerClassName,
  triggerProps,
  contentProps,
}: MoreActionsMenuProps) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild {...triggerProps}>
        <Button variant={variant} size={size} className={triggerClassName}>
          <MoreVertical className={iconClassName} />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align={align} {...contentProps}>
        {children}
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
