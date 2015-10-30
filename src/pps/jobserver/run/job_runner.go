package jobserverrun

import (
	"fmt"
	"strings"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pkg/container"
	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/pachyderm/pachyderm/src/pps/persist"
	"github.com/satori/go.uuid"
	"go.pedge.io/protolog"
	"golang.org/x/net/context"
)

type jobRunner struct {
	persistAPIClient persist.APIClient
	containerClient  container.Client
	pfsMountDir      string
}

func newJobRunner(
	persistAPIClient persist.APIClient,
	containerClient container.Client,
	pfsMountDir string,
) JobRunner {
	return &jobRunner{
		persistAPIClient,
		containerClient,
		pfsMountDir,
	}
}

func (j *jobRunner) Start(persistJobInfo *persist.JobInfo) error {
	if _, err := j.persistAPIClient.CreateJobStatus(
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
		if err := j.runJobInfo(persistJobInfo); err != nil {
			protolog.Errorln(err.Error())
			// TODO(pedge): how to handle the error?
			if _, err = j.persistAPIClient.CreateJobStatus(
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
			if _, err = j.persistAPIClient.CreateJobStatus(
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

func (j *jobRunner) runJobInfo(persistJobInfo *persist.JobInfo) error {
	switch {
	case persistJobInfo.GetTransform() != nil:
		return j.reallyRunJobInfo(
			strings.Replace(uuid.NewV4().String(), "-", "", -1),
			persistJobInfo.JobId,
			persistJobInfo.GetTransform(),
			persistJobInfo.Input,
			persistJobInfo.OutputParent,
			1,
		)
	case persistJobInfo.GetPipelineName() != "":
		persistPipelineInfo, err := j.persistAPIClient.GetPipelineInfo(
			context.Background(),
			&pps.Pipeline{Name: persistJobInfo.GetPipelineName()},
		)
		if err != nil {
			return err
		}
		if persistPipelineInfo.GetTransform() == nil {
			return fmt.Errorf("pachyderm.pps.server: transform not set on pipeline info %v", persistPipelineInfo)
		}
		return j.reallyRunJobInfo(
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

func (j *jobRunner) reallyRunJobInfo(
	name string,
	jobID string,
	transform *pps.Transform,
	input *pfs.Commit,
	outputParent *pfs.Commit,
	numContainers int,
) error {
	image, err := j.buildOrPull(name, transform)
	if err != nil {
		return err
	}
	// TODO(pedge): branch from output parent
	var containerIDs []string
	defer j.removeContainers(containerIDs)
	for i := 0; i < numContainers; i++ {
		containerID, err := j.containerClient.Create(
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
		if err := j.containerClient.Start(
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
		go j.writeContainerLogs(containerID, jobID, errC)
	}
	for _, containerID := range containerIDs {
		if err := j.containerClient.Wait(containerID, container.WaitOptions{}); err != nil {
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
func (j *jobRunner) buildOrPull(name string, transform *pps.Transform) (string, error) {
	image := transform.Image
	//if transform.Build != "" {
	//image = fmt.Sprintf("ppspipelines/%s", name)
	//if err := j.containerClient.Build(
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
	//} else if err := j.containerClient.Pull(
	if err := j.containerClient.Pull(
		transform.Image,
		container.PullOptions{},
	); err != nil {
		return "", err
	}
	return image, nil
}

func (j *jobRunner) removeContainers(containerIDs []string) {
	for _, containerID := range containerIDs {
		_ = j.containerClient.Kill(containerID, container.KillOptions{})
		//_ = j.containerClient.Remove(containerID, container.RemoveOptions{})
	}
}

func (j *jobRunner) writeContainerLogs(containerID string, jobID string, errC chan error) {
	errC <- j.containerClient.Logs(
		containerID,
		container.LogsOptions{
			Stdout: newJobLogWriter(
				jobID,
				pps.OutputStream_OUTPUT_STREAM_STDOUT,
				j.persistAPIClient,
			),
			Stderr: newJobLogWriter(
				jobID,
				pps.OutputStream_OUTPUT_STREAM_STDERR,
				j.persistAPIClient,
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
