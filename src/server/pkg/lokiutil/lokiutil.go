package lokiutil

import (
	"sort"

	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/server/pkg/errutil"

	"github.com/grafana/loki/pkg/loghttp"
)

func ForEachLine(resp *loghttp.QueryResponse, f func(line string) error) error {
	// sort and display entries
	streams, ok := resp.Data.Result.(loghttp.Streams)
	if !ok {
		return errors.Errorf("respo.ResultValue must be of type loghttp.Streams to call ForEachStream on it")
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
		if err := f(e.entry.Line); err != nil {
			if err == errutil.ErrBreak {
				return nil
			}
			return err
		}
	}
	return nil
}

type streamEntryPair struct {
	entry  loghttp.Entry
	labels loghttp.LabelSet
}
