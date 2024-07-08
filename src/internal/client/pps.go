package client

import (
	"context"
	"io"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	// PPSEtcdPrefixEnv is the environment variable that specifies the etcd
	// prefix that PPS uses.
	PPSEtcdPrefixEnv = "PPS_ETCD_PREFIX"
	// PPSWorkerIPEnv is the environment variable that a worker can use to
	// see its own IP.  The IP address is made available through the
	// Kubernetes downward API.
	PPSWorkerIPEnv = "PPS_WORKER_IP"
	// PPSPodNameEnv is the environment variable that a pod can use to
	// see its own name.  The pod name is made available through the
	// Kubernetes downward API.
	PPSPodNameEnv = "PPS_POD_NAME"
	// PPSProjectNameEnv is the env var that sets the name of the project
	// that the workers are running.
	PPSProjectNameEnv = "PPS_PROJECT_NAME"
	// PPSPipelineNameEnv is the env var that sets the name of the pipeline
	// that the workers are running.
	PPSPipelineNameEnv = "PPS_PIPELINE_NAME"
	// PPSJobIDEnv is the env var that sets the ID of the job that the
	// workers are running (if the workers belong to an orphan job, rather than a
	// pipeline).
	PPSJobIDEnv = "PPS_JOB_ID"
	// PPSSpecCommitEnv is the namespace in which pachyderm is deployed
	PPSSpecCommitEnv = "PPS_SPEC_COMMIT"
	// PPSInputPrefix is the prefix of the path where datums are downloaded
	// to.  A datum of an input named `XXX` is downloaded to `/pfs/XXX/`.
	PPSInputPrefix = "/pfs"
	// PPSScratchSpace is where pps workers store data while it's waiting to be
	// processed.
	PPSScratchSpace = ".scratch"
	// PPSWorkerPortEnv is environment variable name for the port that workers
	// use for their gRPC server
	PPSWorkerPortEnv = "PPS_WORKER_GRPC_PORT"
	// PPSWorkerVolume is the name of the volume in which workers store
	// data.
	PPSWorkerVolume = "pachyderm-worker"
	// PPSWorkerUserContainerName is the name of the container that runs
	// the user code to process data.
	PPSWorkerUserContainerName = "user"
	// PPSWorkerSidecarContainerName is the name of the sidecar container
	// that runs alongside of each worker container.
	PPSWorkerSidecarContainerName = "storage"
	// GCGenerationKey is the etcd key that stores a counter that the
	// GC utility increments when it runs, so as to invalidate all cache.
	GCGenerationKey = "gc-generation"
	// JobIDEnv is an env var that is added to the environment of user pipeline
	// code and indicates the id of the job currently being run.
	JobIDEnv = "PACH_JOB_ID"
	// OutputCommitIDEnv is an env var that is added to the environment of user
	// pipelined code and indicates the id of the output commit.
	OutputCommitIDEnv = "PACH_OUTPUT_COMMIT_ID"
	// FilesetIDEnv is an env var that is added to the environment of user
	// pipelined code and indicates the id of the input datum fileset.
	FilesetIDEnv = "FILESET_ID"
	// DatumIDEnv is an env var that is added to the environment of user
	// pipelined code and indicates the id of the datum.
	DatumIDEnv = "PACH_DATUM_ID"
	// PeerPortEnv is the env var that sets a custom peer port
	PeerPortEnv = "PEER_PORT"

	ReprocessSpecUntilSuccess = "until_success"
	ReprocessSpecEveryJob     = "every_job"
)

// NewJob creates a pps.Job.
func NewJob(projectName, pipelineName, jobID string) *pps.Job {
	return &pps.Job{Pipeline: NewPipeline(projectName, pipelineName), Id: jobID}
}

// NewJobSet creates a pps.JobSet.
func NewJobSet(id string) *pps.JobSet {
	return &pps.JobSet{Id: id}
}

// NewPFSInput returns a new PFS input.  It only includes required options.
func NewPFSInput(project, repo, glob string) *pps.Input {
	return &pps.Input{
		Pfs: &pps.PFSInput{
			Project: project,
			Repo:    repo,
			Glob:    glob,
		},
	}
}

