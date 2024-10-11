package bazel

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestLibraryPath(t *testing.T) {
	testData := []struct {
		name       string
		rlocations []string
		environ    []string
		want       string
	}{
		{
			name: "empty",
			want: "",
		},
		{
			name:    "no libraries",
			environ: []string{"FOO=bar", "BAR=baz"},
			want:    "",
		},
		{
			name:    "passthrough library path",
			environ: []string{"FOO=bar", "LD_LIBRARY_PATH=/lib", "BAR=baz"},
			want:    "/lib",
		},
		{
			name:       "create new variable",
			environ:    []string{"FOO=bar", "BAR=baz"},
			rlocations: []string{"/foo/libpq.so.5.17"},
			want:       "/foo",
		},
		{
			name:       "prepend to existing",
			environ:    []string{"FOO=bar", "BAR=baz", "LD_LIBRARY_PATH=/lib:/usr/lib"},
			rlocations: []string{"/foo/libpq.so.5.17"},
			want:       "/foo:/lib:/usr/lib",
		},
		{
			name:       "prepend multiple to existing",
			environ:    []string{"FOO=bar", "BAR=baz", "LD_LIBRARY_PATH=/lib:/usr/lib"},
			rlocations: []string{"/foo/libpq.so.5.17", "/bar/libldap.so.2", ""},
			want:       "/bar:/foo:/lib:/usr/lib",
		},
	}
	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			got := LibraryPath(test.environ, test.rlocations...)
			if diff := cmp.Diff(test.want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("LibraryPath (-want +got):\n%s", diff)
			}
		})
	}
}
