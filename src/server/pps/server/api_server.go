package server

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"fmt"
	"math"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/src/client"
	pfsclient "github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	ppsclient "github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pfs/fuse"
	"github.com/pachyderm/pachyderm/src/server/pkg/lease"
	"github.com/pachyderm/pachyderm/src/server/pkg/metrics"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	pfs_sync "github.com/pachyderm/pachyderm/src/server/pkg/sync"
	ppsserver "github.com/pachyderm/pachyderm/src/server/pps"
	"github.com/pachyderm/pachyderm/src/server/pps/persist"

	"github.com/cenkalti/backoff"
	"go.pedge.io/lion/proto"
	"go.pedge.io/proto/rpclog"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/apis/batch"
	kube "k8s.io/kubernetes/pkg/client/unversioned"
	kube_labels "k8s.io/kubernetes/pkg/labels"
)

const (
	// MaxPodsPerChunk is the maximum number of pods we can schedule for each
	// chunk in case of failures.
	MaxPodsPerChunk = 3
	// DefaultUserImage is the image used for jobs when the user does not specify
	// an image.
	DefaultUserImage = "ubuntu:16.04"
)

var (
	trueVal = true
	suite   = "pachyderm"
)

func newErrJobNotFound(job string) error {
	return fmt.Errorf("job %v not found", job)
}

func newErrPipelineNotFound(pipeline string) error {
	return fmt.Errorf("pipeline %v not found", pipeline)
}

func newErrPipelineExists(pipeline string) error {
	return fmt.Errorf("pipeline %v already exists", pipeline)
}

type errEmptyInput struct {
	error
}

func newErrEmptyInput(commitID string) *errEmptyInput {
	return &errEmptyInput{
		error: fmt.Errorf("job was not started due to empty input at commit %v", commitID),
	}
}

func newErrParentInputsMismatch(parent string) error {
	return fmt.Errorf("job does not have the same set of inputs as its parent %v", parent)
}

type apiServer struct {
	protorpclog.Logger
	hasher                  *ppsserver.Hasher
	address                 string
	pfsAPIClient            pfsclient.APIClient
	pfsClientOnce           sync.Once
	persistAPIClient        persist.APIClient
	persistClientOnce       sync.Once
	kubeClient              *kube.Client
	shardCancelFuncs        map[uint64]func()
	shardCancelFuncsLock    sync.Mutex
	pipelineCancelFuncs     map[string]func()
	pipelineCancelFuncsLock sync.Mutex
	jobCancelFuncs          map[string]func()
	jobCancelFuncsLock      sync.Mutex
	version                 int64
	// versionLock protects the version field.
	// versionLock must be held BEFORE reading from version and UNTIL all
	// requests using version have returned
	versionLock        sync.RWMutex
	namespace          string
	jobShimImage       string
	jobImagePullPolicy string
	reporter           *metrics.Reporter
}

// JobInputs implements sort.Interface so job inputs can be sorted
// We sort job inputs based on repo names
type JobInputs []*ppsclient.JobInput

func (inputs JobInputs) Len() int {
	return len(inputs)
}

func (inputs JobInputs) Less(i, j int) bool {
	return inputs[i].Commit.Repo.Name < inputs[j].Commit.Repo.Name
}

func (inputs JobInputs) Swap(i, j int) {
	inputs[i], inputs[j] = inputs[j], inputs[i]
}

// GetExpectedNumWorkers computes the expected number of workers that pachyderm will start given
// the ParallelismSpec 'spec'.
//
// This is only exported for testing
func GetExpectedNumWorkers(kubeClient *kube.Client, spec *ppsclient.ParallelismSpec) (uint64, error) {
	coefficient := 0.0 // Used if [spec.Strategy == PROPORTIONAL] or [spec.Constant == 0]
	if spec == nil {
		// Unset ParallelismSpec is handled here. Currently we start one worker per
		// node
		coefficient = 1.0
	} else if spec.Strategy == ppsclient.ParallelismSpec_CONSTANT {
		if spec.Constant > 0 {
			return spec.Constant, nil
		}
		// Zero-initialized ParallelismSpec is handled here. Currently we start one
		// worker per node
		coefficient = 1
	} else if spec.Strategy == ppsclient.ParallelismSpec_COEFFICIENT {
		coefficient = spec.Coefficient
	} else {
		return 0, fmt.Errorf("Unable to interpret ParallelismSpec strategy %s", spec.Strategy)
	}
	if coefficient == 0.0 {
		return 0, fmt.Errorf("Ended up with coefficient == 0 (no workers) after interpreting ParallelismSpec %s", spec.Strategy)
	}

	// Start ('coefficient' * 'nodes') workers. Determine number of workers
	nodeList, err := kubeClient.Nodes().List(api.ListOptions{})
	if err != nil {
		return 0, fmt.Errorf("Unable to retrieve node list from k8s to determine parallelism")
	}
	if len(nodeList.Items) == 0 {
		return 0, fmt.Errorf("pachyderm.ppsclient.jobserver: no k8s nodes found")
	}
	result := math.Floor(coefficient * float64(len(nodeList.Items)))
	return uint64(math.Max(result, 1)), nil
}

// onTheSameBranch tests if two commits are on the same branch.
func onTheSameBranch(a, b *pfsclient.Commit) bool {
	aParts := strings.Split(a.ID, "/")
	bParts := strings.Split(b.ID, "/")
	return aParts[0] == bParts[0]
}

