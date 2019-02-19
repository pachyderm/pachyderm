#!/bin/bash
# TODO(msteffen): this is fairly similar to docker-build.sh in
# src/testing/loadtest/split. See if there's some way to unify

set -ex

# Setup build command. The linker flags, along with CGO_ENABLED=0 (set below)
# tell the go compiler to build a fully static binary (see comment at top)
BUILD_PATH=github.com/pachyderm/pachyderm/src/plugin/csi
BUILD_CMD="go install ./${BUILD_PATH}"

# Run compilation inside golang container, with pachyderm mounted
# Explanation of mounts
# ---------------------
# - ${GOPATH}/bin will hold the output produced by the Go compiler in the
#   container
#
# - $HOME/.cache/go-build holds cached docker builds; mounting it makes builds
#   faster:
#     /root/.cache/go-build <- $HOME/.cache/go-build
#
# - Mouting the whole pachyderm repo (including the pachyderm client and this
#   benchmark) makes this benchmark available to be built. Mounting all of it in
#   together (rather than e.g. mounting just this dir and having the pachdyerm
#   repo in ./vendor) means that parallel changes can be made to the benchmark,
#   server, and client simultaneously:
#     /go/src/.../pachyderm <- $GOPATH/src/.../pachyerm/
#
# - Mounting src/server/vendor into src/client/vendor is necessary to build the
#   pachyderm client (which the 'split' supervisor depends on, and which has
#   dependencies but has no vendor directory). It accomplishes a similar goal to
#   the mainline pachyderm's 'make docker-build' directive, which symlinks
#   src/server/vendor into the client:
#     /go/src/../pachyderm/src/client/vendor <- $GOPATH/.../pachyderm/src/server/vendor
#
# - Mounting ${GOPATH}/bin into src/client/vendor/.../pachyderm avoids a stupid
#   bug where src/client/vendor/github.com/pachyderm/pachyderm/src/client fails
#   to build because of recursive dependencies. We break the recursive
#   dependency of the client on itself by mounting a directory over it that
#   contains no go code:
#     /go/src/../pachyderm/src/client/vendor/github.com/pachyderm/ <- ./_out
PACH_PATH=src/github.com/pachyderm/pachyderm
docker run \
  -w /go/src \
  -e CGO_ENABLED=0 \
  -v "${GOPATH}/bin:/go/bin" \
  -v "${HOME}/.cache/go-build:/root/.cache/go-build" \
  -v "${GOPATH}/${PACH_PATH}:/go/${PACH_PATH}" \
  -v "${GOPATH}/${PACH_PATH}/src/server/vendor:/go/${PACH_PATH}/src/client/vendor" \
  -v "${GOPATH}/bin:/go/${PACH_PATH}/src/client/vendor/github.com/pachyderm" \
  golang:1.11 /bin/sh -c "${BUILD_CMD}"
