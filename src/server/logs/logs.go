package logs

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/pachyderm/pachyderm/v2/src/logs"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"

	ilogs "github.com/pachyderm/pachyderm/v2/src/internal/logs"
	loki "github.com/pachyderm/pachyderm/v2/src/internal/lokiutil/client"
)

type ResponsePublisher interface {
	// Publish publishes a single GetLogsResponse to the client.
	Publish(context.Context, *logs.GetLogsResponse) error
}

// LogService implements the core logs functionality.
type LogService struct {
	getLogs       *ilogs.GetLogs
	GetLokiClient func() (*loki.Client, error)
	lineMatchers  *ilogs.MatchLines
	maxLines      uint64
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

	ls.lineMatchers = new(ilogs.MatchLines)

	ls.getLogs = &ilogs.GetLogs{
		Matchers:      ls.lineMatchers,
		GetLokiClient: ls.GetLokiClient,
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

	matcher, err := ilogs.TrueMatcher()
	if err != nil {
		return err
	}
	ls.lineMatchers.Matchers = append(ls.lineMatchers.Matchers, matcher)

	outChannel := make(chan []byte)

	maxLines := ls.maxLines
	var numLines uint64

	ls.getLogs.StartMatching(ctx, func(line []byte) error {

		select {
		case outChannel <- line:
			numLines++
			if numLines == maxLines {
				return errutil.ErrBreak
			}
			return nil
		case <-ctx.Done():
			return errutil.ErrBreak
		}
	})

	for line := range outChannel {

		numLines++
		if numLines == maxLines {
			break
		}
		jsonStruct := new(structpb.Struct)
		err := jsonStruct.UnmarshalJSON([]byte(line))
		if err != nil {
			return errors.WithStack(fmt.Errorf("%w convert json log to struct: %w", ErrPublish, err))
		}

		resp := &logs.GetLogsResponse{ResponseType: &logs.GetLogsResponse_Log{Log: &logs.LogMessage{LogType: &logs.LogMessage_Json{&logs.ParsedJSONLogMessage{Object: jsonStruct}}}}}

		if err := publisher.Publish(ctx, resp); err != nil {
			return errors.WithStack(fmt.Errorf("%w response with parsed json object: %w", ErrPublish, err))
		}

	}

	close(outChannel)
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
