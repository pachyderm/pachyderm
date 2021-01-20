
# Set as you wish
DOCKER_ACCOUNT := seerai
CONTAINER_NAME := pachyderm-rabbitmq-spout
CONTAINER_VERSION := 0.0.1
CONTAINER_TAG := $(DOCKER_ACCOUNT)/$(CONTAINER_NAME):$(CONTAINER_VERSION)

docker-image:
	@docker build -t $(CONTAINER_TAG) .
	@docker push $(CONTAINER_TAG)

create-pipeline:
	pachctl create pipeline -f pipelines/spout.pipeline.json

update-pipeline:
	pachctl update pipeline -f pipelines/spout.pipeline.json

delete-pipeline:
	pachctl delete pipeline -f rabbitmq-spout