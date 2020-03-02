VERSION=v0.0.1
IMAGE=pachyderm/ubuntus3clients:$(VERSION)

image:
	docker build -t $(IMAGE) .

push-to-minikube: image
	../../../../etc/kube/push-to-minikube.sh $(IMAGE)

push-to-dockerhub: image
	docker push $(IMAGE)

.PHONY: \
	image \
	push-to-minikube \
	push-to-dockerhub \
