export DOCKER_BUILDKIT=1 
export COMPOSE_DOCKER_CLI_BUILD=1
export v=10
# Tag and push on Docker Hub
docker image build -t spout101:${v} .
docker tag spout101:${v} npepin/pachyderm-spout101:${v}
docker push npepin/pachyderm-spout101:${v}

