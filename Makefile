.PHONY: \
	all \
	deps \
	update-deps \
	test-deps \
	update-test-deps \
	build \
	install \
	clean \
	lint \
	vet \
	errcheck \
	pretest \
	test \
	shell \
	shell-ppsd \
	launch-pfsd \
	launch-ppsd \
	build-images \
	push-images \
	proto \
	hit-godoc

IMAGES = pfsd ppsd
BINARIES = pfs pfsd pps ppsd

ifdef NOCACHE
NOCACHE_CMD = touch etc/deps/deps.list
endif

all: build

deps:
	go get -d -v ./...

update-deps:
	go get -d -v -u -f ./...
	./bin/deps

test-deps:
	go get -d -v -t ./...

update-test-deps:
	go get -d -v -t -u -f ./...
	./bin/deps

build: deps
	go build ./...

install: deps
	go install ./src/cmd/pfs ./src/cmd/pps

clean:
	go clean ./...
	./bin/clean
	$(foreach image,$(IMAGES),PACHYDERM_IMAGE=$(image) ./bin/clean || exit;)
	$(foreach binary,$(BINARIES),rm -f src/cmd/$(binary)/$(binary);)
	sudo rm -rf _tmp

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

test: pretest
	./bin/run go test $(TESTFLAGS) ./...

build-images:
	$(NOCACHE_CMD)
	$(foreach image,$(IMAGES),PACHYDERM_IMAGE=$(image) ./bin/build || exit;)

build-%:
	$(NOCACHE_CMD)
	PACHYDERM_IMAGE=$* ./bin/build

push-images: build-images
	$(foreach image,$(IMAGES),docker push pachyderm/$(image) || exit;)

shell:
	PACHYDERM_IMAGE=shell ./bin/run

shell-ppsd:
	PACHYDERM_IMAGE=shell-ppsd ./bin/run

launch-pfsd:
	PACHYDERM_IMAGE=pfsd PACHYDERM_DOCKER_OPTS="-d" ./bin/run

launch-ppsd:
	PACHYDERM_IMAGE=ppsd PACHYDERM_DOCKER_OPTS="-d" ./bin/run

proto:
	bin/proto

hit-godoc:
	for pkg in $$(find . -name '*.go' | xargs dirname | sort | uniq); do \
		curl https://godoc.org/github.com/pachyderm/pachyderm/$pkg > /dev/null; \
	done
