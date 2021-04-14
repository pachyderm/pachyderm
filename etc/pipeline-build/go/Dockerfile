FROM golang:latest

LABEL name="Pachyderm" \
      vendor="Pachyderm"

COPY LICENSE /licenses

WORKDIR /app
ADD etc/pipeline-build/go/build.sh .
ADD etc/pipeline-build/go/run.sh .
