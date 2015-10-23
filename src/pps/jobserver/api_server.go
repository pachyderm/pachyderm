package jobserver

import (
	"fmt"
	"strings"
	"time"

	"github.com/satori/go.uuid"

	"go.pachyderm.com/pachyderm/src/pfs"
	"go.pachyderm.com/pachyderm/src/pkg/container"
	"go.pachyderm.com/pachyderm/src/pps"
	"go.pachyderm.com/pachyderm/src/pps/persist"
	"go.pedge.io/google-protobuf"
	"go.pedge.io/proto/rpclog"
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
		Input:        request.Input,
		OutputParent: request.OutputParent,
	}
	if request.GetTransform() != nil {
		persistJobInfo.Spec = &persist.JobInfo_Transform{
			Transform: request.GetTransform(),
		}
	} else if request.GetPipeline() != nil {
		persistJobInfo.Spec = &persist.JobInfo_PipelineName{
			PipelineName: request.GetPipeline().Name,
		}
	} else {
		return nil, fmt.Errorf("pachyderm.pps.jobserver: both transform and pipeline are not set on %v", request)
	}
	persistJobInfo, err = a.persistAPIClient.CreateJobInfo(ctx, persistJobInfo)
	if err != nil {
		return nil, err
	}
	if err := a.startPersistJobInfo(persistJobInfo); err != nil {
		// TODO(pedge): proper rollback
		if _, rollbackErr := a.persistAPIClient.DeleteJobInfo(
			ctx,
			&pps.Job{
				Id: persistJobInfo.JobId,
			},
		); rollbackErr != nil {
			return nil, fmt.Errorf("%v", []error{err, rollbackErr})
		}
		return nil, err
	}
	return &pps.Job{
		Id: persistJobInfo.JobId,
	}, nil
}

func (a *apiServer) InspectJob(ctx context.Context, request *pps.InspectJobRequest) (response *pps.JobInfo, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	persistJobInfo, err := a.persistAPIClient.GetJobInfo(ctx, request.Job)
	if err != nil {
		return nil, err
	}
	return a.persistJobInfoToJobInfo(ctx, persistJobInfo)
}

func (a *apiServer) ListJob(ctx context.Context, request *pps.ListJobRequest) (response *pps.JobInfos, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	var persistJobInfos *persist.JobInfos
	if request.Pipeline == nil {
		persistJobInfos, err = a.persistAPIClient.ListJobInfos(ctx, google_protobuf.EmptyInstance)
	} else {
		persistJobInfos, err = a.persistAPIClient.GetJobInfosByPipeline(ctx, request.Pipeline)
	}
	if err != nil {
		return nil, err
	}
	jobInfos := make([]*pps.JobInfo, len(persistJobInfos.JobInfo))
	for i, persistJobInfo := range persistJobInfos.JobInfo {
		jobInfo, err := a.persistJobInfoToJobInfo(ctx, persistJobInfo)
		if err != nil {
			return nil, err
		}
		jobInfos[i] = jobInfo
	}
	return &pps.JobInfos{
		JobInfo: jobInfos,
	}, nil
}

// TODO(pedge): bulk get
func (a *apiServer) persistJobInfoToJobInfo(ctx context.Context, persistJobInfo *persist.JobInfo) (*pps.JobInfo, error) {
	job := &pps.Job{Id: persistJobInfo.JobId}
	persistJobStatuses, err := a.persistAPIClient.GetJobStatuses(ctx, job)
	if err != nil {
		return nil, err
	}
	persistJobOutput, err := a.persistAPIClient.GetJobOutput(ctx, job)
	if err != nil {
		return nil, err
	}
	jobInfo := &pps.JobInfo{
		Job:   job,
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
			Type:      persistJobStatus.Type,
			Timestamp: persistJobStatus.Timestamp,
			Message:   persistJobStatus.Message,
		}
	}
	if persistJobOutput != nil {
		jobInfo.Output = persistJobOutput.Output
	}
	return jobInfo, nil
}

