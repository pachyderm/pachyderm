package server

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"

	"go.pedge.io/google-protobuf"
	"go.pedge.io/pkg/graph"
	"go.pedge.io/pkg/time"
	"go.pedge.io/proto/time"
	"go.pedge.io/protolog"

	"github.com/satori/go.uuid"
	"go.pachyderm.com/pachyderm/src/pfs"
	"go.pachyderm.com/pachyderm/src/pkg/clone"
	"go.pachyderm.com/pachyderm/src/pkg/container"
	"go.pachyderm.com/pachyderm/src/pps"
	"go.pachyderm.com/pachyderm/src/pps/parse"
	"go.pachyderm.com/pachyderm/src/pps/run"
	"go.pachyderm.com/pachyderm/src/pps/store"
	"golang.org/x/net/context"
)

var (
	emptyInstance = &google_protobuf.Empty{}
)

type apiServer struct {
	pfsAPIClient    pfs.ApiClient
	containerClient container.Client
	storeClient     store.Client
	timer           pkgtime.Timer
}

func newAPIServer(pfsAPIClient pfs.ApiClient, containerClient container.Client, storeClient store.Client, timer pkgtime.Timer) *apiServer {
	return &apiServer{pfsAPIClient, containerClient, storeClient, timer}
}

func (a *apiServer) CreatePipelineSource(ctx context.Context, request *pps.CreatePipelineSourceRequest) (*pps.PipelineSource, error) {
	pipelineSource := request.PipelineSource
	if pipelineSource.Id != "" {
		return nil, fmt.Errorf("cannot set id when creating a pipeline source: %+v", pipelineSource)
	}
	pipelineSource.Id = strings.Replace(uuid.NewV4().String(), "-", "", -1)
	if err := a.storeClient.CreatePipelineSource(pipelineSource); err != nil {
		return nil, err
	}
	return pipelineSource, nil
}

func (a *apiServer) GetPipelineSource(ctx context.Context, request *pps.GetPipelineSourceRequest) (*pps.PipelineSource, error) {
	pipelineSource, err := a.storeClient.GetPipelineSource(request.PipelineSourceId)
	if err != nil {
		return nil, err
	}
	return pipelineSource, nil
}

// TODO(pedge): implement
func (a *apiServer) UpdatePipelineSource(ctx context.Context, request *pps.UpdatePipelineSourceRequest) (*pps.PipelineSource, error) {
	return nil, errors.New("not implemented")
}

func (a *apiServer) ArchivePipelineSource(ctx context.Context, request *pps.ArchivePipelineSourceRequest) (*google_protobuf.Empty, error) {
	if err := a.storeClient.ArchivePipelineSource(request.PipelineSourceId); err != nil {
		return nil, err
	}
	return emptyInstance, nil
}

func (a *apiServer) ListPipelineSources(ctx context.Context, request *pps.ListPipelineSourcesRequest) (*pps.PipelineSources, error) {
	pipelineSources, err := a.storeClient.GetAllPipelineSources()
	if err != nil {
		return nil, err
	}
	if request.Tags == nil || len(request.Tags) == 0 {
		return &pps.PipelineSources{
			PipelineSource: pipelineSources,
		}, nil
	}
	var filteredPipelineSources []*pps.PipelineSource
	for _, pipelineSource := range pipelineSources {
		if tagsMatch(request.Tags, pipelineSource.Tags) {
			filteredPipelineSources = append(filteredPipelineSources, pipelineSource)
		}
	}
	return &pps.PipelineSources{
		PipelineSource: pipelineSources,
	}, nil
}

func tagsMatch(expected map[string]string, tags map[string]string) bool {
	for key, value := range expected {
		if tags[key] != value {
			return false
		}
	}
	return true
}

func (a *apiServer) CreateAndGetPipeline(ctx context.Context, request *pps.CreateAndGetPipelineRequest) (*pps.Pipeline, error) {
	pipelineSource, err := a.storeClient.GetPipelineSource(request.PipelineSourceId)
	if err != nil {
		return nil, err
	}
	_, pipeline, err := getDirPathAndPipeline(pipelineSource)
	if err != nil {
		return nil, err
	}
	if err := a.storeClient.CreatePipeline(pipeline); err != nil {
		return nil, err
	}
	return pipeline, nil
}

