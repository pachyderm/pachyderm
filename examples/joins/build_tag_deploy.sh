#!/bin/bash
export DOCKER_BUILDKIT=1
export COMPOSE_DOCKER_CLI_BUILD=1
export v=2
# Build your image
docker image build -t joins101:${v} .
# Tag your image -> repl
docker tag joins101:${v} npepin/pachyderm-joins101:${v}
# Push your image on Docker hub
docker push npepin/pachyderm-joins101:${v}

