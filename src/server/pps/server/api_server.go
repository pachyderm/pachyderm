package server

import (
	"bufio"
	"bytes"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	client "github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
	"github.com/pachyderm/pachyderm/src/server/pkg/hashtree"
	"github.com/pachyderm/pachyderm/src/server/pkg/metrics"
	"github.com/pachyderm/pachyderm/src/server/pkg/watch"
	ppsserver "github.com/pachyderm/pachyderm/src/server/pps"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/gogo/protobuf/types"
	"go.pedge.io/lion/proto"
	"go.pedge.io/proto/rpclog"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
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

type ctxAndCancel struct {
	ctx    context.Context
	cancel context.CancelFunc
}

type apiServer struct {
	protorpclog.Logger
	etcdPrefix          string
	hasher              *ppsserver.Hasher
	address             string
	etcdClient          *etcd.Client
	pachConn            *grpc.ClientConn
	pachConnOnce        sync.Once
	kubeClient          *kube.Client
	shardLock           sync.RWMutex
	shardCtxs           map[uint64]*ctxAndCancel
	pipelineCancelsLock sync.Mutex
	pipelineCancels     map[string]context.CancelFunc
	jobCancelsLock      sync.Mutex
	jobCancels          map[string]context.CancelFunc
	workerPools         map[string]WorkerPool
	workerPoolsLock     sync.Mutex
	version             int64
	// versionLock protects the version field.
	// versionLock must be held BEFORE reading from version and UNTIL all
	// requests using version have returned
	versionLock           sync.RWMutex
	namespace             string
	workerImage           string
	workerImagePullPolicy string
	reporter              *metrics.Reporter
	// collections
	pipelines col.Collection
	jobs      col.Collection
}

type byInputName struct {
	inputs []*pps.PipelineInput
}

func (b byInputName) Len() int {
	return len(b.inputs)
}

func (b byInputName) Less(i, j int) bool {
	return b.inputs[i].Name < b.inputs[j].Name
}

func (b byInputName) Swap(i, j int) {
	b.inputs[i], b.inputs[j] = b.inputs[j], b.inputs[i]
}

// GetExpectedNumWorkers computes the expected number of workers that pachyderm will start given
// the ParallelismSpec 'spec'.
//
// This is only exported for testing
func GetExpectedNumWorkers(kubeClient *kube.Client, spec *pps.ParallelismSpec) (uint64, error) {
	coefficient := 0.0 // Used if [spec.Strategy == PROPORTIONAL] or [spec.Constant == 0]
	if spec == nil {
		// Unset ParallelismSpec is handled here. Currently we start one worker per
		// node
		coefficient = 1.0
	} else if spec.Strategy == pps.ParallelismSpec_CONSTANT {
		if spec.Constant > 0 {
			return spec.Constant, nil
		}
		// Zero-initialized ParallelismSpec is handled here. Currently we start one
		// worker per node
		coefficient = 1
	} else if spec.Strategy == pps.ParallelismSpec_COEFFICIENT {
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
		return 0, fmt.Errorf("unable to retrieve node list from k8s to determine parallelism: %v", err)
	}
	if len(nodeList.Items) == 0 {
		return 0, fmt.Errorf("pachyderm.pps.jobserver: no k8s nodes found")
	}
	result := math.Floor(coefficient * float64(len(nodeList.Items)))
	return uint64(math.Max(result, 1)), nil
}

func (a *apiServer) validateJob(ctx context.Context, jobInfo *pps.JobInfo) error {
	for _, in := range jobInfo.Inputs {
		switch {
		case len(in.Name) == 0:
			return fmt.Errorf("every job input must specify a name")
		case in.Name == "out":
			return fmt.Errorf("no job input may be named \"out\", as pachyderm " +
				"already creates /pfs/out to collect job output")
		case in.Commit == nil:
			return fmt.Errorf("every job input must specify a commit")
		}
		// check that the input commit exists
		pfsClient, err := a.getPFSClient()
		if err != nil {
			return err
		}
		_, err = pfsClient.InspectCommit(ctx, &pfs.InspectCommitRequest{
			Commit: in.Commit,
		})
		if err != nil {
			return fmt.Errorf("commit %s not found: %s", in.Commit.FullID(), err)
		}
	}
	return nil
}

func (a *apiServer) CreateJob(ctx context.Context, request *pps.CreateJobRequest) (response *pps.Job, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "CreateJob")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	job := &pps.Job{uuid.NewWithoutUnderscores()}
	_, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
		jobInfo := &pps.JobInfo{
			Job:             job,
			Transform:       request.Transform,
			Pipeline:        request.Pipeline,
			ParallelismSpec: request.ParallelismSpec,
			Inputs:          request.Inputs,
			OutputRepo:      request.OutputRepo,
			OutputBranch:    request.OutputBranch,
			Started:         now(),
			Finished:        nil,
			OutputCommit:    nil,
			Service:         request.Service,
			ParentJob:       request.ParentJob,
		}

		if request.Pipeline != nil {
			pipelineInfo := new(pps.PipelineInfo)
			if err := a.pipelines.ReadWrite(stm).Get(request.Pipeline.Name, pipelineInfo); err != nil {
				return err
			}
			jobInfo.PipelineVersion = pipelineInfo.Version
			jobInfo.PipelineID = pipelineInfo.ID
		}
		if err := a.validateJob(ctx, jobInfo); err != nil {
			return err
		}
		return a.updateJobState(stm, jobInfo, pps.JobState_JOB_STARTING)
	})
	if err != nil {
		return nil, err
	}
	return job, nil
}

