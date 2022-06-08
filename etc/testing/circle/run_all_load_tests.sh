#!/bin/bash

set -euxo pipefail

case "${BUCKET}" in
 LOAD1)
    etc/testing/circle/run_load_tests.sh etc/testing/loads/few-commits/few-modifications/few-directories/small-files/load.yaml -s 1630089896104084026 -p etc/testing/loads/pod-patch.json
    ;;
 LOAD2)
    etc/testing/circle/run_load_tests.sh etc/testing/loads/few-commits/few-modifications/many-directories/small-files/load.yaml -s 1630089896104084026 -p etc/testing/loads/pod-patch.json
    ;;
 LOAD3)
    etc/testing/circle/run_load_tests.sh etc/testing/loads/few-commits/few-modifications/one-directory/large-files/load.yaml -s 1630089896104084026 -p etc/testing/loads/pod-patch.json
    ;;
 LOAD4)
    etc/testing/circle/run_load_tests.sh etc/testing/loads/few-commits/few-modifications/one-directory/medium-files/load.yaml -s 1630089896104084026 -p etc/testing/loads/pod-patch.json
    ;;
 LOAD5)
    etc/testing/circle/run_load_tests.sh etc/testing/loads/few-commits/few-modifications/one-directory/mix-files/load.yaml -s 1630089896104084026 -p etc/testing/loads/pod-patch.json
    ;;
 LOAD6)
    etc/testing/circle/run_load_tests.sh etc/testing/loads/few-commits/few-modifications/one-directory/small-files/load.yaml -s 1630089896104084026 -p etc/testing/loads/pod-patch.json
    ;;
 LOAD7)
    etc/testing/circle/run_load_tests.sh etc/testing/loads/few-commits/many-modifications/one-directory/medium-files/load.yaml -s 1630089896104084026 -p etc/testing/loads/pod-patch.json
    ;;
 LOAD8)
    etc/testing/circle/run_load_tests.sh etc/testing/loads/few-commits/many-modifications/one-directory/mix-files/load.yaml -s 1630089896104084026 -p etc/testing/loads/pod-patch.json
    ;;
 LOAD9)
    etc/testing/circle/run_load_tests.sh etc/testing/loads/few-commits/many-modifications/one-directory/small-files/load.yaml -s 1630089896104084026 -p etc/testing/loads/pod-patch.json
    ;;
 LOAD10)
    etc/testing/circle/run_load_tests.sh etc/testing/loads/many-commits/few-modifications/one-directory/medium-files/load.yaml -s 1630089896104084026 -p etc/testing/loads/pod-patch.json
    ;;
 LOAD11)
    etc/testing/circle/run_load_tests.sh etc/testing/loads/many-commits/few-modifications/one-directory/mix-files/load.yaml -s 1630089896104084026 -p etc/testing/loads/pod-patch.json
    ;;
 LOAD12)
    etc/testing/circle/run_load_tests.sh etc/testing/loads/many-commits/few-modifications/one-directory/small-files/load.yaml -s 1630089896104084026 -p etc/testing/loads/pod-patch.json
    ;;
 LOAD13)
    etc/testing/circle/run_load_tests.sh etc/testing/loads/many-commits/many-modifications/one-directory/medium-files/load.yaml -s 1630089896104084026 -p etc/testing/loads/pod-patch.json
    ;;
 LOAD14)
    etc/testing/circle/run_load_tests.sh etc/testing/loads/many-commits/many-modifications/one-directory/mix-files/load.yaml -s 1630089896104084026 -p etc/testing/loads/pod-patch.json
    ;;
 LOAD15)
    etc/testing/circle/run_load_tests.sh etc/testing/loads/many-commits/many-modifications/one-directory/small-files/load.yaml -s 1630089896104084026 -p etc/testing/loads/pod-patch.json
    ;;
esac
