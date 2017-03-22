FROM golang:1.8.0-alpine

ADD . /src
WORKDIR /src

RUN /src/docker-build.sh

ENTRYPOINT ["go", "run", "create_blank_vhd.go"]
