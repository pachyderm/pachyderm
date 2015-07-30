#### VARIABLES
# RUNARGS: arguments for run
# DOCKER_OPTS: docker-compose options for run, test, launch-*
# TESTPKGS: packages for test, default ./...
# TESTFLAGS: flags for test
# NOCACHE: do not use dependency cache for docker
####

TESTPKGS = ./...

ifdef NOCACHE
NOCACHE_CMD = touch etc/deps/deps.list
endif

all: build

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

build: deps
	go build ./...

install: deps
	go install ./src/cmd/pfs ./src/cmd/pps

docker-nocache:
	$(NOCACHE_CMD)

docker-build: docker-nocache
	docker-compose build pachyderm
	docker tag -f pachyderm_pachyderm:latest pachyderm/pachyderm:latest

docker-build-pfsd: docker-build
	docker-compose run pachyderm go build -a -installsuffix netgo -tags netgo -o /compile/pfsd src/cmd/pfsd/main.go
	docker-compose build pfsd
	docker tag -f pachyderm_pfsd:latest pachyderm/pfsd:latest

docker-build-ppsd: docker-build
	docker-compose run pachyderm go build -a -installsuffix netgo -tags netgo -o /compile/ppsd src/cmd/ppsd/main.go
	docker-compose build ppsd
	docker tag -f pachyderm_ppsd:latest pachyderm/ppsd:latest

docker-push-pfsd: docker-build-pfsd
	docker push pachyderm/pfsd

docker-push-ppsd: docker-build-ppsd
	docker push pachyderm/ppsd

docker-push: docker-push-ppsd docker-push-pfsd

btrfs-clean:
	-sudo umount /tmp/pfs/btrfs > /dev/null
	sudo rm -rf /tmp/pfs > /dev/null

btrfs-setup: btrfs-clean
	sudo mkdir -p /tmp/pfs/btrfs
	sudo truncate /tmp/pfs/btrfs.img -s 10G
	sudo mkfs.btrfs /tmp/pfs/btrfs.img
	sudo mount /tmp/pfs/btrfs.img /tmp/pfs/btrfs
	sudo mkdir -p /tmp/pfs/btrfs/global

run: btrfs-setup docker-build
	docker-compose run $(DOCKER_OPTS) pachyderm $(RUNARGS)

launch-pfsd: btrfs-setup docker-build-pfsd
	docker-compose run --service-ports -d $(DOCKER_OPTS) pfsd

launch-ppsd: btrfs-setup docker-build-ppsd
	docker-compose run --service-ports -d $(DOCKER_OPTS) ppsd

proto:
	go get -v github.com/peter-edge/go-tools/docker-protoc-all
	docker pull pedge/protolog
	DEBUG=1 REL_PROTOC_INCLUDE_PATH=src docker-protoc-all github.com/pachyderm/pachyderm

lint:
	go get -v github.com/golang/lint/golint
	for file in $$(find "./src" -name '*.go' | grep -v '\.pb\.go'); do \
		golint $$file | grep -v unexported || true; \
	done;

vet:
	go vet ./...

errcheck:
	go get -v github.com/kisielk/errcheck
	errcheck ./src/cmd ./src/common ./src/pfs ./src/pps

pretest: lint vet errcheck

pre: build pretest

test: pretest btrfs-setup docker-build
	docker-compose run $(DOCKER_OPTS) pachyderm go test $(TESTFLAGS) $(TESTPKGS)

clean: btrfs-clean
	go clean ./...
	rm -f src/cmd/pfsd/pfsd
	rm -f src/cmd/ppsd/ppsd
	sudo rm -rf _tmp

hit-godoc:
	for pkg in $$(find . -name '*.go' | xargs dirname | sort | uniq); do \
		curl https://godoc.org/github.com/pachyderm/pachyderm/$pkg > /dev/null; \
	done

.PHONY: \
	all \
	deps \
	update-deps \
	test-deps \
	update-test-deps \
	update-deps-list \
	build \
	install \
	docker-nocache \
	docker-build \
	docker-build-pfsd \
	docker-build-ppsd \
	docker-push-pfsd \
	docker-push-ppsd \
	docker-push \
	btrfs-clean \
	btrfs-setup \
	run \
	launch-pfsd \
	launch-ppsd \
	proto \
	lint \
	vet \
	errcheck \
	pretest \
	pre \
	test \
	clean \
	hit-godoc
