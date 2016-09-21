package pretty

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
	"text/template"

	"github.com/fatih/color"
	ppsclient "github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/pretty"
)

// PrintJobHeader prints a job header.
func PrintJobHeader(w io.Writer) {
	fmt.Fprint(w, "ID\tOUTPUT\tSTARTED\tDURATION\tSTATE\t\n")
}

// PrintJobInfo pretty-prints job info.
func PrintJobInfo(w io.Writer, jobInfo *ppsclient.JobInfo) {
	fmt.Fprintf(w, "%s\t", jobInfo.Job.ID)
	if jobInfo.OutputCommit != nil {
		fmt.Fprintf(w, "%s/%s\t", jobInfo.OutputCommit.Repo.Name, jobInfo.OutputCommit.ID)
	} else {
		fmt.Fprintf(w, "-\t")
	}
	fmt.Fprintf(w, "%s\t", pretty.Ago(jobInfo.Started))
	if jobInfo.Finished != nil {
		fmt.Fprintf(w, "%s\t", pretty.Duration(jobInfo.Started, jobInfo.Finished))
	} else {
		fmt.Fprintf(w, "-\t")
	}
	fmt.Fprintf(w, "%s\t\n", jobState(jobInfo.State))
}

// PrintPipelineHeader prints a pipeline header.
func PrintPipelineHeader(w io.Writer) {
	fmt.Fprint(w, "NAME\tINPUT\tOUTPUT\tSTATE\t\n")
}

// PrintPipelineInfo pretty-prints pipeline info.
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
	fmt.Fprintf(w, "%s\t\n", pipelineState(pipelineInfo.State))
}

// PrintJobInputHeader pretty prints a job input header.
func PrintJobInputHeader(w io.Writer) {
	fmt.Fprint(w, "NAME\tCOMMIT\tPARTITION\tINCREMENTAL\t\n")
}

// PrintJobInput pretty-prints a job input.
func PrintJobInput(w io.Writer, jobInput *ppsclient.JobInput) {
	fmt.Fprintf(w, "%s\t", jobInput.Commit.Repo.Name)
	fmt.Fprintf(w, "%s\t", jobInput.Commit.ID)
	fmt.Fprintf(w, "%s\t", jobInput.Method.Partition)
	fmt.Fprintf(w, "%t\t\n", jobInput.Method.Incremental)
}

// PrintPipelineInputHeader prints a pipeline input header.
func PrintPipelineInputHeader(w io.Writer) {
	fmt.Fprint(w, "NAME\tPARTITION\tINCREMENTAL\t\n")
}

// PrintPipelineInput pretty-prints a pipeline input.
func PrintPipelineInput(w io.Writer, pipelineInput *ppsclient.PipelineInput) {
	fmt.Fprintf(w, "%s\t", pipelineInput.Repo.Name)
	fmt.Fprintf(w, "%s\t", pipelineInput.Method.Partition)
	fmt.Fprintf(w, "%s\t\n", pipelineInput.Method.Incremental)
}

// PrintJobCountsHeader prints a job counts header.
func PrintJobCountsHeader(w io.Writer) {
	fmt.Fprintf(w, strings.ToUpper(jobState(ppsclient.JobState_JOB_PULLING))+"\t")
	fmt.Fprintf(w, strings.ToUpper(jobState(ppsclient.JobState_JOB_RUNNING))+"\t")
	fmt.Fprintf(w, strings.ToUpper(jobState(ppsclient.JobState_JOB_FAILURE))+"\t")
	fmt.Fprintf(w, strings.ToUpper(jobState(ppsclient.JobState_JOB_SUCCESS))+"\t\n")
}

// PrintDetailedJobInfo pretty-prints detailed job info.
func PrintDetailedJobInfo(jobInfo *ppsclient.JobInfo) error {
	template, err := template.New("JobInfo").Funcs(funcMap).Parse(
		`ID: {{.Job.ID}} {{if .ParentJob}}
Parent: {{.ParentJob.ID}} {{end}}
Started: {{prettyAgo .Started}} {{if .Finished}}
Duration: {{prettyDuration .Started .Finished}} {{end}}
State: {{jobState .State}}
ParallelismSpec: {{.ParallelismSpec}}
Inputs:
{{jobInputs .}}Transform:
{{prettyTransform .Transform}}
`)
	if err != nil {
		return err
	}
	err = template.Execute(os.Stdout, jobInfo)
	if err != nil {
		return err
	}
	return nil
}

