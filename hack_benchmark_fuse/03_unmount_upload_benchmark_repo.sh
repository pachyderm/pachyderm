#!/bin/bash
set -xeuo pipefail
echo "Starting 03_unmount_upload_benchmark_repo.sh with BENCH_NUMJOBS=$BENCH_NUMJOBS, BENCH_FILESIZE=$BENCH_FILESIZE BENCH_NRFILES=$BENCH_NRFILES, RUN_ID=$RUN_ID"

echo "TIME:UNMOUNT"
time curl -XPUT 'localhost:9002/repos/benchmark/master/_unmount?name=benchmark&mode=rw'