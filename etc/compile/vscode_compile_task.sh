#!/usr/bin/env bash
COMPILE_IMAGE="pachyderm/compile:$(cat etc/compile/GO_VERSION)"
COMPILE_RUN_ARGS="-d -v /var/run/socker.sock:/var/run/docker.sock --privileged=true"
VERSION_ADDITIONAL=-$(git log --pretty=format:%H | head -n 1)
LD_FLAGS="-X github.com/pachyderm/pachyderm/src/client/version.AdditionalVersion=${VERSION_ADDITIONAL}"

if [[ -z "${1}" ]]; then
  echo "Name of target must be specified, usage: ${0} ( worker | pachd )"
  exit 1
fi
  
# Normalize paths for windows (some environment variables are already converted in bash for windows)
norm_path () {
  cd ${1} 2>&1 >/dev/null
  echo -n ${PWD}
  cd - 2>&1 >/dev/null
} 

TYPE=$1
NAME=${TYPE}_compile

DOCKER_OPTS=()
# TODO: this is probably needed for the Makefile build to run worker/pachd in parallel
# DOCKER_OPTS+=("-d")
DOCKER_OPTS+=("--name ${NAME}")
DOCKER_OPTS+=("-v ${PWD}:/go/src/github.com/pachyderm/pachyderm")
DOCKER_OPTS+=("-v ${HOME}/.cache/go-build:/root/.cache/go-build")

if [[ "${OS}" == "Windows_NT" ]]; then
  # Convert GOPATH from a normal windows path to a bash path
  DOCKER_OPTS+=("-v $(norm_path ${GOPATH})/pkg:/go/pkg")

  # On windows, we need to copy over the environment variables for connecting to docker
  eval $(minikube docker-env --shell bash)
  DOCKER_OPTS+=("--env DOCKER_HOST=${DOCKER_HOST}")

  if [[ "${DOCKER_TLS_VERIFY}" -eq "1" ]]; then
    DOCKER_OPTS+=("--env DOCKER_TLS_VERIFY=1")
    DOCKER_OPTS+=("-v $(norm_path ${DOCKER_CERT_PATH}):/mnt/docker-certs")
    DOCKER_OPTS+=("--env DOCKER_CERT_PATH=/mnt/docker-certs")
  fi

  # On windows, containers have a default 1GB memory limit, bump it because that is not enough
  DOCKER_OPTS+=("--memory 3gb")
else
  # On linux we can just expose the docker socket
  DOCKER_OPTS+=("-v /var/run/docker.sock:/var/run/docker.sock")
  DOCKER_OPTS+=("-v ${GOPATH}/pkg:/go/pkg")
fi

docker stop ${NAME} > /dev/null
docker rm ${NAME} > /dev/null

MSYS_NO_PATHCONV=1 docker run ${DOCKER_OPTS[@]} ${COMPILE_IMAGE} \
  /go/src/github.com/pachyderm/pachyderm/etc/compile/compile.sh ${TYPE} "${LD_FLAGS}"