FROM golang:1.14.2
LABEL maintainer="msteffen@pachyderm.io"

ENV GOPROXY https://proxy.golang.org

RUN apt-get update -yq
RUN apt-get install -yq unzip

RUN wget https://github.com/google/protobuf/releases/download/v3.7.1/protoc-3.7.1-linux-x86_64.zip -O protoc.zip
RUN unzip protoc.zip -d /
RUN cp -r /include /bin

RUN go get -u -v github.com/golang/protobuf/proto
RUN go get -u -v github.com/gogo/protobuf/protoc-gen-gofast
RUN mkdir -p ${GOPATH}/src/github.com/pachyderm/pachyderm
RUN date +%s >/last_run_time

ADD run.sh /
CMD ["/run.sh"]
