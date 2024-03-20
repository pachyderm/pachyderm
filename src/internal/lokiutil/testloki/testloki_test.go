package testloki

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/pachyderm/pachyderm/v2/src/internal/lokiutil/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
)

func TestTestLoki(t *testing.T) {
	ctx := pctx.TestContext(t)
	loki, err := New(ctx, t.TempDir())
	if err != nil {
		t.Fatalf("set up loki: %v", err)
	}
	t.Cleanup(func() {
		if err := loki.Close(); err != nil {
			t.Fatalf("clean up loki: %v", err)
		}
	})
	now := time.Now()
	if err := loki.AddLog(ctx, &Log{
		Time:    now,
		Message: "hello, loki",
		Labels:  map[string]string{"app": "the_tests", "suite": "pachyderm"},
	}); err != nil {
		t.Fatalf("add log: %v", err)
	}
	want := &client.QueryResponse{
		Status: "success",
		Data: client.QueryResponseData{
			ResultType: "streams",
			Result: client.Streams{
				{
					Labels: client.LabelSet{
						"app":   "the_tests",
						"suite": "pachyderm",
					},
					Entries: []client.Entry{
						{
							Timestamp: now.Truncate(time.Nanosecond),
							Line:      "hello, loki",
						},
					},
				},
			},
		},
	}
	ctx, c := context.WithTimeout(ctx, 5*time.Second)
	got, err := loki.Client.QueryRange(ctx, `{suite="pachyderm"}`, 1, now.Add(-time.Hour), now.Add(time.Hour), "BACKWARD", 0, 0, true)
	c()
	if err != nil {
		t.Fatalf("query range: %v", err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("response (-want +got):\n%s", diff)
	}
}
