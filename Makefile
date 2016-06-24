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
CLUSTER_NAME = pachyderm
MANIFEST = etc/kube/pachyderm-versioned.json
DEV_MANIFEST = etc/kube/pachyderm.json
LD_FLAGS = -X github.com/pachyderm/pachyderm/src/server/vendor/github.com/pachyderm/pachyderm/src/client/version.AdditionalVersion=$(VERSION_ADDITIONAL)

ifndef TRAVIS_BUILD_NUMBER
	# Travis succeeds/fails much faster. If it is a timeout error, no use waiting a long time on travis
	TIMEOUT = 100s
else
	# Locally ... it can take almost this much time to complete
	TIMEOUT = 500s
endif

all: build

version:
	go get go.pedge.io/proto/version
	@echo 'package main; import "github.com/pachyderm/pachyderm/src/client/version"; func main() { println(version.PrettyPrintVersion(version.Version)) }' > /tmp/pachyderm_version.go
	go run /tmp/pachyderm_version.go

deps:
	GO15VENDOREXPERIMENT=0 go get -d -v ./src/... ./.

update-deps:
	GO15VENDOREXPERIMENT=0 go get -d -v -u -f ./src/... ./.

test-deps:
	GO15VENDOREXPERIMENT=0 go get -d -v -t ./src/... ./.

update-test-deps:
	GO15VENDOREXPERIMENT=0 go get -d -v -t -u -f ./src/... ./.

build-clean-vendored-client:
	rm -rf src/server/vendor/github.com/pachyderm/pachyderm/src/client

build:
	GO15VENDOREXPERIMENT=1 go build $$(go list ./src/client/... | grep -v '/src/client$$')
	GO15VENDOREXPERIMENT=1 go build $$(go list ./src/server/... | grep -v '/src/server/vendor/' | grep -v '/src/server$$')

install:
	# GOPATH/bin must be on your PATH to access these binaries:
	GO15VENDOREXPERIMENT=1 go install -ldflags "$(LD_FLAGS)" ./src/server/cmd/pachctl ./src/server/cmd/pach-deploy

install-doc:
	GO15VENDOREXPERIMENT=1 go install ./src/server/cmd/pachctl-doc

# Run via 'make VERSION_ADDITIONAL=RC release' to specify a version string
release: release-version release-pachd release-job-shim release-manifest release-pachctl
	./etc/build/tag_release
	rm VERSION

release-version:
	@# Need to blow away pachctl binary if its already there
	@rm $(GOPATH)/bin/pachctl
	@make install
	@./etc/build/release_version

release-pachd:
	@VERSION="$(shell cat VERSION)" ./etc/build/release_pachd

release-job-shim:
	@VERSION="$(shell cat VERSION)" ./etc/build/release_job_shim

release-manifest:
	@VERSION="$(shell cat VERSION)" ./etc/build/release_manifest

release-pachctl:
	@VERSION="$(shell cat VERSION)" ./etc/build/release_pachctl

docker-build-compile:
	# Running locally, not on travis
	if [ -z $$TRAVIS_BUILD_NUMBER ]; then \
		sed 's/%%PACH_BUILD_NUMBER%%/000/' Dockerfile.pachd_template > Dockerfile.pachd; \
	else \
		sed 's/%%PACH_BUILD_NUMBER%%/${TRAVIS_BUILD_NUMBER}/' Dockerfile.pachd_template > Dockerfile.pachd; \
	fi
	docker build -t pachyderm_compile .

docker-build-job-shim: docker-build-compile
	docker run $(COMPILE_RUN_ARGS) pachyderm_compile sh etc/compile/compile.sh job-shim

docker-build-pachd: docker-build-compile
	docker run $(COMPILE_RUN_ARGS) pachyderm_compile sh etc/compile/compile.sh pachd

docker-build: docker-build-job-shim docker-build-pachd docker-build-fruitstand

docker-build-proto:
	docker build -t pachyderm_proto etc/proto

docker-build-fruitstand:
	docker build -t fruit_stand examples/fruit_stand

docker-push-job-shim: docker-build-job-shim
	docker push pachyderm/job-shim