// PrintDetailedPipelineInfo pretty-prints detailed pipeline info.
func PrintDetailedPipelineInfo(pipelineInfo *ppsclient.PipelineInfo) error {
	template, err := template.New("PipelineInfo").Funcs(funcMap).Parse(
		`Name: {{.Pipeline.Name}}
Created: {{prettyAgo .CreatedAt}}
State: {{pipelineState .State}}
ParallelismSpec: {{.ParallelismSpec}}
Inputs:
{{pipelineInputs .}}Transform:
{{prettyTransform .Transform}}
{{if .RecentError}} Recent Error: {{.RecentError}} {{end}}
Job Counts:
{{jobCounts .JobCounts}}
`)
	if err != nil {
		return err
	}
	err = template.Execute(os.Stdout, pipelineInfo)
	if err != nil {
		return err
	}
	return nil
}

func jobState(jobState ppsclient.JobState) string {
	switch jobState {
	case ppsclient.JobState_JOB_PULLING:
		return color.New(color.FgYellow).SprintFunc()("pulling")
	case ppsclient.JobState_JOB_RUNNING:
		return color.New(color.FgYellow).SprintFunc()("running")
	case ppsclient.JobState_JOB_FAILURE:
		return color.New(color.FgRed).SprintFunc()("failure")
	case ppsclient.JobState_JOB_SUCCESS:
		return color.New(color.FgGreen).SprintFunc()("success")
	case ppsclient.JobState_JOB_EMPTY:
		return color.New(color.FgGreen).SprintFunc()("empty")
	}
	return "-"
}

func pipelineState(pipelineState ppsclient.PipelineState) string {
	switch pipelineState {
	case ppsclient.PipelineState_PIPELINE_IDLE:
		return color.New(color.FgYellow).SprintFunc()("idle")
	case ppsclient.PipelineState_PIPELINE_RUNNING:
		return color.New(color.FgGreen).SprintFunc()("running")
	case ppsclient.PipelineState_PIPELINE_RESTARTING:
		return color.New(color.FgYellow).SprintFunc()("restarting")
	case ppsclient.PipelineState_PIPELINE_FAILURE:
		return color.New(color.FgRed).SprintFunc()("failure")
	case ppsclient.PipelineState_PIPELINE_STOPPED:
		return color.New(color.FgYellow).SprintFunc()("stopped")
	}
	return "-"
}

func jobInputs(jobInfo *ppsclient.JobInfo) string {
	var buffer bytes.Buffer
	writer := tabwriter.NewWriter(&buffer, 20, 1, 3, ' ', 0)
	PrintJobInputHeader(writer)
	for _, input := range jobInfo.Inputs {
		PrintJobInput(writer, input)
	}
	// can't error because buffer can't error on Write
	writer.Flush()
	return buffer.String()
}

func pipelineInputs(pipelineInfo *ppsclient.PipelineInfo) string {
	var buffer bytes.Buffer
	writer := tabwriter.NewWriter(&buffer, 20, 1, 3, ' ', 0)
	PrintPipelineInputHeader(writer)
	for _, input := range pipelineInfo.Inputs {
		PrintPipelineInput(writer, input)
	}
	// can't error because buffer can't error on Write
	writer.Flush()
	return buffer.String()
}

func jobCounts(counts map[int32]int32) string {
	var buffer bytes.Buffer
	for i := int32(ppsclient.JobState_JOB_PULLING); i <= int32(ppsclient.JobState_JOB_SUCCESS); i++ {
		fmt.Fprintf(&buffer, "%s: %d\t", jobState(ppsclient.JobState(i)), counts[i])
	}
	return buffer.String()
}

func prettyTransform(transform *ppsclient.Transform) (string, error) {
	result, err := json.MarshalIndent(transform, "", "  ")
	if err != nil {
		return "", err
	}
	return pretty.UnescapeHTML(string(result)), nil
}

var funcMap = template.FuncMap{
	"pipelineState":   pipelineState,
	"jobState":        jobState,
	"pipelineInputs":  pipelineInputs,
	"jobInputs":       jobInputs,
	"prettyAgo":       pretty.Ago,
	"prettyDuration":  pretty.Duration,
	"jobCounts":       jobCounts,
	"prettyTransform": prettyTransform,
}
