.PHONY: docker-build docker-push

docker-build:
	docker build --tag ysimonson/mnist_kubeflow_pipeline:v1.0.0 .

docker-push: docker-build
	docker push ysimonson/mnist_kubeflow_pipeline:v1.0.0
