FROM golang:1.17.3

LABEL maintainer="msteffen@pachyderm.io"

ARG PROTO_COMPILER_VERSION=3.11.4
ARG TARGETPLATFORM

RUN apt-get update -yq && apt-get install -yq unzip

# Install protoc
RUN if [ "${TARGETPLATFORM}" = "linux/amd64" ]; then wget "https://github.com/protocolbuffers/protobuf/releases/download/v${PROTO_COMPILER_VERSION}/protoc-${PROTO_COMPILER_VERSION}-linux-x86_64.zip" -O protoc.zip; fi
RUN if [ "${TARGETPLATFORM}" = "linux/arm64" ]; then wget "https://github.com/protocolbuffers/protobuf/releases/download/v${PROTO_COMPILER_VERSION}/protoc-${PROTO_COMPILER_VERSION}-linux-aarch_64.zip" -O protoc.zip; fi

RUN unzip protoc.zip -d /
RUN cp -r /include /bin

# if you modify the version of gogo/protobuf. you also need to update the path in run.sh
RUN go install -v github.com/gogo/protobuf/protoc-gen-gofast@v1.3.2 github.com/gogo/protobuf/protoc-gen-gogofast@v1.3.2
RUN mkdir -p ${GOPATH}/src/github.com/pachyderm/pachyderm
RUN date +%s >/last_run_time

# Build the bespoke Pachyderm codegen plugin
COPY pachgen ${GOPATH}/src/github.com/pachyderm/pachyderm/etc/proto/pachgen
WORKDIR ${GOPATH}/src/github.com/pachyderm/pachyderm/etc/proto/pachgen
RUN go mod init && go mod tidy && go build -o "${GOPATH}/bin/protoc-gen-pach"

ADD run.sh /
CMD ["/run.sh"]
