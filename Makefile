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

COMPILE_RUN_ARGS = -d -v /var/run/docker.sock:/var/run/docker.sock --privileged=true
VERSION_ADDITIONAL = $(shell git log --pretty=format:%H | head -n 1)
LD_FLAGS = -X github.com/pachyderm/pachyderm/src/server/vendor/github.com/pachyderm/pachyderm/src/client/version.AdditionalVersion=$(VERSION_ADDITIONAL)

CLUSTER_NAME?=pachyderm
CLUSTER_MACHINE_TYPE?=n1-standard-4
CLUSTER_SIZE?=4

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
	GO15VENDOREXPERIMENT=1 go install -ldflags "$(LD_FLAGS)" ./src/server/cmd/pachctl

install-doc:
	GO15VENDOREXPERIMENT=1 go install ./src/server/cmd/pachctl-doc

point-release:
	@make VERSION_ADDITIONAL= release

# Run via 'make VERSION_ADDITIONAL=RC release' to specify a version string
release: release-version release-pachd release-job-shim release-pachctl doc
	@rm VERSION
	@echo "Release completed"

release-version:
	@# Need to blow away pachctl binary if its already there
	@rm $(GOPATH)/bin/pachctl || true
	@make install
	@./etc/build/release_version

release-pachd:
	@VERSION="$(shell cat VERSION)" ./etc/build/release_pachd

release-job-shim:
	@VERSION="$(shell cat VERSION)" ./etc/build/release_job_shim

release-pachctl:
	@VERSION="$(shell cat VERSION)" ./etc/build/release_pachctl

docker-build-compile:
	docker build -t pachyderm_compile .

docker-clean-job-shim:
	docker stop job_shim_compile || true
	docker rm job_shim_compile || true

docker-build-job-shim: docker-clean-job-shim docker-build-compile
	docker run --name job_shim_compile $(COMPILE_RUN_ARGS) pachyderm_compile sh etc/compile/compile.sh job-shim "$(LD_FLAGS)"

docker-wait-job-shim:
	etc/compile/wait.sh job_shim_compile

docker-clean-pachd:
	docker stop pachd_compile || true
	docker rm pachd_compile || true

docker-build-pachd: docker-clean-pachd docker-build-compile
	docker run --name pachd_compile $(COMPILE_RUN_ARGS) pachyderm_compile sh etc/compile/compile.sh pachd "$(LD_FLAGS)"

docker-wait-pachd:
	etc/compile/wait.sh pachd_compile

docker-build: docker-build-job-shim docker-build-pachd docker-wait-job-shim docker-wait-pachd

docker-build-proto:
	docker build -t pachyderm_proto etc/proto

check-kubectl:
	# check that kubectl is installed
	which kubectl

check-kubectl-connection:
	kubectl $(KUBECTLFLAGS) get all > /dev/null

launch-kube: check-kubectl
	etc/kube/start-kube-docker.sh

clean-launch-kube:
	docker kill $$(docker ps -q)

launch: install check-kubectl
	$(eval STARTTIME := $(shell date +%s))
	pachctl deploy --dry-run | kubectl $(KUBECTLFLAGS) create -f -
	# wait for the pachyderm to come up
	until timeout 1s ./etc/kube/check_pachd_ready.sh; do sleep 1; done
	@echo "pachd launch took $$(($$(date +%s) - $(STARTTIME))) seconds"

launch-dev: check-kubectl check-kubectl-connection install
	$(eval STARTTIME := $(shell date +%s))
	pachctl deploy -d --dry-run | kubectl $(KUBECTLFLAGS) create -f -
	# wait for the pachyderm to come up
	until timeout 1s ./etc/kube/check_pachd_ready.sh; do sleep 1; done
	@echo "pachd launch took $$(($$(date +%s) - $(STARTTIME))) seconds"	

clean-launch: check-kubectl
	pachctl deploy --dry-run | kubectl $(KUBECTLFLAGS) delete --ignore-not-found -f -

