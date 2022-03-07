#!/bin/bash
set -xeuo pipefail
echo "Starting 01_destroy_and_setup.sh with BENCH_NUMJOBS=$BENCH_NUMJOBS, BENCH_FILESIZE=$BENCH_FILESIZE BENCH_NRFILES=$BENCH_NRFILES, RUN_ID=$RUN_ID"

sudo apt update && sudo apt install -y fio

# cleanup
fusermount -u /pfs || true
sudo pkill -f "pachctl mount-server" || true
minikube delete || true
rm -rf /tmp/pfs* || true

(cd ..
 minikube start
 eval $(minikube docker-env)
 make docker-build
 make install
 make launch-dev
)

sudo mkdir -p /pfs
sudo chown $USER /pfs