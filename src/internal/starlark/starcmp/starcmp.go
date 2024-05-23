// Package starcmp provides utilities for running cmp.Diff on Starlark values.  It's not in the
// startest package because startest depends on internal/starlark, but the internal/starlark tests
// want to compare things.
package starcmp

import (
	"fmt"

	"github.com/google/go-cmp/cmp"
	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

// Compare is a cmp.Option that compares Starlark values with Starlark's `==` operator.
var Compare = cmp.Option(cmp.Comparer(func(a, b starlark.Value) bool {
	x, err := starlark.Compare(syntax.EQL, a, b)
	if err != nil {
		panic(fmt.Sprintf("compare failed: %v", err))
	}
	return x
}))