// NewPFSInputOpts returns a new PFS input. It includes all options.
func NewPFSInputOpts(name, project, repo, branch, glob, joinOn, groupBy string, outerJoin, lazy bool, trigger *pfs.Trigger) *pps.Input {
	return &pps.Input{
		Pfs: &pps.PFSInput{
			Name:      name,
			Project:   project,
			Repo:      repo,
			Branch:    branch,
			Glob:      glob,
			JoinOn:    joinOn,
			OuterJoin: outerJoin,
			GroupBy:   groupBy,
			Lazy:      lazy,
			Trigger:   trigger,
		},
	}
}

// NewS3PFSInput returns a new PFS input with 'S3' set.
func NewS3PFSInput(name, project, repo, branch string) *pps.Input {
	return &pps.Input{
		Pfs: &pps.PFSInput{
			Name:    name,
			Project: project,
			Repo:    repo,
			Branch:  branch,
			S3:      true,
		},
	}
}

// NewCrossInput returns an input which is the cross product of other inputs.
// That means that all combination of datums will be seen by the job /
// pipeline.
func NewCrossInput(input ...*pps.Input) *pps.Input {
	return &pps.Input{
		Cross: input,
	}
}

// NewJoinInput returns an input which is the join of other inputs.
// That means that all combination of datums which match on `joinOn` will be seen by the job /
// pipeline.
func NewJoinInput(input ...*pps.Input) *pps.Input {
	return &pps.Input{
		Join: input,
	}
}

// NewUnionInput returns an input which is the union of other inputs. That
// means that all datums from any of the inputs will be seen individually by
// the job / pipeline.
func NewUnionInput(input ...*pps.Input) *pps.Input {
	return &pps.Input{
		Union: input,
	}
}

// NewGroupInput returns an input which groups the inputs by the GroupBy pattern.
// That means that it will return a datum for each group of input datums matching
// a particular GroupBy pattern
func NewGroupInput(input ...*pps.Input) *pps.Input {
	return &pps.Input{
		Group: input,
	}
}

// NewCronInput returns an input which will trigger based on a timed schedule.
// It uses cron syntax to specify the schedule. The input will be exposed to
// jobs as `/pfs/<name>/<timestamp>`. The timestamp uses the RFC 3339 format,
// e.g. `2006-01-02T15:04:05Z07:00`. It only takes required options.
func NewCronInput(name string, spec string) *pps.Input {
	return &pps.Input{
		Cron: &pps.CronInput{
			Name: name,
			Spec: spec,
		},
	}
}

// NewCronInputOpts returns an input which will trigger based on a timed schedule.
// It uses cron syntax to specify the schedule. The input will be exposed to
// jobs as `/pfs/<name>/<timestamp>`. The timestamp uses the RFC 3339 format,
// e.g. `2006-01-02T15:04:05Z07:00`. It includes all the options.
func NewCronInputOpts(name string, repo string, spec string, overwrite bool, start *timestamppb.Timestamp) *pps.Input {
	return &pps.Input{
		Cron: &pps.CronInput{
			Name:      name,
			Repo:      repo,
			Spec:      spec,
			Overwrite: overwrite,
			Start:     start,
		},
	}
}

// NewPipeline creates a pps.Pipeline.
func NewPipeline(projectName, pipelineName string) *pps.Pipeline {
	return &pps.Pipeline{
		Project: NewProject(projectName),
		Name:    pipelineName,
	}
}

// InspectJob returns info about a specific job.
//
// 'details' indicates that the JobInfo.Details field should be filled out.
func (c APIClient) InspectJob(projectName, pipelineName, jobID string, details bool) (_ *pps.JobInfo, retErr error) {
	defer func() { retErr = grpcutil.ScrubGRPC(retErr) }()
	req := &pps.InspectJobRequest{
		Job:     NewJob(projectName, pipelineName, jobID),
		Details: details,
	}
	jobInfo, err := c.PpsAPIClient.InspectJob(c.Ctx(), req)
	return jobInfo, grpcutil.ScrubGRPC(err)
}

// WaitJob is a blocking version of InspectJob that will wait
// until the job has reached a terminal state.
func (c APIClient) WaitJob(projectName, pipelineName, jobID string, details bool) (_ *pps.JobInfo, retErr error) {
	defer func() { retErr = grpcutil.ScrubGRPC(retErr) }()
	req := &pps.InspectJobRequest{
		Job:     NewJob(projectName, pipelineName, jobID),
		Wait:    true,
		Details: details,
	}
	jobInfo, err := c.PpsAPIClient.InspectJob(c.Ctx(), req)
	return jobInfo, grpcutil.ScrubGRPC(err)
}

