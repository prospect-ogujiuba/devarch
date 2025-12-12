# devarch - Context Index
Generated: Fri Dec 12 15:09:02 EST 2025

## Git Status
Branch: dev

Recent commits:
- 10320c9: dashboard
- f0db733: route api to devarch
- e75c681: mount workspace and start server command
- 87fe722: move api to config/devarch
- cf9050d: update scripts config
- cc57b83: add new container configs
- 68f0e41: add new containers compose
- 8612274: dashboard
- 648cd5a: apps
- 330e2d3: split krakend and krakend-designer
- 7c5314b: remove broken health check
- c17f143: add case for static "node" applications
- 07f78aa: create non-root user in node
- 51b8b7a: devarch port 8200 -> 8500
- d6e6890: rework node main app port and pnpm_store
- 09bd3a7: dashboard
- 879766b: dashboard update
- f6b0b93: update laravel script to use composer
- 45bfc59: changes to php and devarch containers and nginx config
- 9feb009: update scripts config

Working directory:
Has uncommitted changes:
M CLAUDE.md
M apps/dashboard

Remote info:
Origin: https://github.com/prospect-ogujiuba/devarch.git
Sync status: Up to date with origin

Latest tag: No tags
Stashes: None

## Environment Configuration
Environment file: .env

Environment variables:
```
# =============================================================================
# MICROSERVICES ENVIRONMENT CONFIGURATION - HYBRID APPROACH
# =============================================================================
# This file contains ONLY sensitive data and truly shared values
# Service-specific non-sensitive configs are in individual compose files

# =============================================================================
# GLOBAL ADMIN CONFIGURATION
# =============================================================================
ADMIN_USER=admin
ADMIN_PASSWORD=***masked***
ADMIN_EMAIL=admin@devarch.test

# =============================================================================
# DOMAIN CONFIGURATION
# =============================================================================
DOMAIN_SUFFIX=test

# =============================================================================
# DEVELOPMENT CREDENTIALS
# =============================================================================
GITHUB_USER=prospect-ogujiuba
GITHUB_TOKEN=***masked***

# =============================================================================
# MARIADB CREDENTIALS
# =============================================================================

MARIADB_HOST=mariadb
MARIADB_PORT=3306
MYSQL_ROOT_PASSWORD=***masked***
MYSQL_USER=mariadb_user
MYSQL_PASSWORD=***masked***

# =============================================================================
# NGINX PROXY MANAGER CREDENTIALS
# =============================================================================

DB_MYSQL_HOST=mariadb
DB_MYSQL_PORT=3306
DB_MYSQL_USER=root
DB_MYSQL_PASSWORD=***masked***
DB_MYSQL_NAME=npm
INITIAL_ADMIN_EMAIL=admin@devarch.test
INITIAL_ADMIN_PASSWORD=***masked***
```

Statistics: 18 total variables, 6 sensitive values masked

## Project Structure
.
./apps
./compose
./config
./context
./scripts
apps/
├── dashboard
├── serverinfo

## Context Files
- compose.txt - Contents of compose/ directory
- config.txt - Contents of config/ directory
- scripts.txt - Contents of scripts/ directory

## Summary
- Total files processed: 233
- Total context size: 413KB
- Folders processed: compose config scripts

# devarch - compose
Generated: Fri Dec 12 15:09:03 EST 2025
Folder: compose

