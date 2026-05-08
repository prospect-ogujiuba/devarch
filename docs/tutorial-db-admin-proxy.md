# DB admin/proxy stack from scratch

This guide starts at the exact first-run failure:

```console
$ devarch doctor
zsh: command not found: devarch
```

It then uses the current DevArch-native workflow wherever DevArch supports it.

## 1. Run or install the CLI

From the DevArch repository root, you can run the CLI without installing it:

```bash
go run ./cmd/devarch --help
go run ./cmd/devarch doctor
```

To install the `devarch` command:

```bash
go install ./cmd/devarch
export PATH="$HOME/go/bin:$PATH"
command -v devarch
devarch --help
```

If `command -v devarch` prints nothing, add the PATH line to your shell config, for example `~/.zshrc`, then reload your shell:

```bash
echo 'export PATH="$HOME/go/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

Optional user-local symlink:

```bash
mkdir -p "$HOME/.local/bin"
ln -sf "$HOME/go/bin/devarch" "$HOME/.local/bin/devarch"
export PATH="$HOME/.local/bin:$PATH"
```

## 2. Check the local runtime

DevArch workspace operations are Podman-oriented.

```bash
devarch doctor
devarch runtime status
podman ps
```

If the socket is unavailable:

```bash
devarch socket status
devarch socket start
```

## 3. Understand the native flow

The native DevArch flow is:

```txt
devarch.workspace.yaml -> workspace plan -> workspace apply -> workspace status
```

Use DevArch commands for planning, applying, status, logs, exec, and restart.

## 4. Create a first native PostgreSQL workspace

PostgreSQL is available in the builtin catalog.

```bash
mkdir -p ~/devarch-workspaces/db-admin-stack
cd ~/devarch-workspaces/db-admin-stack
```

Create `devarch.workspace.yaml`:

```yaml
apiVersion: devarch.io/v2alpha1
kind: Workspace
metadata:
  name: db-admin-stack
  displayName: DB Admin Stack
  description: Local PostgreSQL workspace managed by DevArch.
runtime:
  provider: auto
  isolatedNetwork: true
  namingStrategy: workspace-resource
catalog:
  sources:
    - /path/to/devarch/catalog/builtin
policies:
  autoWire: true
  secretSource: local
resources:
  postgres:
    template: postgres
    enabled: true
    env:
      POSTGRES_DB: devarch
      POSTGRES_USER: devarch
      POSTGRES_PASSWORD: devarch
    ports:
      - host: 8502
        container: 5432
    volumes:
      - source: workspace.postgres-data
        target: /var/lib/postgresql/data
```

Replace `/path/to/devarch` with your DevArch repository path.

## 5. Plan, apply, and inspect

Global flags go before the command.

```bash
devarch \
  --workspace-root ~/devarch-workspaces \
  --catalog-root /path/to/devarch/catalog/builtin \
  workspace plan db-admin-stack
```

Apply only when the plan looks right:

```bash
devarch \
  --workspace-root ~/devarch-workspaces \
  --catalog-root /path/to/devarch/catalog/builtin \
  workspace apply db-admin-stack
```

Check status:

```bash
devarch \
  --workspace-root ~/devarch-workspaces \
  --catalog-root /path/to/devarch/catalog/builtin \
  workspace status db-admin-stack
```

Use DevArch for logs and exec:

```bash
devarch \
  --workspace-root ~/devarch-workspaces \
  workspace logs db-admin-stack postgres
```

```bash
devarch \
  --workspace-root ~/devarch-workspaces \
  workspace exec db-admin-stack postgres -- psql -U devarch -d devarch -c 'SELECT 1;'
```

## 6. MariaDB, Adminer, and Nginx Proxy Manager status

The service library includes Compose definitions for these services:

```txt
services-library/database/mariadb/compose.yml
services-library/dbms/adminer/compose.yml
services-library/proxy/nginx-proxy-manager/compose.yml
```

Current caveat: these are service-library Compose entries, not first-class builtin v2 templates. DevArch can inspect/import-preview them with a native command:

```bash
devarch import v1-library /path/to/devarch/services-library
```

JSON output is useful when you want to review generated template documents:

```bash
devarch --json import v1-library /path/to/devarch/services-library
```

The import preview suggests template paths such as:

```txt
catalog/imported/database/mariadb/template.yaml
catalog/imported/dbms/adminer/template.yaml
catalog/imported/proxy/nginx-proxy-manager/template.yaml
```

After those templates are materialized into a catalog, the workspace shape becomes:

```yaml
resources:
  mariadb:
    template: mariadb
    enabled: true

  adminer:
    template: adminer
    enabled: true
    dependsOn:
      - mariadb

  nginx-proxy-manager:
    template: nginx-proxy-manager
    enabled: true
```

Then run the same native workflow:

```bash
devarch --workspace-root ~/devarch-workspaces --catalog-root ./catalog/imported workspace plan db-admin-stack
devarch --workspace-root ~/devarch-workspaces --catalog-root ./catalog/imported workspace apply db-admin-stack
devarch --workspace-root ~/devarch-workspaces --catalog-root ./catalog/imported workspace status db-admin-stack
```

Do not use a made-up `devarch add service` command; it does not exist in the current CLI. Use `workspace` commands for native runtime operations and `import v1-library` for migration/preview of service-library Compose entries.
