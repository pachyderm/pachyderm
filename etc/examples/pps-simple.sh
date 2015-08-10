#!/bin/sh

set -Ee

DIR="$(cd "$(dirname "${0}")/../.." && pwd)"
cd "${DIR}"

run() {
  echo $@ >&2
  $@
}

if [ -z ${CLIENT} ]; then
  run make launch-ppsd
  run docker run -e CLIENT=1 --link pachyderm_ppsd_1 pachyderm_compile sh "etc/examples/$(basename "${0}")"
  run make docker-clean-launch
else
  export PPS_ADDRESS="$(echo "${PACHYDERM_PPSD_1_PORT}" | sed "s/tcp:\/\///")"
  run make install
  run pps version
  id="$(run pps start github.com/pachyderm/pachyderm src/pps/server/testdata/basic)"
  echo "${id}"
  i=0
  while [ $i -lt 10 ]; do
    run pps status "${id}"
    sleep 1
    i=$(expr ${i} + 1)
  done
fi