docker-push-pachd: docker-build-pachd
	docker push pachyderm/pachd

docker-push: docker-push-job-shim docker-push-pachd

check-kubectl:
	# check that kubectl is installed
	which kubectl

launch-kube: check-kubectl
	etc/kube/start-kube-docker.sh

clean-launch-kube:
	docker kill $$(docker ps -q)

launch: check-kubectl install
	$(eval STARTTIME := $(shell date +%s))
	kubectl $(KUBECTLFLAGS) create -f $(MANIFEST)
	# wait for the pachyderm to come up
	until timeout 1s ./etc/kube/check_pachd_ready.sh; do sleep 1; done
	@echo "pachd launch took $$(($$(date +%s) - $(STARTTIME))) seconds"

launch-dev: check-kubectl install
	$(eval STARTTIME := $(shell date +%s))
	kubectl $(KUBECTLFLAGS) create -f $(DEV_MANIFEST)
	# wait for the pachyderm to come up
	until timeout 1s ./etc/kube/check_pachd_ready.sh; do sleep 1; done
	@echo "pachd launch took $$(($$(date +%s) - $(STARTTIME))) seconds"

clean-launch: check-kubectl
	kubectl $(KUBECTLFLAGS) delete --ignore-not-found -f $(MANIFEST)

clean-launch-dev: check-kubectl
	kubectl $(KUBECTLFLAGS) delete --ignore-not-found -f $(DEV_MANIFEST)

full-clean-launch: check-kubectl
	kubectl $(KUBECTLFLAGS) delete --ignore-not-found job -l suite=pachyderm
	kubectl $(KUBECTLFLAGS) delete --ignore-not-found all -l suite=pachyderm
	kubectl $(KUBECTLFLAGS) delete --ignore-not-found serviceaccount -l suite=pachyderm
	kubectl $(KUBECTLFLAGS) delete --ignore-not-found secret -l suite=pachyderm

clean-pps-storage: check-kubectl
	kubectl $(KUBECTLFLAGS) delete pvc rethink-volume-claim
	kubectl $(KUBECTLFLAGS) delete pv rethink-volume

integration-tests:
	CGOENABLED=0 go test -v ./src/server -timeout $(TIMEOUT)

proto: docker-build-proto
	find src -regex ".*\.proto" \
	| grep -v vendor \
	| xargs tar cf - \
	| docker run -i pachyderm_proto \
	| tar xf -

protofix:
	go install github.com/pachyderm/pachyderm/src/server/cmd/protofix
	protofix fix src
	git checkout src/server/vendor
	sudo chown -R `whoami` src/

pretest:
	go get -v github.com/kisielk/errcheck
	go get -v github.com/golang/lint/golint
	rm -rf src/server/vendor
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
	git checkout src/server/vendor
	#errcheck $$(go list ./src/... | grep -v src/cmd/ppsd | grep -v src/pfs$$ | grep -v src/pps$$)

test: pretest test-client test-fuse test-local docker-build clean-launch-dev launch-dev integration-tests

test-client:
	rm -rf src/client/vendor
	rm -rf src/server/vendor/github.com/pachyderm
	cp -R src/server/vendor src/client/
	GO15VENDOREXPERIMENT=1 go test -cover $$(go list ./src/client/... | grep -v vendor)
	rm -rf src/client/vendor
	git checkout src/server/vendor/github.com/pachyderm

test-fuse:
	CGOENABLED=0 GO15VENDOREXPERIMENT=1 go test -cover $$(go list ./src/server/... | grep -v '/src/server/vendor/' | grep '/src/server/pfs/fuse')

test-local:
	CGOENABLED=0 GO15VENDOREXPERIMENT=1 go test -cover -short $$(go list ./src/server/... | grep -v '/src/server/vendor/' | grep -v '/src/server/pfs/fuse')

clean: clean-launch clean-launch-kube

doc: install-doc
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

logs: check-kubectl
	kubectl get pod -l app=pachd | sed '1d' | cut -f1 -d ' ' | xargs -n 1 -I pod sh -c 'echo pod && kubectl logs pod'

kubectl:
	gcloud config set container/cluster $(CLUSTER_NAME)
	gcloud container clusters get-credentials $(CLUSTER_NAME)

