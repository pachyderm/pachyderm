package pretty

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"

	"github.com/docker/go-units"
	"github.com/fatih/color"
	"github.com/gogo/protobuf/types"
	"github.com/juju/ansiterm"
	"github.com/pachyderm/pachyderm/src/client"
	pfsclient "github.com/pachyderm/pachyderm/src/client/pfs"
	ppsclient "github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/pretty"
)

const (
	// PipelineHeader is the header for pipelines.
	PipelineHeader = "NAME\tVERSION\tINPUT\tCREATED\tSTATE / LAST JOB\tDESCRIPTION\t\n"
	// JobHeader is the header for jobs
	JobHeader = "ID\tPIPELINE\tSTARTED\tDURATION\tRESTART\tPROGRESS\tDL\tUL\tSTATE\t\n"
	// DatumHeader is the header for datums
	DatumHeader = "ID\tSTATUS\tTIME\t\n"
	// jobReasonLen is the amount of the job reason that we print
	jobReasonLen = 25
)

func safeTrim(s string, l int) string {
	if len(s) < l {
		return s
	}
	return strings.TrimSpace(s[:l]) + "..."
}

// PrintJobInfo pretty-prints job info.
func PrintJobInfo(w io.Writer, jobInfo *ppsclient.JobInfo, fullTimestamps bool) {
	fmt.Fprintf(w, "%s\t", jobInfo.Job.ID)
	fmt.Fprintf(w, "%s\t", jobInfo.Pipeline.Name)
	if fullTimestamps {
		fmt.Fprintf(w, "%s\t", jobInfo.Started.String())
	} else {
		fmt.Fprintf(w, "%s\t", pretty.Ago(jobInfo.Started))
	}
	if jobInfo.Finished != nil {
		fmt.Fprintf(w, "%s\t", pretty.TimeDifference(jobInfo.Started, jobInfo.Finished))
	} else {
		fmt.Fprintf(w, "-\t")
	}
	fmt.Fprintf(w, "%d\t", jobInfo.Restart)
	fmt.Fprintf(w, "%s\t", Progress(jobInfo))
	fmt.Fprintf(w, "%s\t", pretty.Size(jobInfo.Stats.DownloadBytes))
	fmt.Fprintf(w, "%s\t", pretty.Size(jobInfo.Stats.UploadBytes))
	if jobInfo.State == ppsclient.JobState_JOB_FAILURE {
		fmt.Fprintf(w, "%s: %s\t", JobState(jobInfo.State), safeTrim(jobInfo.Reason, jobReasonLen))
	} else {
		fmt.Fprintf(w, "%s\t", JobState(jobInfo.State))
	}
	fmt.Fprintln(w)
}

// PrintPipelineInfo pretty-prints pipeline info.
func PrintPipelineInfo(w io.Writer, pipelineInfo *ppsclient.PipelineInfo, fullTimestamps bool) {
	fmt.Fprintf(w, "%s\t", pipelineInfo.Pipeline.Name)
	fmt.Fprintf(w, "%d\t", pipelineInfo.Version)
	fmt.Fprintf(w, "%s\t", ShorthandInput(pipelineInfo.Input))
	if fullTimestamps {
		fmt.Fprintf(w, "%s\t", pipelineInfo.CreatedAt.String())
	} else {
		fmt.Fprintf(w, "%s\t", pretty.Ago(pipelineInfo.CreatedAt))
	}
	fmt.Fprintf(w, "%s / %s\t", pipelineState(pipelineInfo.State), JobState(pipelineInfo.LastJobState))
	fmt.Fprintf(w, "%s\t", pipelineInfo.Description)
	fmt.Fprintln(w)
}

// PrintWorkerStatusHeader pretty prints a worker status header.
func PrintWorkerStatusHeader(w io.Writer) {
	fmt.Fprint(w, "WORKER\tJOB\tDATUM\tSTARTED\tQUEUE\t\n")
}

