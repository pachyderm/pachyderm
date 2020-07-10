.PHONY: docker-build docker-push

docker-build:
	ls | grep -v Makefile | xargs -I {} docker build -t pachyderm/{}-build:$(shell $(GOPATH)/bin/pachctl version --client-only) {}

docker-push:
	ls | grep -v Makefile | xargs -I {} docker push pachyderm/{}-build:$(shell $(GOPATH)/bin/pachctl version --client-only)
