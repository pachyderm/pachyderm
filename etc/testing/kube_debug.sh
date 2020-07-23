#!/bin/bash

set -ex

echo "=== TEST FAILED OR TIMED OUT, DUMPING DEBUG INFO ==="
kubectl get all --all-namespaces

# TODO: Extend this to show kubectl describe output for failed pods, this will
# probably show why things are hanging.
