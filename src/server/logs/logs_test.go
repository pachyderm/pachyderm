package logs_test

import (
	"context"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/logs"
	logservice "github.com/pachyderm/pachyderm/v2/src/server/logs"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

type testPublisher struct {
	responses []*logs.GetLogsResponse
}

func (tp *testPublisher) Publish(ctx context.Context, response *logs.GetLogsResponse) error {
	tp.responses = append(tp.responses, response)
	return nil
}

func TestGetLogsHint(t *testing.T) {
	ctx := pctx.TestContext(t)
	var ls logservice.LogService
	var publisher *testPublisher

	// GetLogs without a hint request should not return hints.
	publisher = new(testPublisher)
	require.NoError(t, ls.GetLogs(ctx, &logs.GetLogsRequest{}, publisher), "GetLogs should succeed")
	for _, r := range publisher.responses {
		_, ok := r.ResponseType.(*logs.GetLogsResponse_PagingHint)
		require.False(t, ok, "paging hints should not be returned when unasked for")
	}

	// GetLogs with a hint request should return a hint.
	publisher = new(testPublisher)
	require.NoError(t, ls.GetLogs(ctx, &logs.GetLogsRequest{WantPagingHint: true}, publisher), "GetLogs must succeed")
	require.True(t, len(publisher.responses) > 0, "there must be at least one response")
	h, ok := publisher.responses[0].ResponseType.(*logs.GetLogsResponse_PagingHint)
	require.True(t, ok, "paging hints should be returned when requested")
	for _, hint := range []*logs.GetLogsRequest{h.PagingHint.Older, h.PagingHint.Newer} {
		from := hint.Filter.TimeRange.From.AsTime()
		until := hint.Filter.TimeRange.Until.AsTime()
		require.Equal(t, 700*time.Hour, until.Sub(from), "default window is 700 hours")
	}

	// GetLogs with a hint request and a non-standard time filter should duplicate that duration.
	publisher = new(testPublisher)
	until := time.Now()
	from := until.Add(-1 * time.Hour)
	require.NoError(t, ls.GetLogs(ctx, &logs.GetLogsRequest{
		WantPagingHint: true,
		Filter: &logs.LogFilter{
			TimeRange: &logs.TimeRangeLogFilter{
				From:  timestamppb.New(from),
				Until: timestamppb.New(until),
			},
		},
	}, publisher), "GetLogs with an explicit time filter must succeed")
	require.True(t, len(publisher.responses) > 0, "there must be at least one response")
	h, ok = publisher.responses[0].ResponseType.(*logs.GetLogsResponse_PagingHint)
	require.True(t, ok, "paging hints should be returned when requested")
	for _, hint := range []*logs.GetLogsRequest{h.PagingHint.Older, h.PagingHint.Newer} {
		from := hint.Filter.TimeRange.From.AsTime()
		until := hint.Filter.TimeRange.Until.AsTime()
		require.Equal(t, time.Hour, until.Sub(from), "explicit window is one hour, not %v", until.Sub(from))
	}

	// GetLogs with a limit and a hint request is currently an error.
	//
	// TODO(CORE-2189): Update with support once logs are actually implemented.
	require.ErrorIs(t, ls.GetLogs(ctx, &logs.GetLogsRequest{WantPagingHint: true, Filter: &logs.LogFilter{Limit: 100}}, publisher), logservice.ErrUnimplemented, "GetLogs with both a limit and a hint request must error")
}
