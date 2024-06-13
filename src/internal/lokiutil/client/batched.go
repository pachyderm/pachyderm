package client

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync/atomic"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/prometheus/common/model"
	"go.uber.org/zap"
)

const (
	defaultRange     = 720 * time.Hour // If Loki can't return its configured max query range, use this one instead.
	defaultBatchSize = 1000            // If Loki can't return its configured batch size, use this one instead.
)

func Time(name string, t time.Time) zap.Field {
	return zap.String(name, t.Format(time.RFC3339Nano))
}

// FetchLimitsConfig returns the maximum query length and the maximum query batch size.
func (c *Client) FetchLimitsConfig(ctx context.Context) (time.Duration, int, error) {
	// query server to get the maximum batch size it supports
	config, err := c.QueryConfig(ctx)
	if err != nil {
		return defaultRange, defaultBatchSize, errors.Wrap(err, "querying config")
	}
	limits, ok := config["limits_config"].(map[any]any)
	if !ok {
		return defaultRange, defaultBatchSize, errors.Wrapf(err, "unknown limits config %#v", config["limits_config"])
	}
	batchSize, ok := limits["max_entries_limit_per_query"].(int)
	if !ok {
		return defaultRange, defaultBatchSize, errors.Wrapf(err, "unknown max_entries_limit_per_query %#v", limits["max_entries_limit_per_query"])
	}
	if maxRangeStr, ok := limits["max_query_length"].(string); ok {
		r, err := model.ParseDuration(maxRangeStr) // Loki uses Prometheus durations here, not Go durations.
		if err != nil {
			return defaultRange, batchSize, errors.Wrapf(err, "unparseable duration %q for max_query_length", maxRangeStr)
		}
		// Loki appears to add a millisecond to the time range that you provide it.  I have
		// no idea why.  This works around that; loki defaults to 721 hours and we have
		// always used 720 hours, so replicate that when possible.
		//
		// Example: start=2024-04-01T23:00:00.000000001Z, end=2024-05-02T00:00:00.000000001Z:
		// > the query time range exceeds the limit (query length: 721h0m0.001s, limit: 721h0m0s)
		var maxRange time.Duration
		if r := time.Duration(r); r > time.Hour {
			maxRange = r - time.Hour
		} else if r > time.Millisecond {
			maxRange = r - time.Millisecond
		} else {
			maxRange = r
		}
		return maxRange, batchSize, nil
	}
	return defaultRange, batchSize, errors.Wrapf(err, "unknown max_query_length %#v", limits["max_query_length"])
}

var (
	// BatchSizeForTesting configures a different batch size, intended for testing.  It's public
	// to avoid a circular dependency between Client and TestLoki.
	BatchSizeForTesting int = 0
	// MaxRangeForTesting configures a different max range, intended for testing.
	MaxRangeForTesting time.Duration = 0
)

// RecvFunc is a function that receives a log entry.  If it doesn't count towards the limit (i.e. is
// discarded by further filtering), the function should return false.
type RecvFunc func(ctx context.Context, labels LabelSet, e *Entry) (count bool, retErr error)

// labeledEntry stores labels and an entry together.  Loki returns logs as a stream of logs with
// fields applying to the entire stream, but we want to process log entries in time order, so we
// copy entries out of streams into this structure, sort them, and then call the RecvFunc in sorted
// order.
type labeledEntry struct {
	labels *LabelSet
	entry  *Entry
}

// batchQuery represents a single batched query.  It is only created by run(), and tracks counters
// across some common operations.
type batchQuery struct {
	client    *Client       // loki client
	query     string        // query strings
	offset    uint          // original offset
	limit     int           // current limit
	newOffset uint          // the offset after this batch ends
	recv      RecvFunc      // where to send matching logs
	maxRange  time.Duration // maximum timespan loki will accept
	batchSize int           // maximum limit loki will accept
	iter      *TimeIterator // the iterator over the query's time range
}

