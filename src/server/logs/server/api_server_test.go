package server_test

import (
	"fmt"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/logs"

	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/lokiutil"
	loki "github.com/pachyderm/pachyderm/v2/src/internal/lokiutil/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachconfig"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testpachd/realenv"
)

func TestVerbatimRequest(t *testing.T) {
	var (
		ctx          = pctx.TestContext(t)
		buildEntries = func() []loki.Entry {
			var entries []loki.Entry
			for i := -99; i <= 0; i++ {
				entries = append(entries, loki.Entry{
					Timestamp: time.Now().Add(time.Duration(i) * time.Second),
					Line:      fmt.Sprintf("%v", i),
				})
			}
			return entries
		}
		srv = httptest.NewServer(&lokiutil.FakeServer{
			Entries: buildEntries(),
		})
	)
	env := realenv.NewRealEnv(ctx, t,
		dockertestenv.NewTestDBConfig(t).PachConfigOption,
		func(c *pachconfig.Configuration) {
			u, err := url.Parse(srv.URL)
			if err != nil {
				panic(err)
			}
			c.LokiHost, c.LokiPort = u.Hostname(), u.Port()
		})
	logsClient := logs.NewAPIClient(env.PachClient.ClientConn())
	respStream, err := logsClient.GetLogs(ctx, &logs.GetLogsRequest{})
	require.NoError(t, err, "logs.GetLogs request error")
	resp, err := respStream.Recv()
	require.NoError(t, err, "logs.GetLogs stream error")
	t.Log(resp)
}