func (a *apiServer) InspectJob(ctx context.Context, request *pps.InspectJobRequest) (response *pps.JobInfo, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "InspectJob")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	jobs := a.jobs.ReadOnly(ctx)

	if request.BlockState {
		watcher, err := jobs.WatchOne(request.Job.ID)
		if err != nil {
			return nil, err
		}
		defer watcher.Close()

		for {
			ev, ok := <-watcher.Watch()
			if !ok {
				return nil, fmt.Errorf("the stream for job updates closed unexpectedly")
			}
			switch ev.Type {
			case watch.EventError:
				return nil, ev.Err
			case watch.EventDelete:
				return nil, fmt.Errorf("job %s was deleted", request.Job.ID)
			case watch.EventPut:
				var jobID string
				var jobInfo pps.JobInfo
				if err := ev.Unmarshal(&jobID, &jobInfo); err != nil {
					return nil, err
				}
				if jobStateToStopped(jobInfo.State) {
					return &jobInfo, nil
				}
			}
		}
	}

	jobInfo := new(pps.JobInfo)
	if err := jobs.Get(request.Job.ID, jobInfo); err != nil {
		return nil, err
	}
	return jobInfo, nil
}

func (a *apiServer) ListJob(ctx context.Context, request *pps.ListJobRequest) (response *pps.JobInfos, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "ListJob")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	jobs := a.jobs.ReadOnly(ctx)
	var iter col.Iterator
	var err error
	if request.Pipeline != nil {
		iter, err = jobs.GetByIndex(jobsPipelineIndex, request.Pipeline)
		if err != nil {
			return nil, err
		}
	} else {
		iter, err = jobs.List()
	}

	var jobInfos []*pps.JobInfo
	for {
		var jobID string
		var jobInfo pps.JobInfo
		ok, err := iter.Next(&jobID, &jobInfo)
		if err != nil {
			return nil, err
		}
		if !ok {
			break
		}
		jobInfos = append(jobInfos, &jobInfo)
	}

	return &pps.JobInfos{jobInfos}, nil
}

func (a *apiServer) DeleteJob(ctx context.Context, request *pps.DeleteJobRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "DeleteJob")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	_, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
		return a.jobs.ReadWrite(stm).Delete(request.Job.ID)
	})
	if err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

func (a *apiServer) GetLogs(request *pps.GetLogsRequest, apiGetLogsServer pps.API_GetLogsServer) (retErr error) {
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

func (a *apiServer) validatePipeline(ctx context.Context, pipelineInfo *pps.PipelineInfo) error {
	names := make(map[string]bool)
	for _, in := range pipelineInfo.Inputs {
		switch {
		case len(in.Name) == 0:
			return fmt.Errorf("every pipeline input must specify a name")
		case in.Name == "out":
			return fmt.Errorf("no pipeline input may be named \"out\", as pachyderm " +
				"already creates /pfs/out to collect pipeline output")
		case len(in.Branch) == 0:
			return fmt.Errorf("every pipeline input must specify a branch")
		case len(in.Glob) == 0:
			return fmt.Errorf("every pipeline input must specify a glob")
		}
		// detect input name conflicts
		if names[in.Name] {
			return fmt.Errorf("conflicting input names: %s", in.Name)
		}
		names[in.Name] = true
		// check that the input repo exists
		pfsClient, err := a.getPFSClient()
		if err != nil {
			return err
		}
		_, err = pfsClient.InspectRepo(ctx, &pfs.InspectRepoRequest{
			Repo: in.Repo,
		})
		if err != nil {
			return fmt.Errorf("repo %s not found: %s", in.Repo.Name, err)
		}
	}
	if strings.Contains(pipelineInfo.Pipeline.Name, "_") {
		return fmt.Errorf("pipeline name %s may not contain underscore", pipelineInfo.Pipeline.Name)
	}
	if pipelineInfo.OutputBranch == "" {
		return fmt.Errorf("pipeline needs to specify an output branch")
	}
	return nil
}

func (a *apiServer) CreatePipeline(ctx context.Context, request *pps.CreatePipelineRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "CreatePipeline")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	sort.Sort(byInputName{request.Inputs})
	pipelineInfo := &pps.PipelineInfo{
		ID:              uuid.NewWithoutDashes(),
		Pipeline:        request.Pipeline,
		Transform:       request.Transform,
		ParallelismSpec: request.ParallelismSpec,
		Inputs:          request.Inputs,
		OutputBranch:    request.OutputBranch,
		GcPolicy:        request.GcPolicy,
		Egress:          request.Egress,
		CreatedAt:       now(),
	}
	setPipelineDefaults(pipelineInfo)
	if err := a.validatePipeline(ctx, pipelineInfo); err != nil {
		return nil, err
	}

	pfsClient, err := a.getPFSClient()
	if err != nil {
		return nil, err
	}

	pipelineName := pipelineInfo.Pipeline.Name

	if request.Update {
		if _, err := a.StopPipeline(ctx, &pps.StopPipelineRequest{request.Pipeline}); err != nil {
			return nil, err
		}
		var oldPipelineInfo pps.PipelineInfo
		_, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
			pipelines := a.pipelines.ReadWrite(stm)
			if err := pipelines.Get(pipelineName, &oldPipelineInfo); err != nil {
				return err
			}
			pipelineInfo.Version = oldPipelineInfo.Version + 1
			pipelines.Put(pipelineName, pipelineInfo)
			return nil
		})
		if err != nil {
			return nil, err
		}

		// Rename the original output branch to `outputBranch-vN`, where N
		// is the previous version number of the pipeline.
		if _, err := pfsClient.SetBranch(ctx, &pfs.SetBranchRequest{
			Commit: &pfs.Commit{
				Repo: &pfs.Repo{pipelineName},
				ID:   oldPipelineInfo.OutputBranch,
			},
			Branch: fmt.Sprintf("%s-v%d", oldPipelineInfo.OutputBranch, oldPipelineInfo.Version),
		}); err != nil {
			return nil, err
		}

		if _, err := pfsClient.DeleteBranch(ctx, &pfs.DeleteBranchRequest{
			Repo:   &pfs.Repo{pipelineName},
			Branch: oldPipelineInfo.OutputBranch,
		}); err != nil {
			return nil, err
		}

		if _, err := a.StartPipeline(ctx, &pps.StartPipelineRequest{request.Pipeline}); err != nil {
			return nil, err
		}
	} else {
		_, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
			pipelines := a.pipelines.ReadWrite(stm)
			err := pipelines.Create(pipelineName, pipelineInfo)
			if isAlreadyExistsErr(err) {
				return newErrPipelineExists(pipelineName)
			}
			return err
		})
		if err != nil {
			return nil, err
		}
	}

	// Create output repo
	// The pipeline manager also creates the output repo, but we want to
	// also create the repo here to make sure that the output repo is
	// guaranteed to be there after CreatePipeline returns.  This is
	// because it's a very common pattern to create many pipelines in a
	// row, some of which depend on the existence of the output repos
	// of upstream pipelines.
	var provenance []*pfs.Repo
	for _, input := range pipelineInfo.Inputs {
		provenance = append(provenance, input.Repo)
	}

	if _, err := pfsClient.CreateRepo(ctx, &pfs.CreateRepoRequest{
		Repo:       &pfs.Repo{pipelineInfo.Pipeline.Name},
		Provenance: provenance,
	}); err != nil && !isAlreadyExistsErr(err) {
		return nil, err
	}

	return &types.Empty{}, err
}

