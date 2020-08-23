#!/bin/sh
set -e
cd /pfs/source
test -f requirements.txt && pip install /pfs/build/*.whl
python main.py "$@"
