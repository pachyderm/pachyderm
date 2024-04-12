package logs

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/pachyderm/pachyderm/v2/src/logs"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	ilogs "github.com/pachyderm/pachyderm/v2/src/internal/logs"
	loki "github.com/pachyderm/pachyderm/v2/src/internal/lokiutil/client"
	"github.com/pachyderm/pachyderm/v2/src/pps"
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
	// ErrLogFormat returned if log line does not match requested log format
	ErrLogFormat = errors.New("error invalid log format")
)

type logDirection string

const (
	forwardLogDirection  logDirection = "forward"
	backwardLogDirection logDirection = "backward"
)

var lokiQuery string

func handlePipelineJob(ctx context.Context, query *logs.PipelineJobLogQuery) ([]ilogs.LineMatcher, error) {
	if query == nil || query.GetPipeline == nil {
		return nil, nil
	}
	matcher, err := ilogs.PipelineJobMatcher(query.GetPipeline().GetProject(), query.GetPipeline().GetPipeline(), query.GetJob())
	if err != nil {
		return nil, err
	}
	return []ilogs.LineMatcher{matcher}, nil
}

func handlePipeline(ctx context.Context, query *logs.PipelineLogQuery) ([]ilogs.LineMatcher, error) {
	if query == nil {
		return nil, nil
	}
	matcher, err := ilogs.PipelineMatcher(query.GetProject(), query.GetPipeline())
	if err != nil {
		return nil, err
	}
	return []ilogs.LineMatcher{matcher}, nil
}

func handleUserLogQuery(ctx context.Context, query *logs.UserLogQuery) ([]ilogs.LineMatcher, error) {
	if query == nil {
		return nil, nil
	}
	matchers := []ilogs.LineMatcher{}

	matcher, err := ilogs.ProjectMatcher(query.GetProject())
	if err != nil {
		return nil, err
	}
	matchers = append(matchers, matcher)

	moreMatchers, err := handlePipeline(ctx, query.GetPipeline())
	if err != nil {
		return nil, err
	}
	matchers = append(matchers, moreMatchers...)

	matcher, err = ilogs.DatumMatcher(query.GetDatum())
	if err != nil {
		return nil, err
	}
	matchers = append(matchers, matcher)

	matcher, err = ilogs.JobMatcher(query.GetJob())
	if err != nil {
		return nil, err
	}
	matchers = append(matchers, matcher)

	moreMatchers, err = handlePipelineJob(ctx, query.GetPipelineJob())
	if err != nil {
		return nil, err
	}
	matchers = append(matchers, moreMatchers...)

	return matchers, nil
}

func handleAdminLogQuery(ctx context.Context, query *logs.AdminLogQuery) ([]ilogs.LineMatcher, error) {
	if query == nil {
		return nil, nil
	}

	switch query.AdminType.(type) {
	case *logs.AdminLogQuery_Logql:
		lokiQuery = query.GetLogql()
	case *logs.AdminLogQuery_Pod:
		lokiQuery = query.GetPod()
	case *logs.AdminLogQuery_PodContainer:
		lokiQuery = query.GetPodContainer().GetPod()
	case *logs.AdminLogQuery_App:
		lokiQuery = query.GetApp()
	case *logs.AdminLogQuery_Master:
		matcher, err := ilogs.MasterMatcher(query.GetStorage().GetProject(), query.GetStorage().GetPipeline())
		if err != nil {
			return nil, err
		}
		return []ilogs.LineMatcher{matcher}, nil
	case *logs.AdminLogQuery_Storage:
		matcher, err := ilogs.StorageMatcher(query.GetStorage().GetProject(), query.GetStorage().GetPipeline())
		if err != nil {
			return nil, err
		}
		return []ilogs.LineMatcher{matcher}, nil
	case *logs.AdminLogQuery_User:
		return handleUserLogQuery(ctx, query.GetUser())
	}

	return nil, nil

}