## Folder Structure
- compose/analytics/elasticsearch.yml
- compose/analytics/grafana.yml
- compose/analytics/jaeger.yml
- compose/analytics/kibana.yml
- compose/analytics/logstash.yml
- compose/analytics/loki.yml
- compose/analytics/matomo.yml
- compose/analytics/otel-collector.yml
- compose/analytics/prometheus.yml
- compose/analytics/tempo.yml
- compose/analytics/victoriametrics.yml
- compose/analytics/zipkin.yml
- compose/backend/bun.yml
- compose/backend/deno.yml
- compose/backend/dotnet.yml
- compose/backend/elixir.yml
- compose/backend/go.yml
- compose/backend/java.yml
- compose/backend/node.yml
- compose/backend/php.yml
- compose/backend/python.yml
- compose/backend/rust.yml
- compose/backend/vite.yml
- compose/backend/zig.yml
- compose/ci/concourse-web.yml
- compose/ci/concourse-worker.yml
- compose/ci/drone-runner.yml
- compose/ci/drone-server.yml
- compose/ci/gitlab-runner.yml
- compose/ci/jenkins.yml
- compose/ci/woodpecker-agent.yml
- compose/ci/woodpecker-server.yml
- compose/collaboration/element-web.yml
- compose/collaboration/matrix-synapse.yml
- compose/collaboration/mattermost.yml
- compose/collaboration/nextcloud.yml
- compose/collaboration/rocketchat.yml
- compose/collaboration/zulip.yml
- compose/database/cassandra.yml
- compose/database/clickhouse.yml
- compose/database/cockroachdb.yml
- compose/database/couchdb.yml
- compose/database/edgedb.yml
- compose/database/mariadb.yml
- compose/database/memcached.yml
- compose/database/mongodb.yml
- compose/database/mssql.yml
- compose/database/mysql.yml
- compose/database/neo4j.yml
- compose/database/postgres.yml
- compose/database/redis.yml
- compose/database/surrealdb.yml
- compose/dbms/adminer.yml
- compose/dbms/beekeeper-studio.yml
- compose/dbms/cloudbeaver.yml
- compose/dbms/dbeaver.yml
- compose/dbms/drawdb.yml
- compose/dbms/memcached-admin.yml
- compose/dbms/metabase.yml
- compose/dbms/mongo-express.yml
- compose/dbms/nocodb.yml
- compose/dbms/pgadmin.yml
- compose/dbms/phpmyadmin.yml
- compose/dbms/redis-commander.yml
- compose/dbms/sqlpad.yml
- compose/docs/README.md
- compose/docs/bookstack.yml
- compose/docs/docusaurus.yml
- compose/docs/outline.yml
- compose/docs/wikijs.yml
- compose/exporters/blackbox-exporter.yml
- compose/exporters/kafka-exporter.yml
- compose/exporters/memcached-exporter.yml
- compose/exporters/mongodb-exporter.yml
- compose/exporters/mysqld-exporter.yml
- compose/exporters/node-exporter.yml
- compose/exporters/postgres-exporter.yml
- compose/exporters/rabbitmq-exporter.yml
- compose/exporters/redis-exporter.yml
- compose/gateway/apisix.yml
- compose/gateway/envoy.yml
- compose/gateway/gravitee.yml
- compose/gateway/kong.yml
- compose/gateway/krakend-designer.yml
- compose/gateway/krakend.yml
- compose/gateway/traefik.yml
- compose/gateway/tyk.yml
- compose/mail/mailhog.yml
- compose/mail/mailpit.yml
- compose/mail/postal-mysql.yml
- compose/mail/postal.yml
- compose/mail/roundcube.yml
- compose/management/devarch.yml
- compose/management/dockge.yml
- compose/management/portainer.yml
- compose/management/rancher.yml
- compose/management/yacht.yml
- compose/messaging/activemq.yml
- compose/messaging/celery.yml
- compose/messaging/kafka-ui.yml
- compose/messaging/kafka.yml
- compose/messaging/nats.yml
- compose/messaging/pulsar.yml
- compose/messaging/rabbitmq.yml
- compose/messaging/redpanda.yml
- compose/messaging/zookeeper.yml
- compose/project/forgejo.yml
- compose/project/gitea.yml
- compose/project/gitlab.yml
- compose/project/openproject-cron.yml
- compose/project/openproject-seeder.yml
- compose/project/openproject-web.yml
- compose/project/openproject-worker.yml
- compose/project/taiga-back.yml
- compose/project/taiga-db.yml
- compose/project/taiga-front.yml
- compose/proxy/caddy.yml
- compose/proxy/haproxy.yml
- compose/proxy/nginx-proxy-manager.yml
- compose/proxy/varnish.yml
- compose/registry/docker-registry.yml
- compose/registry/harbor-core.yml
- compose/registry/harbor-jobservice.yml
- compose/registry/harbor-registry.yml
- compose/registry/nexus.yml
- compose/registry/verdaccio.yml
- compose/search/manticore.yml
- compose/search/meilisearch.yml
- compose/search/solr.yml
- compose/search/sonic.yml
- compose/search/typesense.yml
- compose/security/authelia.yml
- compose/security/authentik-server.yml
- compose/security/authentik-worker.yml
- compose/security/keycloak.yml
- compose/security/trivy.yml
- compose/security/vault.yml
- compose/storage/azurite.yml
- compose/storage/localstack.yml
- compose/storage/minio.yml
- compose/storage/seaweedfs-filer.yml
- compose/storage/seaweedfs-master.yml
- compose/storage/seaweedfs-s3.yml
- compose/storage/seaweedfs-volume.yml
- compose/testing/README.md
- compose/testing/gatling.yml
- compose/testing/k6.yml
- compose/testing/playwright.yml
- compose/testing/selenium-chrome.yml
- compose/testing/selenium-firefox.yml
- compose/testing/selenium-hub.yml
- compose/workflow/README.md
- compose/workflow/airflow-init.yml
- compose/workflow/airflow-scheduler.yml
- compose/workflow/airflow-webserver.yml
- compose/workflow/n8n.yml
- compose/workflow/prefect-agent.yml
- compose/workflow/prefect.yml
- compose/workflow/temporal-server.yml
- compose/workflow/temporal-ui.yml