func (a *apiServer) GetJobLogs(request *pps.GetJobLogsRequest, responseServer pps.JobAPI_GetJobLogsServer) (err error) {
	// TODO(pedge): filter by output stream
	persistJobLogs, err := a.persistAPIClient.GetJobLogs(context.Background(), request.Job)
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

func (a *apiServer) startPersistJobInfo(persistJobInfo *persist.JobInfo) error {
	if _, err := a.persistAPIClient.CreateJobStatus(
		context.Background(),
		&persist.JobStatus{
			JobId: persistJobInfo.JobId,
			Type:  pps.JobStatusType_JOB_STATUS_TYPE_STARTED,
		},
	); err != nil {
		return err
	}
	// TODO(pedge): throttling? worker pool?
	go func() {
		if err := a.runJobInfo(persistJobInfo); err != nil {
			protolog.Errorln(err.Error())
			// TODO(pedge): how to handle the error?
			if _, err = a.persistAPIClient.CreateJobStatus(
				context.Background(),
				&persist.JobStatus{
					JobId:   persistJobInfo.JobId,
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
					JobId: persistJobInfo.JobId,
					Type:  pps.JobStatusType_JOB_STATUS_TYPE_SUCCESS,
				},
			); err != nil {
				protolog.Errorln(err.Error())
			}
		}
	}()
	return nil
}

func (a *apiServer) runJobInfo(persistJobInfo *persist.JobInfo) error {
	switch {
	case persistJobInfo.GetTransform() != nil:
		return a.reallyRunJobInfo(
			strings.Replace(uuid.NewV4().String(), "-", "", -1),
			persistJobInfo.JobId,
			persistJobInfo.GetTransform(),
			persistJobInfo.Input,
			persistJobInfo.OutputParent,
			1,
		)
	case persistJobInfo.GetPipelineName() != "":
		persistPipelineInfo, err := a.persistAPIClient.GetPipelineInfo(
			context.Background(),
			&pps.Pipeline{Name: persistJobInfo.GetPipelineName()},
		)
		if err != nil {
			return err
		}
		if persistPipelineInfo.GetTransform() == nil {
			return fmt.Errorf("pachyderm.pps.server: transform not set on pipeline info %v", persistPipelineInfo)
		}
		return a.reallyRunJobInfo(
			persistPipelineInfo.PipelineName,
			persistJobInfo.JobId,
			persistPipelineInfo.GetTransform(),
			persistJobInfo.Input,
			persistJobInfo.OutputParent,
			1,
		)
	default:
		return fmt.Errorf("pachyderm.pps.server: neither transform or pipeline name set on job info %v", persistJobInfo)
	}
}

func (a *apiServer) reallyRunJobInfo(
	name string,
	jobID string,
	transform *pps.Transform,
	input *pfs.Commit,
	outputParent *pfs.Commit,
	numContainers int,
) error {
	image, err := a.buildOrPull(name, transform)
	if err != nil {
		return err
	}
	// TODO(pedge): branch from output parent
	var containerIDs []string
	defer a.removeContainers(containerIDs)
	for i := 0; i < numContainers; i++ {
		containerID, err := a.containerClient.Create(
			image,
			// TODO(pedge): binds
			container.CreateOptions{
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
	//if transform.Build != "" {
	//image = fmt.Sprintf("ppspipelines/%s", name)
	//if err := a.containerClient.Build(
	//image,
	//transform.Build,
	//// TODO(pedge): this will not work, the path to a dockerfile is not real
	//container.BuildOptions{
	//Dockerfile:   transform.Dockerfile,
	//OutputStream: ioutil.Discard,
	//},
	//); err != nil {
	//return "", err
	//}
	//} else if err := a.containerClient.Pull(
	if err := a.containerClient.Pull(
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

//func getInputBinds(jobInputs []*pps.JobInput) []string {
//var binds []string
//for _, jobInput := range jobInputs {
//if jobInput.GetHostDir() != "" {
//binds = append(binds, getBinds(jobInput.GetHostDir(), filepath.Join("/var/lib/pps/host", jobInput.GetHostDir()), "ro"))
//}
//}
//return binds
//}

//func getOutputBinds(jobOutputs []*pps.JobOutput) []string {
//var binds []string
//for _, jobOutput := range jobOutputs {
//if jobOutput.GetHostDir() != "" {
//binds = append(binds, getBinds(jobOutput.GetHostDir(), filepath.Join("/var/lib/pps/host", jobOutput.GetHostDir()), "rw"))
//}
//}
//return binds
//}

//func getBinds(from string, to string, postfix string) string {
//return fmt.Sprintf("%s:%s:%s", from, to, postfix)
//}
