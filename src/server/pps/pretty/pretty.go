// Package pretty implements pretty-printing for the PPS service.
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
	"github.com/juju/ansiterm"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pretty"
	pfsclient "github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	ppsclient "github.com/pachyderm/pachyderm/v2/src/pps"
	pfspretty "github.com/pachyderm/pachyderm/v2/src/server/pfs/pretty"
)

const (
	// PipelineHeader is the header for pipelines.
	PipelineHeader = "PROJECT\tNAME\tVERSION\tINPUT\tCREATED\tSTATE / LAST JOB\tALERTS\tDESCRIPTION\t\n"
	// JobHeader is the header for jobs
	JobHeader = "PROJECT\tPIPELINE\tID\tSTARTED\tDURATION\tRESTART\tPROGRESS\tDL\tUL\tSTATE\t\n"
	// JobSetHeader is the header for jobsets
	JobSetHeader = "ID\tSUBJOBS\tPROGRESS\tCREATED\tMODIFIED\n"
	// DatumHeader is the header for datums
	DatumHeader = "ID\tFILES\tSTATUS\tTIME\t\n"
	// SecretHeader is the header for secrets
	SecretHeader = "NAME\tTYPE\tCREATED\t\n"
	// jobReasonLen is the amount of the job reason that we print
	jobReasonLen = 25
	// KubeEventsHeader is the header for kubernetes events
	KubeEventsHeader = "LAST SEEN\tTYPE\tREASON\tOBJECT\tMESSAGE\t\n"
	// CheckStatusHeader is the header for pipeline check statuse calls
	CheckStatusHeader = "PROJECT\tPIPELINE\tALERT\t\n"
)

func safeTrim(s string, l int) string {
	if len(s) < l {
		return s
	}
	return strings.TrimSpace(s[:l]) + "..."
}

type KubeEvent struct {
	Message string `json:"msg"`
	Event   struct {
		Type     string `json:"type"`
		LastSeen uint64 `json:"lastTimestamp"`
		Reason   string `json:"reason"`
		Object   struct {
			Name string `json:"name"`
		} `json:"involvedObject"`
	} `json:"kubeEvent"`
}

func PrintKubeEvent(w io.Writer, event string) {
	var kubeEvent KubeEvent
	if err := ParseKubeEvent(event, &kubeEvent); err != nil {
		fmt.Fprintf(w, "parse kube event: %v", errors.EnsureStack(err))
		return
	}
	lastSeen := timestamppb.Timestamp{Seconds: int64(kubeEvent.Event.LastSeen)}
	fmt.Fprintf(w, "%s\t", pretty.Ago(&lastSeen))
	fmt.Fprintf(w, "%s\t", kubeEvent.Event.Type)
	fmt.Fprintf(w, "%s\t", kubeEvent.Event.Reason)
	fmt.Fprintf(w, "%s\t", kubeEvent.Event.Object.Name)
	fmt.Fprintf(w, "%s\t", kubeEvent.Message)
	fmt.Fprintln(w)
}

func ParseKubeEvent(s string, event *KubeEvent) error {
	// CRI formatted log lines have a timestamp at the beginning of the line
	b := []byte(s)
	i := bytes.IndexByte(b, '{')
	if i > 0 {
		s = string(b[i:])
	}
	if err := json.Unmarshal([]byte(s), &event); err != nil {
		return errors.Wrap(err, "unmarshal kube event")
	}
	// if the message is empty, it might be a docker formatted log line
	if event.Message == "" {
		type DockerLog struct {
			Event string `json:"log"`
		}
		var log DockerLog
		// might be docker formatted logs
		if err := json.Unmarshal([]byte(s), &log); err != nil {
			return errors.Wrap(err, "unmarshal docker log")
		}
		if err := json.Unmarshal([]byte(log.Event), &event); err != nil {
			return errors.Wrap(err, "unmarshal kube event")
		}
	}
	return nil
}

