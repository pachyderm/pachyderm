package server

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	pfsclient "github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	ppsclient "github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pfs/fuse"
	"github.com/pachyderm/pachyderm/src/server/pkg/metrics"
	ppsserver "github.com/pachyderm/pachyderm/src/server/pps"
	"github.com/pachyderm/pachyderm/src/server/pps/persist"

	"go.pedge.io/lion/proto"
	"go.pedge.io/pb/go/google/protobuf"
	"go.pedge.io/proto/rpclog"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"k8s.io/kubernetes/pkg/api"
	kube_api "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/apis/extensions"
	kube "k8s.io/kubernetes/pkg/client/unversioned"
	kube_labels "k8s.io/kubernetes/pkg/labels"
)

var (
	trueVal = true
	suite   = "pachyderm"
)

type apiServer struct {
	protorpclog.Logger
	hasher               *ppsserver.Hasher
	address              string
	pfsAPIClient         pfsclient.APIClient
	pfsClientOnce        sync.Once
	persistAPIClient     persist.APIClient
	persistClientOnce    sync.Once
	kubeClient           *kube.Client
	cancelFuncs          map[ppsclient.Pipeline]func()
	cancelFuncsLock      sync.Mutex
	shardCancelFuncs     map[uint64]func()
	shardCancelFuncsLock sync.Mutex
	version              int64
	// versionLock protects the version field.
	// versionLock must be held BEFORE reading from version and UNTIL all
	// requests using version have returned
	versionLock sync.RWMutex
}

func (a *apiServer) CreateJob(ctx context.Context, request *ppsclient.CreateJobRequest) (response *ppsclient.Job, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	defer func() {
		if retErr == nil {
			metrics.AddJobs(1)
		}
	}()

	persistClient, err := a.getPersistClient()
	if err != nil {
		return nil, err
	}

	if request.Shards == 0 {
		nodeList, err := a.kubeClient.Nodes().List(kube_api.ListOptions{})
		if err != nil {
			return nil, fmt.Errorf("pachyderm.ppsclient.jobserver: shards set to zero and unable to retrieve node list from k8s")
		}

		if len(nodeList.Items) == 0 {
			return nil, fmt.Errorf("pachyderm.ppsclient.jobserver: no k8s nodes found")
		}

		request.Shards = uint64(len(nodeList.Items))
	}
	repoSet := make(map[string]bool)
	for _, input := range request.Inputs {
		repoSet[input.Commit.Repo.Name] = true
	}
	if len(repoSet) < len(request.Inputs) {
		return nil, fmt.Errorf("pachyderm.ppsclient.jobserver: duplicate repo in job")
	}

	var parentJobInfo *persist.JobInfo
	if request.ParentJob != nil {
		inspectJobRequest := &ppsclient.InspectJobRequest{Job: request.ParentJob}
		parentJobInfo, err = persistClient.InspectJob(ctx, inspectJobRequest)
		if err != nil {
			return nil, err
		}
	}

	pfsAPIClient, err := a.getPfsClient()
	if err != nil {
		return nil, err
	}

	jobID := getJobID(request)

	startCommitRequest := &pfsclient.StartCommitRequest{}

	// If JobInfo.Pipeline is set, use the pipeline repo
	if request.Pipeline != nil {
		startCommitRequest.Repo = ppsserver.PipelineRepo(&ppsclient.Pipeline{Name: request.Pipeline.Name})
		if parentJobInfo != nil && parentJobInfo.OutputCommit.Repo.Name != startCommitRequest.Repo.Name {
			return nil, fmt.Errorf("Parent job was not part of the same pipeline; this is likely a bug")
		}
	} else {
		// If parent is set, use the parent's repo
		if parentJobInfo != nil {
			startCommitRequest.Repo = parentJobInfo.OutputCommit.Repo
		} else {
			// Otherwise, create a repo for this job
			startCommitRequest.Repo = ppsserver.JobRepo(&ppsclient.Job{
				ID: jobID,
			})
			if _, err := pfsAPIClient.CreateRepo(ctx, &pfsclient.CreateRepoRequest{Repo: startCommitRequest.Repo}); err != nil {
				return nil, err
			}
		}
	}

	// If parent is set...
	if parentJobInfo != nil {
		reduce := false
		for _, jobInput := range request.Inputs {
			if jobInput.Reduce {
				reduce = true
			}
		}
		// ...and if the job is not a reduce job, the parent's output commit
		// should be this commit's parent.
		// Otherwise this commit should have no parent.
		if !reduce {
			startCommitRequest.ParentID = parentJobInfo.OutputCommit.ID
		}
	}

	commit, err := pfsAPIClient.StartCommit(ctx, startCommitRequest)
	if err != nil {
		return nil, err
	}

	// TODO validate job to make sure input commits and output repo exist
	persistJobInfo := &persist.JobInfo{
		JobID:        jobID,
		Shards:       request.Shards,
		Transform:    request.Transform,
		Inputs:       request.Inputs,
		ParentJob:    request.ParentJob,
		OutputCommit: commit,
	}
	if request.Pipeline != nil {
		persistJobInfo.PipelineName = request.Pipeline.Name
	}
	if a.kubeClient == nil {
		return nil, fmt.Errorf("pachyderm.ppsclient.jobserver: no job backend")
	}
	_, err = persistClient.CreateJobInfo(ctx, persistJobInfo)
	if err != nil && !isConflictErr(err) {
		return nil, err
	}

	if err == nil {
		// we only create a kube job if the job did not already exist
		if _, err := a.kubeClient.Jobs(api.NamespaceDefault).Create(job(persistJobInfo)); err != nil {
			return nil, err
		}
	}

	defer func() {
		if retErr != nil {
			if _, err := persistClient.CreateJobState(ctx, &persist.JobState{
				JobID: persistJobInfo.JobID,
				State: ppsclient.JobState_JOB_STATE_FAILURE,
			}); err != nil {
				protolion.Errorf("error from CreateJobState %s", err.Error())
			}
		}
	}()

	return &ppsclient.Job{
		ID: jobID,
	}, nil
}

