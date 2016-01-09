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
	fmt.Fprintf(w, "%s\t", jobInfo.Job.Id)
	fmt.Fprintf(w, "\t")
	if jobInfo.OutputCommit != nil {
		fmt.Fprintf(w, "%s/%s\t", jobInfo.OutputCommit.Repo.Name, jobInfo.OutputCommit.Id)
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
	for i, repo := range pipelineInfo.InputRepo {
		fmt.Fprintf(w, "%s", repo.Name)
		if i == len(pipelineInfo.InputRepo)-1 {
			fmt.Fprintf(w, "\t")
		} else {
			fmt.Fprintf(w, ", ")
		}
	}
	fmt.Fprintf(w, "%s\t\n", pipelineInfo.OutputRepo.Name)
}
