package testing

import (
	"testing"

	"golang.org/x/net/context"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pkg/require"
	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/pachyderm/pachyderm/src/pps/jobserver/run"
)

func TestSimple(t *testing.T) {
	RunTest(t, testSimple)
}

func testSimple(t *testing.T, pfsAPIClient pfs.APIClient, jobAPIClient pps.JobAPIClient, pipelineAPIClient pps.PipelineAPIClient) {
	inputRepoName := "test"
	outputRepoName := "test-output"
	filePathToContent := map[string][]byte{
		"/1": []byte("test1"),
		"/2": []byte("test2"),
		"/3": []byte("test3"),
	}
	transform := &pps.Transform{
		Image: "ubuntu:14.04",
		Cmd: []string{
			fmt.Sprintf("for file in %s/*; do cp ${file} %s/${file}.output; done", jobserverrun.InputMountDir, jobserverRun.OutputMountDir),
		},
	}
	inputCommit, outputParentCommit := setupPFS(t, pfsAPIClient, inputRepoName, outputRepoName, filePathToContent)
	job := createJob(t, jobAPIClient, transform, inputCommit, outputParentCommit)
	waitForJob(t, jobAPIClient, job, 20, true)
}

// TODO: handle directories in filePathToContent
func setupPFS(t *testing.T, pfsAPIClient pfs.APIClient, inputRepoName string, outputRepoName string, filePathToContent map[string]string) (*pfs.Commit, *pfs.Commit) {
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
				Repo: repo,
				Id: pfs.InitialCommitID,
			},
		},
	)
	require.NoError(t, err)
	for filePath, content := range filePathToContent {
		apiPutFileClient, err := pfsAPIClient.PutFile(
			context.Background()
		)
		require.NoError(t, err)
		err = apiPutFileClient.Send(
			&pfs.PutFileRequest {
				File: &pfs.File{
					Commit: commit,
					Path: filePath,
				},
				FileType: pfs.FileType_FILE_TYPE_REGULAR,
				Value: context,
			},
		)
		require.NoError(t, err)
		_, err = apiPutFileClient.CloseAndRecv()
		require.NoError(t, err)
	}
	_, err = pfsAPIClient.FinishCommit(
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
			Input: inputCommit,
			OutputParent: outputParentCommit,
		},
	)
	require.NoError(t, err)
	return job
}

func waitForJob(t *testing.T, jobAPIClient pps.JobAPIClient, job *pps.Job, timeoutSec int, expectSuccess bool) {
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
			if expectSuccess {
				t.Fatalf("job %s had error", job.Id)
			}
			return
		case pps.JobStatusType_JOB_STATUS_TYPE_SUCCESS:
			if !expectSuccess {
				t.Fatalf("job %s did not have error", job.Id)
			}
			return
		}
	}
	t.Fatalf("job %s did not finish in %d seconds", timeoutSec)
}
