#!/bin/bash

set -euxo pipefail

case "${BUCKET}" in
 LOAD1)
    etc/testing/circle/run_load_tests.sh etc/testing/loads/load-1.yaml -s 1630089896104084026
    ;;
 LOAD2)
    etc/testing/circle/run_load_tests.sh etc/testing/loads/load-2.yaml -s 1630089896104084026
    ;;
 LOAD3)
    etc/testing/circle/run_load_tests.sh etc/testing/loads/load-3.yaml -s 1630089896104084026
    ;;
 LOAD4)
    etc/testing/circle/run_load_tests.sh etc/testing/loads/load-4.yaml -s 1630089896104084026
    ;;
 LOAD5)
    etc/testing/circle/run_load_tests.sh etc/testing/loads/load-5.yaml -s 1630089896104084026
    ;;
 LOAD6)
    etc/testing/circle/run_load_tests.sh etc/testing/loads/load-6.yaml -s 1630089896104084026
    ;;
 LOAD7)
    etc/testing/circle/run_load_tests.sh etc/testing/loads/load-7.yaml -s 1630089896104084026
    ;;
esac
