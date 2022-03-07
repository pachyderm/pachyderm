#!/bin/bash
set -xeuo pipefail

pachctl create repo benchmark
curl -XPUT 'localhost:9002/repos/benchmark/master/_mount?name=benchmark&mode=rw'
echo "TIME:WRITE_BENCHMARK"
time fio $(cat current_benchmark) --rw=write --directory=/pfs/benchmark