import { createRootRoute, Outlet } from '@tanstack/react-router'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { Toaster } from '@/components/ui/sonner'
import { ThemeProvider } from '@/lib/theme'
import { Header } from '@/components/layout/header'
import { MobileSidebar } from '@/components/layout/sidebar'
import { useWebSocket } from '@/hooks/use-websocket'
import { useState, useCallback } from 'react'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 5000,
      retry: 1,
    },
  },
})

export const Route = createRootRoute({
  component: RootLayout,
})

function RootLayout() {
  const [sidebarOpen, setSidebarOpen] = useState(false)
  const toggleSidebar = useCallback(() => setSidebarOpen(prev => !prev), [])
  const closeSidebar = useCallback(() => setSidebarOpen(false), [])

  return (
    <ThemeProvider>
      <QueryClientProvider client={queryClient}>
        <WebSocketProvider>
          <div className="min-h-screen">
            <Header onMenuClick={toggleSidebar} />
            <MobileSidebar open={sidebarOpen} onClose={closeSidebar} />
            <main className="container mx-auto px-6 py-6">
              <Outlet />
            </main>
          </div>
          <Toaster richColors />
        </WebSocketProvider>
      </QueryClientProvider>
    </ThemeProvider>
  )
}

function WebSocketProvider({ children }: { children: React.ReactNode }) {
  useWebSocket()
  return <>{children}</>
}
