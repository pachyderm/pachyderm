package pretty

import (
	"fmt"
	"io"

	"github.com/pachyderm/pachyderm/src/pps"
)

func PrintJobHeader(w io.Writer) {
	fmt.Fprint(w, "ID\tOUTPUT\tSTATE\t\n")
}

func PrintJobInfo(w io.Writer, jobInfo *pps.JobInfo) {
	fmt.Fprintf(w, "%s\t", jobInfo.Job.ID)
	if jobInfo.OutputCommit != nil {
		fmt.Fprintf(w, "%s/%s\t", jobInfo.OutputCommit.Repo.Name, jobInfo.OutputCommit.ID)
	} else {
		fmt.Fprintf(w, "-\t")
	}
	fmt.Fprintf(w, "%s\t\n", jobInfo.State.String())
}

func PrintPipelineHeader(w io.Writer) {
	fmt.Fprint(w, "NAME\tINPUT\tOUTPUT\t\n")
}

func PrintPipelineInfo(w io.Writer, pipelineInfo *pps.PipelineInfo) {
	fmt.Fprintf(w, "%s\t", pipelineInfo.Pipeline.Name)
	for i, input := range pipelineInfo.Inputs {
		fmt.Fprintf(w, "%s", input.Repo.Name)
		if i == len(pipelineInfo.Inputs)-1 {
			fmt.Fprintf(w, "\t")
		} else {
			fmt.Fprintf(w, ", ")
		}
	}
	if pipelineInfo.OutputRepo != nil {
		fmt.Fprintf(w, "%s\t\n", pipelineInfo.OutputRepo.Name)
	} else {
		fmt.Fprintf(w, "\t\n")
	}
}