func (a *apiServer) CreateJob(ctx context.Context, request *ppsclient.CreateJobRequest) (response *ppsclient.Job, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "CreateJob")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	persistClient, err := a.getPersistClient()
	if err != nil {
		return nil, err
	}

	// We need to sort job inputs because the following code depends on
	// the invariant that inputs[i] matches parentInputs[i]
	sort.Sort(JobInputs(request.Inputs))

	// In case some inputs have not provided a method, we set the default
	// method for them
	setDefaultJobInputMethod(request.Inputs)

	var pipelineInfo *ppsclient.PipelineInfo
	if request.Pipeline != nil {
		pipelineInfo, err = a.InspectPipeline(ctx, &ppsclient.InspectPipelineRequest{
			Pipeline: request.Pipeline,
		})
		if err != nil {
			return nil, err
		}
	}
	// Currently this happens when someone attempts to run a pipeline once
	if request.Pipeline != nil && request.Transform == nil {
		request.Transform = pipelineInfo.Transform
		request.ParallelismSpec = pipelineInfo.ParallelismSpec
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
		// We Block on our parent's state because the job may access its
		// parent's output while it's running so we want the job and its output
		// commit to have finished.
		inspectJobRequest := &ppsclient.InspectJobRequest{
			Job:        request.ParentJob,
			BlockState: true,
		}
		parentJobInfo, err = persistClient.InspectJob(ctx, inspectJobRequest)
		if err != nil {
			return nil, err
		}

		// Check that the parent job has the same set of inputs as the current job
		if len(parentJobInfo.Inputs) != len(request.Inputs) {
			return nil, newErrParentInputsMismatch(parentJobInfo.JobID)
		}

		for i, input := range request.Inputs {
			if parentJobInfo.Inputs[i].Commit.Repo.Name != input.Commit.Repo.Name {
				return nil, newErrParentInputsMismatch(parentJobInfo.JobID)
			}
		}
	}

	pfsAPIClient, err := a.getPfsClient()
	if err != nil {
		return nil, err
	}

	jobID := getJobID(request)
	if !request.Force {
		_, err = persistClient.InspectJob(ctx, &ppsclient.InspectJobRequest{
			Job: client.NewJob(jobID),
		})
		if err == nil {
			// the job already exists. we simply return
			return &ppsclient.Job{ID: jobID}, nil
		}
	}

	// If JobInfo.Pipeline is set, use the pipeline repo
	var outputRepo *pfsclient.Repo
	if request.Pipeline != nil {
		outputRepo = ppsserver.PipelineRepo(&ppsclient.Pipeline{Name: request.Pipeline.Name})
		if parentJobInfo != nil && parentJobInfo.OutputCommit.Repo.Name != outputRepo.Name {
			return nil, fmt.Errorf("parent job was not part of the same pipeline; this is likely a bug")
		}
	} else {
		if parentJobInfo != nil {
			outputRepo = parentJobInfo.OutputCommit.Repo
		} else {
			// Create a repo for this job
			outputRepo = ppsserver.JobRepo(&ppsclient.Job{
				ID: jobID,
			})
			var provenance []*pfsclient.Repo
			for _, input := range request.Inputs {
				provenance = append(provenance, input.Commit.Repo)
			}
			defer func() {
				if retErr != nil {
					req := &pfsclient.DeleteRepoRequest{
						Repo: outputRepo,
					}
					_, err := pfsAPIClient.DeleteRepo(ctx, req)
					if err != nil {
						protolion.Errorf("could not rollback repo creation %s", err.Error())
						func() { a.Log(req, nil, err, 0) }()
					}
				}
			}()

			if _, err := pfsAPIClient.CreateRepo(ctx,
				&pfsclient.CreateRepoRequest{
					Repo:       outputRepo,
					Provenance: provenance,
				}); err != nil {
				return nil, err
			}
		}
	}

	// We want to fork the output commit to a new branch if
	// this job has a parent, but one or more of the input commits are not on
	// the same branch as their corresponding parent input commits.
	var fork bool
	repoToFromCommit := make(map[string]*pfsclient.Commit)
	if parentJobInfo != nil {
		if len(request.Inputs) != len(parentJobInfo.Inputs) {
			return nil, fmt.Errorf("parent job does not have the same number of inputs as this job does; this is likely a bug")
		}
		for i, jobInput := range request.Inputs {
			if jobInput.Method.Incremental != ppsclient.Incremental_NONE {
				repoToFromCommit[jobInput.Commit.Repo.Name] = parentJobInfo.Inputs[i].Commit
				if !onTheSameBranch(jobInput.Commit, parentJobInfo.Inputs[i].Commit) {
					fork = true
				}
			}
		}
	}

	var outputCommit *pfsclient.Commit
	var provenance []*pfsclient.Commit
	for _, input := range request.Inputs {
		provenance = append(provenance, input.Commit)
	}
	parent := &pfsclient.Commit{
		Repo: outputRepo,
	}
	if fork {
		forkCommitRequest := &pfsclient.ForkCommitRequest{
			Provenance: provenance,
			Parent:     parent,
			Branch:     uuid.NewWithoutDashes(),
		}
		forkCommitRequest.Parent.ID = parentJobInfo.OutputCommit.ID
		outputCommit, err = pfsAPIClient.ForkCommit(ctx, forkCommitRequest)
		if err != nil {
			return nil, err
		}
	} else {
		startCommitRequest := &pfsclient.StartCommitRequest{
			Provenance: provenance,
			Parent:     parent,
		}
		if parentJobInfo != nil {
			startCommitRequest.Parent.ID = parentJobInfo.OutputCommit.ID
		} else {
			// Without a parent, we start the commit on a random branch
			startCommitRequest.Parent.ID = uuid.NewWithoutDashes()
		}
		outputCommit, err = pfsAPIClient.StartCommit(ctx, startCommitRequest)
		if err != nil {
			if strings.Contains(err.Error(), "already exists") {
				if _, err := pfsAPIClient.DeleteCommit(
					ctx,
					&pfsclient.DeleteCommitRequest{
						Commit: client.NewCommit(parent.Repo.Name, strings.Split(parent.ID, "/")[0]),
					},
				); err != nil {
					return nil, err
				}
				outputCommit, err = pfsAPIClient.StartCommit(ctx, startCommitRequest)
				if err != nil {
					return nil, err
				}
			} else {
				return nil, err
			}
		}
	}

	// TODO validate job to make sure input commits and output repo exist
	persistJobInfo := &persist.JobInfo{
		JobID:        jobID,
		Transform:    request.Transform,
		Inputs:       request.Inputs,
		ParentJob:    request.ParentJob,
		OutputCommit: outputCommit,
		Shard: a.hasher.HashJob(&ppsclient.Job{
			ID: jobID,
		}),
		Service: request.Service,
		Output:  request.Output,
	}
	if request.Pipeline != nil {
		persistJobInfo.PipelineName = pipelineInfo.Pipeline.Name
		persistJobInfo.PipelineVersion = pipelineInfo.Version
	}

	// If the job has no input, we respect the specified degree of parallelism
	// Otherwise, we run as many pods as possible given that each pod has some
	// input.
	var shardModuli []uint64
	if len(request.Inputs) == 0 {
		persistJobInfo.ParallelismSpec = request.ParallelismSpec
	} else {
		numWorkers, err := GetExpectedNumWorkers(a.kubeClient, request.ParallelismSpec)
		if err != nil {
			return nil, err
		}
		shardModuli, err = a.shardModuli(ctx, request.Inputs, numWorkers, repoToFromCommit)
		_, ok := err.(*errEmptyInput)
		if err != nil && !ok {
			return nil, err
		}

		if ok {
			// If an input is empty, and RunEmpty flag is not set, then we simply
			// create an empty job and finish the commit.
			persistJobInfo.State = ppsclient.JobState_JOB_EMPTY
			_, err = persistClient.CreateJobInfo(ctx, persistJobInfo)
			if err != nil {
				return nil, err
			}

			_, err = pfsAPIClient.FinishCommit(ctx, &pfsclient.FinishCommitRequest{
				Commit: outputCommit,
			})
			if err != nil {
				return nil, err
			}

			return &ppsclient.Job{
				ID: jobID,
			}, nil
		}

		persistJobInfo.ParallelismSpec = &ppsclient.ParallelismSpec{
			Strategy: ppsclient.ParallelismSpec_CONSTANT,
			Constant: product(shardModuli),
		}
		persistJobInfo.DefaultShardModuli = shardModuli
	}

	if a.kubeClient == nil {
		return nil, fmt.Errorf("pachyderm.ppsclient.jobserver: no job backend")
	}

	// Create chunks for this job
	// We need to create chunks before we create the JobInfo object itself,
	// because once a JobInfo object has been created, a jobManager routine
	// will be kicked off, and it will check to see if all chunks have been
	// finished.  If there are no chunks, then the jobManager will think
	// that the job has been finished, when in reality the chunks haven't
	// even been created.
	//
	// Right now numWorkers == numChunks, but it may not remain that way.
	numChunks, err := GetExpectedNumWorkers(a.kubeClient, persistJobInfo.ParallelismSpec)
	if err != nil {
		return nil, err
	}
	var chunks []*persist.Chunk
	for i := 0; i < int(numChunks); i++ {
		chunk := &persist.Chunk{
			ID:     uuid.New(),
			JobID:  jobID,
			Moduli: shardModuli,
			Index:  uint64(i),
			State:  persist.ChunkState_UNASSIGNED,
		}
		chunks = append(chunks, chunk)
	}
	// TODO: if there are a huge number of chunks, could it be a problem that
	// we are sending all of them in one request?
	_, err = persistClient.AddChunk(ctx, &persist.AddChunkRequest{
		Chunks: chunks,
	})
	if err != nil {
		return nil, err
	}

	if request.Service != nil {
		rc, service, err := service(a.kubeClient, persistJobInfo, a.jobShimImage, a.jobImagePullPolicy, request.Service.InternalPort, request.Service.ExternalPort)
		if err != nil {
			return nil, err
		}
		if _, err := a.kubeClient.ReplicationControllers(a.namespace).Create(rc); err != nil {
			return nil, err
		}
		if _, err := a.kubeClient.Services(a.namespace).Create(service); err != nil {
			return nil, err
		}
	} else {
		job, err := job(a.kubeClient, persistJobInfo, a.jobShimImage, a.jobImagePullPolicy)
		if err != nil {
			return nil, err
		}
		if _, err := a.kubeClient.Extensions().Jobs(a.namespace).Create(job); err != nil {
			return nil, err
		}
	}

	// This should be the last thing the function does
	_, err = persistClient.CreateJobInfo(ctx, persistJobInfo)
	if err != nil {
		return nil, err
	}

	return &ppsclient.Job{
		ID: jobID,
	}, nil
}

// shardModuli computes the modulus to use for each input.  In other words,
// it computes how many shards each input repo should be partitioned into.
//
// The algorithm is as follows:
// 1. Each input starts with a modulus of 1
// 2. Double the modulus of the input that currently has the highest size/modulus
// ratio, but only if doing so does not result in empty shards.  If it does, we
// remove the input from further consideration.
// 3. Repeat step 2, until the product of the moduli hits the given parallelism,
// or until all inputs have been removed from consideration.
func (a *apiServer) shardModuli(ctx context.Context, inputs []*ppsclient.JobInput, parallelism uint64, repoToFromCommit map[string]*pfsclient.Commit) ([]uint64, error) {
	pfsClient, err := a.getPfsClient()
	if err != nil {
		return nil, err
	}

	var shardModuli []uint64
	var inputSizes []uint64
	limitHit := make(map[int]bool)
	for i, input := range inputs {
		if input.Method.Partition == ppsclient.Partition_REPO {
			// A global input shouldn't be partitioned
			limitHit[i] = true
		}

		commitInfo, err := pfsClient.InspectCommit(ctx, &pfsclient.InspectCommitRequest{
			Commit: input.Commit,
		})
		if err != nil {
			return nil, err
		}

		if commitInfo.SizeBytes == 0 {
			if input.RunEmpty {
				// An empty input shouldn't be partitioned
				limitHit[i] = true
			} else {
				return nil, newErrEmptyInput(input.Commit.ID)
			}
		}

		inputSizes = append(inputSizes, commitInfo.SizeBytes)
		shardModuli = append(shardModuli, 1)
	}

	for product(shardModuli) < parallelism && len(limitHit) < len(inputs) {
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
			Mode: pfsclient.ListFileMode_ListFile_RECURSE,
		}
		parentInputCommit := repoToFromCommit[input.Commit.Repo.Name]
		if parentInputCommit != nil && input.Commit.ID != parentInputCommit.ID {
			listFileRequest.DiffMethod = &pfsclient.DiffMethod{
				FromCommit: parentInputCommit,
				FullFile:   input.Method.Incremental == ppsclient.Incremental_FULL,
			}
		}

		switch input.Method.Partition {
		case ppsclient.Partition_BLOCK:
			listFileRequest.Shard.BlockNumber = uint64(i)
			listFileRequest.Shard.BlockModulus = modulus
		case ppsclient.Partition_FILE:
			listFileRequest.Shard.FileNumber = uint64(i)
			listFileRequest.Shard.FileModulus = modulus
		case ppsclient.Partition_REPO:
		default:
			return false, fmt.Errorf("unrecognized partition method; this is likely a bug")
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
	if req.Pipeline != nil && len(req.Inputs) > 0 && !req.Force {
		s := req.Pipeline.Name
		s += req.Transform.String()
		for _, input := range req.Inputs {
			s += "/" + input.String()
		}

		hash := md5.Sum([]byte(s))
		return fmt.Sprintf("%x", hash)
	}

	return uuid.NewWithoutDashes()
}

func (a *apiServer) InspectJob(ctx context.Context, request *ppsclient.InspectJobRequest) (response *ppsclient.JobInfo, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "InspectJob")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	persistClient, err := a.getPersistClient()
	if err != nil {
		return nil, err
	}

	persistJobInfo, err := persistClient.InspectJob(ctx, request)
	if err != nil {
		return nil, err
	}

	jobInfo, err := newJobInfo(persistJobInfo)
	if err != nil {
		return nil, err
	}

	chunks, err := persistClient.GetChunksForJob(ctx, request.Job)
	if err != nil {
		return nil, err
	}

	for _, chunk := range chunks.Chunks {
		var pods []*ppsclient.Pod
		for i, pod := range chunk.Pods {
			pod := &ppsclient.Pod{
				Name:         pod.Name,
				OutputCommit: pod.OutputCommit,
			}
			if i == len(chunk.Pods)-1 && chunk.State == persist.ChunkState_SUCCESS {
				pod.State = ppsclient.PodState_POD_SUCCESS
			} else if i == len(chunk.Pods)-1 && chunk.State == persist.ChunkState_ASSIGNED {
				pod.State = ppsclient.PodState_POD_RUNNING
			} else {
				pod.State = ppsclient.PodState_POD_FAILED
			}
			pods = append(pods, pod)
		}
		c := &ppsclient.Chunk{
			ID:   chunk.ID,
			Pods: pods,
		}
		switch chunk.State {
		case persist.ChunkState_UNASSIGNED:
			c.State = ppsclient.ChunkState_CHUNK_UNASSIGNED
		case persist.ChunkState_ASSIGNED:
			c.State = ppsclient.ChunkState_CHUNK_ASSIGNED
		case persist.ChunkState_SUCCESS:
			c.State = ppsclient.ChunkState_CHUNK_SUCCESS
		case persist.ChunkState_FAILED:
			c.State = ppsclient.ChunkState_CHUNK_FAILURE
		case persist.ChunkState_SPLITTED:
			continue
		}
		jobInfo.Chunks = append(jobInfo.Chunks, c)
	}

	return jobInfo, nil
}

