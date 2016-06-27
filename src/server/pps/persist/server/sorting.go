package server

import (
	"sort"

	"go.pedge.io/proto/time"

	"github.com/pachyderm/pachyderm/src/server/pps/persist"
)

// TODO: this should be a call through the actual persist storage
//
// This does not work:
//
//     func(term gorethink.Term) gorethink.Term {
//         return term.OrderBy(gorethink.Desc("created_at"))
//     }

func sortJobInfosByTimestampDesc(s []*persist.JobInfo) {
	sort.Sort(jobInfosByTimestampDesc(s))
}

type jobInfosByTimestampDesc []*persist.JobInfo

func (s jobInfosByTimestampDesc) Len() int          { return len(s) }
func (s jobInfosByTimestampDesc) Swap(i int, j int) { s[i], s[j] = s[j], s[i] }
func (s jobInfosByTimestampDesc) Less(i int, j int) bool {
	return prototime.TimestampLess(s[j].Started, s[i].Started)
}

func sortPipelineInfosByTimestampDesc(s []*persist.PipelineInfo) {
	sort.Sort(pipelineInfosByTimestampDesc(s))
}

type pipelineInfosByTimestampDesc []*persist.PipelineInfo

func (s pipelineInfosByTimestampDesc) Len() int          { return len(s) }
func (s pipelineInfosByTimestampDesc) Swap(i int, j int) { s[i], s[j] = s[j], s[i] }
func (s pipelineInfosByTimestampDesc) Less(i int, j int) bool {
	return prototime.TimestampLess(s[j].CreatedAt, s[i].CreatedAt)
}
