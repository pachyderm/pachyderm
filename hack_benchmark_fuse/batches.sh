#!/bin/bash
set -xeuo pipefail

# warmup...
BENCH_NUMJOBS=1 BENCH_FILESIZE=10M BENCH_NRFILES=10 ./run.sh
BENCH_NUMJOBS=1 BENCH_FILESIZE=10M BENCH_NRFILES=100 ./run.sh

# large number of small files
BENCH_NUMJOBS=1 BENCH_FILESIZE=10M BENCH_NRFILES=1000 ./run.sh

# smaller number of large files
BENCH_NUMJOBS=1 BENCH_FILESIZE=100M BENCH_NRFILES=100 ./run.sh