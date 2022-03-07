#!/bin/bash
set -xeuo pipefail

cd /pfs/benchmark
time fio $(cat current_benchmark) --rw=read --directory=/pfs/benchmark