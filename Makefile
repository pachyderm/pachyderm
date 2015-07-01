.PHONY: \
	all \
	deps \
	update-deps \
	test-deps \
	update-test-deps \
	build \
	lint \
	vet \
	errcheck \
	pretest \
	test \
	test-short \
	bench \
	container-build \
	container-build-dev \
	container-launch \
	container-launch-dev \
	container-test-short \
	container-clean \
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

lint:
	go get -v github.com/golang/lint/golint
	golint ./...

vet:
	go get -v golang.org/x/tools/cmd/vet
	go vet ./...

errcheck:
	errcheck ./...

pretest: lint vet errcheck

# TODO(pedge): add pretest when fixed
test: testdeps
	go test ./...

# TODO(pedge): add pretest when fixed
test-short: testdeps
	go test -test.short ./...

bench: testdeps
	go test ./... -bench . -timeout $(BENCH_TIMEOUT)

clean:
	go clean -i ./...

container-build:
	docker build -t $(PFS_IMAGE) .

container-build-dev:
	docker build -t $(PFS_DEV_IMAGE) .

container-launch: container-build
	sudo bash bin/launch

container-launch-dev: container-build-dev
	sudo PFS_IMAGE=$(PFS_DEV_IMAGE) bash bin/launch

container-test-short: container-build-dev
	sudo PFS_IMAGE=$(PFS_DEV_IMAGE) bash bin/run make test-short

container-clean:
	sudo bash bin/clean

container-shell:
	sudo bash bin/shell
