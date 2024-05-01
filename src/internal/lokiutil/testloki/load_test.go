package testloki_test

import (
	"strings"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/lokiutil/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/lokiutil/testloki"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

func TestAddLogFile(t *testing.T) {
	ctx := pctx.TestContext(t)
	loki, err := testloki.New(ctx, t.TempDir())
	if err != nil {
		t.Fatalf("new testloki: %v", err)
	}
	t.Cleanup(func() {
		if err := loki.Close(); err != nil {
			t.Fatalf("close: %v", err)
		}
	})
	in := `&map[key:value]

> @pachyderm/dash-backend@0.0.1 start
{"time":"2022-01-02T00:00:00.000Z","message":"message 1"}
2022-01-02 00:00:00.123 UTC [1] LOG message 2
[2022-01-02 00:00:01.003] message 3
{"ts":1641081602.4,"message":"message 4"}
`
	if err := testloki.AddLogFile(ctx, strings.NewReader(in), loki); err != nil {
		t.Fatalf("add log file: %v", err)
	}
	got, err := loki.Client.QueryRange(ctx, `{key="value"}`, 1000, time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2022, 1, 3, 0, 0, 0, 0, time.UTC), "FORWARD", 0, 0, true)
	if err != nil {
		t.Fatalf("query logs: %v", err)
	}
	want := &client.QueryResponse{
		Status: "success",
		Data: client.QueryResponseData{
			ResultType: "streams",
			Result: client.Streams{
				{
					Labels: map[string]string{"key": "value"},
					Entries: []client.Entry{
						{
							Timestamp: time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC),
							Line:      `{"time":"2022-01-02T00:00:00.000Z","message":"message 1"}`,
						},
						{
							Timestamp: time.Date(2022, 1, 2, 0, 0, 0, 123000000, time.UTC),
							Line:      `2022-01-02 00:00:00.123 UTC [1] LOG message 2`,
						},
						{
							Timestamp: time.Date(2022, 1, 2, 0, 0, 1, 3000000, time.UTC),
							Line:      "[2022-01-02 00:00:01.003] message 3",
						},
						{
							Timestamp: time.Date(2022, 1, 2, 0, 0, 2, 400000000, time.UTC),
							Line:      `{"ts":1641081602.4,"message":"message 4"}`,
						},
					},
				},
			},
		},
	}
	require.NoDiff(t, want, got, nil)
}