func (c APIClient) inspectJobSet(id string, wait bool, details bool, cb func(*pps.JobInfo) error) (retErr error) {
	ctx, cf := context.WithCancel(c.Ctx())
	defer cf()
	req := &pps.InspectJobSetRequest{
		JobSet:  NewJobSet(id),
		Wait:    wait,
		Details: details,
	}
	client, err := c.PpsAPIClient.InspectJobSet(ctx, req)
	if err != nil {
		return err
	}
	for {
		ci, err := client.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		if err := cb(ci); err != nil {
			if errors.Is(err, errutil.ErrBreak) {
				return nil
			}
			return err
		}
	}
}

func (c APIClient) InspectJobSet(id string, details bool) (_ []*pps.JobInfo, retErr error) {
	defer func() { retErr = grpcutil.ScrubGRPC(retErr) }()
	result := []*pps.JobInfo{}
	if err := c.inspectJobSet(id, false, details, func(ji *pps.JobInfo) error {
		result = append(result, ji)
		return nil
	}); err != nil {
		return nil, err
	}
	return result, nil
}

func (c APIClient) WaitJobSetAll(id string, details bool) (_ []*pps.JobInfo, retErr error) {
	defer func() { retErr = grpcutil.ScrubGRPC(retErr) }()
	result := []*pps.JobInfo{}
	if err := c.WaitJobSet(id, details, func(ji *pps.JobInfo) error {
		result = append(result, ji)
		return nil
	}); err != nil {
		return nil, err
	}
	return result, nil
}

func (c APIClient) WaitJobSet(id string, details bool, cb func(*pps.JobInfo) error) (retErr error) {
	defer func() { retErr = grpcutil.ScrubGRPC(retErr) }()
	return c.inspectJobSet(id, true, details, cb)
}

// ListJob returns info about all jobs.
//
// If projectName & pipelineName are non empty then only jobs that were started
// by the named pipeline will be returned.
//
// If inputCommit is non-nil then only jobs which took the specific commits as
// inputs will be returned.
//
// The order of the inputCommits doesn't matter.
//
// If outputCommit is non-nil then only the job which created that commit as
// output will be returned.
//
// 'history' controls whether jobs from historical versions of pipelines are
// returned, it has the following semantics:
//
//   - 0: Return jobs from the current version of the pipeline or pipelines.
//   - 1: Return the above and jobs from the next most recent version
//   - 2: etc.
//   - -1: Return jobs from all historical versions.
//
// 'details' controls whether the JobInfo passed to 'f' includes details from
// the pipeline spec (e.g. the transform). Leaving this 'false' can improve
// performance.
func (c APIClient) ListJob(projectName, pipelineName string, inputCommit []*pfs.Commit, history int64, details bool) ([]*pps.JobInfo, error) {
	var result []*pps.JobInfo
	if err := c.ListJobF(projectName, pipelineName, inputCommit, history, details,
		func(ji *pps.JobInfo) error {
			result = append(result, ji)
			return nil
		}); err != nil {
		return nil, err
	}
	return result, nil
}

// ListJobF is a previous version of ListJobFilterF, returning info about all jobs
// and calling f on each JobInfo.
func (c APIClient) ListJobF(projectName, pipelineName string, inputCommit []*pfs.Commit,
	history int64, details bool,
	f func(*pps.JobInfo) error) error {
	return c.ListProjectJobFilterF(projectName, pipelineName, inputCommit, history, details, "", f)
}

