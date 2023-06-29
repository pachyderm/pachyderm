// Code generated by protoc-gen-zap (etc/proto/protoc-gen-zap). DO NOT EDIT.
//
// source: pps/pps.proto

package pps

import (
	fmt "fmt"
	protoextensions "github.com/pachyderm/pachyderm/v2/src/protoextensions"
	zapcore "go.uber.org/zap/zapcore"
)

func (x *SecretMount) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddString("name", x.Name)
	enc.AddString("key", x.Key)
	enc.AddString("mount_path", x.MountPath)
	enc.AddString("env_var", x.EnvVar)
	return nil
}

func (x *Transform) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddString("image", x.Image)
	cmdArrMarshaller := func(enc zapcore.ArrayEncoder) error {
		for _, v := range x.Cmd {
			enc.AppendString(v)
		}
		return nil
	}
	enc.AddArray("cmd", zapcore.ArrayMarshalerFunc(cmdArrMarshaller))
	err_cmdArrMarshaller := func(enc zapcore.ArrayEncoder) error {
		for _, v := range x.ErrCmd {
			enc.AppendString(v)
		}
		return nil
	}
	enc.AddArray("err_cmd", zapcore.ArrayMarshalerFunc(err_cmdArrMarshaller))
	enc.AddObject("env", zapcore.ObjectMarshalerFunc(func(enc zapcore.ObjectEncoder) error {
		for k, v := range x.Env {
			enc.AddString(fmt.Sprintf("%v", k), v)
		}
		return nil
	}))
	secretsArrMarshaller := func(enc zapcore.ArrayEncoder) error {
		for _, v := range x.Secrets {
			enc.AppendObject(v)
		}
		return nil
	}
	enc.AddArray("secrets", zapcore.ArrayMarshalerFunc(secretsArrMarshaller))
	image_pull_secretsArrMarshaller := func(enc zapcore.ArrayEncoder) error {
		for _, v := range x.ImagePullSecrets {
			enc.AppendString(v)
		}
		return nil
	}
	enc.AddArray("image_pull_secrets", zapcore.ArrayMarshalerFunc(image_pull_secretsArrMarshaller))
	stdinArrMarshaller := func(enc zapcore.ArrayEncoder) error {
		for _, v := range x.Stdin {
			enc.AppendString(v)
		}
		return nil
	}
	enc.AddArray("stdin", zapcore.ArrayMarshalerFunc(stdinArrMarshaller))
	err_stdinArrMarshaller := func(enc zapcore.ArrayEncoder) error {
		for _, v := range x.ErrStdin {
			enc.AppendString(v)
		}
		return nil
	}
	enc.AddArray("err_stdin", zapcore.ArrayMarshalerFunc(err_stdinArrMarshaller))
	accept_return_codeArrMarshaller := func(enc zapcore.ArrayEncoder) error {
		for _, v := range x.AcceptReturnCode {
			enc.AppendInt64(v)
		}
		return nil
	}
	enc.AddArray("accept_return_code", zapcore.ArrayMarshalerFunc(accept_return_codeArrMarshaller))
	enc.AddBool("debug", x.Debug)
	enc.AddString("user", x.User)
	enc.AddString("working_dir", x.WorkingDir)
	enc.AddString("dockerfile", x.Dockerfile)
	enc.AddBool("memory_volume", x.MemoryVolume)
	enc.AddBool("datum_batching", x.DatumBatching)
	return nil
}

func (x *TFJob) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddString("tf_job", x.TFJob)
	return nil
}

func (x *Egress) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddString("URL", x.URL)
	enc.AddObject("object_storage", x.GetObjectStorage())
	enc.AddObject("sql_database", x.GetSqlDatabase())
	return nil
}

func (x *Job) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddObject("pipeline", x.Pipeline)
	enc.AddString("id", x.ID)
	return nil
}

func (x *Metadata) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddObject("annotations", zapcore.ObjectMarshalerFunc(func(enc zapcore.ObjectEncoder) error {
		for k, v := range x.Annotations {
			enc.AddString(fmt.Sprintf("%v", k), v)
		}
		return nil
	}))
	enc.AddObject("labels", zapcore.ObjectMarshalerFunc(func(enc zapcore.ObjectEncoder) error {
		for k, v := range x.Labels {
			enc.AddString(fmt.Sprintf("%v", k), v)
		}
		return nil
	}))
	return nil
}

