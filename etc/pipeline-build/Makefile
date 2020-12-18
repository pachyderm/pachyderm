.PHONY: docker-build push-to-minikube

export DOCKER_BUILDKIT = 1

docker-build:
	ls | grep -v Makefile | xargs -I {} bash -c 'cd ../..; docker build -t pachyderm/{}-build:$(shell $(GOPATH)/bin/pachctl version --client-only) -f etc/pipeline-build/{}/Dockerfile .'

push-to-minikube: docker-build
	ls | grep -v Makefile | xargs -I {} ../kube/push-to-minikube.sh pachyderm/{}-build:$(shell $(GOPATH)/bin/pachctl version --client-only)
