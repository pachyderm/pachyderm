package dag

import (
	"testing"

	"github.com/pachyderm/pachyderm/src/pkg/require"
)

func TestLine(t *testing.T) {
	d := NewDAG(map[string][]string{
		"1": {},
		"2": {"1"},
		"3": {"2"},
		"4": {"3"},
	})
	require.Equal(t, []string{"1", "2", "3", "4"}, d.Sorted())
	require.Equal(t, []string{"1", "2", "3", "4"}, d.Ancestors("4", nil))
	require.Equal(t, []string{"1", "2", "3", "4"}, d.Descendants("1", nil))
}
