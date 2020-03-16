.PHONY: docker-build docker-push clean mnist

PACHYDERM_PIPELINE_VERSION = $(shell jq -r .pachyderm_pipeline version.json)

docker-build:
	docker build -t ysimonson/mnist_pachyderm_pipeline:v1.1.8 .

docker-push: docker-build
	docker push ysimonson/mnist_pachyderm_pipeline:v1.1.8

clean:
	echo "y\n" | pachctl delete all

mnist:
	pachctl create repo input-repo
	pachctl put file input-repo@master:/mnist.npz -f mnist.npz
	pachctl create pipeline -f pipeline.yaml
