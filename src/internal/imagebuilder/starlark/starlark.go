// Package starlark exposes the build system to Starlark.
package starlark

import (
	"github.com/pachyderm/pachyderm/v2/src/internal/imagebuilder/jobs"
	ourstar "github.com/pachyderm/pachyderm/v2/src/internal/starlark"
	"go.starlark.net/starlark"
)

var Module = starlark.StringDict{
	"download_file": jobs.MakeStarlarkCommand[jobs.Download]("download_file"),
}

var DebugModule = starlark.StringDict{
	"registry": new(jobs.GlobalRegistry),
}

var ThreadLocals = map[string]any{
	jobs.StarlarkRegistryKey: new(jobs.Registry),
}

func init() {
	ourstar.Modules["build"] = Module
}