func (x *Service) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddInt32("internal_port", x.InternalPort)
	enc.AddInt32("external_port", x.ExternalPort)
	enc.AddString("ip", x.IP)
	enc.AddString("type", x.Type)
	return nil
}

func (x *Spout) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddObject("service", x.Service)
	return nil
}

func (x *PFSInput) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddString("project", x.Project)
	enc.AddString("name", x.Name)
	enc.AddString("repo", x.Repo)
	enc.AddString("repo_type", x.RepoType)
	enc.AddString("branch", x.Branch)
	enc.AddString("commit", x.Commit)
	enc.AddString("glob", x.Glob)
	enc.AddString("join_on", x.JoinOn)
	enc.AddBool("outer_join", x.OuterJoin)
	enc.AddString("group_by", x.GroupBy)
	enc.AddBool("lazy", x.Lazy)
	enc.AddBool("empty_files", x.EmptyFiles)
	enc.AddBool("s3", x.S3)
	enc.AddObject("trigger", x.Trigger)
	return nil
}

func (x *CronInput) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddString("name", x.Name)
	enc.AddString("project", x.Project)
	enc.AddString("repo", x.Repo)
	enc.AddString("commit", x.Commit)
	enc.AddString("spec", x.Spec)
	enc.AddBool("overwrite", x.Overwrite)
	protoextensions.AddTimestamp(enc, "start", x.Start)
	return nil
}

func (x *Input) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddObject("pfs", x.Pfs)
	joinArrMarshaller := func(enc zapcore.ArrayEncoder) error {
		for _, v := range x.Join {
			enc.AppendObject(v)
		}
		return nil
	}
	enc.AddArray("join", zapcore.ArrayMarshalerFunc(joinArrMarshaller))
	groupArrMarshaller := func(enc zapcore.ArrayEncoder) error {
		for _, v := range x.Group {
			enc.AppendObject(v)
		}
		return nil
	}
	enc.AddArray("group", zapcore.ArrayMarshalerFunc(groupArrMarshaller))
	crossArrMarshaller := func(enc zapcore.ArrayEncoder) error {
		for _, v := range x.Cross {
			enc.AppendObject(v)
		}
		return nil
	}
	enc.AddArray("cross", zapcore.ArrayMarshalerFunc(crossArrMarshaller))
	unionArrMarshaller := func(enc zapcore.ArrayEncoder) error {
		for _, v := range x.Union {
			enc.AppendObject(v)
		}
		return nil
	}
	enc.AddArray("union", zapcore.ArrayMarshalerFunc(unionArrMarshaller))
	enc.AddObject("cron", x.Cron)
	return nil
}

func (x *JobInput) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddString("name", x.Name)
	enc.AddObject("commit", x.Commit)
	enc.AddString("glob", x.Glob)
	enc.AddBool("lazy", x.Lazy)
	return nil
}

func (x *ParallelismSpec) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddUint64("constant", x.Constant)
	return nil
}

func (x *InputFile) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddString("path", x.Path)
	protoextensions.AddBytes(enc, "hash", x.Hash)
	return nil
}

func (x *Datum) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddObject("job", x.Job)
	enc.AddString("id", x.ID)
	return nil
}

func (x *DatumInfo) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddObject("datum", x.Datum)
	enc.AddString("state", x.State.String())
	enc.AddObject("stats", x.Stats)
	enc.AddObject("pfs_state", x.PfsState)
	dataArrMarshaller := func(enc zapcore.ArrayEncoder) error {
		for _, v := range x.Data {
			enc.AppendObject(v)
		}
		return nil
	}
	enc.AddArray("data", zapcore.ArrayMarshalerFunc(dataArrMarshaller))
	enc.AddString("image_id", x.ImageId)
	return nil
}

func (x *Aggregate) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddInt64("count", x.Count)
	enc.AddFloat64("mean", x.Mean)
	enc.AddFloat64("stddev", x.Stddev)
	enc.AddFloat64("fifth_percentile", x.FifthPercentile)
	enc.AddFloat64("ninety_fifth_percentile", x.NinetyFifthPercentile)
	return nil
}

