#!/bin/bash
#
# This is the first step in building Pachyderm binaries ('pachd' and 'worker')
# and is called by the 'make docker-build-pachd' and 'make docker-build-worker'
# make targets. It doesn't actually run the go compiler (that's in
# run_go_cmds.sh, called at the bottom) but it sets up an unprivileged user
# with the host caller's user ID, under which the go compiler will run

set -Eex

# Validate env vars
if [[ -z "${CALLING_USER_ID}" ]]; then
  echo "Cannot do docker build without the caller's user ID" >/dev/stderr
  exit 1
fi
if [[ -z "${DOCKER_GROUP_ID}" ]]; then
  echo "Cannot do docker build without the 'docker' group's ID" >/dev/stderr
  exit 1
fi

useradd --uid="${CALLING_USER_ID}" caller
groupadd  --gid="${DOCKER_GROUP_ID}" docker
# Hack: add caller (which we want to have all privileges inside the container,
# but also have the calling user's ID) to the 'root' group, giving it access to
# /root/.cache so that 'go build' works
usermod --append --groups=docker,root caller
usermod --home=/root caller
chmod g+rwx /root

# Run "go build" as "caller", to avoid littering host machine's $GOPATH with
# root-owned files
export ROOT_PATH="${PATH}"
su caller -- "$(dirname "${0}")/run_go_cmds.sh" "${@}"
