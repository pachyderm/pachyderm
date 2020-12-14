# This tool downloads .proto files and generates the corresponding js/ts files
# Everything under /client is part of Pachyderm
# Is is important to preserve file structure here because Pachyderm's proto files use relative imports

PROTOC="$(npm bin)/grpc_tools_node_protoc"
PROTOC_GEN_TS_PATH="$(npm bin)/protoc-gen-ts"
PROTOC_GEN_GRPC_PATH="$(npm bin)/grpc_tools_node_protoc_plugin"
PACHYDERM_VERSION='v1.11.8'
GOGO_VERSION='v1.3.1'
OUT_DIR="pb"

rm -rf $OUT_DIR
mkdir -p dist ${OUT_DIR}/client/pfs ${OUT_DIR}/client/pps ${OUT_DIR}/client/auth ${OUT_DIR}/gogoproto
curl https://raw.githubusercontent.com/pachyderm/pachyderm/${PACHYDERM_VERSION}/src/client/pfs/pfs.proto > ${OUT_DIR}/client/pfs/pfs.proto
curl https://raw.githubusercontent.com/pachyderm/pachyderm/${PACHYDERM_VERSION}/src/client/pps/pps.proto > ${OUT_DIR}/client/pps/pps.proto
curl https://raw.githubusercontent.com/pachyderm/pachyderm/${PACHYDERM_VERSION}/src/client/auth/auth.proto > ${OUT_DIR}/client/auth/auth.proto
curl https://raw.githubusercontent.com/gogo/protobuf/${GOGO_VERSION}/gogoproto/gogo.proto > ${OUT_DIR}/gogoproto/gogo.proto

find $OUT_DIR -name "*.proto" | xargs \
  $PROTOC \
    --proto_path=${OUT_DIR} \
    --plugin=protoc-gen-grpc=${PROTOC_GEN_GRPC_PATH} \
    --plugin=protoc-gen-ts=${PROTOC_GEN_TS_PATH} \
    --grpc_out=grpc_js:${OUT_DIR} \
    --js_out=import_style=commonjs:${OUT_DIR} \
    --ts_out=grpc_js:${OUT_DIR}