// setPipelineDefaults sets the default values for a pipeline info
func setPipelineDefaults(pipelineInfo *pps.PipelineInfo) {
	for _, input := range pipelineInfo.Inputs {
		// Input branches default to master
		if input.Branch == "" {
			input.Branch = "master"
		}
		if input.Name == "" {
			input.Name = input.Repo.Name
		}
	}
	if pipelineInfo.OutputBranch == "" {
		// Output branches default to master
		pipelineInfo.OutputBranch = "master"
	}
}

func (a *apiServer) InspectPipeline(ctx context.Context, request *pps.InspectPipelineRequest) (response *pps.PipelineInfo, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "InspectPipeline")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	pipelineInfo := new(pps.PipelineInfo)
	if err := a.pipelines.ReadOnly(ctx).Get(request.Pipeline.Name, pipelineInfo); err != nil {
		return nil, err
	}
	return pipelineInfo, nil
}

func (a *apiServer) ListPipeline(ctx context.Context, request *pps.ListPipelineRequest) (response *pps.PipelineInfos, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "ListPipeline")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	pipelineIter, err := a.pipelines.ReadOnly(ctx).List()
	if err != nil {
		return nil, err
	}

	pipelineInfos := new(pps.PipelineInfos)

	for {
		var pipelineName string
		pipelineInfo := new(pps.PipelineInfo)
		ok, err := pipelineIter.Next(&pipelineName, pipelineInfo)
		if err != nil {
			return nil, err
		}
		if ok {
			pipelineInfos.PipelineInfo = append(pipelineInfos.PipelineInfo, pipelineInfo)
		} else {
			break
		}
	}
	return pipelineInfos, nil
}

func (a *apiServer) DeletePipeline(ctx context.Context, request *pps.DeletePipelineRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "DeletePipeline")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	_, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
		return a.pipelines.ReadWrite(stm).Delete(request.Pipeline.Name)
	})
	if err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

func (a *apiServer) StartPipeline(ctx context.Context, request *pps.StartPipelineRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "StartPipeline")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	if err := a.updatePipelineState(ctx, request.Pipeline.Name, pps.PipelineState_PIPELINE_RUNNING); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

func (a *apiServer) StopPipeline(ctx context.Context, request *pps.StopPipelineRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "StopPipeline")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	if err := a.updatePipelineState(ctx, request.Pipeline.Name, pps.PipelineState_PIPELINE_STOPPED); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

func (a *apiServer) RerunPipeline(ctx context.Context, request *pps.RerunPipelineRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "RerunPipeline")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	return nil, fmt.Errorf("TODO")
}

func (a *apiServer) DeleteAll(ctx context.Context, request *types.Empty) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "PPSDeleteAll")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	pipelineInfos, err := a.ListPipeline(ctx, &pps.ListPipelineRequest{})
	if err != nil {
		return nil, err
	}

	for _, pipelineInfo := range pipelineInfos.PipelineInfo {
		if _, err := a.DeletePipeline(ctx, &pps.DeletePipelineRequest{pipelineInfo.Pipeline}); err != nil {
			return nil, err
		}
	}

	jobInfos, err := a.ListJob(ctx, &pps.ListJobRequest{})
	if err != nil {
		return nil, err
	}

	for _, jobInfo := range jobInfos.JobInfo {
		if _, err := a.DeleteJob(ctx, &pps.DeleteJobRequest{jobInfo.Job}); err != nil {
			return nil, err
		}
	}

	return &types.Empty{}, err
}

