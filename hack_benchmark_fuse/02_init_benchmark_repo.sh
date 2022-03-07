#!/bin/bash
set -xeuo pipefail
echo "Starting 02_init_benchmark_repo.sh with BENCH_NUMJOBS=$BENCH_NUMJOBS, BENCH_FILESIZE=$BENCH_FILESIZE BENCH_NRFILES=$BENCH_NRFILES, RUN_ID=$RUN_ID"

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

# TODO: remove || true here probably, just for faster iteration...
pachctl create repo benchmark || true
curl -XPUT 'localhost:9002/repos/benchmark/master/_mount?name=benchmark&mode=rw'
(cd /pfs/benchmark
 echo "TIME:WRITE_BENCHMARK"
 time fio --rw=write --numjobs=$BENCH_NUMJOBS --filesize=$BENCH_FILESIZE --nrfiles=$BENCH_NRFILES --name=mount_server
)