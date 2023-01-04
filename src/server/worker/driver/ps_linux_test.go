//go:build linux

package driver

import (
	"fmt"
	"strings"
	"syscall"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/server/worker/logs"
)

type testLogger struct {
	logs.TaggedLogger
	entries []string
}

func (l *testLogger) Logf(format string, args ...any) {
	l.entries = append(l.entries, fmt.Sprintf(format, args...))
}

func TestLogRunningProcesses(t *testing.T) {
	l := new(testLogger)
	logRunningProcesses(l, syscall.Getpgrp())
	t.Logf("entries:\n\t%v", strings.Join(l.entries, "\n\t"))
	if len(l.entries) < 1 {
		t.Errorf("expected logs")
	}
	for _, ent := range l.entries {
		if strings.HasPrefix(ent, "warning: ") {
			t.Errorf("unexpected warning: %q", ent)
		}
	}
}