// PrintWorkerStatus pretty prints a worker status.
func PrintWorkerStatus(w io.Writer, workerStatus *ppsclient.WorkerStatus, fullTimestamps bool) {
	fmt.Fprintf(w, "%s\t", workerStatus.WorkerID)
	fmt.Fprintf(w, "%s\t", workerStatus.JobID)
	for _, datum := range workerStatus.Data {
		fmt.Fprintf(w, datum.Path)
	}
	fmt.Fprintf(w, "\t")
	if fullTimestamps {
		fmt.Fprintf(w, "%s\t", workerStatus.Started.String())
	} else {
		fmt.Fprintf(w, "%s\t", pretty.Ago(workerStatus.Started))
	}
	fmt.Fprintf(w, "%d\t", workerStatus.QueueSize)
	fmt.Fprintln(w)
}

// PrintableJobInfo is a wrapper around JobInfo containing any formatting options
// used within the template to conditionally print information.
type PrintableJobInfo struct {
	*ppsclient.JobInfo
	FullTimestamps bool
}

// NewPrintableJobInfo constructs a PrintableJobInfo from just a JobInfo.
func NewPrintableJobInfo(ji *ppsclient.JobInfo) *PrintableJobInfo {
	return &PrintableJobInfo{
		JobInfo: ji,
	}
}

