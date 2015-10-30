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
	rm -rf /tmp/google-protobuf
	git clone https://github.com/google/protobuf.git /tmp/google-protobuf
	rm -rf google/protobuf
	mkdir -p google/protobuf
	for file in $$(ls /tmp/google-protobuf/src/google/protobuf/*\.proto | grep -v test | grep -v descriptor\.proto); do \
		cp $$file google/protobuf/; \
	done
	protoc --proto_path=. --go_out=. google/protobuf/*\.proto
	mv google/protobuf/* .
	rm -rf google
