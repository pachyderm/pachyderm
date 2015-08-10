#### VARIABLES
# RUNARGS: arguments for run
# DOCKER_OPTS: docker-compose options for run, test, launch-*
# TESTPKGS: packages for test, default ./...
# TESTFLAGS: flags for test
# NOCACHE: do not use dependency cache for docker
####

ifndef TESTPKGS
TESTPKGS = ./...
endif

ifdef NOCACHE
NOCACHE_CMD = touch etc/deps/deps.list
endif

all: build

version:
	@echo 'package main; import "fmt"; import "github.com/pachyderm/pachyderm"; func main() { fmt.Println(pachyderm.Version.VersionString()) }' > /tmp/pachyderm_version.go
	@go run /tmp/pachyderm_version.go

deps:
	go get -d -v ./...

update-deps:
	go get -d -v -u -f ./...

test-deps:
	go get -d -v -t ./...

update-test-deps:
	go get -d -v -t -u -f ./...

update-deps-list: test-deps
	go get -v github.com/peter-edge/go-tools/go-external-deps
	go-external-deps github.com/pachyderm/pachyderm etc/deps/deps.list

install-git2go:
	sh etc/git2go/install.sh

build: deps
	go build ./...

install: deps
	go install ./src/cmd/pfs ./src/cmd/pps

docker-build-btrfs:
	docker-compose build btrfs

docker-build-test: docker-build-btrfs
	$(NOCACHE_CMD)
	docker-compose build test

docker-build-compile:
	$(NOCACHE_CMD)
	docker-compose build compile

docker-build-pfsd: docker-build-compile
	docker-compose run compile sh etc/compile/compile.sh ppsd
	docker-compose build pfsd
	docker tag -f pachyderm_pfsd:latest pachyderm/pfsd:latest

docker-build-ppsd: docker-build-compile
	docker-compose run compile sh etc/compile/compile.sh ppsd
	docker-compose build ppsd
	docker tag -f pachyderm_ppsd:latest pachyderm/ppsd:latest

docker-push-pfsd: docker-build-pfsd
	docker push pachyderm/pfsd

docker-push-ppsd: docker-build-ppsd
	docker push pachyderm/ppsd

docker-push: docker-push-ppsd docker-push-pfsd

run: docker-build-test
	docker-compose run $(DOCKER_OPTS) test $(RUNARGS)

launch-pfsd: docker-build-btrfs docker-build-pfsd
	docker-compose kill pfsd
	docker-compose rm -f pfsd
	docker-compose up -d --force-recreate --no-build pfsd

launch-ppsd: docker-build-ppsd
	docker-compose kill ppsd
	docker-compose rm -f ppsd
	docker-compose up -d --force-recreate --no-build ppsd

launch: launch-pfsd launch-ppsd

proto:
	go get -v github.com/peter-edge/go-tools/docker-protoc-all
	docker pull pedge/protolog
	DEBUG=1 REL_PROTOC_INCLUDE_PATH=src docker-protoc-all github.com/pachyderm/pachyderm

pretest:
	go get -v github.com/kisielk/errcheck
	go get -v github.com/golang/lint/golint
	for file in $$(find "./src" -name '*.go' | grep -v '\.pb\.go'); do \
		golint $$file | grep -v unexported; \
		if [ -n "$$(golint $$file | grep -v unexported)" ]; then \
			exit 1; \
		fi; \
	done;
	go vet ./...
	errcheck ./src/cmd ./src/pfs ./src/pps

docker-clean-test:
	docker-compose kill rethink
	docker-compose rm -f rethink
	docker-compose kill btrfs
	docker-compose rm -f btrfs
	docker-compose kill etcd
	docker-compose rm -f etcd

test: pretest docker-clean-test docker-build-test
	docker-compose run --rm $(DOCKER_OPTS) test go test -test.short $(TESTFLAGS) $(TESTPKGS)

test-long: pretest docker-clean-test docker-build-test
	docker-compose run --rm $(DOCKER_OPTS) test go test $(TESTFLAGS) $(TESTPKGS)

clean: docker-clean-test
	go clean ./...
	rm -f src/cmd/pfsd/pfsd
	rm -f src/cmd/ppsd/ppsd
	sudo rm -rf _tmp

start-kube:
	docker run --net=host -d gcr.io/google_containers/etcd:2.0.12 /usr/local/bin/etcd --addr=127.0.0.1:4001 --bind-addr=0.0.0.0:4001 --data-dir=/var/etcd/data
	docker run --net=host -d -v /var/run/docker.sock:/var/run/docker.sock  gcr.io/google_containers/hyperkube:v1.0.1 /hyperkube kubelet --api_servers=http://localhost:8080 --v=2 --address=0.0.0.0 --enable_server --hostname_override=127.0.0.1 --config=/etc/kubernetes/manifests
	docker run -d --net=host --privileged gcr.io/google_containers/hyperkube:v1.0.1 /hyperkube proxy --master=http://127.0.0.1:8080 --v=2

.PHONY: \
	all \
	version \
	deps \
	update-deps \
	test-deps \
	update-test-deps \
	update-deps-list \
	install-git2go \
	build \
	install \
	docker-build-btrfs \
	docker-build-test \
	docker-build-compile \
	docker-build-pfsd \
	docker-build-ppsd \
	docker-push-pfsd \
	docker-push-ppsd \
	docker-push \
	run \
	launch-pfsd \
	launch-ppsd \
	launch \
	proto \
	pretest \
	docker-clean-test \
	test \
	test-long \
	clean \
	start-kube
