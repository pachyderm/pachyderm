//go:build linux

package driver

import (
	"regexp"
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
	var found bool
	finder := regexp.MustCompile(`^note: about to kill.*driver[.]test`)
	for _, log := range l.Logs {
		if finder.MatchString(log) {
			found = true
			break
		}
	}
	if !found {
		logs := new(strings.Builder)
		for i, log := range l.Logs {
			if i != 0 {
				logs.WriteRune('\n')
			}
			logs.WriteString("    ")
			logs.WriteString(log)
		}
		t.Errorf("did not get info about self (driver.test); all logs:\n%v", logs.String())
	}
}
