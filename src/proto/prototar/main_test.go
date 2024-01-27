package main

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
	"time"

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
	a, b := fstest.MapFS{
		"proto-docs.json": &fstest.MapFile{
			Data:    []byte("{}"),
			Mode:    fs.ModePerm,
			ModTime: now,
		},
		"src/new/new.pb.go": &fstest.MapFile{
			Data:    []byte("new file"),
			Mode:    fs.ModePerm,
			ModTime: now,
		},
		"src/existing/existing.pb.go": &fstest.MapFile{
			Data:    []byte("existing file"),
			Mode:    fs.ModePerm,
			ModTime: now,
		},
	}, fstest.MapFS{
		"src/modified/modified.pb.go": &fstest.MapFile{
			Data:    []byte("new content"),
			Mode:    fs.ModePerm,
			ModTime: now,
		},
	}
	w := new(bytes.Buffer)
	if err := create(ctx, w, a, b); err != nil {
		t.Fatalf("create: %v", err)
	}

	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	write(t, filepath.Join(tmp, "src", "existing", "existing.pb.go"), []byte("existing file"))
	write(t, filepath.Join(tmp, "src", "modified", "modified.pb.go"), []byte("old content"))
	if err := apply(ctx, w); err != nil {
		t.Fatalf("apply: %v", err)
	}
}
