package lokiutil

import (
	"sort"

	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/server/pkg/errutil"

	"github.com/grafana/loki/pkg/loghttp"
)

func ForEachLine(resp *loghttp.QueryResponse, f func(line string) error) error {
	// common := commonLabels(streams)

	// // Remove the labels we want to show from common
	// if len(q.ShowLabelsKey) > 0 {
	// 	common = matchLabels(false, common, q.ShowLabelsKey)
	// }

	// if len(common) > 0 && !q.Quiet {
	// 	log.Println("Common labels:", color.RedString(common.String()))
	// }

	// if len(q.IgnoreLabelsKey) > 0 && !q.Quiet {
	// 	log.Println("Ignoring labels key:", color.RedString(strings.Join(q.IgnoreLabelsKey, ",")))
	// }

	// // Remove ignored and common labels from the cached labels and
	// // calculate the max labels length
	// maxLabelsLen := q.FixedLabelsLen
	// for i, s := range streams {
	// 	// Remove common labels
	// 	ls := subtract(s.Labels, common)

	// 	// Remove ignored labels
	// 	if len(q.IgnoreLabelsKey) > 0 {
	// 		ls = matchLabels(false, ls, q.IgnoreLabelsKey)
	// 	}

	// 	// Overwrite existing Labels
	// 	streams[i].Labels = ls

	// 	// Update max labels length
	// 	len := len(ls.String())
	// 	if maxLabelsLen < len {
	// 		maxLabelsLen = len
	// 	}
	// }

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
