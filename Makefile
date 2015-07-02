.PHONY: \
	all \
	deps \
	update-deps \
	test-deps \
	update-test-deps \
	build \
	install \
	lint \
	vet \
	errcheck \
	pretest \
	test \
	test-short \
	bench \
	clean \
	container-build \
	container-shell \
	container-launch \
	container-test \
	container-test-short \
	container-clean

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

install: deps
	go install ./...

lint:
	go get -v github.com/golang/lint/golint
	golint ./...

vet:
	#go get -v golang.org/x/tools/cmd/vet
	go vet ./...

errcheck:
	errcheck ./...

pretest: lint vet errcheck

# TODO(pedge): add pretest when fixed
test: test-deps
	go test ./...

# TODO(pedge): add pretest when fixed
test-short: test-deps
	go test -test.short ./...

bench: test-deps
	go test ./... -bench . -timeout $(BENCH_TIMEOUT)

clean:
	go clean -i ./...

container-build:
	docker build -t $(PFS_IMAGE) .

container-clean:
	sudo -E bash -c 'bin/clean'

container-shell: container-build
	sudo -E bash -c 'bin/shell'

container-launch: container-build container-clean
	sudo -E bash -c 'bin/launch'

container-test: container-build container-clean
	sudo -E bash -c 'bin/run make test'

container-test-short: container-build container-clean
	sudo -E bash -c 'bin/run make test-short'