// PrintJobInfo pretty-prints job info.
func PrintJobInfo(w io.Writer, jobInfo *ppsclient.JobInfo, fullTimestamps bool) {
	fmt.Fprintf(w, "%s\t", jobInfo.Job.Pipeline.Project.Name)
	fmt.Fprintf(w, "%s\t", jobInfo.Job.Pipeline.Name)
	fmt.Fprintf(w, "%s\t", jobInfo.Job.Id)
	if jobInfo.Started != nil {
		if fullTimestamps {
			fmt.Fprintf(w, "%s\t", pretty.Timestamp(jobInfo.Started))
		} else {
			fmt.Fprintf(w, "%s\t", pretty.Ago(jobInfo.Started))
		}
	} else {
		fmt.Fprintf(w, "-\t")
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
	if jobInfo.Reason != "" {
		fmt.Fprintf(w, "%s: %s\t", JobState(jobInfo.State), safeTrim(jobInfo.Reason, jobReasonLen))
	} else {
		if jobInfo.State == ppsclient.JobState_JOB_SUCCESS && jobInfo.DataSkipped == jobInfo.DataTotal {
			fmt.Fprintf(w, "%s\t", color.New(color.FgGreen).SprintFunc()("success: there were no datums to process"))
		} else {
			fmt.Fprintf(w, "%s\t", JobState(jobInfo.State))
		}
	}
	fmt.Fprintln(w)
}

// PrintJobSetInfo pretty-prints jobset info.
func PrintJobSetInfo(w io.Writer, jobSetInfo *ppsclient.JobSetInfo, fullTimestamps bool) {
	// Aggregate some data to print from the jobs in the jobset
	success := 0
	failure := 0
	var created *timestamppb.Timestamp
	var modified *timestamppb.Timestamp
	for _, job := range jobSetInfo.Jobs {
		if job.State == ppsclient.JobState_JOB_SUCCESS {
			success++
		} else if pps.IsTerminal(job.State) {
			failure++
		}

		if created == nil {
			created = job.Created
			modified = job.Created
		} else {
			if job.Created.AsTime().Before(created.AsTime()) {
				created = job.Created
			}
			if job.Created.AsTime().After(modified.AsTime()) {
				modified = job.Created
			}
		}
	}

	fmt.Fprintf(w, "%s\t", jobSetInfo.JobSet.Id)
	fmt.Fprintf(w, "%d\t", len(jobSetInfo.Jobs))
	fmt.Fprintf(w, "%s\t", pretty.ProgressBar(8, success, len(jobSetInfo.Jobs)-success-failure, failure))
	if created != nil {
		if fullTimestamps {
			fmt.Fprintf(w, "%s\t", pretty.Timestamp(created))
		} else {
			fmt.Fprintf(w, "%s\t", pretty.Ago(created))
		}
	} else {
		fmt.Fprintf(w, "-\t")
	}
	if modified != nil {
		if fullTimestamps {
			fmt.Fprintf(w, "%s\t", pretty.Timestamp(modified))
		} else {
			fmt.Fprintf(w, "%s\t", pretty.Ago(modified))
		}
	} else {
		fmt.Fprintf(w, "-\t")
	}
	fmt.Fprintln(w)
}

// PrintPipelineInfo pretty-prints pipeline info.
func PrintPipelineInfo(w io.Writer, pipelineInfo *ppsclient.PipelineInfo, fullTimestamps bool) {
	fmt.Fprintf(w, "%s\t", pipelineInfo.Pipeline.Project.Name)
	fmt.Fprintf(w, "%s\t", pipelineInfo.Pipeline.Name)
	fmt.Fprintf(w, "%d\t", pipelineInfo.Version)
	if pipelineInfo.Details == nil {
		fmt.Fprint(w, "-\t") // INPUT
		fmt.Fprint(w, "-\t") // CREATED
		fmt.Fprintf(w, "%s / %s\t", pipelineState(pipelineInfo.State), JobState(pipelineInfo.LastJobState))
		fmt.Fprint(w, "pipeline details unavailable\t")
	} else {
		fmt.Fprintf(w, "%s\t", ShorthandInput(pipelineInfo.Details.Input))
		if fullTimestamps {
			fmt.Fprintf(w, "%s\t", pretty.Timestamp(pipelineInfo.Details.CreatedAt))
		} else {
			fmt.Fprintf(w, "%s\t", pretty.Ago(pipelineInfo.Details.CreatedAt))
		}
		fmt.Fprintf(w, "%s / %s\t", pipelineState(pipelineInfo.State), JobState(pipelineInfo.LastJobState))
		if len(pps.GetAlerts(pipelineInfo)) > 0 {
			fmt.Fprintf(w, "%s\t", "*")
		} else {
			fmt.Fprintf(w, "%s\t", "")
		}
		fmt.Fprintf(w, "%s\t", pipelineInfo.Details.Description)
	}
	fmt.Fprintln(w)
}

func PrintCheckStatus(w io.Writer, checkSatusResponse *ppsclient.CheckStatusResponse) {
	for _, alert := range checkSatusResponse.Alerts {
		fmt.Fprintf(w, "%s\t", checkSatusResponse.GetProject())
		fmt.Fprintf(w, "%s\t", checkSatusResponse.GetPipeline())
		fmt.Fprintf(w, "%s\t", alert)
		fmt.Fprintln(w)
	}
}

// PrintWorkerStatusHeader pretty prints a worker status header.
func PrintWorkerStatusHeader(w io.Writer) {
	fmt.Fprint(w, "WORKER\tJOB\tDATUM\tSTARTED\t\n")
}

// PrintWorkerStatus pretty prints a worker status.
func PrintWorkerStatus(w io.Writer, workerStatus *ppsclient.WorkerStatus, fullTimestamps bool) {
	fmt.Fprintf(w, "%s\t", workerStatus.WorkerId)
	fmt.Fprintf(w, "%s\t", workerStatus.JobId)
	if workerStatus.DatumStatus != nil {
		datumStatus := workerStatus.DatumStatus
		for _, datum := range datumStatus.Data {
			fmt.Fprint(w, datum.Path)
		}
		fmt.Fprintf(w, "\t")
		if fullTimestamps {
			fmt.Fprintf(w, "%s\t", pretty.Timestamp(datumStatus.Started))
		} else {
			fmt.Fprintf(w, "%s\t", pretty.Ago(datumStatus.Started))
		}
	}
	fmt.Fprintln(w)
}

// PrintableJobInfo is a wrapper around JobInfo containing any formatting options
// used within the template to conditionally print information.
type PrintableJobInfo struct {
	*ppsclient.JobInfo
	FullTimestamps bool
}

// NewPrintableJobInfo constructs a PrintableJobInfo from just a JobInfo.
func NewPrintableJobInfo(ji *ppsclient.JobInfo, full bool) *PrintableJobInfo {
	return &PrintableJobInfo{
		JobInfo:        ji,
		FullTimestamps: full,
	}
}

// PrintDetailedJobInfo pretty-prints detailed job info.
func PrintDetailedJobInfo(w io.Writer, jobInfo *PrintableJobInfo) error {
	template, err := template.New("JobInfo").Funcs(funcMap).Parse(
		`ID: {{.Job.Id}}
Pipeline: {{.Job.Pipeline.Name}}
Project: {{.Job.Pipeline.Project.Name}}{{if .FullTimestamps}}
Started: {{prettyTime .Started}}{{else}}
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
Datum Timeout: {{.Details.DatumTimeout}}
Job Timeout: {{.Details.JobTimeout}}
Worker Status:
{{workerStatus .}}Restarts: {{.Restart}}
ParallelismSpec: {{.Details.ParallelismSpec}}
{{- if .Details.ResourceRequests }}
ResourceRequests:
  CPU: {{ .Details.ResourceRequests.Cpu }}
  Memory: {{ .Details.ResourceRequests.Memory }} {{end}}
{{- if .Details.ResourceLimits }}
ResourceLimits:
  CPU: {{ .Details.ResourceLimits.Cpu }}
  Memory: {{ .Details.ResourceLimits.Memory }}
  {{ if .Details.ResourceLimits.Gpu }}GPU:
    Type: {{ .Details.ResourceLimits.Gpu.Type }}
    Number: {{ .Details.ResourceLimits.Gpu.Number }} {{end}} {{end}}
{{- if .Details.SidecarResourceRequests }}
SidecarResourceRequests:
  CPU: {{ .Details.SidecarResourceRequests.Cpu }}
  Memory: {{ .Details.SidecarResourceRequests.Memory }} {{end}}
{{- if .Details.SidecarResourceLimits }}
SidecarResourceLimits:
  CPU: {{ .Details.SidecarResourceLimits.Cpu }}
  Memory: {{ .Details.SidecarResourceLimits.Memory }} {{end}}
{{- if .Details.Service }}
Service:
	{{ if .Details.Service.InternalPort }}InternalPort: {{ .Details.Service.InternalPort }} {{end}}
	{{ if .Details.Service.ExternalPort }}ExternalPort: {{ .Details.Service.ExternalPort }} {{end}} {{end}}
Input: {{jobInput . }}
Transform: {{prettyTransform .Details.Transform}} {{if .OutputCommit}}
Output Commit: {{.OutputCommit.Id}} {{end}}{{ if .Details.Egress }}
Egress: {{egress .Details.Egress}} {{end}}
`)
	if err != nil {
		return errors.EnsureStack(err)
	}
	return errors.EnsureStack(template.Execute(w, jobInfo))
}

// PrintablePipelineInfo is a wrapper around PipelineInfo containing any formatting options
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
		`Name: {{.Pipeline.Name}}{{if .Details.Description}}
Description: {{.Details.Description}}{{end}}{{if .FullTimestamps }}
Created: {{prettyTime .Details.CreatedAt}}{{ else }}
Created: {{prettyAgo .Details.CreatedAt}} {{end}}
State: {{pipelineState .State}}
Reason: {{.Reason}}
Workers Available: {{.Details.WorkersAvailable}}/{{.Details.WorkersRequested}}
Stopped: {{ .Stopped }}
Parallelism Spec: {{.Details.ParallelismSpec}}
{{- if .Details.ResourceRequests }}
ResourceRequests:
  CPU: {{ .Details.ResourceRequests.Cpu }}
  Memory: {{ .Details.ResourceRequests.Memory }} {{end}}
{{- if .Details.ResourceLimits }}
ResourceLimits:
  CPU: {{ .Details.ResourceLimits.Cpu }}
  Memory: {{ .Details.ResourceLimits.Memory }}
  {{ if .Details.ResourceLimits.Gpu }}GPU:
    Type: {{ .Details.ResourceLimits.Gpu.Type }}
    Number: {{ .Details.ResourceLimits.Gpu.Number }} {{end}} {{end}}
Datum Timeout: {{.Details.DatumTimeout}}
Job Timeout: {{.Details.JobTimeout}}
Input: {{pipelineInput .PipelineInfo.GetDetails.Input}}
Output Branch: {{.Details.OutputBranch}}
Transform: {{prettyTransform .Details.Transform}}
{{ if .Details.Egress -}}
Egress: {{ egress .Details.Egress }}
{{ end -}}
{{ if .Details.RecentError -}}
Recent Error: {{ .Details.RecentError }}
{{ end -}}
{{ if .UserSpecJson -}}
User Spec: {{ .UserSpecJson | json "  " "  " }}
{{ end -}}
{{ if .EffectiveSpecJson -}}
Effective Spec: {{ .EffectiveSpecJson | json "  " "  " }}
{{ end -}}
`)
	if err != nil {
		return errors.EnsureStack(err)
	}
	return errors.EnsureStack(template.Execute(w, pipelineInfo))
}

