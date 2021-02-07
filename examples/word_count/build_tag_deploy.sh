#!/bin/bash
export DOCKER_BUILDKIT=1
export COMPOSE_DOCKER_CLI_BUILD=1
export v=3
# Build your image
docker image build --platform=linux/amd64 -t word-count:${v} .
# Tag your image -> repl
docker tag word-count:${v} npepin/word-count:${v}
# Push your image on Docker hub
docker push npepin/word-count:${v}

