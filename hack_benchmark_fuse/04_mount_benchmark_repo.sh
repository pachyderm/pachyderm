#!/bin/bash
set -xeuo pipefail

echo "TIME:MOUNT"
time curl -XPUT 'localhost:9002/repos/benchmark/master/_mount?name=benchmark&mode=rw'
echo "TIME:LS"
time ls /pfs/benchmark
echo "TIME:FIND"
time find /pfs/benchmark
echo "TIME:FIND"