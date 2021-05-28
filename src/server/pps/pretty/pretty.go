package pretty

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/template"

	units "github.com/docker/go-units"
	"github.com/fatih/color"
	"github.com/gogo/protobuf/types"
	"github.com/juju/ansiterm"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pretty"
	pfsclient "github.com/pachyderm/pachyderm/v2/src/pfs"
	ppsclient "github.com/pachyderm/pachyderm/v2/src/pps"
	pfspretty "github.com/pachyderm/pachyderm/v2/src/server/pfs/pretty"
)

const (
	// PipelineHeader is the header for pipelines.
	PipelineHeader = "NAME\tVERSION\tINPUT\tCREATED\tSTATE / LAST JOB\tDESCRIPTION\t\n"
	// JobHeader is the header for jobs
	JobHeader = "ID\tPIPELINE\tSTARTED\tDURATION\tRESTART\tPROGRESS\tDL\tUL\tSTATE\t\n"
	// DatumHeader is the header for datums
	DatumHeader = "ID\tFILES\tSTATUS\tTIME\t\n"
	// SecretHeader is the header for secrets
	SecretHeader = "NAME\tTYPE\tCREATED\t\n"
	// jobReasonLen is the amount of the job reason that we print
	jobReasonLen = 25
)

func safeTrim(s string, l int) string {
	if len(s) < l {
		return s
	}
	return strings.TrimSpace(s[:l]) + "..."
}

// PrintPipelineJobInfo pretty-prints job info.
func PrintPipelineJobInfo(w io.Writer, pipelineJobInfo *ppsclient.PipelineJobInfo, fullTimestamps bool) {
	fmt.Fprintf(w, "%s\t", pipelineJobInfo.PipelineJob.ID)
	fmt.Fprintf(w, "%s\t", pipelineJobInfo.Pipeline.Name)
	if pipelineJobInfo.Started != nil {
		if fullTimestamps {
			fmt.Fprintf(w, "%s\t", pipelineJobInfo.Started.String())
		} else {
			fmt.Fprintf(w, "%s\t", pretty.Ago(pipelineJobInfo.Started))
		}
	}
	if pipelineJobInfo.Finished != nil {
		fmt.Fprintf(w, "%s\t", pretty.TimeDifference(pipelineJobInfo.Started, pipelineJobInfo.Finished))
	} else {
		fmt.Fprintf(w, "-\t")
	}
	fmt.Fprintf(w, "%d\t", pipelineJobInfo.Restart)
	fmt.Fprintf(w, "%s\t", Progress(pipelineJobInfo))
	fmt.Fprintf(w, "%s\t", pretty.Size(pipelineJobInfo.Stats.DownloadBytes))
	fmt.Fprintf(w, "%s\t", pretty.Size(pipelineJobInfo.Stats.UploadBytes))
	if pipelineJobInfo.State == ppsclient.PipelineJobState_JOB_FAILURE {
		fmt.Fprintf(w, "%s: %s\t", JobState(pipelineJobInfo.State), safeTrim(pipelineJobInfo.Reason, jobReasonLen))
	} else {
		fmt.Fprintf(w, "%s\t", JobState(pipelineJobInfo.State))
	}
	fmt.Fprintln(w)
}

// PrintPipelineInfo pretty-prints pipeline info.
func PrintPipelineInfo(w io.Writer, pipelineInfo *ppsclient.PipelineInfo, fullTimestamps bool) {
	if pipelineInfo.Transform == nil {
		fmt.Fprintf(w, "%s\t", pipelineInfo.Pipeline.Name)
		fmt.Fprint(w, "-\t")
		fmt.Fprint(w, "-\t")
		fmt.Fprint(w, "-\t")
		fmt.Fprintf(w, "%s / %s\t", pipelineState(pipelineInfo.State), JobState(pipelineInfo.LastJobState))
		fmt.Fprint(w, "could not retrieve pipeline spec\t")
	} else {
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
	}
	fmt.Fprintln(w)
}

// PrintWorkerStatusHeader pretty prints a worker status header.
func PrintWorkerStatusHeader(w io.Writer) {
	fmt.Fprint(w, "WORKER\tJOB\tDATUM\tSTARTED\t\n")
}