clean-launch-dev: check-kubectl
	pachctl deploy -d --dry-run | kubectl $(KUBECTLFLAGS) delete --ignore-not-found -f -

full-clean-launch: check-kubectl
	kubectl $(KUBECTLFLAGS) delete --ignore-not-found job -l suite=pachyderm
	kubectl $(KUBECTLFLAGS) delete --ignore-not-found all -l suite=pachyderm
	kubectl $(KUBECTLFLAGS) delete --ignore-not-found serviceaccount -l suite=pachyderm
	kubectl $(KUBECTLFLAGS) delete --ignore-not-found secret -l suite=pachyderm

launch-test-rethinkdb:
	@# Expose port 8081 so you can connect to the rethink dashboard
	@# (You may need to forward port 8081 if you're running docker machine)
	docker run --name pachyderm-test-rethinkdb -d -p 28015:28015 -p 8081:8080 rethinkdb:2.3.3 
	sleep 20  # wait for rethinkdb to start up

clean-launch-test-rethinkdb:
	docker stop pachyderm-test-rethinkdb || true
	docker rm pachyderm-test-rethinkdb || true

clean-pps-storage: check-kubectl
	kubectl $(KUBECTLFLAGS) delete pvc rethink-volume-claim
	kubectl $(KUBECTLFLAGS) delete pv rethink-volume

integration-tests:
	CGOENABLED=0 go test -v ./src/server $(TESTFLAGS) -timeout $(TIMEOUT)

example-tests:
	CGOENABLED=0 go test -v ./src/server/examples $(TESTFLAGS) -timeout $(TIMEOUT)

proto: docker-build-proto
	find src -regex ".*\.proto" \
	| grep -v vendor \
	| xargs tar cf - \
	| docker run -i pachyderm_proto \
	| tar xf -

# Use this to grab a binary for profiling purposes
pachd-profiling-binary: docker-clean-pachd docker-build-compile
	docker run -i  pachyderm_compile sh etc/compile/compile.sh pachd "$(LD_FLAGS)" PROFILE \
	| tar xf -
	# Binary emitted to ./pachd

protofix:
	go install github.com/pachyderm/pachyderm/src/server/cmd/protofix
	protofix fix src
	git checkout src/server/vendor
	sudo chown -R `whoami` src/

pretest:
	go get -v github.com/kisielk/errcheck
	rm -rf src/server/vendor
	go vet -n ./src/... | while read line; do \
		modified=$$(echo $$line | sed "s/ [a-z0-9_/]*\.pb\.gw\.go//g"); \
		$$modified; \
		if [ -n "$$($$modified)" ]; then \
		exit 1; \
		fi; \
		done
	git checkout src/server/vendor
	#errcheck $$(go list ./src/... | grep -v src/cmd/ppsd | grep -v src/pfs$$ | grep -v src/pps$$)

test: pretest test-client clean-launch-test-rethinkdb launch-test-rethinkdb test-fuse test-local docker-build clean-launch-dev launch-dev integration-tests example-tests

bench:
	go test ./src/server -run=XXX -bench=.

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
	kubectl $(KUBECTLFLAGS) get pod -l app=pachd | sed '1d' | cut -f1 -d ' ' | xargs -n 1 -I pod sh -c 'echo pod && kubectl $(KUBECTLFLAGS) logs pod'

kubectl:
	gcloud config set container/cluster $(CLUSTER_NAME)
	gcloud container clusters get-credentials $(CLUSTER_NAME)

google-cluster-manifest:
	@pachctl deploy --dry-run google $(BUCKET_NAME) $(STORAGE_NAME) $(STORAGE_SIZE)

google-cluster:
	gcloud container clusters create $(CLUSTER_NAME) --scopes storage-rw --machine-type $(CLUSTER_MACHINE_TYPE) --num-nodes $(CLUSTER_SIZE)
	gcloud config set container/cluster $(CLUSTER_NAME)
	gcloud container clusters get-credentials $(CLUSTER_NAME)
	gcloud components install kubectl
	-gcloud compute firewall-rules create pachd --allow=tcp:30650
	gsutil mb gs://$(BUCKET_NAME) # for PFS
	gcloud compute disks create --size=$(STORAGE_SIZE)GB $(STORAGE_NAME) # for PPS