func (a *apiServer) ListJob(ctx context.Context, request *ppsclient.ListJobRequest) (response *ppsclient.JobInfos, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "ListJob")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

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

func (a *apiServer) DeleteJob(ctx context.Context, request *ppsclient.DeleteJobRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "DeleteJob")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	pfsAPIClient, err := a.getPfsClient()
	if err != nil {
		return nil, err
	}
	persistClient, err := a.getPersistClient()
	if err != nil {
		return nil, err
	}
	jobInfo, err := persistClient.InspectJob(ctx, &ppsclient.InspectJobRequest{
		Job: request.Job,
	})
	if err != nil {
		return nil, err
	}
	if jobInfo.PipelineName != "" {
		return nil, fmt.Errorf("cannot delete job (%v) that belongs to a pipeline (%v)", request.Job.ID, jobInfo.PipelineName)
	}
	if err := a.deleteJob(ctx, jobInfo); err != nil {
		return nil, err
	}
	if _, err := pfsAPIClient.DeleteRepo(ctx, &pfsclient.DeleteRepoRequest{Repo: jobInfo.OutputCommit.Repo}); err != nil {
		return nil, err
	}
	if _, err := persistClient.DeleteJobInfo(ctx, request.Job); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

func (a *apiServer) GetLogs(request *ppsclient.GetLogsRequest, apiGetLogsServer ppsclient.API_GetLogsServer) (retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, nil, retErr, time.Since(start)) }(time.Now())
	pods, err := a.jobPods(request.Job)
	if err != nil {
		return err
	}
	if len(pods) == 0 {
		return newErrJobNotFound(request.Job.ID)
	}
	// sort the pods to make sure that the indexes are stable
	sort.Sort(podSlice(pods))
	logs := make([][]byte, len(pods))
	var eg errgroup.Group
	for i, pod := range pods {
		i := i
		pod := pod
		eg.Go(func() error {
			result := a.kubeClient.Pods(a.namespace).GetLogs(
				pod.ObjectMeta.Name, &api.PodLogOptions{}).Do()
			value, err := result.Raw()
			if err != nil {
				return err
			}
			var buffer bytes.Buffer
			scanner := bufio.NewScanner(bytes.NewBuffer(value))
			for scanner.Scan() {
				fmt.Fprintf(&buffer, "%d | %s\n", i, scanner.Text())
			}
			logs[i] = buffer.Bytes()
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return err
	}
	for _, log := range logs {
		if err := apiGetLogsServer.Send(&types.BytesValue{Value: log}); err != nil {
			return err
		}
	}
	return nil
}

func (a *apiServer) StartPod(ctx context.Context, request *ppsserver.StartPodRequest) (response *ppsserver.StartPodResponse, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "StartJob")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	persistClient, err := a.getPersistClient()
	if err != nil {
		return nil, err
	}

	jobInfo, err := persistClient.StartJob(ctx, request.Job)
	if err != nil {
		return nil, err
	}

	if jobInfo.Transform == nil {
		return nil, fmt.Errorf("jobInfo.Transform should not be nil (this is likely a bug)")
	}

	var parentJobInfo *persist.JobInfo
	if jobInfo.ParentJob != nil {
		inspectJobRequest := &ppsclient.InspectJobRequest{
			Job: jobInfo.ParentJob,
		}
		parentJobInfo, err = persistClient.InspectJob(ctx, inspectJobRequest)
		if err != nil {
			return nil, err
		}
	}
	repoToParentJobCommit := make(map[string]*pfsclient.Commit)
	if parentJobInfo != nil {
		for i, jobInput := range jobInfo.Inputs {
			if jobInput.Method.Incremental != ppsclient.Incremental_NONE {
				repoToParentJobCommit[jobInput.Commit.Repo.Name] = parentJobInfo.Inputs[i].Commit
			}
		}
	}

	var provenance []*pfsclient.Commit
	for _, input := range jobInfo.Inputs {
		provenance = append(provenance, input.Commit)
	}

	pfsAPIClient, err := a.getPfsClient()
	if err != nil {
		return nil, err
	}

	var commit *pfsclient.Commit
	if parentJobInfo != nil {
		if len(jobInfo.Inputs) != len(parentJobInfo.Inputs) {
			return nil, fmt.Errorf("parent job does not have the same number of inputs as this job does; this is likely a bug")
		}
		// If we have a parent, then we fork the parent
		forkReq := &pfsclient.ForkCommitRequest{
			Provenance: provenance,
			Parent: &pfsclient.Commit{
				Repo: jobInfo.OutputCommit.Repo,
				ID:   parentJobInfo.OutputCommit.ID,
			},
			Branch: fmt.Sprintf("pod_%v", uuid.NewWithoutDashes()),
		}
		commit, err = pfsAPIClient.ForkCommit(ctx, forkReq)
		if err != nil {
			return nil, err
		}
	} else {
		// If we don't have a parent, then we simply start a new commit
		startCommitReq := &pfsclient.StartCommitRequest{
			Provenance: provenance,
			Parent: &pfsclient.Commit{
				Repo: jobInfo.OutputCommit.Repo,
				ID:   fmt.Sprintf("pod_%v", uuid.NewWithoutDashes()),
			},
		}
		commit, err = pfsAPIClient.StartCommit(ctx, startCommitReq)
		if err != nil {
			return nil, err
		}
	}

	// We archive the commit before we finish it, to ensure that a pipeline
	// that is listing finished commits do not end up seeing this commit
	_, err = pfsAPIClient.ArchiveCommit(ctx, &pfsclient.ArchiveCommitRequest{
		Commits: []*pfsclient.Commit{commit},
	})
	if err != nil {
		return nil, err
	}

	chunk, err := persistClient.ClaimChunk(ctx, &persist.ClaimChunkRequest{
		JobID: request.Job.ID,
		Pod: &persist.Pod{
			Name:         request.PodName,
			OutputCommit: commit,
		},
	})
	if err != nil {
		return nil, err
	}

	var commitMounts []*fuse.CommitMount
	filterNumbers := filterNumber(chunk.Index, chunk.Moduli)
	for i, jobInput := range jobInfo.Inputs {
		commitMount := &fuse.CommitMount{
			Commit: jobInput.Commit,
			Lazy:   jobInput.Lazy,
		}
		parentJobCommit := repoToParentJobCommit[jobInput.Commit.Repo.Name]
		if parentJobCommit != nil && jobInput.Commit.ID != parentJobCommit.ID {
			// We only include a from commit if we have a different commit for a
			// repo than our parent.
			// This means that only the commit that triggered this pipeline will be
			// done incrementally, the other repos will be shown in full
			commitMount.DiffMethod = &pfsclient.DiffMethod{
				FromCommit: parentJobCommit,
				FullFile:   jobInput.Method != nil && jobInput.Method.Incremental == ppsclient.Incremental_FULL,
			}
		}

		switch jobInput.Method.Partition {
		case ppsclient.Partition_BLOCK:
			commitMount.Shard = &pfsclient.Shard{
				BlockNumber:  filterNumbers[i],
				BlockModulus: chunk.Moduli[i],
			}
		case ppsclient.Partition_FILE:
			commitMount.Shard = &pfsclient.Shard{
				FileNumber:  filterNumbers[i],
				FileModulus: chunk.Moduli[i],
			}
		case ppsclient.Partition_REPO:
			// empty shard matches everything
			commitMount.Shard = &pfsclient.Shard{}
		default:
			return nil, fmt.Errorf("unrecognized partition method: %v; this is likely a bug", jobInput.Method.Partition)
		}

		commitMounts = append(commitMounts, commitMount)
	}

	outputCommitMount := &fuse.CommitMount{
		Commit: commit,
		Alias:  "out",
	}
	commitMounts = append(commitMounts, outputCommitMount)

	// If a job has a parent commit, we expose the parent commit
	// to the job under /pfs/prev
	commitInfo, err := pfsAPIClient.InspectCommit(ctx, &pfsclient.InspectCommitRequest{
		Commit: outputCommitMount.Commit,
	})
	if err != nil {
		return nil, err
	}
	if commitInfo.ParentCommit != nil {
		commitMounts = append(commitMounts, &fuse.CommitMount{
			Commit: commitInfo.ParentCommit,
			Alias:  "prev",
		})
	}

	return &ppsserver.StartPodResponse{
		ChunkID:      chunk.ID,
		Transform:    jobInfo.Transform,
		CommitMounts: commitMounts,
	}, nil
}

