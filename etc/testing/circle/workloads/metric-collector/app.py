import io
from google.cloud import bigquery
import python_pachyderm
import datetime
from python_pachyderm.service import pps_proto
from google.protobuf import empty_pb2
from python_pachyderm.proto.v2.version.versionpb import version_pb2, version_pb2_grpc
import os
import sys
import getopt
from datetime import datetime
import pprint
import json
import base64
from google.oauth2 import service_account

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

def main():

    pachClient = python_pachyderm.Client.new_in_cluster()

    # Construct a BigQuery client object.
    client = get_bigquery_client()
    print("BigQuery Client created using default project: {}".format(client.project))

    table = client.get_table("{}.{}.{}".format("build-release-001", "insights", "perf-tests"))

    version = pachClient.get_remote_version()

    version = "".join(f'{version.major}.{version.minor}.{version.micro}{version.additional}'.split())

    pipelines = list(pachClient.list_pipeline())

    workload_id = os.getenv('WORKLOAD_ID')

    for p in pipelines:

        jobs = list(pachClient.list_job(p.pipeline.name))

        for j in jobs:

            rows_to_insert = [{u"version": version, u"run": str(1), u"workflow": str(workload_id), u"jobId": str(j.job.id),
              u"pipeline": j.job.pipeline.name, u"totalDatums": str(j.data_total), u"datumsProcessed": str(j.data_processed),
              u"start": str(j.started.seconds), u"finished": str(j.finished.seconds),
              u"startNano": str(j.started.nanos), u"finishedNano": str(j.finished.nanos)}]

            errors = client.insert_rows_json(table, rows_to_insert)
            if errors == []:
                print("successful insert")

if __name__ == "__main__":
    main()
    print("success")