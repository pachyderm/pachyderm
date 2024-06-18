package testloki_test

import (
	"context"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/lokiutil/testloki"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachd"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/logs"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestTestPachd(t *testing.T) {
	ctx := pctx.TestContext(t)
	l, err := testloki.New(ctx, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := l.Close(); err != nil {
			t.Fatalf("close loki: %v", err)
		}
	})
	start := time.Now()
	pd := pachd.NewTestPachd(t, testloki.WithTestLoki(l))

	tctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	gls, err := pd.LogsClient.GetLogs(tctx, &logs.GetLogsRequest{
		Query: &logs.LogQuery{
			QueryType: &logs.LogQuery_Admin{
				Admin: &logs.AdminLogQuery{
					AdminType: &logs.AdminLogQuery_Logql{
						Logql: `{host=~ ".+"}`,
					},
				},
			},
		},
		Filter: &logs.LogFilter{
			TimeRange: &logs.TimeRangeLogFilter{
				From:  timestamppb.New(start),
				Until: timestamppb.Now(),
			},
			Limit: 1,
		},
	})
	if err != nil {
		t.Fatalf("GetLogs: %v", err)
	}
	res, err := gls.Recv()
	if err != nil {
		// io.EOF is an error here.
		t.Fatalf("Recv: %v", err)
	}
	t.Log(protojson.Format(res))
}
