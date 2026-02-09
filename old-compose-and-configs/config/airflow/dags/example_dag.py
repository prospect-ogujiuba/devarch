"""
Example Airflow DAG for DevArch
Place your workflow DAG files in this directory
"""
from datetime import datetime, timedelta
from airflow import DAG
from airflow.operators.bash import BashOperator

default_args = {
    'owner': 'devarch',
    'depends_on_past': False,
    'start_date': datetime(2024, 1, 1),
    'email_on_failure': False,
    'email_on_retry': False,
    'retries': 1,
    'retry_delay': timedelta(minutes=5),
}

dag = DAG(
    'example_devarch',
    default_args=default_args,
    description='Example DAG for DevArch',
    schedule_interval=timedelta(days=1),
    catchup=False,
)

t1 = BashOperator(
    task_id='print_date',
    bash_command='date',
    dag=dag,
)

t2 = BashOperator(
    task_id='hello_world',
    bash_command='echo "Hello from DevArch workflow!"',
    dag=dag,
)

t1 >> t2
