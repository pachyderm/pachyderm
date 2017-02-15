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
	pfsclient "github.com/pachyderm/pachyderm/src/client/pfs"
	ppsclient "github.com/pachyderm/pachyderm/src/client/pps"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
	"github.com/pachyderm/pachyderm/src/server/pkg/metrics"
	ppsserver "github.com/pachyderm/pachyderm/src/server/pps"

	"github.com/cenkalti/backoff"
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
	etcdPrefix    string
	hasher        *ppsserver.Hasher
	address       string
	etcdClient    *etcd.Client
	pfsAPIClient  pfsclient.APIClient
	pfsClientOnce sync.Once
	kubeClient    *kube.Client
	shardLock     sync.RWMutex
	shardCtxs     map[uint64]*ctxAndCancel
	version       int64
	// versionLock protects the version field.
	// versionLock must be held BEFORE reading from version and UNTIL all
	// requests using version have returned
	versionLock           sync.RWMutex
	namespace             string
	workerShimImage       string
	workerImagePullPolicy string
	reporter              *metrics.Reporter
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

func (a *apiServer) CreateJob(ctx context.Context, request *ppsclient.CreateJobRequest) (response *ppsclient.Job, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "CreateJob")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	return nil, fmt.Errorf("TODO")
}

func (a *apiServer) InspectJob(ctx context.Context, request *ppsclient.InspectJobRequest) (response *ppsclient.JobInfo, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "InspectJob")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	return nil, fmt.Errorf("TODO")
}

func (a *apiServer) ListJob(ctx context.Context, request *ppsclient.ListJobRequest) (response *ppsclient.JobInfos, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "ListJob")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	return nil, fmt.Errorf("TODO")
}

func (a *apiServer) DeleteJob(ctx context.Context, request *ppsclient.DeleteJobRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "DeleteJob")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	return nil, fmt.Errorf("TODO")
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

func (a *apiServer) CreatePipeline(ctx context.Context, request *ppsclient.CreatePipelineRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "CreatePipeline")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	_, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
		pipelineInfo := &ppsclient.PipelineInfo{
			Pipeline:        request.Pipeline,
			Transform:       request.Transform,
			ParallelismSpec: request.ParallelismSpec,
			Inputs:          request.Inputs,
			Output:          request.Output,
			GcPolicy:        request.GcPolicy,
		}
		a.pipelines(stm).Put(pipelineInfo.Pipeline.Name, pipelineInfo)
		return nil
	})
	return &types.Empty{}, err
}

func (a *apiServer) InspectPipeline(ctx context.Context, request *ppsclient.InspectPipelineRequest) (response *ppsclient.PipelineInfo, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "InspectPipeline")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	return nil, fmt.Errorf("TODO")
}

func (a *apiServer) ListPipeline(ctx context.Context, request *ppsclient.ListPipelineRequest) (response *ppsclient.PipelineInfos, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "ListPipeline")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	return nil, fmt.Errorf("TODO")
}

func (a *apiServer) DeletePipeline(ctx context.Context, request *ppsclient.DeletePipelineRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "DeletePipeline")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	return nil, fmt.Errorf("TODO")
}

func (a *apiServer) StartPipeline(ctx context.Context, request *ppsclient.StartPipelineRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "StartPipeline")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	return nil, fmt.Errorf("TODO")
}

func (a *apiServer) StopPipeline(ctx context.Context, request *ppsclient.StopPipelineRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "StopPipeline")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	return nil, fmt.Errorf("TODO")
}

func (a *apiServer) DeleteAll(ctx context.Context, request *types.Empty) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "PPSDeleteAll")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	return nil, fmt.Errorf("TODO")
}

func (a *apiServer) Version(version int64) error {
	a.versionLock.Lock()
	defer a.versionLock.Unlock()
	a.version = version
	return nil
}

// pipelineWatcher watches for pipelines and launch pipelineManager
// when it gets a pipeline that falls into a shard assigned to the
// API server.
func (a *apiServer) pipelineWatcher() {
	b := backoff.NewExponentialBackOff()
	// never stop retrying
	b.MaxElapsedTime = 0
	backoff.RetryNotify(func() error {
		pipelineIter, err := a.pipelinesReadonly(context.Background()).Watch()
		if err != nil {
			return err
		}

		for {
			var pipelineName string
			pipelineInfo := new(ppsclient.PipelineInfo)
			ok, err := pipelineIter.Next(&pipelineName, pipelineInfo)
			if err != nil {
				return err
			}
			if !ok {
				return fmt.Errorf("pipeline stream broken")
			}
			shardCtx := a.getShardCtx(a.hasher.HashPipeline(pipelineInfo.Pipeline))
			if shardCtx != nil {
				go a.pipelineManager(shardCtx, pipelineInfo)
			}
		}
	}, b, func(err error, d time.Duration) {
		protolion.Errorf("error receiving pipeline updates: %v; retrying in %v", err, d)
	})
	panic("pipelineWatcher should never exit")
}

func (a *apiServer) pipelineManager(ctx context.Context, pipelineInfo *ppsclient.PipelineInfo) {
	b := backoff.NewExponentialBackOff()
	// never stop retrying
	b.MaxElapsedTime = 0
	backoff.RetryNotify(func() error {
		// Create a k8s replication controller that runs the workers
		if err := a.createWorkers(pipelineInfo); err != nil {
			return err
		}
		return nil
	}, b, func(err error, d time.Duration) {
		protolion.Errorf("error running pipelineManager: %v; retrying in %v", err, d)
	})
	panic("pipelineManager should never exit")
}

func (a *apiServer) createWorkers(pipelineInfo *ppsclient.PipelineInfo) error {
	options, err := getWorkerOptions(a.kubeClient, pipelineInfo, a.workerShimImage, a.workerImagePullPolicy)
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
			Name:   workerRcName(pipelineInfo.Pipeline.Name),
			Labels: options.labels,
		},
		Spec: api.ReplicationControllerSpec{
			Selector: options.labels,
			Replicas: options.parallelism,
			Template: &api.PodTemplateSpec{
				ObjectMeta: api.ObjectMeta{
					Name:   workerRcName(pipelineInfo.Pipeline.Name),
					Labels: options.labels,
				},
				Spec: workerPodSpec(options),
			},
		},
	}
	_, err = a.kubeClient.ReplicationControllers(a.namespace).Create(rc)
	return err
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

	return fmt.Errorf("TODO")
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
	return nil
}

func (a *apiServer) jobManager(ctx context.Context, job *ppsclient.Job) error {
	return fmt.Errorf("TODO")
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

// RepoNameToEnvString is a helper which uppercases a repo name for
// use in environment variable names.
func RepoNameToEnvString(repoName string) string {
	return strings.ToUpper(repoName)
}

type workerOptions struct {
	labels                map[string]string
	parallelism           int32
	userImage             string
	workerShimImage       string
	workerImagePullPolicy string
	workerEnv             []api.EnvVar
	volumes               []api.Volume
	volumeMounts          []api.VolumeMount
	imagePullSecrets      []api.LocalObjectReference
}

func workerRcName(pipelineName string) string {
	return fmt.Sprintf("pipeline_%s", pipelineName)
}

func getWorkerOptions(kubeClient *kube.Client, pipelineInfo *ppsclient.PipelineInfo, workerShimImage string, workerImagePullPolicy string) (*workerOptions, error) {
	labels := labels(workerRcName(pipelineInfo.Pipeline.Name))
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
		workerShimImage:       workerShimImage,
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
				Image:           options.workerShimImage,
				Command:         []string{"/pach/job-shim.sh"},
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
