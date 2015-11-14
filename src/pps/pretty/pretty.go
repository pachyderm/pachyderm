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
	fmt.Fprint(w, "ID\tINPUT\tOUTPUT\tPIPELINE\tIMAGE\tCOMMAND\tSTATUS\t\n")
}

func PrintJobInfo(w io.Writer, jobInfo *pps.JobInfo) {
	fmt.Fprintf(w, "%s\t", jobInfo.Job.Id)
	fmt.Fprintf(w, "%s/%s\t", jobInfo.InputCommit.Repo.Name, jobInfo.InputCommit.Id)
	if jobInfo.OutputCommit != nil {
		fmt.Fprintf(w, "%s/%s\t", jobInfo.OutputCommit.Repo.Name, jobInfo.OutputCommit.Id)
	} else {
		fmt.Fprintf(w, "-\t")
	}
	if jobInfo.GetPipeline() != nil {
		fmt.Fprintf(w, "%s\t", jobInfo.GetPipeline().Name)
	} else {
		fmt.Fprintf(w, "-\t")
	}
	if jobInfo.GetTransform() != nil {
		fmt.Fprintf(w, "%s\t", jobInfo.GetTransform().Image)
		fmt.Fprintf(w, "%s\t", strings.Join(jobInfo.GetTransform().Cmd, " "))
	} else {
		fmt.Fprintf(w, "-\t")
		fmt.Fprintf(w, "-\t")
	}
	if len(jobInfo.JobStatus) > 0 {
		fmt.Fprintf(w, "%s\t\n", jobInfo.JobStatus[0].Message)
	} else {
		fmt.Fprintf(w, "-\t\n")
	}
}

func PrintPipelineHeader(w io.Writer) {
	fmt.Fprint(w, "NAME\tINPUT\tOUTPUT\tIMAGE\tCOMMAND\t\n")
}

func PrintPipelineInfo(w io.Writer, pipelineInfo *pps.PipelineInfo) {
	fmt.Fprintf(w, "%s\t", pipelineInfo.Pipeline.Name)
	fmt.Fprintf(w, "%s\t", pipelineInfo.Input.Name)
	fmt.Fprintf(w, "%s\t", pipelineInfo.Output.Name)
	fmt.Fprintf(w, "%s\t", pipelineInfo.Transform.Image)
	fmt.Fprintf(w, "%s\t\n", strings.Join(pipelineInfo.Transform.Cmd, " "))
}

type uint64Slice []uint64

func (s uint64Slice) Len() int           { return len(s) }
func (s uint64Slice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s uint64Slice) Less(i, j int) bool { return s[i] < s[j] }
