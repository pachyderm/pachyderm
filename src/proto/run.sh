#!/usr/bin/env bash

set -euo pipefail
IFS=$'\n\t'

# --- begin runfiles.bash initialization v3 ---
# Copy-pasted from the Bazel Bash runfiles library v3.
set -uo pipefail; set +e; f=bazel_tools/tools/bash/runfiles/runfiles.bash
# shellcheck source=/dev/null.
source "${RUNFILES_DIR:-/dev/null}/$f" 2>/dev/null || \
source "$(grep -sm1 "^$f " "${RUNFILES_MANIFEST_FILE:-/dev/null}" | cut -f2- -d' ')" 2>/dev/null || \
source "$0.runfiles/$f" 2>/dev/null || \
source "$(grep -sm1 "^$f " "$0.runfiles_manifest" | cut -f2- -d' ')" 2>/dev/null || \
source "$(grep -sm1 "^$f " "$0.exe.runfiles_manifest" | cut -f2- -d' ')" 2>/dev/null || \
{ echo>&2 "ERROR: cannot find $f; this script must be run with 'bazel run'"; exit 1; }; f=; set -e
# --- end runfiles.bash initialization v3 ---

cd "$BUILD_WORKSPACE_DIRECTORY" # Where your working copy is.

OUT=/tmp/pachyderm-gen-proto-out

rm -rf $OUT
function cleanup(){
    rm -rf $OUT
}
trap cleanup EXIT

mkdir -p $OUT/src
mkdir -p $OUT/src/internal/jsonschema
mkdir -p $OUT/src/openapi
mkdir -p $OUT/src/typescript

PROTOS=("$@")

for i in "${PROTOS[@]}"; do \
    if ! grep -q 'go_package' "${i}"; then
        echo -e "\e[1;31mError:\e[0m missing \"go_package\" declaration in ${i}" >/dev/stderr
        exit 1
    fi
done

"$(rlocation _main/src/proto/protoc)" \
    -I"$(dirname "$(dirname "$(dirname "$(rlocation com_google_protobuf/src/google/protobuf/any.proto)")")")" \
    -Isrc \
    --plugin=protoc-gen-go="$(rlocation _main/src/proto/protoc-gen-go)" \
    --plugin=protoc-gen-go-grpc="$(rlocation org_golang_google_grpc_cmd_protoc_gen_go_grpc/protoc-gen-go-grpc_/protoc-gen-go-grpc)" \
    --plugin=protoc-gen-zap="$(rlocation _main/src/proto/protoc-gen-zap/protoc-gen-zap_/protoc-gen-zap)" \
    --plugin=protoc-gen-pach="$(rlocation _main/src/proto/pachgen/pachgen_/pachgen)" \
    --plugin=protoc-gen-doc="$(rlocation _main/src/proto/protoc-gen-doc)" \
    --plugin=protoc-gen-doc2="$(rlocation _main/src/proto/protoc-gen-doc)" \
    --plugin=protoc-gen-jsonschema="$(rlocation com_github_chrusty_protoc_gen_jsonschema/cmd/protoc-gen-jsonschema/protoc-gen-jsonschema_/protoc-gen-jsonschema)" \
    --plugin=protoc-gen-validate="$(rlocation _main/src/proto/protoc-gen-validate-go)" \
    --plugin=protoc-gen-openapiv2="$(rlocation _main/src/proto/protoc-gen-openapiv2)" \
    --plugin=protoc-gen-grpc-gateway="$(rlocation _main/src/proto/protoc-gen-grpc-gateway)" \
    --plugin=protoc-gen-grpc-gateway-ts="$(rlocation _main/src/proto/protoc-gen-grpc-gateway-ts)" \
    --zap_out="$OUT" \
    --pach_out="$OUT/src" \
    --go_out="$OUT" \
    --go-grpc_out="$OUT" \
    --jsonschema_out="$OUT/src/internal/jsonschema" \
    --validate_out="$OUT" \
    --doc_out="$OUT" \
    --doc2_out="$OUT" \
    --openapiv2_out="$OUT/src/openapi" \
    --grpc-gateway_out="$OUT" \
    --grpc-gateway-ts_out="$OUT/src/typescript" \
    --jsonschema_opt="enforce_oneof" \
    --jsonschema_opt="file_extension=schema.json" \
    --jsonschema_opt="disallow_additional_properties" \
    --jsonschema_opt="enums_as_strings_only" \
    --jsonschema_opt="disallow_bigints_as_strings" \
    --jsonschema_opt="prefix_schema_files_with_package" \
    --jsonschema_opt="json_fieldnames" \
    --doc_opt="json,proto-docs.json" \
    --doc2_opt="markdown,proto-docs.md" \
    --grpc-gateway_opt logtostderr=true \
    --grpc-gateway_opt generate_unbound_methods=true \
    --openapiv2_opt logtostderr=true \
    --openapiv2_opt generate_unbound_methods=true \
    --openapiv2_opt merge_file_name=pachyderm_api \
    --openapiv2_opt disable_service_tags=true \
    --openapiv2_opt preserve_rpc_order=true \
    --openapiv2_opt allow_merge=true \
    --openapiv2_opt merge_file_name=pachyderm_api \
    "${PROTOS[@]}"

pushd $OUT >/dev/null
echo -n "gopatch..."
"$(rlocation _main/src/proto/gopatch)" ./... -p="$(rlocation _main/src/proto/proto.patch)"
echo "done."
echo -n "gofmt..."
"$(rlocation go_sdk/bin/gofmt)" -w .
echo "done."
popd >/dev/null

echo -n "copy generated files into workspace..."
find src/internal/jsonschema -name \*.schema.json -exec rm {} '+'
cp -a $OUT/src/ .
cp -a $OUT/github.com/pachyderm/pachyderm/v2/src/ .
echo "done."
