#!/bin/bash

set -euxo pipefail

case "${BUCKET}" in
 LOAD1)
    etc/testing/circle/run_hub_tests.sh etc/testing/loads/load-1.yaml 
    ;;
 LOAD2)
    etc/testing/circle/run_hub_tests.sh etc/testing/loads/load-2.yaml 
    ;;
 LOAD3)
    etc/testing/circle/run_hub_tests.sh etc/testing/loads/load-3.yaml 
    ;;
esac