func (x *ProcessStats) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	protoextensions.AddDuration(enc, "download_time", x.DownloadTime)
	protoextensions.AddDuration(enc, "process_time", x.ProcessTime)
	protoextensions.AddDuration(enc, "upload_time", x.UploadTime)
	enc.AddInt64("download_bytes", x.DownloadBytes)
	enc.AddInt64("upload_bytes", x.UploadBytes)
	return nil
}

func (x *AggregateProcessStats) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddObject("download_time", x.DownloadTime)
	enc.AddObject("process_time", x.ProcessTime)
	enc.AddObject("upload_time", x.UploadTime)
	enc.AddObject("download_bytes", x.DownloadBytes)
	enc.AddObject("upload_bytes", x.UploadBytes)
	return nil
}

func (x *WorkerStatus) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddString("worker_id", x.WorkerID)
	enc.AddString("job_id", x.JobID)
	enc.AddObject("datum_status", x.DatumStatus)
	return nil
}

func (x *DatumStatus) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	protoextensions.AddTimestamp(enc, "started", x.Started)
	dataArrMarshaller := func(enc zapcore.ArrayEncoder) error {
		for _, v := range x.Data {
			enc.AppendObject(v)
		}
		return nil
	}
	enc.AddArray("data", zapcore.ArrayMarshalerFunc(dataArrMarshaller))
	return nil
}

func (x *ResourceSpec) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddFloat32("cpu", x.Cpu)
	enc.AddString("memory", x.Memory)
	enc.AddObject("gpu", x.Gpu)
	enc.AddString("disk", x.Disk)
	return nil
}

func (x *GPUSpec) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddString("type", x.Type)
	enc.AddInt64("number", x.Number)
	return nil
}

func (x *JobSetInfo) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddObject("job_set", x.JobSet)
	jobsArrMarshaller := func(enc zapcore.ArrayEncoder) error {
		for _, v := range x.Jobs {
			enc.AppendObject(v)
		}
		return nil
	}
	enc.AddArray("jobs", zapcore.ArrayMarshalerFunc(jobsArrMarshaller))
	return nil
}

func (x *JobInfo) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddObject("job", x.Job)
	enc.AddUint64("pipeline_version", x.PipelineVersion)
	enc.AddObject("output_commit", x.OutputCommit)
	enc.AddUint64("restart", x.Restart)
	enc.AddInt64("data_processed", x.DataProcessed)
	enc.AddInt64("data_skipped", x.DataSkipped)
	enc.AddInt64("data_total", x.DataTotal)
	enc.AddInt64("data_failed", x.DataFailed)
	enc.AddInt64("data_recovered", x.DataRecovered)
	enc.AddObject("stats", x.Stats)
	enc.AddString("state", x.State.String())
	enc.AddString("reason", x.Reason)
	protoextensions.AddTimestamp(enc, "created", x.Created)
	protoextensions.AddTimestamp(enc, "started", x.Started)
	protoextensions.AddTimestamp(enc, "finished", x.Finished)
	enc.AddObject("details", x.Details)
	enc.AddString("auth_token", x.AuthToken)
	return nil
}

func (x *JobInfo_Details) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddObject("transform", x.Transform)
	enc.AddObject("parallelism_spec", x.ParallelismSpec)
	enc.AddObject("egress", x.Egress)
	enc.AddObject("service", x.Service)
	enc.AddObject("spout", x.Spout)
	worker_statusArrMarshaller := func(enc zapcore.ArrayEncoder) error {
		for _, v := range x.WorkerStatus {
			enc.AppendObject(v)
		}
		return nil
	}
	enc.AddArray("worker_status", zapcore.ArrayMarshalerFunc(worker_statusArrMarshaller))
	enc.AddObject("resource_requests", x.ResourceRequests)
	enc.AddObject("resource_limits", x.ResourceLimits)
	enc.AddObject("sidecar_resource_limits", x.SidecarResourceLimits)
	enc.AddObject("input", x.Input)
	enc.AddString("salt", x.Salt)
	enc.AddObject("datum_set_spec", x.DatumSetSpec)
	protoextensions.AddDuration(enc, "datum_timeout", x.DatumTimeout)
	protoextensions.AddDuration(enc, "job_timeout", x.JobTimeout)
	enc.AddInt64("datum_tries", x.DatumTries)
	enc.AddObject("scheduling_spec", x.SchedulingSpec)
	enc.AddString("pod_spec", x.PodSpec)
	enc.AddString("pod_patch", x.PodPatch)
	enc.AddObject("sidecar_resource_requests", x.SidecarResourceRequests)
	return nil
}

