package persist

import (
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
)

func TestNewChild(t *testing.T) {
	fullClock := []*Clock{NewClock("master")}
	child := NewChild(fullClock)

	expected := FullClock{
		&Clock{
			Branch: "master",
			Clock:  1,
		},
	}
	require.Equal(t, expected, child)
}

func TestClockRange(t *testing.T) {
	clockRangeList := NewClockRangeList(
		[]*Clock{
			{
				Branch: "master",
				Clock:  1,
			},
		},
		[]*Clock{
			{
				Branch: "master",
				Clock:  3,
			},
			{
				Branch: "foo",
				Clock:  5,
			},
		},
	)
	require.Equal(t, 2, len(clockRangeList.ranges))
	require.Equal(t, &ClockRange{Branch: "master", Left: 2, Right: 3}, clockRangeList.ranges[0])
	require.Equal(t, &ClockRange{Branch: "foo", Left: 0, Right: 5}, clockRangeList.ranges[1])
}

func TestGetClockIntervals(t *testing.T) {
	c, err := StringToClock("master/0")
	require.NoError(t, err)

	_, err = StringToClock("master/2-foo/0")
	require.YesError(t, err)

	_, err = StringToClock("master/2")
	require.NoError(t, err)
	rangeList := NewClockRangeList(
		[]*Clock{
			c,
		},
		[]*Clock{
			{
				Branch: "master",
				Clock:  2,
			},
			{
				Branch: "foo",
				Clock:  0,
			},
		},
	)
	require.Equal(t, 2, len(rangeList.ranges))
	require.Equal(
		t,
		&ClockRange{
			Branch: "master",
			Left:   1,
			Right:  2,
		},
		rangeList.ranges[0],
	)
	require.Equal(
		t,
		&ClockRange{
			Branch: "foo",
			Left:   0,
			Right:  0,
		},
		rangeList.ranges[1],
	)
	//	require.Equal(t, [][]*BranchClock{{b1, e1}, {e2, b2}}, intervals)
	/*
		b1, err = StringToClock("master/2")
		require.NoError(t, err)
		b2, err = StringToClock("master/5")
		require.NoError(t, err)
		intervals, err = GetClockIntervals(b1, b2)
		require.NoError(t, err)
		require.Equal(t, [][]*BranchClock{{b1, b2}}, intervals)

		b1, err = StringToClock("master/1-foo/2")
		require.NoError(t, err)
		b2, err = StringToClock("master/1-foo/5")
		require.NoError(t, err)
		intervals, err = GetClockIntervals(b1, b2)
		require.NoError(t, err)
		require.Equal(t, [][]*BranchClock{{b1, b2}}, intervals)

		b1, err = StringToClock("master/1-foo/2")
		require.NoError(t, err)
		b2, err = StringToClock("master/1-bar/1")
		require.NoError(t, err)
		intervals, err = GetClockIntervals(b1, b2)
		require.YesError(t, err)

		b1, err = StringToClock("master/1-foo/2")
		require.NoError(t, err)
		b2, err = StringToClock("master/0")
		require.NoError(t, err)
		intervals, err = GetClockIntervals(b1, b2)
		require.YesError(t, err)
	*/
}
