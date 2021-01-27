package dag

import (
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
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
	require.Equal(t, 0, len(d.Ghosts()))
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
			[]string{"2", "3", "4"},
			[]string{"3", "2", "4"},
		},
		d.Ancestors("4", []string{"1"}),
	)
	require.EqualOneOf(
		t,
		[]interface{}{
			[]string{"1", "2", "3"},
			[]string{"1", "3", "2"},
		},
		d.Descendants("1", []string{"4"}),
	)
	require.Equal(t, []string{"4"}, d.Leaves())
	require.Equal(t, 0, len(d.Ghosts()))
}

func TestGhosts(t *testing.T) {
	d := NewDAG(map[string][]string{
		"1": {},
		"2": {"1", "3"},
		"3": {"4", "1"},
		"5": {"4", "6"},
	})
	require.EqualOneOf(
		t,
		[]interface{}{
			[]string{"4", "6"},
			[]string{"6", "4"},
		},
		d.Ghosts(),
	)
}
