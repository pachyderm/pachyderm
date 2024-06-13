package client

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

type timeRange [2]time.Time

const maxNanoseconds = 999_999_999

func TestTimeIterator(t *testing.T) {
	testData := []struct {
		name     string
		iterator *TimeIterator
		limit    int
		logs     []time.Time
		// Note about want; want uses Loki's time conventions (where end is exclusive), but
		// we use inclusive start and end times.  So if you see tests that look like they're
		// off by a nanosecond, that's what's going on there.
		want                              []timeRange
		wantBackwardHint, wantForwardHint time.Time
	}{
		{
			name: "exactly one nanosecond",
			iterator: &TimeIterator{
				Start: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
				Step:  24 * time.Hour,
			},
			want: []timeRange{
				{time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2020, 1, 1, 0, 0, 0, 1, time.UTC)},
			},
			wantBackwardHint: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).Add(-time.Nanosecond),
			wantForwardHint:  time.Date(2020, 1, 1, 0, 0, 0, 1, time.UTC),
		},
		{
			name: "bounded forward traversal",
			iterator: &TimeIterator{
				Start: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
				Step:  24 * time.Hour,
			},
			want: []timeRange{
				{time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)},
				{time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC), time.Date(2020, 1, 2, 0, 0, 0, 1, time.UTC)},
			},
			wantBackwardHint: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).Add(-time.Nanosecond),
			wantForwardHint:  time.Date(2020, 1, 2, 0, 0, 0, 1, time.UTC),
		},
		{
			name: "bounded forward traversal without last nanosecond",
			iterator: &TimeIterator{
				Start: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2020, 1, 1, 23, 59, 59, maxNanoseconds, time.UTC),
				Step:  24 * time.Hour,
			},
			want: []timeRange{
				{time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)},
			},
			wantBackwardHint: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).Add(-time.Nanosecond),
			wantForwardHint:  time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "bounded slower forward traversal",
			iterator: &TimeIterator{
				Start: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
				Step:  6 * time.Hour,
			},
			want: []timeRange{
				{time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2020, 1, 1, 6, 0, 0, 0, time.UTC)},
				{time.Date(2020, 1, 1, 6, 0, 0, 0, time.UTC), time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC)},
				{time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC), time.Date(2020, 1, 1, 18, 0, 0, 0, time.UTC)},
				{time.Date(2020, 1, 1, 18, 0, 0, 0, time.UTC), time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)},
				{time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC), time.Date(2020, 1, 2, 0, 0, 0, 1, time.UTC)},
			},
			wantBackwardHint: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).Add(-time.Nanosecond),
			wantForwardHint:  time.Date(2020, 1, 2, 0, 0, 0, 1, time.UTC),
		},
		{
			name: "bounded slower forward traversal with weird interval",
			iterator: &TimeIterator{
				Start: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
				Step:  9*time.Hour - time.Nanosecond,
			},
			want: []timeRange{
				{time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2020, 1, 1, 8, 59, 59, maxNanoseconds, time.UTC)},
				{time.Date(2020, 1, 1, 8, 59, 59, maxNanoseconds, time.UTC),
					time.Date(2020, 1, 1, 17, 59, 59, maxNanoseconds-1, time.UTC)},
				{time.Date(2020, 1, 1, 17, 59, 59, maxNanoseconds-1, time.UTC),
					time.Date(2020, 1, 2, 0, 0, 0, 1, time.UTC)},
			},
			wantBackwardHint: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).Add(-time.Nanosecond),
			wantForwardHint:  time.Date(2020, 1, 2, 0, 0, 0, 1, time.UTC),
		},
		{
			name: "forward traversal from before the start of time",
			iterator: &TimeIterator{
				Start: time.Date(2019, 12, 31, 23, 0, 0, 0, time.UTC),
				End:   time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
				Step:  24 * time.Hour,
			},
			want: []timeRange{
				{time.Date(2019, 12, 31, 23, 0, 0, 0, time.UTC), time.Date(2020, 1, 1, 23, 0, 0, 0, time.UTC)},
				{time.Date(2020, 1, 1, 23, 0, 0, 0, time.UTC), time.Date(2020, 1, 2, 0, 0, 0, 1, time.UTC)},
			},
			wantBackwardHint: time.Date(2019, 12, 31, 22, 59, 59, maxNanoseconds, time.UTC),
			wantForwardHint:  time.Date(2020, 1, 2, 0, 0, 0, 1, time.UTC),
		},
		{
			name: "forward traversal to the end of time",
			iterator: &TimeIterator{
				Start: time.Date(2020, 1, 1, 6, 0, 0, 0, time.UTC),
				Step:  24 * time.Hour,
				now:   time.Date(2020, 1, 2, 23, 59, 59, 0, time.UTC),
			},
			want: []timeRange{
				{time.Date(2020, 1, 1, 6, 0, 0, 0, time.UTC), time.Date(2020, 1, 2, 6, 0, 0, 0, time.UTC)},
				{time.Date(2020, 1, 2, 6, 0, 0, 0, time.UTC), time.Date(2020, 1, 2, 23, 59, 59, 0, time.UTC)},
			},
			wantBackwardHint: time.Date(2020, 1, 1, 5, 59, 59, maxNanoseconds, time.UTC),
			wantForwardHint:  time.Date(2020, 1, 2, 23, 59, 59, 0, time.UTC),
		},
		{
			name: "nanosecond forward traversal",
			iterator: &TimeIterator{
				Start: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2020, 1, 1, 0, 0, 0, 3, time.UTC),
				Step:  time.Nanosecond,
			},
			want: []timeRange{
				{time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2020, 1, 1, 0, 0, 0, 1, time.UTC)},
				{time.Date(2020, 1, 1, 0, 0, 0, 1, time.UTC), time.Date(2020, 1, 1, 0, 0, 0, 2, time.UTC)},
				{time.Date(2020, 1, 1, 0, 0, 0, 2, time.UTC), time.Date(2020, 1, 1, 0, 0, 0, 3, time.UTC)},
				{time.Date(2020, 1, 1, 0, 0, 0, 3, time.UTC), time.Date(2020, 1, 1, 0, 0, 0, 4, time.UTC)},
			},
			wantBackwardHint: time.Date(2019, 12, 31, 23, 59, 59, maxNanoseconds, time.UTC),
			wantForwardHint:  time.Date(2020, 1, 1, 0, 0, 0, 4, time.UTC),
		},
		{
			name: "backward traversal",
			iterator: &TimeIterator{
				Start: time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
				Step:  24 * time.Hour,
			},
			want: []timeRange{
				{time.Date(2020, 1, 2, 0, 0, 0, 1, time.UTC), time.Date(2020, 1, 3, 0, 0, 0, 1, time.UTC)},
				{time.Date(2020, 1, 1, 0, 0, 0, 1, time.UTC), time.Date(2020, 1, 2, 0, 0, 0, 1, time.UTC)},
				{time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2020, 1, 1, 0, 0, 0, 1, time.UTC)},
			},
			wantBackwardHint: time.Date(2019, 12, 31, 23, 59, 59, maxNanoseconds, time.UTC),
			wantForwardHint:  time.Date(2020, 1, 3, 0, 0, 0, 1, time.UTC),
		},
		{
			name: "nanosecond backward traversal",
			iterator: &TimeIterator{
				Start: time.Date(2020, 1, 1, 0, 0, 0, 2, time.UTC),
				End:   time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			want: []timeRange{
				{time.Date(2020, 1, 1, 0, 0, 0, 2, time.UTC), time.Date(2020, 1, 1, 0, 0, 0, 3, time.UTC)},
				{time.Date(2020, 1, 1, 0, 0, 0, 1, time.UTC), time.Date(2020, 1, 1, 0, 0, 0, 2, time.UTC)},
				{time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2020, 1, 1, 0, 0, 0, 1, time.UTC)},
			},
			wantBackwardHint: time.Date(2019, 12, 31, 23, 59, 59, maxNanoseconds, time.UTC),
			wantForwardHint:  time.Date(2020, 1, 1, 0, 0, 0, 3, time.UTC),
		},
		{
			name: "backwards traversal towards start of time",
			iterator: &TimeIterator{
				End:  time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
				Step: 24 * time.Hour,
			},
			want: []timeRange{
				{time.Date(2020, 1, 1, 0, 0, 0, 1, time.UTC), time.Date(2020, 1, 2, 0, 0, 0, 1, time.UTC)},
				{time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2020, 1, 1, 0, 0, 0, 1, time.UTC)},
			},
			wantBackwardHint: time.Date(2019, 12, 31, 23, 59, 59, maxNanoseconds, time.UTC),
			wantForwardHint:  time.Date(2020, 1, 2, 0, 0, 0, 1, time.UTC),
		},
		{
			name:  "bounded forward traversal with limits",
			limit: 1,
			logs: []time.Time{
				time.Date(2020, 1, 1, 9, 0, 0, 0, time.UTC),
			},
			iterator: &TimeIterator{
				Start: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
				Step:  24 * time.Hour,
			},
			want: []timeRange{
				{time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)},
				{time.Date(2020, 1, 1, 9, 0, 0, 1, time.UTC), time.Date(2020, 1, 2, 0, 0, 0, 1, time.UTC)},
			},
			wantBackwardHint: time.Date(2019, 12, 31, 23, 59, 59, maxNanoseconds, time.UTC),
			wantForwardHint:  time.Date(2020, 1, 2, 0, 0, 0, 1, time.UTC),
		},
		{
			name:  "backward traversal with limits",
			limit: 1,
			logs: []time.Time{
				time.Date(2020, 1, 1, 9, 0, 0, 0, time.UTC),
			},
			iterator: &TimeIterator{
				Start: time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
				Step:  24 * time.Hour,
			},
			want: []timeRange{
				{time.Date(2020, 1, 2, 0, 0, 0, 1, time.UTC), time.Date(2020, 1, 3, 0, 0, 0, 1, time.UTC)},
				{time.Date(2020, 1, 1, 0, 0, 0, 1, time.UTC), time.Date(2020, 1, 2, 0, 0, 0, 1, time.UTC)},
				{time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2020, 1, 1, 9, 0, 0, 0, time.UTC)},
			},
			wantBackwardHint: time.Date(2019, 12, 31, 23, 59, 59, maxNanoseconds, time.UTC),
			wantForwardHint:  time.Date(2020, 1, 3, 0, 0, 0, 1, time.UTC),
		},
		{
			name:  "backward traversal with larger limit", // limits do not affect traversal
			limit: 2,
			logs: []time.Time{
				time.Date(2020, 1, 1, 9, 0, 0, 0, time.UTC),
			},
			iterator: &TimeIterator{
				Start: time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
				Step:  24 * time.Hour,
			},
			want: []timeRange{
				{time.Date(2020, 1, 2, 0, 0, 0, 1, time.UTC), time.Date(2020, 1, 3, 0, 0, 0, 1, time.UTC)},
				{time.Date(2020, 1, 1, 0, 0, 0, 1, time.UTC), time.Date(2020, 1, 2, 0, 0, 0, 1, time.UTC)},
				{time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2020, 1, 1, 0, 0, 0, 1, time.UTC)},
			},
			wantBackwardHint: time.Date(2019, 12, 31, 23, 59, 59, maxNanoseconds, time.UTC),
			wantForwardHint:  time.Date(2020, 1, 3, 0, 0, 0, 1, time.UTC),
		},
		{
			name:  "backward traversal with larger limit that affects traversal",
			limit: 2,
			logs: []time.Time{
				time.Date(2020, 1, 1, 10, 0, 0, 0, time.UTC),
				time.Date(2020, 1, 1, 9, 0, 0, 0, time.UTC),
			},
			iterator: &TimeIterator{
				Start: time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
				Step:  24 * time.Hour,
			},
			want: []timeRange{
				{time.Date(2020, 1, 2, 0, 0, 0, 1, time.UTC), time.Date(2020, 1, 3, 0, 0, 0, 1, time.UTC)},
				{time.Date(2020, 1, 1, 0, 0, 0, 1, time.UTC), time.Date(2020, 1, 2, 0, 0, 0, 1, time.UTC)},
				{time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2020, 1, 1, 9, 0, 0, 0, time.UTC)},
			},
			wantBackwardHint: time.Date(2019, 12, 31, 23, 59, 59, maxNanoseconds, time.UTC),
			wantForwardHint:  time.Date(2020, 1, 3, 0, 0, 0, 1, time.UTC),
		},
		{
			name:  "unbounded backward traversal with larger limit that affects traversal",
			limit: 2,
			logs: []time.Time{
				time.Date(2020, 1, 1, 10, 0, 0, 0, time.UTC),
				time.Date(2020, 1, 1, 9, 0, 0, 0, time.UTC),
			},
			iterator: &TimeIterator{
				End:  time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC),
				Step: 24 * time.Hour,
			},
			want: []timeRange{
				{time.Date(2020, 1, 2, 0, 0, 0, 1, time.UTC), time.Date(2020, 1, 3, 0, 0, 0, 1, time.UTC)},
				{time.Date(2020, 1, 1, 0, 0, 0, 1, time.UTC), time.Date(2020, 1, 2, 0, 0, 0, 1, time.UTC)},
				{time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2020, 1, 1, 9, 0, 0, 0, time.UTC)},
			},
			wantBackwardHint: time.Date(2019, 12, 31, 23, 59, 59, maxNanoseconds, time.UTC),
			wantForwardHint:  time.Date(2020, 1, 3, 0, 0, 0, 1, time.UTC),
		},
		{
			name:  "defaults",
			limit: 1,
			logs: []time.Time{
				time.Date(2024, 6, 1, 0, 0, 0, 123, time.UTC),
				time.Date(2024, 6, 1, 0, 0, 0, 124, time.UTC),
			},
			iterator: &TimeIterator{
				Step: 365 * 24 * time.Hour,
				now:  time.Date(2024, 6, 4, 0, 0, 0, 0, time.UTC),
			},
			want: []timeRange{
				{time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2020, 12, 31, 0, 0, 0, 0, time.UTC)},
				{time.Date(2020, 12, 31, 0, 0, 0, 0, time.UTC), time.Date(2021, 12, 31, 0, 0, 0, 0, time.UTC)},
				{time.Date(2021, 12, 31, 0, 0, 0, 0, time.UTC), time.Date(2022, 12, 31, 0, 0, 0, 0, time.UTC)},
				{time.Date(2022, 12, 31, 0, 0, 0, 0, time.UTC), time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC)},
				{time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC), time.Date(2024, 6, 4, 0, 0, 0, 0, time.UTC)},
				{time.Date(2024, 6, 1, 0, 0, 0, 124, time.UTC), time.Date(2024, 6, 4, 0, 0, 0, 0, time.UTC)},
			},
			wantBackwardHint: time.Date(2019, 12, 31, 23, 59, 59, maxNanoseconds, time.UTC),
			wantForwardHint:  time.Date(2024, 6, 4, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			var got []timeRange
			var logsEverUsed bool
			n := len(test.want) + 5
			limit := test.limit
			for test.iterator.Next() {
				s, e := test.iterator.Interval()
				got = append(got, timeRange{s, e})
				for _, l := range test.logs { // test author must sort test.logs in traversal order.
					if !l.Before(s) && l.Before(e) { // ( !(l < s) && l < e ) -> (s <= l < e)
						logsEverUsed = true
						limit--
						if limit == 0 {
							test.iterator.ObserveLast(l)
							break
						}
					}
				}
				n--
				if n < 0 {
					t.Error("iterator may be out of control; bailing out")
					break
				}
			}

			// some checks to ensure that the test is working like the author expects
			if len(test.logs) > 0 && !logsEverUsed {
				t.Error("no test logs were ever consumed")
			}
			if test.limit > 0 && limit == test.limit {
				t.Error("the provided limit was never used")
			}

			// what we actually care about
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("time ranges (-expected +actual):\n%s", diff)
			}
			backwardEnd := test.iterator.BackwardHint()
			if diff := cmp.Diff(test.wantBackwardHint, backwardEnd); diff != "" {
				t.Errorf("backward hint (-expected +actual):\n%s", diff)
			}
			forwardStart := test.iterator.ForwardHint()
			if diff := cmp.Diff(test.wantForwardHint, forwardStart); diff != "" {
				t.Errorf("forward hint (-expected +actual):\n%s", diff)
			}
		})
	}
}