// ListJobFilterF returns info about all jobs, calling f with each JobInfo.
//
// If f returns an error iteration of jobs will stop and ListJobF will return
// that error, unless the error is errutil.ErrBreak in which case it will return
// nil.
//
// If pipelineName is non empty then only jobs that were started by the named
// pipeline will be returned.
//
// If inputCommit is non-nil then only jobs which took the specific commits as
// inputs will be returned.
//
// The order of the inputCommits doesn't matter.
//
// If outputCommit is non-nil then only the job which created that commit as
// output will be returned.
//
// 'history' controls whether jobs from historical versions of pipelines are
// returned, it has the following semantics:
//
//   - 0: Return jobs from the current version of the pipeline or pipelines.
//   - 1: Return the above and jobs from the next most recent version
//   - 2: etc.
//   - -1: Return jobs from all historical versions.
//
// 'details' controls whether the JobInfo passed to 'f' includes details from
// the pipeline spec--setting this to 'false' can improve performance.
func (c APIClient) ListJobFilterF(pipelineName string, inputCommit []*pfs.Commit,
	history int64, details bool, jqFilter string,
	f func(*pps.JobInfo) error) error {
	return c.ListProjectJobFilterF(pfs.DefaultProjectName, pipelineName, inputCommit, history, details, jqFilter, f)
}

// ListProjectJobFilterF returns info about all jobs, calling f with each JobInfo.
//
// If f returns an error iteration of jobs will stop and ListJobF will return
// that error, unless the error is errutil.ErrBreak in which case it will return
// nil.
//
// If projectName & pipelineName are both non-empty then only jobs that were
// started by the named pipeline will be returned.
//
// If inputCommit is non-nil then only jobs which took the specific commits as
// inputs will be returned.
//
// The order of the inputCommits doesn't matter.
//
// If outputCommit is non-nil then only the job which created that commit as
// output will be returned.
//
// 'history' controls whether jobs from historical versions of pipelines are
// returned, it has the following semantics:
//
//   - 0: Return jobs from the current version of the pipeline or pipelines.
//   - 1: Return the above and jobs from the next most recent version
//   - 2: etc.
//   - -1: Return jobs from all historical versions.
//
// 'details' controls whether the JobInfo passed to 'f' includes details from
// the pipeline spec--setting this to 'false' can improve performance.
func (c APIClient) ListProjectJobFilterF(projectName, pipelineName string, inputCommit []*pfs.Commit,
	history int64, details bool, jqFilter string,
	f func(*pps.JobInfo) error) error {
	var pipeline *pps.Pipeline
	if projectName != "" && pipelineName != "" {
		pipeline = NewPipeline(projectName, pipelineName)
	}
	ctx, cf := context.WithCancel(c.Ctx())
	defer cf()
	client, err := c.PpsAPIClient.ListJob(
		ctx,
		&pps.ListJobRequest{
			Pipeline:    pipeline,
			InputCommit: inputCommit,
			History:     history,
			Details:     details,
			JqFilter:    jqFilter,
		})
	if err != nil {
		return grpcutil.ScrubGRPC(err)
	}
	for {
		ji, err := client.Recv()
		if errors.Is(err, io.EOF) {
			return nil
		} else if err != nil {
			return grpcutil.ScrubGRPC(err)
		}
		if err := f(ji); err != nil {
			if errors.Is(err, errutil.ErrBreak) {
				return nil
			}
			return err
		}
	}
}

// SubscribeJob calls the given callback with each open job in the given
// pipeline until cancelled.
func (c APIClient) SubscribeJob(projectName, pipelineName string, details bool, cb func(*pps.JobInfo) error) error {
	ctx, cf := context.WithCancel(c.Ctx())
	defer cf()
	client, err := c.PpsAPIClient.SubscribeJob(
		ctx,
		&pps.SubscribeJobRequest{
			Pipeline: NewPipeline(projectName, pipelineName),
			Details:  details,
		})
	if err != nil {
		return grpcutil.ScrubGRPC(err)
	}
	for {
		ji, err := client.Recv()
		if errors.Is(err, io.EOF) {
			return nil
		} else if err != nil {
			return grpcutil.ScrubGRPC(err)
		}
		if err := cb(ji); err != nil {
			if errors.Is(err, errutil.ErrBreak) {
				return nil
			}
			return err
		}
	}
}

// DeleteJob deletes a job.
func (c APIClient) DeleteJob(projectName, pipelineName, jobID string) error {
	_, err := c.PpsAPIClient.DeleteJob(
		c.Ctx(),
		&pps.DeleteJobRequest{
			Job: NewJob(projectName, pipelineName, jobID),
		},
	)
	return grpcutil.ScrubGRPC(err)
}

// StopJob stops a job.
func (c APIClient) StopJob(projectName, pipelineName, jobID string) error {
	_, err := c.PpsAPIClient.StopJob(
		c.Ctx(),
		&pps.StopJobRequest{
			Job: NewJob(projectName, pipelineName, jobID),
		},
	)
	return grpcutil.ScrubGRPC(err)
}