// isConflictErr returns true if the error is non-nil and the query failed
// due to a duplicate primary key.
func isConflictErr(err error) bool {
	if err == nil {
		return false
	}

	return strings.Contains(err.Error(), "Duplicate primary key")
}

func getJobID(req *ppsclient.CreateJobRequest) string {
	// If the job belongs to a pipeline, we want to make sure that the same
	// job does now run twice.  We ensure that by generating the job id by
	// hashing the pipeline name and input commits.  That way, two same jobs
	// will have the sam job IDs, therefore won't be created in the database
	// twice.
	if req.Pipeline != nil {
		s := req.Pipeline.Name
		for _, input := range req.Inputs {
			s += "/" + input.Commit.ID
		}

		hash := md5.Sum([]byte(s))
		return fmt.Sprintf("%x", hash)
	}

	return uuid.NewWithoutDashes()
}

func (a *apiServer) InspectJob(ctx context.Context, request *ppsclient.InspectJobRequest) (response *ppsclient.JobInfo, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	persistClient, err := a.getPersistClient()
	if err != nil {
		return nil, err
	}

	persistJobInfo, err := persistClient.InspectJob(ctx, request)
	if err != nil {
		return nil, err
	}
	return newJobInfo(persistJobInfo)
}

func (a *apiServer) ListJob(ctx context.Context, request *ppsclient.ListJobRequest) (response *ppsclient.JobInfos, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	persistClient, err := a.getPersistClient()
	if err != nil {
		return nil, err
	}

	persistJobInfos, err := persistClient.ListJobInfos(ctx, request)
	if err != nil {
		return nil, err
	}
	jobInfos := make([]*ppsclient.JobInfo, len(persistJobInfos.JobInfo))
	for i, persistJobInfo := range persistJobInfos.JobInfo {
		jobInfo, err := newJobInfo(persistJobInfo)
		if err != nil {
			return nil, err
		}
		jobInfos[i] = jobInfo
	}
	return &ppsclient.JobInfos{
		JobInfo: jobInfos,
	}, nil
}

