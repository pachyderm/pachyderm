#!/bin/bash
# Build the 'split' pipeline and supervisor images, to run in the load test

set -ex

# Clear _out, to hold build output
rm -rf ./_out || true
mkdir _out

# Set up build command.
# The linker flags, along with CGO_ENABLED=0 (set below) tell the go compiler
# to build a fully static binary. This allows us to run the supervisor
# statically (i.e. in a scratch containiner, with no libc), so that its image
# is as small as possible.
#
# The pipeline container is currently based on the "busybox" image (so that we
# can run a shell in the pipeline container and inspect its behavior), but
# because busybox includes musl libc, the pipeline binary doesn't need to be
# entirely static
LD_FLAGS="-extldflags -static"
BUILD_PATH=github.com/pachyderm/pachyderm/src/testing/loadtest/split
# -a tells go install to ignore the existing build artifact and rebuild.
# TODO(msteffen): I've added this flag out of desparation--I don't think it
# should be necessary, but it seems to be the only way to force source changes
# to take effect
BUILD_CMD="
go install -a -ldflags \"${LD_FLAGS}\" ./${BUILD_PATH}/cmd/pipeline && \
go install -a ./${BUILD_PATH}/cmd/supervisor && \
mv ../bin/* /out/"

# Compile 'supervisor' and 'worker' inside golang container, with pachyderm
# mounted.
#
# Explanation of mounts
# ---------------------
# - _out is where the binaries built in this docker image are written out for
#   us to use:
#     /out <- ./_out
#
# - $HOME/.cache/go-build holds cached go builds; mounting it makes builds
#   faster:
#     /root/.cache/go-build <- $HOME/.cache/go-build
#
# - Mouting the whole pachyderm repo (including the pachyderm client and this
#   benchmark) makes all other dependencies in the Pachyderm repo available to
#   the compiler. Mounting all of it into the container (rather than e.g.
#   mounting just this dir and having the pachdyerm repo in ./vendor) means
#   that parallel changes can be made to the benchmark, server, and client
#   (which must be vendored) simultaneously:
#     /go/src/.../pachyderm <- $GOPATH/src/.../pachyerm/
#
# - Mounting src/server/vendor into src/client/vendor is necessary to build the
#   pachyderm client (which the 'split' supervisor depends on, and which has
#   dependencies but has no vendor directory). It accomplishes a similar goal to
#   the mainline pachyderm's 'make docker-build' directive, which symlinks
#   src/server/vendor into the client:
#     /go/src/../pachyderm/src/client/vendor <- $GOPATH/.../pachyderm/src/server/vendor
#
# - Mounting ./_out into src/client/vendor/.../pachyderm avoids a stupid bug
#   where src/client/vendor/github.com/pachyderm/pachyderm/src/client fails to
#   build because of recursive dependencies. We break the recursive dependency
#   of the client on itself by mounting a directory over it that contains no go
#   code:
#     /go/src/../pachyderm/src/client/vendor/github.com/pachyderm/ <- ./_out
#  -v "${GOPATH}/${PACH_PATH}/src/server/vendor:/go/${PACH_PATH}/src/client/vendor" \
#  -v "${PWD}/_out:/go/${PACH_PATH}/src/client/vendor/github.com/pachyderm" \
PACH_PATH=src/github.com/pachyderm/pachyderm
docker run \
  -w /go/src \
  -e CGO_ENABLED=0 \
  -e GO111MODULE=off \
  -v "${PWD}/_out:/out" \
  -v "${HOME}/.cache/go-build:/root/.cache/go-build" \
  -v "${GOPATH}/${PACH_PATH}:/go/${PACH_PATH}" \
  golang:1.11 /bin/sh -c "${BUILD_CMD}"

# Now that the 'pipeline' and 'supervisor' binaries have been built, build The
# docker images containing them that we'll deploy
# Note: if BENCH_VERSION is set, then this adds :$BENCH_VERSION to the image tag
# (to allow deploying multiple versions of the pipeline/supervisor)
declare -A base_image=(
  [pipeline]=busybox:1.31.1-musl
  [supervisor]=scratch
)
for bin in supervisor pipeline; do
	image_name="pachyderm/split-loadtest-${bin}${BENCH_VERSION:+:$BENCH_VERSION}"
	echo -e "FROM ${base_image[$bin]}\nCOPY ${bin} /" >_out/Dockerfile
	docker build -t ${image_name} ./_out
	docker push ${image_name}
done

rm -rf ./_out
