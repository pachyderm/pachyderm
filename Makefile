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
	GO15VENDOREXPERIMENT=1 go install ./src/cmd/pachctl ./src/cmd/pachctl-doc

docker-build-test:
	docker-compose build test
	docker tag -f pachyderm_test:latest pachyderm/test:latest
	mkdir -p /tmp/pachyderm-test

docker-build-compile:
	docker-compose build compile

docker-build-pfs-roler: docker-build-compile
	docker-compose run --rm compile sh etc/compile/compile.sh pfs-roler

docker-build-pfsd: docker-build-compile
	docker-compose run --rm compile sh etc/compile/compile.sh pfsd

docker-build-ppsd: docker-build-compile
	docker-compose run --rm compile sh etc/compile/compile.sh ppsd

docker-build-objd: docker-build-compile
	docker-compose run --rm compile sh etc/compile/compile.sh objd

docker-build-job-shim: docker-build-compile
	docker-compose run --rm compile sh etc/compile/compile.sh job-shim

docker-build: docker-build-test docker-build-pfs-roler docker-build-pfsd docker-build-ppsd docker-build-objd docker-build-job-shim

docker-push-test: docker-build-test
	docker push pachyderm/test

docker-push-pfs-roler: docker-build-pfs-roler
	docker push pachyderm/pfs-roler

docker-push-pfsd: docker-build-pfsd
	docker push pachyderm/pfsd

docker-push-ppsd: docker-build-ppsd
	docker push pachyderm/ppsd

docker-push-objd: docker-build-objd
	docker push pachyderm/objd

docker-push-job-shim: docker-build-job-shim
	docker push pachyderm/job-shim

docker-push: docker-push-pfs-roler docker-push-ppsd docker-push-objd docker-push-pfsd docker-push-job-shim

run: docker-build-test
	docker-compose run --rm $(DOCKER_OPTS) test $(RUNARGS)

launch-kube:
	etc/kube/start-kube-docker.sh

clean-launch-kube:
	docker kill $$(docker ps -q)

kube-cluster-assets: install
	pachctl manifest -s 32 >etc/kube/pachyderm.json

launch: install
	kubectl $(KUBECTLFLAGS) create -f etc/kube/pachyderm.json
	until pachctl version 2>/dev/null >/dev/null; do sleep 5; done

launch-dev: launch-kube launch

clean-job:

clean-launch:
	kubectl $(KUBECTLFLAGS) delete --ignore-not-found job -l suite=pachyderm
	kubectl $(KUBECTLFLAGS) delete --ignore-not-found all -l suite=pachyderm
	kubectl $(KUBECTLFLAGS) delete --ignore-not-found serviceaccount -l suite=pachyderm
	kubectl $(KUBECTLFLAGS) delete --ignore-not-found secret -l suite=pachyderm

run-integration-test:
	kubectl $(KUBECTLFLAGS) delete --ignore-not-found -f etc/kube/test-pod.yml
	kubectl $(KUBECTLFLAGS) create -f etc/kube/test-pod.yml

integration-test: launch run-integration-test

proto:
	go get -v go.pedge.io/protoeasy/cmd/protoeasy
	protoeasy --grpc --grpc-gateway --go --go-import-path github.com/pachyderm/pachyderm/src src

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

docker-clean-launch: docker-clean-test
	docker-compose kill pfs-roler
	docker-compose rm -f pfs-roler
	docker-compose kill pfsd
	docker-compose rm -f pfsd
	docker-compose kill ppsd
	docker-compose rm -f ppsd

go-test: docker-clean-test docker-build-test
	docker-compose run --rm $(DOCKER_OPTS) test sh -c "go test -test.short $(TESTFLAGS) $(TESTPKGS)"

go-test-long: docker-clean-test docker-build-test
	docker-compose run --rm $(DOCKER_OPTS) test sh -c "go test $(TESTFLAGS) $(TESTPKGS)"

test: pretest go-test docker-clean-test

test-long: pretest go-test-long docker-clean-test

clean: docker-clean-launch clean-launch clean-launch-kube
	go clean ./src/... ./.
	rm -f src/cmd/pfs/pfs-roler
	rm -f src/cmd/pfsd/pfsd
	rm -f src/cmd/ppsd/ppsd
	rm -f src/cmd/objd/objd

doc: install
	# we rename to pachctl because the program name is used in generating docs
	cp $(GOPATH)/bin/pachctl-doc ./pachctl
	rm -rf doc/pachctl && mkdir doc/pachctl
	./pachctl
	rm ./pachctl

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
	docker-build-pfs-roler \
	docker-build-pfsd \
	docker-build-ppsd \
	docker-build \
	docker-push-pfs-roler \
	docker-push-pfsd \
	docker-push-ppsd \
	docker-push \
	run \
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
