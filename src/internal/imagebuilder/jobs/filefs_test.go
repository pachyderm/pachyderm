package jobs

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"io/fs"

	"github.com/google/go-cmp/cmp"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/fsutil"
	"golang.org/x/exp/maps"
)

func TestFileFS(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "abc123"), []byte("hello from disk"), 0o600); err != nil {
		t.Fatal(err)
	}
	testData := []struct {
		name string
		fs   fs.FS
		want map[string]string
	}{
		{
			name: "mem",
			fs: &FileFS{
				Name: "x",
				Data: []byte("hello from memory"),
			},
			want: map[string]string{"x": "hello from memory"},
		},
		{
			name: "disk",
			fs: &FileFS{
				Name: "x",
				Path: filepath.Join(dir, "abc123"),
			},
			want: map[string]string{"x": "hello from disk"},
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			if err := fstest.TestFS(test.fs, maps.Keys(test.want)...); err != nil {
				t.Errorf("TestFS: %v", err)
			}
			got := make(map[string]string)
			err := fs.WalkDir(test.fs, ".", func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return errors.Wrap(err, path)
				}
				if !d.Type().IsRegular() {
					return nil
				}
				content, err := fs.ReadFile(test.fs, path)
				if err != nil {
					return errors.Wrap(err, path)
				}
				got[path] = string(content)
				return nil
			})
			if err != nil {
				t.Fatalf("walk: %v", err)
			}
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("files (-want +got):\n%s", diff)
			}
		})
	}
}
