package lokiutil

import (
	"context"
	"encoding/json"
	"sort"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	loki "github.com/pachyderm/pachyderm/v2/src/internal/lokiutil/client"
)

const (
	// maxLogMessages is the maximum number of log messages the loki server
	// will send us, it will error if this is made higher.
	maxLogMessages = 5000
)

func forEachLine(resp loki.QueryResponse, f func(t time.Time, line string) error) error {
	// sort and display entries
	streams, ok := resp.Data.Result.(loki.Streams)
	if !ok {
		return errors.Errorf("resp.Data.Result must be of type loghttp.Streams to call ForEachStream on it")
	}
	allEntries := make([]streamEntryPair, 0)

	for _, s := range streams {
		for _, e := range s.Entries {
			allEntries = append(allEntries, streamEntryPair{
				entry:  e,
				labels: s.Labels,
			})
		}
	}

	sort.Slice(allEntries, func(i, j int) bool { return allEntries[i].entry.Timestamp.Before(allEntries[j].entry.Timestamp) })

	for _, e := range allEntries {
		if err := f(e.entry.Timestamp, e.entry.Line); err != nil {
			if errors.Is(err, errutil.ErrBreak) {
				return nil
			}
			return err
		}
	}
	return nil
}

func wrapEntryIfNotJSON(entry streamEntry) (streamEntry, error) {
	err := json.Unmarshal([]byte(entry.Log), &map[string]interface{}{})
	if err == nil {
		return entry, nil
	}
	// Entries from Klog may not be in JSON and so should be wrapped to maintain structured logging.
	wrappedLog := logWrapper{}
	wrappedLog.Message = entry.Log
	wrappedLogJSON, err := json.Marshal(wrappedLog)
	if err != nil {
		return streamEntry{}, errors.EnsureStack(err)
	}
	entry.Log = string(wrappedLogJSON)
	return entry, nil
}

// QueryRange calls QueryRange on the passed loki.Client and calls f with each logline.
func QueryRange(ctx context.Context, c *loki.Client, queryStr string, from, through time.Time, follow bool, f func(t time.Time, line string) error) error {
	for {
		resp, err := c.QueryRange(ctx, queryStr, maxLogMessages, from, through, "FORWARD", 0, 0, true)
		if err != nil {
			return errors.EnsureStack(err)
		}
		nMsgs := 0
		if err := forEachLine(*resp, func(t time.Time, line string) error {
			from = t
			nMsgs++
			entry := streamEntry{}
			err := json.Unmarshal([]byte(line), &entry)
			if err != nil {
				return errors.EnsureStack(err)
			}
			entry, err = wrapEntryIfNotJSON(entry)
			if err != nil {
				return errors.EnsureStack(err)
			}
			return f(t, entry.Log)
		}); err != nil {
			return err
		}
		if !follow && nMsgs < maxLogMessages {
			return nil
		}
		from = from.Add(time.Nanosecond)
	}
}

type streamEntry struct {
	Log    string    `json:"log"`
	Stream string    `json:"stream"`
	Time   time.Time `json:"time"`
}

type streamEntryPair struct {
	entry  loki.Entry
	labels loki.LabelSet
}

type logWrapper struct {
	Message string `json:"message"`
}