func (x *Worker) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddString("name", x.Name)
	enc.AddString("state", x.State.String())
	return nil
}

func (x *Pipeline) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddObject("project", x.Project)
	enc.AddString("name", x.Name)
	return nil
}

func (x *Toleration) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddString("key", x.Key)
	enc.AddString("operator", x.Operator.String())
	enc.AddString("value", x.Value)
	enc.AddString("effect", x.Effect.String())
	protoextensions.AddInt64Value(enc, "toleration_seconds", x.TolerationSeconds)
	return nil
}

func (x *PipelineInfo) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddObject("pipeline", x.Pipeline)
	enc.AddUint64("version", x.Version)
	enc.AddObject("spec_commit", x.SpecCommit)
	enc.AddBool("stopped", x.Stopped)
	enc.AddString("state", x.State.String())
	enc.AddString("reason", x.Reason)
	enc.AddString("last_job_state", x.LastJobState.String())
	enc.AddUint64("parallelism", x.Parallelism)
	enc.AddString("type", x.Type.String())
	enc.AddString("auth_token", x.AuthToken)
	enc.AddObject("details", x.Details)
	return nil
}

func (x *PipelineInfo_Details) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddObject("transform", x.Transform)
	enc.AddObject("tf_job", x.TFJob)
	enc.AddObject("parallelism_spec", x.ParallelismSpec)
	enc.AddObject("egress", x.Egress)
	protoextensions.AddTimestamp(enc, "created_at", x.CreatedAt)
	enc.AddString("recent_error", x.RecentError)
	enc.AddInt64("workers_requested", x.WorkersRequested)
	enc.AddInt64("workers_available", x.WorkersAvailable)
	enc.AddString("output_branch", x.OutputBranch)
	enc.AddObject("resource_requests", x.ResourceRequests)
	enc.AddObject("resource_limits", x.ResourceLimits)
	enc.AddObject("sidecar_resource_limits", x.SidecarResourceLimits)
	enc.AddObject("input", x.Input)
	enc.AddString("description", x.Description)
	enc.AddString("salt", x.Salt)
	enc.AddString("reason", x.Reason)
	enc.AddObject("service", x.Service)
	enc.AddObject("spout", x.Spout)
	enc.AddObject("datum_set_spec", x.DatumSetSpec)
	protoextensions.AddDuration(enc, "datum_timeout", x.DatumTimeout)
	protoextensions.AddDuration(enc, "job_timeout", x.JobTimeout)
	enc.AddInt64("datum_tries", x.DatumTries)
	enc.AddObject("scheduling_spec", x.SchedulingSpec)
	enc.AddString("pod_spec", x.PodSpec)
	enc.AddString("pod_patch", x.PodPatch)
	enc.AddBool("s3_out", x.S3Out)
	enc.AddObject("metadata", x.Metadata)
	enc.AddString("reprocess_spec", x.ReprocessSpec)
	enc.AddInt64("unclaimed_tasks", x.UnclaimedTasks)
	enc.AddString("worker_rc", x.WorkerRc)
	enc.AddBool("autoscaling", x.Autoscaling)
	tolerationsArrMarshaller := func(enc zapcore.ArrayEncoder) error {
		for _, v := range x.Tolerations {
			enc.AppendObject(v)
		}
		return nil
	}
	enc.AddArray("tolerations", zapcore.ArrayMarshalerFunc(tolerationsArrMarshaller))
	enc.AddObject("sidecar_resource_requests", x.SidecarResourceRequests)
	return nil
}

func (x *PipelineInfos) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	pipeline_infoArrMarshaller := func(enc zapcore.ArrayEncoder) error {
		for _, v := range x.PipelineInfo {
			enc.AppendObject(v)
		}
		return nil
	}
	enc.AddArray("pipeline_info", zapcore.ArrayMarshalerFunc(pipeline_infoArrMarshaller))
	return nil
}

func (x *JobSet) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddString("id", x.ID)
	return nil
}

func (x *InspectJobSetRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddObject("job_set", x.JobSet)
	enc.AddBool("wait", x.Wait)
	enc.AddBool("details", x.Details)
	return nil
}

