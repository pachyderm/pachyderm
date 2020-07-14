#!/bin/sh
protoc \
    -I${GOPATH}/src/github.com/gogo/protobuf \
    -Isrc \
    --gofast_out=plugins=grpc,\
Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types,\
Mgoogle/protobuf/empty.proto=github.com/gogo/protobuf/types,\
Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types,\
Mgoogle/protobuf/wrappers.proto=github.com/gogo/protobuf/types,\
Mgogoproto/gogo.proto=github.com/gogo/protobuf/gogoproto,\
Mgoogle/protobuf/any.proto=github.com/gogo/protobuf/types:. \
--js_out=import_style=commonjs:. \
--grpc-web_out=import_style=commonjs,mode=grpcwebtext:. \
src/client/pfs/pfs.proto
