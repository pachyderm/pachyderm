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
	require.Equal(t, []string{"3", "4"}, d.Ancestors("4", []string{"2"}))
	require.Equal(t, []string{"1", "2"}, d.Ancestors("2", nil))
	require.Equal(t, []string{"1", "2", "3", "4"}, d.Descendants("1", nil))
	require.Equal(t, []string{"3", "4"}, d.Descendants("3", nil))
	require.Equal(t, []string{"1", "2"}, d.Descendants("1", []string{"3"}))
	require.Equal(t, []string{"4"}, d.Leaves())
}

func TestDiamond(t *testing.T) {
	d := NewDAG(map[string][]string{
		"1": {},
		"2": {"1"},
		"3": {"1"},
		"4": {"2", "3"},
	})
	require.EqualOneOf(
		t,
		[]interface{}{
			[]string{"1", "2", "3", "4"},
			[]string{"1", "3", "2", "4"},
		},
		d.Sorted(),
	)
	require.EqualOneOf(
		t,
		[]interface{}{
			[]string{"1", "2", "3", "4"},
			[]string{"1", "3", "2", "4"},
		},
		d.Ancestors("4", nil),
	)
	require.EqualOneOf(
		t,
		[]interface{}{
			[]string{"1", "2", "3", "4"},
			[]string{"1", "3", "2", "4"},
		},
		d.Descendants("1", nil),
	)
}
