FROM golang:1.7.0
MAINTAINER derek@pachyderm.io

RUN apt-get update -yq
RUN apt-get install -yq unzip

RUN wget https://github.com/google/protobuf/releases/download/v3.0.0-beta-2/protoc-3.0.0-beta-2-linux-x86_64.zip -O protoc.zip
RUN unzip protoc.zip -d /bin

RUN go get -u -v github.com/golang/protobuf/proto
RUN go get -u -v github.com/golang/protobuf/protoc-gen-go
RUN go get -u github.com/gengo/grpc-gateway/protoc-gen-grpc-gateway
RUN go get -v --insecure go.pedge.io/protoeasy/cmd/protoeasy
RUN go get -v github.com/pachyderm/pachyderm/src/server/cmd/protofix

ADD run /
ENTRYPOINT ["/run"]
WORKDIR /work
