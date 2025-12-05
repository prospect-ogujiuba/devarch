import Link from 'next/link'

export const metadata = {
  title: 'About - DevArch Next.js',
  description: 'About this DevArch Next.js application',
}

export default function AboutPage() {
  return (
    <main className="container">
      <h1>About This Application</h1>
      <p>
        This application was created using the DevArch Next.js template.
        It follows the standardized public/ directory structure required by
        the DevArch web server configuration.
      </p>

      <h2>Application Structure</h2>
      <pre className="code-block">
{`apps/your-app/
├── public/              # Built assets served by web server
│   ├── .next/          # Next.js build output
│   └── api/            # Optional API endpoints
├── src/
│   └── app/            # App router pages
│       ├── page.jsx    # Home page
│       └── about/
│           └── page.jsx
└── next.config.js      # Build config (outputs to public/)`}
      </pre>

      <h2>Build Configuration</h2>
      <p>
        The next.config.js is configured to build static exports to the public/ directory:
      </p>
      <pre className="code-block">
{`{
  distDir: 'public/.next',
  output: 'export',
  images: { unoptimized: true }
}`}
      </pre>

      <div className="actions">
        <Link href="/" className="button">
          Back to Home
        </Link>
      </div>
    </main>
  )
}
