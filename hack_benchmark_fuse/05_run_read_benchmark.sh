#!/bin/bash
set -xeuo pipefail
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

(cd /pfs/benchmark
 echo "TIME:READ_BENCHMARK"
 time fio $SCRIPT_DIR/$(cat current_benchmark) --rw=read
)