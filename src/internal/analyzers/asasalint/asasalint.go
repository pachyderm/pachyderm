// Package asasalint aliases github.com/alingse/asasalint.
package asasalint

import (
	"fmt"

	"github.com/alingse/asasalint"
	"golang.org/x/tools/go/analysis"
)

var Analyzer *analysis.Analyzer

func init() {
	var err error
	Analyzer, err = asasalint.NewAnalyzer(asasalint.LinterSetting{})
	if err != nil {
		panic(fmt.Sprintf("create asasalint linter: %v", err))
	}
}
