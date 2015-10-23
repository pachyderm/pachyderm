package jobserver

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/satori/go.uuid"

	"go.pachyderm.com/pachyderm/src/pkg/container"
	"go.pachyderm.com/pachyderm/src/pps"
	"go.pachyderm.com/pachyderm/src/pps/convert"
	"go.pachyderm.com/pachyderm/src/pps/persist"
	"go.pedge.io/google-protobuf"
	"go.pedge.io/proto/rpclog"
	"go.pedge.io/proto/time"
	"go.pedge.io/protolog"
	"golang.org/x/net/context"
)

type apiServer struct {
	protorpclog.Logger
	persistAPIClient persist.APIClient
	containerClient  container.Client
}

func newAPIServer(
	persistAPIClient persist.APIClient,
	containerClient container.Client,
) *apiServer {
	return &apiServer{
		protorpclog.NewLogger("pachyderm.pps.JobAPI"),
		persistAPIClient,
		containerClient,
	}
}

func (a *apiServer) CreateJob(ctx context.Context, request *pps.CreateJobRequest) (response *pps.Job, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	persistJobInfo := &persist.JobInfo{
		Input: request.Input,
		OutputParent: request.OutputParent
	}
	if request.GetTransform() != nil {
		persist.Spec = &persist.JobInfo_Transform{
			Transform: request.GetTransform(),
		}
	} else if request.GetPipeline() != nil {
		persist.Spec = &persist.JobInfo_PipelineName{
			PipelineName: request.GetPipeline().Name,
		}
	} else {
		return nil, fmt.Errorf("pachyderm.pps.jobserver: both transform and pipeline are not set on %v", request)
	}
	persistJobInfo, err := a.persistAPIClient.CreateJobInfo(ctx, persistJobInfo)
	if err != nil {
		return nil, err
	}
	return &pps.Job{
		Id: persistJobInfo.Id,
	}, nil
}

func (a *apiServer) InspectJob(ctx context.Context, request *pps.InspectJobRequest) (response *pps.JobInfo, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	persistJobInfo, err := a.persistAPIClient.GetJobInfo(ctx, request.Job)
	if err != nil {
		return nil, err
	}
	persistJobStatuses, err := a.persistAPIClient.GetJobStatuses(ctx, request.Job)
	if err != nil {
		return nil, err
	}
	persistJobOutput, err := a.persistAPI.GetJobOutput(ctx, request.Job)
	if err != nil {
		return nil, err
	}
	jobInfo := &pps.JobInfo{
		Job: request.Job,
		Input: persistJobInfo.Input,
	}
	if persistJobInfo.GetTransform() != nil {
		jobInfo.Spec = &pps.JobInfo_Transform{
			Transform: persistJobInfo.GetTransform(),
		}
	}
	if persistJobInfo.GetPipelineName() != "" {
		jobInfo.Spec = &pps.JobInfo_Pipeline{
			Pipeline: &pps.Pipeline{
				Name: persistJobInfo.GetPipelineName(),
			},
		}
	}
	jobInfo.JobStatus = make([]*pps.JobStatus, len(persistJobStatuses.JobStatus))
	for i, persistJobStatus := range persistJobStatuses.JobStatus {
		jobInfo.JobStatus[i] = &pps.JobStatus{
			Type: persistJobStatus.Type,
			Timestamp: persistJobStatus.Timestamp,
			Message: persistJobStatus.Message,
		}
	}
	if persistJobOutput != nil {
		jobInfo.Output = persistJobOutput.Output
	}
	return jobInfo, nil
}

func (a *apiServer) GetJobsByPipelineName(ctx context.Context, request *pps.GetJobsByPipelineNameRequest) (response *pps.Jobs, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	persistPipelines, err := a.persistAPIClient.GetPipelinesByName(ctx, &google_protobuf.StringValue{Value: request.PipelineName})
	if err != nil {
		return nil, err
	}
	var persistJobs []*persist.Job
	for _, persistPipeline := range persistPipelines.Pipeline {
		iPersistJobs, err := a.persistAPIClient.GetJobsByPipelineID(ctx, &google_protobuf.StringValue{Value: persistPipeline.Id})
		if err != nil {
			return nil, err
		}
		persistJobs = append(persistJobs, iPersistJobs.Job...)
	}
	// TODO(pedge): could do a smart merge since jobs are already sorted in this order for each call,
	// or if we eliminate many pipelines per name, this is not needed
	sort.Sort(jobsByCreatedAtDesc(persistJobs))
	jobs := convert.PersistToJobs(&persist.Jobs{Job: persistJobs})
	return jobs, nil
}

