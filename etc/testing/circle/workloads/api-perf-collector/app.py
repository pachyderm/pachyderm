import csv
import json
import re
import time
from google.cloud import bigquery
import os
from google.oauth2 import service_account


def main():

    client = get_bigquery_client()
    common_columns = {
        "Workload_Id": os.getenv('CIRCLE_WORKFLOW_ID'),
        "Job_Id": os.getenv('CIRCLE_WORKFLOW_JOB_ID'),
        "Version": os.getenv('PACHD_VERSION'),
        "Upload_Timestamp": time.time()
    }

    for key, value in common_columns.items():
        if not value:
            raise Exception(
                f'Missing a required value to export api performance data. Need {key}.')

    rows_to_insert = get_file_rows("api-perf_stats.csv", common_columns)
    insert_to_bigquery(client, rows_to_insert, 'api-perf-stats')


# Log in to big wuery and return the authenticated client
def get_bigquery_client() -> bigquery.Client:
    bigquery_auth_json = json.loads(os.getenv('BIGQUERY_AUTH_JSON'))
    bigquery_credentials = \
        service_account.Credentials \
        .from_service_account_info(bigquery_auth_json)
    return bigquery.Client(credentials=bigquery_credentials)


# Collects the rows to insert from a csv and prepares them for insertion into big query. returns the rows to insert
def get_file_rows(file_name: str, common_columns: dict[str, any]) -> list[dict[str, any]]:
    rows_to_insert = []
    with open(file_name, mode='r') as c:
        csv_reader = csv.reader(c, delimiter=',')
        columns = format_column_names(next(csv_reader))
        for row in csv_reader:
            row_dict = {}
            row_dict.update(common_columns)
            for i, value in enumerate(row):
                row_dict[columns[i]] = value
            rows_to_insert.append(row_dict)
    return rows_to_insert


# format column names to be compatible with bigquery's rules
def format_column_names(raw: list[str]) -> list[str]:
    formatted = []
    for name in raw:
        # strip non-alphamnumeric column titles
        tmp = re.sub('[^a-zA-Z\d:]', '_', name)
        # append _ instead of starting with a number
        if re.match('\d', tmp[0]) is not None:
            tmp = '_' + tmp
        formatted.append(tmp)
    return formatted


def insert_to_bigquery(client: bigquery.Client, rows_to_insert: list[dict[str, any]], table_name: str):
    table = client.get_table("{}.{}.{}".format(
        "build-release-001", "insights", table_name))
    errors = client.insert_rows_json(table, rows_to_insert)
    if errors == []:
        print(f'successful insert to {table_name}')


if __name__ == "__main__":
    main()