// RestartDatum restarts a datum that's being processed as part of a job.
//
// datumFilter is a slice of strings which are matched against either the Path
// or Hash of the datum, the order of the strings in datumFilter is irrelevant.
func (c APIClient) RestartDatum(projectName, pipelineName, jobID string, datumFilter []string) error {
	_, err := c.PpsAPIClient.RestartDatum(
		c.Ctx(),
		&pps.RestartDatumRequest{
			Job:         NewJob(projectName, pipelineName, jobID),
			DataFilters: datumFilter,
		},
	)
	return grpcutil.ScrubGRPC(err)
}

// ListDatum returns info about datums in a job.
func (c APIClient) ListDatum(projectName, pipelineName, jobID string, cb func(*pps.DatumInfo) error) (retErr error) {
	defer func() {
		retErr = grpcutil.ScrubGRPC(retErr)
	}()
	req := &pps.ListDatumRequest{
		Job: NewJob(projectName, pipelineName, jobID),
	}
	return c.listDatum(req, cb)
}

// ListDatumAll returns info about datums in a job.
func (c APIClient) ListDatumAll(projectName, pipelineName, jobID string) (_ []*pps.DatumInfo, retErr error) {
	defer func() {
		retErr = grpcutil.ScrubGRPC(retErr)
	}()
	var dis []*pps.DatumInfo
	if err := c.ListDatum(projectName, pipelineName, jobID, func(di *pps.DatumInfo) error {
		dis = append(dis, di)
		return nil
	}); err != nil {
		return nil, err
	}
	return dis, nil
}

// ListDatumInput returns info about datums for a pipeline with input.  The
// pipeline doesn't need to exist.
func (c APIClient) ListDatumInput(input *pps.Input, cb func(*pps.DatumInfo) error) (retErr error) {
	defer func() {
		retErr = grpcutil.ScrubGRPC(retErr)
	}()
	req := &pps.ListDatumRequest{
		Input: input,
	}
	return c.listDatum(req, cb)
}

// ListDatumInputAll returns info about datums for a pipeline with input. The
// pipeline doesn't need to exist.
func (c APIClient) ListDatumInputAll(input *pps.Input) (_ []*pps.DatumInfo, retErr error) {
	defer func() {
		retErr = grpcutil.ScrubGRPC(retErr)
	}()
	var dis []*pps.DatumInfo
	if err := c.ListDatumInput(input, func(di *pps.DatumInfo) error {
		dis = append(dis, di)
		return nil
	}); err != nil {
		return nil, err
	}
	return dis, nil
}

func (c APIClient) listDatum(req *pps.ListDatumRequest, cb func(*pps.DatumInfo) error) (retErr error) {
	ctx, cf := context.WithCancel(c.Ctx())
	defer cf()
	client, err := c.PpsAPIClient.ListDatum(ctx, req)
	if err != nil {
		return err
	}
	for {
		di, err := client.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		if err := cb(di); err != nil {
			if errors.Is(err, errutil.ErrBreak) {
				return nil
			}
			return err
		}
	}
}

// InspectDatum returns info about a single datum.
func (c APIClient) InspectDatum(projectName, pipelineName, jobID, datumID string) (*pps.DatumInfo, error) {
	datumInfo, err := c.PpsAPIClient.InspectDatum(
		c.Ctx(),
		&pps.InspectDatumRequest{
			Datum: &pps.Datum{
				Id:  datumID,
				Job: NewJob(projectName, pipelineName, jobID),
			},
		},
	)
	if err != nil {
		return nil, grpcutil.ScrubGRPC(err)
	}
	return datumInfo, nil
}

// LogsIter iterates through log messages returned from pps.GetLogs. Logs can
// be fetched with 'Next()'. The log message received can be examined with
// 'Message()', and any errors can be examined with 'Err()'.
type LogsIter struct {
	logsClient pps.API_GetLogsClient
	msg        *pps.LogMessage
	err        error
}

// Next retrieves the next relevant log message from pachd
func (l *LogsIter) Next() bool {
	if l.err != nil {
		l.msg = nil
		return false
	}
	l.msg, l.err = l.logsClient.Recv()
	return l.err == nil
}