// filterNumber essentially computes a representation of the number N
// as if the base for each digit is the corresponding number in the moduli array
func filterNumber(n uint64, moduli []uint64) []uint64 {
	res := make([]uint64, len(moduli), len(moduli))
	for i := len(moduli) - 1; i >= 0; i-- {
		res[i] = n % moduli[i]
		n = n / moduli[i]
	}
	return res
}

func (a *apiServer) ContinuePod(ctx context.Context, request *ppsserver.ContinuePodRequest) (response *ppsserver.ContinuePodResponse, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	persistClient, err := a.getPersistClient()
	if err != nil {
		return nil, err
	}
	chunk, err := persistClient.RenewChunk(ctx, &persist.RenewChunkRequest{
		ChunkID: request.ChunkID,
		PodName: request.PodName,
	})
	if err != nil {
		return nil, err
	}
	response = &ppsserver.ContinuePodResponse{}
	// A pod should only continue executing if it still owns the chunk, or the chunk
	// has been finished (in which case the pod just continues executing and eventually peacefully exit).
	// Otherwise, we tell the pod to exit and restart, so it can claim another chunk
	// to process.
	if (chunk.Owner == request.PodName && chunk.State == persist.ChunkState_ASSIGNED) || (chunk.State == persist.ChunkState_SUCCESS) {
		response.Restart = false
	} else {
		response.Restart = true
	}
	return response, nil
}

func (a *apiServer) FinishPod(ctx context.Context, request *ppsserver.FinishPodRequest) (response *ppsserver.FinishPodResponse, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "FinishJob")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	persistClient, err := a.getPersistClient()
	if err != nil {
		return nil, err
	}
	var chunk *persist.Chunk
	if request.Success {
		chunk, err = persistClient.FinishChunk(ctx, &persist.FinishChunkRequest{
			ChunkID: request.ChunkID,
			PodName: request.PodName,
		})
		if err != nil {
			return nil, err
		}
	} else {
		chunk, err = persistClient.RevokeChunk(ctx, &persist.RevokeChunkRequest{
			ChunkID: request.ChunkID,
			PodName: request.PodName,
			MaxPods: MaxPodsPerChunk,
		})
		if err != nil {
			return nil, err
		}
	}

	pfsAPIClient, err := a.getPfsClient()
	if err != nil {
		return nil, err
	}
	if _, err := pfsAPIClient.FinishCommit(ctx, &pfsclient.FinishCommitRequest{
		Commit: chunk.Pods[len(chunk.Pods)-1].OutputCommit,
		Cancel: !request.Success,
	}); err != nil {
		return nil, err
	}

	// Restart the pod if it fails, and the chunk has not exceeded the maximum
	// number of retries (at which point its state will be set to FAILED)
	response = &ppsserver.FinishPodResponse{
		Restart: !request.Success && chunk.State != persist.ChunkState_FAILED,
	}

	return response, nil
}

func (a *apiServer) CreatePipeline(ctx context.Context, request *ppsclient.CreatePipelineRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "CreatePipeline")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	pfsAPIClient, err := a.getPfsClient()
	if err != nil {
		return nil, err
	}

	persistClient, err := a.getPersistClient()
	if err != nil {
		return nil, err
	}

	setDefaultPipelineInputMethod(request.Inputs)

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
	var provenance []*pfsclient.Repo
	for _, input := range request.Inputs {
		provenance = append(provenance, input.Repo)
	}
	if !request.Update { // repo exists if it's an update
		// This function needs to return newErrPipelineExists if the pipeline
		// already exists
		if _, err := a.InspectPipeline(
			ctx,
			&ppsclient.InspectPipelineRequest{Pipeline: request.Pipeline},
		); err == nil {
			return nil, newErrPipelineExists(request.Pipeline.Name)
		}
		if _, err := pfsAPIClient.CreateRepo(
			ctx,
			&pfsclient.CreateRepoRequest{
				Repo:       repo,
				Provenance: provenance,
			}); err != nil {
			return nil, err
		}
		defer func() {
			if retErr != nil {
				// we don't return the error here because the function has
				// already errored, if this fails there's nothing we can do but
				// log it
				if _, err := pfsAPIClient.DeleteRepo(ctx, &pfsclient.DeleteRepoRequest{Repo: repo}); err != nil {
					protolion.Errorf("error deleting repo %s: %s", repo, err.Error())
				}
			}
		}()
	}

	persistPipelineInfo := &persist.PipelineInfo{
		PipelineName:    request.Pipeline.Name,
		Transform:       request.Transform,
		ParallelismSpec: request.ParallelismSpec,
		Inputs:          request.Inputs,
		OutputRepo:      repo,
		Shard:           a.hasher.HashPipeline(request.Pipeline),
		State:           ppsclient.PipelineState_PIPELINE_IDLE,
		GcPolicy:        request.GcPolicy,
		Output:          request.Output,
	}
	if persistPipelineInfo.GcPolicy == nil {
		persistPipelineInfo.GcPolicy = DefaultGCPolicy
	}

	if !request.Update {
		if _, err := persistClient.CreatePipelineInfo(ctx, persistPipelineInfo); err != nil {
			if strings.Contains(err.Error(), "Duplicate primary key `PipelineName`") {
				return nil, fmt.Errorf("pipeline %v already exists", request.Pipeline.Name)
			}
		}
	} else {
		if !request.NoArchive {
			if _, err := a.StopPipeline(ctx, &ppsclient.StopPipelineRequest{Pipeline: request.Pipeline}); err != nil {
				return nil, err
			}
			// archive the existing commits from the pipeline
			commitInfos, err := pfsAPIClient.ListCommit(
				ctx,
				&pfsclient.ListCommitRequest{
					Include: []*pfsclient.Commit{&pfsclient.Commit{
						Repo: ppsserver.PipelineRepo(request.Pipeline),
					}},
				})
			if err != nil {
				return nil, err
			}
			var commits []*pfsclient.Commit
			for _, commitInfo := range commitInfos.CommitInfo {
				commits = append(commits, commitInfo.Commit)
			}
			_, err = pfsAPIClient.ArchiveCommit(
				ctx,
				&pfsclient.ArchiveCommitRequest{
					Commits: commits,
				})
			if err != nil {
				return nil, err
			}
		}
		if _, err := persistClient.UpdatePipelineInfo(ctx, persistPipelineInfo); err != nil {
			return nil, err
		}
		if !request.NoArchive {
			// Downstream pipelines need to be restarted as well so that they know
			// there's new stuff to process.
			repoInfos, err := pfsAPIClient.ListRepo(
				ctx,
				&pfsclient.ListRepoRequest{
					Provenance: []*pfsclient.Repo{client.NewRepo(request.Pipeline.Name)},
				})
			if err != nil {
				return nil, err
			}
			var eg errgroup.Group
			for _, repoInfo := range repoInfos.RepoInfo {
				repoInfo := repoInfo
				eg.Go(func() error {
					// here we use the fact that pipelines have the same names as their output repos
					request := &persist.UpdatePipelineStoppedRequest{PipelineName: repoInfo.Repo.Name}
					request.Stopped = true
					if _, err := persistClient.UpdatePipelineStopped(ctx, request); err != nil {
						return err
					}
					request.Stopped = false
					if _, err := persistClient.UpdatePipelineStopped(ctx, request); err != nil {
						return err
					}
					return nil
				})
			}
			if err := eg.Wait(); err != nil {
				return nil, err
			}
		}
	}
	return &types.Empty{}, nil
}

// setDefaultPipelineInputMethod sets method to the default for the inputs
// that do not specify a method
func setDefaultPipelineInputMethod(inputs []*ppsclient.PipelineInput) {
	for _, input := range inputs {
		if input.Method == nil {
			input.Method = client.DefaultMethod
		}
	}
}

// setDefaultJobInputMethod sets method to the default for the inputs
// that do not specify a method
func setDefaultJobInputMethod(inputs []*ppsclient.JobInput) {
	for _, input := range inputs {
		if input.Method == nil {
			input.Method = client.DefaultMethod
		}
	}
}

func (a *apiServer) InspectPipeline(ctx context.Context, request *ppsclient.InspectPipelineRequest) (response *ppsclient.PipelineInfo, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "InspectPipeline")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

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