// PrintCreatePipelineRequest pretty-prints a create pipeline request.
func PrintCreatePipelineRequest(w io.Writer, req *ppsclient.CreatePipelineRequest) error {
	template, err := template.New("CreatePipelineRequest").Funcs(funcMap).Parse(
		`Name: {{.Pipeline.Name}}{{if .Description}}
Description: {{.Description}}{{end}}
Parallelism Spec: {{.ParallelismSpec}}
{{- if .ResourceRequests }}
ResourceRequests:
  CPU: {{ .ResourceRequests.Cpu }}
  Memory: {{ .ResourceRequests.Memory }} {{end}}
{{- if .ResourceLimits }}
ResourceLimits:
  CPU: {{ .ResourceLimits.Cpu }}
  Memory: {{ .ResourceLimits.Memory }}
  {{ if .ResourceLimits.Gpu }}GPU:
    Type: {{ .ResourceLimits.Gpu.Type }}
    Number: {{ .ResourceLimits.Gpu.Number }} {{end}} {{end}}
Datum Timeout: {{.DatumTimeout}}
Job Timeout: {{.JobTimeout}}
Input: {{pipelineInput .Input}}
Output Branch: {{.OutputBranch}}
Transform: {{prettyTransform .Transform}}
{{ if .Egress }}Egress: {{egress .Egress}} {{end}}
`)
	if err != nil {
		return errors.EnsureStack(err)
	}
	return errors.EnsureStack(template.Execute(w, req))
}

