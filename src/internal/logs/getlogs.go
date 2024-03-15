package logs

import (
	"context"
	"io"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/lokiutil"
	lokiclient "github.com/pachyderm/pachyderm/v2/src/internal/lokiutil/client"
)

type GetLogs struct {
	Matchers      LineMatcher
	GetLokiClient func() (*lokiclient.Client, error)
}

type LogSource interface {
	Stream(context.Context) (io.ReadCloser, error)
}

// LineMatcher Match() selects lines that match its criteria and returns them through sendMatch
type LineMatcher interface {
	Match(line string, sendMatch func([]byte) error) error
}

func (gl *GetLogs) StartMatching(ctx context.Context, sendMatch func([]byte) error) error {

	loki, err := gl.GetLokiClient()
	if err != nil {
		return errors.EnsureStack(err)
	}
	_ = lokiutil.QueryRange(ctx, loki, `{app="pachd"}`, time.Time{}, time.Time{}, false, func(t time.Time, line string) error {
		gl.Matchers.Match(line, sendMatch)

		return nil
	})

	return nil
}
