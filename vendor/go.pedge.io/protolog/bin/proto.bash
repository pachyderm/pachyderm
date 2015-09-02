#!/bin/bash

PROTO_FILE=${1}
PB_GO_FILE=$(echo ${PROTO_FILE} | sed "s/\.proto/\.pb.go/")
PB_LOG_GO_FILE=$(echo ${PROTO_FILE} | sed "s/\.proto/\.pb.log.go/")

protoc -I "$(dirname $(which protoc))/../include" -I $(dirname ${PROTO_FILE}) --go_out=Mgoogle/protobuf/timestamp.proto=github.com/peter-edge/go-google-protobuf:$(dirname ${PROTO_FILE}) --protolog_out=$(dirname ${PROTO_FILE}) ${PROTO_FILE}

rm -f ${PB_GO_FILE}.tmp
head -4 ${PB_GO_FILE} > ${PB_GO_FILE}.tmp
tail -n +$(grep -n 'package ' ${PB_GO_FILE} | cut -f 1 -d :) ${PB_GO_FILE} >> ${PB_GO_FILE}.tmp
mv ${PB_GO_FILE}.tmp ${PB_GO_FILE}

if [ "$(dirname ${PROTO_FILE})" == "." ]; then
  rm -f ${PB_LOG_GO_FILE}.tmp
  cat ${PB_LOG_GO_FILE} | grep -v import | sed "s/protolog.Register/Register/" | sed "s/protolog.Message/Message/" > ${PB_LOG_GO_FILE}.tmp
  mv ${PB_LOG_GO_FILE}.tmp ${PB_LOG_GO_FILE}
fi
