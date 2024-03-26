package main_test

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/bazelbuild/rules_go/go/tools/bazel"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
)

func TestMain_(t *testing.T) {
	ctx := pctx.TestContext(t)
	testpachd, ok := bazel.FindBinary("//src/testing/testpachd", "testpachd")
	if !ok {
		t.Log("cannot find testpachd binary; using version in $PATH")
		testpachd = "testpachd"
	}
	r, w := io.Pipe()
	cmd := exec.CommandContext(ctx, testpachd, "-context=")
	cmd.Stderr = log.WriterAt(pctx.Child(ctx, "testpachd"), log.DebugLevel)
	cmd.Stdout = w
	cmd.Cancel = func() error {
		// This gives testpachd a chance to clean up even if this test fails.
		return cmd.Process.Signal(os.Interrupt)
	}
	if err := cmd.Start(); err != nil {
		t.Fatalf("start testpachd: %v", err)
	}
	var addr string
	s := bufio.NewScanner(r)
	for s.Scan() {
		t.Logf("read input from child: %v", s.Text())
		addr = s.Text()
		break
	}
	if err := s.Err(); err != nil {
		t.Errorf("scanner encountered an error: %v", err)
	}
	addr = strings.TrimSpace(addr)
	pachAddr, err := grpcutil.ParsePachdAddress(string(addr))
	if err != nil {
		t.Fatalf("parse pachd address %s: %v", addr, err)
	}
	client, err := client.NewFromPachdAddress(ctx, pachAddr)
	if err != nil {
		t.Fatalf("new pach client: %v", err)
	}
	v, err := client.Version()
	if err != nil {
		t.Fatalf("read version: %v", err)
	}
	t.Logf("server version: %v", v)
	if err := cmd.Process.Signal(os.Interrupt); err != nil {
		t.Errorf("kill testpachd: %v", err)
	}
	cmd.Cancel = func() error { return cmd.Process.Kill() }
	if err := cmd.Wait(); err != nil {
		t.Errorf("unexpectedly exited non-zero: %v", err)
	}
}
