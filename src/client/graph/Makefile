generate:
	protoc --gogo_out=Mgoogle/protobuf/descriptor.proto=github.com/gogo/protobuf/protoc-gen-gogo/descriptor:. --proto_path=$(GOPATH)/src:. *.proto
	protoc --gogo_out=Mgoogle/protobuf/descriptor.proto=github.com/gogo/protobuf/protoc-gen-gogo/descriptor:. --proto_path=$(GOPATH)/src:. types/*.proto

.PHONY:
	generate