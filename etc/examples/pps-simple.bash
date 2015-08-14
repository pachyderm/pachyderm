#!/bin/bash

set -Ee

DIR="$(cd "$(dirname "${0}")" && pwd)"
source "${DIR}/source.bash"

simple() {
  run pps version
  id="$(run pps start github.com/pachyderm/pachyderm src/pps/server/testdata/basic)"
  echo "${id}"
  status="$(run pps status "${id}")"
  echo "${status}"
  while ! echo "${status}" | grep -E 'SUCCESS|ERROR' > /dev/null; do
    sleep 1
    status="$(run pps status "${id}")"
    echo "${status}"
  done
  run pps logs "${id}"
}

do_all "${0}" simple
