module github.com/pachyderm/pachyderm/v2/etc/proto

// This is not a real go.mod file.  It only exists to provide a file named go.mod for Gazelle before
// we migrate the rest of Pachyderm to Bazel.
//
// Hopefully you will never have to change this file, but if you do... run go mod tidy and remove
// pachyderm as a direct dependency.  (Gazelle can't handle an on-disk replace directives, but
// eventually we won't need it.)

go 1.21.4

require (
	github.com/chrusty/protoc-gen-jsonschema v0.0.0-20230806074516-0ca6ba213e83
	github.com/golang/protobuf v1.5.3
	go.uber.org/zap v1.24.0
	google.golang.org/grpc/cmd/protoc-gen-go-grpc v1.3.0
	google.golang.org/protobuf v1.31.0
)

require (
	github.com/alecthomas/jsonschema v0.0.0-20210918223802-a1d3f4b43d7b // indirect
	github.com/envoyproxy/protoc-gen-validate v1.0.2 // indirect
	github.com/fatih/camelcase v1.0.0 // indirect
	github.com/iancoleman/orderedmap v0.2.0 // indirect
	github.com/iancoleman/strcase v0.2.0 // indirect
	github.com/sirupsen/logrus v1.9.0 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190809123943-df4f5c81cb3b // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	golang.org/x/sys v0.13.0 // indirect
)

// Temporary hack: uncomment this to run "go mod tidy"
// replace github.com/pachyderm/pachyderm/v2 => ../..