# devarch - config
Generated: Fri Dec 12 15:09:04 EST 2025
Folder: config

## Folder Structure
- config/airflow/README.md
- config/airflow/dags/example_dag.py
- config/apisix/config.yaml
- config/caddy/Caddyfile
- config/devarch/Dockerfile
- config/devarch/api/endpoints/apps.php
- config/devarch/api/endpoints/bulk.php
- config/devarch/api/endpoints/categories.php
- config/devarch/api/endpoints/category-containers.php
- config/devarch/api/endpoints/category-refresh.php
- config/devarch/api/endpoints/containers.php
- config/devarch/api/endpoints/control.php
- config/devarch/api/endpoints/domains.php
- config/devarch/api/endpoints/logs.php
- config/devarch/api/lib/apps.php
- config/devarch/api/lib/categories.php
- config/devarch/api/lib/common.php
- config/devarch/api/lib/containers.php
- config/devarch/api/public/index.php
- config/dotnet/Dockerfile
- config/envoy/envoy.yaml
- config/go/Dockerfile
- config/haproxy/haproxy.cfg
- config/kafka/server.properties
- config/kong/kong.yml
- config/krakend/krakend.json
- config/logstash/logstash.yml
- config/logstash/pipeline.conf
- config/loki/loki-config.yml
- config/nginx/Dockerfile
- config/nginx/certs/local.crt
- config/nginx/certs/local.key
- config/nginx/custom/events.conf
- config/nginx/custom/http.conf
- config/nginx/custom/http_top.conf
- config/nginx/custom/root_top.conf
- config/nginx/custom/server_proxy.conf
- config/nginx/npm-import.sql
- config/node/Dockerfile
- config/node/ecosystem.config.js
- config/otel-collector/otel-collector.yml
- config/php/Dockerfile
- config/php/php.ini
- config/php/start.sh
- config/phpmyadmin/config.inc.php
- config/prometheus/.my.cnf
- config/prometheus/blackbox.yml
- config/prometheus/prometheus.yml
- config/prometheus/rules/alerts.yml
- config/python/Dockerfile
- config/python/requirements.txt
- config/rabbitmq/enabled_plugins
- config/rabbitmq/rabbitmq.conf
- config/rust/Dockerfile
- config/sonic/config.cfg
- config/supervisord/supervisord.conf
- config/synapse/homeserver.yaml
- config/tempo/tempo-config.yml
- config/traefik/dynamic/routes.yml
- config/traefik/traefik.yml
- config/tyk/tyk.conf
- config/typesense/typesense-server.ini
- config/varnish/default.vcl
- config/vault/config.hcl
- config/vite/Dockerfile
