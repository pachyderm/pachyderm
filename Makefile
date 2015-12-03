#### VARIABLES
# RUNARGS: arguments for run
# DOCKER_OPTS: docker-compose options for run, test, launch-*
# TESTPKGS: packages for test, default ./src/...
# TESTFLAGS: flags for test
# VENDOR_ALL: do not ignore some vendors when updating vendor directory
# VENDOR_IGNORE_DIRS: ignore vendor dirs
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

all: build

version:
	@echo 'package main; import "fmt"; import "github.com/pachyderm/pachyderm"; func main() { fmt.Println(pachyderm.Version.VersionString()) }' > /tmp/pachyderm_version.go
	@go run /tmp/pachyderm_version.go

deps:
	GO15VENDOREXPERIMENT=0 go get -d -v ./src/...

update-deps:
	GO15VENDOREXPERIMENT=0 go get -d -v -u -f ./src/...

test-deps:
	GO15VENDOREXPERIMENT=0 go get -d -v -t ./src/...

update-test-deps:
	GO15VENDOREXPERIMENT=0 go get -d -v -t -u -f ./src/...

vendor-update:
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 GO15VENDOREXPERIMENT=0 go get -d -v -t -u -f ./src/...

vendor-without-update:
	go get -u github.com/tools/godep
	rm -rf Godeps
	rm -rf vendor
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 godep save ./src/...
	rm -rf Godeps
	$(foreach vendor_dir, $(VENDOR_IGNORE_DIRS), rm -rf vendor/$(vendor_dir) || exit; git checkout vendor/$(vendor_dir) || exit;)

vendor: vendor-update vendor-without-update

build:
	GO15VENDOREXPERIMENT=1 go build ./src/...

install:
	GO15VENDOREXPERIMENT=1 go install ./src/cmd/pachctl

docker-build-btrfs:
	docker-compose build btrfs

docker-build-test: docker-build-btrfs
	docker-compose build test
	mkdir -p /tmp/pachyderm-test

docker-build-compile:
	docker-compose build compile

docker-build-pfs-roler: docker-build-compile
	docker-compose run --rm compile sh etc/compile/compile.sh pfs-roler

docker-build-pfsd: docker-build-btrfs docker-build-compile
	docker-compose run --rm compile sh etc/compile/compile.sh pfsd

docker-build-ppsd: docker-build-compile
	docker-compose run --rm compile sh etc/compile/compile.sh ppsd

docker-build-pachctl: docker-build-compile
	docker-compose run --rm compile sh etc/compile/compile.sh pachctl

docker-build-job-shim: docker-build-compile
	docker-compose run --rm compile sh etc/compile/compile.sh job-shim

docker-build: docker-build-pfs-roler docker-build-pfsd docker-build-ppsd docker-build-pachctl docker-build-job-shim

docker-push-pfs-roler: docker-build-pfs-roler
	docker push pachyderm/pfs-roler

docker-push-pfsd: docker-build-pfsd
	docker push pachyderm/pfsd

docker-push-ppsd: docker-build-ppsd
	docker push pachyderm/ppsd

docker-push-pachctl: docker-build-pachctl
	docker push pachyderm/pachctl

docker-push-job-shim: docker-build-job-shim
	docker push pachyderm/job-shim

docker-push: docker-push-pfs-roler docker-push-ppsd docker-push-pfsd docker-push-pachctl docker-push-job-shim

run: docker-build-test
	docker-compose run --rm $(DOCKER_OPTS) test $(RUNARGS)

launch-kube:
	etc/kube/start-kube-docker.sh

launch: install docker-build launch-kube
	pachctl create-cluster

clean-launch:
	docker kill $$(docker ps -q)

run-integration-test: docker-build-test docker-build-job-shim
	kubectl create -f etc/kube/test-pod.yml

integration-test: launch run-integration-test

proto:
	go get -u -v go.pedge.io/protolog/cmd/protoc-gen-protolog go.pedge.io/tools/protoc-all
	PROTOC_INCLUDE_PATH=src protoc-all github.com/pachyderm/pachyderm

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

docker-clean-test:
	docker-compose kill rethink
	docker-compose rm -f rethink
	docker-compose kill etcd
	docker-compose rm -f etcd
	docker-compose kill btrfs
	docker-compose rm -f btrfs

docker-clean-launch: docker-clean-test
	docker-compose kill pfs-roler
	docker-compose rm -f pfs-roler
	docker-compose kill pfsd
	docker-compose rm -f pfsd
	docker-compose kill ppsd
	docker-compose rm -f ppsd

go-test: docker-clean-test docker-build-test
	docker-compose run --rm $(DOCKER_OPTS) test sh -c "sh etc/btrfs/btrfs-mount.sh go test -test.short $(TESTFLAGS) $(TESTPKGS)"

go-test-long: docker-clean-test docker-build-test
	docker-compose run --rm $(DOCKER_OPTS) test sh -c "sh etc/btrfs/btrfs-mount.sh go test $(TESTFLAGS) $(TESTPKGS)"

test: pretest go-test docker-clean-test

test-long: pretest go-test-long docker-clean-test

clean: docker-clean-launch
	go clean ./src/...
	rm -f src/cmd/pfs/pfs-roler
	rm -f src/cmd/pfsd/pfsd
	rm -f src/cmd/ppsd/ppsd

.PHONY: \
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
	docker-build-btrfs \
	docker-build-test \
	docker-build-compile \
	docker-build-pfs-roler \
	docker-build-pfsd \
	docker-build-ppsd \
	docker-build \
	docker-push-pfs-roler \
	docker-push-pfsd \
	docker-push-ppsd \
	docker-push \
	run \
	launch-pfsd \
	launch \
	proto \
	pretest \
	docker-clean-test \
	docker-clean-launch \
	go-test \
	go-test-long \
	test \
	test-long \
	clean
