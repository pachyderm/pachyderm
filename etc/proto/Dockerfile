ARG GOVERSION=golang:1.20.5
FROM $GOVERSION

LABEL maintainer="msteffen@pachyderm.io"

ARG PROTO_COMPILER_VERSION=23.4
ARG GOPATCH_VERSION=v0.2.0
ARG TARGETPLATFORM

RUN apt-get update -yq && apt-get install -yq unzip

# Install protoc
RUN if [ "${TARGETPLATFORM}" = "linux/amd64" ]; then wget "https://github.com/protocolbuffers/protobuf/releases/download/v${PROTO_COMPILER_VERSION}/protoc-${PROTO_COMPILER_VERSION}-linux-x86_64.zip" -O protoc.zip; fi
RUN if [ "${TARGETPLATFORM}" = "linux/arm64" ]; then wget "https://github.com/protocolbuffers/protobuf/releases/download/v${PROTO_COMPILER_VERSION}/protoc-${PROTO_COMPILER_VERSION}-linux-aarch_64.zip" -O protoc.zip; fi
RUN unzip protoc.zip -d /
RUN cp -r /include /bin

# Instal protoc-gen-go
RUN go install -v google.golang.org/protobuf/cmd/protoc-gen-go@v1.30.0
RUN go install -v google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0
RUN mkdir -p ${GOPATH}/src/github.com/pachyderm/pachyderm
RUN date +%s >/last_run_time

# Install gopatch.
RUN go install github.com/uber-go/gopatch@${GOPATCH_VERSION}

# Build the bespoke Pachyderm codegen plugin
COPY pachgen ${GOPATH}/src/github.com/pachyderm/pachyderm/etc/proto/pachgen
WORKDIR ${GOPATH}/src/github.com/pachyderm/pachyderm/etc/proto/pachgen
RUN go mod init && go mod tidy && go build -o "${GOPATH}/bin/protoc-gen-pach"

# Build the bespoke MarshalLogObject generator.
COPY protoc-gen-zap ${GOPATH}/src/github.com/pachyderm/pachyderm/etc/proto/protoc-gen-zap
WORKDIR ${GOPATH}/src/github.com/pachyderm/pachyderm/etc/proto/protoc-gen-zap
RUN go build -mod=readonly -o "${GOPATH}/bin/protoc-gen-zap"

ADD run.sh /

# Copy over the proto modification patch.
COPY proto.patch /proto.patch

CMD ["/run.sh"]
