#!/usr/bin/env bash
set -e

if [[ "${OS}" == "Windows_NT" ]]; then
  # On windows, we need to copy over the environment variables for connecting to docker
  eval "$(minikube docker-env --shell bash)"
fi

REPO_DIR=$(git rev-parse --show-toplevel)
COMPILE_DIR=$(dirname "${BASH_SOURCE[0]}")
VERSION_ADDITIONAL=-$(git log --pretty=format:%H | head -n 1)

export DOCKER_BUILDKIT=1
export GOVERSION=$(cat "${COMPILE_DIR}/GO_VERSION")
export LD_FLAGS="-X github.com/pachyderm/pachyderm/src/client/version.AdditionalVersion=${VERSION_ADDITIONAL}"

goreleaser release -p 1 --snapshot --skip-publish --rm-dist -f "${REPO_DIR}/goreleaser/docker.yml"