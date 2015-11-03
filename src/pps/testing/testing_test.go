package testing

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"go.pedge.io/proto/stream"
	"go.pedge.io/protolog"

	"golang.org/x/net/context"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pkg/require"
	"github.com/pachyderm/pachyderm/src/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/pachyderm/pachyderm/src/pps/jobserver/run"
)

const (
	defaultJobTimeoutSec = 20
)

func TestSimple(t *testing.T) {
	runSimpleJobTest(
		t,
		&simpleJobTest{
			inputFilePathToContent: map[string][]byte{
				"/1": []byte("test1"),
				"/2": []byte("test2"),
				"/3": []byte("test3"),
			},
			outputFilePathToContent: map[string][]byte{
				"/1.output": []byte("test1"),
				"/2.output": []byte("test2"),
				"/3.output": []byte("test3"),
			},
			transform: &pps.Transform{
				Image: "ubuntu:14.04",
				Cmd: []string{
					fmt.Sprintf("for file in %s/*; do cp \"${file}\" %s/$(basename \"${file}\").output; done", jobserverrun.InputMountDir, jobserverrun.OutputMountDir),
				},
			},
		},
	)
}

func runSimpleJobTest(t *testing.T, simpleJobTest *simpleJobTest) {
	RunTest(t, simpleJobTest.run)
}

type simpleJobTest struct {
	inputRepoName           string
	outputRepoName          string
	inputFilePathToContent  map[string][]byte
	outputFilePathToContent map[string][]byte
	transform               *pps.Transform
	jobTimeoutSec           int
	expectError             bool
}

func (s *simpleJobTest) run(
	t *testing.T,
	pfsAPIClient pfs.APIClient,
	jobAPIClient pps.JobAPIClient,
	pipelineAPIClient pps.PipelineAPIClient,
) {
	inputRepoName := s.inputRepoName
	if inputRepoName == "" {
		inputRepoName = uuid.NewWithoutDashes()
	}
	outputRepoName := s.outputRepoName
	if outputRepoName == "" {
		outputRepoName = uuid.NewWithoutDashes()
	}
	jobTimeoutSec := s.jobTimeoutSec
	if jobTimeoutSec == 0 {
		jobTimeoutSec = defaultJobTimeoutSec
	}
	inputCommit, outputParentCommit := setupPFS(t, pfsAPIClient, inputRepoName, outputRepoName, s.inputFilePathToContent)
	job := createJob(t, jobAPIClient, s.transform, inputCommit, outputParentCommit)
	jobInfo := waitForJob(t, jobAPIClient, job, jobTimeoutSec, s.expectError)
	checkPFSOutput(t, pfsAPIClient, jobInfo.Output, s.outputFilePathToContent)
}

// TODO: handle directories in filePathToContent
func setupPFS(t *testing.T, pfsAPIClient pfs.APIClient, inputRepoName string, outputRepoName string, filePathToContent map[string][]byte) (*pfs.Commit, *pfs.Commit) {
	inputRepo := &pfs.Repo{
		Name: inputRepoName,
	}
	_, err := pfsAPIClient.CreateRepo(
		context.Background(),
		&pfs.CreateRepoRequest{
			Repo: inputRepo,
		},
	)
	require.NoError(t, err)
	outputRepo := &pfs.Repo{
		Name: outputRepoName,
	}
	_, err = pfsAPIClient.CreateRepo(
		context.Background(),
		&pfs.CreateRepoRequest{
			Repo: outputRepo,
		},
	)
	require.NoError(t, err)
	commit, err := pfsAPIClient.StartCommit(
		context.Background(),
		&pfs.StartCommitRequest{
			Commit: &pfs.Commit{
				Repo: inputRepo,
				Id:   pfs.InitialCommitID,
			},
		},
	)
	require.NoError(t, err)
	for filePath, content := range filePathToContent {
		apiPutFileClient, err := pfsAPIClient.PutFile(
			context.Background(),
		)
		require.NoError(t, err)
		err = apiPutFileClient.Send(
			&pfs.PutFileRequest{
				File: &pfs.File{
					Commit: commit,
					Path:   filePath,
				},
				FileType: pfs.FileType_FILE_TYPE_REGULAR,
				Value:    content,
			},
		)
		require.NoError(t, err)
		_, err = apiPutFileClient.CloseAndRecv()
		require.NoError(t, err)
	}
	_, err = pfsAPIClient.FinishCommit(
		context.Background(),
		&pfs.FinishCommitRequest{
			Commit: commit,
		},
	)
	require.NoError(t, err)
	return commit, &pfs.Commit{Repo: outputRepo, Id: pfs.InitialCommitID}
}

func createJob(t *testing.T, jobAPIClient pps.JobAPIClient, transform *pps.Transform, inputCommit *pfs.Commit, outputParentCommit *pfs.Commit) *pps.Job {
	job, err := jobAPIClient.CreateJob(
		context.Background(),
		&pps.CreateJobRequest{
			Spec: &pps.CreateJobRequest_Transform{
				Transform: transform,
			},
			Input:        inputCommit,
			OutputParent: outputParentCommit,
		},
	)
	require.NoError(t, err)
	return job
}

func waitForJob(t *testing.T, jobAPIClient pps.JobAPIClient, job *pps.Job, timeoutSec int, expectError bool) *pps.JobInfo {
	for i := 0; i < timeoutSec; i++ {
		time.Sleep(1 * time.Second)
		jobInfo, err := jobAPIClient.InspectJob(
			context.Background(),
			&pps.InspectJobRequest{
				Job: job,
			},
		)
		require.NoError(t, err)
		if len(jobInfo.JobStatus) == 0 {
			continue
		}
		jobStatus := jobInfo.JobStatus[0]
		protolog.Infof("status of job %s at %d seconds: %v", job.Id, i+1, jobInfo.JobStatus)
		switch jobStatus.Type {
		case pps.JobStatusType_JOB_STATUS_TYPE_ERROR:
			if !expectError {
				t.Fatalf("job %s had error", job.Id)
			}
			return jobInfo
		case pps.JobStatusType_JOB_STATUS_TYPE_SUCCESS:
			if expectError {
				t.Fatalf("job %s did not have error", job.Id)
			}
			return jobInfo
		}
	}
	t.Fatalf("job %s did not finish in %d seconds", job.Id, timeoutSec)
	return nil
}

func checkPFSOutput(t *testing.T, pfsAPIClient pfs.APIClient, outputCommit *pfs.Commit, filePathToContent map[string][]byte) {
	for filePath, content := range filePathToContent {
		getContent, err := getPFSContent(pfsAPIClient, outputCommit, filePath)
		require.NoError(t, err)
		require.Equal(t, content, getContent)
	}
}

func getPFSContent(pfsAPIClient pfs.APIClient, commit *pfs.Commit, filePath string) ([]byte, error) {
	apiGetFileClient, err = pfsAPIClient.GetFile(
		context.Background(),
		&pfs.GetFileRequest{
			File: &pfs.File{
				Commit: commit,
				Path:   filePath,
			},
		},
	)
	if err != nil {
		return nil, err
	}
	buffer := bytes.NewBuffer(nil)
	if err := protostream.WriteFromStreamingBytesClient(apiGetFileClient, buffer); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
