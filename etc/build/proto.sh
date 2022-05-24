# This tool downloads .proto files and generates the corresponding js/ts files
# Is is important to mirror Pachyderm's file structure here because Pachyderm's proto files use relative imports

PROTOC="$(npm bin)/grpc_tools_node_protoc"
PROTOC_GEN_TS_PATH="$(npm bin)/protoc-gen-ts"
PROTOC_GEN_GRPC_PATH="$(npm bin)/grpc_tools_node_protoc_plugin"
PACHYDERM_VERSION="v$(jq -r .pachyderm version.json)"
GOGO_VERSION="v1.3.2"
OUT_DIR="src/proto"

rm -rf $OUT_DIR
mkdir -p ${OUT_DIR}/license ${OUT_DIR}/enterprise ${OUT_DIR}/task ${OUT_DIR}/pfs ${OUT_DIR}/pps ${OUT_DIR}/auth ${OUT_DIR}/projects ${OUT_DIR}/gogoproto ${OUT_DIR}/admin
curl https://raw.githubusercontent.com/pachyderm/pachyderm/${PACHYDERM_VERSION}/src/license/license.proto > ${OUT_DIR}/license/license.proto
curl https://raw.githubusercontent.com/pachyderm/pachyderm/${PACHYDERM_VERSION}/src/enterprise/enterprise.proto > ${OUT_DIR}/enterprise/enterprise.proto
curl https://raw.githubusercontent.com/pachyderm/pachyderm/${PACHYDERM_VERSION}/src/task/task.proto > ${OUT_DIR}/task/task.proto
curl https://raw.githubusercontent.com/pachyderm/pachyderm/${PACHYDERM_VERSION}/src/pfs/pfs.proto > ${OUT_DIR}/pfs/pfs.proto
curl https://raw.githubusercontent.com/pachyderm/pachyderm/${PACHYDERM_VERSION}/src/pps/pps.proto > ${OUT_DIR}/pps/pps.proto
curl https://raw.githubusercontent.com/pachyderm/pachyderm/${PACHYDERM_VERSION}/src/auth/auth.proto > ${OUT_DIR}/auth/auth.proto
curl https://raw.githubusercontent.com/pachyderm/pachyderm/${PACHYDERM_VERSION}/src/admin/admin.proto > ${OUT_DIR}/admin/admin.proto
curl https://raw.githubusercontent.com/gogo/protobuf/${GOGO_VERSION}/gogoproto/gogo.proto > ${OUT_DIR}/gogoproto/gogo.proto
cp src/mock/projects.proto ${OUT_DIR}/projects/projects.proto

find $OUT_DIR -name "*.proto" | xargs \
  $PROTOC \
    --proto_path=${OUT_DIR} \
    --plugin=protoc-gen-grpc=${PROTOC_GEN_GRPC_PATH} \
    --plugin=protoc-gen-ts=${PROTOC_GEN_TS_PATH} \
    --grpc_out=grpc_js:${OUT_DIR} \
    --js_out=import_style=commonjs:${OUT_DIR} \
    --ts_out=grpc_js:${OUT_DIR}
