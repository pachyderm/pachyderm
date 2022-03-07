#!/bin/bash
set -xeuo pipefail
echo "Starting 05_run_read_benchmark.sh with BENCH_NUMJOBS=$BENCH_NUMJOBS, BENCH_FILESIZE=$BENCH_FILESIZE BENCH_NRFILES=$BENCH_NRFILES, RUN_ID=$RUN_ID"

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

(cd /pfs/benchmark
 echo "TIME:READ_BENCHMARK"
 time fio --rw=read --numjobs=$BENCH_NUMJOBS --filesize=$BENCH_FILESIZE --nrfiles=$BENCH_NRFILES --name=mount_server
)