package main

import (
	"github.com/hashicorp/go-plugin"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pps"
)

func Filter(inputs []*pfs.FileInfo) bool {
	var path string
	for _, in := range inputs {
		if path == "" {
			path = in.File.Path
		} else if path != in.File.Path {
			return false
		}
	}
	return true
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: pps.FilterHandshake,
		Plugins: map[string]plugin.Plugin{
			"Filter": &pps.FilterGRPCPlugin{Impl: Filter},
		},

		// A non-nil value here enables gRPC serving for this plugin...
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