// PrintDatumInfo pretty-prints file info.
// If recurse is false and directory size is 0, display "-" instead
// If fast is true and file size is 0, display "-" instead
func PrintDatumInfo(w io.Writer, datumInfo *ppsclient.DatumInfo) {
	totalTime := "-"
	if datumInfo.Stats != nil {
		totalTime = units.HumanDuration(client.GetDatumTotalTime(datumInfo.Stats))
	}
	if datumInfo.Datum.Id == "" {
		datumInfo.Datum.Id = "-"
	}
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t", datumInfo.Datum.Id, datumFiles(datumInfo), datumState(datumInfo.State), totalTime)
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
	fmt.Fprintf(w, "ID\t%s\n", datumInfo.Datum.Id)
	fmt.Fprintf(w, "Job ID\t%s\n", datumInfo.Datum.Job.Id)
	fmt.Fprintf(w, "Image ID\t%s\n", datumInfo.ImageId)
	fmt.Fprintf(w, "State\t%s\n", datumInfo.State)
	fmt.Fprintf(w, "Data Downloaded\t%s\n", pretty.Size(datumInfo.Stats.DownloadBytes))
	fmt.Fprintf(w, "Data Uploaded\t%s\n", pretty.Size(datumInfo.Stats.UploadBytes))

	totalTime := client.GetDatumTotalTime(datumInfo.Stats).String()
	fmt.Fprintf(w, "Total Time\t%s\n", totalTime)
	fmt.Fprintf(w, "Download Time\t%s\n", datumInfo.Stats.DownloadTime.AsDuration())
	fmt.Fprintf(w, "Process Time\t%s\n", datumInfo.Stats.ProcessTime.AsDuration())
	fmt.Fprintf(w, "Upload Time\t%s\n", datumInfo.Stats.UploadTime.AsDuration())

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
	fmt.Fprintf(w, "  %s\t%s\t%s\t\n", file.Commit.Repo, file.Commit.Id, file.Path)
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
	case ppsclient.DatumState_UNKNOWN:
		return color.New(color.FgGreen).SprintFunc()("-")
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
	case ppsclient.JobState_JOB_FAILURE:
		return color.New(color.FgRed).SprintFunc()("failure")
	case ppsclient.JobState_JOB_SUCCESS:
		return color.New(color.FgGreen).SprintFunc()("success")
	case ppsclient.JobState_JOB_KILLED:
		return color.New(color.FgRed).SprintFunc()("killed")
	case ppsclient.JobState_JOB_EGRESSING:
		return color.New(color.FgYellow).SprintFunc()("egressing")
	case ppsclient.JobState_JOB_FINISHING:
		return color.New(color.FgYellow).SprintFunc()("finishing")
	case ppsclient.JobState_JOB_UNRUNNABLE:
		return color.New(color.FgRed).SprintFunc()("unrunnable")

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
	case ppsclient.PipelineState_PIPELINE_CRASHING:
		return color.New(color.FgRed).SprintFunc()("crashing")
	}
	return "-"
}

