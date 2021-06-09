#! /bin/bash

export DOCKER_BUILDKIT=1 
export COMPOSE_DOCKER_CLI_BUILD=1
export v=1
# Build your image
docker image build -t spout101:${v} .
# Tag your image -> repl
docker tag spout101:${v} pachyderm/pachyderm-spout101:${v}
# Push your image on Docker hub
docker push pachyderm/pachyderm-spout101:${v}

