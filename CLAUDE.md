### CLAUDE: Project Operating Guide

This document tells Claude how to interact with and assist on this repository effectively. It defines goals, guardrails, conventions, and quick references tailored to this codebase and environment.

---

### 1) Purpose and Success Criteria

- Primary goal: Help maintain and evolve the DevArch workspace: a containerized developer platform with a React dashboard and many Docker Compose service recipes.
- Succeed by:
  - Producing minimal, correct changes aligned with the project’s structure and conventions.
  - Respecting platform constraints: Windows host, PowerShell shell, WSL paths, and non-destructive defaults.
  - Explaining reasoning succinctly when asked; otherwise keep outputs concise and actionable.

---

### 2) Environment and Constraints

- Host OS: Windows (PowerShell); WSL is available under `\\wsl.localhost\Arch\...` paths.
- Project root: `//wsl.localhost/Arch/home/fhcadmin/projects/devarch`
- Path conventions:
  - When showing terminal commands for the user, prefer PowerShell-compatible syntax.
  - Use Windows backslashes `\` when referencing local Windows paths; use forward slashes for Docker/Linux/WSL examples.
- Git hygiene:
  - `.gitignore` ignores most `apps/**` except `apps/dashboard` and `apps/serverinfo`.
  - Place repository-wide docs (like this file) at the repo root.

---

### 3) Repository Overview (What’s here)

- apps/
  - dashboard/ — React + Vite dashboard UI
    - `src/` components, hooks, pages, utils
    - `api/` simple PHP endpoints for local environment inspection
    - `vite.config.js`, `tailwind.config.js`
- compose/ — Modular Docker Compose definitions by category
  - analytics/ (ELK, Prometheus, Grafana, Matomo, OTEL)
  - backend/ (node, python, go, rust, dotnet, php, vite, celery)
  - database/ (mysql, mariadb, postgres, redis, mongodb, mssql, memcached)
  - dbms/ (pgadmin, phpmyadmin, adminer, metabase, nocodb, mongo-express, cloudbeaver, drawdb)
  - exporters/ (prometheus exporters for various services)
  - mail/ (mailpit)
  - management/ (portainer)
  - messaging/ (kafka, zookeeper, rabbitmq, kafka-ui)
  - project/ (gitea, openproject)
  - proxy/ (nginx-proxy-manager)
  - search/ (meilisearch, typesense)
- config/ — Service configs & Dockerfiles used by compose units
- scripts/ — Utility scripts (auto-discovery; service/runtime management; language app launchers)

Key intent: assemble stacks by combining compose files; operate them with scripts and configs; view/manage via the Dashboard.

---

### 4) How to Run (typical flows)

Note: Tailor commands to the user’s OS/shell. On Windows, use PowerShell. For Docker, commands run inside WSL or PowerShell depending on the user’s setup.

- Dashboard (development):
  1) Navigate to the dashboard app
     - PowerShell: `cd "\\wsl.localhost\Arch\home\fhcadmin\projects\devarch\apps\dashboard"`
     - WSL: `cd /home/fhcadmin/projects/devarch/apps/dashboard`
  2) Install deps: `npm install`
  3) Start dev server: `npm run dev`

  The dashboard is a Vite React app. If an API proxy is configured in `vite.config.js`, it will forward API calls to local PHP endpoints in `apps/dashboard/api/`.

- Launching services with Compose:
  This repo uses many single-purpose compose files you can stack together:
  - Example (PowerShell):
    ```powershell
    cd "\\wsl.localhost\Arch\home\fhcadmin\projects\devarch"
    docker compose -f compose\database\postgres.yml -f compose\dbms\pgadmin.yml up -d
    ```
  - Example (WSL):
    ```bash
    cd /home/fhcadmin/projects/devarch
    docker compose -f compose/database/postgres.yml -f compose/dbms/pgadmin.yml up -d
    ```

  You can combine any number of files to assemble a stack. Shut down with `docker compose down` using the same set of `-f` files.

- Service configs:
  - Service-specific Dockerfiles and configs live under `config/*`. Compose files reference these.

- Utility scripts (bash):
  - `scripts/service-manager.sh` — generic service controls
  - `scripts/runtime-switcher.sh` — runtime selection helper
  - `scripts/*-apps-launcher.sh` — language-specific app runners
  - `scripts/*-auto-discover.sh` — find and manage apps by language

These scripts are bash-oriented; run them inside WSL. If you need Windows-native equivalents, provide PowerShell alternatives or instructions.

---

### 5) Code Pointers (Dashboard)

- UI entry: `apps/dashboard/src/main.jsx`, `src/App.jsx`
- UI components: `src/components/*`
- Hooks: `src/hooks/*` (e.g., `useApps.js`, `useContainers.js`)
- Utilities: `src/utils/*`
- API (local PHP):
  - `apps/dashboard/api/*.php`
  - `apps/dashboard/api/lib/*.php` (helpers; includes `detection.php`, `containers.php`)

When modifying the dashboard, maintain Tailwind and Vite conventions already present.

---

### 6) Operating Rules for Claude

- Keep changes minimal and safe by default; prefer docs/PR suggestions unless explicit permission to refactor.
- Respect Windows + PowerShell command syntax in instructions; use `;` to chain commands in PowerShell.
- When referencing files:
  - Use repo-root relative paths for clarity.
  - Be mindful of `.gitignore` — don’t add ignored files unless the user approves.
- For Docker instructions:
  - Provide both PowerShell (Windows path separators) and WSL/Linux examples when helpful.
  - Avoid destructive operations (`docker system prune -a`) unless requested and confirmed.
- Testing/verification:
  - If changing code, request or run focused checks (e.g., build dashboard with `npm run build`) when the user asks for verification.
- Security/secrets:
  - Do not commit `.env` files or secrets.
  - Assume local-only dev endpoints in `apps/dashboard/api` are not exposed publicly.

---

### 7) Common Tasks Cheat Sheet

- Start a DB + Admin UI (example):
  - PowerShell:
    ```powershell
    docker compose -f compose\database\mariadb.yml -f compose\dbms\phpmyadmin.yml up -d
    ```
  - WSL:
    ```bash
    docker compose -f compose/database/mariadb.yml -f compose/dbms/phpmyadmin.yml up -d
    ```

- Start observability basics (Prometheus + Grafana):
  ```bash
  docker compose \
    -f compose/analytics/prometheus.yml \
    -f compose/analytics/grafana.yml up -d
  ```

- Start Kafka with UI:
  ```bash
  docker compose \
    -f compose/messaging/zookeeper.yml \
    -f compose/messaging/kafka.yml \
    -f compose/messaging/kafka-ui.yml up -d
  ```

- Start a search engine (Typesense or Meilisearch):
  ```bash
  docker compose -f compose/search/typesense.yml up -d
  # or
  docker compose -f compose/search/meilisearch.yml up -d
  ```

---

### 8) Coding Style & Conventions

- Mirror existing patterns:
  - React components: functional components, hooks pattern, Tailwind classes
  - PHP utilities under `apps/dashboard/api/lib`: keep small, procedural helpers consistent with current style
  - Compose files: keep services modular (one purpose per file) and reference configs from `config/`
- Documentation: concise, sectioned, with copy-pasteable commands for both Windows PowerShell and WSL/Linux when relevant.

---

### 9) When to Ask for Clarification

Ask the user before proceeding when:
- Combining many compose files into a large stack (resource-heavy).
- Introducing new services or changing ports/volumes.
- Making non-trivial refactors in dashboard or scripts.
- Handling sensitive data or `.env` values.

---

### 10) Output Formatting (Claude responses)

- Keep answers succinct unless the user asks for detail.
- Use fenced code blocks for commands/configs:
  - Use ```powershell for Windows PowerShell commands
  - Use ```bash for WSL/Linux
- Use inline code for file names, paths, and identifiers.
- Provide step-by-step lists for operational tasks.

---

### 11) Quick File Index (high value targets)

- Dashboard UI: `apps/dashboard/src/*`
- Dashboard API helpers: `apps/dashboard/api/lib/*`
- Compose recipes: `compose/**.yml`
- Service configs/Dockerfiles: `config/**/*`
- Orchestration scripts: `scripts/*.sh`

---

### 12) Maintenance Notes

- Prefer adding new compose files over modifying existing ones when introducing services; it preserves modularity.
- Keep `config/` and `compose/` in sync: if a compose file references a `config/*` path, ensure it exists and is committed.
- For the dashboard, confirm dev server proxies (if any) still point at available endpoints.

---

Authored for Claude on 2025-12-08.