// PrintDetailedJobInfo pretty-prints detailed job info.
func PrintDetailedJobInfo(jobInfo *PrintableJobInfo) error {
	template, err := template.New("JobInfo").Funcs(funcMap).Parse(
		`ID: {{.Job.ID}} {{if .Pipeline}}
Pipeline: {{.Pipeline.Name}} {{end}} {{if .ParentJob}}
Parent: {{.ParentJob.ID}} {{end}}{{if .FullTimestamps}}
Started: {{.Started}}{{else}}
Started: {{prettyAgo .Started}} {{end}}{{if .Finished}}
Duration: {{prettyTimeDifference .Started .Finished}} {{end}}
State: {{jobState .State}}
Reason: {{.Reason}}
Processed: {{.DataProcessed}}
Failed: {{.DataFailed}}
Skipped: {{.DataSkipped}}
Recovered: {{.DataRecovered}}
Total: {{.DataTotal}}
Data Downloaded: {{prettySize .Stats.DownloadBytes}}
Data Uploaded: {{prettySize .Stats.UploadBytes}}
Download Time: {{prettyDuration .Stats.DownloadTime}}
Process Time: {{prettyDuration .Stats.ProcessTime}}
Upload Time: {{prettyDuration .Stats.UploadTime}}
Datum Timeout: {{.DatumTimeout}}
Job Timeout: {{.JobTimeout}}
Worker Status:
{{workerStatus .}}Restarts: {{.Restart}}
ParallelismSpec: {{.ParallelismSpec}}
{{ if .ResourceRequests }}ResourceRequests:
  CPU: {{ .ResourceRequests.Cpu }}
  Memory: {{ .ResourceRequests.Memory }} {{end}}
{{ if .ResourceLimits }}ResourceLimits:
  CPU: {{ .ResourceLimits.Cpu }}
  Memory: {{ .ResourceLimits.Memory }}
  {{ if .ResourceLimits.Gpu }}GPU:
    Type: {{ .ResourceLimits.Gpu.Type }}
    Number: {{ .ResourceLimits.Gpu.Number }} {{end}} {{end}}
{{ if .Service }}Service:
	{{ if .Service.InternalPort }}InternalPort: {{ .Service.InternalPort }} {{end}}
	{{ if .Service.ExternalPort }}ExternalPort: {{ .Service.ExternalPort }} {{end}} {{end}}Input:
{{jobInput .}}
Transform:
{{prettyTransform .Transform}} {{if .OutputCommit}}
Output Commit: {{.OutputCommit.ID}} {{end}} {{ if .StatsCommit }}
Stats Commit: {{.StatsCommit.ID}} {{end}} {{ if .Egress }}
Egress: {{.Egress.URL}} {{end}}
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

// PrintablePipelineInfo is a wrapper around PipelinInfo containing any formatting options
// used within the template to conditionally print information.
type PrintablePipelineInfo struct {
	*ppsclient.PipelineInfo
	FullTimestamps bool
}

// NewPrintablePipelineInfo constructs a PrintablePipelineInfo from just a PipelineInfo.
func NewPrintablePipelineInfo(pi *ppsclient.PipelineInfo) *PrintablePipelineInfo {
	return &PrintablePipelineInfo{
		PipelineInfo: pi,
	}
}

// PrintDetailedPipelineInfo pretty-prints detailed pipeline info.
func PrintDetailedPipelineInfo(pipelineInfo *PrintablePipelineInfo) error {
	template, err := template.New("PipelineInfo").Funcs(funcMap).Parse(
		`Name: {{.Pipeline.Name}}{{if .Description}}
Description: {{.Description}}{{end}}{{if .FullTimestamps }}
Created: {{.CreatedAt}}{{ else }}
Created: {{prettyAgo .CreatedAt}} {{end}}
State: {{pipelineState .State}}
Stopped: {{ .Stopped }}
Reason: {{.Reason}}
Parallelism Spec: {{.ParallelismSpec}}
{{ if .ResourceRequests }}ResourceRequests:
  CPU: {{ .ResourceRequests.Cpu }}
  Memory: {{ .ResourceRequests.Memory }} {{end}}
{{ if .ResourceLimits }}ResourceLimits:
  CPU: {{ .ResourceLimits.Cpu }}
  Memory: {{ .ResourceLimits.Memory }}
  {{ if .ResourceLimits.Gpu }}GPU:
    Type: {{ .ResourceLimits.Gpu.Type }} 
    Number: {{ .ResourceLimits.Gpu.Number }} {{end}} {{end}}
Datum Timeout: {{.DatumTimeout}}
Job Timeout: {{.JobTimeout}}
Input:
{{pipelineInput .PipelineInfo}}
{{ if .GithookURL }}Githook URL: {{.GithookURL}} {{end}}
Output Branch: {{.OutputBranch}}
Transform:
{{prettyTransform .Transform}}
{{ if .Egress }}Egress: {{.Egress.URL}} {{end}}
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

// PrintDatumInfo pretty-prints file info.
// If recurse is false and directory size is 0, display "-" instead
// If fast is true and file size is 0, display "-" instead
func PrintDatumInfo(w io.Writer, datumInfo *ppsclient.DatumInfo) {
	totalTime := "-"
	if datumInfo.Stats != nil {
		totalTime = units.HumanDuration(client.GetDatumTotalTime(datumInfo.Stats))
	}
	fmt.Fprintf(w, "%s\t%s\t%s\t", datumInfo.Datum.ID, datumState(datumInfo.State), totalTime)
	fmt.Fprintln(w)
}

// PrintDetailedDatumInfo pretty-prints detailed info about a datum
func PrintDetailedDatumInfo(w io.Writer, datumInfo *ppsclient.DatumInfo) {
	fmt.Fprintf(w, "ID\t%s\n", datumInfo.Datum.ID)
	fmt.Fprintf(w, "Job ID\t%s\n", datumInfo.Datum.Job.ID)
	fmt.Fprintf(w, "State\t%s\n", datumInfo.State)
	fmt.Fprintf(w, "Data Downloaded\t%s\n", pretty.Size(datumInfo.Stats.DownloadBytes))
	fmt.Fprintf(w, "Data Uploaded\t%s\n", pretty.Size(datumInfo.Stats.UploadBytes))

	totalTime := client.GetDatumTotalTime(datumInfo.Stats).String()
	fmt.Fprintf(w, "Total Time\t%s\n", totalTime)

	var downloadTime string
	dl, err := types.DurationFromProto(datumInfo.Stats.DownloadTime)
	if err != nil {
		downloadTime = err.Error()
	} else {
		downloadTime = dl.String()
	}
	fmt.Fprintf(w, "Download Time\t%s\n", downloadTime)

	var procTime string
	proc, err := types.DurationFromProto(datumInfo.Stats.ProcessTime)
	if err != nil {
		procTime = err.Error()
	} else {
		procTime = proc.String()
	}
	fmt.Fprintf(w, "Process Time\t%s\n", procTime)

	var uploadTime string
	ul, err := types.DurationFromProto(datumInfo.Stats.UploadTime)
	if err != nil {
		uploadTime = err.Error()
	} else {
		uploadTime = ul.String()
	}
	fmt.Fprintf(w, "Upload Time\t%s\n", uploadTime)

	fmt.Fprintf(w, "PFS State:\n")
	tw := ansiterm.NewTabWriter(w, 10, 1, 3, ' ', 0)
	PrintFileHeader(tw)
	PrintFile(tw, datumInfo.PfsState)
	tw.Flush()
	fmt.Fprintf(w, "Inputs:\n")
	tw = ansiterm.NewTabWriter(w, 10, 1, 3, ' ', 0)
	PrintFileHeader(tw)
	for _, d := range datumInfo.Data {
		PrintFile(tw, d.File)
	}
	tw.Flush()
}

// PrintFileHeader prints the header for a pfs file.
func PrintFileHeader(w io.Writer) {
	fmt.Fprintf(w, "  REPO\tCOMMIT\tPATH\t\n")
}

// PrintFile values for a pfs file.
func PrintFile(w io.Writer, file *pfsclient.File) {
	fmt.Fprintf(w, "  %s\t%s\t%s\t\n", file.Commit.Repo.Name, file.Commit.ID, file.Path)
}

func datumState(datumState ppsclient.DatumState) string {
	switch datumState {
	case ppsclient.DatumState_SKIPPED:
		return color.New(color.FgYellow).SprintFunc()("skipped")
	case ppsclient.DatumState_FAILED:
		return color.New(color.FgRed).SprintFunc()("failed")
	case ppsclient.DatumState_RECOVERED:
		return color.New(color.FgYellow).SprintFunc()("recovered")
	case ppsclient.DatumState_SUCCESS:
		return color.New(color.FgGreen).SprintFunc()("success")
	}
	return "-"
}

// JobState returns the state of a job as a pretty printed string.
func JobState(jobState ppsclient.JobState) string {
	switch jobState {
	case ppsclient.JobState_JOB_STARTING:
		return color.New(color.FgYellow).SprintFunc()("starting")
	case ppsclient.JobState_JOB_RUNNING:
		return color.New(color.FgYellow).SprintFunc()("running")
	case ppsclient.JobState_JOB_MERGING:
		return color.New(color.FgYellow).SprintFunc()("merging")
	case ppsclient.JobState_JOB_FAILURE:
		return color.New(color.FgRed).SprintFunc()("failure")
	case ppsclient.JobState_JOB_SUCCESS:
		return color.New(color.FgGreen).SprintFunc()("success")
	case ppsclient.JobState_JOB_KILLED:
		return color.New(color.FgRed).SprintFunc()("killed")
	}
	return "-"
}

// Progress pretty prints the datum progress of a job.
func Progress(ji *ppsclient.JobInfo) string {
	if ji.DataRecovered != 0 {
		return fmt.Sprintf("%d + %d + %d / %d", ji.DataProcessed, ji.DataSkipped, ji.DataRecovered, ji.DataTotal)
	}
	return fmt.Sprintf("%d + %d / %d", ji.DataProcessed, ji.DataSkipped, ji.DataTotal)
}

func pipelineState(pipelineState ppsclient.PipelineState) string {
	switch pipelineState {
	case ppsclient.PipelineState_PIPELINE_STARTING:
		return color.New(color.FgYellow).SprintFunc()("starting")
	case ppsclient.PipelineState_PIPELINE_RUNNING:
		return color.New(color.FgGreen).SprintFunc()("running")
	case ppsclient.PipelineState_PIPELINE_RESTARTING:
		return color.New(color.FgYellow).SprintFunc()("restarting")
	case ppsclient.PipelineState_PIPELINE_FAILURE:
		return color.New(color.FgRed).SprintFunc()("failure")
	case ppsclient.PipelineState_PIPELINE_PAUSED:
		return color.New(color.FgYellow).SprintFunc()("paused")
	case ppsclient.PipelineState_PIPELINE_STANDBY:
		return color.New(color.FgYellow).SprintFunc()("standby")
	}
	return "-"
}

func jobInput(jobInfo PrintableJobInfo) string {
	if jobInfo.Input == nil {
		return ""
	}
	input, err := json.MarshalIndent(jobInfo.Input, "", "  ")
	if err != nil {
		panic(fmt.Errorf("error marshalling input: %+v", err))
	}
	return string(input) + "\n"
}

func workerStatus(jobInfo PrintableJobInfo) string {
	var buffer bytes.Buffer
	writer := ansiterm.NewTabWriter(&buffer, 20, 1, 3, ' ', 0)
	PrintWorkerStatusHeader(writer)
	for _, workerStatus := range jobInfo.WorkerStatus {
		PrintWorkerStatus(writer, workerStatus, jobInfo.FullTimestamps)
	}
	// can't error because buffer can't error on Write
	writer.Flush()
	return buffer.String()
}

func pipelineInput(pipelineInfo *ppsclient.PipelineInfo) string {
	if pipelineInfo.Input == nil {
		return ""
	}
	input, err := json.MarshalIndent(pipelineInfo.Input, "", "  ")
	if err != nil {
		panic(fmt.Errorf("error marshalling input: %+v", err))
	}
	return string(input) + "\n"
}

func jobCounts(counts map[int32]int32) string {
	var buffer bytes.Buffer
	for i := int32(ppsclient.JobState_JOB_STARTING); i <= int32(ppsclient.JobState_JOB_SUCCESS); i++ {
		fmt.Fprintf(&buffer, "%s: %d\t", JobState(ppsclient.JobState(i)), counts[i])
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

// ShorthandInput renders a pps.Input as a short, readable string
func ShorthandInput(input *ppsclient.Input) string {
	switch {
	case input == nil:
		return "none"
	case input.Pfs != nil:
		return fmt.Sprintf("%s:%s", input.Pfs.Repo, input.Pfs.Glob)
	case input.Cross != nil:
		var subInput []string
		for _, input := range input.Cross {
			subInput = append(subInput, ShorthandInput(input))
		}
		return "(" + strings.Join(subInput, " ⨯ ") + ")"
	case input.Join != nil:
		var subInput []string
		for _, input := range input.Join {
			subInput = append(subInput, ShorthandInput(input))
		}
		return "(" + strings.Join(subInput, " ⋈ ") + ")"
	case input.Union != nil:
		var subInput []string
		for _, input := range input.Union {
			subInput = append(subInput, ShorthandInput(input))
		}
		return "(" + strings.Join(subInput, " ∪ ") + ")"
	case input.Cron != nil:
		return fmt.Sprintf("%s:%s", input.Cron.Name, input.Cron.Spec)
	}
	return ""
}

var funcMap = template.FuncMap{
	"pipelineState":        pipelineState,
	"jobState":             JobState,
	"datumState":           datumState,
	"workerStatus":         workerStatus,
	"pipelineInput":        pipelineInput,
	"jobInput":             jobInput,
	"prettyAgo":            pretty.Ago,
	"prettyTimeDifference": pretty.TimeDifference,
	"prettyDuration":       pretty.Duration,
	"prettySize":           pretty.Size,
	"jobCounts":            jobCounts,
	"prettyTransform":      prettyTransform,
}
