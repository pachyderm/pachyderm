#!/bin/bash
set -xeuo pipefail
echo "Starting 04_mount_benchmark_repo.sh with BENCH_NUMJOBS=$BENCH_NUMJOBS, BENCH_FILESIZE=$BENCH_FILESIZE BENCH_NRFILES=$BENCH_NRFILES, RUN_ID=$RUN_ID"

echo "TIME:MOUNT"
time curl -XPUT 'localhost:9002/repos/benchmark/master/_mount?name=benchmark&mode=rw'
echo "TIME:LS"
time ls /pfs/benchmark
echo "TIME:FIND"
time find /pfs/benchmark
echo "TIME:FIND"