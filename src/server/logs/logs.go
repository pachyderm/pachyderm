package logs

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/pachyderm/pachyderm/v2/src/logs"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"

	loki "github.com/pachyderm/pachyderm/v2/src/internal/lokiutil/client"
)

type ResponsePublisher interface {
	// Publish publishes a single GetLogsResponse to the client.
	Publish(context.Context, *logs.GetLogsResponse) error
}

// LogService implements the core logs functionality.
type LogService struct {
	GetLokiClient func() (*loki.Client, error)
}

var (
	// ErrUnimplemented is returned whenever requested functionality is planned but unimplemented.
	ErrUnimplemented = errors.New("unimplemented")
	// ErrPublish is returned whenever publishing fails (say, due to a closed client).
	ErrPublish = errors.New("error publishing")
)

// GetLogs gets logs according its request and publishes them.  The pattern is
// similar to that used when handling an HTTP request.
func (ls LogService) GetLogs(ctx context.Context, request *logs.GetLogsRequest, publisher ResponsePublisher) error {
	if err := validateGetLogsRequest(request); err != nil {
		return errors.Wrap(err, "invalid GetLogs request")
	}

	if request == nil {
		request = &logs.GetLogsRequest{}
	}

	filter := request.Filter

	if filter == nil {
		filter = new(logs.LogFilter)
		request.Filter = filter
	}
	switch {
	case filter.TimeRange == nil || (filter.TimeRange.From == nil && filter.TimeRange.Until == nil):
		now := time.Now()
		filter.TimeRange = &logs.TimeRangeLogFilter{
			From:  timestamppb.New(now.Add(-700 * time.Hour)),
			Until: timestamppb.New(now),
		}
	case filter.TimeRange.From == nil:
		filter.TimeRange.From = timestamppb.New(filter.TimeRange.Until.AsTime().Add(-700 * time.Hour))
	case filter.TimeRange.Until == nil:
		filter.TimeRange.Until = timestamppb.New(filter.TimeRange.From.AsTime().Add(700 * time.Hour))
	}

	if request.WantPagingHint {
		hint := &logs.PagingHint{
			Older: proto.Clone(request).(*logs.GetLogsRequest),
			Newer: proto.Clone(request).(*logs.GetLogsRequest),
		}
		from, until := request.Filter.TimeRange.From.AsTime(), request.Filter.TimeRange.Until.AsTime()
		hint.Older.Filter.TimeRange.From = timestamppb.New(from.Add(from.Sub(until)))
		hint.Older.Filter.TimeRange.Until = timestamppb.New(from)
		hint.Newer.Filter.TimeRange.From = timestamppb.New(until)
		hint.Newer.Filter.TimeRange.Until = timestamppb.New(until.Add(until.Sub(from)))
		if err := publisher.Publish(ctx, &logs.GetLogsResponse{ResponseType: &logs.GetLogsResponse_PagingHint{PagingHint: hint}}); err != nil {
			return errors.WithStack(fmt.Errorf("%w paging hint: %w", ErrPublish, err))
		}
	}

	outChannel := make(chan []byte)

	go func() {
		c, err := ls.GetLokiClient()
		if err != nil {
			panic(fmt.Sprintf("loki client error: %v", err))
		}
		resp, err := c.QueryRange(ctx, request.GetQuery().GetAdmin().GetLogql(), int(filter.GetLimit()), filter.GetTimeRange().GetFrom().AsTime(), filter.GetTimeRange().GetUntil().AsTime(), "forward", 0, 0, true)
		if err != nil {
			panic(fmt.Sprintf("error: %v", err))
		}
		streams, ok := resp.Data.Result.(loki.Streams)
		if !ok {
			panic("resp.Data.Result must be of type loghttp.Streams to call ForEachStream on it")
		}

		for _, s := range streams {
			for _, e := range s.Entries {
				outChannel <- []byte(e.Line)
			}
		}

		close(outChannel)
	}()

	for line := range outChannel {
		var resp *logs.GetLogsResponse
		switch request.LogFormat {
		case logs.LogFormat_LOG_FORMAT_UNKNOWN:
			return errors.New("unknown log format not support") // TODO: return unimplemented
		case logs.LogFormat_LOG_FORMAT_VERBATIM_WITH_TIMESTAMP:
			resp = &logs.GetLogsResponse{ResponseType: &logs.GetLogsResponse_Log{Log: &logs.LogMessage{LogType: &logs.LogMessage_Verbatim{&logs.VerbatimLogMessage{Line: line}}}}}
		default:
			return errors.Errorf("%v not supported", request.LogFormat)
		}

		if err := publisher.Publish(ctx, resp); err != nil {
			return errors.WithStack(fmt.Errorf("%w response with parsed json object: %w", ErrPublish, err))
		}

	}

	return nil

}

func validateGetLogsRequest(request *logs.GetLogsRequest) error {
	if request.GetWantPagingHint() {
		// TODO(CORE-2189): actually need to have logs to implement the logic with limits
		if request.GetFilter().GetLimit() > 0 {
			return errors.Wrap(ErrUnimplemented, "paging hints with limit > 0")
		}
	}
	return nil
}