func (x *ListJobSetRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddBool("details", x.Details)
	projectsArrMarshaller := func(enc zapcore.ArrayEncoder) error {
		for _, v := range x.Projects {
			enc.AppendObject(v)
		}
		return nil
	}
	enc.AddArray("projects", zapcore.ArrayMarshalerFunc(projectsArrMarshaller))
	protoextensions.AddTimestamp(enc, "paginationMarker", x.PaginationMarker)
	enc.AddInt64("number", x.Number)
	enc.AddBool("reverse", x.Reverse)
	enc.AddString("jqFilter", x.JqFilter)
	return nil
}

func (x *InspectJobRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddObject("job", x.Job)
	enc.AddBool("wait", x.Wait)
	enc.AddBool("details", x.Details)
	return nil
}

func (x *ListJobRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	projectsArrMarshaller := func(enc zapcore.ArrayEncoder) error {
		for _, v := range x.Projects {
			enc.AppendObject(v)
		}
		return nil
	}
	enc.AddArray("projects", zapcore.ArrayMarshalerFunc(projectsArrMarshaller))
	enc.AddObject("pipeline", x.Pipeline)
	input_commitArrMarshaller := func(enc zapcore.ArrayEncoder) error {
		for _, v := range x.InputCommit {
			enc.AppendObject(v)
		}
		return nil
	}
	enc.AddArray("input_commit", zapcore.ArrayMarshalerFunc(input_commitArrMarshaller))
	enc.AddInt64("history", x.History)
	enc.AddBool("details", x.Details)
	enc.AddString("jqFilter", x.JqFilter)
	protoextensions.AddTimestamp(enc, "paginationMarker", x.PaginationMarker)
	enc.AddInt64("number", x.Number)
	enc.AddBool("reverse", x.Reverse)
	return nil
}

func (x *SubscribeJobRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddObject("pipeline", x.Pipeline)
	enc.AddBool("details", x.Details)
	return nil
}

func (x *DeleteJobRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddObject("job", x.Job)
	return nil
}

func (x *StopJobRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddObject("job", x.Job)
	enc.AddString("reason", x.Reason)
	return nil
}

func (x *UpdateJobStateRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddObject("job", x.Job)
	enc.AddString("state", x.State.String())
	enc.AddString("reason", x.Reason)
	enc.AddUint64("restart", x.Restart)
	enc.AddInt64("data_processed", x.DataProcessed)
	enc.AddInt64("data_skipped", x.DataSkipped)
	enc.AddInt64("data_failed", x.DataFailed)
	enc.AddInt64("data_recovered", x.DataRecovered)
	enc.AddInt64("data_total", x.DataTotal)
	enc.AddObject("stats", x.Stats)
	return nil
}

func (x *GetLogsRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddObject("pipeline", x.Pipeline)
	enc.AddObject("job", x.Job)
	data_filtersArrMarshaller := func(enc zapcore.ArrayEncoder) error {
		for _, v := range x.DataFilters {
			enc.AppendString(v)
		}
		return nil
	}
	enc.AddArray("data_filters", zapcore.ArrayMarshalerFunc(data_filtersArrMarshaller))
	enc.AddObject("datum", x.Datum)
	enc.AddBool("master", x.Master)
	enc.AddBool("follow", x.Follow)
	enc.AddInt64("tail", x.Tail)
	enc.AddBool("use_loki_backend", x.UseLokiBackend)
	protoextensions.AddDuration(enc, "since", x.Since)
	return nil
}

func (x *LogMessage) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddString("project_name", x.ProjectName)
	enc.AddString("pipeline_name", x.PipelineName)
	enc.AddString("job_id", x.JobID)
	enc.AddString("worker_id", x.WorkerID)
	enc.AddString("datum_id", x.DatumID)
	enc.AddBool("master", x.Master)
	dataArrMarshaller := func(enc zapcore.ArrayEncoder) error {
		for _, v := range x.Data {
			enc.AppendObject(v)
		}
		return nil
	}
	enc.AddArray("data", zapcore.ArrayMarshalerFunc(dataArrMarshaller))
	enc.AddBool("user", x.User)
	protoextensions.AddTimestamp(enc, "ts", x.Ts)
	enc.AddString("message", x.Message)
	return nil
}

