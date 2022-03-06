#!/bin/bash
set -xeuo pipefail
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

pachctl create repo benchmark
curl -XPUT 'localhost:9002/repos/benchmark/master/_mount?name=benchmark&mode=rw'
(cd /pfs/benchmark
 fio $SCRIPT_DIR/small.ini
)