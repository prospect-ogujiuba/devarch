import { Link } from '@tanstack/react-router'
import { Server, Moon, Sun, Menu } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useTheme } from '@/lib/theme'
import { navItems } from '@/lib/nav-items'

const settingsItem = navItems.find((item) => item.to === '/settings')
const primaryNavItems = navItems.filter((item) => item.to !== '/settings')

export function Header({ onMenuClick }: { onMenuClick?: () => void }) {
  const { resolvedTheme, setTheme } = useTheme()

  return (
    <header className="sticky top-0 z-50 border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div className="flex h-12 items-center gap-3 px-3 sm:h-14 sm:gap-4 sm:px-6">
        <Link to="/" className="flex items-center gap-2 font-semibold">
          <Server className="size-5" />
          <span className="hidden sm:inline">DevArch</span>
        </Link>

        <nav className="hidden md:flex items-center gap-1 ml-6">
          {primaryNavItems.map(item => (
            <Button key={item.to} variant="ghost" size="sm" asChild>
              <Link to={item.to} activeProps={{ className: 'bg-accent' }} activeOptions={item.to === '/' ? { exact: true } : undefined}>
                <item.icon className="size-4" />
                {item.label}
              </Link>
            </Button>
          ))}
        </nav>

        <div className="ml-auto flex items-center gap-2">
          {settingsItem && (
            <Button variant="ghost" size="sm" className="hidden md:inline-flex" asChild>
              <Link to={settingsItem.to} activeProps={{ className: 'bg-accent' }}>
                <settingsItem.icon className="size-4" />
                {settingsItem.label}
              </Link>
            </Button>
          )}
          <Button
            variant="ghost"
            size="icon-sm"
            onClick={() => setTheme(resolvedTheme === 'dark' ? 'light' : 'dark')}
          >
            {resolvedTheme === 'dark' ? (
              <Sun className="size-4" />
            ) : (
              <Moon className="size-4" />
            )}
          </Button>
          <span className="text-muted-foreground/60 text-sm md:hidden">|</span>
          <Button
            variant="ghost"
            size="icon-sm"
            className="md:hidden"
            onClick={onMenuClick}
          >
            <Menu className="size-5" />
          </Button>
        </div>
      </div>
    </header>
  )
}