// Message returns the most recently retrieve log message (as an annotated log
// line, in the form of a pps.LogMessage)
func (l *LogsIter) Message() *pps.LogMessage {
	return l.msg
}

// Err retrieves any errors encountered in the course of calling 'Next()'.
func (l *LogsIter) Err() error {
	if errors.Is(l.err, io.EOF) {
		return nil
	}
	return grpcutil.ScrubGRPC(l.err)
}

// GetLogs gets logs from a job (logs includes stdout and stderr).
// 'pipelineName', 'jobID', 'data', and 'datumID', are all filters.  To forego
// any filter, simply pass an empty value, though one of 'pipelineName' and
// 'jobID' must be set.  Responses are written to 'messages'.
func (c APIClient) GetLogs(projectName, pipelineName, jobID string, data []string, datumID string, master, follow bool, since time.Duration) *LogsIter {
	return c.getLogs(projectName, pipelineName, jobID, data, datumID, master, follow, since, false)
}

// GetLogsLoki gets logs from a job (logs includes stdout and stderr).
// 'pipelineName', 'jobID', 'data', and 'datumID', are all filters.  To forego
// any filter, simply pass an empty value, though one of 'pipelineName' and
// 'jobID' must be set.  Responses are written to 'messages'.
func (c APIClient) GetLogsLoki(
	pipelineName string,
	jobID string,
	data []string,
	datumID string,
	master bool,
	follow bool,
	since time.Duration,
) *LogsIter {
	return c.GetProjectLogsLoki(pfs.DefaultProjectName, pipelineName, jobID, data, datumID, master, follow, since)
}

// GetProjectLogsLoki gets logs from a job (logs includes stdout and stderr).
// 'pipelineName', 'jobID', 'data', and 'datumID', are all filters.  To forego
// any filter, simply pass an empty value, though one of 'pipelineName' and
// 'jobID' must be set.  Responses are written to 'messages'.
func (c APIClient) GetProjectLogsLoki(projectName, pipelineName, jobID string, data []string, datumID string, master, follow bool, since time.Duration) *LogsIter {
	return c.getLogs("", pipelineName, jobID, data, datumID, master, follow, since, true)
}

func (c APIClient) getLogs(projectName, pipelineName, jobID string, data []string, datumID string, master, follow bool, since time.Duration, useLoki bool) *LogsIter {
	request := pps.GetLogsRequest{
		Master:         master,
		Follow:         follow,
		UseLokiBackend: useLoki,
		Since:          durationpb.New(since),
	}
	if pipelineName != "" {
		request.Pipeline = NewPipeline(projectName, pipelineName)
	}
	if jobID != "" {
		request.Job = NewJob(projectName, pipelineName, jobID)
	}
	request.DataFilters = data
	if datumID != "" {
		request.Datum = &pps.Datum{
			Job: NewJob(projectName, pipelineName, jobID),
			Id:  datumID,
		}
	}
	resp := &LogsIter{}
	resp.logsClient, resp.err = c.PpsAPIClient.GetLogs(c.Ctx(), &request)
	resp.err = grpcutil.ScrubGRPC(resp.err)
	return resp
}

// CreatePipeline creates a new pipeline, pipelines are the main computation
// object in PPS they create a flow of data from a set of input Repos to an
// output Repo (which has the same name as the pipeline).  Whenever new data is
// committed to one of the input repos the pipelines will create jobs to bring
// the output Repo up to data.
//
// image is the Docker image to run the jobs in.
//
// cmd is the command passed to the Docker run invocation.  NOTE as with Docker
// cmd is not run inside a shell that means that things like wildcard globbing
// (*), pipes (|) and file redirects (> and >>) will not work.  To get that
// behavior you should have your command be a shell of your choice and pass a
// shell script to stdin.
//
// stdin is a slice of lines that are sent to your command on stdin.  Lines need
// not end in newline characters.
//
// parallelism is how many copies of your container should run in parallel.  You
// may pass 0 for parallelism in which case PPS will set the parallelism based
// on available resources.
//
// input specifies a set of Repos that will be visible to the jobs during
// runtime.  commits to these repos will cause the pipeline to create new jobs
// to process them.  update indicates that you want to update an existing
// pipeline.
func (c APIClient) CreatePipeline(projectName, pipelineName, image string, cmd []string, stdin []string, parallelismSpec *pps.ParallelismSpec, input *pps.Input, outputBranch string, update bool) error {
	if image == "" {
		image = c.defaultTransformImage
	}
	_, err := c.PpsAPIClient.CreatePipeline(
		c.Ctx(),
		&pps.CreatePipelineRequest{
			Pipeline: NewPipeline(projectName, pipelineName),
			Transform: &pps.Transform{
				Image: image,
				Cmd:   cmd,
				Stdin: stdin,
			},
			ParallelismSpec: parallelismSpec,
			Input:           input,
			OutputBranch:    outputBranch,
			Update:          update,
		},
	)
	return grpcutil.ScrubGRPC(err)
}