func (a *apiServer) CreatePipelineRun(ctx context.Context, request *pps.CreatePipelineRunRequest) (*pps.PipelineRun, error) {
	pipelineRun := &pps.PipelineRun{
		Id:         strings.Replace(uuid.NewV4().String(), "-", "", -1),
		PipelineId: request.PipelineId,
	}
	// TODO(pedge): should be transactional with call to CreatePipelineRunStatus
	if err := a.storeClient.CreatePipelineRun(pipelineRun); err != nil {
		return nil, err
	}
	if err := a.storeClient.CreatePipelineRunStatus(pipelineRun.Id, pps.PipelineRunStatusType_PIPELINE_RUN_STATUS_TYPE_CREATED); err != nil {
		return nil, err
	}
	protolog.Info(
		&CreatedPipelineRun{
			PipelineRun: pipelineRun,
		},
	)
	return pipelineRun, nil
}

func (a *apiServer) StartPipelineRun(ctx context.Context, request *pps.StartPipelineRunRequest) (*google_protobuf.Empty, error) {
	runner := run.NewRunner(
		pkggraph.NewGrapher(),
		a.containerClient,
		a.storeClient,
		a.timer,
	)
	if err := runner.Start(request.PipelineRunId); err != nil {
		return nil, err
	}
	return emptyInstance, nil
}

// TODO(pedge): implement
func (a *apiServer) ListPipelineRuns(ctx context.Context, request *pps.ListPipelineRunsRequest) (*pps.PipelineRuns, error) {
	return nil, errors.New("not implemented")
}

func (a *apiServer) GetPipelineRunStatus(ctx context.Context, request *pps.GetPipelineRunStatusRequest) (*pps.PipelineRunStatuses, error) {
	pipelineRunStatuses, err := a.storeClient.GetAllPipelineRunStatuses(request.PipelineRunId)
	if err != nil {
		return nil, err
	}
	if !request.All {
		pipelineRunStatuses = []*pps.PipelineRunStatus{pipelineRunStatuses[0]}
	}
	return &pps.PipelineRunStatuses{
		PipelineRunStatus: pipelineRunStatuses,
	}, nil
}

func (a *apiServer) GetPipelineRunLogs(ctx context.Context, getRunLogsRequest *pps.GetPipelineRunLogsRequest) (*pps.PipelineRunLogs, error) {
	pipelineRunLogs, err := a.storeClient.GetPipelineRunLogs(getRunLogsRequest.PipelineRunId)
	if err != nil {
		return nil, err
	}
	filteredPipelineRunLogs := pipelineRunLogs
	if getRunLogsRequest.Node != "" {
		filteredPipelineRunLogs = make([]*pps.PipelineRunLog, 0)
		for _, pipelineRunLog := range pipelineRunLogs {
			if pipelineRunLog.Node == getRunLogsRequest.Node {
				filteredPipelineRunLogs = append(filteredPipelineRunLogs, pipelineRunLog)
			}
		}
	}
	sort.Sort(sortByTimestamp(filteredPipelineRunLogs))
	return &pps.PipelineRunLogs{
		PipelineRunLog: filteredPipelineRunLogs,
	}, nil
}

func getDirPathAndPipeline(pipelineSource *pps.PipelineSource) (string, *pps.Pipeline, error) {
	if pipelineSource.GetGithubPipelineSource() != nil {
		dirPath, err := githubClone(pipelineSource.GetGithubPipelineSource())
		if err != nil {
			return "", nil, err
		}
		dirPath = filepath.Join(dirPath, pipelineSource.GetGithubPipelineSource().ContextDir)
		pipeline, err := parse.NewParser().ParsePipeline(dirPath)
		if err != nil {
			return "", nil, err
		}
		pipeline.PipelineSourceId = pipelineSource.Id
		pipeline.Id = strings.Replace(uuid.NewV4().String(), "-", "", -1)
		return dirPath, pipeline, nil
	}
	return "", nil, fmt.Errorf("must specify pipeline source")
}

func githubClone(githubPipelineSource *pps.GithubPipelineSource) (string, error) {
	dirPath, err := makeTempDir()
	if err != nil {
		return "", err
	}
	if err := clone.GithubClone(
		dirPath,
		githubPipelineSource.User,
		githubPipelineSource.Repository,
		githubPipelineSource.Branch,
		githubPipelineSource.CommitId,
		githubPipelineSource.AccessToken,
	); err != nil {
		return "", err
	}
	return dirPath, nil
}

func makeTempDir() (string, error) {
	return ioutil.TempDir("", "pachyderm")
}

type sortByTimestamp []*pps.PipelineRunLog

func (s sortByTimestamp) Len() int          { return len(s) }
func (s sortByTimestamp) Swap(i int, j int) { s[i], s[j] = s[j], s[i] }
func (s sortByTimestamp) Less(i int, j int) bool {
	return prototime.TimestampLess(s[i].Timestamp, s[j].Timestamp)
}