func handleQuery(ctx context.Context, query *logs.LogQuery) ([]ilogs.LineMatcher, error) {

	if query == nil {
		return nil, nil
	}

	log.Info(ctx, "handleQuery: entry")
	switch queryType := query.QueryType.(type) {
	case *logs.LogQuery_User:
		log.Info(ctx, fmt.Sprintf("handleQuery: got user query type %T", queryType))
		return handleUserLogQuery(ctx, query.GetUser())
	case *logs.LogQuery_Admin:
		log.Info(ctx, fmt.Sprintf("handleQuery: got admin query type %v", queryType))
		return handleAdminLogQuery(ctx, query.GetAdmin())
	}

	log.Info(ctx, fmt.Sprintf("handleQuery: got unknown query type"))
	return nil, errors.Errorf("unexpected query type")

}

func handleFilters(ctx context.Context, filters *logs.LogFilter) ([]ilogs.LineMatcher, error) {
	if filters == nil {
		return nil, nil
	}
	matchers := []ilogs.LineMatcher{}

	matcher, err := ilogs.LogLevelMatcher(uint32(filters.Level))
	if err != nil {
		return nil, err
	}
	matchers = append(matchers, matcher)

	if filters.Regex != nil {
		matcher, err := ilogs.RegexMatcher(filters.Regex.Pattern, filters.Regex.Negate)
		if err != nil {
			return nil, err
		}
		matchers = append(matchers, matcher)
	}

	return matchers, nil

}

