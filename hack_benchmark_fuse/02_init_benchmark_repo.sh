#!/bin/bash
set -xeuo pipefail
echo "Starting 02_init_benchmark_repo.sh with BENCH_NUMJOBS=$BENCH_NUMJOBS, BENCH_FILESIZE=$BENCH_FILESIZE BENCH_NRFILES=$BENCH_NRFILES, RUN_ID=$RUN_ID"

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

# TODO: remove || true here probably, just for faster iteration...
pachctl create repo benchmark || true
curl -XPUT 'localhost:9002/repos/benchmark/master/_mount?name=benchmark&mode=rw'
(cd /pfs/benchmark
 echo "TIME:WRITE_BENCHMARK"
 time strace fio --rw=write --numjobs=1 --filesize=10M --nrfiles=10 --name=mount_server
 #time strace fio $SCRIPT_DIR/$(cat $SCRIPT_DIR/current_benchmark)
)