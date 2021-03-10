# This tool downloads .proto files and generates the corresponding js/ts files
# Is is important to mirror Pachyderm's file structure here because Pachyderm's proto files use relative imports

PROTOC="$(npm bin)/grpc_tools_node_protoc"
PROTOC_GEN_TS_PATH="$(npm bin)/protoc-gen-ts"
PROTOC_GEN_GRPC_PATH="$(npm bin)/grpc_tools_node_protoc_plugin"
PACHYDERM_VERSION='v2.0.0-alpha.5'
GOGO_VERSION='v1.3.1'
OUT_DIR="pb"

rm -rf $OUT_DIR
mkdir -p dist ${OUT_DIR}/pfs ${OUT_DIR}/pps ${OUT_DIR}/auth ${OUT_DIR}/projects ${OUT_DIR}/gogoproto
curl https://raw.githubusercontent.com/pachyderm/pachyderm/${PACHYDERM_VERSION}/src/pfs/pfs.proto > ${OUT_DIR}/pfs/pfs.proto
curl https://raw.githubusercontent.com/pachyderm/pachyderm/${PACHYDERM_VERSION}/src/pps/pps.proto > ${OUT_DIR}/pps/pps.proto
curl https://raw.githubusercontent.com/pachyderm/pachyderm/${PACHYDERM_VERSION}/src/auth/auth.proto > ${OUT_DIR}/auth/auth.proto
curl https://raw.githubusercontent.com/gogo/protobuf/${GOGO_VERSION}/gogoproto/gogo.proto > ${OUT_DIR}/gogoproto/gogo.proto
cp mock/projects.proto ${OUT_DIR}/projects/projects.proto

find $OUT_DIR -name "*.proto" | xargs \
  $PROTOC \
    --proto_path=${OUT_DIR} \
    --plugin=protoc-gen-grpc=${PROTOC_GEN_GRPC_PATH} \
    --plugin=protoc-gen-ts=${PROTOC_GEN_TS_PATH} \
    --grpc_out=grpc_js:${OUT_DIR} \
    --js_out=import_style=commonjs:${OUT_DIR} \
    --ts_out=grpc_js:${OUT_DIR}