// GetLogs gets logs according its request and publishes them.  The pattern is
// similar to that used when handling an HTTP request.
func (ls LogService) GetLogs(ctx context.Context, request *logs.GetLogsRequest, publisher ResponsePublisher) error {
	var direction = forwardLogDirection

	if request == nil {
		request = &logs.GetLogsRequest{}
	}

	lineMatchers := []ilogs.LineMatcher{}

	matchers, err := handleQuery(ctx, request.Query)
	if err != nil {
	}
	lineMatchers = append(lineMatchers, matchers...)

	matchers, err = handleFilters(ctx, request.Filter)
	if err != nil {
	}
	lineMatchers = append(lineMatchers, matchers...)

	filter := request.Filter
	if filter == nil {
		filter = new(logs.LogFilter)
		request.Filter = filter
	}
	if filter.Limit > math.MaxInt {
		return errors.Wrapf(ErrBadRequest, "limit %d > maxint", filter.Limit)
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
		direction = backwardLogDirection
		start, end = end, start
	}

	adapter := newAdapter(publisher, request.LogFormat, lineMatchers)
	if err = doQuery(ctx, c, request.GetQuery().GetAdmin().GetLogql(), int(filter.Limit), start, end, direction, adapter.publish); err != nil {
		var invalidBatchSizeErr ErrInvalidBatchSize
		switch {
		case errors.As(err, &invalidBatchSizeErr):
			// try to requery
			err = doQuery(ctx, c, request.GetQuery().GetAdmin().GetLogql(), invalidBatchSizeErr.RecommendedBatchSize(), start, end, direction, adapter.publish)
			if err != nil {
				return errors.Wrap(err, "invalid batch size requery failed")
			}
		default:
			return errors.Wrap(err, "doQuery failed")
		}
	}
	if request.WantPagingHint {
		var newer, older loki.Entry
		switch direction {
		case forwardLogDirection:
			newer = adapter.last
		case backwardLogDirection:
			newer = adapter.first
		default:
			return errors.Errorf("invalid direction %q", direction)
		}
		// request a record immediately prior to the page
		err := doQuery(ctx, c, request.GetQuery().GetAdmin().GetLogql(), 1, start.Add(-700*time.Hour), start, backwardLogDirection, func(ctx context.Context, e loki.Entry) error {
			older = e
			return nil
		})
		if err != nil {
			return errors.Wrap(err, "hint doQuery failed")
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
				hint.Newer.Filter.TimeRange.Until = timestamppb.New(newer.Timestamp.Add(delta))
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

	return nil
}

// An adapter publishes log entries to a ResponsePublisher in a specified format.
type adapter struct {
	responsePublisher ResponsePublisher
	logFormat         logs.LogFormat
  lineMatchers      []ilogs.LineMatcher
	first, last       loki.Entry
	gotFirst          bool
}

func newAdapter(p ResponsePublisher, f logs.LogFormat, matchers []ilogs.LineMatcher) *adapter {
	return &adapter{
		responsePublisher: p,
		logFormat:         f,
    lineMatchers: matchers,
	}
}

func (a *adapter) publish(ctx context.Context, entry loki.Entry) error {
	if !a.gotFirst {
		a.gotFirst = true
		a.first = entry
	}
	a.last = entry
	var resp *logs.GetLogsResponse
	switch a.logFormat {
	case logs.LogFormat_LOG_FORMAT_UNKNOWN:
		return errors.Wrap(ErrUnimplemented, "unknown log format not supported")
	case logs.LogFormat_LOG_FORMAT_VERBATIM_WITH_TIMESTAMP:
		resp = &logs.GetLogsResponse{
			ResponseType: &logs.GetLogsResponse_Log{
				Log: &logs.LogMessage{
					LogType: &logs.LogMessage_Verbatim{
						Verbatim: &logs.VerbatimLogMessage{
							Line: []byte(entry.Line),
              Timestamp: timestamppb.New(entry.Timestamp),
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
                  Line:      []byte(entry.Line),
                  Timestamp: timestamppb.New(entry.Timestamp),
                },
                NativeTimestamp: timestamppb.New(entry.Timestamp),
              },
            },
          },
        },
      }
      jsonStruct := new(structpb.Struct)
      if err := jsonStruct.UnmarshalJSON([]byte(entry.Line)); err != nil {
        log.Error(ctx, "failed to unmarshal json into protobuf Struct", zap.Error(err), zap.String("line", entry.Line))
      } else {
        resp.GetLog().GetJson().Object = jsonStruct
      }
      ppsLog := new(pps.LogMessage)
      m := protojson.UnmarshalOptions{
        AllowPartial:   true,
        DiscardUnknown: true,
      }
      if err := m.Unmarshal([]byte(entry.Line), ppsLog); err != nil {
        log.Error(ctx, "failed to unmarshal json into PpsLogMessage", zap.Error(err), zap.String("line", entry.Line))
      } else {
        resp.GetLog().GetJson().PpsLogMessage = ppsLog
      }
    case logs.LogFormat_LOG_FORMAT_PPS_LOGMESSAGE:
      ppsLog := new(pps.LogMessage)
      m := protojson.UnmarshalOptions{
        AllowPartial:   true,
        DiscardUnknown: true,
      }
      if err := m.Unmarshal([]byte(entry.Line), ppsLog); err != nil {
        return errors.Wrapf(ErrLogFormat, "log line cannot be formatted as %v", a.logFormat, zap.String("line", entry.Line))
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
		return errors.Wrapf(ErrUnimplemented, "%v not supported", a.logFormat)
	}

  logFields := new(ilogs.LogFields)
  json.Unmarshal([]byte(entry.Line), logFields)

  if len(a.lineMatchers) == 0 {
    if err := a.responsePublisher.Publish(ctx, resp); err != nil {
      return errors.WithStack(fmt.Errorf("%w response with parsed json object: %w", ErrPublish, err))
    }
  }

  for _, matcher := range a.lineMatchers {
    if matcher([]byte(entry.Line), logFields) {
      if err := a.responsePublisher.Publish(ctx, resp); err != nil {
        return errors.WithStack(fmt.Errorf("%w response with parsed json object: %w", ErrPublish, err))
      }
      break
    }
  }

	if err := a.responsePublisher.Publish(ctx, resp); err != nil {
		return errors.WithStack(fmt.Errorf("%w response with parsed json object: %w", ErrPublish, err))
	}
	return nil
}