func (a *apiServer) ListPipeline(ctx context.Context, request *ppsclient.ListPipelineRequest) (response *ppsclient.PipelineInfos, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "ListPipeline")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

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

func (a *apiServer) DeletePipeline(ctx context.Context, request *ppsclient.DeletePipelineRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "DeletePipeline")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	if err := a.deletePipeline(ctx, request.Pipeline); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

func (a *apiServer) StartPipeline(ctx context.Context, request *ppsclient.StartPipelineRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "StartPipeline")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	persistClient, err := a.getPersistClient()
	if err != nil {
		return nil, err
	}
	_, err = persistClient.UpdatePipelineStopped(ctx, &persist.UpdatePipelineStoppedRequest{
		PipelineName: request.Pipeline.Name,
		Stopped:      false,
	})
	if err != nil {
		return nil, err
	}
	return persistClient.BlockPipelineState(ctx, &persist.BlockPipelineStateRequest{
		PipelineName: request.Pipeline.Name,
		State:        ppsclient.PipelineState_PIPELINE_RUNNING,
	})
}

func (a *apiServer) StopPipeline(ctx context.Context, request *ppsclient.StopPipelineRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "StopPipeline")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	persistClient, err := a.getPersistClient()
	if err != nil {
		return nil, err
	}
	_, err = persistClient.UpdatePipelineStopped(ctx, &persist.UpdatePipelineStoppedRequest{
		PipelineName: request.Pipeline.Name,
		Stopped:      true,
	})
	if err != nil {
		return nil, err
	}
	return persistClient.BlockPipelineState(ctx, &persist.BlockPipelineStateRequest{
		PipelineName: request.Pipeline.Name,
		State:        ppsclient.PipelineState_PIPELINE_STOPPED,
	})
}

func (a *apiServer) RerunPipeline(ctx context.Context, request *ppsclient.RerunPipelineRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "RerunPipeline")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	if _, err := a.StopPipeline(ctx, &ppsclient.StopPipelineRequest{Pipeline: request.Pipeline}); err != nil {
		return nil, err
	}
	defer func() {
		if _, err := a.StartPipeline(ctx, &ppsclient.StartPipelineRequest{Pipeline: request.Pipeline}); err != nil && retErr == nil {
			retErr = err
		}
	}()
	// archive the commits the user wants to rerun
	pfsAPIClient, err := a.getPfsClient()
	if err != nil {
		return nil, err
	}
	for _, commit := range append(request.Include, request.Exclude...) {
		if commit.Repo.Name != ppsserver.PipelineRepo(request.Pipeline).Name {
			return nil, fmt.Errorf("commit %s/%s must be on repo %s",
				commit.Repo.Name, commit.ID, ppsserver.PipelineRepo(request.Pipeline).Name)
		}
	}
	commitInfos, err := pfsAPIClient.ListCommit(
		ctx,
		&pfsclient.ListCommitRequest{
			Include: request.Include,
			Exclude: request.Exclude,
		})
	if err != nil {
		return nil, err
	}
	var commits []*pfsclient.Commit
	for _, commitInfo := range commitInfos.CommitInfo {
		commits = append(commits, commitInfo.Commit)
	}
	if _, err := pfsAPIClient.ArchiveCommit(
		ctx,
		&pfsclient.ArchiveCommitRequest{
			Commits: commits,
		},
	); err != nil {
		return nil, err
	}

	persistClient, err := a.getPersistClient()
	if err != nil {
		return nil, err
	}
	if _, err := persistClient.TouchPipelineInfo(ctx, request.Pipeline); err != nil {
		return nil, err
	}

	return &types.Empty{}, nil
}

func (a *apiServer) DeleteAll(ctx context.Context, request *types.Empty) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "PPSDeleteAll")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	persistClient, err := a.getPersistClient()
	if err != nil {
		return nil, err
	}
	persistPipelineInfos, err := persistClient.ListPipelineInfos(ctx, &persist.ListPipelineInfosRequest{})
	if err != nil {
		return nil, err
	}
	for _, persistPipelineInfo := range persistPipelineInfos.PipelineInfo {
		if err := a.deletePipeline(ctx, client.NewPipeline(persistPipelineInfo.PipelineName)); err != nil {
			return nil, err
		}
	}
	if _, err := persistClient.DeleteAll(ctx, request); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

func (a *apiServer) Version(version int64) error {
	a.versionLock.Lock()
	defer a.versionLock.Unlock()
	a.version = version
	return nil
}

func (a *apiServer) newPipelineCtx(ctx context.Context, pipelineName string) context.Context {
	ctx, cancel := context.WithCancel(ctx)
	a.pipelineCancelFuncsLock.Lock()
	defer a.pipelineCancelFuncsLock.Unlock()
	if oldCancel, ok := a.pipelineCancelFuncs[pipelineName]; ok {
		oldCancel()
	}
	a.pipelineCancelFuncs[pipelineName] = cancel
	return ctx
}

func (a *apiServer) cancelPipeline(pipelineName string) error {
	a.pipelineCancelFuncsLock.Lock()
	defer a.pipelineCancelFuncsLock.Unlock()
	cancel, ok := a.pipelineCancelFuncs[pipelineName]
	if ok {
		cancel()
		delete(a.pipelineCancelFuncs, pipelineName)
	} else {
		return fmt.Errorf("trying to cancel a pipeline %s which has not been started; this is likely a bug", pipelineName)
	}
	return nil
}

func (a *apiServer) newJobCtx(ctx context.Context, jobID string) context.Context {
	ctx, cancel := context.WithCancel(ctx)
	a.jobCancelFuncsLock.Lock()
	defer a.jobCancelFuncsLock.Unlock()
	if oldCancel, ok := a.jobCancelFuncs[jobID]; ok {
		oldCancel()
	}
	a.jobCancelFuncs[jobID] = cancel
	return ctx
}

func (a *apiServer) cancelJob(jobID string) error {
	a.jobCancelFuncsLock.Lock()
	defer a.jobCancelFuncsLock.Unlock()
	cancel, ok := a.jobCancelFuncs[jobID]
	if ok {
		cancel()
		delete(a.jobCancelFuncs, jobID)
	} else {
		return fmt.Errorf("trying to cancel a job %s which has not been started; this is likely a bug", jobID)
	}
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

	go func() {
		b := backoff.NewExponentialBackOff()
		// never stop retrying
		b.MaxElapsedTime = 0
		backoff.RetryNotify(func() error {
			pipelineClient, err := persistClient.SubscribePipelineInfos(ctx, &persist.SubscribePipelineInfosRequest{
				IncludeInitial: true,
				Shard:          &persist.Shard{Number: shard},
			})
			if err != nil {
				return err
			}
			for {
				pipelineChange, err := pipelineClient.Recv()
				if err != nil {
					return err
				}
				pipelineName := pipelineChange.Pipeline.PipelineName

				switch pipelineChange.Type {
				case persist.ChangeType_DELETE:
					if err := a.cancelPipeline(pipelineName); err != nil {
						protolion.Errorf("error cancelling pipeline %v: %s", pipelineName, err.Error())
					}
				case persist.ChangeType_UPDATE:
					if err := a.cancelPipeline(pipelineName); err != nil {
						protolion.Errorf("error cancelling pipeline %v: %s", pipelineName, err.Error())
					}
					fallthrough
				case persist.ChangeType_CREATE:
					pipelineCtx := a.newPipelineCtx(ctx, pipelineName)
					if pipelineChange.Pipeline.Stopped {
						if _, err = persistClient.UpdatePipelineState(pipelineCtx, &persist.UpdatePipelineStateRequest{
							PipelineName: pipelineName,
							State:        ppsclient.PipelineState_PIPELINE_STOPPED,
						}); err != nil {
							return err
						}
						continue
					}
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
						err = backoff.RetryNotify(func() error {
							if err := a.runPipeline(pipelineCtx, newPipelineInfo(pipelineChange.Pipeline)); err != nil && !isContextCancelled(err) {
								return err
							}
							return nil
						}, b, func(err error, d time.Duration) {
							protolion.Errorf("error running pipeline %v: %v; retrying in %s", pipelineName, err, d)
							if _, err = persistClient.UpdatePipelineState(pipelineCtx, &persist.UpdatePipelineStateRequest{
								PipelineName: pipelineName,
								State:        ppsclient.PipelineState_PIPELINE_RESTARTING,
								RecentError:  err.Error(),
							}); err != nil {
								protolion.Errorf("error updating pipeline state: %v", err)
							}
						})
						// At this point we stop retrying and update the pipeline state
						// to FAILED
						if err != nil {
							if _, err = persistClient.UpdatePipelineState(pipelineCtx, &persist.UpdatePipelineStateRequest{
								PipelineName: pipelineName,
								State:        ppsclient.PipelineState_PIPELINE_FAILURE,
								RecentError:  err.Error(),
							}); err != nil {
								protolion.Errorf("error updating pipeline state: %v", err)
							}
						}
					}()
				}
			}
		}, b, func(err error, d time.Duration) {
			protolion.Errorf("error receiving pipeline updates: %v; retrying in %v", err, d)
		})
	}()

	go func() {
		b := backoff.NewExponentialBackOff()
		// never stop retrying
		b.MaxElapsedTime = 0
		backoff.RetryNotify(func() error {
			jobInfoClient, err := persistClient.SubscribeJobInfos(ctx, &persist.SubscribeJobInfosRequest{
				IncludeInitial: true,
				IncludeChanges: false,
				Shard:          &persist.Shard{Number: shard},
				State:          []ppsclient.JobState{ppsclient.JobState_JOB_RUNNING},
			})
			if err != nil {
				return err
			}
			for {
				jobChange, err := jobInfoClient.Recv()
				if err != nil {
					return fmt.Errorf("error from receive: %s", err.Error())
				}
				jobID := jobChange.JobInfo.JobID

				switch jobChange.Type {
				case persist.ChangeType_DELETE:
					if err := a.cancelJob(jobID); err != nil {
						return fmt.Errorf("error cancelling job %v: %s", jobID, err.Error())
					}
				case persist.ChangeType_CREATE:
					// If we see a job that's running or creating, we start a job
					// manager for it.
					if jobChange.JobInfo.State == ppsclient.JobState_JOB_RUNNING || jobChange.JobInfo.State == ppsclient.JobState_JOB_CREATING {
						jobCtx := a.newJobCtx(ctx, jobID)
						go func() {
							b := backoff.NewExponentialBackOff()
							b.MaxElapsedTime = 0
							backoff.RetryNotify(func() error {
								if err := a.jobManager(jobCtx, &ppsclient.Job{jobID}); err != nil && !isContextCancelled(err) {
									return err
								}
								return nil
							}, b, func(err error, d time.Duration) {
								protolion.Errorf("error running jobManager for job %v: %v; retrying in %s", jobID, err, d)
							})
						}()
					}
				}
			}
		}, b, func(err error, d time.Duration) {
			protolion.Errorf("error receiving job updates: %v; retrying in %v", err, d)
		})
	}()

	return nil
}

