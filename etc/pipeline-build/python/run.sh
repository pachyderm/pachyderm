#!/bin/sh
set -e
cd /pfs/build
pip install ./*.whl
cd /pfs/source
python main.py $@