clean-google-cluster:
	gcloud container clusters delete $(CLUSTER_NAME)
	gcloud compute firewall-rules delete pachd
	gsutil -m rm -r gs://$(BUCKET_NAME)
	gcloud compute disks delete $(STORAGE_NAME)

amazon-cluster-manifest: install
	@pachctl deploy --dry-run amazon $(BUCKET_NAME) $(AWS_ID) $(AWS_KEY) $(AWS_TOKEN) $(AWS_REGION) $(STORAGE_NAME) $(STORAGE_SIZE)

amazon-cluster:
	aws s3api create-bucket --bucket $(BUCKET_NAME) --region $(AWS_REGION)
	aws ec2 create-volume --size $(STORAGE_SIZE) --region $(AWS_REGION) --availability-zone $(AWS_AVAILABILITY_ZONE) --volume-type gp2

amazon-clean-cluster:
	aws s3api delete-bucket --bucket $(BUCKET_NAME) --region $(AWS_REGION)
	aws ec2 detach-volume --force --volume-id $(STORAGE_NAME)
	sleep 20
	aws ec2 delete-volume --volume-id $(STORAGE_NAME)

amazon-clean-launch: clean-launch
	kubectl $(KUBECTLFLAGS) delete --ignore-not-found secrets amazon-secret
	kubectl $(KUBECTLFLAGS) delete --ignore-not-found persistentvolumes rethink-volume
	kubectl $(KUBECTLFLAGS) delete --ignore-not-found persistentvolumeclaims rethink-volume-claim

amazon-clean: 
	@while :; \
        do if echo "The following script will delete your AWS bucket and volume. The action cannot be undone. Do you want to proceed? (Y/n)";read REPLY; then \
        case $$REPLY in Y|y) make amazon-clean-launch;make amazon-clean-cluster;break;; \
	N|n) echo "The amazon clean process has been cancelled by user!";break;; \ 
	*) echo "input parameter error, please input again ";continue;;esac; \
        fi;done;

install-go-bindata:
	go get -u github.com/jteeuwen/go-bindata/...

assets: install-go-bindata
	go-bindata -o assets.go -pkg pachyderm doc/...

lint:
	@go get -u github.com/golang/lint/golint
	@for file in $$(find "./src" -name '*.go' | grep -v '/vendor/' | grep -v '\.pb\.go'); do \
		golint $$file; \
		if [ -n "$$(golint $$file)" ]; then \
			echo "golint errors!" && echo && exit 1; \
		fi; \
	done;

goxc-generate-local:
	@if [ -z $$GITHUB_OAUTH_TOKEN ]; then \
		echo "Missing token. Please run via: 'make GITHUB_OAUTH_TOKEN=12345 goxc-generate-local'"; \
		exit 1; \
	fi
	goxc -wlc default publish-github -apikey=$(GITHUB_OAUTH_TOKEN)

goxc-release:
	@if [ -z $$VERSION ]; then \
		@echo "Missing version. Please run via: 'make VERSION=v1.2.3-4567 VERSION_ADDITIONAL=4567 goxc-release'"; \
		@exit 1; \
	fi
	sed 's/%%VERSION_ADDITIONAL%%/$(VERSION_ADDITIONAL)/' .goxc.json.template > .goxc.json
	goxc -pv="$(VERSION)" -wd=./src/server/cmd/pachctl

goxc-build:
	sed 's/%%VERSION_ADDITIONAL%%/$(VERSION_ADDITIONAL)/' .goxc.json.template > .goxc.json
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
	amazon-clean-cluster \
	amazon-clean-launch \
	amazon-clean \
	install-go-bindata \
	assets \
	lint \
	goxc-generate-local \
	goxc-release \
	goxc-build \
	launch-test-rethinkdb \
	clean-launch-test-rethinkdb
