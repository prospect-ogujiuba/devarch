# JavaScript framework app scaffolding

`scripts/javascript/bootstrap.sh` creates a fresh application under `apps/<app-name>` from a curated **framework + project profile** combination. Profiles invoke each framework's official `@latest` scaffolder rather than copying DevArch templates, then add the `devarch` package script required by the shared Node runtime.

## Quick start

```bash
scripts/javascript/bootstrap.sh --list-frameworks
scripts/javascript/bootstrap.sh --list-profiles --framework next
scripts/javascript/bootstrap.sh storefront --framework next --profile fullstack --dry-run
scripts/javascript/bootstrap.sh storefront --framework next --profile fullstack --start
# open https://storefront.test
```

The app name must be lowercase DNS-safe text, such as `storefront` or `admin-ui`. Creation uses npm and requires Node/npm on the host. Generated apps run in DevArch's isolated Node 22 container through [`scripts/node/bootstrap.sh`](../node/README.md).

## Curated combinations

| Framework | Profile | Best suited to |
| --- | --- | --- |
| `angular` | `spa` (default) | Strict client-rendered application with routing and SCSS |
| `angular` | `ssr` | Server rendering and hydration |
| `astro` | `minimal` (default) | Lean content or marketing site |
| `astro` | `blog` | Blog with content collections and example pages |
| `next` | `fullstack` (default) | App Router application with Tailwind CSS and server capabilities |
| `next` | `minimal` | Empty App Router project without a CSS framework |
| `next` | `api` | Route-handler-only HTTP API |
| `nuxt` | `minimal` (default) | Universal Vue application without selected modules |
| `nuxt` | `content` | Documentation, editorial, or content-driven site |
| `qwik` | `app` (default) | Lean resumable application |
| `qwik` | `playground` | Evaluation, learning, and examples |
| `qwik` | `library` | Publishable Qwik component package |
| `react-router` | `fullstack` (default) | Maintained full-stack framework template |
| `react-router` | `custom-server` | Application needing a custom Node server |
| `sveltekit` | `minimal` (default) | Lean TypeScript application |
| `sveltekit` | `tested` | Application preconfigured with ESLint, Prettier, Vitest, and Playwright |
| `sveltekit` | `library` | Publishable Svelte component package |
| `vite-lit` | `typescript` (default), `javascript` | Lit web-component application |
| `vite-preact` | `typescript` (default), `javascript` | Lightweight Preact SPA |
| `vite-react` | `typescript` (default), `javascript` | Conventional React SPA |
| `vite-react` | `compiler` | React TypeScript SPA with React Compiler enabled |
| `vite-solid` | `typescript` (default), `javascript` | Fine-grained Solid SPA |
| `vite-vue` | `typescript` (default), `javascript` | Vue SPA |

The matrix is intentionally curated rather than a Cartesian product. Only combinations with a meaningful upstream template, mode, module, or add-on are included.

## Compatibility aliases

The original framework-only form remains available and selects the framework's documented default:

```bash
# Equivalent commands
scripts/javascript/bootstrap.sh dashboard --profile vite-react
scripts/javascript/bootstrap.sh dashboard --framework vite-react --profile typescript
```

New automation should use the explicit framework + profile form.

## Safety and lifecycle

```bash
# Preview without running npm or changing files
scripts/javascript/bootstrap.sh dashboard --framework vite-react --profile compiler --dry-run

# Create now, start later
scripts/javascript/bootstrap.sh dashboard --framework astro --profile blog
scripts/node/bootstrap.sh dashboard

# Replace only after backing up the existing app
scripts/javascript/bootstrap.sh dashboard --framework nuxt --profile content --force

# Start without changing the hosts file
scripts/javascript/bootstrap.sh dashboard --framework sveltekit --profile tested --start --no-hosts
```

An existing target is never changed unless `--force` is supplied. Forced replacement moves it to `apps/.devarch-backups/<name>-<timestamp>` before installing the new app. Scaffolding happens in a temporary directory and is moved into place only after a valid `package.json` exists.

The command removes a scaffolded `.git` directory because `apps/` workspaces are managed separately. Initialize a repository in the generated app when desired.

## Adding a framework or profile

Each framework has a directory under `scripts/javascript/profiles/`. Its `framework.conf` declares listing metadata and the default compatibility profile:

```bash
# scripts/javascript/profiles/vite-react/framework.conf
FRAMEWORK_DESCRIPTION='Vite with React'
DEFAULT_PROFILE='typescript'
```

Project profiles are auditable shell data files named `<profile>.profile`:

```bash
# scripts/javascript/profiles/vite-react/typescript.profile
PROFILE_DESCRIPTION='React SPA with TypeScript'
SCAFFOLD=(npm create vite@latest -- __APP__ --template react-ts --no-interactive)
DEVARCH_SCRIPT='vite --host 0.0.0.0 --port 3000'
POST_INSTALL=1
```

- `SCAFFOLD` is an argv array, not an evaluated command string. `__APP__` becomes the staging directory name.
- `DEVARCH_SCRIPT` must bind the development server to `0.0.0.0:3000`.
- `POST_INSTALL=1` runs `npm install`; use `0` when the upstream scaffolder installs dependencies.
- A profile may define `configure_app APP_DIR APP_NAME` for narrow runtime integration such as Next.js development-origin configuration.

Adding a valid profile file automatically exposes it through `--list-profiles`; bootstrap control flow does not change. Because upstream CLIs evolve, update only the affected profile when a flag changes.

## Tests

```bash
bash scripts/javascript/bootstrap.test.sh
bash scripts/node/bootstrap.test.sh
```
