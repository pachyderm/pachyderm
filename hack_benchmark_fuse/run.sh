#!/bin/bash
set -xeuo pipefail
if [ "$1" != "" ]; then
    echo "$1" > current_benchmark
fi
outputdir="output-$(date +%s)"
mkdir -p $outputdir
echo "setup"
time ./01_destroy_and_setup.sh &> $outputdir/01_destroy_and_setup.log
echo "mount-server"
pachctl mount-server &
pid=$!
sleep 1
echo "init"
time ./02_init_benchmark_repo.sh &> $outputdir/02_init_benchmark_repo.log
echo "unmount"
time ./03_unmount_upload_benchmark_repo.sh &> $outputdir/03_unmount_upload_benchmark_repo.log
echo "mount"
time ./04_mount_benchmark_repo.sh &> $outputdir/04_mount_benchmark_repo.log
echo "run"
time ./05_run_read_benchmark.sh &> $outputdir/05_run_read_benchmark.log
echo "teardown"
time ./10_teardown.sh &> $outputdir/10_teardown.log
kill $pid