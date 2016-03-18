#### VARIABLES
# RUNARGS: arguments for run
# DOCKER_OPTS: docker-compose options for run, test, launch-*
# TESTPKGS: packages for test, default ./src/...
# TESTFLAGS: flags for test
# VENDOR_ALL: do not ignore some vendors when updating vendor directory
# VENDOR_IGNORE_DIRS: ignore vendor dirs
# KUBECTLFLAGS: flags for kubectl
####

ifndef TESTPKGS
	TESTPKGS = ./src/...
endif
ifndef VENDOR_IGNORE_DIRS
	VENDOR_IGNORE_DIRS = go.pedge.io
endif
ifdef VENDOR_ALL
	VENDOR_IGNORE_DIRS =
endif

COMPILE_RUN_ARGS = -v /var/run/docker.sock:/var/run/docker.sock --privileged=true

all: build

version:
	@echo 'package main; import "fmt"; import "github.com/pachyderm/pachyderm"; func main() { fmt.Println(pachyderm.Version.VersionString()) }' > /tmp/pachyderm_version.go
	@go run /tmp/pachyderm_version.go

deps:
	GO15VENDOREXPERIMENT=0 go get -d -v ./src/... ./.

update-deps:
	GO15VENDOREXPERIMENT=0 go get -d -v -u -f ./src/... ./.

test-deps:
	GO15VENDOREXPERIMENT=0 go get -d -v -t ./src/... ./.

update-test-deps:
	GO15VENDOREXPERIMENT=0 go get -d -v -t -u -f ./src/... ./.

vendor-update:
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 GO15VENDOREXPERIMENT=0 go get -d -v -t -u -f ./src/... ./.

vendor-without-update:
	go get -v github.com/kardianos/govendor
	rm -rf vendor
	govendor init
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 govendor add +external
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 govendor update +vendor
	$(foreach vendor_dir, $(VENDOR_IGNORE_DIRS), rm -rf vendor/$(vendor_dir) || exit; git checkout vendor/$(vendor_dir) || exit;)

vendor: vendor-update vendor-without-update

build:
	GO15VENDOREXPERIMENT=1 go build ./src/... ./.

install:
	# GOPATH/bin must be on your PATH to access these binaries:
	GO15VENDOREXPERIMENT=1 go install ./src/cmd/pachctl ./src/cmd/pachctl-doc

docker-build-test:
	docker build -t pachyderm/testqwertyuio .

docker-build-compile:
	docker build -t pachyderm_compile .

docker-build-job-shim: docker-build-compile
	docker run $(COMPILE_RUN_ARGS) pachyderm_compile sh etc/compile/compile.sh job-shim

docker-build-pachd: docker-build-compile
	docker run $(COMPILE_RUN_ARGS) pachyderm_compile sh etc/compile/compile.sh pachd

docker-build-hyperkube:
	docker build -t privileged_hyperkube etc/kube

docker-build: docker-build-test docker-build-job-shim docker-build-pachd


docker-push-test: docker-build-test
	docker push pachyderm/testqwertyuio

docker-push-job-shim: docker-build-job-shim
	docker push pachyderm/job-shim

docker-push-pachd: docker-build-pachd
	docker push pachyderm/pachd

docker-push: docker-push-job-shim docker-push-pachd

launch-kube: docker-build-hyperkube
	etc/kube/start-kube-docker.sh

clean-launch-kube:
	docker kill $$(docker ps -q)

kube-cluster-assets: install
	pachctl manifest -s 32 >etc/kube/pachyderm.json

launch:
	kubectl $(KUBECTLFLAGS) create -f etc/kube/pachyderm.json

launch-dev: launch-kube launch

clean-launch:
	kubectl $(KUBECTLFLAGS) delete --ignore-not-found job -l suite=pachyderm
	kubectl $(KUBECTLFLAGS) delete --ignore-not-found all -l suite=pachyderm
	kubectl $(KUBECTLFLAGS) delete --ignore-not-found serviceaccount -l suite=pachyderm
	kubectl $(KUBECTLFLAGS) delete --ignore-not-found secret -l suite=pachyderm

integration-tests-local: docker-push-test integration-tests

integration-tests: 
	kubectl $(KUBECTLFLAGS) delete --ignore-not-found pod integrationtests
	#kubectl $(KUBECTLFLAGS) run integrationtests -i --image pachyderm/testqwertyuio --restart=Never --command -- echo $$PFSD_PORT_650_TCP_ADDR
	kubectl $(KUBECTLFLAGS) run integrationtests -i --image pachyderm/testqwertyuio --restart=Never --command -- go test . -timeout 60s

proto:
	go get -v go.pedge.io/protoeasy/cmd/protoeasy
	protoeasy --grpc --grpc-gateway --go --go-import-path github.com/pachyderm/pachyderm/src src
	go install github.com/pachyderm/pachyderm/src/cmd/protofix
	protofix fix src

pretest:
	go get -v github.com/kisielk/errcheck
	go get -v github.com/golang/lint/golint
	for file in $$(find "./src" -name '*.go' | grep -v '\.pb\.go' | grep -v '\.pb\.gw\.go'); do \
		golint $$file | grep -v unexported; \
		if [ -n "$$(golint $$file | grep -v unexported)" ]; then \
		exit 1; \
		fi; \
		done;
	go vet -n ./src/... | while read line; do \
		modified=$$(echo $$line | sed "s/ [a-z0-9_/]*\.pb\.gw\.go//g"); \
		$$modified; \
		if [ -n "$$($$modified)" ]; then \
		exit 1; \
		fi; \
		done
	#errcheck $$(go list ./src/... | grep -v src/cmd/ppsd | grep -v src/pfs$$ | grep -v src/pps$$)

test: pretest docker-build clean-launch launch integration-tests

localtest: 
	GO15VENDOREXPERIMENT=1 go test -v -short $$(go list ./... | grep -v '/vendor/')

clean: clean-launch clean-launch-kube

doc: install
	# we rename to pachctl because the program name is used in generating docs
	cp $(GOPATH)/bin/pachctl-doc ./pachctl
	rm -rf doc/pachctl && mkdir doc/pachctl
	./pachctl
	rm ./pachctl

grep-data:
	go run examples/grep/generate.go >examples/grep/set1.txt
	go run examples/grep/generate.go >examples/grep/set2.txt

grep-example:
	sh examples/grep/run.sh

logs:
	kubectl get pod -l app=pachd | sed '1d' | cut -f1 -d ' ' | xargs -n 1 -I pod sh -c 'kubectl logs pod >pod'

.PHONY: \
	doc \
	all \
	version \
	deps \
	update-deps \
	test-deps \
	update-test-deps \
	vendor-update \
	vendor-without-update \
	vendor \
	build \
	install \
	docker-build-test \
	docker-build-compile \
	docker-build \
	docker-build-pachd \
	docker-push \
	docker-push-pachd \
	run \
	launch \
	proto \
	pretest \
	docker-clean-test \
	go-test \
	go-test-long \
	test \
	test-long \
	clean
