# Set as you wish
DOCKER_ACCOUNT := pachyderm
CONTAINER_NAME := example-wordcount
CONTAINER_VERSION := 1.12.2
CONTAINER_TAG := $(DOCKER_ACCOUNT)/$(CONTAINER_NAME):$(CONTAINER_VERSION)

docker-image:
	@docker build -t $(CONTAINER_TAG) .
	@docker push $(CONTAINER_TAG)


wordcount:
	pachctl create repo urls
	cd data && pachctl put file urls@master -f Wikipedia
	pachctl create pipeline -f pipelines/scraper.json
	pachctl create pipeline -f pipelines/map.json
	pachctl create pipeline -f pipelines/reduce.json

clean:
	pachctl delete pipeline reduce
	pachctl delete pipeline map
	pachctl delete pipeline scraper
	pachctl delete repo urls
	pachctl delete repo scraper	
	pachctl delete repo map	
	pachctl delete repo reduce


.PHONY:
	wordcount
