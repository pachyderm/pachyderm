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
	go build ./...

lint: testdeps
	go get -v github.com/golang/lint/golint
	for file in $$(find . -name '*.go' | grep -v '\.pb\.go' | grep -v 'constants.go' | grep -v 'ttypes.go' | grep -v 'testing/'); do \
		golint $$file | grep -v underscore; \
		if [ -n "$$(golint $$file | grep -v underscore)" ]; then \
			exit 1; \
		fi; \
	done

vet: testdeps
	go vet ./...

errcheck: testdeps
	go get -v github.com/kisielk/errcheck
	errcheck ./...

pretest: lint vet errcheck

prototest: testdeps
	go test -v ./proto/testing

thrifttest: testdeps
	go test -v ./thrift/testing

test: pretest prototest thrifttest

clean:
	go clean -i ./...

proto:
	go get -v go.pedge.io/protoeasy/cmd/protoeasy
	go get -v go.pedge.io/pkg/cmd/strip-package-comments
	protoeasy --go --go-import-path go.pedge.io/lion .
	find . -name *\.pb\*\.go | xargs strip-package-comments

thrift:
	rm -rf gen-go
	thrift --strict --gen go thrift/thriftlion.thrift
	mv gen-go/thriftlion/* thrift/
	rm -rf gen-go
	thrift --strict --gen go thrift/testing/testing.thrift
	mv gen-go/thriftliontesting/* thrift/testing/
	rm -rf gen-go

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
	prototest \
	thrifttest \
	test \
	clean \
	proto \
	thrift
