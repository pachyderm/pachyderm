import csv
import json
import re
import time
from typing import Sequence
from google.cloud import bigquery
import os
from google.oauth2 import service_account


def main():
    results_folder = os.getenv('API_PERF_RESULTS_FOLDER')
    common_columns = {
        'Workflow_Id': os.getenv('CIRCLE_WORKFLOW_ID'),
        'Job_Id': os.getenv('CIRCLE_WORKFLOW_JOB_ID'),
        'Version': normalize_version(os.getenv('PACHD_PERF_VERSION')),
        'Upload_Timestamp': time.time()
    }
    for key, value in common_columns.items():
        if not value:
            raise Exception(
                f'Missing a required value to export api performance data. Need {key}.')

    client = get_bigquery_client()

    rows_to_insert = get_file_rows(
        "api-perf_stats.csv", results_folder, common_columns)
    insert_to_bigquery(client, rows_to_insert, 'api-perf-stats')

    rows_to_insert = get_file_rows(
        "api-perf_stats_history.csv", results_folder, common_columns)
    insert_to_bigquery(client, rows_to_insert, 'api-perf-stats-history')

    rows_to_insert = get_file_rows(
        "api-perf_exceptions.csv", results_folder, common_columns)
    insert_to_bigquery(client, rows_to_insert, 'api-perf-exceptions')

    rows_to_insert = get_file_rows(
        "api-perf_failures.csv", results_folder, common_columns)
    insert_to_bigquery(client, rows_to_insert, 'api-perf-failures')


# Log in to big query and return the authenticated client
def get_bigquery_client() -> bigquery.Client:
    auth_string = os.getenv('BIGQUERY_AUTH_JSON')
    if not auth_string:
        raise Exception(
            f'Missing bigquery authorization to export api performance data. Need BIGQUERY_AUTH_JSON.')
    bigquery_auth_json = json.loads(auth_string)
    bigquery_credentials = \
        service_account.Credentials \
        .from_service_account_info(bigquery_auth_json)
    return bigquery.Client(credentials=bigquery_credentials)


# Collects the rows to insert from a csv and prepares them for insertion into big query. returns the rows to insert
def get_file_rows(file_name: str, results_folder: str, common_columns: dict[str, any]) -> list[dict[str, any]]:
    if results_folder:
        file_path = os.path.join(results_folder, file_name)
    rows_to_insert = []
    with open(file_path, mode='r') as c:
        csv_reader = csv.reader(c, delimiter=',')
        columns = format_column_names(next(csv_reader))
        for row in csv_reader:
            row_dict = {}
            for i, value in enumerate(row):
                row_dict[columns[i]] = value
            row_dict.update(common_columns)
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


# We don't want some versions with v in front and some without in the DB, so just remove v if it's there
def normalize_version(version: str):
    if version[0] == 'v':
        return version[1:]
    return version


# do the actual call to bigquery to insert rows. returns any errors
def insert_to_bigquery(client: bigquery.Client, rows_to_insert: list[dict[str, any]], table_name: str) -> Sequence[dict]:
    errors = []
    if len(rows_to_insert)>0:
        table = client.get_table("{}.{}.{}".format(
            "build-release-001", "insights", table_name))
        errors = client.insert_rows_json(table, rows_to_insert)
        if errors == []:
            print(f'Successful insert to {table_name}')
        else:
            print(f'Errors inserting into {table_name}: {errors}')
    else: 
        print(f'No rows to insert into {table_name}, skipping.')
    return errors


if __name__ == "__main__":
    main()