func (a *apiServer) StartJob(ctx context.Context, request *pps.StartJobRequest) (response *google_protobuf.Empty, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	persistJob, err := a.persistAPIClient.GetJobByID(ctx, &google_protobuf.StringValue{Value: request.JobId})
	if err != nil {
		return nil, err
	}
	if err := a.startPersistJob(persistJob); err != nil {
		return nil, err
	}
	return google_protobuf.EmptyInstance, nil
}

func (a *apiServer) GetJobStatus(ctx context.Context, request *pps.GetJobStatusRequest) (response *pps.JobStatus, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	persistJobStatuses, err := a.persistAPIClient.GetJobStatusesByJobID(ctx, &google_protobuf.StringValue{Value: request.JobId})
	if err != nil {
		return nil, err
	}
	if len(persistJobStatuses.JobStatus) == 0 {
		return nil, fmt.Errorf("pachyderm.pps.server: no job statuses for %s", request.JobId)
	}
	return convert.PersistToJobStatus(persistJobStatuses.JobStatus[0]), nil
}

func (a *apiServer) GetJobLogs(request *pps.GetJobLogsRequest, responseServer pps.JobAPI_GetJobLogsServer) (err error) {
	persistJobLogs, err := a.persistAPIClient.GetJobLogsByJobID(context.Background(), &google_protobuf.StringValue{Value: request.JobId})
	if err != nil {
		return err
	}
	for _, persistJobLog := range persistJobLogs.JobLog {
		if persistJobLog.OutputStream == request.OutputStream {
			if err := responseServer.Send(&google_protobuf.BytesValue{Value: persistJobLog.Value}); err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *apiServer) startPersistJob(persistJob *persist.Job) error {
	if _, err := a.persistAPIClient.CreateJobStatus(
		context.Background(),
		&persist.JobStatus{
			JobId: persistJob.Id,
			Type:  pps.JobStatusType_JOB_STATUS_TYPE_STARTED,
		},
	); err != nil {
		return err
	}
	// TODO(pedge): throttling? worker pool?
	go func() {
		if err := a.runJob(persistJob); err != nil {
			protolog.Errorln(err.Error())
			// TODO(pedge): how to handle the error?
			if _, err = a.persistAPIClient.CreateJobStatus(
				context.Background(),
				&persist.JobStatus{
					JobId:   persistJob.Id,
					Type:    pps.JobStatusType_JOB_STATUS_TYPE_ERROR,
					Message: err.Error(),
				},
			); err != nil {
				protolog.Errorln(err.Error())
			}
		} else {
			// TODO(pedge): how to handle the error?
			if _, err = a.persistAPIClient.CreateJobStatus(
				context.Background(),
				&persist.JobStatus{
					JobId: persistJob.Id,
					Type:  pps.JobStatusType_JOB_STATUS_TYPE_SUCCESS,
				},
			); err != nil {
				protolog.Errorln(err.Error())
			}
		}
	}()
	return nil
}

func (a *apiServer) runJob(persistJob *persist.Job) error {
	switch {
	case persistJob.GetTransform() != nil:
		return a.reallyRunJob(strings.Replace(uuid.NewV4().String(), "-", "", -1), persistJob.Id, persistJob.GetTransform(), persistJob.JobInput, persistJob.JobOutput, 1)
	case persistJob.GetPipelineId() != "":
		persistPipeline, err := a.persistAPIClient.GetPipelineByID(
			context.Background(),
			&google_protobuf.StringValue{Value: persistJob.GetPipelineId()},
		)
		if err != nil {
			return err
		}
		if persistPipeline.Transform == nil {
			return fmt.Errorf("pachyderm.pps.server: transform not set on pipeline %v", persistPipeline)
		}
		return a.reallyRunJob(persistPipeline.Name, persistJob.Id, persistPipeline.Transform, persistJob.JobInput, persistJob.JobOutput, 1)
	default:
		return fmt.Errorf("pachyderm.pps.server: neither transform or pipeline id set on job %v", persistJob)
	}
}

func (a *apiServer) reallyRunJob(
	name string,
	jobID string,
	transform *pps.Transform,
	jobInputs []*pps.JobInput,
	jobOutputs []*pps.JobOutput,
	numContainers int,
) error {
	image, err := a.buildOrPull(name, transform)
	if err != nil {
		return err
	}
	var containerIDs []string
	defer a.removeContainers(containerIDs)
	for i := 0; i < numContainers; i++ {
		containerID, err := a.containerClient.Create(
			image,
			container.CreateOptions{
				Binds:      append(getInputBinds(jobInputs), getOutputBinds(jobOutputs)...),
				HasCommand: len(transform.Cmd) > 0,
			},
		)
		if err != nil {
			return err
		}
		containerIDs = append(containerIDs, containerID)
	}
	for _, containerID := range containerIDs {
		if err := a.containerClient.Start(
			containerID,
			container.StartOptions{
				Commands: transform.Cmd,
			},
		); err != nil {
			return err
		}
	}
	errC := make(chan error, len(containerIDs))
	for _, containerID := range containerIDs {
		go a.writeContainerLogs(containerID, jobID, errC)
	}
	for _, containerID := range containerIDs {
		if err := a.containerClient.Wait(containerID, container.WaitOptions{}); err != nil {
			return err
		}
	}
	err = nil
	for _ = range containerIDs {
		if logsErr := <-errC; logsErr != nil && err == nil {
			err = logsErr
		}
	}
	return err
}

// return image name
func (a *apiServer) buildOrPull(name string, transform *pps.Transform) (string, error) {
	image := transform.Image
	if transform.Build != "" {
		image = fmt.Sprintf("ppspipelines/%s", name)
		if err := a.containerClient.Build(
			image,
			transform.Build,
			// TODO(pedge): this will not work, the path to a dockerfile is not real
			container.BuildOptions{
				Dockerfile:   transform.Dockerfile,
				OutputStream: ioutil.Discard,
			},
		); err != nil {
			return "", err
		}
	} else if err := a.containerClient.Pull(
		transform.Image,
		container.PullOptions{},
	); err != nil {
		return "", err
	}
	return image, nil
}

func (a *apiServer) removeContainers(containerIDs []string) {
	for _, containerID := range containerIDs {
		_ = a.containerClient.Kill(containerID, container.KillOptions{})
		//_ = a.containerClient.Remove(containerID, container.RemoveOptions{})
	}
}

func (a *apiServer) writeContainerLogs(containerID string, jobID string, errC chan error) {
	errC <- a.containerClient.Logs(
		containerID,
		container.LogsOptions{
			Stdout: newJobLogWriter(
				jobID,
				pps.OutputStream_OUTPUT_STREAM_STDOUT,
				a.persistAPIClient,
			),
			Stderr: newJobLogWriter(
				jobID,
				pps.OutputStream_OUTPUT_STREAM_STDERR,
				a.persistAPIClient,
			),
		},
	)
}

func getInputBinds(jobInputs []*pps.JobInput) []string {
	var binds []string
	for _, jobInput := range jobInputs {
		if jobInput.GetHostDir() != "" {
			binds = append(binds, getBinds(jobInput.GetHostDir(), filepath.Join("/var/lib/pps/host", jobInput.GetHostDir()), "ro"))
		}
	}
	return binds
}

func getOutputBinds(jobOutputs []*pps.JobOutput) []string {
	var binds []string
	for _, jobOutput := range jobOutputs {
		if jobOutput.GetHostDir() != "" {
			binds = append(binds, getBinds(jobOutput.GetHostDir(), filepath.Join("/var/lib/pps/host", jobOutput.GetHostDir()), "rw"))
		}
	}
	return binds
}

func getBinds(from string, to string, postfix string) string {
	return fmt.Sprintf("%s:%s:%s", from, to, postfix)
}

type jobsByCreatedAtDesc []*persist.Job

func (s jobsByCreatedAtDesc) Len() int      { return len(s) }
func (s jobsByCreatedAtDesc) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s jobsByCreatedAtDesc) Less(i, j int) bool {
	return prototime.TimestampLess(s[j].CreatedAt, s[i].CreatedAt)
}
