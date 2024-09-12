package wrapcheck

import (
	"github.com/tomarrell/wrapcheck/v2/wrapcheck"
	"golang.org/x/tools/go/analysis"
)

var Analyzer *analysis.Analyzer

func init() {
	cfg := wrapcheck.NewDefaultConfig()
	cfg.IgnoreSigs = []string{
		"github.com/pachyderm/pachyderm/v2/src/internal/errors.Errorf",
		"github.com/pachyderm/pachyderm/v2/src/internal/errors.New",
		"github.com/pachyderm/pachyderm/v2/src/internal/errors.Unwrap",
		"github.com/pachyderm/pachyderm/v2/src/internal/errors.EnsureStack",
		"github.com/pachyderm/pachyderm/v2/src/internal/errors.Join",
		"github.com/pachyderm/pachyderm/v2/src/internal/errors.JoinInto",
		"github.com/pachyderm/pachyderm/v2/src/internal/errors.Close",
		"github.com/pachyderm/pachyderm/v2/src/internal/errors.Invoke",
		"github.com/pachyderm/pachyderm/v2/src/internal/errors.Invoke1",
		"google.golang.org/grpc/status.Error",
		"google.golang.org/grpc/status.Errorf",
		"(*google.golang.org/grpc/internal/status.Status).Err",
		"google.golang.org/protobuf/types/known/anypb.New",
		".Wrap(",
		".Wrapf(",
		".WithMessage(",
		".WithMessagef(",
		".WithStack(",
	}
	cfg.IgnorePackageGlobs = []string{
		"github.com/pachyderm/pachyderm/v2/src/*",
	}
	cfg.IgnoreInterfaceRegexps = []string{
		`^fileset\.`,
		`^collection\.`,
		`^track\.`,
	}
	Analyzer = wrapcheck.NewAnalyzer(cfg)
}
