package proto

import (
	_ "github.com/chrusty/protoc-gen-jsonschema/cmd/protoc-gen-jsonschema"
	_ "google.golang.org/grpc/cmd/protoc-gen-go-grpc"

	// Because the fake go.mod doesn't consider this package to be part of
	// pachyderm/pachyderm/v2, direct dependencies of protosupport become indirect dependencies
	// of this.  But we need them to be direct, because we are using Bazel to build them when
	// they are included in ./protoc-gen-*.
	_ "go.uber.org/zap"
)
