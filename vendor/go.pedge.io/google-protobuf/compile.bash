#!/bin/bash

set -e

if [ "$#" -ne 2 ]; then
  echo "usage: ${0} include_dir output_dir" >&2
  exit 1
fi

INCLUDE_DIR="${1}"
OUTPUT_DIR="${2}"

PROTO_DIR="${INCLUDE_DIR}/google/protobuf"
PROTO_FILES="$(ls "${PROTO_DIR}" | grep "\.proto$")"
NUM_PROTO_FILES="$(echo "${PROTO_FILES}" | wc -w)"

for proto_file in ${PROTO_FILES}; do
  output_file="${OUTPUT_DIR}/$(echo "${proto_file}" | sed "s/\.proto$/.pb.go/")"
  proto_file="${PROTO_DIR}/${proto_file}"
  tmp_file="/tmp/tmp.$$"
  protoc "-I${PROTO_DIR}" "--go_out=${OUTPUT_DIR}" "${proto_file}"
  cat "${output_file}" | grep -v "import google_protobuf" > "${tmp_file}"
  mv "${tmp_file}" "${output_file}"
  i=1
  while [ "${i}" -lt "${NUM_PROTO_FILES}" ]; do
    if grep "google_protobuf${i}" "${output_file}" > /dev/null; then
      sed -i "s/google_protobuf${i}\.//g" ${output_file}
    fi
    i="$((i+1))"
  done
done
