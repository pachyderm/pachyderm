# Set as you wish
DOCKER_ACCOUNT := pachyderm
CONTAINER_NAME := example-no-skip
CONTAINER_VERSION := 11
CONTAINER_TAG := $(DOCKER_ACCOUNT)/$(CONTAINER_NAME):$(CONTAINER_VERSION)

docker-image:
	@docker build -t $(CONTAINER_TAG) . --platform linux/amd64
	@docker push $(CONTAINER_TAG)

setup:
	mkdir -p ./model
	printf "this is file 1" \
	>./model/file1.txt
	printf "this is file 2" \
	>./model/file2.txt
	printf "this is file 3" \
	>./model/file3.txt

create:
	pachctl create repo model
	pachctl put file -r model@master:/ -f model
	pachctl create pipeline -f no-skip.json
	pachctl create pipeline -f skip.json

update:
	printf "modifying file 2 again" \
	>./model/file2.txt
	cd model && pachctl put file model@master -f file2.txt -o

addnewfile:
	printf "this is file 4" \
	>./model/file4.txt
	cd model && pachctl put file model@master -f file4.txt 



