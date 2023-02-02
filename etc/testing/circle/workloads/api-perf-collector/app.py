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

    rows_to_insert = get_csv_file_rows(
        'api-perf_stats.csv', results_folder, common_columns)
    insert_to_bigquery(client, rows_to_insert, 'api-perf-stats')

    rows_to_insert = get_csv_file_rows(
        'api-perf_stats_history.csv', results_folder, common_columns)
    insert_to_bigquery(client, rows_to_insert, 'api-perf-stats-history')

    rows_to_insert = get_csv_file_rows(
        'api-perf_exceptions.csv', results_folder, common_columns)
    insert_to_bigquery(client, rows_to_insert, 'api-perf-exceptions')

    rows_to_insert = get_csv_file_rows(
        'api-perf_failures.csv', results_folder, common_columns)
    insert_to_bigquery(client, rows_to_insert, 'api-perf-failures')

    rows_to_insert = get_log_file_rows(
        'pachctl_logs.jsonl', results_folder, common_columns)
    insert_to_bigquery(client, rows_to_insert, 'api-perf-pachd-logs')

    rows_to_insert = get_kubeconfig_rows(
        'pachd-k8s-config.json', results_folder, 'pachd', common_columns)
    insert_to_bigquery(client, rows_to_insert, 'api-perf-kube-config')

    rows_to_insert = get_kubeconfig_rows(
        'pg-bouncer-k8s-config.json', results_folder, 'pg-bouncer', common_columns)
    insert_to_bigquery(client, rows_to_insert, 'api-perf-kube-config')

    rows_to_insert = get_kubeconfig_rows(
        'postgres-k8s-config.json', results_folder, 'postgres', common_columns)
    insert_to_bigquery(client, rows_to_insert, 'api-perf-kube-config')

    rows_to_insert = get_sadf_rows(
        'sadf_stats', results_folder, common_columns)
    insert_to_bigquery(client, rows_to_insert, 'api-perf-sar-stats')


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
# returns a list of rows to insert. Each row is represented by a dictionary with `column_name: column_value` format.
def get_csv_file_rows(file_name: str, results_folder: str, common_columns: dict[str, any]) -> list[dict[str, any]]:
    if results_folder:
        file_path = os.path.join(results_folder, file_name)
    rows = []
    with open(file_path, mode='r') as c:
        csv_reader = csv.reader(c, delimiter=',')
        columns = format_column_names(next(csv_reader))
        for row in csv_reader:
            row_dict = {}
            for i, value in enumerate(row):
                if not value == 'N/A' : # Just don't insert NA for data typing
                    row_dict[columns[i]] = value
            row_dict.update(common_columns)
            rows.append(row_dict)
    return rows


# Each log is json within json, so load that into a list of dictionaries for storage.
def get_log_file_rows(file_name: str, results_folder: str, common_columns: dict[str, any]) -> list[dict[str, any]]:
    if results_folder:
        file_path = os.path.join(results_folder, file_name)
    rows = []
    with open(file_path, 'r') as f:
        for line in f:
            try:
                json_log = json.loads(line)['log']
                log_json = json.loads(json_log)
                if log_json['severity'] != 'info' and log_json['severity'] != 'debug':
                    for field in log_json.keys():
                        # flatten json for insertion to BQ
                        if type(log_json[field]) is dict:
                            log_json[field] = json.dumps(log_json[field])
                    log_json.update(common_columns)
                    rows.append(log_json)
            except ValueError:  # in the case the log is not parsable json, we have to toss it
                pass
    return rows


def get_kubeconfig_rows(file_name: str, results_folder: str, app_name: str, common_columns: dict[str, any]) -> list[dict[str, any]]:
    if results_folder:
        file_path = os.path.join(results_folder, file_name)
    rows = []
    with open(file_path, 'r') as f:
        # f should already be json so no need to json.loads/dumps
        kubeconfig = {'pod': app_name, 'kubeconfig': f.read()}
        kubeconfig.update(common_columns)
        rows.append(kubeconfig)
    return rows


def get_sadf_rows(file_name: str, results_folder: str, common_columns: dict[str, any]) -> list[dict[str, any]]:
    if results_folder:
        file_path = os.path.join(results_folder, file_name)
    rows = []
    # we want one row per timestamp so we flatten the report by timestamps before finalizing
    time_sorted_rows = {}
    with open(file_path, 'r') as f:
        for line in f:
            if line.startswith('#'):  # then they are columns
                column_names = format_column_names(line[1:].strip().split(';'))
            else:
                values = line.split(';')
                row = {}
                timestamp = ""
                for i, name in enumerate(column_names):
                    row[name] = values[i]
                    if name == 'timestamp':
                        timestamp = values[i].strip()
                row.update(common_columns)
                if not timestamp in time_sorted_rows.keys():
                    time_sorted_rows[timestamp] = {}
                time_sorted_rows[timestamp].update(row)
    for row in time_sorted_rows.values():
        rows.append(row)
    return rows


# We don't want some versions with v in front and some without in the DB, so just remove v if it's there
def normalize_version(version: str):
    # handle master sha
    if not re.search('^v?(\d.\d.\d).*', version):
        return 'master'
    if version[0] == 'v':
        return version[1:]
    return version


# do the actual call to bigquery to insert rows. returns any errors
def insert_to_bigquery(client: bigquery.Client, rows_to_insert: list[dict[str, any]], table_name: str) -> Sequence[dict]:
    errors = []
    if len(rows_to_insert) > 0:
        table = client.get_table('{}.{}.{}'.format(
            'build-release-001', 'insights', table_name))
        errors = client.insert_rows_json(table, rows_to_insert)
        if errors == []:
            print(f'Successful insert to {table_name}')
        else:
            print(f'Errors inserting into {table_name}: {errors}')
    else:
        print(f'No rows to insert into {table_name}, skipping.')
    return errors


if __name__ == '__main__':
    main()
