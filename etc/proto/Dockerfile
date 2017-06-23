FROM golang:1.8.3
MAINTAINER derek@pachyderm.io

RUN apt-get update -yq
RUN apt-get install -yq unzip

RUN wget https://github.com/google/protobuf/releases/download/v3.3.0/protoc-3.3.0-linux-x86_64.zip -O protoc.zip
RUN unzip protoc.zip -d /
RUN cp -r /include /bin

RUN go get -u -v github.com/golang/protobuf/proto
RUN go get -u -v github.com/gogo/protobuf/protoc-gen-gofast
RUN go get -u github.com/gengo/grpc-gateway/protoc-gen-grpc-gateway

ADD run /
ENTRYPOINT ["/run"]
WORKDIR /work
