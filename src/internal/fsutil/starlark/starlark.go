package starlark

import (
	"io/fs"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/fsutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/imagebuilder/jobs"
	ourstar "github.com/pachyderm/pachyderm/v2/src/internal/starlark"
	"go.starlark.net/starlark"
)

var Module = starlark.StringDict{
	"find": starlark.NewBuiltin("find", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		if len(kwargs) != 0 {
			return nil, errors.New("unexpected kwargs")
		}
		var fss []fs.FS
		for i, arg := range args {
			r, ok := arg.(*ourstar.Reflect)
			if !ok {
				return nil, errors.Errorf("arg %v: not a Reflect", i)
			}
			withfs, ok := r.Any.(jobs.WithFS)
			if !ok {
				return nil, errors.Errorf("arg %v: not a WithFS", i)
			}
			fss = append(fss, withfs.FS())
		}
		var infos []fs.FileInfo
		for i, fs := range fss {
			info, err := fsutil.Find(fs)
			if err != nil {
				return nil, errors.Wrapf(err, "arg %v: find", i)
			}
			infos = append(infos, info...)
		}
		return ourstar.ReflectList(infos), nil
	}),
}

func init() {
	ourstar.Modules["fsutil"] = Module
}