func (a *apiServer) GetLogs(request *ppsclient.GetLogsRequest, apiGetLogsServer ppsclient.API_GetLogsServer) (retErr error) {
	defer func(start time.Time) { a.Log(request, nil, retErr, time.Since(start)) }(time.Now())
	podList, err := a.kubeClient.Pods(api.NamespaceDefault).List(kube_api.ListOptions{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "ListOptions",
			APIVersion: "v1",
		},
		LabelSelector: kube_labels.SelectorFromSet(labels(request.Job.ID)),
	})
	if err != nil {
		return err
	}
	// sort the pods to make sure that the indexes are stable
	sort.Sort(podSlice(podList.Items))
	logs := make([][]byte, len(podList.Items))
	var wg sync.WaitGroup
	errCh := make(chan error, 1)
	for i, pod := range podList.Items {
		i := i
		pod := pod
		wg.Add(1)
		go func() {
			defer wg.Done()
			result := a.kubeClient.Pods(api.NamespaceDefault).GetLogs(
				pod.ObjectMeta.Name, &kube_api.PodLogOptions{}).Do()
			value, err := result.Raw()
			if err != nil {
				select {
				default:
				case errCh <- err:
				}
			}
			var buffer bytes.Buffer
			scanner := bufio.NewScanner(bytes.NewBuffer(value))
			for scanner.Scan() {
				fmt.Fprintf(&buffer, "%d | %s\n", i, scanner.Text())
			}
			logs[i] = buffer.Bytes()
		}()
	}
	wg.Wait()
	select {
	default:
	case err := <-errCh:
		return err
	}
	for _, log := range logs {
		if err := apiGetLogsServer.Send(&google_protobuf.BytesValue{Value: log}); err != nil {
			return err
		}
	}
	return nil
}

func (a *apiServer) StartJob(ctx context.Context, request *ppsserver.StartJobRequest) (response *ppsserver.StartJobResponse, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	persistClient, err := a.getPersistClient()
	if err != nil {
		return nil, err
	}

	jobInfo, err := persistClient.StartShard(ctx, request.Job)
	if err != nil {
		return nil, err
	}

	if jobInfo.ShardsStarted > jobInfo.Shards {
		return nil, fmt.Errorf("job %s already has %d shards", request.Job.ID, jobInfo.Shards)
	}

	if jobInfo.Transform == nil {
		return nil, fmt.Errorf("jobInfo.Transform should not be nil (this is likely a bug)")
	}

	var parentJobInfo *persist.JobInfo
	if jobInfo.ParentJob != nil {
		inspectJobRequest := &ppsclient.InspectJobRequest{Job: jobInfo.ParentJob}
		parentJobInfo, err = persistClient.InspectJob(ctx, inspectJobRequest)
		if err != nil {
			return nil, err
		}
	}
	repoToFromCommit := make(map[string]*pfsclient.Commit)
	if parentJobInfo != nil {
		for _, jobInput := range parentJobInfo.Inputs {
			if !jobInput.Reduce {
				// input isn't being reduced, do it incrementally
				repoToFromCommit[jobInput.Commit.Repo.Name] = jobInput.Commit
			}
		}
	}

	if jobInfo.OutputCommit == nil {
		return nil, fmt.Errorf("jobInfo.OutputCommit should not be nil (this is likely a bug)")
	}
	var commitMounts []*fuse.CommitMount
	for _, jobInput := range jobInfo.Inputs {
		commitMount := &fuse.CommitMount{
			Commit:     jobInput.Commit,
			FromCommit: repoToFromCommit[jobInput.Commit.Repo.Name],
			Shard: &pfsclient.Shard{
				FileModulus:  1,
				BlockModulus: 1,
			},
		}
		if jobInput.Reduce {
			commitMount.Shard.FileNumber = jobInfo.ShardsStarted - 1
			commitMount.Shard.FileModulus = jobInfo.Shards
		} else {
			commitMount.Shard.BlockNumber = jobInfo.ShardsStarted - 1
			commitMount.Shard.BlockModulus = jobInfo.Shards
		}
		commitMounts = append(commitMounts, commitMount)
	}
	outputCommitMount := &fuse.CommitMount{
		Commit: jobInfo.OutputCommit,
		Alias:  "out",
	}
	commitMounts = append(commitMounts, outputCommitMount)
	return &ppsserver.StartJobResponse{
		Transform:    jobInfo.Transform,
		CommitMounts: commitMounts,
		Index:        jobInfo.ShardsStarted - 1,
	}, nil
}

