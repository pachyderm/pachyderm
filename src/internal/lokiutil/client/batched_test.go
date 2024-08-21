package client_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmputil"
	"github.com/pachyderm/pachyderm/v2/src/internal/lokiutil/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/lokiutil/testloki"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

func TestFetchLimitsConfig(t *testing.T) {
	ctx := pctx.TestContext(t)
	l, err := testloki.New(ctx, t.TempDir())
	if err != nil {
		t.Fatalf("testloki.New: %v", err)
	}
	t.Cleanup(func() {
		if err := l.Close(); err != nil {
			t.Fatalf("testloki.Close: %v", err)
		}
	})
	gotRange, gotBatch, err := l.Client.FetchLimitsConfig(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := gotBatch, 5000; got != want {
		t.Errorf("batch size from testloki:\n  got: %v\n want: %v", got, want)
		t.Log("note: batch size is configured in ../testloki/config.yaml")
	}
	if got, want := gotRange, 720*time.Hour; got != want {
		t.Errorf("max range from testloki:\n  got: %v\n want: %v", got, want)
	}
}

func TestBatchedQueryRange(t *testing.T) {
	ctx := pctx.TestContext(t)
	l, err := testloki.New(ctx, t.TempDir())
	if err != nil {
		t.Fatalf("testloki.New: %v", err)
	}
	t.Cleanup(func() {
		if err := l.Close(); err != nil {
			t.Fatalf("testloki.Close: %v", err)
		}
	})
	fh, err := os.Open("testdata/simple.txt")
	if err != nil {
		t.Fatalf("open testdata: %v", err)
	}
	defer fh.Close() //nolint:errcheck
	if err := testloki.AddLogFile(ctx, fh, l); err != nil {
		t.Fatalf("add testdata: %v", err)
	}
	for i := 0; i < 100; i++ {
		if err := l.AddLog(ctx, &testloki.Log{
			Time:    time.Date(2024, 5, 1, 0, 0, 0, 80, time.UTC),
			Labels:  map[string]string{"app": "simple", "suite": "pachyderm"},
			Message: fmt.Sprintf(`{"severity":"debug","time":"2024-05-01T00:00:00.00000008Z","message":"duplicate %d"}`, i),
		}); err != nil {
			t.Fatalf("add duplicate log #%d: %v", i, err)
		}
	}

	testData := []struct {
		name                 string
		start, end           time.Time
		limit                int
		offset               uint
		maxRange             time.Duration
		batchSize            int
		pick                 func(e *client.Entry) bool
		want                 []string
		wantOffset           uint
		wantBatchSizeWarning bool
	}{
		{
			name: "end time before logs start",
			end:  time.Date(2024, 4, 30, 23, 59, 59, 999999999, time.UTC),
		},
		{
			name:  "start time after logs end",
			start: time.Date(2024, 5, 2, 0, 0, 0, 0, time.UTC),
		},
		{
			name:  "start of logs, forward",
			start: time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC),
			end:   time.Date(2024, 5, 1, 0, 0, 0, 4, time.UTC),
			want: []string{
				`/"message":"0"/`,
				`/"message":"1"/`,
				`/"message":"2"/`,
				`/"message":"3"/`,
				`/"message":"4"/`,
			},
		},
		{
			name:  "start of logs, limited",
			start: time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC),
			limit: 5,
			want: []string{
				`/"message":"0"/`,
				`/"message":"1"/`,
				`/"message":"2"/`,
				`/"message":"3"/`,
				`/"message":"4"/`,
			},
		},
		{
			name:      "start of logs, small batch, limited",
			start:     time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC),
			batchSize: 5,
			limit:     10,
			want: []string{
				`/"message":"0"/`,
				`/"message":"1"/`,
				`/"message":"2"/`,
				`/"message":"3"/`,
				`/"message":"4"/`,
				`/"message":"5"/`,
				`/"message":"6"/`,
				`/"message":"7"/`,
				`/"message":"8"/`,
				`/"message":"9"/`,
			},
			// Because the limit was hit at the end of the batch, we don't know if there
			// are more logs or not.
			wantOffset: 1,
		},
		{
			name:  "start of logs, backward",
			end:   time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC),
			start: time.Date(2024, 5, 1, 0, 0, 0, 4, time.UTC),
			want: []string{
				`/"message":"4"/`,
				`/"message":"3"/`,
				`/"message":"2"/`,
				`/"message":"1"/`,
				`/"message":"0"/`,
			},
		},
		{
			name:      "backwards, small batch",
			end:       time.Date(2024, 5, 1, 0, 0, 0, 10, time.UTC),
			batchSize: 5,
			want: []string{
				`/"message":"10"/`,
				`/"message":"9"/`,
				`/"message":"8"/`,
				`/"message":"7"/`,
				`/"message":"6"/`,
				`/"message":"5"/`,
				`/"message":"4"/`,
				`/"message":"3"/`,
				`/"message":"2"/`,
				`/"message":"1"/`,
				`/"message":"0"/`,
			},
		},
		{
			name:      "backwards, small batch, limited",
			end:       time.Date(2024, 5, 1, 0, 0, 0, 9, time.UTC),
			batchSize: 5,
			limit:     6,
			want: []string{
				`/"message":"9"/`,
				`/"message":"8"/`,
				`/"message":"7"/`,
				`/"message":"6"/`,
				`/"message":"5"/`,
				`/"message":"4"/`,
			},
		},
		{
			name:  "end of logs, backwards",
			end:   time.Date(2024, 5, 2, 0, 0, 0, 0, time.UTC),
			limit: 5,
			want: []string{
				`/"message":"999"/`,
				`/"message":"998"/`,
				`/"message":"997"/`,
				`/"message":"996"/`,
				`/"message":"995"/`,
			},
		},
		{
			name:  "middle of logs, forward, hitting a long string of duplicate timestamps",
			start: time.Date(2024, 5, 1, 0, 0, 0, 79, time.UTC),
			limit: 5,
			want: []string{
				`/"message":"79"/`,
				`/"message":"80"/`,          // next log has offset = 1
				`/"message":"duplicate 0"/`, // offset = 2
				`/"message":"duplicate 1"/`, // offset = 3
				`/"message":"duplicate 2"/`, // offset = 4
			},
			wantOffset: 4,
		},
		{
			name:   "offset into a long string of duplicate timestamps",
			start:  time.Date(2024, 5, 1, 0, 0, 0, 80, time.UTC),
			offset: 4,
			limit:  5,
			want: []string{
				`/"message":"duplicate 3"/`, // offset = 5
				`/"message":"duplicate 4"/`,
				`/"message":"duplicate 5"/`,
				`/"message":"duplicate 6"/`,
				`/"message":"duplicate 7"/`,
			},
			wantOffset: 9,
		},
		{
			name:      "offset into a long string of duplicate timestamps, limited by batch size",
			start:     time.Date(2024, 5, 1, 0, 0, 0, 80, time.UTC),
			end:       time.Date(2024, 5, 1, 0, 0, 0, 82, time.UTC),
			offset:    4,
			batchSize: 6,
			want: []string{
				`/"message":"duplicate 3"/`,
				`/"message":"duplicate 4"/`,
				`/"message":"81"/`,
				`/"message":"82"/`,
			},
			wantBatchSizeWarning: true,
		},
		{
			name:      "offset at a long string of duplicate timestamps, limited by batch size",
			start:     time.Date(2024, 5, 1, 0, 0, 0, 80, time.UTC),
			end:       time.Date(2024, 5, 1, 0, 0, 0, 80, time.UTC),
			offset:    4,
			batchSize: 6,
			want: []string{
				`/"message":"duplicate 3"/`,
				`/"message":"duplicate 4"/`,
			},
			wantBatchSizeWarning: true,
		},
		{
			name:  "middle of logs, backward, into a long string of duplicate timestamps",
			end:   time.Date(2024, 5, 1, 0, 0, 0, 81, time.UTC),
			limit: 5,
			want: []string{
				`/"message":"81"/`,
				`/"message":"duplicate 99"/`, // next log has offset = 1
				`/"message":"duplicate 98"/`, // offset = 2
				`/"message":"duplicate 97"/`, // offset = 3
				`/"message":"duplicate 96"/`, // offset = 4
			},
			wantOffset: 4,
		},
		{
			name:   "middle of logs, backward, with offset",
			start:  time.Date(2024, 5, 1, 0, 0, 0, 80, time.UTC),
			end:    time.Date(2024, 5, 1, 0, 0, 0, 79, time.UTC),
			offset: 4,
			limit:  1,
			want: []string{
				`/"message":"duplicate 95"/`, // offset = 5
			},
			wantOffset: 5,
		},
		{
			name:      "backward offset into a long string of duplicate timestamps, limited by batch size",
			start:     time.Date(2024, 5, 1, 0, 0, 0, 81, time.UTC),
			end:       time.Date(2024, 5, 1, 0, 0, 0, 79, time.UTC),
			batchSize: 2,
			want: []string{
				`/"message":"81"/`,           // start of first QueryRange
				`/"message":"duplicate 99"/`, // end of first QueryRange
				`/"message":"duplicate 98"/`, // queryOneNanosecond is called, skips 99, returns 98, and hits the limit
				`/"message":"79"/`,           // last matching entry (based on end time)
			},
			wantBatchSizeWarning: true,
		},
		{
			name:      "backward offset into a long string of duplicate timestamps, limited by batch size of 1",
			start:     time.Date(2024, 5, 1, 0, 0, 0, 81, time.UTC),
			end:       time.Date(2024, 5, 1, 0, 0, 0, 79, time.UTC),
			batchSize: 1,
			want: []string{
				`/"message":"81"/`,
				`/"message":"duplicate 99"/`,
				`/"message":"79"/`,
			},
			wantBatchSizeWarning: true,
		},
		{
			name:      "backwards offset into a long string of duplicate timestamps, limited by limit",
			start:     time.Date(2024, 5, 1, 0, 0, 0, 82, time.UTC),
			end:       time.Date(2024, 5, 1, 0, 0, 0, 80, time.UTC),
			batchSize: 3,
			limit:     4,
			want: []string{
				`/"message":"82"/`,
				`/"message":"81"/`,
				`/"message":"duplicate 99"/`, // next log has offset=1
				`/"message":"duplicate 98"/`, // next log has offset=2
			},
			wantOffset: 2,
		},
		{
			name:      "continuation of above",
			end:       time.Date(2024, 5, 1, 0, 0, 0, 80, time.UTC),
			offset:    2,
			batchSize: 3,
			limit:     2,
			want: []string{
				`/"message":"duplicate 97"/`,
				`/"message":"79"/`,
			},
			wantBatchSizeWarning: true,
		},
		{
			name:   "offset beyond the end",
			start:  time.Date(2024, 5, 1, 0, 0, 0, 80, time.UTC),
			offset: 1000,
		},
		{
			name:                 "offset beyond the end, limited by batch size",
			start:                time.Date(2024, 5, 1, 0, 0, 0, 80, time.UTC),
			batchSize:            10,
			offset:               1000,
			wantBatchSizeWarning: true,
		},
		{
			name:   "offset beyond the end, with limit",
			start:  time.Date(2024, 5, 1, 0, 0, 0, 80, time.UTC),
			limit:  10,
			offset: 1000,
		},
		{
			name:                 "offset beyond the end, limited by batch size and limit",
			start:                time.Date(2024, 5, 1, 0, 0, 0, 80, time.UTC),
			batchSize:            10,
			limit:                10,
			offset:               1000,
			wantBatchSizeWarning: true,
		},
		{
			name:  "middle of logs, forward, hitting the first duplicate timestamp",
			start: time.Date(2024, 5, 1, 0, 0, 0, 79, time.UTC),
			limit: 2,
			want: []string{
				`/"message":"79"/`,
				`/"message":"80"/`, // offset = 1
			},
			wantOffset: 1,
		},
		{
			name:      "middle of logs, forward, no duplicate",
			start:     time.Date(2024, 5, 1, 0, 0, 0, 78, time.UTC),
			batchSize: 3,
			limit:     2,
			want: []string{
				`/"message":"78"/`,
				`/"message":"79"/`,
			},
		},
		{
			name:      "middle of logs, forward, no duplicate, but we can't know",
			start:     time.Date(2024, 5, 1, 0, 0, 0, 78, time.UTC),
			batchSize: 2,
			limit:     2,
			want: []string{
				`/"message":"78"/`,
				`/"message":"79"/`,
			},
			wantOffset: 1,
		},
		{
			name:      "middle of logs, forward, no limit",
			start:     time.Date(2024, 5, 1, 0, 0, 0, 998, time.UTC),
			batchSize: 100,
			limit:     0,
			want: []string{
				`/"message":"998"/`,
				`/"message":"999"/`,
			},
			wantOffset: 0,
		},
	}
	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			ctx := pctx.TestContext(t)
			client.BatchSizeWarning.Store(false)
			if x := test.batchSize; x > 0 {
				t.Log("overriding batch size")
				client.BatchSizeForTesting = x
				t.Cleanup(func() {
					t.Log("restoring batch size")
					client.BatchSizeForTesting = 0
				})
			}
			if x := test.maxRange; x > 0 {
				t.Log("overriding max range")
				client.MaxRangeForTesting = x
				t.Cleanup(func() {
					t.Log("restoring max range")
					client.MaxRangeForTesting = 0
				})
			}
			var got []string
			_, _, prevOffset, nextOffset, err := l.Client.BatchedQueryRange(ctx, `{app="simple"}`, test.start, test.end, test.offset, test.limit,
				func(ctx context.Context, labels client.LabelSet, e *client.Entry) (count bool, retErr error) {
					if pick := test.pick; pick != nil {
						if keep := pick(e); keep {
							got = append(got, e.Line)
							return true, nil
						}
						return false, nil
					}
					got = append(got, e.Line)
					return true, nil
				})
			if err != nil {
				t.Errorf("BatchedQueryRange: %v", err)
			}
			if got, want := max(prevOffset, nextOffset), test.wantOffset; got != want {
				t.Errorf("offset:\n  got: %v\n want: %v", got, want)
			}
			if got, want := client.BatchSizeWarning.Load(), test.wantBatchSizeWarning; got != want {
				t.Errorf("batch size warning:\n  got: %v\n want: %v", got, want)
			}
			require.NoDiff(t, test.want, got, []cmp.Option{cmputil.RegexpStrings()})
		})
	}
}