func (x *RestartDatumRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddObject("job", x.Job)
	data_filtersArrMarshaller := func(enc zapcore.ArrayEncoder) error {
		for _, v := range x.DataFilters {
			enc.AppendString(v)
		}
		return nil
	}
	enc.AddArray("data_filters", zapcore.ArrayMarshalerFunc(data_filtersArrMarshaller))
	return nil
}

func (x *InspectDatumRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddObject("datum", x.Datum)
	return nil
}

func (x *ListDatumRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddObject("job", x.Job)
	enc.AddObject("input", x.Input)
	enc.AddObject("filter", x.Filter)
	enc.AddString("paginationMarker", x.PaginationMarker)
	enc.AddInt64("number", x.Number)
	enc.AddBool("reverse", x.Reverse)
	return nil
}

func (x *ListDatumRequest_Filter) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	stateArrMarshaller := func(enc zapcore.ArrayEncoder) error {
		for _, v := range x.State {
			enc.AppendString(v.String())
		}
		return nil
	}
	enc.AddArray("state", zapcore.ArrayMarshalerFunc(stateArrMarshaller))
	return nil
}

func (x *DatumSetSpec) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddInt64("number", x.Number)
	enc.AddInt64("size_bytes", x.SizeBytes)
	enc.AddInt64("per_worker", x.PerWorker)
	return nil
}

func (x *SchedulingSpec) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddObject("node_selector", zapcore.ObjectMarshalerFunc(func(enc zapcore.ObjectEncoder) error {
		for k, v := range x.NodeSelector {
			enc.AddString(fmt.Sprintf("%v", k), v)
		}
		return nil
	}))
	enc.AddString("priority_class_name", x.PriorityClassName)
	return nil
}

func (x *CreatePipelineRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddObject("pipeline", x.Pipeline)
	enc.AddObject("tf_job", x.TFJob)
	enc.AddObject("transform", x.Transform)
	enc.AddObject("parallelism_spec", x.ParallelismSpec)
	enc.AddObject("egress", x.Egress)
	enc.AddBool("update", x.Update)
	enc.AddString("output_branch", x.OutputBranch)
	enc.AddBool("s3_out", x.S3Out)
	enc.AddObject("resource_requests", x.ResourceRequests)
	enc.AddObject("resource_limits", x.ResourceLimits)
	enc.AddObject("sidecar_resource_limits", x.SidecarResourceLimits)
	enc.AddObject("input", x.Input)
	enc.AddString("description", x.Description)
	enc.AddBool("reprocess", x.Reprocess)
	enc.AddObject("service", x.Service)
	enc.AddObject("spout", x.Spout)
	enc.AddObject("datum_set_spec", x.DatumSetSpec)
	protoextensions.AddDuration(enc, "datum_timeout", x.DatumTimeout)
	protoextensions.AddDuration(enc, "job_timeout", x.JobTimeout)
	enc.AddString("salt", x.Salt)
	enc.AddInt64("datum_tries", x.DatumTries)
	enc.AddObject("scheduling_spec", x.SchedulingSpec)
	enc.AddString("pod_spec", x.PodSpec)
	enc.AddString("pod_patch", x.PodPatch)
	enc.AddObject("spec_commit", x.SpecCommit)
	enc.AddObject("metadata", x.Metadata)
	enc.AddString("reprocess_spec", x.ReprocessSpec)
	enc.AddBool("autoscaling", x.Autoscaling)
	tolerationsArrMarshaller := func(enc zapcore.ArrayEncoder) error {
		for _, v := range x.Tolerations {
			enc.AppendObject(v)
		}
		return nil
	}
	enc.AddArray("tolerations", zapcore.ArrayMarshalerFunc(tolerationsArrMarshaller))
	enc.AddObject("sidecar_resource_requests", x.SidecarResourceRequests)
	return nil
}

func (x *InspectPipelineRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddObject("pipeline", x.Pipeline)
	enc.AddBool("details", x.Details)
	return nil
}

func (x *ListPipelineRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddObject("pipeline", x.Pipeline)
	enc.AddInt64("history", x.History)
	enc.AddBool("details", x.Details)
	enc.AddString("jqFilter", x.JqFilter)
	enc.AddObject("commit_set", x.CommitSet)
	projectsArrMarshaller := func(enc zapcore.ArrayEncoder) error {
		for _, v := range x.Projects {
			enc.AppendObject(v)
		}
		return nil
	}
	enc.AddArray("projects", zapcore.ArrayMarshalerFunc(projectsArrMarshaller))
	return nil
}

