#!/bin/sh

set -Ee

DIR="$(cd "$(dirname "${0}")/../.." && pwd)"
cd "${DIR}"

func run() {
  echo $@
  $@
}

run make launch-ppsd install
run pps version
run pps start github.com/pachyderm/pachyderm src/pps/server/testdata/basic
