package logs

// SPDX-FileCopyrightText: © 2024 Hewlett Packard Enterprise
// SPDX-FileCopyrightText: © 2021 Grafana Labs
// SPDX-License-Identifier: Apache-2.0

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	loki "github.com/pachyderm/pachyderm/v2/src/internal/lokiutil/client"
)

type ErrInvalidBatchSize struct {
	batchSize, overlappingCount int
}

func (err ErrInvalidBatchSize) Error() string {
	return fmt.Sprintf("invalid batch size %v, the next query will have %v overlapping entries "+
		"(there will always be 1 overlapping entry but Loki allows multiple entries to have "+
		"the same timestamp, so when a batch ends in this scenario the next query will include "+
		"all the overlapping entries again).  Please increase your batch size to at least %v to account "+
		"for overlapping entryes\n", err.batchSize, err.overlappingCount, err.RecommendedBatchSize())
}

func (err ErrInvalidBatchSize) RecommendedBatchSize() int {
	return err.overlappingCount + 1
}

// doQuery executes a query.
//
// Adapted from <URL:https://github.com/grafana/loki/blob/3c78579676562b06e73791d71fcf6e3abf50a014/pkg/logcli/query/query.go>.
//
// License: Apache 2.0 <URL:https://github.com/grafana/loki/blob/3c78579676562b06e73791d71fcf6e3abf50a014/LICENSE>.
func doQuery(ctx context.Context, client *loki.Client, logQL string, limit int, start, end time.Time, direction logDirection) ([]loki.Entry, error) {
	var (
		batchSize    = limit
		resultLength int
		total        int
		results      []loki.Entry
		lastEntry    []loki.Entry
	)
	for total < limit {
		var result []loki.Entry
		bs := batchSize
		// We want to truncate the batch size if the remaining number
		// of items needed to reach the limit is less than the batch size
		if limit-total < batchSize {
			// Truncated batchsize is limit - total, however we add to this
			// the length of the overlap from the last query to make sure we get the
			// correct amount of new logs knowing there will be some overlapping logs returned.
			bs = limit - total + len(lastEntry)
		}
		resp, err := client.QueryRange(ctx, logQL, bs, start, end, string(direction), 0, 0, true)
		if err != nil {
			// It would be nice to distinguish user query errors
			// from internal errors here, but … any invalid user
			// query passed through by the log server is effectively
			// an internal error.
			return nil, errors.Wrap(err, "query failed")
		}

		streams, ok := resp.Data.Result.(loki.Streams)
		if !ok {
			return nil, errors.Errorf("resp.Data.Result must be of type loghttp.Streams, not %T", resp.Data.Result)
		}

		if result, resultLength, lastEntry, err = publishEntries(ctx, streams, direction, lastEntry); err != nil {
			return nil, errors.Wrap(err, "could not publish entries")
		}
		results = append(results, result...)
		if resultLength <= 0 {
			// Was not a log stream query, or no results, no more batching
			break
		} else if len(lastEntry) == 0 {
			// Also no result, wouldn't expect to hit this.
			break
		} else if resultLength == limit {
			// Can only happen if all the results return in one request
			break
		}
		if len(lastEntry) >= batchSize {
			return nil, ErrInvalidBatchSize{batchSize, len(lastEntry)}
		}

		// Batching works by taking the timestamp of the last query and using it in the next query,
		// because Loki supports multiple entries with the same timestamp it's possible for a batch to have
		// fallen in the middle of a list of entries for the same time, so to make sure we get all entries
		// we start the query on the same time as the last entry from the last batch, and then we keep this last
		// entry and remove the duplicate when printing the results.
		// Because of this duplicate entry, we have to subtract it here from the total for each batch
		// to get the desired limit.
		total += resultLength
		// Based on the query direction we either set the start or end for the next query.
		// If there are multiple entries in `lastEntry` they have to have the same timestamp so we can pick just the first
		if direction == forwardLogDirection {
			start = lastEntry[0].Timestamp
		} else {
			// The end timestamp is exclusive on a backward query, so to make sure we get back an overlapping result
			// fudge the timestamp forward in time to make sure to get the last entry from this batch in the next query
			end = lastEntry[0].Timestamp.Add(1 * time.Nanosecond)
		}
	}
	return results, nil
}

// Adapted from <URL:https://github.com/grafana/loki/blob/3c78579676562b06e73791d71fcf6e3abf50a014/pkg/logcli/query/query.go#L259>.
//
// License: Apache 2.0 <URL:https://github.com/grafana/loki/blob/3c78579676562b06e73791d71fcf6e3abf50a014/LICENSE>.
func publishEntries(ctx context.Context, streams loki.Streams, direction logDirection, lastEntry []loki.Entry) ([]loki.Entry, int, []loki.Entry, error) {
	var (
		entries, result []loki.Entry
		published       int
	)
	for _, s := range streams {
		entries = append(entries, s.Entries...)
	}
	if len(entries) == 0 {
		return nil, 0, nil, nil
	}
	switch direction {
	case forwardLogDirection:
		sort.Slice(entries, func(i, j int) bool { return entries[i].Timestamp.Before(entries[j].Timestamp) })
	case backwardLogDirection:
		sort.Slice(entries, func(i, j int) bool { return entries[i].Timestamp.After(entries[j].Timestamp) })
	default:
		return nil, 0, nil, errors.Errorf("invalid direction %q", direction)
	}
	for _, e := range entries {
		if len(lastEntry) > 0 && e.Timestamp == lastEntry[0].Timestamp {
			skip := false
			for _, le := range lastEntry {
				if e.Line == le.Line {
					skip = true
				}
			}
			if skip {
				continue
			}
		}
		result = append(result, e)
		published++
	}
	var lel []loki.Entry
	le := entries[len(entries)-1]
	for _, e := range entries {
		if e.Timestamp.Equal(le.Timestamp) {
			lel = append(lel, e)
		}
	}
	return result, published, lel, nil
}
