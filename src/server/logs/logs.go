package logs

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/pachyderm/pachyderm/v2/src/logs"
	"github.com/pachyderm/pachyderm/v2/src/pps"

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
		if err := publisher.Publish(ctx, &logs.GetLogsResponse{
			ResponseType: &logs.GetLogsResponse_PagingHint{
				PagingHint: hint,
			},
		}); err != nil {
			return errors.WithStack(fmt.Errorf("%w paging hint: %w", ErrPublish, err))
		}
	}

	c, err := ls.GetLokiClient()
	if err != nil {
		return errors.Wrap(err, "loki client error")
	}
	resp, err := c.QueryRange(ctx, request.GetQuery().GetAdmin().GetLogql(), int(filter.GetLimit()), filter.GetTimeRange().GetFrom().AsTime(), filter.GetTimeRange().GetUntil().AsTime(), "forward", 0, 0, true)
	if err != nil {
		return errors.Wrap(err, "Loki QueryRange")
	}
	streams, ok := resp.Data.Result.(loki.Streams)
	if !ok {
		return errors.Errorf("resp.Data.Result must be of type loghttp.Streams, not %T, to call ForEachStream on it", resp.Data.Result)
	}
	for _, s := range streams {
		for _, e := range s.Entries {
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
									Line:      []byte(e.Line),
									Timestamp: timestamppb.New(e.Timestamp),
								},
							},
						},
					},
				}
			case logs.LogFormat_LOG_FORMAT_PARSED_JSON:
				resp = &logs.GetLogsResponse{
					ResponseType: &logs.GetLogsResponse_Log{
						Log: &logs.LogMessage{
							LogType: &logs.LogMessage_Json{
								Json: &logs.ParsedJSONLogMessage{
									Verbatim: &logs.VerbatimLogMessage{
										Line:      []byte(e.Line),
										Timestamp: timestamppb.New(e.Timestamp),
									},
									NativeTimestamp: timestamppb.New(e.Timestamp),
								},
							},
						},
					},
				}
				jsonStruct := new(structpb.Struct)
				if err := jsonStruct.UnmarshalJSON([]byte(e.Line)); err == nil {
					resp.GetLog().GetJson().Object = jsonStruct
				}
				ppsLog := new(pps.LogMessage)
				m := protojson.UnmarshalOptions{
					AllowPartial:   true,
					DiscardUnknown: true,
				}
				if err := m.Unmarshal([]byte(e.Line), ppsLog); err == nil {
					resp.GetLog().GetJson().PpsLogMessage = ppsLog
				}
			case logs.LogFormat_LOG_FORMAT_PPS_LOGMESSAGE:
				ppsLog := new(pps.LogMessage)
				m := protojson.UnmarshalOptions{
					AllowPartial:   true,
					DiscardUnknown: true,
				}
				if err := m.Unmarshal([]byte(e.Line), ppsLog); err != nil {
					continue
				}
				resp = &logs.GetLogsResponse{
					ResponseType: &logs.GetLogsResponse_Log{
						Log: &logs.LogMessage{
							LogType: &logs.LogMessage_PpsLogMessage{
								PpsLogMessage: ppsLog,
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
