.PHONY: \
	all \
	deps \
	updatedeps \
	testdeps \
	updatetestdeps \
	build \
	lint \
	vet \
	errcheck \
	pretest \
	test \
	clean

include etc/env/pfs.env

BENCH_TIMEOUT = "20m"

all: test

print-%:
	@echo $* = $($*)

deps:
	go get -d -v ./...

update-deps:
	go get -d -v -u -f ./...

test-deps:
	go get -d -v -t ./...

update-test-deps:
	go get -d -v -t -u -f ./...

build: deps
	go build ./...

build-container:
	docker build -t $(PFS_IMAGE) .

build-container-dev:
	docker build -t $(PFS_DEV_IMAGE) .

launch-clean:
	-umount $(PFS_HOST_VOLUME)
	-rm $(PFS_DATA_IMG)
	-docker kill $(PFS_CONTAINER_NAME)
	-docker rm $(PFS_CONTAINER_NAME)

launch: build-container launch-clean
	bash bin/launch

launch-dev: build-container-dev launch-clean
	PFS_IMAGE=$(PFS_DEV_IMAGE) bash bin/launch

lint: testdeps
	go get -v github.com/golang/lint/golint
	golint ./...

vet:
	go get -v golang.org/x/tools/cmd/vet
	go vet ./...

errcheck:
	errcheck ./...

pretest: lint vet errcheck

test: pretest testdeps
	go test github.com/pachyderm/pfs/services/...

test-short: pretest testdeps
	go test -test.short github.com/pachyderm/pfs/services/...

bench: testdeps
	go test github.com/pachyderm/pfs/services/... -bench . -timeout $(BENCH_TIMEOUT)

container-test-short: build-container-dev
	PFS_IMAGE=$(PFS_DEV_IMAGE) bash bin/run make test-short

clean:
	go clean -i ./...
