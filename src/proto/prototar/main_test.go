package main

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

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

	// Create a proto bundle from this directory structure.
	f := fstest.MapFS{
		"out/forgotten": {
			Data: []byte("this file was not included in the archive"),
		},
		"out/pachyderm/proto-docs.json": {
			Data: []byte("{}"),
			Mode: fs.ModePerm,
		},
		"out/pachyderm/src/new/new.pb.go": {
			Data: []byte("new file"),
			Mode: fs.ModePerm,
		},
		"out/pachyderm/src/internal/jsonschema/new_v2/New.schema.json": {},
		"out/pachyderm/src/existing/existing.pb.go": {
			Data: []byte("existing file"),
			Mode: fs.ModePerm,
		},
		"out/github.com/whatever/src/modified/modified.pb.go": {
			Data: []byte("new content"),
			Mode: fs.ModePerm,
		},
	}

	tarW := new(bytes.Buffer)
	forgotten := new(bytes.Buffer)
	req := &createRequest{
		tar:       tarW,
		forgotten: forgotten,
		root:      f,
		dirs:      []string{"out/pachyderm", "out/github.com/whatever"},
	}
	if err := create(ctx, req); err != nil {
		t.Fatalf("create: %v", err)
	}
	tar := tarW.Bytes()

	// Check that we found the "forgotten" file, not in out/pachyderm or
	// out/github.com/whatever.
	if got, want := forgotten.String(), "out/forgotten\n"; got != want {
		diff := cmp.Diff(want, got)
		t.Errorf("forgotten files: (-want +got)\n%s", diff)
	}

	// Apply only works on the current working directory, so switch there.  At the top of the
	// test, we have a cleanup that reverts this chdir.
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	// Generated a representative directory tree.  This is supposed to look like your working
	// copy of Pachyderm.
	write(t, filepath.Join("src", "existing", "existing.pb.go"), []byte("existing file"))
	write(t, filepath.Join("src", "modified", "modified.pb.go"), []byte("old content"))

	// Apply the tar we generated above
	t.Run("apply_1", func(t *testing.T) {
		ctx := pctx.TestContext(t)
		gotReport, err := apply(ctx, bytes.NewReader(tar), false)
		if err != nil {
			t.Fatalf("apply: %v", err)
		}
		wantReport := &ApplicationReport{
			Modified:  []string{"src/modified/modified.pb.go"},
			Unchanged: []string{"src/existing/existing.pb.go"},
			Added:     []string{"src/new/new.pb.go", "proto-docs.json", "src/internal/jsonschema/new_v2/New.schema.json"},
		}
		if diff := cmp.Diff(wantReport, gotReport, cmpopts.SortSlices(func(a, b string) bool { return a < b })); diff != "" {
			t.Errorf("application report: (-want +got)\n%s", diff)
		}
	})

	// The case above tests what happens when src/internal/jsonschema doesn't exist, which is
	// not an error.  (If you're mad, you might rm -rf it.  That's fine.)  This case tests what
	// happens when there is a stale file in there.
	t.Run("apply_2", func(t *testing.T) {
		ctx := pctx.TestContext(t)
		write(t, filepath.Join("src", "internal", "jsonschema", "old_v2", "Old.schema.json"), []byte(`{"old":"schema"}`))
		write(t, filepath.Join("src", "internal", "jsonschema", "jsonschema.go"), []byte("some go code to not delete"))
		gotReport, err := apply(ctx, bytes.NewReader(tar), false)
		if err != nil {
			t.Fatalf("apply: %v", err)
		}
		wantReport := &ApplicationReport{
			Unchanged: []string{"src/existing/existing.pb.go", "src/modified/modified.pb.go", "src/new/new.pb.go", "proto-docs.json", "src/internal/jsonschema/new_v2/New.schema.json"},
			Deleted:   []string{"src/internal/jsonschema/old_v2/Old.schema.json"},
		}
		if diff := cmp.Diff(wantReport, gotReport, cmpopts.SortSlices(func(a, b string) bool { return a < b })); diff != "" {
			t.Errorf("application report: (-want +got)\n%s", diff)
		}
	})

	// Finally we should be able to run this again and see only unchanged files.
	t.Run("apply_3", func(t *testing.T) {
		ctx := pctx.TestContext(t)
		gotReport, err := apply(ctx, bytes.NewReader(tar), false)
		if err != nil {
			t.Fatalf("apply: %v", err)
		}
		wantReport := &ApplicationReport{
			Unchanged: []string{"src/existing/existing.pb.go", "src/modified/modified.pb.go", "src/new/new.pb.go", "proto-docs.json", "src/internal/jsonschema/new_v2/New.schema.json"},
		}
		if diff := cmp.Diff(wantReport, gotReport, cmpopts.SortSlices(func(a, b string) bool { return a < b })); diff != "" {
			t.Errorf("application report: (-want +got)\n%s", diff)
		}
	})

	// Try a dry run in an unwritable directory to make sure we don't touch any files.
	t.Run("dry_run", func(t *testing.T) {
		if err := os.Chdir("/"); err != nil {
			t.Fatalf("chdir /: %v", err)
		}
		ctx := pctx.TestContext(t)
		gotReport, err := apply(ctx, bytes.NewReader(tar), true)
		if err != nil {
			t.Fatalf("apply: %v", err)
		}
		wantReport := &ApplicationReport{
			Added: []string{"src/new/new.pb.go", "proto-docs.json", "src/internal/jsonschema/new_v2/New.schema.json", "src/existing/existing.pb.go", "src/modified/modified.pb.go"},
		}
		if diff := cmp.Diff(wantReport, gotReport, cmpopts.SortSlices(func(a, b string) bool { return a < b })); diff != "" {
			t.Errorf("application report: (-want +got)\n%s", diff)
		}
	})
}