func (a *apiServer) FinishJob(ctx context.Context, request *ppsserver.FinishJobRequest) (response *google_protobuf.Empty, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	persistClient, err := a.getPersistClient()
	if err != nil {
		return nil, err
	}

	var jobInfo *persist.JobInfo
	if request.Success {
		jobInfo, err = persistClient.SucceedShard(ctx, request.Job)
		if err != nil {
			return nil, err
		}
	} else {
		jobInfo, err = persistClient.FailShard(ctx, request.Job)
		if err != nil {
			return nil, err
		}
	}
	if jobInfo.ShardsSucceeded+jobInfo.ShardsFailed == jobInfo.Shards {
		if jobInfo.OutputCommit == nil {
			return nil, fmt.Errorf("jobInfo.OutputCommit should not be nil (this is likely a bug)")
		}
		pfsAPIClient, err := a.getPfsClient()
		if err != nil {
			return nil, err
		}
		if _, err := pfsAPIClient.FinishCommit(ctx, &pfsclient.FinishCommitRequest{
			Commit: jobInfo.OutputCommit,
		}); err != nil {
			return nil, err
		}
		jobState := ppsclient.JobState_JOB_STATE_FAILURE
		if jobInfo.ShardsSucceeded == jobInfo.Shards {
			jobState = ppsclient.JobState_JOB_STATE_SUCCESS
		}
		if _, err := persistClient.CreateJobState(ctx, &persist.JobState{
			JobID: request.Job.ID,
			State: jobState,
		}); err != nil {
			return nil, err
		}
	}
	return google_protobuf.EmptyInstance, nil
}

func (a *apiServer) CreatePipeline(ctx context.Context, request *ppsclient.CreatePipelineRequest) (response *google_protobuf.Empty, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	defer func() {
		if retErr == nil {
			metrics.AddPipelines(1)
		}
	}()
	pfsAPIClient, err := a.getPfsClient()
	if err != nil {
		return nil, err
	}

	persistClient, err := a.getPersistClient()
	if err != nil {
		return nil, err
	}

	if request.Pipeline == nil {
		return nil, fmt.Errorf("pachyderm.ppsclient.pipelineserver: request.Pipeline cannot be nil")
	}
	repoSet := make(map[string]bool)
	for _, input := range request.Inputs {
		if _, err := pfsAPIClient.InspectRepo(ctx, &pfsclient.InspectRepoRequest{Repo: input.Repo}); err != nil {
			return nil, err
		}
		repoSet[input.Repo.Name] = true
	}
	if len(repoSet) < len(request.Inputs) {
		return nil, fmt.Errorf("pachyderm.ppsclient.pipelineserver: duplicate input repos")
	}
	repo := ppsserver.PipelineRepo(request.Pipeline)
	if _, err := pfsAPIClient.CreateRepo(ctx, &pfsclient.CreateRepoRequest{Repo: repo}); err != nil {
		return nil, err
	}
	persistPipelineInfo := &persist.PipelineInfo{
		PipelineName: request.Pipeline.Name,
		Transform:    request.Transform,
		Shards:       request.Shards,
		Inputs:       request.Inputs,
		OutputRepo:   repo,
		Shard:        a.hasher.HashPipeline(request.Pipeline),
	}
	if _, err := persistClient.CreatePipelineInfo(ctx, persistPipelineInfo); err != nil {
		return nil, err
	}
	return google_protobuf.EmptyInstance, nil
}

func (a *apiServer) InspectPipeline(ctx context.Context, request *ppsclient.InspectPipelineRequest) (response *ppsclient.PipelineInfo, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	persistClient, err := a.getPersistClient()
	if err != nil {
		return nil, err
	}

	persistPipelineInfo, err := persistClient.GetPipelineInfo(ctx, request.Pipeline)
	if err != nil {
		return nil, err
	}
	return newPipelineInfo(persistPipelineInfo), nil
}

