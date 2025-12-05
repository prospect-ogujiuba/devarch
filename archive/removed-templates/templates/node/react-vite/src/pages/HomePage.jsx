import { Link } from 'react-router-dom'
import Header from '@components/Header'

function HomePage() {
  return (
    <div className="page">
      <Header />
      <main className="container">
        <section className="hero">
          <h1>Welcome to DevArch</h1>
          <p className="subtitle">
            A containerized microservices development environment
          </p>
        </section>

        <section className="features">
          <div className="feature-card">
            <h2>Standardized Structure</h2>
            <p>All apps follow the public/ directory pattern for consistent web serving.</p>
          </div>
          <div className="feature-card">
            <h2>Multiple Runtimes</h2>
            <p>Support for PHP, Node.js, Python, Go, and .NET applications.</p>
          </div>
          <div className="feature-card">
            <h2>Quick Setup</h2>
            <p>Templates and scaffolding scripts for rapid application development.</p>
          </div>
        </section>

        <div className="actions">
          <Link to="/about" className="button">Learn More</Link>
          <a
            href={import.meta.env.VITE_API_BASE_URL || '/api'}
            className="button button-secondary"
          >
            API Documentation
          </a>
        </div>
      </main>
    </div>
  )
}

export default HomePage