// BatchedQueryRange executes multiple QueryRange calls to return logs from the entire time
// interval, ending when the range is exhausted or the limit is reached.  recv is called with log
// entries sorted in traversal order (forwards or backwards; see TimeIterator), not one stream at a
// time.
//
// nextOffset and prevOffset is how deep into a stream of identical timestamps the querying is at
// the time this function returns; a value > 0 implies that logs may be being dropped.  Passing the
// returned offset in the relevant direction as offset to a query in the same direction that starts
// where this one ends will extract the remaining logs without duplication.  nextPage and nextOffset
// are the start time and offset for a forward paging hint; prevPage and prevOffset are the end time
// and offset for a backward paging hint.
func (c *Client) BatchedQueryRange(ctx context.Context, query string, start, end time.Time, offset uint, limit int, recv RecvFunc) (prevPage, nextPage time.Time, prevOffset, nextOffset uint, retErr error) {
	ctx, done := log.SpanContext(ctx, "BatchedQueryRange", zap.String("query", query))
	defer done(log.Errorp(&retErr))

	// We fetch the limits every time rather than caching them because Loki can be reconfigured
	// without restarting pachd.
	tctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	maxRange, batchSize, err := c.FetchLimitsConfig(tctx)
	cancel()
	if BatchSizeForTesting > 0 {
		batchSize = BatchSizeForTesting
	}
	if MaxRangeForTesting > 0 {
		maxRange = MaxRangeForTesting
	}
	if err != nil {
		log.Info(ctx, "failed to determine loki's maximum request size; assuming reasonable defaults", zap.Error(err), zap.Duration("range", maxRange), zap.Int("batch_size", batchSize))
	}
	iter := &TimeIterator{Start: start, End: end, Step: maxRange}
	// People tend to think that limit = 0 implies "unlimited", so we adjust for that here.
	if limit == 0 {
		limit = math.MaxInt
	}
	q := &batchQuery{
		client:    c,
		query:     query,
		offset:    offset,
		limit:     limit,
		recv:      recv,
		maxRange:  maxRange,
		batchSize: batchSize,
		iter:      iter,
	}
	defer func() {
		prevPage = iter.BackwardHint()
		nextPage = iter.ForwardHint()
		if iter.forward() {
			nextOffset = q.newOffset
		} else {
			prevOffset = q.newOffset
		}
	}()
	if err := q.run(ctx); err != nil {
		retErr = errors.Wrap(err, "run")
		return
	}
	return
}

func (q *batchQuery) run(rctx context.Context) error {
	for i := 0; q.iter.Next(); i++ {
		ctx := pctx.Child(rctx, fmt.Sprintf("run(%v)", i))
		s, e := q.iter.Interval()
		if i == 0 && q.offset > 0 {
			var firstNano time.Time
			if q.iter.forward() {
				// If forward, get the nanosecond at the query start.
				firstNano = s
			} else {
				// If backward, get the nanosecond at the query end.  (Loki's end is exclusive, so actually the nanosecond before.)
				firstNano = e.Add(-time.Nanosecond)
			}
			// If the query has an offset, resolve that first.
			if err := q.queryOneNanosecond(ctx, firstNano, q.offset); err != nil {
				return errors.Wrapf(err, "initial queryOneNanosecond(i=0, time=%v)", s.Format(time.RFC3339Nano))
			}
			if q.limit <= 0 {
				break
			}
			// queryOneNanosecond calls handleBatch calls iter.ObserveLast, which will
			// cause iter.Next() to advance 1 nanosecond in the correct direction.
			continue
		}
		q.newOffset = 0
		resp, err := q.client.QueryRange(ctx, q.query, q.batchSize, s, e, q.iter.Direction(), 0, 0, true)
		if err != nil {
			msg := "<nil>"
			if resp != nil {
				msg = resp.Status
			}
			return errors.Wrapf(err, "QueryRange(i=%v, start=%s, end=%s, limit=%v, dir=%v): (status=%v)", i, s.Format(time.RFC3339Nano), e.Format(time.RFC3339Nano), q.batchSize, q.iter.Direction(), msg)
		}

		last, err := q.handleBatch(ctx, resp, 0)
		if err != nil {
			return errors.Wrapf(err, "handleBatch(i=%v)", i)
		}
		if q.limit <= 0 {
			break
		}
		if o := q.newOffset; o > 0 {
			// If there's an offset remaining, we need to re-query the last millisecond.
			q.newOffset = 0
			if err := q.queryOneNanosecond(ctx, last, o); err != nil {
				return errors.Wrapf(err, "queryOneNanosecond(i=%v, time=%v)", i, s.Format(time.RFC3339Nano))
			}
			if q.limit <= 0 {
				break
			}
		}
	}
	return nil
}