// PrintWorkerStatus pretty prints a worker status.
func PrintWorkerStatus(w io.Writer, workerStatus *ppsclient.WorkerStatus, fullTimestamps bool) {
	fmt.Fprintf(w, "%s\t", workerStatus.WorkerID)
	fmt.Fprintf(w, "%s\t", workerStatus.PipelineJobID)
	if workerStatus.DatumStatus != nil {
		datumStatus := workerStatus.DatumStatus
		for _, datum := range datumStatus.Data {
			fmt.Fprintf(w, datum.Path)
		}
		fmt.Fprintf(w, "\t")
		if fullTimestamps {
			fmt.Fprintf(w, "%s\t", datumStatus.Started.String())
		} else {
			fmt.Fprintf(w, "%s\t", pretty.Ago(datumStatus.Started))
		}
	}
	fmt.Fprintln(w)
}

// PrintablePipelineJobInfo is a wrapper around PipelineJobInfo containing any formatting options
// used within the template to conditionally print information.
type PrintablePipelineJobInfo struct {
	*ppsclient.PipelineJobInfo
	FullTimestamps bool
}

// NewPrintablePipelineJobInfo constructs a PrintablePipelineJobInfo from just a PipelineJobInfo.
func NewPrintablePipelineJobInfo(pji *ppsclient.PipelineJobInfo) *PrintablePipelineJobInfo {
	return &PrintablePipelineJobInfo{
		PipelineJobInfo: pji,
	}
}

