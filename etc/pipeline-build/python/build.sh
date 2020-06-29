#!/bin/sh
python --version
pip --version
cd /pfs/__source__
test -f requirements.txt && pip wheel -r requirements.txt -w /pfs/out
