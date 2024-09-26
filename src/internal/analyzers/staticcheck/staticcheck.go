// Package staticcheck implements static checks based on https://github.com/sluongng/nogo-analyzer
package staticcheck

import (
	"fmt"

	"golang.org/x/tools/go/analysis"
	"honnef.co/go/tools/analysis/lint"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
	"honnef.co/go/tools/unused"
)

var (
	// Value to be added during stamping
	name = "dummy value please replace using x_defs"

	// Exported analyzer to be consumed by rules_go's nogo
	Analyzer = FindAnalyzerByName(name)
)

var analyzers = func() map[string]*analysis.Analyzer {
	resMap := make(map[string]*analysis.Analyzer)

	for _, analyzers := range [][]*lint.Analyzer{
		simple.Analyzers,
		staticcheck.Analyzers,
		stylecheck.Analyzers,
		{unused.Analyzer},
	} {
		for _, a := range analyzers {
			resMap[a.Analyzer.Name] = a.Analyzer
		}
	}

	return resMap
}()

func FindAnalyzerByName(name string) *analysis.Analyzer {
	if a, ok := analyzers[name]; ok {
		return a
	}
	panic(fmt.Sprintf("no analzyer %v: run `bazel run @co_honnef_go_tools//cmd/staticcheck -- -list-checks` for a list", name))
}
