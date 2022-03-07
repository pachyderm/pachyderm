#!/bin/bash
set -xeuo pipefail
echo "TIME:UNMOUNT"
time curl -XPUT 'localhost:9002/repos/benchmark/master/_unmount?name=benchmark&mode=rw'