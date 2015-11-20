package testing

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"go.pedge.io/proto/stream"

	"golang.org/x/net/context"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pkg/require"
	"github.com/pachyderm/pachyderm/src/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/pps"
)

const (
	defaultJobTimeoutSec = 30
)

var (
	baseSimpleTest = &simpleTest{
		inputFilePathToContent: map[string][]byte{
			"1": []byte("test1"),
			"2": []byte("test2"),
			"3": []byte("test3"),
		},
		outputFilePathToContent: map[string][]byte{
			"1.output": []byte("test1"),
			"2.output": []byte("test2"),
			"3.output": []byte("test3"),
		},
		transform: &pps.Transform{
			Image: "ubuntu:14.04",
			Cmd: []string{
				fmt.Sprintf("for file in /pfs/*; do cp \"${file}\" /pfs/$(basename \"${file}\").output; done"),
			},
		},
	}
)

func TestSimpleJob(t *testing.T) {
	t.Skip()
	runSimpleJobTest(t, baseSimpleTest.copy())
}

func TestSimplePipeline(t *testing.T) {
	t.Skip()
	runSimplePipelineTest(t, baseSimpleTest.copy())
}

func runSimpleJobTest(t *testing.T, simpleTest *simpleTest) {
	RunTest(t, simpleTest.runJob)
}

func runSimplePipelineTest(t *testing.T, simpleTest *simpleTest) {
	RunTest(t, simpleTest.runPipeline)
}

type simpleTest struct {
	inputRepoName           string
	outputRepoName          string
	inputFilePathToContent  map[string][]byte
	outputFilePathToContent map[string][]byte
	transform               *pps.Transform
	pipelineName            string
	jobTimeoutSec           int
	expectError             bool
}

func (s *simpleTest) copy() *simpleTest {
	c := &simpleTest{
		inputRepoName:           s.inputRepoName,
		outputRepoName:          s.outputRepoName,
		inputFilePathToContent:  make(map[string][]byte),
		outputFilePathToContent: make(map[string][]byte),
		pipelineName:            s.pipelineName,
		jobTimeoutSec:           s.jobTimeoutSec,
		expectError:             s.expectError,
	}
	for key, value := range s.inputFilePathToContent {
		buf := make([]byte, len(value))
		copy(buf, value)
		c.inputFilePathToContent[key] = buf
	}
	for key, value := range s.outputFilePathToContent {
		buf := make([]byte, len(value))
		copy(buf, value)
		c.outputFilePathToContent[key] = buf
	}
	if s.transform != nil {
		c.transform = &pps.Transform{
			Image: s.transform.Image,
			Cmd:   make([]string, len(s.transform.Cmd)),
		}
		for i, value := range s.transform.Cmd {
			c.transform.Cmd[i] = value
		}
	}
	return c
}

func (s *simpleTest) setup() {
	if s.inputRepoName == "" {
		s.inputRepoName = uuid.NewWithoutDashes()
	}
	if s.outputRepoName == "" {
		s.outputRepoName = uuid.NewWithoutDashes()
	}
	if s.pipelineName == "" {
		s.pipelineName = uuid.NewWithoutDashes()
	}
	if s.jobTimeoutSec == 0 {
		s.jobTimeoutSec = defaultJobTimeoutSec
	}
}

func (s *simpleTest) runJob(
	t *testing.T,
	pfsAPIClient pfs.APIClient,
	jobAPIClient pps.JobAPIClient,
	pipelineAPIClient pps.PipelineAPIClient,
) {
	s.setup()
	inputRepo := setupPFSRepo(t, pfsAPIClient, s.inputRepoName)
	outputRepo := setupPFSRepo(t, pfsAPIClient, s.outputRepoName)
	inputCommit := setupPFSInputCommit(t, pfsAPIClient, inputRepo, s.inputFilePathToContent)
	outputParentCommit := setupPFSOutputParentCommit(t, pfsAPIClient, outputRepo)
	job := createJob(t, jobAPIClient, s.transform, inputCommit, outputParentCommit)
	jobInfo := waitForJob(t, jobAPIClient, job, s.jobTimeoutSec, s.expectError)
	checkPFSOutput(t, pfsAPIClient, jobInfo.OutputCommit, s.outputFilePathToContent)
}