func (a *apiServer) ListPipeline(ctx context.Context, request *ppsclient.ListPipelineRequest) (response *ppsclient.PipelineInfos, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	persistClient, err := a.getPersistClient()
	if err != nil {
		return nil, err
	}

	persistPipelineInfos, err := persistClient.ListPipelineInfos(ctx, &persist.ListPipelineInfosRequest{})
	if err != nil {
		return nil, err
	}
	pipelineInfos := make([]*ppsclient.PipelineInfo, len(persistPipelineInfos.PipelineInfo))
	for i, persistPipelineInfo := range persistPipelineInfos.PipelineInfo {
		pipelineInfos[i] = newPipelineInfo(persistPipelineInfo)
	}
	return &ppsclient.PipelineInfos{
		PipelineInfo: pipelineInfos,
	}, nil
}

func (a *apiServer) DeletePipeline(ctx context.Context, request *ppsclient.DeletePipelineRequest) (response *google_protobuf.Empty, err error) {
	persistClient, err := a.getPersistClient()
	if err != nil {
		return nil, err
	}

	if _, err := persistClient.DeletePipelineInfo(ctx, request.Pipeline); err != nil {
		return nil, err
	}
	a.cancelFuncsLock.Lock()
	defer a.cancelFuncsLock.Unlock()
	a.cancelFuncs[*request.Pipeline]()
	delete(a.cancelFuncs, *request.Pipeline)
	return google_protobuf.EmptyInstance, nil
}

func (a *apiServer) Version(version int64) error {
	a.versionLock.Lock()
	defer a.versionLock.Unlock()
	a.version = version
	return nil
}

func (a *apiServer) AddShard(shard uint64) error {
	persistClient, err := a.getPersistClient()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())

	a.shardCancelFuncsLock.Lock()
	defer a.shardCancelFuncsLock.Unlock()
	if _, ok := a.shardCancelFuncs[shard]; ok {
		return fmt.Errorf("shard %d is being added twice; this is likely a bug", shard)
	}
	a.shardCancelFuncs[shard] = cancel

	client, err := persistClient.SubscribePipelineInfos(ctx, &persist.SubscribePipelineInfosRequest{
		Shard: &persist.Shard{shard},
	})
	if err != nil {
		return err
	}

	go func() {
		pipelineChan := make(chan *ppsclient.PipelineInfo)
		go func() {
			for {
				pipelineInfo, err := client.Recv()
				if err != nil {
					return
				}
				pipelineChan <- newPipelineInfo(pipelineInfo)
			}
		}()
		for {
			select {
			case <-ctx.Done():
				return
			case pipelineInfo := <-pipelineChan:
				go func() {
					if err := a.runPipeline(pipelineInfo); err != nil {
						protolion.Printf("error running pipeline: %v", err)
					}
				}()
			}
		}
	}()

	// If we did the following before we subscribe to changes, a new pipeline
	// could come in after we did the following but before we subscribe to
	// changes, so we might miss that pipeline.
	pipelineInfos, err := a.persistAPIClient.ListPipelineInfos(ctx, &persist.ListPipelineInfosRequest{
		Shard: &persist.Shard{shard},
	})
	if err != nil {
		return err
	}

	for _, pipelineInfo := range pipelineInfos.PipelineInfo {
		pipelineInfo := pipelineInfo
		go func() {
			if err := a.runPipeline(newPipelineInfo(pipelineInfo)); err != nil {
				protolion.Printf("error running pipeline: %v", err)
			}
		}()
	}

	return nil
}

func (a *apiServer) DeleteShard(shard uint64) error {
	a.cancelFuncsLock.Lock()
	defer a.cancelFuncsLock.Unlock()
	for pipeline, cancelFunc := range a.cancelFuncs {
		if a.hasher.HashPipeline(&pipeline) == shard {
			cancelFunc()
			delete(a.cancelFuncs, pipeline)
		}
	}

	a.shardCancelFuncsLock.Lock()
	defer a.shardCancelFuncsLock.Unlock()
	cancel, ok := a.shardCancelFuncs[shard]
	if !ok {
		return fmt.Errorf("shard %d is being deleted, but it was never added; this is likely a bug", shard)
	}
	cancel()

	return nil
}

