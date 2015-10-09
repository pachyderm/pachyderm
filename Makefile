#### VARIABLES
# RUNARGS: arguments for run
# DOCKER_OPTS: docker-compose options for run, test, launch-*
# TESTPKGS: packages for test, default ./src/...
# TESTFLAGS: flags for test
####

ifndef TESTPKGS
TESTPKGS = ./src/...
endif

all: build

version:
	@echo 'package main; import "fmt"; import "go.pachyderm.com/pachyderm"; func main() { fmt.Println(pachyderm.Version.VersionString()) }' > /tmp/pachyderm_version.go
	@go run /tmp/pachyderm_version.go

deps:
	GO15VENDOREXPERIMENT=0 go get -d -v ./src/...

update-deps:
	GO15VENDOREXPERIMENT=0 go get -d -v -u -f ./src/...

test-deps:
	GO15VENDOREXPERIMENT=0 go get -d -v -t ./src/...

update-test-deps:
	GO15VENDOREXPERIMENT=0 go get -d -v -t -u -f ./src/...

vendor:
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 GO15VENDOREXPERIMENT=0 go get -d -v -t -u -f ./src/...
	go get -u github.com/tools/godep
	rm -rf Godeps
	rm -rf vendor
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 godep save ./src/...
	rm -rf Godeps
	rm -rf vendor/gopkg.in/libgit2

install-git2go:
	sh etc/git2go/install.sh

build: deps
	go build ./src/...

install: deps
	go install ./src/cmd/pfs-volume-driver ./src/cmd/pfs ./src/cmd/pps
	go install ./vendor/go.pedge.io/dockervolume/cmd/dockervolume

docker-build-btrfs:
	docker-compose build btrfs

docker-build-test: docker-build-btrfs
	docker-compose build test
	mkdir -p /tmp/pachyderm-test

docker-build-compile:
	docker-compose build compile

docker-build-pfs-mount: docker-build-compile
	docker-compose run compile sh etc/compile/compile.sh pfs pfs-mount

docker-build-pfs-volume-driver: docker-build-compile
	docker-compose run compile sh etc/compile/compile.sh pfs-volume-driver

docker-build-pfs-roler: docker-build-compile
	docker-compose run compile sh etc/compile/compile.sh pfs-roler

docker-build-pfsd: docker-build-btrfs docker-build-compile
	docker-compose run compile sh etc/compile/compile.sh pfsd

docker-build-ppsd: docker-build-compile
	docker-compose run compile sh etc/compile/compile.sh ppsd

docker-build: docker-build-pfs-mount docker-build-pfs-volume-driver docker-build-pfs-roler docker-build-pfsd docker-build-ppsd

docker-push-pfs-mount: docker-build-pfs-mount
	docker push pachyderm/pfs-mount

docker-push-pfs-volume-driver: docker-build-pfs-volume-driver
	docker push pachyderm/pfs-volume-driver

docker-push-pfs-roler: docker-build-pfs-roler
	docker push pachyderm/pfs-roler

docker-push-pfsd: docker-build-pfsd
	docker push pachyderm/pfsd

docker-push-ppsd: docker-build-ppsd
	docker push pachyderm/ppsd

docker-push: docker-push-pfs-mount docker-push-pfs-roler docker-push-ppsd docker-push-pfsd docker-push-pfs-volume-driver

run: docker-build-test
	docker-compose run $(DOCKER_OPTS) test $(RUNARGS)

launch-pfsd: docker-clean-launch docker-build-pfs-roler docker-build-pfsd
	docker-compose up -d --force-recreate --no-build pfsd

launch: docker-clean-launch docker-build-pfs-roler docker-build-pfsd docker-build-ppsd
	docker-compose up -d --force-recreate --no-build ppsd

proto:
	go get -v go.pedge.io/tools/protoc-all
	PROTOC_INCLUDE_PATH=src protoc-all go.pachyderm.com/pachyderm

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
	errcheck $$(go list ./src/... | grep -v src/cmd/ppsd | grep -v src/pfs$$ | grep -v src/pkg/clone | grep -v src/pps$$ | grep -v src/pps/server)

docker-clean-test:
	docker-compose kill rethink
	docker-compose rm -f rethink
	docker-compose kill etcd
	docker-compose rm -f etcd
	docker-compose kill btrfs
	docker-compose rm -f btrfs

docker-clean-launch: docker-clean-test
	docker-compose kill pfs-volume-driver
	docker-compose rm -f pfs-volume-driver
	docker-compose kill pfs-roler
	docker-compose rm -f pfs-roler
	docker-compose kill pfsd
	docker-compose rm -f pfsd
	docker-compose kill ppsd
	docker-compose rm -f ppsd

test: pretest docker-clean-test docker-build-test
	docker-compose run --rm $(DOCKER_OPTS) test sh -c "sh etc/btrfs/btrfs-mount.sh go test -test.short $(TESTFLAGS) $(TESTPKGS)"

test-pfs-extra: pretest docker-clean-test docker-build-test
	docker-compose run --rm $(DOCKER_OPTS) test sh -c "sh etc/btrfs/btrfs-mount.sh go test $(TESTFLAGS) ./src/pfs/server"

test-pps-extra: pretest docker-clean-test docker-build-test
	docker-compose run --rm $(DOCKER_OPTS) test sh -c "sh etc/btrfs/btrfs-mount.sh go test $(TESTFLAGS) ./src/pps/server"

clean: docker-clean-launch
	go clean ./src/...
	rm -f src/cmd/pfs/pfs
	rm -f src/cmd/pfs/pfs-mount
	rm -f src/cmd/pfs/pfs-volume-driver
	rm -f src/cmd/pfs/pfs-roler
	rm -f src/cmd/pfsd/pfsd
	rm -f src/cmd/pps/pps
	rm -f src/cmd/ppsd/ppsd

.PHONY: \
	all \
	version \
	deps \
	update-deps \
	test-deps \
	update-test-deps \
	vendor \
	install-git2go \
	build \
	install \
	docker-build-btrfs \
	docker-build-test \
	docker-build-compile \
	docker-build-pfs-volume-driver \
	docker-build-pfs-roler \
	docker-build-pfsd \
	docker-build-ppsd \
	docker-build \
	docker-push-pfs-volume-driver \
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
	test \
	test-pfs-extra \
	test-pps-extra \
	clean
