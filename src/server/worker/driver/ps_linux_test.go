//go:build linux

package driver

import (
	"strings"
	"syscall"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/logs"
)

func TestLogRunningProcesses(t *testing.T) {
	ctx := pctx.TestContext(t)
	l := logs.NewTest(ctx)
	logRunningProcesses(l, syscall.Getpgrp())
	if len(l.Logs) < 1 {
		t.Errorf("expected logs")
	}
	for _, ent := range l.Logs {
		if strings.HasPrefix(ent, "warning: ") {
			t.Errorf("unexpected warning: %q", ent)
		}
	}
}