func isContextCancelled(err error) bool {
	return err == context.Canceled || strings.Contains(err.Error(), context.Canceled.Error())
}

func (a *apiServer) DeleteShard(shard uint64) error {
	a.shardCancelFuncsLock.Lock()
	defer a.shardCancelFuncsLock.Unlock()
	cancel, ok := a.shardCancelFuncs[shard]
	if !ok {
		return fmt.Errorf("shard %d is being deleted, but it was never added; this is likely a bug", shard)
	}
	cancel()
	delete(a.shardCancelFuncs, shard)

	return nil
}

func newPipelineInfo(persistPipelineInfo *persist.PipelineInfo) *ppsclient.PipelineInfo {
	return &ppsclient.PipelineInfo{
		Pipeline: &ppsclient.Pipeline{
			Name: persistPipelineInfo.PipelineName,
		},
		Version:         persistPipelineInfo.Version,
		Transform:       persistPipelineInfo.Transform,
		ParallelismSpec: persistPipelineInfo.ParallelismSpec,
		Inputs:          persistPipelineInfo.Inputs,
		OutputRepo:      persistPipelineInfo.OutputRepo,
		CreatedAt:       persistPipelineInfo.CreatedAt,
		State:           persistPipelineInfo.State,
		RecentError:     persistPipelineInfo.RecentError,
		JobCounts:       persistPipelineInfo.JobCounts,
		GcPolicy:        persistPipelineInfo.GcPolicy,
		Output:          persistPipelineInfo.Output,
	}
}

func isCommitCancelledErr(err error) bool {
	return strings.Contains(err.Error(), "cancelled")
}

func (a *apiServer) runPipeline(ctx context.Context, pipelineInfo *ppsclient.PipelineInfo) error {
	if len(pipelineInfo.Inputs) == 0 {
		// this pipeline does not have inputs; there is nothing to be done
		return nil
	}

	persistClient, err := a.getPersistClient()
	if err != nil {
		return err
	}

	_, err = persistClient.UpdatePipelineState(ctx, &persist.UpdatePipelineStateRequest{
		PipelineName: pipelineInfo.Pipeline.Name,
		State:        ppsclient.PipelineState_PIPELINE_RUNNING,
	})
	if err != nil {
		return err
	}

	gcCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	go a.runGC(gcCtx, pipelineInfo)

	repoToLeaves := make(map[string]map[string]bool)
	rawInputRepos, err := a.rawInputs(ctx, pipelineInfo)
	if err != nil {
		return err
	}
	for _, repo := range rawInputRepos {
		repoToLeaves[repo.Name] = make(map[string]bool)
	}
	pfsAPIClient, err := a.getPfsClient()
	if err != nil {
		return err
	}
	for {
		var exclude []*pfsclient.Commit
		for repo, leaves := range repoToLeaves {
			if len(leaves) > 0 {
				for leaf := range leaves {
					exclude = append(
						exclude,
						&pfsclient.Commit{
							Repo: &pfsclient.Repo{Name: repo},
							ID:   leaf,
						})
				}
			} else {
				exclude = append(
					exclude,
					&pfsclient.Commit{
						Repo: &pfsclient.Repo{Name: repo},
					})
			}
		}
		listCommitRequest := &pfsclient.ListCommitRequest{
			Exclude:    exclude,
			CommitType: pfsclient.CommitType_COMMIT_TYPE_READ,
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
						newCommitSet[len(commitSet)] = client.NewCommit(repoName, leaf)
						newCommitSets = append(newCommitSets, newCommitSet)
					}
				}
				commitSets = newCommitSets
			}
			for _, commitSet := range commitSets {
				rawInputs := append(commitSet, commitInfo.Commit)
				// + 1 as the commitSet doesn't contain the commit we just got
				if len(rawInputs) < len(rawInputRepos) {
					continue
				}
				trueInputs, err := a.trueInputs(ctx, rawInputs, pipelineInfo)
				if err != nil {
					if isCommitCancelledErr(err) {
						protolion.Errorf("could not process raw commit set (%v) due to commit cancellation: %s", rawInputs, err)
						continue
					}
					return err
				}
				var parentJob *ppsclient.Job
				if commitInfo.ParentCommit != nil {
					parentRawInputs := append(commitSet, commitInfo.ParentCommit)
					parentJob, err = a.parentJob(ctx, trueInputs, parentRawInputs, pipelineInfo)
					if err != nil {
						return err
					}
				}
				_, err = a.CreateJob(
					ctx,
					&ppsclient.CreateJobRequest{
						Transform:       pipelineInfo.Transform,
						Pipeline:        pipelineInfo.Pipeline,
						ParallelismSpec: pipelineInfo.ParallelismSpec,
						Inputs:          trueInputs,
						ParentJob:       parentJob,
						Output:          pipelineInfo.Output,
					},
				)
				if err != nil {
					return err
				}
			}
		}
	}
}

// jobManager manages a job.  Specifically, it subscribes to status updates of
// chunks (units of work that make up of a job) and updates the state of the job
// as chunks are being finished.  It also squashes all the output commits of
// the chunks into the final output commit.
func (a *apiServer) jobManager(ctx context.Context, job *ppsclient.Job) error {
	persistClient, err := a.getPersistClient()
	if err != nil {
		return err
	}

	chunkClient, err := persistClient.SubscribeChunks(ctx, &persist.SubscribeChunksRequest{
		Job:            job,
		IncludeInitial: true,
	})
	if err != nil {
		return err
	}

	// a set that stores the chunk IDs that we've seen
	var podCommits []*pfsclient.Commit
	var failed bool
	var totalChunks int
	// ready is used to indicate if we've already received all chunks.
	// This is to prevent a scenario where we receive 1 chunk, see that
	// it's a SUCCESS, and be like, oh we are done!, without knowing that
	// there are more chunks to come, some of which might have not been
	// finished.
	var ready bool
	lm := lease.NewLeaser()
	for {
		chunkChange, err := chunkClient.Recv()
		if err != nil {
			return err
		}

		ready = ready || chunkChange.Ready

		chunk := chunkChange.Chunk
		if chunk != nil {
			switch chunkChange.Type {
			case persist.ChangeType_DELETE:
				totalChunks--
			case persist.ChangeType_CREATE, persist.ChangeType_UPDATE:
				if chunkChange.Type == persist.ChangeType_CREATE {
					totalChunks++
				}
				switch chunk.State {
				case persist.ChunkState_SUCCESS:
					lm.Return(chunk.ID)
					podCommits = append(podCommits, chunk.Pods[len(chunk.Pods)-1].OutputCommit)
				case persist.ChunkState_FAILED:
					lm.Return(chunk.ID)
					podCommits = append(podCommits, chunk.Pods[len(chunk.Pods)-1].OutputCommit)
					failed = true
				case persist.ChunkState_ASSIGNED:
					lm.Lease(chunk.ID, client.PPSLeasePeriod, func() {
						b := backoff.NewExponentialBackOff()
						b.MaxElapsedTime = 0
						backoff.Retry(func() error {
							if _, err := persistClient.RevokeChunk(ctx, &persist.RevokeChunkRequest{
								ChunkID: chunk.ID,
								PodName: chunk.Owner,
								MaxPods: MaxPodsPerChunk,
							}); err != nil && !isContextCancelled(err) {
								return err
							}
							return nil
						}, b)
					})
				}
			}
		}

		if ready && totalChunks == len(podCommits) {
			break
		}
	}

	jobInfo, err := persistClient.InspectJob(ctx, &ppsclient.InspectJobRequest{
		Job: job,
	})
	if err != nil {
		return err
	}

	pfsAPIClient, err := a.getPfsClient()
	if err != nil {
		return err
	}

	// Note that there's a failure mode that's not accounted for: if this process
	// fails after SquashCommit completes, but before CreateJobState completes,
	// then another process will attempt to run this jobManager again, causing
	// the pod commits to be squashed into the output commit again, meaning that
	// we might get duplicated data in the output commit.
	squashReq := &pfsclient.SquashCommitRequest{
		FromCommits: podCommits,
		ToCommit:    jobInfo.OutputCommit,
	}
	if _, err := pfsAPIClient.SquashCommit(
		ctx,
		squashReq,
	); err != nil {
		return err
	}

	if _, err = pfsAPIClient.FinishCommit(ctx, &pfsclient.FinishCommitRequest{
		Commit: jobInfo.OutputCommit,
		Cancel: failed,
	}); err != nil {
		return err
	}

	if !failed && jobInfo.Output != nil {
		// We use a new context here because as soon as we update the job state,
		// the original context will be cancelled.
		if _, err := persistClient.CreateJobState(context.Background(), &persist.JobState{
			JobID: job.ID,
			State: ppsclient.JobState_JOB_OUTPUTTING,
		}); err != nil {
			return err
		}
		pClient := client.APIClient{PfsAPIClient: pfsAPIClient}
		objClient, err := obj.NewClientFromURLAndSecret(context.Background(), jobInfo.Output.URL)
		if err != nil {
			return err
		}
		url, err := url.Parse(jobInfo.Output.URL)
		if err != nil {
			return err
		}
		if err := pfs_sync.PushObj(pClient, jobInfo.OutputCommit, objClient, url.Path); err != nil {
			return err
		}
	}

	var state ppsclient.JobState
	if failed {
		state = ppsclient.JobState_JOB_FAILURE
	} else {
		state = ppsclient.JobState_JOB_SUCCESS
	}
	// We use a new context here because as soon as we update the job state,
	// the original context will be cancelled.
	if _, err := persistClient.CreateJobState(context.Background(), &persist.JobState{
		JobID: job.ID,
		State: state,
	}); err != nil {
		return err
	}

	return nil
}

