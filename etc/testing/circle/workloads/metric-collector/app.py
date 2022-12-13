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

def main():
    
    pachClient = python_pachyderm.Client.new_in_cluster()

    # Construct a BigQuery client object.
    client = bigquery.Client()

    table = client.get_table("{}.{}.{}".format("build-release-001", "insights", "perf-tests"))

    version = pachClient.get_remote_version()

    version = "".join(f'{version.major}.{version.minor}.{version.micro}{version.additional}'.split())

    pipelines = list(pachClient.list_pipeline())

    for p in pipelines:

        jobs = list(pachClient.list_job(p.pipeline.name))

        for j in jobs:

            rows_to_insert = [{u"pachydermVersion": version, u"workloadName": "test", u"jobId": str(j.job.id),
              u"pipeline": j.job.pipeline.name, u"totalDatums": str(j.data_total), u"datumsProcessed": str(j.data_processed),
              u"startTime": j.started, u"endTime": j.finished}]

            errors = client.insert_rows_json(table, rows_to_insert)
            if errors == []:
                print("successful insert")

if __name__ == "__main__":
    main()
    print("success")