func (s *simpleTest) runPipeline(
	t *testing.T,
	pfsAPIClient pfs.APIClient,
	jobAPIClient pps.JobAPIClient,
	pipelineAPIClient pps.PipelineAPIClient,
) {
	s.setup()
	inputRepo := setupPFSRepo(t, pfsAPIClient, s.inputRepoName)
	outputRepo := setupPFSRepo(t, pfsAPIClient, s.outputRepoName)
	pipeline := createPipeline(t, pipelineAPIClient, s.pipelineName, s.transform, inputRepo, outputRepo)
	setupPFSInputCommit(t, pfsAPIClient, inputRepo, s.inputFilePathToContent)
	// TODO: all the waits heh
	time.Sleep(5 * time.Second)
	job := getJobForPipeline(t, jobAPIClient, pipeline)
	jobInfo := waitForJob(t, jobAPIClient, job, s.jobTimeoutSec, s.expectError)
	checkPFSOutput(t, pfsAPIClient, jobInfo.OutputCommit, s.outputFilePathToContent)
}

func setupPFSRepo(t *testing.T, pfsAPIClient pfs.APIClient, repoName string) *pfs.Repo {
	repo := &pfs.Repo{
		Name: repoName,
	}
	_, err := pfsAPIClient.CreateRepo(
		context.Background(),
		&pfs.CreateRepoRequest{
			Repo: repo,
		},
	)
	require.NoError(t, err)
	return repo
}

// TODO: handle directories in filePathToContent
func setupPFSInputCommit(t *testing.T, pfsAPIClient pfs.APIClient, repo *pfs.Repo, filePathToContent map[string][]byte) *pfs.Commit {
	commit, err := pfsAPIClient.StartCommit(
		context.Background(),
		&pfs.StartCommitRequest{
			Parent: &pfs.Commit{
				Repo: repo,
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
		_, _ = apiPutFileClient.CloseAndRecv()
	}
	_, err = pfsAPIClient.FinishCommit(
		context.Background(),
		&pfs.FinishCommitRequest{
			Commit: commit,
		},
	)
	require.NoError(t, err)
	return commit
}

func setupPFSOutputParentCommit(t *testing.T, pfsAPIClient pfs.APIClient, repo *pfs.Repo) *pfs.Commit {
	return &pfs.Commit{Repo: repo, Id: pfs.InitialCommitID}
}

func createJob(t *testing.T, jobAPIClient pps.JobAPIClient, transform *pps.Transform, inputCommit *pfs.Commit, outputParentCommit *pfs.Commit) *pps.Job {
	job, err := jobAPIClient.CreateJob(
		context.Background(),
		&pps.CreateJobRequest{
			Spec: &pps.CreateJobRequest_Transform{
				Transform: transform,
			},
			InputCommit:  []*pfs.Commit{inputCommit},
			OutputParent: outputParentCommit,
		},
	)
	require.NoError(t, err)
	return job
}

func waitForJob(t *testing.T, jobAPIClient pps.JobAPIClient, job *pps.Job, timeoutSec int, expectError bool) *pps.JobInfo {
	// TODO
	return nil
}

func createPipeline(t *testing.T, pipelineAPIClient pps.PipelineAPIClient, pipelineName string, transform *pps.Transform, inputRepo *pfs.Repo, outputRepo *pfs.Repo) *pps.Pipeline {
	pipeline := &pps.Pipeline{
		Name: pipelineName,
	}
	_, err := pipelineAPIClient.CreatePipeline(
		context.Background(),
		&pps.CreatePipelineRequest{
			Pipeline:   pipeline,
			Transform:  transform,
			InputRepo:  []*pfs.Repo{inputRepo},
			OutputRepo: outputRepo,
		},
	)
	require.NoError(t, err)
	return pipeline
}

func getJobForPipeline(t *testing.T, jobAPIClient pps.JobAPIClient, pipeline *pps.Pipeline) *pps.Job {
	jobInfos, err := jobAPIClient.ListJob(
		context.Background(),
		&pps.ListJobRequest{
			Pipeline: pipeline,
		},
	)
	require.NoError(t, err)
	require.NotNil(t, jobInfos)
	require.Equal(t, 1, len(jobInfos.JobInfo))
	return jobInfos.JobInfo[0].Job
}

func checkPFSOutput(t *testing.T, pfsAPIClient pfs.APIClient, outputCommit *pfs.Commit, filePathToContent map[string][]byte) {
	for filePath, content := range filePathToContent {
		getContent, err := getPFSContent(pfsAPIClient, outputCommit, filePath)
		require.NoError(t, err)
		require.Equal(t, content, getContent)
	}
}

func getPFSContent(pfsAPIClient pfs.APIClient, commit *pfs.Commit, filePath string) ([]byte, error) {
	apiGetFileClient, err := pfsAPIClient.GetFile(
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
