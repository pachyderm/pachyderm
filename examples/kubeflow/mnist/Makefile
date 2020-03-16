.PHONY: docker-build docker-push kubeflow-build delete-old-kubeflow-jobs mnist

MNIST_PIPELINE_VERSION = $(shell jq -r .kubeflow_pipeline version.json)
PACHYDERM_PIPELINE_VERSION = $(shell jq -r .pachyderm_pipeline version.json)

docker-build:
	docker build -t ysimonson/mnist_kubeflow_pipeline:v$(MNIST_PIPELINE_VERSION) -f Dockerfile.kf .
	docker build -t ysimonson/mnist_pachyderm_pipeline:v$(PACHYDERM_PIPELINE_VERSION) .

docker-push: docker-build
	docker push ysimonson/mnist_kubeflow_pipeline:v$(MNIST_PIPELINE_VERSION)
	docker push ysimonson/mnist_pachyderm_pipeline:v$(PACHYDERM_PIPELINE_VERSION)

clean:
	kubectl delete pod $$(kubectl get pod --output=json | jq '.items[].metadata.name' -r | grep mnist-kubeflow-pipeline) || exit 0
	echo "y\n" | pachctl delete all

mnist:
	pachctl create repo input-repo
	pachctl put file input-repo@master:/mnist.npz -f mnist.npz
	pachctl create pipeline -f pipeline.yaml
