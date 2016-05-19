package server

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/pachyderm/pachyderm/src/client"
	pfsclient "github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	ppsclient "github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pfs/fuse"
	"github.com/pachyderm/pachyderm/src/server/pkg/metrics"
	ppsserver "github.com/pachyderm/pachyderm/src/server/pps"
	"github.com/pachyderm/pachyderm/src/server/pps/persist"

	"github.com/cenkalti/backoff"
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

var (
	ErrEmptyInput = errors.New("job was not started due to empty input")
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
	cancelFuncs          map[string]func()
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

	setDefaultJobInputStrategy(request.Inputs)

	// Currently this happens when someone attempts to run a pipeline once
	if request.Pipeline != nil && request.Transform == nil {
		pipelineInfo, err := a.InspectPipeline(ctx, &ppsclient.InspectPipelineRequest{
			Pipeline: request.Pipeline,
		})
		if err != nil {
			return nil, err
		}
		request.Transform = pipelineInfo.Transform
		request.Parallelism = pipelineInfo.Parallelism
	}

	if request.Parallelism == 0 {
		nodeList, err := a.kubeClient.Nodes().List(kube_api.ListOptions{})
		if err != nil {
			return nil, fmt.Errorf("pachyderm.ppsclient.jobserver: parallelism set to zero and unable to retrieve node list from k8s")
		}

		if len(nodeList.Items) == 0 {
			return nil, fmt.Errorf("pachyderm.ppsclient.jobserver: no k8s nodes found")
		}

		request.Parallelism = uint64(len(nodeList.Items))
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
	_, err = persistClient.InspectJob(ctx, &ppsclient.InspectJobRequest{
		Job: &ppsclient.Job{jobID},
	})
	if err == nil {
		// the job already exists. we simply return
		return &ppsclient.Job{jobID}, nil
	}

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

	repoToFromCommit := make(map[string]*pfsclient.Commit)
	if parentJobInfo != nil {
		if len(request.Inputs) != len(parentJobInfo.Inputs) {
			return nil, fmt.Errorf("parent job does not have the same number of inputs as this job does; this is likely a bug")
		}
		startCommitRequest.ParentID = parentJobInfo.OutputCommit.ID
		for i, jobInput := range request.Inputs {
			if jobInput.Strategy.Incrementality {
				// input isn't being reduced, do it incrementally
				repoToFromCommit[jobInput.Commit.Repo.Name] = parentJobInfo.Inputs[i].Commit
			}
		}
	}

	commit, err := pfsAPIClient.StartCommit(ctx, startCommitRequest)
	if err != nil {
		return nil, err
	}

	// TODO validate job to make sure input commits and output repo exist
	persistJobInfo := &persist.JobInfo{
		JobID:        jobID,
		Transform:    request.Transform,
		Inputs:       request.Inputs,
		ParentJob:    request.ParentJob,
		OutputCommit: commit,
	}
	if request.Pipeline != nil {
		persistJobInfo.PipelineName = request.Pipeline.Name
	}

	// If the job has no input, we respect the specified degree of parallelism
	// Otherwise, we run as many pods as possible given that each pod has some
	// input.
	if len(request.Inputs) == 0 {
		persistJobInfo.Parallelism = request.Parallelism
	} else {
		shardModuli, err := a.computeShardModuli(ctx, request.Inputs, request.Parallelism, repoToFromCommit)
		if err != nil {
			return nil, err
		}

		persistJobInfo.Parallelism = product(shardModuli)
		persistJobInfo.ShardModuli = shardModuli
	}

	if a.kubeClient == nil {
		return nil, fmt.Errorf("pachyderm.ppsclient.jobserver: no job backend")
	}

	_, err = persistClient.CreateJobInfo(ctx, persistJobInfo)
	if err != nil {
		return nil, err
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

	if _, err := a.kubeClient.Jobs(api.NamespaceDefault).Create(job(persistJobInfo)); err != nil {
		return nil, err
	}

	return &ppsclient.Job{
		ID: jobID,
	}, nil
}

// computeShardModuli computes the modulus to use for each input.  In other words,
// it computes how many shards each input repo should be partitioned into.
//
// The algorithm is as follows:
// 1. Each input starts with a modulus of 1
// 2. Double the modulus of the input that currently has the highest size/modulus
// ratio, but only if doing so does not result in empty shards.  If it does, we
// remove the input from further consideration.
// 3. Repeat step 2, until the product of the moduli hits the given parallelism,
// or until all inputs have been removed from consideration.
func (a *apiServer) computeShardModuli(ctx context.Context, inputs []*ppsclient.JobInput, parallelism uint64, repoToFromCommit map[string]*pfsclient.Commit) ([]uint64, error) {
	pfsClient, err := a.getPfsClient()
	if err != nil {
		return nil, err
	}

	var shardModuli []uint64
	var inputSizes []uint64
	for _, input := range inputs {
		commitInfo, err := pfsClient.InspectCommit(ctx, &pfsclient.InspectCommitRequest{
			Commit: input.Commit,
		})
		if err != nil {
			return nil, err
		}

		if commitInfo.SizeBytes == 0 {
			return nil, ErrEmptyInput
		}

		inputSizes = append(inputSizes, commitInfo.SizeBytes)
		shardModuli = append(shardModuli, 1)
	}

	limitHit := make(map[int]bool)
	for {
		max := float64(0)
		modulusIndex := 0
		// Find the modulus to double
		// It should maximize the decrease in size per shard
		for i, inputSize := range inputSizes {
			if !limitHit[i] {
				diff := float64(inputSize) / float64(shardModuli[i])
				if diff > max {
					max = diff
					modulusIndex = i
				}
			}
		}

		b, err := a.noEmptyShards(ctx, inputs[modulusIndex], shardModuli[modulusIndex]*2, repoToFromCommit)
		if err != nil {
			return nil, err
		}

		if b {
			shardModuli[modulusIndex] *= 2
		} else {
			limitHit[modulusIndex] = true
		}

		if product(shardModuli) >= parallelism || len(limitHit) == len(inputs) {
			break
		}
	}

	return shardModuli, nil
}

// product computes the product of a list of integers
//
// The algorithm, originally discovered at Pachyderm, is as follows:
// 1. Set p to 1
// 2. Set p to the product of itself and the first unprocessed number in the list
// 3. Repeat step 2 until we run out of numbers
// 4. Return p
func product(numbers []uint64) uint64 {
	p := uint64(1)
	for _, n := range numbers {
		p *= n
	}
	return p
}

// noEmptyShards computes if every shard will have some input data given an
// input and a modulus number.
//
// TODO: it's very inefficient as of now, since it involves many calls to ListFile
func (a *apiServer) noEmptyShards(ctx context.Context, input *ppsclient.JobInput, modulus uint64, repoToFromCommit map[string]*pfsclient.Commit) (bool, error) {
	pfsClient, err := a.getPfsClient()
	if err != nil {
		return false, err
	}

	for i := 0; i < int(modulus); i++ {
		listFileRequest := &pfsclient.ListFileRequest{
			File: &pfsclient.File{
				Commit: input.Commit,
				Path:   "", // the root directory
			},
			Shard: &pfsclient.Shard{
				FileModulus:  1,
				BlockModulus: 1,
			},
			Recurse: true,
		}
		parentInputCommit := repoToFromCommit[input.Commit.Repo.Name]
		if parentInputCommit != nil && input.Commit.ID != parentInputCommit.ID {
			listFileRequest.FromCommit = parentInputCommit
		}

		switch input.Strategy.Partition {
		case ppsclient.Partition_BLOCK:
			listFileRequest.Shard.BlockNumber = uint64(i)
			listFileRequest.Shard.BlockModulus = modulus
		case ppsclient.Partition_FILE:
			listFileRequest.Shard.FileNumber = uint64(i)
			listFileRequest.Shard.FileModulus = modulus
		case ppsclient.Partition_REPO:
		default:
			return false, fmt.Errorf("unrecognized partition strategy; this is likely a bug")
		}

		fileInfos, err := pfsClient.ListFile(ctx, listFileRequest)
		if err != nil {
			return false, err
		}

		var totalSize uint64
		for _, fileInfo := range fileInfos.FileInfo {
			totalSize += fileInfo.SizeBytes
		}

		if totalSize == 0 {
			return false, nil
		}
	}

	return true, nil
}

func getJobID(req *ppsclient.CreateJobRequest) string {
	// If the job belongs to a pipeline, and the pipeline has inputs,
	// we want to make sure that the same
	// job does not run twice.  We ensure that by generating the job id by
	// hashing the pipeline name and input commits.  That way, two same jobs
	// will have the sam job IDs, therefore won't be created in the database
	// twice.
	if req.Pipeline != nil && len(req.Inputs) > 0 {
		s := req.Pipeline.Name
		for _, input := range req.Inputs {
			s += "/" + input.String()
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

	jobInfo, err := persistClient.StartPod(ctx, request.Job)
	if err != nil {
		return nil, err
	}

	if jobInfo.PodsStarted > jobInfo.Parallelism {
		return nil, fmt.Errorf("job %s already has %d pods", request.Job.ID, jobInfo.Parallelism)
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
	repoToParentJobCommit := make(map[string]*pfsclient.Commit)
	if parentJobInfo != nil {
		for i, jobInput := range jobInfo.Inputs {
			if jobInput.Strategy.Incrementality {
				// input isn't being reduced, do it incrementally
				repoToParentJobCommit[jobInput.Commit.Repo.Name] = parentJobInfo.Inputs[i].Commit
			}
		}
	}

	if jobInfo.OutputCommit == nil {
		return nil, fmt.Errorf("jobInfo.OutputCommit should not be nil (this is likely a bug)")
	}

	var commitMounts []*fuse.CommitMount
	filterNumbers := computeFilterNumber(jobInfo.PodsStarted-1, jobInfo.ShardModuli)
	for i, jobInput := range jobInfo.Inputs {
		commitMount := &fuse.CommitMount{
			Commit: jobInput.Commit,
		}
		parentJobCommit := repoToParentJobCommit[jobInput.Commit.Repo.Name]
		if parentJobCommit != nil && jobInput.Commit.ID != parentJobCommit.ID {
			// We only include a from commit if we have a different commit for a
			// repo than our parent.
			// This means that only the commit that triggered this pipeline will be
			// done incrementally, the other repos will be shown in full
			commitMount.FromCommit = parentJobCommit
		}

		switch jobInput.Strategy.Partition {
		case ppsclient.Partition_BLOCK:
			commitMount.Shard = &pfsclient.Shard{
				BlockNumber:  filterNumbers[i],
				BlockModulus: jobInfo.ShardModuli[i],
			}
		case ppsclient.Partition_FILE:
			commitMount.Shard = &pfsclient.Shard{
				FileNumber:  filterNumbers[i],
				FileModulus: jobInfo.ShardModuli[i],
			}
		case ppsclient.Partition_REPO:
			// empty shard matches everything
			commitMount.Shard = &pfsclient.Shard{}
		default:
			return nil, fmt.Errorf("unrecognized partition strategy: %v; this is likely a bug", jobInput.Strategy.Partition)
		}

		commitMounts = append(commitMounts, commitMount)
	}

	outputCommitMount := &fuse.CommitMount{
		Commit: jobInfo.OutputCommit,
		Alias:  "out",
	}

	// We want to set the commit mount for the output commit such that
	// its FromCommit is its direct parent.  By doing so, we ensure that
	// the files written in previous commits are completely invisible
	// to the job.  Files being written in the current commit will be
	// invisible too due to the way PFS works.  Therefore, /pfs/out/
	// will essentially be a "black box" that can only be written to,
	// but never read from.
	pfsAPIClient, err := a.getPfsClient()
	if err != nil {
		return nil, err
	}
	commitInfo, err := pfsAPIClient.InspectCommit(ctx, &pfsclient.InspectCommitRequest{
		Commit: outputCommitMount.Commit,
	})
	if err != nil {
		return nil, err
	}

	if commitInfo.ParentCommit != nil {
		outputCommitMount.FromCommit = commitInfo.ParentCommit
	}

	commitMounts = append(commitMounts, outputCommitMount)

	// If a job has a parent commit, we expose the parent commit
	// to the job under /pfs/self
	if commitInfo.ParentCommit != nil {
		commitMounts = append(commitMounts, &fuse.CommitMount{
			Commit: commitInfo.ParentCommit,
			Alias:  "self",
		})
	}

	return &ppsserver.StartJobResponse{
		Transform:    jobInfo.Transform,
		CommitMounts: commitMounts,
	}, nil
}

// computeFilterNumber essentially computes a representation of the number N
// as if the base for each digit is the corresponding number in the moduli array
func computeFilterNumber(n uint64, moduli []uint64) []uint64 {
	res := make([]uint64, len(moduli), len(moduli))
	for i := len(moduli) - 1; i >= 0; i-- {
		res[i] = n % moduli[i]
		n = n / moduli[i]
	}
	return res
}

func (a *apiServer) FinishJob(ctx context.Context, request *ppsserver.FinishJobRequest) (response *google_protobuf.Empty, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	persistClient, err := a.getPersistClient()
	if err != nil {
		return nil, err
	}

	var jobInfo *persist.JobInfo
	if request.Success {
		jobInfo, err = persistClient.SucceedPod(ctx, request.Job)
		if err != nil {
			return nil, err
		}
	} else {
		jobInfo, err = persistClient.FailPod(ctx, request.Job)
		if err != nil {
			return nil, err
		}
	}
	if jobInfo.PodsSucceeded+jobInfo.PodsFailed == jobInfo.Parallelism {
		if jobInfo.OutputCommit == nil {
			return nil, fmt.Errorf("jobInfo.OutputCommit should not be nil (this is likely a bug)")
		}
		failed := jobInfo.PodsSucceeded != jobInfo.Parallelism
		pfsAPIClient, err := a.getPfsClient()
		if err != nil {
			return nil, err
		}
		if _, err := pfsAPIClient.FinishCommit(ctx, &pfsclient.FinishCommitRequest{
			Commit: jobInfo.OutputCommit,
			Cancel: failed,
		}); err != nil {
			return nil, err
		}
		commitInfo, err := pfsAPIClient.InspectCommit(ctx, &pfsclient.InspectCommitRequest{
			Commit: jobInfo.OutputCommit,
		})
		if err != nil {
			return nil, err
		}
		jobState := ppsclient.JobState_JOB_STATE_SUCCESS
		if failed || commitInfo.Cancelled {
			jobState = ppsclient.JobState_JOB_STATE_FAILURE
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

	setDefaultPipelineInputStrategy(request.Inputs)

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
		Parallelism:  request.Parallelism,
		Inputs:       request.Inputs,
		OutputRepo:   repo,
		Shard:        a.hasher.HashPipeline(request.Pipeline),
	}
	if _, err := persistClient.CreatePipelineInfo(ctx, persistPipelineInfo); err != nil {
		return nil, err
	}
	return google_protobuf.EmptyInstance, nil
}

// setDefaultPipelineInputStrategy sets strategy to the default for the inputs
// that do not specify a strategy
func setDefaultPipelineInputStrategy(inputs []*ppsclient.PipelineInput) {
	for _, input := range inputs {
		if input.Strategy == nil {
			input.Strategy = client.DefaultStrategy
		}
	}
}

// setDefaultJobInputStrategy sets strategy to the default for the inputs
// that do not specify a strategy
func setDefaultJobInputStrategy(inputs []*ppsclient.JobInput) {
	for _, input := range inputs {
		if input.Strategy == nil {
			input.Strategy = client.DefaultStrategy
		}
	}
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

	if request.Pipeline == nil {
		return nil, fmt.Errorf("Pipeline cannot be nil")
	}

	if _, err := persistClient.DeletePipelineInfo(ctx, request.Pipeline); err != nil {
		return nil, err
	}
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
		IncludeInitial: true,
		Shard:          &persist.Shard{shard},
	})
	if err != nil {
		return err
	}

	go func() {
		for {
			pipelineChange, err := client.Recv()
			if err != nil {
				return
			}

			if pipelineChange.Removed {
				a.cancelFuncsLock.Lock()
				cancel, ok := a.cancelFuncs[pipelineChange.Pipeline.PipelineName]
				if ok {
					cancel()
					delete(a.cancelFuncs, pipelineChange.Pipeline.PipelineName)
				} else {
					protolion.Printf("trying to cancel a pipeline that we are not assigned to; this is likely a bug")
				}
				a.cancelFuncsLock.Unlock()
			} else {
				// We only want to start a goro to run the pipeline if the
				// pipeline has more than one inputs
				go func() {
					b := backoff.NewExponentialBackOff()
					// We set MaxElapsedTime to 0 because we want the retry to
					// never stop.
					// However, ideally we should crash this pps server so the
					// pipeline gets reassigned to another pps server.
					// The reason we don't do that right now is that pps and
					// pfs are bundled together, so by crashing this program
					// we will be crashing a pfs node too, which might cause
					// cascading failures as other pps nodes might be depending
					// on it.
					b.MaxElapsedTime = 0
					backoff.Retry(func() error {
						if err := a.runPipeline(newPipelineInfo(pipelineChange.Pipeline)); err != nil && !isContextCancelled(err) {
							protolion.Printf("error running pipeline: %v", err)
							return err
						}
						return nil
					}, b)
				}()
			}
		}
	}()

	return nil
}

func isContextCancelled(err error) bool {
	return err == context.Canceled || strings.Contains(err.Error(), context.Canceled.Error())
}

func (a *apiServer) DeleteShard(shard uint64) error {
	a.cancelFuncsLock.Lock()
	defer a.cancelFuncsLock.Unlock()
	for pipeline, cancelFunc := range a.cancelFuncs {
		if a.hasher.HashPipeline(&ppsclient.Pipeline{
			Name: pipeline,
		}) == shard {
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
		Transform:   persistPipelineInfo.Transform,
		Parallelism: persistPipelineInfo.Parallelism,
		Inputs:      persistPipelineInfo.Inputs,
		OutputRepo:  persistPipelineInfo.OutputRepo,
	}
}

func (a *apiServer) runPipeline(pipelineInfo *ppsclient.PipelineInfo) error {
	ctx, cancel := context.WithCancel(context.Background())
	returnNil := func() bool {
		a.cancelFuncsLock.Lock()
		defer a.cancelFuncsLock.Unlock()
		if _, ok := a.cancelFuncs[pipelineInfo.Pipeline.Name]; ok {
			// The pipeline is already being run
			return true
		}
		if len(pipelineInfo.Inputs) == 0 {
			// this pipeline does not have inputs; there is nothing to be done
			return true
		}

		a.cancelFuncs[pipelineInfo.Pipeline.Name] = cancel
		return false
	}()
	if returnNil {
		return nil
	}
	repoToLeaves := make(map[string]map[string]bool)
	repoToInput := make(map[string]*ppsclient.PipelineInput)
	repoIsIncremental := make(map[string]bool)
	var inputRepos []*pfsclient.Repo
	for _, input := range pipelineInfo.Inputs {
		repoToLeaves[input.Repo.Name] = make(map[string]bool)
		repoToInput[input.Repo.Name] = input
		inputRepos = append(inputRepos, &pfsclient.Repo{Name: input.Repo.Name})
		repoIsIncremental[input.Repo.Name] = input.Strategy.Incrementality
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
				if commitInfo.ParentCommit != nil && repoIsIncremental[commitInfo.Commit.Repo.Name] {
					parentJob, err = a.parentJob(ctx, pipelineInfo, commitSet, commitInfo)
					if err != nil {
						return err
					}
				}
				var inputs []*ppsclient.JobInput
				for _, commit := range append(commitSet, commitInfo.Commit) {
					inputs = append(inputs, &ppsclient.JobInput{
						Commit:   commit,
						Strategy: repoToInput[commit.Repo.Name].Strategy,
					})
				}
				if _, err = a.CreateJob(
					ctx,
					&ppsclient.CreateJobRequest{
						Transform:   pipelineInfo.Transform,
						Pipeline:    pipelineInfo.Pipeline,
						Parallelism: pipelineInfo.Parallelism,
						Inputs:      inputs,
						ParentJob:   parentJob,
					},
				); err != nil && err != ErrEmptyInput {
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
		Parallelism:  persistJobInfo.Parallelism,
		Inputs:       persistJobInfo.Inputs,
		ParentJob:    persistJobInfo.ParentJob,
		CreatedAt:    persistJobInfo.CreatedAt,
		OutputCommit: persistJobInfo.OutputCommit,
		State:        persistJobInfo.State,
	}, nil
}

func RepoNameToEnvString(repoName string) string {
	return strings.ToUpper(repoName)
}

func job(jobInfo *persist.JobInfo) *extensions.Job {
	app := jobInfo.JobID
	parallelism := int(jobInfo.Parallelism)
	image := "pachyderm/job-shim"
	if jobInfo.Transform.Image != "" {
		image = jobInfo.Transform.Image
	}

	var jobEnv []api.EnvVar
	jobEnv = append(
		jobEnv,
		api.EnvVar{
			Name:  "PACH_OUTPUT_COMMIT_ID",
			Value: jobInfo.OutputCommit.ID,
		},
	)
	for _, input := range jobInfo.Inputs {
		jobEnv = append(
			jobEnv,
			api.EnvVar{
				Name:  fmt.Sprintf("PACH_%v_COMMIT_ID", RepoNameToEnvString(input.Commit.Repo.Name)),
				Value: input.Commit.ID,
			},
		)
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
			Parallelism: &parallelism,
			Completions: &parallelism,
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
							Env:             jobEnv,
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
