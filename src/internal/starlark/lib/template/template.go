// Package template is a text/template binding for Starlark.
package template

import (
	"bytes"
	"text/template"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	ourstar "github.com/pachyderm/pachyderm/v2/src/internal/starlark"
	"go.starlark.net/starlark"
)

var Module = starlark.StringDict{
	"run": starlark.NewBuiltin("run", func(th *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var tpl string
		var data ourstar.Any
		if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "template", &tpl, "data", &data); err != nil {
			return nil, errors.Wrap(err, "unpack args")
		}
		t, err := template.New("<arg>").Parse(tpl)
		if err != nil {
			return nil, errors.Wrap(err, "parse template")
		}
		var result bytes.Buffer
		if err := t.Execute(&result, data.Value); err != nil {
			return nil, errors.Wrap(err, "execute template")
		}
		return starlark.String(result.String()), nil
	},
	),
}

func init() {
	ourstar.RegisterPersonality("template", ourstar.Options{
		Predefined: Module,
	})
}
