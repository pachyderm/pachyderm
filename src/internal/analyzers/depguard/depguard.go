package depguard

import (
	"fmt"

	"github.com/OpenPeeDeeP/depguard/v2"
	"golang.org/x/tools/go/analysis"
)

var Analyzer *analysis.Analyzer

func init() {
	var err error
	Analyzer, err = depguard.NewAnalyzer(&depguard.LinterSettings{
		"main": &depguard.List{
			Files: []string{
				"$all",
				"!**/src/internal/log/*.go",
				"!**/etc/**/*.go",
			},
			Deny: map[string]string{
				"github.com/sirupsen/logrus":         "use the internal/log package",
				"go.uber.org/multierr":               "use the internal/errors package",
				"github.com/hashicorp/go-multierror": "use the internal/errors package",
				"github.com/gogo/protobuf/proto":     "wrong proto package; use google.golang.org/protobuf/proto",
				"github.com/gogo/protobuf/types":     "wrong proto package; use google.golang.org/protobuf/types/...",
				"github.com/gogo/protobuf/jsonpb":    "wrong proto package; use google.golang.org/protobuf/encoding/protojson",
				"github.com/golang/protobuf/proto":   "wrong proto package; use google.golang.org/protobuf/proto",
				"github.com/golang/protobuf/ptypes":  "wrong proto package; use google.golang.org/protobuf/types/...",
				"github.com/golang/protobuf/jsonpb":  "wrong proto package; use google.golang.org/protobuf/encoding/protojson",
			},
		},
	})
	if err != nil {
		panic(fmt.Sprintf("create depguard analyzer: %v", err))
	}
}