google-cluster-manifest:
	@pach-deploy google $(BUCKET_NAME) $(STORAGE_NAME) $(STORAGE_SIZE)

google-cluster:
	gcloud container clusters create $(CLUSTER_NAME) --scopes storage-rw
	gcloud config set container/cluster $(CLUSTER_NAME)
	gcloud container clusters get-credentials $(CLUSTER_NAME)
	gcloud components update kubectl
	gcloud compute firewall-rules create pachd --allow=tcp:30650
	gsutil mb gs://$(BUCKET_NAME) # for PFS
	gcloud compute disks create --size=$(STORAGE_SIZE)GB $(STORAGE_NAME) # for PPS

clean-google-cluster:
	gcloud container clusters delete $(CLUSTER_NAME)
	gsutil -m rm -r gs://$(BUCKET_NAME)
	gcloud compute disks delete $(STORAGE_NAME)

amazon-cluster-manifest:
	@pach-deploy amazon $(BUCKET_NAME) $(AWS_ID) $(AWS_KEY) $(AWS_TOKEN) $(AWS_REGION) $(STORAGE_NAME) $(STORAGE_SIZE)

amazon-cluster:
	aws s3api create-bucket --bucket $(BUCKET_NAME) --region $(AWS_REGION)
	aws ec2 create-volume --size $(STORAGE_SIZE) --region $(AWS_REGION) --availability-zone $(AWS_AVAILABILITY_ZONE) --volume-type gp2

clean-amazon-cluster:
	aws s3api delete-bucket --bucket $(BUCKET_NAME) --region $(AWS_REGION)
	aws ec2 delete-volume --volume-id $(STORAGE_NAME)

install-go-bindata:
	go get -u github.com/jteeuwen/go-bindata/...

assets: install-go-bindata
	go-bindata -o assets.go -pkg pachyderm doc/

lint:
	@for pkg in $$(go list ./src/... | grep -v '/vendor/' ) ; do \
		if [ "`golint $$pkg | tee /dev/stderr`" ] ; then \
			echo "golint errors!" && echo && exit 1; \
		fi \
	done

goxc-generate-local:
	@if [ -z $$GITHUB_OAUTH_TOKEN ]; then \
		echo "Missing token. Please run via: 'make GITHUB_OAUTH_TOKEN=12345 goxc-generate-local'"; \
		exit 1; \
	fi
	goxc -wlc default publish-github -apikey=$(GITHUB_OAUTH_TOKEN)

goxc-release:
	@if [ -z $$VERSION ]; then \
		echo "Missing version. Please run via: 'make VERSION=v1.2.3-4567 goxc-release'"; \
		exit 1; \
	fi
	goxc -pv="$(VERSION)" -wd=./src/server/cmd/pachctl

goxc-build:
	goxc -tasks=xc -wd=./src/server/cmd/pachctl

.PHONY:
	all \
	version \
	deps \
	deps-client \
	update-deps \
	test-deps \
	update-test-deps \
	build-clean-vendored-client \
	build \
	install \
	install-doc \
	homebrew \
	tag-release \
	release \
	release-job-shim \
	release-manifest \
	release-pachd \
	release-version \
	docker-build-compile \
	docker-build-job-shim \
	docker-build-pachd \
	docker-build \
	docker-build-proto \
	docker-build-fruitstand \
	docker-push-job-shim \
	docker-push-pachd \
	docker-push \
	launch-kube \
	clean-launch-kube \
	kube-cluster-assets \
	launch \
	launch-dev \
	clean-launch \
	full-clean-launch \
	clean-pps-storage \
	integration-tests \
	proto \
	protofix \
	pretest \
	test \
	test-client \
	test-fuse \
	test-local \
	clean \
	doc \
	grep-data \
	grep-example \
	logs \
	kubectl \
	google-cluster-manifest \
	google-cluster \
	clean-google-cluster \
	amazon-cluster-manifest \
	amazon-cluster \
	clean-amazon-cluster \
	install-go-bindata \
	assets \
	lint \
	goxc-generate-local \
	goxc-release \
	goxc-build
