# This tool downloads .proto files and generates the corresponding js/ts files
# Is is important to mirror Pachyderm's file structure here because Pachyderm's proto files use relative imports

PROTOC="npx grpc_tools_node_protoc"
PROTOC_GEN_TS_PATH="$(pwd)/node_modules/.bin/protoc-gen-ts"
PROTOC_GEN_GRPC_PATH="$(pwd)/node_modules/.bin/grpc_tools_node_protoc_plugin"
PACHYDERM_VERSION="$(jq -r .proto ../../../version.json)"
GOGO_VERSION="v1.3.2"
OUT_DIR="proto"

echo DEBUG INFO
echo -----------------------------------------------------------------------------------------------------------------------
echo This is a string interpolated URL this script is downloading from:
echo https://raw.githubusercontent.com/pachyderm/pachyderm/${PACHYDERM_VERSION}/src/license/license.proto
echo
echo The following is a list of valid URLS.
echo If using a tag: 
echo https://raw.githubusercontent.com/pachyderm/pachyderm/v2.5.0-nightly.20230104/src/license/license.proto
echo
echo If using a released version:
echo https://raw.githubusercontent.com/pachyderm/pachyderm/v2.4.2/src/license/license.proto
echo
echo If using a sha:
echo https://raw.githubusercontent.com/pachyderm/pachyderm/a7f229de3963235f5561855d00fe11a5d2550d95/src/license/license.proto
echo -----------------------------------------------------------------------------------------------------------------------
echo

rm -rf $OUT_DIR
mkdir -p ${OUT_DIR}/license ${OUT_DIR}/enterprise ${OUT_DIR}/task ${OUT_DIR}/pfs ${OUT_DIR}/pps ${OUT_DIR}/auth ${OUT_DIR}/gogoproto ${OUT_DIR}/admin ${OUT_DIR}/version/versionpb ${OUT_DIR}/protoextensions
curl https://raw.githubusercontent.com/pachyderm/pachyderm/${PACHYDERM_VERSION}/src/license/license.proto > ${OUT_DIR}/license/license.proto
curl https://raw.githubusercontent.com/pachyderm/pachyderm/${PACHYDERM_VERSION}/src/enterprise/enterprise.proto > ${OUT_DIR}/enterprise/enterprise.proto
curl https://raw.githubusercontent.com/pachyderm/pachyderm/${PACHYDERM_VERSION}/src/task/task.proto > ${OUT_DIR}/task/task.proto
curl https://raw.githubusercontent.com/pachyderm/pachyderm/${PACHYDERM_VERSION}/src/pfs/pfs.proto > ${OUT_DIR}/pfs/pfs.proto
curl https://raw.githubusercontent.com/pachyderm/pachyderm/${PACHYDERM_VERSION}/src/pps/pps.proto > ${OUT_DIR}/pps/pps.proto
curl https://raw.githubusercontent.com/pachyderm/pachyderm/${PACHYDERM_VERSION}/src/auth/auth.proto > ${OUT_DIR}/auth/auth.proto
curl https://raw.githubusercontent.com/pachyderm/pachyderm/${PACHYDERM_VERSION}/src/admin/admin.proto > ${OUT_DIR}/admin/admin.proto
curl https://raw.githubusercontent.com/pachyderm/pachyderm/${PACHYDERM_VERSION}/src/version/versionpb/version.proto > ${OUT_DIR}/version/versionpb/version.proto
curl https://raw.githubusercontent.com/pachyderm/pachyderm/${PACHYDERM_VERSION}/src/protoextensions/log.proto > ${OUT_DIR}/protoextensions/log.proto
curl https://raw.githubusercontent.com/pachyderm/pachyderm/${PACHYDERM_VERSION}/src/protoextensions/validate.proto > ${OUT_DIR}/protoextensions/validate.proto
curl https://raw.githubusercontent.com/pachyderm/pachyderm/${PACHYDERM_VERSION}/src/protoextensions/json-schema-options.proto > ${OUT_DIR}/protoextensions/json-schema-options.proto
curl https://raw.githubusercontent.com/gogo/protobuf/${GOGO_VERSION}/gogoproto/gogo.proto > ${OUT_DIR}/gogoproto/gogo.proto

find $OUT_DIR -name "*.proto" | xargs \
  $PROTOC \
    --proto_path=${OUT_DIR} \
    --plugin=protoc-gen-grpc=${PROTOC_GEN_GRPC_PATH} \
    --plugin=protoc-gen-ts=${PROTOC_GEN_TS_PATH} \
    --grpc_out=grpc_js:${OUT_DIR} \
    --js_out=import_style=commonjs:${OUT_DIR} \
    --ts_out=grpc_js:${OUT_DIR}
