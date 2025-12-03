import './globals.css'

export const metadata = {
  title: 'DevArch Next.js App',
  description: 'A Next.js application in DevArch',
}

export default function RootLayout({ children }) {
  return (
    <html lang="en">
      <body>
        <header className="header">
          <nav className="nav container">
            <a href="/" className="logo">DevArch</a>
            <ul className="nav-links">
              <li><a href="/">Home</a></li>
              <li><a href="/about">About</a></li>
            </ul>
          </nav>
        </header>
        {children}
      </body>
    </html>
  )
}
