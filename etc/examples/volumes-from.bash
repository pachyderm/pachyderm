#!/bin/bash

set -Ee

DIR="$(cd "$(dirname "${0}")" && pwd)"
source "${DIR}/source.bash"

kill_pfs_example_mount() {
  run docker kill pfs_example_mount || true
  run docker rm pfs_example_mount || true
}

if [ -n "${PFS_MOUNT_EXAMPLE}" ]; then
  export PFS_ADDRESS="$(echo "${PACHYDERM_PFSD_1_PORT}" | sed "s/tcp:\/\///")"
  run_make install
  run pfs init test-repository
  commit_id="$(run pfs branch test-repository scratch)"
  echo "${commit_id}"
  echo hello | pfs put test-repository "${commit_id}" foo.txt
  run pfs commit "${commit_id}"
  cd /in
  run pfs mount test-repository &
  run sleep 1
  run touch /watch/touch
elif [ -n "${PFS_CLIENT_EXAMPLE}" ]; then
  while [ ! -f '/watch/touch' ]; do
    echo "no /watch/touch yet, sleeping..."
    sleep 1
  done
  run ls /in
else
  kill_pfs_example_mount
  run_make launch-pfsd
  run docker run \
    -e PFS_MOUNT_EXAMPLE=1 \
    --name pfs_example_mount \
    --link pachyderm_pfsd_1 \
    --volume '/in' \
    --volume '/watch' \
    -d \
    pachyderm_compile \
      bash "etc/examples/$(basename "${0}")"
  run docker run \
    -e PFS_CLIENT_EXAMPLE=1 \
    --volumes-from pfs_example_mount \
    pachyderm_compile \
    bash "etc/examples/$(basename "${0}")"
  kill_pfs_example_mount
fi
