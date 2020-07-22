.PHONY: docker-build docker-push push-to-minikube

docker-build:
	ls | grep -v Makefile | xargs -I {} docker build -t pachyderm/{}-build:$(shell $(GOPATH)/bin/pachctl version --client-only) {}

push-to-minikube: docker-build
	ls | grep -v Makefile | xargs -I {} ../kube/push-to-minikube.sh pachyderm/{}-build:$(shell $(GOPATH)/bin/pachctl version --client-only)

docker-push: docker-build
	ls | grep -v Makefile | xargs -I {} docker push pachyderm/{}-build:$(shell $(GOPATH)/bin/pachctl version --client-only)
