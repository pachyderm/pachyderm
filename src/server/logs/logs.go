package logs

import (
	"context"
	"fmt"
	"math"
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
	ErrPublish    = errors.New("error publishing")
	ErrBadRequest = errors.New("bad request")
)

// GetLogs gets logs according its request and publishes them.  The pattern is
// similar to that used when handling an HTTP request.
func (ls LogService) GetLogs(ctx context.Context, request *logs.GetLogsRequest, publisher ResponsePublisher) error {
	var direction = "forward"
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
	if filter.Limit+1 > math.MaxInt {
		return errors.Wrapf(ErrBadRequest, "limit %d > maxint", filter.Limit)
	}
	if filter.Limit == 0 {
		filter.Limit = 100
	}
	// retrieve one extra record for paging
	filter.Limit++
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

	c, err := ls.GetLokiClient()
	if err != nil {
		return errors.Wrap(err, "loki client error")
	}
	start := filter.TimeRange.From.AsTime()
	end := filter.TimeRange.Until.AsTime()
	if start.Equal(end) {
		return errors.Errorf("start equals end (%v)", start)
	}
	if start.After(end) {
		direction = "backward"
		start, end = end, start
	}
	// FIXME: retry with new batch size when necessary, or return error?

	entries, err := doQuery(ctx, c, request.GetQuery().GetAdmin().GetLogql(), int(filter.Limit+1), start, end, direction)
	if err != nil {
		return errors.Wrap(err, "doQuery failed")
	}
	if request.WantPagingHint {
		var newer, older loki.Entry
		if len(entries) > 1 {
			switch direction {
			case "forward":
				var l = len(entries) - 1
				newer = entries[l]
				entries = entries[:l]
			case "backward":
				newer = entries[0]
				entries = entries[1:]
			default:
				return errors.Errorf("invalid direction %q", direction)
			}
		}
		// request a record immediately prior to the page
		entries, err := doQuery(ctx, c, request.GetQuery().GetAdmin().GetLogql(), 1, start.Add(-700*time.Hour), start, "backward")
		if err != nil {
			return errors.Wrap(err, "hint doQuery failed")
		}
		if len(entries) > 0 {
			older = entries[len(entries)-1]
		}
		hint := &logs.PagingHint{
			Older: proto.Clone(request).(*logs.GetLogsRequest),
			Newer: proto.Clone(request).(*logs.GetLogsRequest),
		}
		if !older.Timestamp.IsZero() {
			hint.Older.Filter.TimeRange.Until = timestamppb.New(older.Timestamp)
		}
		if !newer.Timestamp.IsZero() {
			hint.Newer.Filter.TimeRange.From = timestamppb.New(newer.Timestamp)
		}
		if request.Filter.TimeRange.From != nil && request.Filter.TimeRange.Until != nil {
			delta := request.Filter.TimeRange.Until.AsTime().Sub(request.Filter.TimeRange.From.AsTime())
			if !older.Timestamp.IsZero() {
				hint.Older.Filter.TimeRange.From = timestamppb.New(older.Timestamp.Add(-delta))
			}
			if !newer.Timestamp.IsZero() {
				hint.Newer.Filter.TimeRange.Until = timestamppb.New(older.Timestamp.Add(delta))
			}
		}
		if err := publisher.Publish(ctx, &logs.GetLogsResponse{
			ResponseType: &logs.GetLogsResponse_PagingHint{
				PagingHint: hint,
			},
		}); err != nil {
			return errors.WithStack(fmt.Errorf("%w paging hint: %w", ErrPublish, err))
		}
	}
	for _, entry := range entries {
		var resp *logs.GetLogsResponse
		switch request.LogFormat {
		case logs.LogFormat_LOG_FORMAT_UNKNOWN:
			return errors.Wrap(ErrUnimplemented, "unknown log format not supported")
		case logs.LogFormat_LOG_FORMAT_VERBATIM_WITH_TIMESTAMP:
			resp = &logs.GetLogsResponse{
				ResponseType: &logs.GetLogsResponse_Log{
					Log: &logs.LogMessage{
						LogType: &logs.LogMessage_Verbatim{
							Verbatim: &logs.VerbatimLogMessage{
								Line: []byte(entry.Line),
							},
						},
					},
				},
			}
		default:
			return errors.Wrapf(ErrUnimplemented, "%v not supported", request.LogFormat)
		}

		if err := publisher.Publish(ctx, resp); err != nil {
			return errors.WithStack(fmt.Errorf("%w response with parsed json object: %w", ErrPublish, err))
		}
	}

	return nil
}

func validateGetLogsRequest(request *logs.GetLogsRequest) error {
	// FIXME: any validation required?
	return nil
}
