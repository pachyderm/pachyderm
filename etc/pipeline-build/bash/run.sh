#!/bin/sh
set -e
cd /pfs/source
test -f requirements.sh && bash requirements.sh
bash main.bash "$@"
