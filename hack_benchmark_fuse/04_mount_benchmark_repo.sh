#!/bin/bash
set -xeuo pipefail

curl -XPUT 'localhost:9002/repos/benchmark/master/_mount?name=benchmark&mode=rw'
ls /pfs/benchmark