import { Link } from 'react-router-dom'
import Header from '@components/Header'

function AboutPage() {
  return (
    <div className="page">
      <Header />
      <main className="container">
        <h1>About This Application</h1>
        <p>
          This application was created using the DevArch React + Vite template.
          It follows the standardized public/ directory structure required by
          the DevArch web server configuration.
        </p>

        <h2>Application Structure</h2>
        <pre className="code-block">
{`apps/your-app/
├── public/              # Built assets served by web server
│   ├── index.html
│   ├── assets/         # Compiled JS, CSS, images
│   └── api/            # Optional API endpoints
├── src/                # Source code
│   ├── components/
│   ├── pages/
│   ├── utils/
│   └── styles/
└── vite.config.js      # Build config (outputs to public/)`}
        </pre>

        <h2>Environment Variables</h2>
        <ul>
          <li><strong>App Name:</strong> {import.meta.env.VITE_APP_NAME || 'Not set'}</li>
          <li><strong>Environment:</strong> {import.meta.env.VITE_APP_ENV || 'production'}</li>
          <li><strong>API Base URL:</strong> {import.meta.env.VITE_API_BASE_URL || '/api'}</li>
        </ul>

        <div className="actions">
          <Link to="/" className="button">Back to Home</Link>
        </div>
      </main>
    </div>
  )
}

export default AboutPage
