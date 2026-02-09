# Airflow Configuration

## DAGs Directory
Place your Airflow DAG Python files in the `dags/` subdirectory.

## Usage
Start Airflow with:
```bash
# Initialize database and create admin user
docker compose -f compose/workflow/airflow-init.yml up

# Start webserver and scheduler
docker compose \
  -f compose/workflow/airflow-webserver.yml \
  -f compose/workflow/airflow-scheduler.yml \
  up -d
```

Access UI: http://localhost:9900
- Username: admin
- Password: admin
