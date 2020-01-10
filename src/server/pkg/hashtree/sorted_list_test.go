package hashtree

import (
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
)

func TestInsertStr(t *testing.T) {
	// Don't expand slice
	x := make([]string, 3, 4)
	copy(x, []string{"a", "b", "d"})
	insertStr(&x, "c")
	require.Equal(t, []string{"a", "b", "c", "d"}, x)
	require.Equal(t, 4, len(x))
	require.Equal(t, 4, cap(x))

	// Expand slice by constant amount
	x = make([]string, 3)
	copy(x, []string{"a", "b", "d"})
	insertStr(&x, "c")
	require.Equal(t, []string{"a", "b", "c", "d"}, x)
	require.Equal(t, 4, len(x))
	require.True(t, cap(x) >= 4)

	// Expand slice by factor (may fail if constant grows)
	x = make([]string, 25)
	copy(x, []string{"a", "b", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m",
		"n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"})
	insertStr(&x, "c")
	require.Equal(t, []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j",
		"k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y",
		"z"}, x)
	require.Equal(t, 26, len(x))
	require.True(t, cap(x) >= 26)

	// insert at beginning
	x = make([]string, 3)
	copy(x, []string{"b", "c", "d"})
	insertStr(&x, "a")
	require.Equal(t, []string{"a", "b", "c", "d"}, x)
	require.Equal(t, 4, len(x))
	require.True(t, cap(x) >= 4)

	// insert at end
	x = make([]string, 3)
	copy(x, []string{"a", "b", "c"})
	insertStr(&x, "d")
	require.Equal(t, []string{"a", "b", "c", "d"}, x)
	require.Equal(t, 4, len(x))
	require.True(t, cap(x) >= 4)
}
