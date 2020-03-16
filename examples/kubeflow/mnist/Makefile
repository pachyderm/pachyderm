.PHONY: docker-build docker-push mnist

docker-build:
	docker build -t ysimonson/mnist_pachyderm_pipeline:v1.1.9 .

docker-push: docker-build
	docker push ysimonson/mnist_pachyderm_pipeline:v1.1.9

mnist:
	pachctl create repo input-repo
	pachctl put file input-repo@master:/mnist.npz -f mnist.npz
	pachctl create pipeline -f pipeline.yaml
