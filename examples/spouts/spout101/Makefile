# Set as you wish
DOCKER_ACCOUNT := pachyderm
CONTAINER_NAME := example-spout101
CONTAINER_VERSION := 2.0.1
CONTAINER_TAG := $(DOCKER_ACCOUNT)/$(CONTAINER_NAME):$(CONTAINER_VERSION)

docker-image:
	@docker build -t $(CONTAINER_TAG) . --platform linux/amd64
	@docker push $(CONTAINER_TAG)

deploy:
	pachctl create pipeline -f ./pipelines/spout.pipeline.json
	pachctl create pipeline -f ./pipelines/processor.pipeline.json