func (a *apiServer) parentJob(
	ctx context.Context,
	trueInputs []*ppsclient.JobInput,
	rawInputs []*pfsclient.Commit,
	pipelineInfo *ppsclient.PipelineInfo,
) (*ppsclient.Job, error) {
	parentTrueInputs, err := a.trueInputs(ctx, rawInputs, pipelineInfo)
	if err != nil {
		return nil, err
	}
	parental, err := inputsAreParental(trueInputs, parentTrueInputs)
	if err != nil {
		return nil, err
	}
	if !parental {
		return nil, nil
	}
	var parentTrueInputCommits []*pfsclient.Commit
	for _, input := range parentTrueInputs {
		parentTrueInputCommits = append(parentTrueInputCommits, input.Commit)
	}
	jobInfos, err := a.ListJob(
		ctx,
		&ppsclient.ListJobRequest{
			Pipeline:    pipelineInfo.Pipeline,
			InputCommit: parentTrueInputCommits,
		})
	if err != nil {
		return nil, err
	}
	for _, jobInfo := range jobInfos.JobInfo {
		if jobInfo.PipelineVersion == pipelineInfo.Version {
			return jobInfo.Job, nil
		}
	}
	return nil, nil
}

// inputsAreParental returns true if a job run from oldTrueInputs can be used
// as a parent for one run from newTrueInputs
func inputsAreParental(
	trueInputs []*ppsclient.JobInput,
	parentTrueInputs []*ppsclient.JobInput,
) (bool, error) {
	if len(trueInputs) != len(parentTrueInputs) {
		return false, fmt.Errorf("true inputs have different lengths (this is likely a bug)")
	}
	sort.Sort(JobInputs(trueInputs))
	sort.Sort(JobInputs(parentTrueInputs))
	for i, trueInput := range trueInputs {
		parentTrueInput := parentTrueInputs[i]
		if trueInput.Commit.ID != parentTrueInput.Commit.ID &&
			trueInput.Method.Incremental == ppsclient.Incremental_NONE {
			return false, nil
		}
	}
	return true, nil
}

// rawInputs tracks provenance for a pipeline back to its raw sources of
// data
// rawInputs is much efficient less than it could be because it does a lot of
// duplicate work computing provenance. It could be made more efficient by
// adding a special purpose rpc to the pfs api but that call wouldn't be useful
// for much other than this.
func (a *apiServer) rawInputs(
	ctx context.Context,
	pipelineInfo *ppsclient.PipelineInfo,
) ([]*pfsclient.Repo, error) {
	pfsClient, err := a.getPfsClient()
	if err != nil {
		return nil, err
	}
	repoInfo, err := pfsClient.InspectRepo(
		ctx,
		&pfsclient.InspectRepoRequest{Repo: ppsserver.PipelineRepo(pipelineInfo.Pipeline)},
	)
	if err != nil {
		return nil, err
	}
	var result []*pfsclient.Repo
	for _, repo := range repoInfo.Provenance {
		repoInfo, err := pfsClient.InspectRepo(
			ctx,
			&pfsclient.InspectRepoRequest{Repo: repo},
		)
		if err != nil {
			return nil, err
		}
		if len(repoInfo.Provenance) == 0 {
			result = append(result, repoInfo.Repo)
		}
	}
	return result, nil
}

// trueInputs returns the JobInputs for a set of raw input commits
func (a *apiServer) trueInputs(
	ctx context.Context,
	rawInputs []*pfsclient.Commit,
	pipelineInfo *ppsclient.PipelineInfo,
) ([]*ppsclient.JobInput, error) {
	pfsClient, err := a.getPfsClient()
	if err != nil {
		return nil, err
	}
	var toRepo []*pfsclient.Repo
	repoToInput := make(map[string]*ppsclient.PipelineInput)
	for _, input := range pipelineInfo.Inputs {
		toRepo = append(toRepo, input.Repo)
		repoToInput[input.Repo.Name] = input
	}
	var result []*ppsclient.JobInput
	for _, commit := range rawInputs {
		pipelineInput, ok := repoToInput[commit.Repo.Name]
		if ok {
			result = append(result,
				&ppsclient.JobInput{
					Commit:   commit,
					Method:   pipelineInput.Method,
					RunEmpty: pipelineInput.RunEmpty,
					Lazy:     pipelineInput.Lazy,
				})
			delete(repoToInput, commit.Repo.Name)
		}
	}
	if len(result) == len(pipelineInfo.Inputs) {
		// our pipeline only has raw inputs
		// no need to flush them, we can return them as is
		return result, nil
	}
	// Flush the rawInputs up to true input repos of the pipeline
	commitInfos, err := pfsClient.FlushCommit(
		ctx,
		&pfsclient.FlushCommitRequest{
			Commit: rawInputs,
			ToRepo: toRepo,
		},
	)
	if err != nil {
		return nil, err
	}
	for _, commitInfo := range commitInfos.CommitInfo {
		pipelineInput, ok := repoToInput[commitInfo.Commit.Repo.Name]
		if ok {
			result = append(result,
				&ppsclient.JobInput{
					Commit:   commitInfo.Commit,
					Method:   pipelineInput.Method,
					RunEmpty: pipelineInput.RunEmpty,
				})
		}
	}
	return result, nil
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
		Job:             job,
		Transform:       persistJobInfo.Transform,
		Pipeline:        &ppsclient.Pipeline{Name: persistJobInfo.PipelineName},
		PipelineVersion: persistJobInfo.PipelineVersion,
		ParallelismSpec: persistJobInfo.ParallelismSpec,
		Inputs:          persistJobInfo.Inputs,
		ParentJob:       persistJobInfo.ParentJob,
		Started:         persistJobInfo.Started,
		Finished:        persistJobInfo.Finished,
		OutputCommit:    persistJobInfo.OutputCommit,
		State:           persistJobInfo.State,
		Service:         persistJobInfo.Service,
		Output:          persistJobInfo.Output,
	}, nil
}

// RepoNameToEnvString is a helper which uppercases a repo name for
// use in environment variable names.
func RepoNameToEnvString(repoName string) string {
	return strings.ToUpper(repoName)
}

type jobOptions struct {
	labels             map[string]string
	parallelism        int32
	userImage          string
	jobShimImage       string
	jobImagePullPolicy string
	jobEnv             []api.EnvVar
	volumes            []api.Volume
	volumeMounts       []api.VolumeMount
	imagePullSecrets   []api.LocalObjectReference
}

