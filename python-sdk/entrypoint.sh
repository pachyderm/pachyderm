#!/bin/bash
set -e

# will extract from stdin and to ./api
tar xf /dev/stdin

# VERSION matches the MajorVersion of pachyderm
OUTDIR="api"

mkdir -p ${OUTDIR}
mv src/* ${OUTDIR}

# Make sure to remove the gogoproto line, as that is Go specific
for i in $(find ${OUTDIR} -name "*.proto"); do
    # remove the import
    sed -i 's/import.*gogo.proto.*\;//' ${i}
    # remove usages of gogoproto types
    sed -i 's/\[.*gogoproto.*\]//' ${i}
    sed -i 's/.*gogoproto.*//' ${i}
done

# fix imports to be relative to OUTDIR
for i in $(find ${OUTDIR} -name "*.proto"); do
    perl -pi -e "s/import \"((?!google).*)\"/import \"api\/\$1\"/" ${i}
done

# Refactor IDP -> Idp, OIDC -> Oidc (for BetterProto)
sed -i 's/IDP/Idp/g' ${OUTDIR}/identity/identity.proto
sed -i 's/OIDC/Oidc/g' ${OUTDIR}/identity/identity.proto
find ${OUTDIR} -name '*.proto' | xargs poetry run python3 -m grpc_tools.protoc -I. --python_betterproto_out=${OUTDIR} --python_betterproto_opt="grpc=grpcio"

# Clean up
find ${OUTDIR} -name '*.proto' | xargs rm
find ${OUTDIR} -empty -type d -delete

tar cf - ${OUTDIR}
