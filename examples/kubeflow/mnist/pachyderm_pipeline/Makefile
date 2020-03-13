VERSION = $(shell jq -r .pachyderm_pipeline ../version.json)

.PHONY: docker-build docker-push

docker-build:
	docker build -t ysimonson/mnist_pachyderm_pipeline:v$(VERSION) .

docker-push: docker-build
	docker push ysimonson/mnist_pachyderm_pipeline:v$(VERSION)
