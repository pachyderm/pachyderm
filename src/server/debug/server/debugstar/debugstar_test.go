package debugstar

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/google/go-cmp/cmp"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
)

type testDumpFS fstest.MapFS

var _ DumpFS = (*testDumpFS)(nil)

func (fs testDumpFS) Write(file string, f func(w io.Writer) error) error {
	buf := new(bytes.Buffer)
	if err := f(buf); err != nil {
		return errors.Wrap(err, "write.f")
	}
	fs[file] = &fstest.MapFile{
		Data: buf.Bytes(),
	}
	return nil
}

func TestStarlark(t *testing.T) {
	testData := []struct {
		script string
		want   fstest.MapFS
	}{
		{
			script: "test.star",
			want: fstest.MapFS{
				"foo.txt": &fstest.MapFile{
					Data: []byte(string("this is foo")),
				},
				"bar.txt": &fstest.MapFile{
					Data: []byte(string("this is bar")),
				},
				"baz/baz.txt": &fstest.MapFile{
					Data: []byte(string("this is baz")),
				},
				"quux.txt": &fstest.MapFile{
					Data: []byte(string("this is quux")),
				},
			},
		},
	}

	for _, test := range testData {
		t.Run(test.script, func(t *testing.T) {
			script, err := os.ReadFile(filepath.Join("testdata", test.script))
			if err != nil {
				t.Fatalf("read script: %v", err)
			}
			ctx := pctx.TestContext(t)
			got := make(fstest.MapFS)
			env := &Env{FS: testDumpFS(got)}
			if err := env.RunStarlark(ctx, test.script, string(script)); err != nil {
				t.Errorf("run script: %v", err)
			}
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("produced files (-want +got):\n%s", diff)
			}
		})
	}
}

func TestInteractiveDumpFS(t *testing.T) {
	base := t.TempDir()
	fs := &InteractiveDumpFS{Base: base}
	want := []byte("Hello, world.\n")
	if err := fs.Write("test.txt", func(w io.Writer) error {
		fmt.Fprintf(w, "%s", want)
		return nil
	}); err != nil {
		t.Fatalf("write test.txt: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(base, "test.txt"))
	if err != nil {
		t.Fatalf("read produced file: %v", err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("content (-want +got):\n%v", diff)
	}
}

// io.NopCloser is a nop closer for readers, not writers.
type nopCloser struct{ io.Writer }

func (nopCloser) Close() error { return nil }

func TestLocalDumpFS(t *testing.T) {
	buf := new(bytes.Buffer)
	fs := &LocalDumpFS{w: nopCloser{buf}}
	want := map[string][]byte{
		"a.txt":                           []byte("this is a"),
		"deeply/nested/file/called/b.txt": []byte("this is b"),
	}
	for name, data := range want {
		if err := fs.Write(name, func(w io.Writer) error {
			_, err := w.Write(data)
			return errors.Wrap(err, "write")
		}); err != nil {
			t.Fatalf("write %v: %v", name, err)
		}
	}
	if err := fs.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	gzr, err := gzip.NewReader(buf)
	if err != nil {
		t.Fatalf("new gzip reader: %v", err)
	}
	tr := tar.NewReader(gzr)
	got := map[string][]byte{}
	for {
		h, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			t.Fatalf("read header: %v", err)
		}
		b, err := io.ReadAll(tr)
		if err != nil {
			t.Fatalf("read %v data: %v", h.Name, err)
		}
		got[h.Name] = b
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("content (-want +got):\n%s", diff)
	}
}
