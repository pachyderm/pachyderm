package worker

import (
	"github.com/pachyderm/pachyderm/src/client/pps"
)

func MatchDatum(filter []string, data []*pps.Datum) bool {
	// All paths in request.DataFilters must appear somewhere in the log
	// line's inputs, or it's filtered
	matchesData := true
dataFilters:
	for _, dataFilter := range filter {
		for _, datum := range data {
			if dataFilter == datum.Path || dataFilter == string(datum.Hash) {
				continue dataFilters // Found, move to next filter
			}
		}
		matchesData = false
		break
	}
	return matchesData
}
