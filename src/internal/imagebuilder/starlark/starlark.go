// Package starlark exposes the build system to Starlark.
package starlark

import (
	"path/filepath"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/imagebuilder/jobs"
	ourstar "github.com/pachyderm/pachyderm/v2/src/internal/starlark"
	"go.starlark.net/starlark"
)

var Module = starlark.StringDict{
	"path": starlark.NewBuiltin("path", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		if len(kwargs) > 0 {
			return nil, errors.New("unexpected kwargs")
		}
		var fragment string
		if err := starlark.UnpackPositionalArgs(fn.Name(), args, kwargs, 1, &fragment); err != nil {
			return nil, errors.Wrap(err, "unpack args")
		}
		got := filepath.Join(filepath.Dir(thread.CallFrame(1).Pos.Filename()), fragment)
		result, err := filepath.Abs(got)
		if err != nil {
			return nil, errors.Wrapf(err, "resolve abolute path %v", got)
		}
		return starlark.String(result), nil
	}),
	"download_file":      jobs.MakeStarlarkCommand[jobs.Download]("download_file"),
	"go_binary":          jobs.MakeStarlarkCommand[jobs.GoBinary]("go_binary"),
	"oci_layer":          jobs.MakeStarlarkCommand[jobs.FSLayer]("oci_layer"),
	"oci_image_config":   starlark.NewBuiltin("oci_image_config", jobs.NewImageConfigFromStarlark),
	"oci_image_manifest": jobs.MakeStarlarkCommand[jobs.BuildManifest]("oci_image_manifest"),
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