func newPipelineInfo(persistPipelineInfo *persist.PipelineInfo) *ppsclient.PipelineInfo {
	return &ppsclient.PipelineInfo{
		Pipeline: &ppsclient.Pipeline{
			Name: persistPipelineInfo.PipelineName,
		},
		Transform:  persistPipelineInfo.Transform,
		Shards:     persistPipelineInfo.Shards,
		Inputs:     persistPipelineInfo.Inputs,
		OutputRepo: persistPipelineInfo.OutputRepo,
	}
}

func (a *apiServer) runPipeline(pipelineInfo *ppsclient.PipelineInfo) error {
	ctx, cancel := context.WithCancel(context.Background())
	a.cancelFuncsLock.Lock()
	if _, ok := a.cancelFuncs[*pipelineInfo.Pipeline]; ok {
		// The pipeline is already being run
		a.cancelFuncsLock.Unlock()
		return nil
	}
	a.cancelFuncs[*pipelineInfo.Pipeline] = cancel
	a.cancelFuncsLock.Unlock()
	repoToLeaves := make(map[string]map[string]bool)
	repoToInput := make(map[string]*ppsclient.PipelineInput)
	var inputRepos []*pfsclient.Repo
	for _, input := range pipelineInfo.Inputs {
		repoToLeaves[input.Repo.Name] = make(map[string]bool)
		repoToInput[input.Repo.Name] = input
		inputRepos = append(inputRepos, &pfsclient.Repo{Name: input.Repo.Name})
	}
	pfsAPIClient, err := a.getPfsClient()
	if err != nil {
		return err
	}
	for {
		var fromCommits []*pfsclient.Commit
		for repo, leaves := range repoToLeaves {
			for leaf := range leaves {
				fromCommits = append(
					fromCommits,
					&pfsclient.Commit{
						Repo: &pfsclient.Repo{Name: repo},
						ID:   leaf,
					})
			}
		}
		listCommitRequest := &pfsclient.ListCommitRequest{
			Repo:       inputRepos,
			CommitType: pfsclient.CommitType_COMMIT_TYPE_READ,
			FromCommit: fromCommits,
			Block:      true,
		}
		commitInfos, err := pfsAPIClient.ListCommit(ctx, listCommitRequest)
		if err != nil {
			return err
		}
		for _, commitInfo := range commitInfos.CommitInfo {
			repoToLeaves[commitInfo.Commit.Repo.Name][commitInfo.Commit.ID] = true
			if commitInfo.ParentCommit != nil {
				delete(repoToLeaves[commitInfo.ParentCommit.Repo.Name], commitInfo.ParentCommit.ID)
			}
			// generate all the permutations of leaves we could use this commit with
			commitSets := [][]*pfsclient.Commit{[]*pfsclient.Commit{}}
			for repoName, leaves := range repoToLeaves {
				if repoName == commitInfo.Commit.Repo.Name {
					continue
				}
				var newCommitSets [][]*pfsclient.Commit
				for _, commitSet := range commitSets {
					for leaf := range leaves {
						newCommitSet := make([]*pfsclient.Commit, len(commitSet)+1)
						copy(newCommitSet, commitSet)
						newCommitSet[len(commitSet)] = &pfsclient.Commit{
							Repo: &pfsclient.Repo{Name: repoName},
							ID:   leaf,
						}
						newCommitSets = append(newCommitSets, newCommitSet)
					}
				}
				commitSets = newCommitSets
			}
			for _, commitSet := range commitSets {
				// + 1 as the commitSet doesn't contain the commit we just got
				if len(commitSet)+1 < len(pipelineInfo.Inputs) {
					continue
				}
				var parentJob *ppsclient.Job
				if commitInfo.ParentCommit != nil {
					parentJob, err = a.parentJob(ctx, pipelineInfo, commitSet, commitInfo)
					if err != nil {
						return err
					}
				}
				var inputs []*ppsclient.JobInput
				for _, commit := range append(commitSet, commitInfo.Commit) {
					inputs = append(inputs, &ppsclient.JobInput{
						Commit: commit,
						Reduce: repoToInput[commit.Repo.Name].Reduce,
					})
				}
				if _, err = a.CreateJob(
					ctx,
					&ppsclient.CreateJobRequest{
						Transform: pipelineInfo.Transform,
						Pipeline:  pipelineInfo.Pipeline,
						Shards:    pipelineInfo.Shards,
						Inputs:    inputs,
						ParentJob: parentJob,
					},
				); err != nil {
					return err
				}
			}
		}
	}
}

