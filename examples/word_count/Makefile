# Set as you wish
DOCKER_ACCOUNT := pachyderm
CONTAINER_NAME := example-wordcount
CONTAINER_VERSION := 2.0.3
CONTAINER_TAG := $(DOCKER_ACCOUNT)/$(CONTAINER_NAME):$(CONTAINER_VERSION)
PROJECT ?= default

docker-image:
	@docker buildx build --platform linux/amd64 -t $(CONTAINER_TAG)-amd64 .
	@docker buildx build --platform linux/arm64 -t $(CONTAINER_TAG)-arm64 .
	@docker push $(CONTAINER_TAG)-amd64
	@docker push $(CONTAINER_TAG)-arm64
	#@docker manifest rm $(CONTAINER_TAG)
	@docker manifest create $(CONTAINER_TAG) $(CONTAINER_TAG)-arm64 $(CONTAINER_TAG)-amd64
	@docker manifest push $(CONTAINER_TAG)


wordcount:
	pachctl create repo urls --project $(PROJECT)
	cd data && pachctl put file urls@master -f Wikipedia --project $(PROJECT)
	pachctl create pipeline -f pipelines/scraper.pipeline.json --project $(PROJECT)
	pachctl create pipeline -f pipelines/map.pipeline.json --project $(PROJECT)
	pachctl create pipeline -f pipelines/reduce.pipeline.json --project $(PROJECT)

clean:
	pachctl delete pipeline reduce --project $(PROJECT)
	pachctl delete pipeline map --project $(PROJECT)
	pachctl delete pipeline scraper --project $(PROJECT)
	pachctl delete repo urls --project $(PROJECT)
	pachctl delete repo scraper	--project $(PROJECT)
	pachctl delete repo map	--project $(PROJECT)
	pachctl delete repo reduce --project $(PROJECT)


.PHONY:
	wordcount
