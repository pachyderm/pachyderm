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
	clean \
	proto

all: test

deps:
	go get -d -v ./...

updatedeps:
	go get -d -v -u -f ./...

testdeps:
	go get -d -v -t ./...

updatetestdeps:
	go get -d -v -t -u -f ./...

build: deps
	GOOS=linux go build ./...

install: deps
	go install ./...

lint: testdeps
	go get -v github.com/golang/lint/golint
	for file in $$(find . -name '*.go' | grep -v '\.pb\.go' | grep -v '\.pb\.gw\.go' | grep -v '\.pb\.log\.go'); do \
		golint $$file; \
		if [ -n "$$(golint $$file)" ]; then \
			exit 1; \
		fi; \
	done

vet: testdeps
	go vet ./... 2>&1 | grep -v 'not a string in call to Errorf'

errcheck: testdeps
	go get -v github.com/kisielk/errcheck
	errcheck ./...

pretest: lint vet errcheck

test: testdeps pretest
	GOOS=linux go test -test.v ./...

clean:
	go clean ./...
	GOOS=linux go clean ./...
	go clean -i ./...

proto:
	go get -v go.pedge.io/tools/protoc-all
	STRIP_PACKAGE_COMMENTS=1 protoc-all go.pedge.io/dockervolume
