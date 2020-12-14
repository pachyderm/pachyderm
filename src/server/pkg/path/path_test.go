package path

import (
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
)

func TestClean(t *testing.T) {
	for input, expected := range map[string]string{
		`/`:          ``,
		`.`:          ``,
		`/.`:         ``,
		`foo/`:       `/foo`,
		`foo/.`:      `/foo`,
		`/foo/`:      `/foo`,
		`/foo/.`:     `/foo`,
		`/foo/..`:    ``,
		`/foo/../..`: ``,
		`/(*)`:       `/(*)`,
		`/(*)/`:      `/(*)`,
	} {
		require.Equal(t, expected, Clean(input), "for %q", input)
	}
}

func TestBase(t *testing.T) {
	for input, expected := range map[string]string{
		`/`:          ``,
		`.`:          ``,
		`foo/`:       `foo`,
		`foo/.`:      `foo`, // pre-refactor, was ""
		`/foo/`:      `foo`,
		`/foo/.`:     `foo`,
		`/foo/..`:    ``,
		`/foo/../..`: ``,
		`/(*)`:       `(*)`,
		`/foo/bar`:   `bar`,
		`/foo/*`:     `*`,
	} {
		require.Equal(t, expected, Base(input), "for %q", input)
	}
}

func TestDir(t *testing.T) {
	for input, expected := range map[string]string{
		`/`:          ``,
		`.`:          ``,
		`foo/`:       ``, // pre-refactor, was "/foo"
		`foo/.`:      ``,
		`/foo/`:      ``,
		`/foo/.`:     ``,
		`/foo/..`:    ``,
		`/foo/../..`: ``,
		`/(*)`:       ``,
		`/foo/bar`:   `/foo`,
		`/foo/*`:     `/foo`,
	} {
		require.Equal(t, expected, Dir(input), "for %q", input)
	}
}

func TestSplit(t *testing.T) {
	for input, expected := range map[string]struct{ base, dir string }{
		`/`:          {``, ``},
		`.`:          {``, ``},
		`foo/`:       {``, `foo`}, // pre-refactor, was {"foo", "foo"}
		`foo/.`:      {``, `foo`}, // pre-refactor, was {"/foo", ""}
		`/foo/`:      {``, `foo`},
		`/foo/.`:     {``, `foo`},
		`/foo/..`:    {``, ``},
		`/foo/../..`: {``, ``},
		`/(*)`:       {``, `(*)`},
		`/foo/bar`:   {`/foo`, `bar`},
		`/foo/*`:     {`/foo`, `*`},
	} {
		b, d := Split(input)
		require.Equal(t, expected, struct{ base, dir string }{b, d}, "for %q", input)
	}
}

func TestIsGlob(t *testing.T) {
	for input, expected := range map[string]bool{
		`/`:           false,
		`.`:           false,
		`/foo/`:       false,
		`/foo/.`:      false,
		`/foo/..`:     false,
		`/foo/../..`:  false,
		`(*)/`:        true,
		`/(*)`:        true,
		`/foo/bar`:    false,
		`/foo/*`:      true,
		`/foo-?/`:     true,
		`/foo-(a|b)/`: true,

		// These are carried over from hashtree/hashtree_test.go
		`*`:                       true,
		`path/to*/file`:           true,
		`path/**/file`:            true,
		`path/to/f?le`:            true,
		`pa!h/to/file`:            true,
		`pa[th]/to/file`:          true,
		`pa{th}/to/file`:          true,
		`*/*`:                     true,
		`path`:                    false,
		`path/to/file1.txt`:       false,
		`path/to_test-a/file.txt`: false,
	} {
		require.Equal(t, expected, IsGlob(input), "for %q", input)
	}
}

func TestGlobLiteralPrefix(t *testing.T) {
	for input, expected := range map[string]string{
		`/`:           ``,
		`.`:           ``,
		`/foo/`:       `/foo`,
		`/foo/.`:      `/foo`,
		`/foo/..`:     ``,
		`/foo/../..`:  ``,
		`/(*)`:        `/`,
		`(*)/`:        `/`,
		`/foo/bar`:    `/foo/bar`,
		`/foo/*`:      `/foo/`,
		`/foo-?/`:     `/foo-`,
		`/foo-(a|b)/`: `/foo-`,

		// These are carried over from hashtree/hashtree_test.go
		`*`:             `/`,
		`**`:            `/`,
		`dir/*`:         `/dir/`,
		`dir/**`:        `/dir/`,
		`dir/(*)`:       `/dir/`,
		`di?/(*)`:       `/di`,
		`di?[rg]`:       `/di`,
		`dir/@(a)`:      `/dir/`,
		`dir/+(a)`:      `/dir/`,
		`dir/{foo,bar}`: `/dir/`,
		`dir/(a|b)`:     `/dir/`,
		`dir/^[a-z]`:    `/dir/`,
		`dir/[!abc]`:    `/dir/`,
	} {
		require.Equal(t, expected, GlobLiteralPrefix(input), "for %q", input)
	}
}

func TestGlobPrefix(t *testing.T) {
}