func (a *apiServer) Version(version int64) error {
	a.versionLock.Lock()
	defer a.versionLock.Unlock()
	a.version = version
	return nil
}

func (a *apiServer) getPipelineCancel(pipelineName string) context.CancelFunc {
	a.pipelineCancelsLock.Lock()
	defer a.pipelineCancelsLock.Unlock()
	return a.pipelineCancels[pipelineName]
}

func (a *apiServer) deletePipelineCancel(pipelineName string) context.CancelFunc {
	a.pipelineCancelsLock.Lock()
	defer a.pipelineCancelsLock.Unlock()
	cancel := a.pipelineCancels[pipelineName]
	delete(a.pipelineCancels, pipelineName)
	return cancel
}

func (a *apiServer) setPipelineCancel(pipelineName string, cancel context.CancelFunc) {
	a.pipelineCancelsLock.Lock()
	defer a.pipelineCancelsLock.Unlock()
	a.pipelineCancels[pipelineName] = cancel
}

func (a *apiServer) getJobCancel(jobID string) context.CancelFunc {
	a.jobCancelsLock.Lock()
	defer a.jobCancelsLock.Unlock()
	return a.jobCancels[jobID]
}

func (a *apiServer) deleteJobCancel(jobID string) context.CancelFunc {
	a.jobCancelsLock.Lock()
	defer a.jobCancelsLock.Unlock()
	cancel := a.jobCancels[jobID]
	delete(a.jobCancels, jobID)
	return cancel
}

func (a *apiServer) setJobCancel(jobID string, cancel context.CancelFunc) {
	a.jobCancelsLock.Lock()
	defer a.jobCancelsLock.Unlock()
	a.jobCancels[jobID] = cancel
}

// pipelineWatcher watches for pipelines and launch pipelineManager
// when it gets a pipeline that falls into a shard assigned to the
// API server.
func (a *apiServer) pipelineWatcher() {
	b := backoff.NewInfiniteBackOff()
	backoff.RetryNotify(func() error {
		pipelineWatcher, err := a.pipelines.ReadOnly(context.Background()).WatchByIndex(stoppedIndex, false)
		if err != nil {
			return err
		}
		for {
			event := <-pipelineWatcher.Watch()
			if event.Err != nil {
				return event.Err
			}
			pipelineName := string(event.Key)
			shardCtx := a.getShardCtx(a.hasher.HashPipeline(pipelineName))
			if shardCtx == nil {
				// Skip pipelines that don't fall into my shards
				continue
			}
			switch event.Type {
			case watch.EventPut:
				var pipelineInfo pps.PipelineInfo
				if err := event.Unmarshal(&pipelineName, &pipelineInfo); err != nil {
					return err
				}
				if cancel := a.deletePipelineCancel(pipelineName); cancel != nil {
					protolion.Infof("Appear to be running a pipeline (%s) that's already being run; this may be a bug", pipelineName)
					protolion.Infof("cancelling pipeline: %s", pipelineName)
					cancel()
				}
				pipelineCtx, cancel := context.WithCancel(shardCtx)
				a.setPipelineCancel(pipelineName, cancel)
				protolion.Infof("launching pipeline manager for pipeline %s", pipelineInfo.Pipeline.Name)
				go a.pipelineManager(pipelineCtx, &pipelineInfo)
			case watch.EventDelete:
				if cancel := a.deletePipelineCancel(pipelineName); cancel != nil {
					protolion.Infof("cancelling pipeline: %s", pipelineName)
					cancel()
					// delete the worker pool here, so we don't end up with a
					// race where the same pipeline is immediately recreated
					// and ends up using the defunct worker pool.
					a.delWorkerPool(pipelineName)
				}
			}
		}
	}, b, func(err error, d time.Duration) error {
		protolion.Errorf("error receiving pipeline updates: %v; retrying in %v", err, d)
		return nil
	})
	panic("pipelineWatcher should never exit")
}

// jobWatcher watches for unfinished jobs and launches jobManagers for
// the jobs that fall into this server's shards
func (a *apiServer) jobWatcher() {
	b := backoff.NewInfiniteBackOff()
	backoff.RetryNotify(func() error {
		// Wait for job events where JobInfo.Stopped is set to "false", and then
		// start JobManagers for those jobs
		jobWatcher, err := a.jobs.ReadOnly(context.Background()).WatchByIndex(stoppedIndex, false)
		if err != nil {
			return err
		}

		for {
			event := <-jobWatcher.Watch()
			if event.Err != nil {
				return event.Err
			}
			jobID := string(event.Key)
			shardCtx := a.getShardCtx(a.hasher.HashJob(jobID))
			if shardCtx == nil {
				// Skip jobs that don't fall into my shards
				continue
			}
			switch event.Type {
			case watch.EventPut:
				var jobInfo pps.JobInfo
				if err := event.Unmarshal(&jobID, &jobInfo); err != nil {
					return err
				}
				if cancel := a.deleteJobCancel(jobID); cancel != nil {
					protolion.Infof("Appear to be running a job (%s) that's already being run; this may be a bug", jobID)
					protolion.Infof("cancelling job: %s", jobID)
					cancel()
				}
				jobCtx, cancel := context.WithCancel(shardCtx)
				a.setJobCancel(jobID, cancel)
				protolion.Infof("launching job manager for job %s", jobInfo.Job.ID)
				go a.jobManager(jobCtx, &jobInfo)
			case watch.EventDelete:
				if cancel := a.deleteJobCancel(jobID); cancel != nil {
					cancel()
					protolion.Infof("cancelling job: %s", jobID)
				}
			}
		}
	}, b, func(err error, d time.Duration) error {
		protolion.Errorf("error receiving job updates: %v; retrying in %v", err, d)
		return nil
	})
	panic("jobWatcher should never exit")
}

