.PHONY: \
	all \
	deps \
	update-deps \
	build \
	clean \
	proto

all: proto build

deps:
	go get -d -v ./...

update-deps:
	go get -d -v -u -f ./...

build: deps
	go build ./...

clean:
	go clean ./...

proto:
	docker pull pedge/proto3grpc
	docker run --volume $$(pwd):/compile pedge/proto3grpc bash /compile/compile.bash /usr/include /compile
	for file in *.pb.go; do \
		sudo chown $$(whoami) $$file; \
		sudo chgrp $$(whoami) $$file; \
	done
