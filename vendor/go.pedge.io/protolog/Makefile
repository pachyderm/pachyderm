.PHONY: \
	all \
	deps \
	updatedeps \
	testdeps \
	updatetestdeps \
	build \
	install \
	lint \
	vet \
	errcheck \
	pretest \
	test \
	bench \
	clean \
	proto

all: proto test

deps:
	go get -d -v ./...

updatedeps:
	go get -d -v -u -f ./...

testdeps:
	go get -d -v -t ./...

updatetestdeps:
	go get -d -v -t -u -f ./...

build: deps
	go build ./...

install: deps
	go install ./cmd/protoc-gen-protolog

lint: testdeps
	go get -v github.com/golang/lint/golint
	for file in $$(find . -name '*.go' | grep -v '\.pb.go$$' | grep -v '\.pb.log.go$$' | grep -v 'testing/'); do \
		golint $$file; \
		if [ -n "$$(golint $$file)" ]; then \
			exit 1; \
		fi; \
	done

vet: testdeps
	-go get -v golang.org/x/tools/cmd/vet
	go vet ./...

errcheck: testdeps
	go get -v github.com/kisielk/errcheck
	errcheck ./...

pretest: lint vet errcheck

test: testdeps pretest
	go test -test.v ./...

bench: testdeps proto
	go get -v github.com/peter-edge/go-tools/go-benchmark-columns
	go test -test.v -bench . ./benchmark | go-benchmark-columns

clean:
	go clean -i ./...

proto: install
	bin/proto.bash protolog.proto
	bin/proto.bash testing/testing.proto
	bin/proto.bash benchmark/benchmark.proto
