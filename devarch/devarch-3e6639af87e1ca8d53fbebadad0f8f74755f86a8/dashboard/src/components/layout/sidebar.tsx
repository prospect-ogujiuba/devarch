import { Link, useLocation } from '@tanstack/react-router'
import { X, Server } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { navItems } from './header'
import { useEffect } from 'react'

export function MobileSidebar({ open, onClose }: { open: boolean; onClose: () => void }) {
  const location = useLocation()

  useEffect(() => {
    onClose()
  }, [location.pathname])

  useEffect(() => {
    if (open) {
      document.body.style.overflow = 'hidden'
    } else {
      document.body.style.overflow = ''
    }
    return () => {
      document.body.style.overflow = ''
    }
  }, [open])

  return (
    <>
      {open && (
        <div
          className="fixed inset-0 z-50 bg-black/50 md:hidden"
          onClick={onClose}
        />
      )}

      <aside
        className={`fixed inset-y-0 left-0 z-50 w-72 bg-background border-r transform transition-transform duration-200 ease-in-out md:hidden ${
          open ? 'translate-x-0' : '-translate-x-full'
        }`}
      >
        <div className="flex items-center justify-between h-14 px-6 border-b">
          <Link to="/" className="flex items-center gap-2 font-semibold" onClick={onClose}>
            <Server className="size-5" />
            <span>DevArch</span>
          </Link>
          <Button variant="ghost" size="icon-sm" onClick={onClose}>
            <X className="size-5" />
          </Button>
        </div>

        <nav className="flex flex-col gap-1 p-4">
          {navItems.map(item => {
            const isActive = item.to === '/'
              ? location.pathname === '/'
              : location.pathname.startsWith(item.to)

            return (
              <Link
                key={item.to}
                to={item.to}
                className={`flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-colors ${
                  isActive
                    ? 'bg-accent text-accent-foreground'
                    : 'text-muted-foreground hover:text-foreground hover:bg-accent/50'
                }`}
              >
                <item.icon className="size-5" />
                {item.label}
              </Link>
            )
          })}
        </nav>
      </aside>
    </>
  )
}