func isAlreadyExistsErr(err error) bool {
	return err != nil && strings.Contains(err.Error(), "already exists")
}

func isNotFoundErr(err error) bool {
	return err != nil && strings.Contains(err.Error(), "not found")
}

func isContextCancelledErr(err error) bool {
	return err != nil && strings.Contains(err.Error(), context.Canceled.Error())
}

func (a *apiServer) pipelineManager(ctx context.Context, pipelineInfo *pps.PipelineInfo) {
	pipelineName := pipelineInfo.Pipeline.Name
	go func() {
		// Clean up workers if the pipeline gets cancelled
		<-ctx.Done()
		if err := a.deleteWorkers(pipelineInfo); err != nil {
			protolion.Errorf("error deleting workers for pipeline: %v", pipelineInfo.Pipeline.Name)
		}
		protolion.Infof("deleted workers for pipeline: %v", pipelineInfo.Pipeline.Name)
	}()

	b := backoff.NewInfiniteBackOff()
	if err := backoff.RetryNotify(func() error {
		if err := a.updatePipelineState(ctx, pipelineName, pps.PipelineState_PIPELINE_RUNNING); err != nil {
			return err
		}

		// Start worker pool
		a.workerPool(ctx, pipelineName)

		var provenance []*pfs.Repo
		for _, input := range pipelineInfo.Inputs {
			provenance = append(provenance, input.Repo)
		}

		pfsClient, err := a.getPFSClient()
		if err != nil {
			return err
		}
		// Create the output repo; if it already exists, do nothing
		if _, err := pfsClient.CreateRepo(ctx, &pfs.CreateRepoRequest{
			Repo:       &pfs.Repo{pipelineName},
			Provenance: provenance,
		}); err != nil {
			if !isAlreadyExistsErr(err) {
				return err
			}
		}

		// Create a k8s replication controller that runs the workers
		if err := a.createWorkers(pipelineInfo); err != nil {
			if !isAlreadyExistsErr(err) {
				return err
			}
		}

		branchSetCtx, cancel := context.WithCancel(ctx)
		defer cancel()
		branchSets, err := newBranchSetFactory(branchSetCtx, pfsClient, pipelineInfo.Inputs)
		if err != nil {
			return err
		}

		var job *pps.Job
		for {
			// Block until new input commit comes in, then gather input branches for processing
			branchSet, err := branchSets.Next()
			if err != nil {
				return err
			}

			// Create JobInput for new processing job
			var jobInputs []*pps.JobInput
			for _, pipelineInput := range pipelineInfo.Inputs {
				for _, branch := range branchSet {
					if pipelineInput.Repo.Name == branch.Head.Repo.Name && pipelineInput.Branch == branch.Name {
						jobInputs = append(jobInputs, &pps.JobInput{
							Name:   pipelineInput.Name,
							Commit: branch.Head,
							Glob:   pipelineInput.Glob,
							Lazy:   pipelineInput.Lazy,
						})
					}
				}
			}

			// Check if this input set has already been processed
			jobIter, err := a.jobs.ReadOnly(ctx).GetByIndex(jobsInputsIndex, jobInputs)
			if err != nil {
				return err
			}
			var jobExists bool
			for {
				var jobID string
				var jobInfo pps.JobInfo
				ok, err := jobIter.Next(&jobID, &jobInfo)
				if !ok {
					break
				}
				if err != nil {
					return err
				}
				if jobInfo.PipelineID == pipelineInfo.ID && jobInfo.PipelineVersion == pipelineInfo.Version {
					// TODO: we should check if the output commit exists.
					// If the output commit has been deleted, we should
					// re-run the job.
					// Also,
					jobExists = true
					break
				}
			}
			if jobExists {
				continue
			}

			job, err = a.CreateJob(ctx, &pps.CreateJobRequest{
				Transform:       pipelineInfo.Transform,
				Pipeline:        pipelineInfo.Pipeline,
				ParallelismSpec: pipelineInfo.ParallelismSpec,
				Inputs:          jobInputs,
				OutputRepo:      &pfs.Repo{pipelineInfo.Pipeline.Name},
				OutputBranch:    pipelineInfo.OutputBranch,
				Egress:          pipelineInfo.Egress,
				// TODO
				// Note that once the pipeline restarts, the `job` variable
				// is lost and we don't know who is our parent job.
				ParentJob: job,
			})
			if err != nil {
				return err
			}
			protolion.Infof("pipeline %s created job %v with the following input commits: %v", pipelineName, job.ID, jobInputs)
		}
		panic("unreachable")
		return nil
	}, b, func(err error, d time.Duration) error {
		if isContextCancelledErr(err) {
			return err
		}
		protolion.Errorf("error running pipelineManager: %v; retrying in %v", err, d)
		if err := a.updatePipelineState(ctx, pipelineName, pps.PipelineState_PIPELINE_RESTARTING); err != nil {
			protolion.Errorf("error updating pipeline state: %v", err)
		}
		return nil
	}); err != nil && !isContextCancelledErr(err) {
		panic(fmt.Sprintf("the retry loop should not exit with a non-context-cancelled error: %v", err))
	}
}