func getJobOptions(kubeClient *kube.Client, jobInfo *persist.JobInfo, jobShimImage string, jobImagePullPolicy string) (*jobOptions, error) {
	labels := labels(jobInfo.JobID)
	parallelism64, err := GetExpectedNumWorkers(kubeClient, jobInfo.ParallelismSpec)
	if err != nil {
		return nil, err
	}
	parallelism := int32(parallelism64)
	userImage := jobInfo.Transform.Image
	if userImage == "" {
		userImage = DefaultUserImage
	}
	if jobImagePullPolicy == "" {
		jobImagePullPolicy = "IfNotPresent"
	}

	var jobEnv []api.EnvVar
	for _, input := range jobInfo.Inputs {
		jobEnv = append(
			jobEnv,
			api.EnvVar{
				Name:  fmt.Sprintf("PACH_%v_COMMIT_ID", RepoNameToEnvString(input.Commit.Repo.Name)),
				Value: input.Commit.ID,
			},
		)
	}
	for name, value := range jobInfo.Transform.Env {
		jobEnv = append(
			jobEnv,
			api.EnvVar{
				Name:  name,
				Value: value,
			},
		)
	}
	// We use Kubernetes' "Downward API" so the pod is aware of its name.
	// This is so that the pod can include its name in future requests
	// to PPS.
	// http://kubernetes.io/docs/user-guide/downward-api/
	jobEnv = append(jobEnv, api.EnvVar{
		Name: client.PPSPodNameEnv,
		ValueFrom: &api.EnvVarSource{
			FieldRef: &api.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath:  "metadata.name",
			},
		},
	})

	var volumes []api.Volume
	var volumeMounts []api.VolumeMount
	for _, secret := range jobInfo.Transform.Secrets {
		volumes = append(volumes, api.Volume{
			Name: secret.Name,
			VolumeSource: api.VolumeSource{
				Secret: &api.SecretVolumeSource{
					SecretName: secret.Name,
				},
			},
		})
		volumeMounts = append(volumeMounts, api.VolumeMount{
			Name:      secret.Name,
			MountPath: secret.MountPath,
		})
	}

	volumes = append(volumes, api.Volume{
		Name: "pach-bin",
		VolumeSource: api.VolumeSource{
			EmptyDir: &api.EmptyDirVolumeSource{},
		},
	})
	volumeMounts = append(volumeMounts, api.VolumeMount{
		Name:      "pach-bin",
		MountPath: "/pach-bin",
	})
	var imagePullSecrets []api.LocalObjectReference
	for _, secret := range jobInfo.Transform.ImagePullSecrets {
		imagePullSecrets = append(imagePullSecrets, api.LocalObjectReference{Name: secret})
	}

	return &jobOptions{
		labels:             labels,
		parallelism:        parallelism,
		userImage:          userImage,
		jobShimImage:       jobShimImage,
		jobImagePullPolicy: jobImagePullPolicy,
		jobEnv:             jobEnv,
		volumes:            volumes,
		volumeMounts:       volumeMounts,
		imagePullSecrets:   imagePullSecrets,
	}, nil
}

func podSpec(options *jobOptions, jobID string, restartPolicy api.RestartPolicy) api.PodSpec {
	return api.PodSpec{
		InitContainers: []api.Container{
			{
				Name:            "init",
				Image:           options.jobShimImage,
				Command:         []string{"/pach/job-shim.sh"},
				ImagePullPolicy: api.PullPolicy(options.jobImagePullPolicy),
				Env:             options.jobEnv,
				VolumeMounts:    options.volumeMounts,
			},
		},
		Containers: []api.Container{
			{
				Name:    "user",
				Image:   options.userImage,
				Command: []string{"/pach-bin/guest.sh", jobID},
				SecurityContext: &api.SecurityContext{
					Privileged: &trueVal, // god is this dumb
				},
				ImagePullPolicy: api.PullPolicy(options.jobImagePullPolicy),
				Env:             options.jobEnv,
				VolumeMounts:    options.volumeMounts,
			},
		},
		RestartPolicy:    restartPolicy,
		Volumes:          options.volumes,
		ImagePullSecrets: options.imagePullSecrets,
	}
}

func serviceName(jobID string) string {
	return fmt.Sprintf("pach-%v", jobID)
}

func service(kubeClient *kube.Client, jobInfo *persist.JobInfo, jobShimImage string, jobImagePullPolicy string, internalPort int32, externalPort int32) (*api.ReplicationController, *api.Service, error) {

	options, err := getJobOptions(kubeClient, jobInfo, jobShimImage, jobImagePullPolicy)
	if err != nil {
		return nil, nil, err
	}
	if options.parallelism != int32(1) {
		return nil, nil, fmt.Errorf("pachyderm service only supports parallelism of 1, got %v", options.parallelism)
	}
	rc := &api.ReplicationController{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "ReplicationController",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   jobInfo.JobID,
			Labels: options.labels,
		},
		Spec: api.ReplicationControllerSpec{
			Selector: options.labels,
			Replicas: options.parallelism,
			Template: &api.PodTemplateSpec{
				ObjectMeta: api.ObjectMeta{
					Name:   jobInfo.JobID,
					Labels: options.labels,
				},
				Spec: podSpec(options, jobInfo.JobID, "Always"),
			},
		},
	}
	service := &api.Service{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   serviceName(jobInfo.JobID),
			Labels: options.labels,
		},
		Spec: api.ServiceSpec{
			Type:     api.ServiceTypeNodePort,
			Selector: options.labels,
			Ports: []api.ServicePort{
				{
					Port:     internalPort,
					Name:     "user-service-port",
					NodePort: externalPort,
				},
			},
		},
	}
	return rc, service, nil
}

// Convert a persist.JobInfo into a Kubernetes batch.Job spec
func job(kubeClient *kube.Client, jobInfo *persist.JobInfo, jobShimImage string, jobImagePullPolicy string) (*batch.Job, error) {
	options, err := getJobOptions(kubeClient, jobInfo, jobShimImage, jobImagePullPolicy)
	if err != nil {
		return nil, err
	}
	return &batch.Job{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Job",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   jobInfo.JobID,
			Labels: options.labels,
		},
		Spec: batch.JobSpec{
			ManualSelector: &trueVal,
			Selector: &unversioned.LabelSelector{
				MatchLabels: options.labels,
			},
			Parallelism: &options.parallelism,
			Completions: &options.parallelism,
			Template: api.PodTemplateSpec{
				ObjectMeta: api.ObjectMeta{
					Name:   jobInfo.JobID,
					Labels: options.labels,
				},
				Spec: podSpec(options, jobInfo.JobID, "Never"),
			},
		},
	}, nil
}

func (a *apiServer) jobPods(job *ppsclient.Job) ([]api.Pod, error) {
	podList, err := a.kubeClient.Pods(a.namespace).List(api.ListOptions{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "ListOptions",
			APIVersion: "v1",
		},
		LabelSelector: kube_labels.SelectorFromSet(labels(job.ID)),
	})
	if err != nil {
		return nil, err
	}
	return podList.Items, nil
}

func (a *apiServer) deleteJob(ctx context.Context, jobInfo *persist.JobInfo) error {
	falseVal := false
	deleteOptions := &api.DeleteOptions{
		OrphanDependents: &falseVal,
	}
	if jobInfo.Service != nil {
		if err := a.kubeClient.ReplicationControllers(a.namespace).Delete(jobInfo.JobID, deleteOptions); err != nil {
			return err
		}
		if err := a.kubeClient.Services(a.namespace).Delete(serviceName(jobInfo.JobID)); err != nil {
			return err
		}
	} else {
		if err := a.kubeClient.Extensions().Jobs(a.namespace).Delete(jobInfo.JobID, nil); err != nil {
			// we don't return on failure here because jobs may get deleted
			// through other means and we don't want that to prevent users from
			// deleting pipelines.
			protolion.Errorf("error deleting job %s: %s", jobInfo.JobID, err.Error())
		}
		pods, jobPodsErr := a.jobPods(client.NewJob(jobInfo.JobID))
		for _, pod := range pods {
			if err := a.kubeClient.Pods(a.namespace).Delete(pod.Name, deleteOptions); err != nil {
				// we don't return on failure here because pods may get deleted
				// through other means and we don't want that to prevent users from
				// deleting pipelines.
				protolion.Errorf("error deleting pod %s: %s", pod.Name, err.Error())
			}
		}
		if jobPodsErr != nil {
			return jobPodsErr
		}
	}
	persistClient, err := a.getPersistClient()
	if err != nil {
		return err
	}
	// Remove the chunks for this job
	_, err = persistClient.DeleteChunksForJob(ctx, &ppsclient.Job{
		ID: jobInfo.JobID,
	})
	return err
}

func (a *apiServer) deletePipeline(ctx context.Context, pipeline *ppsclient.Pipeline) error {
	persistClient, err := a.getPersistClient()
	if err != nil {
		return err
	}

	if pipeline == nil {
		return fmt.Errorf("pipeline cannot be nil")
	}

	// Delete kubernetes jobs.  Otherwise we won't be able to create jobs with
	// the same IDs, since kubernetes jobs simply use these IDs as their names
	jobInfos, err := persistClient.ListJobInfos(ctx, &ppsclient.ListJobRequest{
		Pipeline: pipeline,
	})
	if err != nil {
		return err
	}
	var eg errgroup.Group
	for _, jobInfo := range jobInfos.JobInfo {
		jobInfo := jobInfo
		eg.Go(func() error { return a.deleteJob(ctx, jobInfo) })
	}
	if err := eg.Wait(); err != nil {
		return err
	}

	// The reason we need to do this, is that if we don't, then if the very same
	// pipeline is recreated, we won't actually create new jobs due to the fact
	// that we de-duplicate jobs by obtaining JobIDs through hashing pipeline
	// name + inputs.  So the jobs will already be in the database, resulting
	// in no new jobs being created, even though the output of those existing
	// jobs might have already being removed.
	// Therefore, we delete the job infos.
	if _, err := persistClient.DeleteJobInfosForPipeline(ctx, pipeline); err != nil {
		return err
	}

	if _, err := persistClient.DeletePipelineInfo(ctx, pipeline); err != nil {
		return err
	}

	return nil
}

func labels(app string) map[string]string {
	return map[string]string{
		"app":   app,
		"suite": suite,
	}
}

type podSlice []api.Pod

func (s podSlice) Len() int {
	return len(s)
}
func (s podSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s podSlice) Less(i, j int) bool {
	return s[i].ObjectMeta.Name < s[j].ObjectMeta.Name
}
