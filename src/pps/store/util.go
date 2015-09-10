package store

import (
	"sort"

	"go.pedge.io/proto/time"

	"github.com/pachyderm/pachyderm/src/pps"
)

func getPfsCommitMappingLatestInMemory(
	pfsCommitMappings []*pps.PfsCommitMapping,
	inputRepositoryName string,
) (*pps.PfsCommitMapping, error) {
	var filteredPfsCommitMappings []*pps.PfsCommitMapping
	for _, pfsCommitMapping := range pfsCommitMappings {
		if pfsCommitMapping.InputRepository == inputRepositoryName {
			filteredPfsCommitMappings = append(filteredPfsCommitMappings, pfsCommitMapping)
		}
	}
	if len(filteredPfsCommitMappings) == 0 {
		return nil, nil
	}
	sort.Sort(sortByTimestamp(filteredPfsCommitMappings))
	return filteredPfsCommitMappings[len(filteredPfsCommitMappings)-1], nil
}

type sortByTimestamp []*pps.PfsCommitMapping

func (s sortByTimestamp) Len() int          { return len(s) }
func (s sortByTimestamp) Swap(i int, j int) { s[i], s[j] = s[j], s[i] }
func (s sortByTimestamp) Less(i int, j int) bool {
	return prototime.TimestampLess(s[i].Timestamp, s[j].Timestamp)
}