func (q *batchQuery) handleBatch(ctx context.Context, resp *QueryResponse, offset uint) (time.Time, error) {
	streams, ok := resp.Data.Result.(Streams)
	if !ok {
		return time.Time{}, errors.Errorf(`unexpected response type from loki; got %q want "streams"`, resp.Data.ResultType)
	}
	var entries []labeledEntry
	for _, stream := range streams {
		for _, entry := range stream.Entries {
			entries = append(entries, labeledEntry{labels: &stream.Labels, entry: &entry})
		}
	}
	if len(entries) == 0 {
		q.newOffset = 0
		return time.Time{}, nil
	}
	sort.Slice(entries, func(i, j int) bool {
		if !q.iter.forward() {
			i, j = j, i
		}
		return entries[i].entry.Timestamp.Before(entries[j].entry.Timestamp)
	})
	last := entries[0].entry.Timestamp
	for i, e := range entries {
		if t := e.entry.Timestamp; t.Equal(last) {
			q.newOffset++
		} else {
			last = t
			q.newOffset = 1
		}
		if offset > 0 {
			offset--
			continue
		}
		// Note: this constrains the iterator more than is strictly necessary; if we
		// recevied less than batchSize logs, we know that `end` is really the start of the
		// next iteration, not the time of the last log we saw.  But as you go forward in
		// time, some logs may not have arrived yet, and since we use the iterator to
		// provide paging hints, doing this optimization would cause some of those
		// new-arriving entries to be skipped over when paging.
		q.iter.ObserveLast(last)
		count, err := q.recv(ctx, *e.labels, e.entry)
		if err != nil {
			return time.Time{}, errors.Wrapf(err, "handle entry %v", i)
		}
		if count {
			q.limit--
			if q.limit <= 0 {
				if i+1 < len(entries) && !e.entry.Timestamp.Equal(entries[i+1].entry.Timestamp) {
					// Peek at the next log, even though we're returning now.
					// If it's at a different time than this, we know the offset
					// is 0.
					q.newOffset = 0
				}
				return last, nil
			}
		}
	}
	return last, nil
}

// BatchSizeWarning is set to true if we have to warn about Loki's batch size being too small.  It's
// intended to be read by tests, which live in a different package.
var BatchSizeWarning atomic.Bool

func (q *batchQuery) queryOneNanosecond(ctx context.Context, t time.Time, offset uint) (retErr error) {
	ctx, done := log.SpanContext(ctx, "queryOneNanosecond", Time("time", t))
	defer done(log.Errorp(&retErr))

	resp, err := q.client.QueryRange(ctx, q.query, q.batchSize, t, t.Add(time.Nanosecond), q.iter.Direction(), 0, 0, true)
	if err != nil {
		msg := "<nil>"
		if resp != nil {
			msg = resp.Status
		}
		return errors.Wrapf(err, "QueryRange(nanosecond=%s, limit=%v, dir=%v): (status=%v)", t.Format(time.RFC3339Nano), q.batchSize, q.iter.Direction(), msg)
	}

	if _, err := q.handleBatch(ctx, resp, offset); err != nil {
		return errors.Wrap(err, "handleBatch")
	}
	if q.newOffset >= uint(q.batchSize) {
		BatchSizeWarning.Store(true)
		log.Error(ctx, "skipping logs, because Loki's batch size is too small to retrieve all logs at a duplicate timestamp; this likely indicates a configuration issue with log ingestion", zap.Int("batchSize", q.batchSize))
	}

	if q.limit > 0 {
		// After querying a nanosecond, we always reset the offset, since there is nothing
		// else Loki can give us.  The exception is if we stopped iteration because of the
		// limit, in which case the offset is correct.
		q.newOffset = 0
	}
	return nil
}
