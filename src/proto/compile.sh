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

# args: <output tar> <forgotten files report> <proto1.proto> <proto2.proto> ...
TAR="$1"
FORGOTTEN="$2"
shift
shift

PROTOS=("$@")

for i in "${PROTOS[@]}"; do \
    if ! grep -q 'go_package' "${i}"; then
        echo -e "\e[1;31mError:\e[0m missing \"go_package\" declaration in ${i}" >/dev/stderr
        exit 1
    fi
done

mkdir -p out/pachyderm/src/internal/jsonschema
mkdir -p out/pachyderm/src/openapi
mkdir -p out/pachyderm/src/typescript

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
    --zap_out="out" \
    --pach_out="out/pachyderm/src" \
    --go_out="out" \
    --go-grpc_out="out" \
    --jsonschema_out="out/pachyderm/src/internal/jsonschema" \
    --validate_out="out" \
    --doc_out="out/pachyderm" \
    --doc2_out="out/pachyderm" \
    --openapiv2_out="out/pachyderm/src/openapi" \
    --grpc-gateway_out="out" \
    --grpc-gateway-ts_out="out/pachyderm/src/typescript" \
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

echo -n "gopatch..."
"$(rlocation _main/src/proto/gopatch)" ./out/pachyderm/... ./out/github.com/pachyderm/pachyderm/v2/... -p="$(rlocation _main/src/proto/proto.patch)"
echo "done."

echo -n "gofmt..."
"$(rlocation go_sdk/bin/gofmt)" -w .
echo "done."

echo "package result..."
"$(rlocation _main/src/proto/prototar/prototar_/prototar)" create "$TAR" "$FORGOTTEN" out/pachyderm out/github.com/pachyderm/pachyderm/v2
echo "done."
