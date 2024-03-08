package logs

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/pachyderm/pachyderm/v2/src/logs"
	"github.com/pachyderm/pachyderm/v2/src/pps"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

type ResponsePublisher interface {
	// Publish publishes a single GetLogsResponse to the client.
	Publish(context.Context, *logs.GetLogsResponse) error
}

// LogService implements the core logs functionality.
type LogService struct{}

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
	if filter.TimeRange == nil {
		now := time.Now()
		filter.TimeRange = &logs.TimeRangeLogFilter{
			From:  timestamppb.New(now.Add(-700 * time.Hour)),
			Until: timestamppb.New(now),
		}
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
	// TODO(CORE-2189): return all the actual logs
	if err := publisher.Publish(ctx, &logs.GetLogsResponse{ResponseType: &logs.GetLogsResponse_Log{Log: &logs.LogMessage{LogType: &logs.LogMessage_PpsLogMessage{PpsLogMessage: &pps.LogMessage{Message: "GetLogs dummy response"}}}}}); err != nil {
		return errors.WithStack(fmt.Errorf("%w dummy: %w", ErrPublish, err))
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
