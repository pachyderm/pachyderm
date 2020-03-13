.PHONY: docker-build docker-push

docker-build:
	docker build -t ysimonson/mnist_pachyderm_pipeline:v1.0.0 .

docker-push: docker-build
	docker push ysimonson/mnist_pachyderm_pipeline:v1.0.0
