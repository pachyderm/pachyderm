package pretty

import (
	"fmt"
	"io"
	"strings"
	//"time"

	//"go.pedge.io/proto/time"

	//"github.com/docker/docker/pkg/units"
	"github.com/pachyderm/pachyderm/src/pps"
)

func PrintJobHeader(w io.Writer) {
	fmt.Fprint(w, "ID\tINPUT\tOUTPUT\tIMAGE\tCOMMAND\t\n")
}

func PrintJobInfo(w io.Writer, jobInfo *pps.JobInfo) {
	fmt.Fprintf(w, "%s\t", jobInfo.Job.Id)
	for i, commit := range jobInfo.InputCommit {
		fmt.Fprintf(w, "%s/%s", commit.Repo.Name, commit.Id)
		if i == len(jobInfo.InputCommit)-1 {
			fmt.Fprintf(w, "\t")
		} else {
			fmt.Fprintf(w, ", ")
		}
	}
	fmt.Fprintf(w, "\t")
	if jobInfo.OutputCommit != nil {
		fmt.Fprintf(w, "%s/%s\t", jobInfo.OutputCommit.Repo.Name, jobInfo.OutputCommit.Id)
	} else {
		fmt.Fprintf(w, "-\t")
	}
	fmt.Fprintf(w, "%s\t", jobInfo.Transform.Image)
	fmt.Fprintf(w, "%s\t\n", strings.Join(jobInfo.Transform.Cmd, " "))
}

func PrintPipelineHeader(w io.Writer) {
	fmt.Fprint(w, "NAME\tINPUT\tOUTPUT\tIMAGE\tCOMMAND\t\n")
}

func PrintPipelineInfo(w io.Writer, pipelineInfo *pps.PipelineInfo) {
	fmt.Fprintf(w, "%s\t", pipelineInfo.Pipeline.Name)
	for i, repo := range pipelineInfo.InputRepo {
		fmt.Fprintf(w, "%s", repo.Name)
		if i == len(pipelineInfo.InputRepo)-1 {
			fmt.Fprintf(w, "\t")
		} else {
			fmt.Fprintf(w, ", ")
		}
	}
	fmt.Fprintf(w, "%s\t", pipelineInfo.OutputRepo.Name)
	fmt.Fprintf(w, "%s\t", pipelineInfo.Transform.Image)
	fmt.Fprintf(w, "%s\t\n", strings.Join(pipelineInfo.Transform.Cmd, " "))
}

type uint64Slice []uint64

func (s uint64Slice) Len() int           { return len(s) }
func (s uint64Slice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s uint64Slice) Less(i, j int) bool { return s[i] < s[j] }
