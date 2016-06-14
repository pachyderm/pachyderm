package pretty

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/Jeffail/gabs"
	"github.com/fatih/color"
	ppsclient "github.com/pachyderm/pachyderm/src/client/pps"
)

func PrintJobHeader(w io.Writer) {
	fmt.Fprint(w, "ID\tOUTPUT\tSTATE\t\n")
}

func PrintJobInfo(w io.Writer, jobInfo *ppsclient.JobInfo) {
	fmt.Fprintf(w, "%s\t", jobInfo.Job.ID)
	if jobInfo.OutputCommit != nil {
		fmt.Fprintf(w, "%s/%s\t", jobInfo.OutputCommit.Repo.Name, jobInfo.OutputCommit.ID)
	} else {
		fmt.Fprintf(w, "-\t")
	}
	fmt.Fprintf(w, "%s\t\n", jobState(jobInfo))
}

func PrintPipelineHeader(w io.Writer) {
	fmt.Fprint(w, "NAME\tINPUT\tOUTPUT\tSTATE\t\n")
}

func PrintPipelineInfo(w io.Writer, pipelineInfo *ppsclient.PipelineInfo) {
	fmt.Fprintf(w, "%s\t", pipelineInfo.Pipeline.Name)
	if len(pipelineInfo.Inputs) == 0 {
		fmt.Fprintf(w, "\t")
	} else {
		for i, input := range pipelineInfo.Inputs {
			fmt.Fprintf(w, "%s", input.Repo.Name)
			if i == len(pipelineInfo.Inputs)-1 {
				fmt.Fprintf(w, "\t")
			} else {
				fmt.Fprintf(w, ", ")
			}
		}
	}
	if pipelineInfo.OutputRepo != nil {
		fmt.Fprintf(w, "%s\t", pipelineInfo.OutputRepo.Name)
	} else {
		fmt.Fprintf(w, "\t")
	}
	fmt.Fprintf(w, "%s\t\n", pipelineState(pipelineInfo))
}

func PrintDetailedJobInfo(jobInfo *ppsclient.JobInfo) {
	bytes, err := json.Marshal(jobInfo)
	if err != nil {
		fmt.Println(err.Error())
	}

	obj, err := gabs.ParseJSON(bytes)
	if err != nil {
		fmt.Println(err.Error())
	}

	// state is an integer; we want to print a string
	_, err = obj.Set(ppsclient.JobState_name[int32(jobInfo.State)], "state")
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println(obj.StringIndent("", "    "))
}

func PrintDetailedPipelineInfo(pipelineInfo *ppsclient.PipelineInfo) {
	bytes, err := json.Marshal(pipelineInfo)
	if err != nil {
		fmt.Println(err.Error())
	}

	obj, err := gabs.ParseJSON(bytes)
	if err != nil {
		fmt.Println(err.Error())
	}

	// state is an integer; we want to print a string
	_, err = obj.Set(ppsclient.PipelineState_name[int32(pipelineInfo.State)], "state")
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println(obj.StringIndent("", "    "))
}

func jobState(jobInfo *ppsclient.JobInfo) string {
	switch jobInfo.State {
	case ppsclient.JobState_JOB_PULLING:
		return color.New(color.FgYellow).SprintFunc()("pulling")
	case ppsclient.JobState_JOB_RUNNING:
		return color.New(color.FgYellow).SprintFunc()("running")
	case ppsclient.JobState_JOB_FAILURE:
		return color.New(color.FgRed).SprintFunc()("failure")
	case ppsclient.JobState_JOB_SUCCESS:
		return color.New(color.FgGreen).SprintFunc()("success")
	}
	return "-"
}

func pipelineState(pipelineInfo *ppsclient.PipelineInfo) string {
	switch pipelineInfo.State {
	case ppsclient.PipelineState_PIPELINE_IDLE:
		return color.New(color.FgYellow).SprintFunc()("idle")
	case ppsclient.PipelineState_PIPELINE_RUNNING:
		return color.New(color.FgGreen).SprintFunc()("running")
	case ppsclient.PipelineState_PIPELINE_RESTARTING:
		return color.New(color.FgYellow).SprintFunc()("restarting")
	case ppsclient.PipelineState_PIPELINE_FAILURE:
		return color.New(color.FgRed).SprintFunc()("failure")
	}
	return "-"
}
