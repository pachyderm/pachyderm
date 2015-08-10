#!/bin/bash

set -Ee

DIR="$(cd "$(dirname "${0}")" && pwd)"
source "${DIR}/source.bash"

simple() {
  run pps version
  id="$(run pps start github.com/pachyderm/pachyderm src/pps/server/testdata/basic)"
  echo "${id}"
  i=0
  while [ $i -lt 10 ]; do
    run pps status "${id}"
    sleep 1
    i=$(expr ${i} + 1)
  done
}

do_pps "${0}" simple
