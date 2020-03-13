VERSION = $(shell jq -r .kubeflow_pipeline ../version.json)

.PHONY: docker-build docker-push kubeflow-build

docker-build:
	docker build --tag ysimonson/mnist_kubeflow_pipeline:v$(VERSION) .

docker-push: docker-build
	docker push ysimonson/mnist_kubeflow_pipeline:v$(VERSION)

kubeflow-build:
	dsl-compile --py pipeline.py --output pipeline.tar.gz
	@echo "pipeline built, please upload ${PWD}/pipeline.tar.gz to kubeflow"