// pipelineStateToStopped defines what pipeline states are "stopped"
// states, meaning that pipelines in this state should not be managed
// by pipelineManager
func pipelineStateToStopped(state pps.PipelineState) bool {
	switch state {
	case pps.PipelineState_PIPELINE_STARTING:
		return false
	case pps.PipelineState_PIPELINE_RUNNING:
		return false
	case pps.PipelineState_PIPELINE_RESTARTING:
		return false
	case pps.PipelineState_PIPELINE_STOPPED:
		return true
	case pps.PipelineState_PIPELINE_FAILURE:
		return true
	default:
		panic(fmt.Sprintf("unrecognized pipeline state: %s", state))
	}
}

func (a *apiServer) updatePipelineState(ctx context.Context, pipelineName string, state pps.PipelineState) error {
	_, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
		pipelines := a.pipelines.ReadWrite(stm)
		pipelineInfo := new(pps.PipelineInfo)
		if err := pipelines.Get(pipelineName, pipelineInfo); err != nil {
			return err
		}
		pipelineInfo.State = state
		pipelineInfo.Stopped = pipelineStateToStopped(state)
		pipelines.Put(pipelineName, pipelineInfo)
		return nil
	})
	if isNotFoundErr(err) {
		return newErrPipelineNotFound(pipelineName)
	}
	return err
}

func (a *apiServer) updateJobState(stm col.STM, jobInfo *pps.JobInfo, state pps.JobState) error {
	// Update job counts
	if jobInfo.Pipeline != nil {
		pipelines := a.pipelines.ReadWrite(stm)
		pipelineInfo := new(pps.PipelineInfo)
		if err := pipelines.Get(jobInfo.Pipeline.Name, pipelineInfo); err != nil {
			return err
		}
		if pipelineInfo.JobCounts == nil {
			pipelineInfo.JobCounts = make(map[int32]int32)
		}
		if pipelineInfo.JobCounts[int32(jobInfo.State)] != 0 {
			pipelineInfo.JobCounts[int32(jobInfo.State)] -= 1
		}
		pipelineInfo.JobCounts[int32(state)] += 1
		pipelines.Put(pipelineInfo.Pipeline.Name, pipelineInfo)
	}
	jobInfo.State = state
	jobInfo.Stopped = jobStateToStopped(state)
	jobs := a.jobs.ReadWrite(stm)
	jobs.Put(jobInfo.Job.ID, jobInfo)
	return nil
}

func (a *apiServer) jobManager(ctx context.Context, jobInfo *pps.JobInfo) {
	jobID := jobInfo.Job.ID
	b := backoff.NewInfiniteBackOff()
	if err := backoff.RetryNotify(func() error {
		_, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
			jobs := a.jobs.ReadWrite(stm)
			jobInfo := new(pps.JobInfo)
			if err := jobs.Get(jobID, jobInfo); err != nil {
				return err
			}
			return a.updateJobState(stm, jobInfo, pps.JobState_JOB_RUNNING)
		})
		if err != nil {
			return err
		}

		pfsClient, err := a.getPFSClient()
		if err != nil {
			return err
		}
		df, err := newDatumFactory(ctx, pfsClient, jobInfo.Inputs, nil)
		if err != nil {
			return err
		}
		var wp WorkerPool
		if jobInfo.Pipeline != nil {
			wp = a.workerPool(ctx, jobInfo.Pipeline.Name)
		} else {
			wp = a.workerPool(ctx, jobInfo.Job.ID)
		}
		// process all datums
		var numData int
		tree := hashtree.NewHashTree()
		respCh := make(chan hashtree.HashTree)
		errCh := make(chan string)
		datum := df.Next()
		for {
			var resp hashtree.HashTree
			var datumErr string
			if datum != nil {
				select {
				case wp.DataCh() <- datumAndResp{
					datum:  datum,
					respCh: respCh,
					errCh:  errCh,
				}:
					datum = df.Next()
					numData++
				case resp = <-respCh:
					numData--
				case datumErr = <-errCh:
					numData--
				}
			} else {
				if numData == 0 {
					break
				}
				select {
				case resp = <-respCh:
				case datumErr = <-errCh:
				}
				numData--
			}
			if resp != nil {
				if err := tree.Merge(resp); err != nil {
					return err
				}
			}
			if datumErr != "" {
				_, err = col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
					jobs := a.jobs.ReadWrite(stm)
					jobInfo := new(pps.JobInfo)
					if err := jobs.Get(jobID, jobInfo); err != nil {
						return err
					}
					jobInfo.Error = datumErr
					jobInfo.Finished = now()
					return a.updateJobState(stm, jobInfo, pps.JobState_JOB_FAILURE)
				})
				if err != nil {
					return err
				}
			}
		}

		finishedTree, err := tree.Finish()
		if err != nil {
			return err
		}

		data, err := hashtree.Serialize(finishedTree)
		if err != nil {
			return err
		}

		objClient, err := a.getObjectClient()
		if err != nil {
			return err
		}

		putObjClient, err := objClient.PutObject(ctx)
		if err != nil {
			return err
		}
		if err := putObjClient.Send(&pfs.PutObjectRequest{
			Value: data,
		}); err != nil {
			return err
		}
		obj, err := putObjClient.CloseAndRecv()
		if err != nil {
			return err
		}

		var provenance []*pfs.Commit
		for _, input := range jobInfo.Inputs {
			provenance = append(provenance, input.Commit)
		}

		var outputCommit *pfs.Commit
		// If the branch does not exist, create it
		_, err = pfsClient.InspectCommit(ctx, &pfs.InspectCommitRequest{
			Commit: &pfs.Commit{
				Repo: jobInfo.OutputRepo,
				ID:   jobInfo.OutputBranch,
			},
		})

		if jobInfo.ParentJob != nil {
			// Wait for the parent job to finish, to ensure that output commits
			// are ordered correctly.
			// Right now we don't care if the parent job succeeded or not; we
			// just output a commit anyways.  But maybe it makes sense to only
			// output a commit if the parent job succeeded?
			// TODO
			if _, err := a.InspectJob(ctx, &pps.InspectJobRequest{
				Job:        jobInfo.ParentJob,
				BlockState: true,
			}); err != nil {
				return err
			}
		}

		if isNotFoundErr(err) {
			outputCommit, err = pfsClient.BuildCommit(ctx, &pfs.BuildCommitRequest{
				Parent: &pfs.Commit{
					Repo: jobInfo.OutputRepo,
				},
				Provenance: provenance,
				Tree:       obj,
			})
			if err != nil {
				return err
			}
			if _, err := pfsClient.SetBranch(ctx, &pfs.SetBranchRequest{
				Commit: outputCommit,
				Branch: jobInfo.OutputBranch,
			}); err != nil {
				return err
			}
		} else {
			outputCommit, err = pfsClient.BuildCommit(ctx, &pfs.BuildCommitRequest{
				Parent: &pfs.Commit{
					Repo: jobInfo.OutputRepo,
					ID:   jobInfo.OutputBranch,
				},
				Provenance: provenance,
				Tree:       obj,
			})
			if err != nil {
				return err
			}
		}

		_, err = col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
			jobs := a.jobs.ReadWrite(stm)
			jobInfo := new(pps.JobInfo)
			if err := jobs.Get(jobID, jobInfo); err != nil {
				return err
			}
			jobInfo.OutputCommit = outputCommit
			jobInfo.Finished = now()
			return a.updateJobState(stm, jobInfo, pps.JobState_JOB_SUCCESS)
		})
		return err
	}, b, func(err error, d time.Duration) error {
		if isContextCancelledErr(err) {
			return err
		}
		protolion.Errorf("error running jobManager: %v; retrying in %v", err, d)
		return nil
	}); err != nil && !isContextCancelledErr(err) {
		panic(fmt.Sprintf("the retry loop should not exit with a non-context-cancelled error: %v", err))
	}
}

