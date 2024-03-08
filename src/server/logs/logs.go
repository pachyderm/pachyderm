package logs

import (
	"context"
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

type LogService struct{}

var ErrUnimplemented = errors.New("unimplemented")

func (ls LogService) GetLogs(ctx context.Context, request *logs.GetLogsRequest, responseWriter ResponsePublisher) error {
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
		hint.Newer.Filter.TimeRange.From = timestamppb.New(until.Add(until.Sub(from)))
		if err := responseWriter.Publish(ctx, &logs.GetLogsResponse{ResponseType: &logs.GetLogsResponse_PagingHint{PagingHint: hint}}); err != nil {
			return errors.Wrap(err, "error publishing paging hint")
		}
	}
	// TODO(CORE-2189): return all the actual logs
	if err := responseWriter.Publish(ctx, &logs.GetLogsResponse{ResponseType: &logs.GetLogsResponse_Log{Log: &logs.LogMessage{LogType: &logs.LogMessage_PpsLogMessage{PpsLogMessage: &pps.LogMessage{Message: "GetLogs dummy response"}}}}}); err != nil {
		return errors.Wrap(err, "error publishing dummy log")
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