// InspectPipeline returns info about a specific pipeline.  The name may
// include ancestry syntax or be a bare name.
func (c APIClient) InspectPipeline(projectName, pipelineName string, details bool) (*pps.PipelineInfo, error) {
	pipelineInfo, err := c.PpsAPIClient.InspectPipeline(
		c.Ctx(),
		&pps.InspectPipelineRequest{
			Pipeline: NewPipeline(projectName, pipelineName),
			Details:  details,
		},
	)
	return pipelineInfo, grpcutil.ScrubGRPC(err)
}

// ListPipeline returns info about all pipelines.
func (c APIClient) ListPipeline() ([]*pps.PipelineInfo, error) {
	ctx, cf := context.WithCancel(c.Ctx())
	defer cf()
	client, err := c.PpsAPIClient.ListPipeline(
		ctx,
		&pps.ListPipelineRequest{},
	)
	if err != nil {
		return nil, grpcutil.ScrubGRPC(err)
	}
	return grpcutil.Collect[*pps.PipelineInfo](client, 1000)
}

// ListProjectPipelineHistory returns historical information about pipelines.
//
// `pipelineName` specifies which pipeline to return history about, if it's equal
// to "" then ListPipelineHistory returns historical information about all
// pipelines.
//
// `history` specifies how many historical revisions to return:

// - 0: Return the current version of the pipeline or pipelines.
// - 1: Return the above and the next most recent version
// - 2: etc.
// - -1: Return all historical versions.
func (c APIClient) ListPipelineHistory(projectName, pipelineName string, history int64) ([]*pps.PipelineInfo, error) {
	var pipeline *pps.Pipeline
	if pipelineName != "" {
		pipeline = NewPipeline(projectName, pipelineName)
	}
	ctx, cf := context.WithCancel(c.Ctx())
	defer cf()
	client, err := c.PpsAPIClient.ListPipeline(
		ctx,
		&pps.ListPipelineRequest{
			Pipeline: pipeline,
			History:  history,
		},
	)
	if err != nil {
		return nil, grpcutil.ScrubGRPC(err)
	}
	return grpcutil.Collect[*pps.PipelineInfo](client, 1000)
}

// DeletePipeline deletes a pipeline along with its output Repo.
func (c APIClient) DeletePipeline(projectName, pipelineName string, force bool) error {
	req := &pps.DeletePipelineRequest{
		Pipeline: NewPipeline(projectName, pipelineName),
		Force:    force,
	}
	_, err := c.PpsAPIClient.DeletePipeline(
		c.Ctx(),
		req,
	)
	return grpcutil.ScrubGRPC(err)
}

// StartPipeline restarts a stopped pipeline.
func (c APIClient) StartPipeline(projectName, pipelineName string) error {
	_, err := c.PpsAPIClient.StartPipeline(
		c.Ctx(),
		&pps.StartPipelineRequest{
			Pipeline: NewPipeline(projectName, pipelineName),
		},
	)
	return grpcutil.ScrubGRPC(err)
}

// StopPipeline prevents a pipeline from processing things; it can be
// restarted with StartProjectPipeline.
func (c APIClient) StopPipeline(projectName, pipelineName string) error {
	_, err := c.PpsAPIClient.StopPipeline(
		c.Ctx(),
		&pps.StopPipelineRequest{
			Pipeline: NewPipeline(projectName, pipelineName),
		},
	)
	return grpcutil.ScrubGRPC(err)
}

