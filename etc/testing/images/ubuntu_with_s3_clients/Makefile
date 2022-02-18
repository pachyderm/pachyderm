VERSION=v0.0.2
IMAGE=pachyderm/ubuntu-with-s3-clients:$(VERSION)

image:
	docker buildx build --platform linux/amd64 -t $(IMAGE)-amd64 .
	docker buildx build --platform linux/arm64 -t $(IMAGE)-arm64 .

push-to-minikube: image
	../../../../etc/kube/push-to-minikube.sh $(IMAGE)

push-to-dockerhub: image
	docker push $(IMAGE)-amd64
	docker push $(IMAGE)-arm64
	docker manifest create $(IMAGE) $(IMAGE)-arm64 $(IMAGE)-amd64 --amend

.PHONY: \
	image \
	push-to-minikube \
	push-to-dockerhub \
