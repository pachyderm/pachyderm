#!/bin/sh
set -e
python --version
pip --version
cd /pfs/source
test -f requirements.txt && pip wheel -r requirements.txt -w /pfs/out
cp /app/run /pfs/out/run