// jobStateToStopped defines what job states are "stopped" states,
// meaning that jobs in this state should not be managed by jobManager
func jobStateToStopped(state pps.JobState) bool {
	switch state {
	case pps.JobState_JOB_STARTING:
		return false
	case pps.JobState_JOB_RUNNING:
		return false
	case pps.JobState_JOB_SUCCESS:
		return true
	case pps.JobState_JOB_FAILURE:
		return true
	default:
		panic(fmt.Sprintf("unrecognized job state: %s", state))
	}
}

func (a *apiServer) createWorkers(pipelineInfo *pps.PipelineInfo) error {
	options, err := a.getWorkerOptions(a.kubeClient, pipelineInfo, a.workerImage, a.workerImagePullPolicy)
	if err != nil {
		return err
	}
	if options.parallelism != int32(1) {
		return fmt.Errorf("pachyderm service only supports parallelism of 1, got %v", options.parallelism)
	}
	rc := &api.ReplicationController{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "ReplicationController",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   workerRcName(pipelineInfo),
			Labels: options.labels,
		},
		Spec: api.ReplicationControllerSpec{
			Selector: options.labels,
			Replicas: options.parallelism,
			Template: &api.PodTemplateSpec{
				ObjectMeta: api.ObjectMeta{
					Name:   workerRcName(pipelineInfo),
					Labels: options.labels,
				},
				Spec: workerPodSpec(options),
			},
		},
	}
	_, err = a.kubeClient.ReplicationControllers(a.namespace).Create(rc)
	return err
}

func (a *apiServer) deleteWorkers(pipelineInfo *pps.PipelineInfo) error {
	falseVal := false
	deleteOptions := &api.DeleteOptions{
		OrphanDependents: &falseVal,
	}
	return a.kubeClient.ReplicationControllers(a.namespace).Delete(workerRcName(pipelineInfo), deleteOptions)
}

// getShardCtx returns the context associated with a shard that this server
// manages.  It can also be used to determine if a shard is managed by this
// server
func (a *apiServer) getShardCtx(shard uint64) context.Context {
	a.shardLock.RLock()
	defer a.shardLock.RUnlock()
	ctxAndCancel := a.shardCtxs[shard]
	if ctxAndCancel != nil {
		return ctxAndCancel.ctx
	}
	return nil
}

func (a *apiServer) AddShard(shard uint64) error {
	ctx, cancel := context.WithCancel(context.Background())
	a.shardLock.Lock()
	defer a.shardLock.Unlock()
	if _, ok := a.shardCtxs[shard]; ok {
		return fmt.Errorf("shard %d is being added twice; this is likely a bug", shard)
	}
	a.shardCtxs[shard] = &ctxAndCancel{
		ctx:    ctx,
		cancel: cancel,
	}
	protolion.Infof("adding shard %d", shard)
	return nil
}

func (a *apiServer) DeleteShard(shard uint64) error {
	a.shardLock.Lock()
	defer a.shardLock.Unlock()
	ctxAndCancel, ok := a.shardCtxs[shard]
	if !ok {
		return fmt.Errorf("shard %d is being deleted, but it was never added; this is likely a bug", shard)
	}
	ctxAndCancel.cancel()
	delete(a.shardCtxs, shard)
	protolion.Infof("removing shard %d", shard)
	return nil
}

