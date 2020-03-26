#!/usr/bin/env bash
set -e

REPO_DIR=$(git rev-parse --show-toplevel)
COMPILE_DIR=$(dirname "${BASH_SOURCE[0]}")
GO_VERSION=$(cat "${COMPILE_DIR}/GO_VERSION")
VERSION_ADDITIONAL=-$(git log --pretty=format:%H | head -n 1)
LD_FLAGS="-X github.com/pachyderm/pachyderm/src/client/version.AdditionalVersion=${VERSION_ADDITIONAL}"

DOCKER_OPTS=()
DOCKER_OPTS+=("--build-arg=GO_VERSION=${GO_VERSION}")
DOCKER_OPTS+=("--build-arg=LD_FLAGS=${LD_FLAGS}")
DOCKER_OPTS+=("--memory=3gb")

if [[ "${OS}" == "Windows_NT" ]]; then
  # On windows, we need to copy over the environment variables for connecting to docker
  eval "$(minikube docker-env --shell bash)"
fi

DOCKER_BUILDKIT=1 docker build "${DOCKER_OPTS[@]}" --progress plain -t pachyderm_build ${REPO_DIR}
docker build "${DOCKER_OPTS[@]}" -t pachyderm/pachd ${REPO_DIR}/etc/pachd
docker tag pachyderm/pachd pachyderm/pachd:local
docker build "${DOCKER_OPTS[@]}" -t pachyderm/worker ${REPO_DIR}/etc/worker
docker tag pachyderm/worker pachyderm/worker:local