// PrintDetailedPipelineJobInfo pretty-prints detailed job info.
func PrintDetailedPipelineJobInfo(w io.Writer, pipelineJobInfo *PrintablePipelineJobInfo) error {
	template, err := template.New("PipelineJobInfo").Funcs(funcMap).Parse(
		`ID: {{.PipelineJob.ID}} {{if .Pipeline}}
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
{{ if .SidecarResourceLimits }}SidecarResourceLimits:
  CPU: {{ .SidecarResourceLimits.Cpu }}
  Memory: {{ .SidecarResourceLimits.Memory }} {{end}}
{{ if .Service }}Service:
	{{ if .Service.InternalPort }}InternalPort: {{ .Service.InternalPort }} {{end}}
	{{ if .Service.ExternalPort }}ExternalPort: {{ .Service.ExternalPort }} {{end}} {{end}}Input:
{{jobInput .}}
Transform:
{{prettyTransform .Transform}} {{if .OutputCommit}}
Output Commit: {{.OutputCommit.ID}} {{end}}{{ if .Egress }}
Egress: {{.Egress.URL}} {{end}}
`)
	if err != nil {
		return err
	}
	return template.Execute(w, pipelineJobInfo)
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
func PrintDetailedPipelineInfo(w io.Writer, pipelineInfo *PrintablePipelineInfo) error {
	template, err := template.New("PipelineInfo").Funcs(funcMap).Parse(
		`Name: {{.Pipeline.Name}}{{if .Description}}
Description: {{.Description}}{{end}}{{if .FullTimestamps }}
Created: {{.CreatedAt}}{{ else }}
Created: {{prettyAgo .CreatedAt}} {{end}}
State: {{pipelineState .State}}
Reason: {{.Reason}}
Workers Available: {{.WorkersAvailable}}/{{.WorkersRequested}}
Stopped: {{ .Stopped }}
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
	err = template.Execute(w, pipelineInfo)
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
	if datumInfo.Datum.ID == "" {
		datumInfo.Datum.ID = "-"
	}
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t", datumInfo.Datum.ID, datumFiles(datumInfo), datumState(datumInfo.State), totalTime)
	fmt.Fprintln(w)
}

func datumFiles(datumInfo *ppsclient.DatumInfo) string {
	builder := &strings.Builder{}
	for i, fi := range datumInfo.Data {
		if i != 0 {
			builder.WriteString(", ")
		}
		fmt.Fprint(builder, pfspretty.CompactPrintFile(fi.File))
	}
	return builder.String()
}

// PrintDetailedDatumInfo pretty-prints detailed info about a datum
func PrintDetailedDatumInfo(w io.Writer, datumInfo *ppsclient.DatumInfo) {
	fmt.Fprintf(w, "ID\t%s\n", datumInfo.Datum.ID)
	fmt.Fprintf(w, "Pipeline Job ID\t%s\n", datumInfo.Datum.PipelineJob.ID)
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

// PrintSecretInfo pretty-prints secret info.
func PrintSecretInfo(w io.Writer, secretInfo *ppsclient.SecretInfo) {
	fmt.Fprintf(w, "%s\t%s\t%s\t\n", secretInfo.Secret.Name, secretInfo.Type, pretty.Ago(secretInfo.CreationTimestamp))
}

// PrintFileHeader prints the header for a pfs file.
func PrintFileHeader(w io.Writer) {
	fmt.Fprintf(w, "  REPO\tCOMMIT\tPATH\t\n")
}

// PrintFile values for a pfs file.
func PrintFile(w io.Writer, file *pfsclient.File) {
	fmt.Fprintf(w, "  %s\t%s\t%s\t\n", pfspretty.CompactPrintRepo(file.Commit.Branch.Repo), file.Commit.ID, file.Path)
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
func JobState(pipelineJobState ppsclient.PipelineJobState) string {
	switch pipelineJobState {
	case ppsclient.PipelineJobState_JOB_STARTING:
		return color.New(color.FgYellow).SprintFunc()("starting")
	case ppsclient.PipelineJobState_JOB_RUNNING:
		return color.New(color.FgYellow).SprintFunc()("running")
	case ppsclient.PipelineJobState_JOB_FAILURE:
		return color.New(color.FgRed).SprintFunc()("failure")
	case ppsclient.PipelineJobState_JOB_SUCCESS:
		return color.New(color.FgGreen).SprintFunc()("success")
	case ppsclient.PipelineJobState_JOB_KILLED:
		return color.New(color.FgRed).SprintFunc()("killed")
	case ppsclient.PipelineJobState_JOB_EGRESSING:
		return color.New(color.FgYellow).SprintFunc()("egressing")

	}
	return "-"
}

// Progress pretty prints the datum progress of a job.
func Progress(pji *ppsclient.PipelineJobInfo) string {
	if pji.DataRecovered != 0 {
		return fmt.Sprintf("%d + %d + %d / %d", pji.DataProcessed, pji.DataSkipped, pji.DataRecovered, pji.DataTotal)
	}
	return fmt.Sprintf("%d + %d / %d", pji.DataProcessed, pji.DataSkipped, pji.DataTotal)
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
	case ppsclient.PipelineState_PIPELINE_CRASHING:
		return color.New(color.FgRed).SprintFunc()("crashing")
	}
	return "-"
}

func jobInput(ppji PrintablePipelineJobInfo) string {
	if ppji.Input == nil {
		return ""
	}
	input, err := json.MarshalIndent(ppji.Input, "", "  ")
	if err != nil {
		panic(errors.Wrapf(err, "error marshalling input"))
	}
	return string(input) + "\n"
}

func workerStatus(ppji PrintablePipelineJobInfo) string {
	var buffer bytes.Buffer
	writer := ansiterm.NewTabWriter(&buffer, 20, 1, 3, ' ', 0)
	PrintWorkerStatusHeader(writer)
	for _, workerStatus := range ppji.WorkerStatus {
		PrintWorkerStatus(writer, workerStatus, ppji.FullTimestamps)
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
		panic(errors.Wrapf(err, "error marshalling input"))
	}
	return string(input) + "\n"
}

func jobCounts(counts map[int32]int32) string {
	var buffer bytes.Buffer
	for i := int32(ppsclient.PipelineJobState_JOB_STARTING); i <= int32(ppsclient.PipelineJobState_JOB_SUCCESS); i++ {
		fmt.Fprintf(&buffer, "%s: %d\t", JobState(ppsclient.PipelineJobState(i)), counts[i])
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
	case input.Group != nil:
		var subInput []string
		for _, input := range input.Group {
			subInput = append(subInput, ShorthandInput(input))
		}
		return "(Group: " + strings.Join(subInput, ", ") + ")"
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