func jobInput(pji PrintableJobInfo) string {
	if pji.Details.Input == nil {
		return ""
	}
	input, err := json.MarshalIndent(pji.Details.Input, "", "  ")
	if err != nil {
		panic(errors.Wrapf(err, "error marshalling input"))
	}
	return string(input)
}

func workerStatus(pji PrintableJobInfo) string {
	var buffer bytes.Buffer
	writer := ansiterm.NewTabWriter(&buffer, 20, 1, 3, ' ', 0)
	PrintWorkerStatusHeader(writer)
	for _, workerStatus := range pji.Details.WorkerStatus {
		PrintWorkerStatus(writer, workerStatus, pji.FullTimestamps)
	}
	// can't error because buffer can't error on Write
	writer.Flush()
	return buffer.String()
}

func pipelineInput(i *ppsclient.Input) string {
	if i == nil {
		return ""
	}
	input, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		panic(errors.Wrapf(err, "error marshalling input"))
	}
	return string(input)
}

func prettyTransform(transform *ppsclient.Transform) (string, error) {
	result, err := json.MarshalIndent(transform, "", "  ")
	if err != nil {
		return "", errors.EnsureStack(err)
	}
	return pretty.UnescapeHTML(string(result)), nil
}

// ShorthandInput renders a pps.Input as a short, readable string
func ShorthandInput(input *ppsclient.Input) string {
	switch {
	case input == nil:
		return "none"
	case input.Pfs != nil:
		return fmt.Sprintf("%s/%s:%s", input.Pfs.Project, input.Pfs.Repo, input.Pfs.Glob)
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

func egress(e *ppsclient.Egress) string {
	target := e.GetTarget()
	if target == nil {
		return e.GetURL()
	}
	s, err := json.MarshalIndent(target, "", "  ")
	if err != nil {
		panic(errors.Wrapf(err, "error marshalling egress"))
	}
	return string(s)
}

func js(prefix, indent, s string) (string, error) {
	var buf bytes.Buffer
	if err := json.Indent(&buf, []byte(s), prefix, indent); err != nil {
		return "", errors.Wrapf(err, "could not indent JSON %q", s)
	}
	return buf.String(), nil
}

var funcMap = template.FuncMap{
	"pipelineState":        pipelineState,
	"jobState":             JobState,
	"datumState":           datumState,
	"workerStatus":         workerStatus,
	"pipelineInput":        pipelineInput,
	"jobInput":             jobInput,
	"prettyAgo":            pretty.Ago,
	"prettyDuration":       pretty.Duration,
	"prettySize":           pretty.Size,
	"prettyTime":           pretty.Timestamp,
	"prettyTimeDifference": pretty.TimeDifference,
	"prettyTransform":      prettyTransform,
	"egress":               egress,
	"json":                 js,
}
