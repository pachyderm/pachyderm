#!/bin/bash
set -xeuo pipefail
echo "Starting 10_teardown.sh with BENCH_NUMJOBS=$BENCH_NUMJOBS, BENCH_FILESIZE=$BENCH_FILESIZE BENCH_NRFILES=$BENCH_NRFILES, RUN_ID=$RUN_ID"

minikube delete