func (x *DeletePipelineRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddObject("pipeline", x.Pipeline)
	enc.AddBool("all", x.All)
	enc.AddBool("force", x.Force)
	enc.AddBool("keep_repo", x.KeepRepo)
	return nil
}

func (x *DeletePipelinesRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	projectsArrMarshaller := func(enc zapcore.ArrayEncoder) error {
		for _, v := range x.Projects {
			enc.AppendObject(v)
		}
		return nil
	}
	enc.AddArray("projects", zapcore.ArrayMarshalerFunc(projectsArrMarshaller))
	enc.AddBool("force", x.Force)
	enc.AddBool("keep_repo", x.KeepRepo)
	enc.AddBool("all", x.All)
	return nil
}

func (x *DeletePipelinesResponse) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	pipelinesArrMarshaller := func(enc zapcore.ArrayEncoder) error {
		for _, v := range x.Pipelines {
			enc.AppendObject(v)
		}
		return nil
	}
	enc.AddArray("pipelines", zapcore.ArrayMarshalerFunc(pipelinesArrMarshaller))
	return nil
}

func (x *StartPipelineRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddObject("pipeline", x.Pipeline)
	return nil
}

func (x *StopPipelineRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddObject("pipeline", x.Pipeline)
	return nil
}

func (x *RunPipelineRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddObject("pipeline", x.Pipeline)
	provenanceArrMarshaller := func(enc zapcore.ArrayEncoder) error {
		for _, v := range x.Provenance {
			enc.AppendObject(v)
		}
		return nil
	}
	enc.AddArray("provenance", zapcore.ArrayMarshalerFunc(provenanceArrMarshaller))
	enc.AddString("job_id", x.JobID)
	return nil
}

func (x *RunCronRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddObject("pipeline", x.Pipeline)
	return nil
}

func (x *CreateSecretRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	protoextensions.AddBytes(enc, "file", x.File)
	return nil
}

func (x *DeleteSecretRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddObject("secret", x.Secret)
	return nil
}

func (x *InspectSecretRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddObject("secret", x.Secret)
	return nil
}

func (x *Secret) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddString("name", x.Name)
	return nil
}

func (x *SecretInfo) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddObject("secret", x.Secret)
	enc.AddString("type", x.Type)
	protoextensions.AddTimestamp(enc, "creation_timestamp", x.CreationTimestamp)
	return nil
}

func (x *SecretInfos) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	secret_infoArrMarshaller := func(enc zapcore.ArrayEncoder) error {
		for _, v := range x.SecretInfo {
			enc.AppendObject(v)
		}
		return nil
	}
	enc.AddArray("secret_info", zapcore.ArrayMarshalerFunc(secret_infoArrMarshaller))
	return nil
}

func (x *ActivateAuthRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	return nil
}

func (x *ActivateAuthResponse) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	return nil
}

func (x *RunLoadTestRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddString("dag_spec", x.DagSpec)
	enc.AddString("load_spec", x.LoadSpec)
	enc.AddInt64("seed", x.Seed)
	enc.AddInt64("parallelism", x.Parallelism)
	enc.AddString("pod_patch", x.PodPatch)
	enc.AddString("state_id", x.StateId)
	return nil
}

func (x *RunLoadTestResponse) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddString("error", x.Error)
	enc.AddString("state_id", x.StateId)
	return nil
}

func (x *RenderTemplateRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddString("template", x.Template)
	enc.AddObject("args", zapcore.ObjectMarshalerFunc(func(enc zapcore.ObjectEncoder) error {
		for k, v := range x.Args {
			enc.AddString(fmt.Sprintf("%v", k), v)
		}
		return nil
	}))
	return nil
}

func (x *RenderTemplateResponse) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddString("json", x.Json)
	specsArrMarshaller := func(enc zapcore.ArrayEncoder) error {
		for _, v := range x.Specs {
			enc.AppendObject(v)
		}
		return nil
	}
	enc.AddArray("specs", zapcore.ArrayMarshalerFunc(specsArrMarshaller))
	return nil
}

func (x *LokiRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	protoextensions.AddDuration(enc, "since", x.Since)
	enc.AddString("query", x.Query)
	return nil
}

func (x *LokiLogMessage) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddString("message", x.Message)
	return nil
}