// RunPipeline runs a pipeline.  It can be passed a list of commit
// provenance.  This will trigger a new job provenant on those commits,
// effectively running the pipeline on the data in those commits.
func (c APIClient) RunPipeline(projectName, pipelineName string, provenance []*pfs.Commit, jobID string) error {
	_, err := c.PpsAPIClient.RunPipeline(
		c.Ctx(),
		&pps.RunPipelineRequest{
			Pipeline:   NewPipeline(projectName, pipelineName),
			Provenance: provenance,
			JobId:      jobID,
		},
	)
	return grpcutil.ScrubGRPC(err)
}

// RunCron runs a pipeline.  It can be passed a list of commit
// provenance.  This will trigger a new job provenant on those commits,
// effectively running the pipeline on the data in those commits.
func (c APIClient) RunCron(projectName, pipelineName string) error {
	_, err := c.PpsAPIClient.RunCron(
		c.Ctx(),
		&pps.RunCronRequest{
			Pipeline: NewPipeline(projectName, pipelineName),
		},
	)
	return grpcutil.ScrubGRPC(err)
}

// CreateSecret creates a secret on the cluster.
func (c APIClient) CreateSecret(file []byte) error {
	_, err := c.PpsAPIClient.CreateSecret(
		c.Ctx(),
		&pps.CreateSecretRequest{
			File: file,
		},
	)
	return grpcutil.ScrubGRPC(err)
}

// DeleteSecret deletes a secret from the cluster.
func (c APIClient) DeleteSecret(secret string) error {
	_, err := c.PpsAPIClient.DeleteSecret(
		c.Ctx(),
		&pps.DeleteSecretRequest{
			Secret: &pps.Secret{Name: secret},
		},
	)
	return grpcutil.ScrubGRPC(err)
}

// InspectSecret returns info about a specific secret.
func (c APIClient) InspectSecret(secret string) (*pps.SecretInfo, error) {
	secretInfo, err := c.PpsAPIClient.InspectSecret(
		c.Ctx(),
		&pps.InspectSecretRequest{
			Secret: &pps.Secret{Name: secret},
		},
	)
	return secretInfo, grpcutil.ScrubGRPC(err)
}

// ListSecret returns info about all Pachyderm secrets.
func (c APIClient) ListSecret() ([]*pps.SecretInfo, error) {
	secretInfos, err := c.PpsAPIClient.ListSecret(
		c.Ctx(),
		&emptypb.Empty{},
	)
	if err != nil {
		return nil, grpcutil.ScrubGRPC(err)
	}
	return secretInfos.SecretInfo, nil
}

// CreatePipelineService creates a new pipeline service.
func (c APIClient) CreatePipelineService(projectName, pipelineName, image string, cmd, stdin []string, parallelismSpec *pps.ParallelismSpec, input *pps.Input, update bool, internalPort, externalPort int32, annotations map[string]string) error {
	if image == "" {
		image = c.defaultTransformImage
	}
	_, err := c.PpsAPIClient.CreatePipeline(
		c.Ctx(),
		&pps.CreatePipelineRequest{
			Pipeline: NewPipeline(projectName, pipelineName),
			Metadata: &pps.Metadata{
				Annotations: annotations,
			},
			Transform: &pps.Transform{
				Image: image,
				Cmd:   cmd,
				Stdin: stdin,
			},
			ParallelismSpec: parallelismSpec,
			Input:           input,
			Update:          update,
			Service: &pps.Service{
				InternalPort: internalPort,
				ExternalPort: externalPort,
			},
		},
	)
	return grpcutil.ScrubGRPC(err)
}

// WithDefaultTransformImage sets the image used when the empty string ""
// is passed as the image in calls to CreatePipeline*
func (c APIClient) WithDefaultTransformImage(x string) *APIClient {
	c2 := c
	c2.defaultTransformImage = x
	return &c2
}

// WithDefaultTransformUser sets the user to run the transform container as.
// This overrides the user set by the image.
func (c APIClient) WithDefaultTransformUser(x string) *APIClient {
	c2 := c
	c2.defaultTransformUser = x
	return &c2
}

// GetDatumTotalTime sums the timing stats from a DatumInfo
func GetDatumTotalTime(s *pps.ProcessStats) time.Duration {
	totalDuration := time.Duration(0)
	totalDuration += s.DownloadTime.AsDuration()
	totalDuration += s.ProcessTime.AsDuration()
	totalDuration += s.UploadTime.AsDuration()
	return totalDuration
}
