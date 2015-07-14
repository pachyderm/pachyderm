.PHONY: \
	all \
	deps \
	update-deps \
	test-deps \
	update-test-deps \
	build \
	install \
	clean \
	container-build \
	container-clean \
	container-shell \
	container-launch \
	lint \
	vet \
	errcheck \
	pretest \
	test-long \
	test \
	bench \
	proto

include etc/env/pfs.env

BENCH_TIMEOUT = "20m"

ifndef GOMAXPROCS
GOMAXPROCS = 20
endif

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

clean:
	go clean -i ./...

container-build:
	docker build -t $(PFS_IMAGE) .

container-clean:
	sudo -E bash -c 'bin/clean'

container-shell:
	sudo -E bash -c 'bin/shell'

container-launch:
	sudo -E bash -c 'bin/launch'

lint:
	go get -v github.com/golang/lint/golint
	golint ./...

vet:
	go vet ./...

errcheck:
	errcheck ./...

pretest: lint vet errcheck

# TODO(pedge): add pretest when fixed
test:
	sudo -E bash -c 'bin/run go test -parallel $(GOMAXPROCS) -test.short ./...'

# TODO(pedge): add pretest when fixed
test-long:
	sudo -E bash -c 'bin/run go test -parallel $(GOMAXPROCS) ./...'

# TODO(pedge): add pretest when fixed
bench:
	sudo -E bash -c 'bin/run go test -parallel $(GOMAXPROCS) -bench . -timeout $(BENCH_TIMEOUT) ./...'

proto:
	docker pull pedge/proto3grpc
	docker run \
		--volume $(shell pwd):/compile \
		--workdir /compile \
		pedge/proto3grpc \
		protoc \
		-I /usr/include \
		-I /compile/src/pfs \
		--go_out=plugins=grpc,Mgoogle/protobuf/wrappers.proto=github.com/peter-edge/go-google-protobuf:/compile/src/pfs \
		/compile/src/pfs/pfs.proto
