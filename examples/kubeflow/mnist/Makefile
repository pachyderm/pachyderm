.PHONY: docker-build docker-push mnist

docker-build:
	docker build -t pachyderm/mnist_pipeline:v1.0.1 .

docker-push: docker-build
	docker push pachyderm/mnist_pipeline:v1.0.1

mnist:
	pachctl create repo input-repo
	pachctl put file input-repo@master:/mnist.npz -f mnist.npz
	pachctl create pipeline -f pipeline.yaml
