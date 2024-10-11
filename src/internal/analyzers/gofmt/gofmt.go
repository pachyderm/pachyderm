// Package gofmt provides an analyzer which checks that files have been run
// through gofmt.
package gofmt

import (
	"go/token"
	"path/filepath"

	"github.com/golangci/gofmt/gofmt"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "gofmt",
	Doc:  "checks if files have been run through `gofmt -s`",
	Run: func(pass *analysis.Pass) (any, error) {
		files := map[string]token.Pos{}
		for _, f := range pass.Files {
			fileName := pass.Fset.PositionFor(f.Pos(), true).Filename
			ext := filepath.Ext(fileName)
			if ext != "" && ext != ".go" {
				// position has been adjusted to a non-go file, revert to original file
				fileName = pass.Fset.PositionFor(f.Pos(), false).Filename
			}
			files[fileName] = f.Pos()
		}

		for f, pos := range files {
			diff, err := gofmt.RunRewrite(f, true, nil)
			if err != nil {
				return nil, errors.Wrap(err, "gofmt.RunRewrite")
			}
			if len(diff) > 0 {
				pass.Report(analysis.Diagnostic{
					Pos:     pos,
					Message: "not formatted with gofmt -s:\n" + string(diff),
				})
			}
		}
		return nil, nil
	},
}
