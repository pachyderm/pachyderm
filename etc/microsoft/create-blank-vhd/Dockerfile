FROM golang:1.7.1-alpine

ADD . /src
WORKDIR /src

RUN /src/docker-build.sh

ENTRYPOINT ["go", "run", "create_blank_vhd.go"]
