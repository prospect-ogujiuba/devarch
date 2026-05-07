# Small DB tutorial: MariaDB + Adminer

This tutorial demonstrates DevArch end to end with small public images:

- `mariadb:11`
- `adminer:4`

It intentionally avoids custom application containers.

## Goal

Create a temporary DevArch catalog and workspace, plan it, apply it, verify MariaDB/Adminer, then clean it up.

## 1. Create a temporary tutorial catalog

```bash
mkdir -p .model-artifacts/devarch-e2e/catalog/database/mariadb
mkdir -p .model-artifacts/devarch-e2e/catalog/dbms/adminer
mkdir -p .model-artifacts/devarch-e2e/workspaces/tiny-db
```

MariaDB template:

```bash
cat > .model-artifacts/devarch-e2e/catalog/database/mariadb/template.yaml <<'YAML'
apiVersion: devarch.io/v2alpha1
kind: Template
metadata:
  name: mariadb-small
  tags: [database, sql, mariadb]
  description: Small MariaDB database for DevArch smoke tests.
spec:
  runtime:
    image: mariadb:11
  env:
    MARIADB_ROOT_PASSWORD: devarch
    MARIADB_DATABASE: devarch
  ports:
    - container: 3306
  volumes:
    - target: /var/lib/mysql
      kind: data
  exports:
    - contract: mysql
      env:
        DB_HOST: "${resource.host}"
        DB_PORT: "${resource.port.3306}"
        DB_DATABASE: "${env.MARIADB_DATABASE}"
        DB_USERNAME: root
        DB_PASSWORD: "${env.MARIADB_ROOT_PASSWORD}"
  health:
    test: [CMD, healthcheck.sh, --connect, --innodb_initialized]
    interval: 10s
    timeout: 5s
    retries: 5
    startPeriod: 45s
YAML
```

Adminer template:

```bash
cat > .model-artifacts/devarch-e2e/catalog/dbms/adminer/template.yaml <<'YAML'
apiVersion: devarch.io/v2alpha1
kind: Template
metadata:
  name: adminer-small
  tags: [dbms, adminer]
  description: Small Adminer UI for DevArch smoke tests.
spec:
  runtime:
    image: adminer:4
  env:
    ADMINER_DEFAULT_SERVER: devarch-tiny-db-mariadb
    ADMINER_DESIGN: pepa-linha-dark
  ports:
    - container: 8080
  imports:
    - contract: mysql
  exports:
    - contract: http
      env:
        ADMINER_URL: "http://${resource.host}:${resource.port.8080}"
  health:
    test:
      - CMD-SHELL
      - php -r '$c=@fsockopen("localhost",8080);if(!$c)exit(1);'
    interval: 10s
    timeout: 5s
    retries: 5
    startPeriod: 20s
YAML
```

## 2. Create the workspace

```bash
cat > .model-artifacts/devarch-e2e/workspaces/tiny-db/devarch.workspace.yaml <<'YAML'
apiVersion: devarch.io/v2alpha1
kind: Workspace
metadata:
  name: tiny-db
  displayName: Tiny DB Smoke
  description: Small MariaDB plus Adminer DevArch end-to-end scenario.
runtime:
  provider: podman
  isolatedNetwork: true
  namingStrategy: workspace-resource
catalog:
  sources:
    - ../../catalog
policies:
  autoWire: true
  secretSource: local
resources:
  mariadb:
    template: mariadb-small
    enabled: true
    env:
      MARIADB_ROOT_PASSWORD: devarch
      MARIADB_DATABASE: devarch
    volumes:
      - source: workspace.mariadb-data
        target: /var/lib/mysql
    exports:
      - mysql
  adminer:
    template: adminer-small
    enabled: true
    dependsOn:
      - mariadb
    imports:
      - contract: mysql
        from: mariadb
    ports:
      - host: 8082
        container: 8080
YAML
```

## 3. Verify catalog discovery

```bash
devarch \
  --workspace-root .model-artifacts/devarch-e2e/workspaces \
  --catalog-root .model-artifacts/devarch-e2e/catalog \
  catalog list
```

Expected templates:

```txt
adminer-small
mariadb-small
```

## 4. Plan

```bash
devarch \
  --workspace-root .model-artifacts/devarch-e2e/workspaces \
  --catalog-root .model-artifacts/devarch-e2e/catalog \
  workspace plan tiny-db
```

On a clean runtime, expect add actions for:

- `devarch-tiny-db-net`
- `devarch-tiny-db-adminer`
- `devarch-tiny-db-mariadb`

## 5. Apply

```bash
devarch \
  --workspace-root .model-artifacts/devarch-e2e/workspaces \
  --catalog-root .model-artifacts/devarch-e2e/catalog \
  workspace apply tiny-db
```

This creates the Podman network and containers.

## 6. Check status

```bash
devarch \
  --workspace-root .model-artifacts/devarch-e2e/workspaces \
  --catalog-root .model-artifacts/devarch-e2e/catalog \
  workspace status tiny-db
```

You can also check Podman directly:

```bash
podman ps -a --filter label=devarch.workspace=tiny-db --format '{{.Names}} {{.Status}} {{.Ports}}'
```

Wait until both containers are healthy.

## 7. Verify MariaDB

```bash
podman exec devarch-tiny-db-mariadb mariadb-admin ping -uroot -pdevarch
podman exec devarch-tiny-db-mariadb mariadb -uroot -pdevarch -e 'SHOW DATABASES LIKE "devarch";'
```

Expected:

```txt
mysqld is alive
```

and a `devarch` database row.

You can also use the DevArch exec path:

```bash
devarch \
  --workspace-root .model-artifacts/devarch-e2e/workspaces \
  --catalog-root .model-artifacts/devarch-e2e/catalog \
  workspace exec tiny-db mariadb -- mariadb -uroot -pdevarch -e 'SELECT 1;'
```

## 8. Verify Adminer

```bash
curl -fsS -o /dev/null -w '%{http_code} %{content_type}\n' http://127.0.0.1:8082/
```

Expected:

```txt
200 text/html; charset=utf-8
```

Open Adminer at:

```txt
http://127.0.0.1:8082/
```

Login values:

- System: MySQL/MariaDB
- Server: `devarch-tiny-db-mariadb`
- Username: `root`
- Password: `devarch`
- Database: `devarch`

## 9. Clean up

DevArch currently creates/replaces resources but does not expose a dedicated workspace delete command. Clean this tutorial manually:

```bash
podman rm -f devarch-tiny-db-adminer devarch-tiny-db-mariadb
podman network rm devarch-tiny-db-net
podman volume rm workspace.mariadb-data
```

Confirm cleanup:

```bash
podman ps -a --filter label=devarch.workspace=tiny-db
podman network ls --filter name=devarch-tiny-db-net
podman volume ls --filter name=workspace.mariadb-data
```

## Known noisy follow-up plans

After apply, a follow-up `workspace plan` may report `modify` because runtime inspect includes normalized image names, default entrypoints, command, environment, ports, or volumes that differ from the compact desired template. That is an adapter normalization gap, not a failure of this tutorial.
