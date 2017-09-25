package match

import (
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
)

func TestMatch(t *testing.T) {
	shouldMatch(t, "**/*", "a/b/c")
	shouldMatch(t, "/**/*", "a/b/c")
	shouldMatch(t, "/**/*", "/a/b/c")
	shouldMatch(t, "**/*", "/a/b/c")
	shouldMatch(t, "**/*", "a")
	shouldMatch(t, "a/**/*", "/a/b/c/d")
	shouldMatch(t, "a/**/d", "/a/b/c/d")
	shouldMatch(t, "a/**/c/d", "/a/b/c/d")
	shouldMatch(t, "a/**/b/c/d", "/a/b/c/d")
	shouldMatch(t, "**/a/b/c/d", "/a/b/c/d")
	shouldMatch(t, "**/*/b", "a/b")
	shouldMatch(t, "*/**/b", "a/b")

	shouldNotMatch(t, "*/**/*/b", "a/b")

	shouldErr(t, "**/**", "a/b/c")
	shouldErr(t, "***", "a/b/c")
	shouldErr(t, "*****", "a/b/c")
}

func shouldMatch(t *testing.T, pattern string, name string) {
	t.Helper()
	matched, err := Match(pattern, name)
	require.NoError(t, err)
	require.True(t, matched)
}

func shouldNotMatch(t *testing.T, pattern string, name string) {
	t.Helper()
	matched, err := Match(pattern, name)
	require.NoError(t, err)
	require.False(t, matched)
}

func shouldErr(t *testing.T, pattern string, name string) {
	t.Helper()
	matched, err := Match(pattern, name)
	require.YesError(t, err)
	require.False(t, matched)
}
