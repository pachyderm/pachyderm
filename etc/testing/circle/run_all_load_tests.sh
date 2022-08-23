#!/bin/bash

set -euxo pipefail

case "${BUCKET}" in
 LOAD1)
    etc/testing/circle/run_load_tests.sh etc/testing/loads/few-commits/few-modifications/few-directories/small-files/load.yaml -s 1630089896104084026
    ;;
 LOAD2)
    etc/testing/circle/run_load_tests.sh etc/testing/loads/few-commits/few-modifications/many-directories/small-files/load.yaml -s 1630089896104084026
    ;;
 LOAD3)
    etc/testing/circle/run_load_tests.sh etc/testing/loads/few-commits/few-modifications/one-directory/large-files/load.yaml -s 1630089896104084026
    ;;
 LOAD4)
    etc/testing/circle/run_load_tests.sh etc/testing/loads/few-commits/few-modifications/one-directory/medium-files/load.yaml -s 1630089896104084026
    ;;
 LOAD5)
    etc/testing/circle/run_load_tests.sh etc/testing/loads/few-commits/few-modifications/one-directory/mix-files/load.yaml -s 1630089896104084026
    ;;
 LOAD6)
    etc/testing/circle/run_load_tests.sh etc/testing/loads/few-commits/few-modifications/one-directory/small-files/load.yaml -s 1630089896104084026
    ;;
 LOAD7)
    etc/testing/circle/run_load_tests.sh etc/testing/loads/few-commits/many-modifications/one-directory/medium-files/load.yaml -s 1630089896104084026
    ;;
 LOAD8)
    etc/testing/circle/run_load_tests.sh etc/testing/loads/few-commits/many-modifications/one-directory/mix-files/load.yaml -s 1630089896104084026
    ;;
 LOAD9)
    etc/testing/circle/run_load_tests.sh etc/testing/loads/few-commits/many-modifications/one-directory/small-files/load.yaml -s 1630089896104084026
    ;;
 LOAD10)
    etc/testing/circle/run_load_tests.sh etc/testing/loads/many-commits/few-modifications/one-directory/medium-files/load.yaml -s 1630089896104084026
    ;;
 LOAD11)
    etc/testing/circle/run_load_tests.sh etc/testing/loads/many-commits/few-modifications/one-directory/mix-files/load.yaml -s 1630089896104084026
    ;;
 LOAD12)
    etc/testing/circle/run_load_tests.sh etc/testing/loads/many-commits/few-modifications/one-directory/small-files/load.yaml -s 1630089896104084026
    ;;
 LOAD13)
    etc/testing/circle/run_load_tests.sh etc/testing/loads/many-commits/many-modifications/one-directory/medium-files/load.yaml -s 1630089896104084026
    ;;
 LOAD14)
    etc/testing/circle/run_load_tests.sh etc/testing/loads/many-commits/many-modifications/one-directory/mix-files/load.yaml -s 1630089896104084026
    ;;
 LOAD15)
    etc/testing/circle/run_load_tests.sh etc/testing/loads/many-commits/many-modifications/one-directory/small-files/load.yaml -s 1630089896104084026
    ;;
esac
