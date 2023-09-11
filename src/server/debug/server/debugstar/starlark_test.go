package debugstar

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/google/go-cmp/cmp"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"k8s.io/client-go/kubernetes/fake"
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
			script: "basic.star",
			want: fstest.MapFS{
				"foo": &fstest.MapFile{
					Data: []byte(string("this is foo")),
				},
				"bar": &fstest.MapFile{
					Data: []byte(string("this is bar")),
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
			env := &Env{
				Kubernetes: fake.NewSimpleClientset(),
			}
			got := make(fstest.MapFS)
			if err := env.RunStarlark(ctx, test.script, string(script), testDumpFS(got)); err != nil {
				t.Errorf("run script: %v", err)
			}
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("produced files (-want +got):\n%s", diff)
			}
		})
	}
}
