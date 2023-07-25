#!/bin/bash
# This script reads a tar of protobuf files from stdin and prints the
# generated python files as a tarball to stdout.
#
# Note: If you are trying to develop on/debug this script, you should
# print to stderr as this will not corrupt the output.
set -e

# will extract from stdin and move to ./api
tar xf /dev/stdin

OUTDIR="api"
mkdir -p ${OUTDIR}
mv src/* ${OUTDIR}

# Rearrange/rename some files.
mkdir ${OUTDIR}/taskapi
mv ${OUTDIR}/task/task.proto ${OUTDIR}/taskapi
mv ${OUTDIR}/version/versionpb/version.proto ${OUTDIR}/version

PROTO_FILES=$(find ${OUTDIR} -name "*.proto")

# Remove protobuf extensions, that are Go specific.
for i in ${PROTO_FILES}; do
    # remove the protoextensions/log.proto
    sed -i 's/import.*protoextensions\/log.proto.*\;//' "${i}"
    sed -i 's/\[.*log.*\]//' "${i}"
done

# fix imports to be relative to $OUTDIR
for i in ${PROTO_FILES}; do
    perl -pi -e "s/import \"((?!google).*)\"/import \"api\/\$1\"/" "${i}"
done

# Fix naming things for cleaner output.
for i in ${PROTO_FILES}; do
    sed -i 's/_v2//' "${i}"
    sed -i 's/task\/task.proto/taskapi\/task.proto/' "${i}"
    sed -i 's/version\/versionpb/version/' "${i}"
    sed -i 's/versionpb/version/' "${i}"
    sed -i 's/pachyderm.worker/worker/' "${i}"
done

# Generate python files.
echo "${PROTO_FILES}" | xargs python3 -m grpc_tools.protoc -I. --python_betterproto_out=${OUTDIR}

# Fix routing addresses.
V2_APIS="admin auth enterprise identity license pfs pps transaction"
for name in ${V2_APIS}; do
  sed -i "s/${name}.API/${name}_v2.API/" ${OUTDIR}/"${name}"/__init__.py
done
sed -i "s/debug.Debug/debug_v2.Debug/" ${OUTDIR}/debug/__init__.py
sed -i "s/version.API/versionpb_v2.API/" ${OUTDIR}/version/__init__.py
sed -i "s/worker.Worker/pachyderm.worker.Worker/" ${OUTDIR}/worker/__init__.py

# Clean up
find ${OUTDIR} -empty -type d -delete

tar cf - ${OUTDIR}
