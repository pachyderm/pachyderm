#!/bin/bash

set -euxo pipefail

case "${BUCKET}" in
 LOAD1)
    etc/testing/circle/run_hub_tests.sh etc/testing/loads/load-1.yaml 
    ;;
esac
