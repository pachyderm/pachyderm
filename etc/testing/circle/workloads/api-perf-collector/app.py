import json
import time
from google.cloud import bigquery
import os
from google.oauth2 import service_account


def main():
    bigquery_auth_json=json.loads(os.getenv('BIGQUERY_AUTH_JSON'))
    bigquery_credentials = service_account.Credentials.from_service_account_info(bigquery_auth_json)
    client = bigquery.Client(credentials=bigquery_credentials)

    common_columns = {
        "Workload_Id": os.getenv('CIRCLE_WORKFLOW_ID'),
        "Job_Id": os.getenv('CIRCLE_WORKFLOW_JOB_ID'),
        "Version": os.getenv('PACHD_VERSION'),
        "Upload_Timestamp": time.time()
    }
    for key, value in common_columns.items():
        if not value:
            raise Exception(f'Missing a required value to export api performance data. Need {key}.')
    
    # table = client.get_table("{}.{}.{}".format("build-release-001", "insights", "perf-tests"))


if __name__ == "__main__":
    main()
