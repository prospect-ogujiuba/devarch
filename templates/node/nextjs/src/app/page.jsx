import Link from 'next/link'
import './globals.css'

export default function Home() {
  return (
    <main className="container">
      <section className="hero">
        <h1>Welcome to DevArch</h1>
        <p className="subtitle">Next.js Application</p>
      </section>

      <section className="features">
        <div className="feature-card">
          <h2>Server Components</h2>
          <p>React Server Components for optimal performance</p>
        </div>
        <div className="feature-card">
          <h2>File-based Routing</h2>
          <p>Intuitive routing based on file structure</p>
        </div>
        <div className="feature-card">
          <h2>Static Export</h2>
          <p>Build to public/ directory for DevArch compatibility</p>
        </div>
      </section>

      <div className="actions">
        <Link href="/about" className="button">
          Learn More
        </Link>
        <a
          href={process.env.NEXT_PUBLIC_API_URL || '/api'}
          className="button button-secondary"
        >
          API Documentation
        </a>
      </div>
    </main>
  )
}
