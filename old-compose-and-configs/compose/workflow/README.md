# Workflow Automation Services

Orchestration and automation engines for DevArch.

## Services

### Apache Airflow (DAG-based workflows)
- **Files**: `airflow-init.yml`, `airflow-webserver.yml`, `airflow-scheduler.yml`
- **Ports**: 9900 (webserver)
- **Dependencies**: PostgreSQL (database: airflow)
- **Usage**:
  ```bash
  # Initialize (run once)
  docker compose -f compose/workflow/airflow-init.yml up

  # Start services
  docker compose \
    -f compose/workflow/airflow-webserver.yml \
    -f compose/workflow/airflow-scheduler.yml \
    up -d
  ```
- **Access**: http://localhost:9900 (admin/admin)

### n8n (Visual workflow builder)
- **File**: `n8n.yml`
- **Port**: 9910
- **Dependencies**: None (standalone)
- **Usage**:
  ```bash
  docker compose -f compose/workflow/n8n.yml up -d
  ```
- **Access**: http://localhost:9910

### Prefect (Python-native workflows)
- **Files**: `prefect.yml`, `prefect-agent.yml`
- **Port**: 9920 (server)
- **Dependencies**: None (standalone)
- **Usage**:
  ```bash
  # Start server and agent
  docker compose \
    -f compose/workflow/prefect.yml \
    -f compose/workflow/prefect-agent.yml \
    up -d
  ```
- **Access**: http://localhost:9920

### Temporal (Durable execution)
- **Files**: `temporal-server.yml`, `temporal-ui.yml`
- **Ports**: 9930 (web), 9931 (server)
- **Dependencies**: PostgreSQL
- **Usage**:
  ```bash
  docker compose \
    -f compose/workflow/temporal-server.yml \
    -f compose/workflow/temporal-ui.yml \
    up -d
  ```
- **Access**: http://localhost:9930

## Port Range
9900-9999
