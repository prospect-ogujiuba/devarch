import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { EnableToggle, LifecycleButtons, MoreActionsMenu } from './entity-actions'
import { DropdownMenuItem } from '@/components/ui/dropdown-menu'

describe('mobile action controls', () => {
  it('applies responsive classes to lifecycle buttons', () => {
    const onStart = vi.fn()
    render(
      <LifecycleButtons
        isRunning={false}
        onStart={onStart}
        onStop={vi.fn()}
        className="col-span-2"
        buttonClassName="w-full sm:w-auto"
      />,
    )

    const startButton = screen.getByRole('button', { name: /start/i })
    expect(startButton).toHaveClass('w-full')
    expect(startButton.parentElement).toHaveClass('col-span-2')
  })

  it('renders mobile label for more actions trigger', () => {
    render(
      <MoreActionsMenu mobileLabel="Actions">
        <DropdownMenuItem>Item</DropdownMenuItem>
      </MoreActionsMenu>,
    )

    expect(screen.getByRole('button', { name: /actions/i })).toBeInTheDocument()
  })

  it('keeps enable toggle interactive with custom width class', async () => {
    const user = userEvent.setup()
    const onToggle = vi.fn()
    render(
      <EnableToggle enabled={false} onToggle={onToggle} isPending={false} className="w-full sm:w-auto" />,
    )

    const button = screen.getByRole('button', { name: /enable/i })
    expect(button).toHaveClass('w-full')
    await user.click(button)
    expect(onToggle).toHaveBeenCalledTimes(1)
  })
})
