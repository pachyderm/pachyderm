#!/bin/sh
set -e
cd /pfs/build
chmod +x main
./main $@