func (a *apiServer) getPFSClient() (pfs.APIClient, error) {
	if a.pachConn == nil {
		var onceErr error
		a.pachConnOnce.Do(func() {
			pachConn, err := grpc.Dial(a.address, grpc.WithInsecure())
			if err != nil {
				onceErr = err
			}
			a.pachConn = pachConn
		})
		if onceErr != nil {
			return nil, onceErr
		}
	}
	return pfs.NewAPIClient(a.pachConn), nil
}

func (a *apiServer) getObjectClient() (pfs.ObjectAPIClient, error) {
	if a.pachConn == nil {
		var onceErr error
		a.pachConnOnce.Do(func() {
			pachConn, err := grpc.Dial(a.address, grpc.WithInsecure())
			if err != nil {
				onceErr = err
			}
			a.pachConn = pachConn
		})
		if onceErr != nil {
			return nil, onceErr
		}
	}
	return pfs.NewObjectAPIClient(a.pachConn), nil
}

// RepoNameToEnvString is a helper which uppercases a repo name for
// use in environment variable names.
func RepoNameToEnvString(repoName string) string {
	return strings.ToUpper(repoName)
}

type workerOptions struct {
	labels                map[string]string
	parallelism           int32
	userImage             string
	workerImage           string
	workerImagePullPolicy string
	workerEnv             []api.EnvVar
	volumes               []api.Volume
	volumeMounts          []api.VolumeMount
	imagePullSecrets      []api.LocalObjectReference
}

func workerRcName(pipelineInfo *pps.PipelineInfo) string {
	// k8s won't allow RC names that contain upper-case letters
	// TODO: deal with name collision
	return fmt.Sprintf("pipeline-%s-v%d", strings.ToLower(pipelineInfo.Pipeline.Name), pipelineInfo.Version)
}

func (a *apiServer) getWorkerOptions(kubeClient *kube.Client, pipelineInfo *pps.PipelineInfo, workerImage string, workerImagePullPolicy string) (*workerOptions, error) {
	labels := labels(workerRcName(pipelineInfo))
	parallelism, err := GetExpectedNumWorkers(kubeClient, pipelineInfo.ParallelismSpec)
	if err != nil {
		return nil, err
	}
	userImage := pipelineInfo.Transform.Image
	if userImage == "" {
		userImage = DefaultUserImage
	}
	if workerImagePullPolicy == "" {
		workerImagePullPolicy = "IfNotPresent"
	}

	var workerEnv []api.EnvVar
	for name, value := range pipelineInfo.Transform.Env {
		workerEnv = append(
			workerEnv,
			api.EnvVar{
				Name:  name,
				Value: value,
			},
		)
	}
	// We use Kubernetes' "Downward API" so the workers know their IP
	// addresses, which they will then post on etcd so the job managers
	// can discover the workers.
	workerEnv = append(workerEnv, api.EnvVar{
		Name: client.PPSWorkerIPEnv,
		ValueFrom: &api.EnvVarSource{
			FieldRef: &api.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath:  "status.podIP",
			},
		},
	})
	// Set the etcd prefix env
	workerEnv = append(workerEnv, api.EnvVar{
		Name:  client.PPSEtcdPrefixEnv,
		Value: a.etcdPrefix,
	})
	// Set the pipline name env
	workerEnv = append(workerEnv, api.EnvVar{
		Name:  client.PPSPipelineNameEnv,
		Value: pipelineInfo.Pipeline.Name,
	})

	var volumes []api.Volume
	var volumeMounts []api.VolumeMount
	for _, secret := range pipelineInfo.Transform.Secrets {
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
	for _, secret := range pipelineInfo.Transform.ImagePullSecrets {
		imagePullSecrets = append(imagePullSecrets, api.LocalObjectReference{Name: secret})
	}

	return &workerOptions{
		labels:                labels,
		parallelism:           int32(parallelism),
		userImage:             userImage,
		workerImage:           workerImage,
		workerImagePullPolicy: workerImagePullPolicy,
		workerEnv:             workerEnv,
		volumes:               volumes,
		volumeMounts:          volumeMounts,
		imagePullSecrets:      imagePullSecrets,
	}, nil
}

func workerPodSpec(options *workerOptions) api.PodSpec {
	return api.PodSpec{
		InitContainers: []api.Container{
			{
				Name:            "init",
				Image:           options.workerImage,
				Command:         []string{"/pach/worker.sh"},
				ImagePullPolicy: api.PullPolicy(options.workerImagePullPolicy),
				Env:             options.workerEnv,
				VolumeMounts:    options.volumeMounts,
			},
		},
		Containers: []api.Container{
			{
				Name:    "user",
				Image:   options.userImage,
				Command: []string{"/pach-bin/guest.sh"},
				SecurityContext: &api.SecurityContext{
					Privileged: &trueVal, // god is this dumb
				},
				ImagePullPolicy: api.PullPolicy(options.workerImagePullPolicy),
				Env:             options.workerEnv,
				VolumeMounts:    options.volumeMounts,
			},
		},
		RestartPolicy:    "Always",
		Volumes:          options.volumes,
		ImagePullSecrets: options.imagePullSecrets,
	}
}

func (a *apiServer) jobPods(job *pps.Job) ([]api.Pod, error) {
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

func now() *types.Timestamp {
	t, err := types.TimestampProto(time.Now())
	if err != nil {
		panic(err)
	}
	return t
}
