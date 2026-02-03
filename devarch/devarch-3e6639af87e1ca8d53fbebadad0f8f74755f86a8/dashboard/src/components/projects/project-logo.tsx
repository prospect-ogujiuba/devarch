import { cn } from '@/lib/utils'

const DEVICON_BASE = 'https://cdn.jsdelivr.net/gh/devicons/devicon/icons'

const logoMap: Record<string, string> = {
  'laravel': `${DEVICON_BASE}/laravel/laravel-original.svg`,
  'wordpress': `${DEVICON_BASE}/wordpress/wordpress-plain.svg`,
  'next.js': `${DEVICON_BASE}/nextjs/nextjs-original.svg`,
  'nuxt': `${DEVICON_BASE}/nuxtjs/nuxtjs-original.svg`,
  'react': `${DEVICON_BASE}/react/react-original.svg`,
  'vue': `${DEVICON_BASE}/vuejs/vuejs-original.svg`,
  'angular': `${DEVICON_BASE}/angularjs/angularjs-original.svg`,
  'svelte': `${DEVICON_BASE}/svelte/svelte-original.svg`,
  'express': `${DEVICON_BASE}/express/express-original.svg`,
  'fastify': `${DEVICON_BASE}/fastify/fastify-original.svg`,
  'nestjs': `${DEVICON_BASE}/nestjs/nestjs-original.svg`,
  'django': `${DEVICON_BASE}/django/django-plain.svg`,
  'flask': `${DEVICON_BASE}/flask/flask-original.svg`,
  'fastapi': `${DEVICON_BASE}/fastapi/fastapi-original.svg`,
  'gin': `${DEVICON_BASE}/go/go-original-wordmark.svg`,
  'echo': `${DEVICON_BASE}/go/go-original-wordmark.svg`,
  'fiber': `${DEVICON_BASE}/go/go-original-wordmark.svg`,
  'chi': `${DEVICON_BASE}/go/go-original-wordmark.svg`,
  'actix web': `${DEVICON_BASE}/rust/rust-original.svg`,
  'axum': `${DEVICON_BASE}/rust/rust-original.svg`,
  'rocket': `${DEVICON_BASE}/rust/rust-original.svg`,
  'remix': `${DEVICON_BASE}/remix/remix-original.svg`,
  'gatsby': `${DEVICON_BASE}/gatsby/gatsby-original.svg`,
  'go': `${DEVICON_BASE}/go/go-original-wordmark.svg`,
  'rust': `${DEVICON_BASE}/rust/rust-original.svg`,
  'python': `${DEVICON_BASE}/python/python-original.svg`,
  'php': `${DEVICON_BASE}/php/php-original.svg`,
  'javascript': `${DEVICON_BASE}/javascript/javascript-original.svg`,
  'typescript': `${DEVICON_BASE}/typescript/typescript-original.svg`,
  'node': `${DEVICON_BASE}/nodejs/nodejs-original.svg`,
}

function getLogoUrl(projectType: string, framework?: string, language?: string): string | null {
  if (framework) {
    const fwLower = framework.toLowerCase()
    if (logoMap[fwLower]) return logoMap[fwLower]
    const firstWord = fwLower.split(' ')[0]
    if (logoMap[firstWord]) return logoMap[firstWord]
  }
  if (logoMap[projectType]) return logoMap[projectType]
  if (language && logoMap[language]) return logoMap[language]
  return null
}

const typeColors: Record<string, string> = {
  laravel: 'text-red-500',
  wordpress: 'text-blue-500',
  node: 'text-green-500',
  go: 'text-cyan-500',
  rust: 'text-orange-500',
  python: 'text-yellow-500',
  php: 'text-purple-500',
}

interface ProjectLogoProps {
  projectType: string
  framework?: string
  language?: string
  className?: string
}

export function ProjectLogo({ projectType, framework, language, className }: ProjectLogoProps) {
  const url = getLogoUrl(projectType, framework, language)

  if (url) {
    return (
      <img
        src={url}
        alt={framework || projectType}
        className={cn('size-6 object-contain', className)}
        onError={(e) => {
          const target = e.currentTarget
          target.style.display = 'none'
          if (target.nextElementSibling) {
            target.nextElementSibling.classList.remove('hidden')
          }
        }}
      />
    )
  }

  return (
    <span className={cn(
      'size-6 rounded flex items-center justify-center text-xs font-bold bg-muted',
      typeColors[projectType] || 'text-muted-foreground',
      className,
    )}>
      {(framework || projectType).charAt(0).toUpperCase()}
    </span>
  )
}
