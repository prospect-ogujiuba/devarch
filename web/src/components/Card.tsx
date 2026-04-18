import type { HTMLAttributes, ReactNode } from 'react'
import { classNames } from '../lib/utils'

interface CardProps extends HTMLAttributes<HTMLElement> {
  title?: string
  description?: string
  actions?: ReactNode
}

export function Card({ title, description, actions, className, children, ...props }: CardProps) {
  return (
    <section className={classNames('card', className)} {...props}>
      {(title || description || actions) && (
        <header className="card__header">
          <div>
            {title ? <h2 className="card__title">{title}</h2> : null}
            {description ? <p className="card__description">{description}</p> : null}
          </div>
          {actions ? <div className="card__actions">{actions}</div> : null}
        </header>
      )}
      <div className="card__body">{children}</div>
    </section>
  )
}
