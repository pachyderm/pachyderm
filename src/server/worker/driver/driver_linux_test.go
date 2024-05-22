//go:build linux

package driver

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmputil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/logs"
)

func TestLogUnexpectedSubprocesses(t *testing.T) {
	ctx := pctx.TestContext(t)
	d := &driver{
		pipelineInfo: &pps.PipelineInfo{
			Details: &pps.PipelineInfo_Details{
				Transform: &pps.Transform{
					Cmd: []string{
						"sh",
						"-c",
						"sleep infinity & echo hello",
					},
				},
			},
		},
	}
	logger := logs.NewTest(ctx)
	if err := d.RunUserCode(ctx, logger, nil); err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(logger.Errors, ([]string)(nil), cmpopts.EquateEmpty()); diff != "" {
		t.Errorf("unexpected errors:\n%s", diff)
	}
	var got []string
	for _, x := range logger.Logs {
		if !strings.HasPrefix(x, "warning:") { // Sometimes this races; as long as we see the expected messages, we're happy.
			got = append(got, x)
		}
	}
	want := []string{
		"/^beginning to run user code/",
		"/^note: about to kill unexpectedly-remaining subprocess.*sleep/",
		"/^finished running user code/",
	}
	if diff := cmp.Diff(got, want, cmputil.RegexpStrings()); diff != "" {
		t.Errorf("logs (-got +want):\n%s", diff)
	}
}