func (a *apiServer) parentJob(
	ctx context.Context,
	pipelineInfo *ppsclient.PipelineInfo,
	commitSet []*pfsclient.Commit,
	newCommit *pfsclient.CommitInfo,
) (*ppsclient.Job, error) {
	jobInfo, err := a.ListJob(
		ctx,
		&ppsclient.ListJobRequest{
			Pipeline:    pipelineInfo.Pipeline,
			InputCommit: append(commitSet, newCommit.ParentCommit),
		})
	if err != nil {
		return nil, err
	}
	if len(jobInfo.JobInfo) == 0 {
		return nil, nil
	}
	return jobInfo.JobInfo[0].Job, nil
}

func (a *apiServer) getPfsClient() (pfsclient.APIClient, error) {
	if a.pfsAPIClient == nil {
		var onceErr error
		a.pfsClientOnce.Do(func() {
			clientConn, err := grpc.Dial(a.address, grpc.WithInsecure())
			if err != nil {
				onceErr = err
			}
			a.pfsAPIClient = pfsclient.NewAPIClient(clientConn)
		})
		if onceErr != nil {
			return nil, onceErr
		}
	}
	return a.pfsAPIClient, nil
}

func (a *apiServer) getPersistClient() (persist.APIClient, error) {
	if a.persistAPIClient == nil {
		var onceErr error
		a.persistClientOnce.Do(func() {
			clientConn, err := grpc.Dial(a.address, grpc.WithInsecure())
			if err != nil {
				onceErr = err
			}
			a.persistAPIClient = persist.NewAPIClient(clientConn)
		})
		if onceErr != nil {
			return nil, onceErr
		}
	}
	return a.persistAPIClient, nil
}

func newJobInfo(persistJobInfo *persist.JobInfo) (*ppsclient.JobInfo, error) {
	job := &ppsclient.Job{ID: persistJobInfo.JobID}
	return &ppsclient.JobInfo{
		Job:          job,
		Transform:    persistJobInfo.Transform,
		Pipeline:     &ppsclient.Pipeline{Name: persistJobInfo.PipelineName},
		Shards:       persistJobInfo.Shards,
		Inputs:       persistJobInfo.Inputs,
		ParentJob:    persistJobInfo.ParentJob,
		CreatedAt:    persistJobInfo.CreatedAt,
		OutputCommit: persistJobInfo.OutputCommit,
		State:        persistJobInfo.State,
	}, nil
}

func job(jobInfo *persist.JobInfo) *extensions.Job {
	app := jobInfo.JobID
	shards := int(jobInfo.Shards)
	image := "pachyderm/job-shim"
	if jobInfo.Transform.Image != "" {
		image = jobInfo.Transform.Image
	}
	return &extensions.Job{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Job",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   jobInfo.JobID,
			Labels: labels(app),
		},
		Spec: extensions.JobSpec{
			Selector: &unversioned.LabelSelector{
				MatchLabels: labels(app),
			},
			Parallelism: &shards,
			Completions: &shards,
			Template: api.PodTemplateSpec{
				ObjectMeta: api.ObjectMeta{
					Name:   jobInfo.JobID,
					Labels: labels(app),
				},
				Spec: api.PodSpec{
					Containers: []api.Container{
						{
							Name:    "user",
							Image:   image,
							Command: []string{"/job-shim", jobInfo.JobID},
							SecurityContext: &api.SecurityContext{
								Privileged: &trueVal, // god is this dumb
							},
							ImagePullPolicy: "IfNotPresent",
						},
					},
					RestartPolicy: "OnFailure",
				},
			},
		},
	}
}

func labels(app string) map[string]string {
	return map[string]string{
		"app":   app,
		"suite": suite,
	}
}

type podSlice []kube_api.Pod

func (s podSlice) Len() int {
	return len(s)
}
func (s podSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s podSlice) Less(i, j int) bool {
	return s[i].ObjectMeta.Name < s[j].ObjectMeta.Name
}
