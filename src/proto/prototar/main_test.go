package main

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
)

func write(t *testing.T, name string, content []byte) {
	t.Helper()
	dir, _ := filepath.Split(name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("make parent: %v", err)
	}
	if err := os.WriteFile(name, content, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func TestCreateAndApply(t *testing.T) {
	ctx := pctx.TestContext(t)
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(orig); err != nil {
			t.Fatalf("change back to original working directory: %v", err)
		}
	})

	now := time.Now()
	f := fstest.MapFS{
		"out/forgotten": &fstest.MapFile{
			Data: []byte("this file was not included in the archive"),
		},
		"out/pachyderm/proto-docs.json": &fstest.MapFile{
			Data:    []byte("{}"),
			Mode:    fs.ModePerm,
			ModTime: now,
		},
		"out/pachyderm/src/new/new.pb.go": &fstest.MapFile{
			Data:    []byte("new file"),
			Mode:    fs.ModePerm,
			ModTime: now,
		},
		"out/pachyderm/src/existing/existing.pb.go": &fstest.MapFile{
			Data:    []byte("existing file"),
			Mode:    fs.ModePerm,
			ModTime: now,
		},
		"out/github.com/whatever/src/modified/modified.pb.go": &fstest.MapFile{
			Data:    []byte("new content"),
			Mode:    fs.ModePerm,
			ModTime: now,
		},
	}
	tar := new(bytes.Buffer)
	forgotten := new(bytes.Buffer)
	req := &createRequest{
		tar:       tar,
		forgotten: forgotten,
		root:      f,
		dirs:      []string{"out/pachyderm", "out/github.com/whatever"},
	}
	if err := create(ctx, req); err != nil {
		t.Fatalf("create: %v", err)
	}
	if got, want := forgotten.String(), "out/forgotten\n"; got != want {
		diff := cmp.Diff(want, got)
		t.Errorf("forgotten files: (-want +got)\n%s", diff)
	}

	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	write(t, filepath.Join(tmp, "src", "existing", "existing.pb.go"), []byte("existing file"))
	write(t, filepath.Join(tmp, "src", "modified", "modified.pb.go"), []byte("old content"))
	applyReport = new(applicationReport)
	if err := apply(ctx, tar); err != nil {
		t.Fatalf("apply: %v", err)
	}
	wantReport := &applicationReport{
		Modified:  []string{"src/modified/modified.pb.go"},
		Unchanged: []string{"src/existing/existing.pb.go"},
		Added:     []string{"src/new/new.pb.go", "proto-docs.json"},
	}
	if diff := cmp.Diff(wantReport, applyReport, cmpopts.SortSlices(func(a, b string) bool { return a < b })); diff != "" {
		t.Errorf("application report: (-want +got)\n%s", diff)
	}
}
