package worker

import (
	"encoding/base64"
	"encoding/hex"
	"github.com/pachyderm/pachyderm/src/client/pps"
)

// MatchDatum checks if a datum matches a filter.  To match each string in
// filter must correspond match at least 1 datum's Path or Hash. Order of
// filter and data is irrelevant.
func MatchDatum(filter []string, data []*pps.Datum) bool {
	// All paths in request.DataFilters must appear somewhere in the log
	// line's inputs, or it's filtered
	matchesData := true
dataFilters:
	for _, dataFilter := range filter {
		for _, datum := range data {
			if dataFilter == datum.Path ||
				dataFilter == base64.StdEncoding.EncodeToString(datum.Hash) ||
				dataFilter == hex.EncodeToString(datum.Hash) {
				continue dataFilters // Found, move to next filter
			}
		}
		matchesData = false
		break
	}
	return matchesData
}
