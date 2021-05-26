package server

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"path"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"github.com/golang/protobuf/ptypes"
	"github.com/itchyny/gojq"
	"github.com/jmoiron/sqlx"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/robfig/cron"
	logrus "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/metadata"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube "k8s.io/client-go/kubernetes"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/client"
	enterpriseclient "github.com/pachyderm/pachyderm/v2/src/enterprise"
	"github.com/pachyderm/pachyderm/v2/src/internal/ancestry"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/lokiutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/metrics"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsconsts"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/serde"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/tracing"
	"github.com/pachyderm/pachyderm/v2/src/internal/tracing/extended"
	txnenv "github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/internal/watch"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	pfsServer "github.com/pachyderm/pachyderm/v2/src/server/pfs"
	ppsServer "github.com/pachyderm/pachyderm/v2/src/server/pps"
	"github.com/pachyderm/pachyderm/v2/src/server/pps/server/githook"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/common"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/datum"
	workerserver "github.com/pachyderm/pachyderm/v2/src/server/worker/server"
)

const (
	// DefaultUserImage is the image used for jobs when the user does not specify
	// an image.
	DefaultUserImage = "ubuntu:20.04"
	// DefaultDatumTries is the default number of times a datum will be tried
	// before we give up and consider the pipeline job failed.
	DefaultDatumTries = 3

	// DefaultLogsFrom is the default duration to return logs from, i.e. by
	// default we return logs from up to 24 hours ago.
	DefaultLogsFrom = time.Hour * 24
)

var (
	suite = "pachyderm"
)

func newErrPipelineNotFound(pipeline string) error {
	return errors.Errorf("pipeline %v not found", pipeline)
}

func newErrPipelineExists(pipeline string) error {
	return errors.Errorf("pipeline %v already exists", pipeline)
}

func newErrPipelineUpdate(pipeline string, reason string) error {
	return errors.Errorf("pipeline %v update error: %s", pipeline, reason)
}

type errGithookServiceNotFound struct {
	error
}

// apiServer implements the public interface of the Pachyderm Pipeline System,
// including all RPCs defined in the protobuf spec.
type apiServer struct {
	log.Logger
	etcdPrefix            string
	env                   serviceenv.ServiceEnv
	txnEnv                *txnenv.TransactionEnv
	namespace             string
	workerImage           string
	workerSidecarImage    string
	workerImagePullPolicy string
	storageRoot           string
	storageBackend        string
	storageHostPath       string
	iamRole               string
	imagePullSecret       string
	noExposeDockerSocket  bool
	reporter              *metrics.Reporter
	workerUsesRoot        bool
	workerGrpcPort        uint16
	port                  uint16
	httpPort              uint16
	peerPort              uint16
	gcPercent             int
	// collections
	pipelines    col.PostgresCollection
	pipelineJobs col.PostgresCollection
}

func merge(from, to map[string]bool) {
	for s := range from {
		to[s] = true
	}
}

func validateNames(names map[string]bool, input *pps.Input) error {
	switch {
	case input == nil:
		return nil // spouts can have nil input
	case input.Pfs != nil:
		if names[input.Pfs.Name] {
			return errors.Errorf(`name "%s" was used more than once`, input.Pfs.Name)
		}
		names[input.Pfs.Name] = true
	case input.Cron != nil:
		if names[input.Cron.Name] {
			return errors.Errorf(`name "%s" was used more than once`, input.Cron.Name)
		}
		names[input.Cron.Name] = true
	case input.Union != nil:
		for _, input := range input.Union {
			namesCopy := make(map[string]bool)
			merge(names, namesCopy)
			if err := validateNames(namesCopy, input); err != nil {
				return err
			}
			// we defer this because subinputs of a union input are allowed to
			// have conflicting names but other inputs that are, for example,
			// crossed with this union cannot conflict with any of the names it
			// might present
			defer merge(namesCopy, names)
		}
	case input.Cross != nil:
		for _, input := range input.Cross {
			if err := validateNames(names, input); err != nil {
				return err
			}
		}
	case input.Join != nil:
		for _, input := range input.Join {
			if err := validateNames(names, input); err != nil {
				return err
			}
		}
	case input.Group != nil:
		for _, input := range input.Group {
			if err := validateNames(names, input); err != nil {
				return err
			}
		}
	case input.Git != nil:
		if names[input.Git.Name] {
			return errors.Errorf(`name "%s" was used more than once`, input.Git.Name)
		}
		names[input.Git.Name] = true
	}
	return nil
}

func (a *apiServer) validateInput(pipelineName string, input *pps.Input) error {
	if err := validateNames(make(map[string]bool), input); err != nil {
		return err
	}
	return pps.VisitInput(input, func(input *pps.Input) error {
		set := false
		if input.Pfs != nil {
			set = true
			switch {
			case len(input.Pfs.Name) == 0:
				return errors.Errorf("input must specify a name")
			case input.Pfs.Name == "out":
				return errors.Errorf("input cannot be named \"out\", as pachyderm " +
					"already creates /pfs/out to collect job output")
			case input.Pfs.Repo == "":
				return errors.Errorf("input must specify a repo")
			case input.Pfs.Repo == "out" && input.Pfs.Name == "":
				return errors.Errorf("inputs based on repos named \"out\" must have " +
					"'name' set, as pachyderm already creates /pfs/out to collect " +
					"job output")
			case input.Pfs.Branch == "":
				return errors.Errorf("input must specify a branch")
			case !input.Pfs.S3 && len(input.Pfs.Glob) == 0:
				return errors.Errorf("input must specify a glob")
			case input.Pfs.S3 && input.Pfs.Glob != "/":
				return errors.Errorf("inputs that set 's3' to 'true' must also set " +
					"'glob', to \"/\", as the S3 gateway is only able to expose data " +
					"at the commit level")
			case input.Pfs.S3 && input.Pfs.Lazy:
				return errors.Errorf("input cannot specify both 's3' and 'lazy', as " +
					"'s3' requires input data to be accessed via Pachyderm's S3 " +
					"gateway rather than the file system")
			case input.Pfs.S3 && input.Pfs.EmptyFiles:
				return errors.Errorf("input cannot specify both 's3' and " +
					"'empty_files', as 's3' requires input data to be accessed via " +
					"Pachyderm's S3 gateway rather than the file system")
			}
		}
		if input.Cross != nil {
			if set {
				return errors.Errorf("multiple input types set")
			}
			set = true
		}
		if input.Join != nil {
			if set {
				return errors.Errorf("multiple input types set")
			}
			set = true
			if ppsutil.ContainsS3Inputs(input) {
				// The best datum semantics for s3 inputs embedded in join expressions
				// are not yet clear, and we see no use case for them yet, so block
				// them until we know how they should work
				return errors.Errorf("S3 inputs in join expressions are not supported")
			}
		}
		if input.Group != nil {
			if set {
				return errors.Errorf("multiple input types set")
			}
			set = true
			if ppsutil.ContainsS3Inputs(input) {
				// See above for "joins"; block s3 inputs in group expressions until
				// we know how they should work
				return errors.Errorf("S3 inputs in group expressions are not supported")
			}
		}
		if input.Union != nil {
			if set {
				return errors.Errorf("multiple input types set")
			}
			set = true
			if ppsutil.ContainsS3Inputs(input) {
				// See above for "joins"; block s3 inputs in union expressions until
				// we know how they should work
				return errors.Errorf("S3 inputs in union expressions are not supported")
			}
		}
		if input.Cron != nil {
			if set {
				return errors.Errorf("multiple input types set")
			}
			set = true
			if len(input.Cron.Name) == 0 {
				return errors.Errorf("input must specify a name")
			}
			if _, err := cron.ParseStandard(input.Cron.Spec); err != nil {
				return errors.Wrapf(err, "error parsing cron-spec")
			}
		}
		if input.Git != nil {
			if set {
				return errors.Errorf("multiple input types set")
			}
			logrus.Warn("githooks are deprecated and will be removed in a future version - see pipeline build steps for an alternative")
			set = true
			if err := pps.ValidateGitCloneURL(input.Git.URL); err != nil {
				return err
			}
		}
		if !set {
			return errors.Errorf("no input set")
		}
		return nil
	})
}

func validateTransform(transform *pps.Transform) error {
	if transform == nil {
		return errors.Errorf("pipeline must specify a transform")
	}
	if transform.Image == "" {
		return errors.Errorf("pipeline transform must contain an image")
	}
	return nil
}

func (a *apiServer) validateKube() {
	errors := false
	kubeClient := a.env.GetKubeClient()
	_, err := kubeClient.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		errors = true
		logrus.Errorf("unable to access kubernetes nodeslist, Pachyderm will continue to work but it will not be possible to use COEFFICIENT parallelism. error: %v", err)
	}
	_, err = kubeClient.CoreV1().Pods(a.namespace).Watch(metav1.ListOptions{Watch: true})
	if err != nil {
		errors = true
		logrus.Errorf("unable to access kubernetes pods, Pachyderm will continue to work but certain pipeline errors will result in pipelines being stuck indefinitely in \"starting\" state. error: %v", err)
	}
	pods, err := a.rcPods("pachd")
	if err != nil || len(pods) == 0 {
		errors = true
		logrus.Errorf("unable to access kubernetes pods, Pachyderm will continue to work but 'pachctl logs' will not work. error: %v", err)
	} else {
		// No need to check all pods since we're just checking permissions.
		pod := pods[0]
		_, err = kubeClient.CoreV1().Pods(a.namespace).GetLogs(
			pod.ObjectMeta.Name, &v1.PodLogOptions{
				Container: "pachd",
			}).Timeout(10 * time.Second).Do().Raw()
		if err != nil {
			errors = true
			logrus.Errorf("unable to access kubernetes logs, Pachyderm will continue to work but 'pachctl logs' will not work. error: %v", err)
		}
	}
	name := uuid.NewWithoutDashes()
	labels := map[string]string{"app": name}
	rc := &v1.ReplicationController{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ReplicationController",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
		Spec: v1.ReplicationControllerSpec{
			Selector: labels,
			Replicas: new(int32),
			Template: &v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   name,
					Labels: labels,
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:    "name",
							Image:   DefaultUserImage,
							Command: []string{"true"},
						},
					},
				},
			},
		},
	}
	if _, err := kubeClient.CoreV1().ReplicationControllers(a.namespace).Create(rc); err != nil {
		if err != nil {
			errors = true
			logrus.Errorf("unable to create kubernetes replication controllers, Pachyderm will not function properly until this is fixed. error: %v", err)
		}
	}
	if err := kubeClient.CoreV1().ReplicationControllers(a.namespace).Delete(name, nil); err != nil {
		if err != nil {
			errors = true
			logrus.Errorf("unable to delete kubernetes replication controllers, Pachyderm function properly but pipeline cleanup will not work. error: %v", err)
		}
	}
	if !errors {
		logrus.Infof("validating kubernetes access returned no errors")
	}
}

// authorizing a pipeline operation varies slightly depending on whether the
// pipeline is being created, updated, or deleted
type pipelineOperation uint8

const (
	// pipelineOpCreate is required for CreatePipeline
	pipelineOpCreate pipelineOperation = iota
	// pipelineOpListDatum is required for ListDatum
	pipelineOpListDatum
	// pipelineOpGetLogs is required for GetLogs
	pipelineOpGetLogs
	// pipelineOpUpdate is required for UpdatePipeline
	pipelineOpUpdate
	// pipelineOpUpdate is required for DeletePipeline
	pipelineOpDelete
)

// authorizePipelineOp checks if the user indicated by 'ctx' is authorized
// to perform 'operation' on the pipeline in 'info'
func (a *apiServer) authorizePipelineOp(ctx context.Context, operation pipelineOperation, input *pps.Input, output string) error {
	return a.txnEnv.WithReadContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
		return a.authorizePipelineOpInTransaction(txnCtx, operation, input, output)
	})
}

// authorizePipelineOpInTransaction is identical to authorizePipelineOp, but runs in the provided transaction
func (a *apiServer) authorizePipelineOpInTransaction(txnCtx *txncontext.TransactionContext, operation pipelineOperation, input *pps.Input, output string) error {
	_, err := a.env.AuthServer().WhoAmI(txnCtx.ClientContext, &auth.WhoAmIRequest{})
	if auth.IsErrNotActivated(err) {
		return nil // Auth isn't activated, skip authorization completely
	} else if err != nil {
		return err
	}

	if input != nil && operation != pipelineOpDelete {
		// Check that the user is authorized to read all input repos, and write to the
		// output repo (which the pipeline needs to be able to do on the user's
		// behalf)
		done := make(map[string]struct{}) // don't double-authorize repos
		if err := pps.VisitInput(input, func(in *pps.Input) error {
			var repo string
			if in.Pfs != nil {
				repo = in.Pfs.Repo
			} else {
				return nil
			}

			if _, ok := done[repo]; ok {
				return nil
			}
			done[repo] = struct{}{}
			return a.env.AuthServer().CheckRepoIsAuthorizedInTransaction(txnCtx, repo, auth.Permission_REPO_READ)
		}); err != nil {
			return err
		}
	}

	// Check that the user is authorized to write to the output repo.
	// Note: authorizePipelineOp is called before CreateRepo creates a
	// PipelineInfo proto in etcd, so PipelineManager won't have created an output
	// repo yet, and it's possible to check that the output repo doesn't exist
	// (if it did exist, we'd have to check that the user has permission to write
	// to it, and this is simpler)
	if output != "" {
		var required auth.Permission
		switch operation {
		case pipelineOpCreate:
			if _, err := a.env.PfsServer().InspectRepoInTransaction(txnCtx, &pfs.InspectRepoRequest{
				Repo: client.NewRepo(output),
			}); err == nil {
				// the repo already exists, so we need the same permissions as update
				required = auth.Permission_REPO_WRITE
			} else if isNotFoundErr(err) {
				return nil
			} else {
				return err
			}
		case pipelineOpListDatum, pipelineOpGetLogs:
			required = auth.Permission_REPO_READ
		case pipelineOpUpdate:
			required = auth.Permission_REPO_WRITE
		case pipelineOpDelete:
			if _, err := a.env.PfsServer().InspectRepoInTransaction(txnCtx, &pfs.InspectRepoRequest{
				Repo: client.NewRepo(output),
			}); isNotFoundErr(err) {
				// special case: the pipeline output repo has been deleted (so the
				// pipeline is now invalid). It should be possible to delete the pipeline.
				return nil
			}
			required = auth.Permission_REPO_DELETE
		default:
			return errors.Errorf("internal error, unrecognized operation %v", operation)
		}
		if err := a.env.AuthServer().CheckRepoIsAuthorizedInTransaction(txnCtx, output, required); err != nil {
			return err
		}
	}
	return nil
}

func (a *apiServer) UpdatePipelineJobState(ctx context.Context, request *pps.UpdatePipelineJobStateRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	if err := a.txnEnv.WithTransaction(ctx, func(txn txnenv.Transaction) error {
		return txn.UpdatePipelineJobState(request)
	}); err != nil {
		return nil, err
	}

	return &types.Empty{}, nil
}

func (a *apiServer) UpdatePipelineJobStateInTransaction(txnCtx *txncontext.TransactionContext, request *pps.UpdatePipelineJobStateRequest) error {
	pipelineJobs := a.pipelineJobs.ReadWrite(txnCtx.SqlTx)
	pipelineJobPtr := &pps.StoredPipelineJobInfo{}
	if err := pipelineJobs.Get(request.PipelineJob.ID, pipelineJobPtr); err != nil {
		return err
	}
	if ppsutil.IsTerminal(pipelineJobPtr.State) {
		return ppsServer.ErrPipelineJobFinished{pipelineJobPtr.PipelineJob}
	}

	pipelineJobPtr.Restart = request.Restart
	pipelineJobPtr.DataProcessed = request.DataProcessed
	pipelineJobPtr.DataSkipped = request.DataSkipped
	pipelineJobPtr.DataFailed = request.DataFailed
	pipelineJobPtr.DataRecovered = request.DataRecovered
	pipelineJobPtr.DataTotal = request.DataTotal
	pipelineJobPtr.Stats = request.Stats

	return ppsutil.UpdatePipelineJobState(a.pipelines.ReadWrite(txnCtx.SqlTx), pipelineJobs, pipelineJobPtr, request.State, request.Reason)
}

// CreatePipelineJob implements the protobuf pps.CreatePipelineJob RPC
func (a *apiServer) CreatePipelineJob(ctx context.Context, request *pps.CreatePipelineJobRequest) (response *pps.PipelineJob, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	var result *pps.PipelineJob
	if err := a.txnEnv.WithWriteContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
		commitInfo, err := a.env.PfsServer().InspectCommitInTransaction(txnCtx, &pfs.InspectCommitRequest{
			Commit: request.OutputCommit,
		})
		if err != nil {
			return err
		}
		if commitInfo.Finished != nil {
			return pfsServer.ErrCommitFinished{commitInfo.Commit}
		}
		if request.Stats == nil {
			request.Stats = &pps.ProcessStats{}
		}
		pipelines := a.pipelines.ReadWrite(txnCtx.SqlTx)
		pipelineJobs := a.pipelineJobs.ReadWrite(txnCtx.SqlTx)
		pipelineJobPtr := &pps.StoredPipelineJobInfo{
			PipelineJob:   client.NewPipelineJob(uuid.NewWithoutDashes()),
			OutputCommit:  request.OutputCommit,
			Pipeline:      request.Pipeline,
			Stats:         request.Stats,
			Restart:       request.Restart,
			DataProcessed: request.DataProcessed,
			DataSkipped:   request.DataSkipped,
			DataTotal:     request.DataTotal,
			DataFailed:    request.DataFailed,
			DataRecovered: request.DataRecovered,
			StatsCommit:   request.StatsCommit,
			Started:       request.Started,
			Finished:      request.Finished,
		}
		result = pipelineJobPtr.PipelineJob
		return ppsutil.UpdatePipelineJobState(pipelines, pipelineJobs, pipelineJobPtr, pps.PipelineJobState_JOB_STARTING, "")
	}); err != nil {
		return nil, err
	}
	return result, nil
}

// InspectPipelineJob implements the protobuf pps.InspectPipelineJob RPC
func (a *apiServer) InspectPipelineJob(ctx context.Context, request *pps.InspectPipelineJobRequest) (response *pps.PipelineJobInfo, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	if request.PipelineJob == nil && request.OutputCommit == nil {
		return nil, errors.Errorf("must specify either a PipelineJob or an OutputCommit")
	}
	pipelineJobs := a.pipelineJobs.ReadOnly(ctx)
	if request.OutputCommit != nil {
		if request.PipelineJob != nil {
			return nil, errors.Errorf("can't set both PipelineJob and OutputCommit")
		}
		pachClient := a.env.GetPachClient(ctx)
		ci, err := pachClient.InspectCommit(request.OutputCommit.Branch.Repo.Name, request.OutputCommit.Branch.Name, request.OutputCommit.ID)
		if err != nil {
			return nil, err
		}
		if err := a.listPipelineJob(ctx, nil, ci.Commit, nil, -1, false, "", func(pji *pps.PipelineJobInfo) error {
			if request.PipelineJob != nil {
				return errors.Errorf("internal error, more than 1 PipelineJob has output commit: %v (this is likely a bug)", request.OutputCommit)
			}
			request.PipelineJob = pji.PipelineJob
			return nil
		}); err != nil {
			return nil, err
		}
		if request.PipelineJob == nil {
			return nil, errors.Errorf("pipeline job with output commit %s not found", request.OutputCommit.ID)
		}
	}
	if request.BlockState {
		watcher, err := pipelineJobs.WatchOne(request.PipelineJob.ID)
		if err != nil {
			return nil, err
		}
		defer watcher.Close()

		for {
			ev, ok := <-watcher.Watch()
			if !ok {
				return nil, errors.Errorf("the stream for pipeline job updates closed unexpectedly")
			}
			switch ev.Type {
			case watch.EventError:
				return nil, ev.Err
			case watch.EventDelete:
				return nil, errors.Errorf("pipeline job %s was deleted", request.PipelineJob.ID)
			case watch.EventPut:
				var pipelineJobID string
				pipelineJobPtr := &pps.StoredPipelineJobInfo{}
				if err := ev.Unmarshal(&pipelineJobID, pipelineJobPtr); err != nil {
					return nil, err
				}
				if ppsutil.IsTerminal(pipelineJobPtr.State) {
					return a.pipelineJobInfoFromPtr(ctx, pipelineJobPtr, true)
				}
			}
		}
	}
	pipelineJobPtr := &pps.StoredPipelineJobInfo{}
	if err := pipelineJobs.Get(request.PipelineJob.ID, pipelineJobPtr); err != nil {
		return nil, err
	}
	pipelineJobInfo, err := a.pipelineJobInfoFromPtr(ctx, pipelineJobPtr, true)
	if err != nil {
		return nil, err
	}
	if request.Full {
		// If the pipeline job is running, we fill in WorkerStatus field, otherwise
		// we just return the pipelineJobInfo.
		if pipelineJobInfo.State != pps.PipelineJobState_JOB_RUNNING {
			return pipelineJobInfo, nil
		}
		workerPoolID := ppsutil.PipelineRcName(pipelineJobInfo.Pipeline.Name, pipelineJobInfo.PipelineVersion)
		workerStatus, err := workerserver.Status(ctx, workerPoolID, a.env.GetEtcdClient(), a.etcdPrefix, a.workerGrpcPort)
		if err != nil {
			logrus.Errorf("failed to get worker status with err: %s", err.Error())
		} else {
			// It's possible that the workers might be working on datums for other
			// jobs, we omit those since they're not part of the status for this
			// job.
			for _, status := range workerStatus {
				if status.PipelineJobID == pipelineJobInfo.PipelineJob.ID {
					pipelineJobInfo.WorkerStatus = append(pipelineJobInfo.WorkerStatus, status)
				}
			}
		}
	}
	return pipelineJobInfo, nil
}

// listPipelineJob is the internal implementation of ListPipelineJob shared
// between ListPipelineJob and ListPipelineJobStream. When ListPipelineJob is
// removed, this should be inlined into ListPipelineJobStream.
func (a *apiServer) listPipelineJob(
	ctx context.Context,
	pipeline *pps.Pipeline,
	outputCommit *pfs.Commit,
	inputCommits []*pfs.Commit,
	history int64,
	full bool,
	jqFilter string,
	f func(*pps.PipelineJobInfo) error,
) error {
	if pipeline != nil {
		// If 'pipeline is set, check that caller has access to the pipeline's
		// output repo; currently, that's all that's required for ListPipelineJob.
		//
		// If 'pipeline' isn't set, then we don't return an error (otherwise, a
		// caller without access to a single pipeline's output repo couldn't run
		// `pachctl list job` at all) and instead silently skip jobs where the user
		// doesn't have access to the job's output repo.
		if err := a.env.AuthServer().CheckRepoIsAuthorized(ctx, pipeline.Name, auth.Permission_PIPELINE_LIST_JOB); err != nil && !auth.IsErrNotActivated(err) {
			return err
		}
	}
	var err error
	if outputCommit != nil {
		outputCommit, err = a.resolveCommit(ctx, outputCommit)
		if err != nil {
			return err
		}
	}
	for i, inputCommit := range inputCommits {
		inputCommits[i], err = a.resolveCommit(ctx, inputCommit)
		if err != nil {
			return err
		}
	}
	var jqCode *gojq.Code
	var enc serde.Encoder
	var jsonBuffer bytes.Buffer
	if jqFilter != "" {
		jqQuery, err := gojq.Parse(jqFilter)
		if err != nil {
			return err
		}
		jqCode, err = gojq.Compile(jqQuery)
		if err != nil {
			return err
		}
		// ensure field names and enum values match with --raw output
		enc = serde.NewJSONEncoder(&jsonBuffer, serde.WithOrigName(true))
	}
	// specCommits holds the specCommits of pipelines that we're interested in
	specCommits := make(map[string]bool)
	if err := a.listPipelinePtr(ctx, pipeline, history,
		func(ptr *pps.StoredPipelineInfo) error {
			specCommits[ptr.SpecCommit.ID] = true
			return nil
		}); err != nil {
		return err
	}
	pipelineJobs := a.pipelineJobs.ReadOnly(ctx)
	pipelineJobPtr := &pps.StoredPipelineJobInfo{}
	_f := func(string) error {
		pipelineJobInfo, err := a.pipelineJobInfoFromPtr(ctx, pipelineJobPtr, len(inputCommits) > 0 || full)
		if err != nil {
			if isNotFoundErr(err) {
				// This can happen if a user deletes an upstream commit and thereby
				// deletes this job's output commit, but doesn't delete the etcdJobInfo.
				// In this case, the job is effectively deleted, but isn't removed from
				// etcd yet.
				return nil
			} else if auth.IsErrNotAuthorized(err) {
				return nil // skip job--see note under 'authIsActive && pipeline != nil'
			}
			return err
		}
		if len(inputCommits) > 0 {
			found := make([]bool, len(inputCommits))
			pps.VisitInput(pipelineJobInfo.Input, func(in *pps.Input) error {
				if in.Pfs != nil {
					for i, inputCommit := range inputCommits {
						if in.Pfs.Commit == inputCommit.ID {
							found[i] = true
						}
					}
				}
				return nil
			})
			for _, found := range found {
				if !found {
					return nil
				}
			}
		}
		if !specCommits[pipelineJobInfo.SpecCommit.ID] {
			return nil
		}
		if jqCode != nil {
			jsonBuffer.Reset()
			// convert pipelineJobInfo to a map[string]interface{} for use with gojq
			enc.EncodeProto(pipelineJobInfo)
			var jobInterface interface{}
			json.Unmarshal(jsonBuffer.Bytes(), &jobInterface)
			iter := jqCode.Run(jobInterface)
			// treat either jq false-y value as rejection
			if v, _ := iter.Next(); v == false || v == nil {
				return nil
			}
		}
		return f(pipelineJobInfo)
	}
	if pipeline != nil {
		return pipelineJobs.GetByIndex(ppsdb.PipelineJobsPipelineIndex, pipeline.Name, pipelineJobPtr, col.DefaultOptions(), _f)
	} else if outputCommit != nil {
		return pipelineJobs.GetByIndex(ppsdb.PipelineJobsOutputIndex, pfsdb.CommitKey(outputCommit), pipelineJobPtr, col.DefaultOptions(), _f)
	} else {
		return pipelineJobs.List(pipelineJobPtr, col.DefaultOptions(), _f)
	}
}

func (a *apiServer) pipelineJobInfoFromPtr(ctx context.Context, pipelineJobPtr *pps.StoredPipelineJobInfo, full bool) (*pps.PipelineJobInfo, error) {
	result := &pps.PipelineJobInfo{
		PipelineJob:   pipelineJobPtr.PipelineJob,
		Pipeline:      pipelineJobPtr.Pipeline,
		OutputRepo:    client.NewRepo(pipelineJobPtr.Pipeline.Name),
		OutputCommit:  pipelineJobPtr.OutputCommit,
		Restart:       pipelineJobPtr.Restart,
		DataProcessed: pipelineJobPtr.DataProcessed,
		DataSkipped:   pipelineJobPtr.DataSkipped,
		DataTotal:     pipelineJobPtr.DataTotal,
		DataFailed:    pipelineJobPtr.DataFailed,
		DataRecovered: pipelineJobPtr.DataRecovered,
		Stats:         pipelineJobPtr.Stats,
		StatsCommit:   pipelineJobPtr.StatsCommit,
		State:         pipelineJobPtr.State,
		Reason:        pipelineJobPtr.Reason,
		Started:       pipelineJobPtr.Started,
		Finished:      pipelineJobPtr.Finished,
	}

	pachClient := a.env.GetPachClient(ctx)
	commitInfo, err := pachClient.InspectCommit(pipelineJobPtr.OutputCommit.Branch.Repo.Name, pipelineJobPtr.OutputCommit.Branch.Name, pipelineJobPtr.OutputCommit.ID)
	if err != nil {
		if isNotFoundErr(err) {
			if err := col.NewSQLTx(ctx, a.env.GetDBClient(), func(sqlTx *sqlx.Tx) error {
				return a.pipelineJobs.ReadWrite(sqlTx).Delete(pipelineJobPtr.PipelineJob.ID)
			}); err != nil {
				return nil, err
			}
			return nil, errors.Errorf("job %s not found", pipelineJobPtr.PipelineJob.ID)
		}
		return nil, err
	}
	var specCommit *pfs.Commit
	for _, prov := range commitInfo.Provenance {
		if prov.Commit.Branch.Repo.Name == ppsconsts.SpecRepo && prov.Commit.Branch.Name == pipelineJobPtr.Pipeline.Name {
			specCommit = prov.Commit
			break
		}
	}
	if specCommit == nil {
		return nil, errors.Errorf("couldn't find spec commit for job %s, (this is likely a bug)", pipelineJobPtr.PipelineJob.ID)
	}
	result.SpecCommit = specCommit
	pipelinePtr := &pps.StoredPipelineInfo{}
	if err := a.pipelines.ReadOnly(ctx).Get(pipelineJobPtr.Pipeline.Name, pipelinePtr); err != nil {
		return nil, err
	}
	// Override the SpecCommit for the pipeline to be what it was when this job
	// was created, this prevents races between updating a pipeline and
	// previous jobs running.
	pipelinePtr.SpecCommit = specCommit
	if full {
		pipelineInfo, err := ppsutil.GetPipelineInfo(pachClient, pipelinePtr)
		if err != nil {
			return nil, err
		}
		result.Transform = pipelineInfo.Transform
		result.PipelineVersion = pipelineInfo.Version
		result.ParallelismSpec = pipelineInfo.ParallelismSpec
		result.Egress = pipelineInfo.Egress
		result.Service = pipelineInfo.Service
		result.Spout = pipelineInfo.Spout
		result.OutputBranch = pipelineInfo.OutputBranch
		result.ResourceRequests = pipelineInfo.ResourceRequests
		result.ResourceLimits = pipelineInfo.ResourceLimits
		result.SidecarResourceLimits = pipelineInfo.SidecarResourceLimits
		result.Input = ppsutil.PipelineJobInput(pipelineInfo, commitInfo)
		result.EnableStats = pipelineInfo.EnableStats
		result.Salt = pipelineInfo.Salt
		result.ChunkSpec = pipelineInfo.ChunkSpec
		result.DatumTimeout = pipelineInfo.DatumTimeout
		result.JobTimeout = pipelineInfo.JobTimeout
		result.DatumTries = pipelineInfo.DatumTries
		result.SchedulingSpec = pipelineInfo.SchedulingSpec
		result.PodSpec = pipelineInfo.PodSpec
		result.PodPatch = pipelineInfo.PodPatch
	}
	return result, nil
}

// ListPipelineJob implements the protobuf pps.ListPipelineJob RPC
func (a *apiServer) ListPipelineJob(request *pps.ListPipelineJobRequest, resp pps.API_ListPipelineJobServer) (retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	sent := 0
	defer func(start time.Time) {
		a.Log(request, fmt.Sprintf("stream containing %d PipelineJobInfos", sent), retErr, time.Since(start))
	}(time.Now())
	return a.listPipelineJob(resp.Context(), request.Pipeline, request.OutputCommit, request.InputCommit, request.History, request.Full, request.JqFilter, func(pji *pps.PipelineJobInfo) error {
		if err := resp.Send(pji); err != nil {
			return err
		}
		sent++
		return nil
	})
}

// FlushPipelineJob implements the protobuf pps.FlushPipelineJob RPC
func (a *apiServer) FlushPipelineJob(request *pps.FlushPipelineJobRequest, resp pps.API_FlushPipelineJobServer) (retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	sent := 0
	defer func(start time.Time) {
		a.Log(request, fmt.Sprintf("stream containing %d PipelineJobInfos", sent), retErr, time.Since(start))
	}(time.Now())
	var toRepos []*pfs.Repo
	for _, pipeline := range request.ToPipelines {
		toRepos = append(toRepos, client.NewRepo(pipeline.Name))
	}

	pachClient := a.env.GetPachClient(resp.Context())
	return pachClient.FlushCommit(request.Commits, toRepos, func(ci *pfs.CommitInfo) error {
		var pjis []*pps.PipelineJobInfo
		// FlushPipelineJob passes -1 for history because we don't know which version
		// of the pipeline created the output commit.
		if err := a.listPipelineJob(resp.Context(), nil, ci.Commit, nil, -1, false, "", func(pji *pps.PipelineJobInfo) error {
			pjis = append(pjis, pji)
			return nil
		}); err != nil {
			return err
		}
		if len(pjis) == 0 {
			// This is possible because the commit may be part of the stats
			// branch of a pipeline, in which case it's not the output commit
			// of any job, thus we ignore it, the job will be returned in
			// another call to this function, the one for the job's output
			// commit.
			return nil
		}
		if len(pjis) > 1 {
			return errors.Errorf("found too many jobs (%d) for output commit: %s", len(pjis), pfsdb.CommitKey(ci.Commit))
		}
		// Even though the commit has been finished the job isn't necessarily
		// finished yet, so we block on its state as well.
		pji, err := a.InspectPipelineJob(resp.Context(), &pps.InspectPipelineJobRequest{PipelineJob: pjis[0].PipelineJob, BlockState: true})
		if err != nil {
			return err
		}
		return resp.Send(pji)
	})
}

// DeletePipelineJob implements the protobuf pps.DeletePipelineJob RPC
func (a *apiServer) DeletePipelineJob(ctx context.Context, request *pps.DeletePipelineJobRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	if request.PipelineJob == nil {
		return nil, errors.New("Job cannot be nil")
	}
	if err := a.txnEnv.WithWriteContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
		if err := a.stopPipelineJob(txnCtx, request.PipelineJob, nil, "job deleted"); err != nil {
			return err
		}
		return a.pipelineJobs.ReadWrite(txnCtx.SqlTx).Delete(request.PipelineJob.ID)
	}); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

// StopPipelineJob implements the protobuf pps.StopPipelineJob RPC
func (a *apiServer) StopPipelineJob(ctx context.Context, request *pps.StopPipelineJobRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	if err := a.txnEnv.WithTransaction(ctx, func(txn txnenv.Transaction) error {
		return txn.StopPipelineJob(request)
	}); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

// StopPipelineJobInTransaction is identical to StopPipelineJob except that it can run inside an
// existing postgres transaction.  This is not an RPC.
func (a *apiServer) StopPipelineJobInTransaction(txnCtx *txncontext.TransactionContext, request *pps.StopPipelineJobRequest) error {
	reason := request.Reason
	if reason == "" {
		reason = "job stopped"
	}
	return a.stopPipelineJob(txnCtx, request.PipelineJob, request.OutputCommit, reason)
}

func (a *apiServer) stopPipelineJob(txnCtx *txncontext.TransactionContext, pipelineJob *pps.PipelineJob, outputCommit *pfs.Commit, reason string) error {
	pipelineJobs := a.pipelineJobs.ReadWrite(txnCtx.SqlTx)
	if (pipelineJob == nil) == (outputCommit == nil) {
		return errors.New("Exactly one of PipelineJob or OutputCommit must be specified")
	}

	pipelineJobInfo := &pps.StoredPipelineJobInfo{}
	if pipelineJob != nil {
		if err := pipelineJobs.Get(pipelineJob.ID, pipelineJobInfo); err != nil {
			return err
		}
		outputCommit = pipelineJobInfo.OutputCommit
	}

	commitInfo, err := a.env.PfsServer().InspectCommitInTransaction(txnCtx, &pfs.InspectCommitRequest{
		Commit: proto.Clone(outputCommit).(*pfs.Commit),
	})
	if err != nil && !pfsServer.IsCommitNotFoundErr(err) && !pfsServer.IsCommitDeletedErr(err) {
		return err
	}
	if commitInfo != nil {
		if err := a.env.PfsServer().FinishCommitInTransaction(txnCtx, &pfs.FinishCommitRequest{
			Commit: commitInfo.Commit,
			Empty:  true,
		}); err != nil && !pfsServer.IsCommitNotFoundErr(err) && !pfsServer.IsCommitDeletedErr(err) && !pfsServer.IsCommitFinishedErr(err) {
			return err
		}

		if err := a.env.PfsServer().FinishCommitInTransaction(txnCtx, &pfs.FinishCommitRequest{
			Commit: ppsutil.GetStatsCommit(commitInfo),
			Empty:  true,
		}); err != nil && !pfsServer.IsCommitNotFoundErr(err) && !pfsServer.IsCommitDeletedErr(err) && !pfsServer.IsCommitFinishedErr(err) {
			return err
		}
	}

	handleJob := func(pji *pps.StoredPipelineJobInfo) error {
		// TODO: We can still not update a job's state if we fail here. This is
		// probably fine for now since we are likely to have a more comprehensive
		// solution to this with global ids.
		if err := ppsutil.UpdatePipelineJobState(a.pipelines.ReadWrite(txnCtx.SqlTx), pipelineJobs, pji, pps.PipelineJobState_JOB_KILLED, reason); err != nil && !ppsServer.IsPipelineJobFinishedErr(err) {
			return err
		}
		return nil
	}

	if pipelineJob != nil {
		return handleJob(pipelineJobInfo)
	}

	// Continue on idempotently if we find multiple pipeline jobs or none.
	return pipelineJobs.GetByIndex(ppsdb.PipelineJobsOutputIndex, pfsdb.CommitKey(outputCommit), pipelineJobInfo, col.DefaultOptions(), func(string) error {
		return handleJob(pipelineJobInfo)
	})
}

// RestartDatum implements the protobuf pps.RestartDatum RPC
func (a *apiServer) RestartDatum(ctx context.Context, request *pps.RestartDatumRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	pipelineJobInfo, err := a.InspectPipelineJob(ctx, &pps.InspectPipelineJobRequest{
		PipelineJob: request.PipelineJob,
	})
	if err != nil {
		return nil, err
	}
	workerPoolID := ppsutil.PipelineRcName(pipelineJobInfo.Pipeline.Name, pipelineJobInfo.PipelineVersion)
	if err := workerserver.Cancel(ctx, workerPoolID, a.env.GetEtcdClient(), a.etcdPrefix, a.workerGrpcPort, request.PipelineJob.ID, request.DataFilters); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

func (a *apiServer) InspectDatum(ctx context.Context, request *pps.InspectDatumRequest) (response *pps.DatumInfo, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	// TODO: Auth?
	if err := a.collectDatums(ctx, request.Datum.PipelineJob, func(meta *datum.Meta, pfsState *pfs.File) error {
		if common.DatumID(meta.Inputs) == request.Datum.ID {
			response = convertDatumMetaToInfo(meta)
			response.PfsState = pfsState
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return response, nil
}

func (a *apiServer) ListDatum(request *pps.ListDatumRequest, server pps.API_ListDatumServer) (retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, nil, retErr, time.Since(start)) }(time.Now())
	// TODO: Auth?
	if request.Input != nil {
		return a.listDatumInput(server.Context(), request.Input, func(meta *datum.Meta) error {
			return server.Send(convertDatumMetaToInfo(meta))
		})
	}
	return a.collectDatums(server.Context(), request.PipelineJob, func(meta *datum.Meta, _ *pfs.File) error {
		return server.Send(convertDatumMetaToInfo(meta))
	})
}

func (a *apiServer) listDatumInput(ctx context.Context, input *pps.Input, cb func(*datum.Meta) error) error {
	setInputDefaults("", input)
	if visitErr := pps.VisitInput(input, func(input *pps.Input) error {
		if input.Pfs != nil {
			pachClient := a.env.GetPachClient(ctx)
			ci, err := pachClient.InspectCommit(input.Pfs.Repo, input.Pfs.Branch, "")
			if err != nil {
				return err
			}
			input.Pfs.Commit = ci.Commit.ID
		}
		if input.Cron != nil {
			return errors.Errorf("can't list datums with a cron input, there will be no datums until the pipeline is created")
		}
		return nil
	}); visitErr != nil {
		return visitErr
	}
	pachClient := a.env.GetPachClient(ctx)
	di, err := datum.NewIterator(pachClient, input)
	if err != nil {
		return err
	}
	return di.Iterate(func(meta *datum.Meta) error {
		return cb(meta)
	})
}

func convertDatumMetaToInfo(meta *datum.Meta) *pps.DatumInfo {
	di := &pps.DatumInfo{
		Datum: &pps.Datum{
			PipelineJob: client.NewPipelineJob(meta.PipelineJobID),
			ID:          common.DatumID(meta.Inputs),
		},
		State: convertDatumState(meta.State),
		Stats: meta.Stats,
	}
	for _, input := range meta.Inputs {
		di.Data = append(di.Data, input.FileInfo)
	}
	return di
}

// TODO: this is a bit wonky, but it is necessary based on the dependency graph.
func convertDatumState(state datum.State) pps.DatumState {
	switch state {
	case datum.State_FAILED:
		return pps.DatumState_FAILED
	case datum.State_RECOVERED:
		return pps.DatumState_RECOVERED
	default:
		return pps.DatumState_SUCCESS
	}
}

func (a *apiServer) collectDatums(ctx context.Context, pipelineJob *pps.PipelineJob, cb func(*datum.Meta, *pfs.File) error) error {
	pipelineJobInfo, err := a.InspectPipelineJob(ctx, &pps.InspectPipelineJobRequest{
		PipelineJob: client.NewPipelineJob(pipelineJob.ID),
	})
	if err != nil {
		return err
	}
	pachClient := a.env.GetPachClient(ctx)
	fsi := datum.NewCommitIterator(pachClient, pipelineJobInfo.StatsCommit)
	return fsi.Iterate(func(meta *datum.Meta) error {
		// TODO: Potentially refactor into datum package (at least the path).
		pfsState := &pfs.File{
			Commit: pipelineJobInfo.StatsCommit,
			Path:   "/" + path.Join(datum.PFSPrefix, common.DatumID(meta.Inputs)),
		}
		return cb(meta, pfsState)
	})
}

func (a *apiServer) GetLogs(request *pps.GetLogsRequest, apiGetLogsServer pps.API_GetLogsServer) (retErr error) {
	// Set the default for the `Since` field.
	if request.Since == nil || (request.Since.Seconds == 0 && request.Since.Nanos == 0) {
		request.Since = types.DurationProto(DefaultLogsFrom)
	}
	if a.env.Config().LokiLogging || request.UseLokiBackend {
		pachClient := a.env.GetPachClient(apiGetLogsServer.Context())
		resp, err := pachClient.Enterprise.GetState(context.Background(),
			&enterpriseclient.GetStateRequest{})
		if err != nil {
			return errors.Wrapf(grpcutil.ScrubGRPC(err), "could not get enterprise status")
		}
		if resp.State == enterpriseclient.State_ACTIVE {
			return a.getLogsLoki(request, apiGetLogsServer)
		}
		return errors.Errorf("enterprise must be enabled to use loki logging")
	}
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, nil, retErr, time.Since(start)) }(time.Now())

	// Authorize request and get list of pods containing logs we're interested in
	// (based on pipeline and job filters)
	var rcName, containerName string
	if request.Pipeline == nil && request.PipelineJob == nil {
		if len(request.DataFilters) > 0 || request.Datum != nil {
			return errors.Errorf("must specify the PipelineJob or Pipeline that the datum is from to get logs for it")
		}
		containerName, rcName = "pachd", "pachd"
	} else {
		containerName = client.PPSWorkerUserContainerName

		// 1) Lookup the PipelineInfo for this pipeline/job, for auth and to get the
		// RC name
		var pipelineInfo *pps.PipelineInfo
		var err error
		if request.Pipeline != nil {
			pipelineInfo, err = a.inspectPipeline(apiGetLogsServer.Context(), request.Pipeline.Name)
			if err != nil {
				return errors.Wrapf(err, "could not get pipeline information for %s", request.Pipeline.Name)
			}
		} else if request.PipelineJob != nil {
			// If user provides a pipeline job, lookup the pipeline from the
			// PipelineJobInfo, and then get the pipeline RC
			pipelineJobPtr := &pps.StoredPipelineJobInfo{}
			err = a.pipelineJobs.ReadOnly(apiGetLogsServer.Context()).Get(request.PipelineJob.ID, pipelineJobPtr)
			if err != nil {
				return errors.Wrapf(err, "could not get pipeline job information for \"%s\"", request.PipelineJob.ID)
			}
			pipelineInfo, err = a.inspectPipeline(apiGetLogsServer.Context(), pipelineJobPtr.Pipeline.Name)
			if err != nil {
				return errors.Wrapf(err, "could not get pipeline information for %s", pipelineJobPtr.Pipeline.Name)
			}
		}

		// 2) Check whether the caller is authorized to get logs from this pipeline/job
		if err := a.authorizePipelineOp(apiGetLogsServer.Context(), pipelineOpGetLogs, pipelineInfo.Input, pipelineInfo.Pipeline.Name); err != nil {
			return err
		}

		// 3) Get rcName for this pipeline
		rcName = ppsutil.PipelineRcName(pipelineInfo.Pipeline.Name, pipelineInfo.Version)
		if err != nil {
			return err
		}
	}

	// Get pods managed by the RC we're scraping (either pipeline or pachd)
	pods, err := a.rcPods(rcName)
	if err != nil {
		return errors.Wrapf(err, "could not get pods in rc \"%s\" containing logs", rcName)
	}
	if len(pods) == 0 {
		return errors.Errorf("no pods belonging to the rc \"%s\" were found", rcName)
	}
	// Convert request.From to a usable timestamp.
	since, err := types.DurationFromProto(request.Since)
	if err != nil {
		return errors.Wrapf(err, "invalid from time")
	}
	sinceSeconds := int64(since.Seconds())

	// Spawn one goroutine per pod. Each goro writes its pod's logs to a channel
	// and channels are read into the output server in a stable order.
	// (sort the pods to make sure that the order of log lines is stable)
	sort.Sort(podSlice(pods))
	logCh := make(chan *pps.LogMessage)
	var eg errgroup.Group
	var mu sync.Mutex
	eg.Go(func() error {
		for _, pod := range pods {
			pod := pod
			if !request.Follow {
				mu.Lock()
			}
			eg.Go(func() (retErr error) {
				if !request.Follow {
					defer mu.Unlock()
				}
				tailLines := &request.Tail
				if *tailLines <= 0 {
					tailLines = nil
				}
				// Get full set of logs from pod i
				stream, err := a.env.GetKubeClient().CoreV1().Pods(a.namespace).GetLogs(
					pod.ObjectMeta.Name, &v1.PodLogOptions{
						Container:    containerName,
						Follow:       request.Follow,
						TailLines:    tailLines,
						SinceSeconds: &sinceSeconds,
					}).Timeout(10 * time.Second).Stream()
				if err != nil {
					return err
				}
				defer func() {
					if err := stream.Close(); err != nil && retErr == nil {
						retErr = err
					}
				}()

				// Parse pods' log lines, and filter out irrelevant ones
				scanner := bufio.NewScanner(stream)
				for scanner.Scan() {
					msg := new(pps.LogMessage)
					if containerName == "pachd" {
						msg.Message = scanner.Text()
					} else {
						logBytes := scanner.Bytes()
						if err := jsonpb.Unmarshal(bytes.NewReader(logBytes), msg); err != nil {
							continue
						}

						// Filter out log lines that don't match on pipeline or job
						if request.Pipeline != nil && request.Pipeline.Name != msg.PipelineName {
							continue
						}
						if request.PipelineJob != nil && request.PipelineJob.ID != msg.PipelineJobID {
							continue
						}
						if request.Datum != nil && request.Datum.ID != msg.DatumID {
							continue
						}
						if request.Master != msg.Master {
							continue
						}
						if !common.MatchDatum(request.DataFilters, msg.Data) {
							continue
						}
					}
					msg.Message = strings.TrimSuffix(msg.Message, "\n")

					// Log message passes all filters -- return it
					select {
					case logCh <- msg:
					case <-apiGetLogsServer.Context().Done():
						return nil
					}
				}
				return nil
			})
		}
		return nil
	})
	var egErr error
	go func() {
		egErr = eg.Wait()
		close(logCh)
	}()

	for msg := range logCh {
		if err := apiGetLogsServer.Send(msg); err != nil {
			return err
		}
	}
	return egErr
}

func (a *apiServer) getLogsLoki(request *pps.GetLogsRequest, apiGetLogsServer pps.API_GetLogsServer) (retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, nil, retErr, time.Since(start)) }(time.Now())

	// Authorize request and get list of pods containing logs we're interested in
	// (based on pipeline and job filters)
	loki, err := a.env.GetLokiClient()
	if err != nil {
		return err
	}
	since, err := types.DurationFromProto(request.Since)
	if err != nil {
		return errors.Wrapf(err, "invalid from time")
	}
	if request.Pipeline == nil && request.PipelineJob == nil {
		if len(request.DataFilters) > 0 || request.Datum != nil {
			return errors.Errorf("must specify the Job or Pipeline that the datum is from to get logs for it")
		}
		// no authorization is done to get logs from master
		return lokiutil.QueryRange(apiGetLogsServer.Context(), loki, `{app="pachd"}`, time.Now().Add(-since), time.Now(), request.Follow, func(t time.Time, line string) error {
			return apiGetLogsServer.Send(&pps.LogMessage{
				Message: strings.TrimSuffix(line, "\n"),
			})
		})
	}

	// 1) Lookup the PipelineInfo for this pipeline/job, for auth and to get the
	// RC name
	var pipelineInfo *pps.PipelineInfo
	if request.Pipeline != nil {
		pipelineInfo, err = a.inspectPipeline(apiGetLogsServer.Context(), request.Pipeline.Name)
		if err != nil {
			return errors.Wrapf(err, "could not get pipeline information for %s", request.Pipeline.Name)
		}
	} else if request.PipelineJob != nil {
		// If user provides a pipeline job, lookup the pipeline from the
		// PipelineJobInfo, and then get the pipeline RC
		pipelineJobPtr := &pps.StoredPipelineJobInfo{}
		err = a.pipelineJobs.ReadOnly(apiGetLogsServer.Context()).Get(request.PipelineJob.ID, pipelineJobPtr)
		if err != nil {
			return errors.Wrapf(err, "could not get pipeline job information for \"%s\"", request.PipelineJob.ID)
		}
		pipelineInfo, err = a.inspectPipeline(apiGetLogsServer.Context(), pipelineJobPtr.Pipeline.Name)
		if err != nil {
			return errors.Wrapf(err, "could not get pipeline information for %s", pipelineJobPtr.Pipeline.Name)
		}
	}

	// 2) Check whether the caller is authorized to get logs from this pipeline/job
	if err := a.authorizePipelineOp(apiGetLogsServer.Context(), pipelineOpGetLogs, pipelineInfo.Input, pipelineInfo.Pipeline.Name); err != nil {
		return err
	}
	query := fmt.Sprintf(`{pipelineName=%q, container="user"}`, pipelineInfo.Pipeline.Name)
	if request.Master {
		query += contains("master")
	}
	if request.PipelineJob != nil {
		query += contains(request.PipelineJob.ID)
	}
	if request.Datum != nil {
		query += contains(request.Datum.ID)
	}
	for _, filter := range request.DataFilters {
		query += contains(filter)
	}
	return lokiutil.QueryRange(apiGetLogsServer.Context(), loki, query, time.Now().Add(-since), time.Now(), request.Follow, func(t time.Time, line string) error {
		msg := &pps.LogMessage{}
		// These filters are almost always unnecessary because we apply
		// them in the Loki request, but many of them are just done with
		// string matching so there technically could be some false
		// positive matches (although it's pretty unlikely), checking here
		// just makes sure we don't accidentally intersperse unrelated log
		// messages.
		if err := jsonpb.Unmarshal(strings.NewReader(line), msg); err != nil {
			return nil
		}
		if request.Pipeline != nil && request.Pipeline.Name != msg.PipelineName {
			return nil
		}
		if request.PipelineJob != nil && request.PipelineJob.ID != msg.PipelineJobID {
			return nil
		}
		if request.Datum != nil && request.Datum.ID != msg.DatumID {
			return nil
		}
		if request.Master != msg.Master {
			return nil
		}
		if !common.MatchDatum(request.DataFilters, msg.Data) {
			return nil
		}
		msg.Message = strings.TrimSuffix(msg.Message, "\n")
		return apiGetLogsServer.Send(msg)
	})
}

func contains(s string) string {
	return fmt.Sprintf(" |= %q", s)
}

type podSlice []v1.Pod

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

func (a *apiServer) validatePipelineRequest(request *pps.CreatePipelineRequest) error {
	// TODO: Remove when at feature parity.
	var err error
	request, err = a.validateV2Features(request)
	if err != nil {
		return err
	}
	if request.Pipeline == nil {
		return errors.New("invalid pipeline spec: request.Pipeline cannot be nil")
	}
	if request.Pipeline.Name == "" {
		return errors.New("invalid pipeline spec: request.Pipeline.Name cannot be empty")
	}
	if err := ancestry.ValidateName(request.Pipeline.Name); err != nil {
		return errors.Wrapf(err, "invalid pipeline name")
	}
	if len(request.Pipeline.Name) > 63 {
		return errors.Errorf("pipeline name is %d characters long, but must have at most 63: %q",
			len(request.Pipeline.Name), request.Pipeline.Name)
	}
	// TODO(msteffen) eventually TFJob and Transform will be alternatives, but
	// currently TFJob isn't supported
	if request.TFJob != nil {
		return errors.New("embedding TFJobs in pipelines is not supported yet")
	}
	if request.S3Out && ((request.Service != nil) || (request.Spout != nil)) {
		return errors.New("s3 output is not supported in spouts or services")
	}
	if request.Transform == nil {
		return errors.Errorf("pipeline must specify a transform")
	}
	if request.ReprocessSpec != "" &&
		request.ReprocessSpec != client.ReprocessSpecUntilSuccess &&
		request.ReprocessSpec != client.ReprocessSpecEveryJob {
		return errors.Errorf("invalid pipeline spec: ReprocessSpec must be one of '%s' or '%s'",
			client.ReprocessSpecUntilSuccess, client.ReprocessSpecEveryJob)
	}
	return nil
}

// TODO: Implement the appropriate features.
func (a *apiServer) validateV2Features(request *pps.CreatePipelineRequest) (*pps.CreatePipelineRequest, error) {
	if request.CacheSize != "" {
		return nil, errors.Errorf("CacheSize not implemented")
	}
	if request.Service == nil && request.Spout == nil {
		request.EnableStats = true
	}
	if request.MaxQueueSize != 0 {
		return nil, errors.Errorf("MaxQueueSize not implemented")
	}
	return request, nil
}

func (a *apiServer) validatePipeline(pipelineInfo *pps.PipelineInfo) error {
	if pipelineInfo.Pipeline == nil {
		return errors.New("invalid pipeline spec: Pipeline field cannot be nil")
	}
	if pipelineInfo.Pipeline.Name == "" {
		return errors.New("invalid pipeline spec: Pipeline.Name cannot be empty")
	}
	if err := ancestry.ValidateName(pipelineInfo.Pipeline.Name); err != nil {
		return errors.Wrapf(err, "invalid pipeline name")
	}
	first := rune(pipelineInfo.Pipeline.Name[0])
	if !unicode.IsLetter(first) && !unicode.IsDigit(first) {
		return errors.Errorf("pipeline names must start with an alphanumeric character")
	}
	if len(pipelineInfo.Pipeline.Name) > 63 {
		return errors.Errorf("pipeline name is %d characters long, but must have at most 63: %q",
			len(pipelineInfo.Pipeline.Name), pipelineInfo.Pipeline.Name)
	}
	if err := validateTransform(pipelineInfo.Transform); err != nil {
		return errors.Wrapf(err, "invalid transform")
	}
	if err := a.validateInput(pipelineInfo.Pipeline.Name, pipelineInfo.Input); err != nil {
		return err
	}
	if pipelineInfo.ParallelismSpec != nil {
		if pipelineInfo.ParallelismSpec.Coefficient < 0 {
			return errors.New("ParallelismSpec.Coefficient cannot be negative")
		}
		if pipelineInfo.ParallelismSpec.Constant != 0 &&
			pipelineInfo.ParallelismSpec.Coefficient != 0 {
			return errors.New("contradictory parallelism strategies: must set at " +
				"most one of ParallelismSpec.Constant and ParallelismSpec.Coefficient")
		}
		if pipelineInfo.Service != nil && pipelineInfo.ParallelismSpec.Constant != 1 {
			return errors.New("services can only be run with a constant parallelism of 1")
		}
	}
	if pipelineInfo.OutputBranch == "" {
		return errors.New("pipeline needs to specify an output branch")
	}
	if _, err := resource.ParseQuantity(pipelineInfo.CacheSize); err != nil {
		return errors.Wrapf(err, "could not parse cacheSize '%s'", pipelineInfo.CacheSize)
	}
	if pipelineInfo.JobTimeout != nil {
		_, err := types.DurationFromProto(pipelineInfo.JobTimeout)
		if err != nil {
			return err
		}
	}
	if pipelineInfo.DatumTimeout != nil {
		_, err := types.DurationFromProto(pipelineInfo.DatumTimeout)
		if err != nil {
			return err
		}
	}
	if pipelineInfo.PodSpec != "" && !json.Valid([]byte(pipelineInfo.PodSpec)) {
		return errors.Errorf("malformed PodSpec")
	}
	if pipelineInfo.PodPatch != "" && !json.Valid([]byte(pipelineInfo.PodPatch)) {
		return errors.Errorf("malformed PodPatch")
	}
	if pipelineInfo.Service != nil {
		validServiceTypes := map[v1.ServiceType]bool{
			v1.ServiceTypeClusterIP:    true,
			v1.ServiceTypeLoadBalancer: true,
			v1.ServiceTypeNodePort:     true,
		}

		if !validServiceTypes[v1.ServiceType(pipelineInfo.Service.Type)] {
			return errors.Errorf("the following service type %s is not allowed", pipelineInfo.Service.Type)
		}
	}
	if pipelineInfo.Spout != nil {
		if pipelineInfo.EnableStats {
			return errors.Errorf("spouts are not allowed to have a stats branch")
		}
		if pipelineInfo.Spout.Service == nil && pipelineInfo.Input != nil {
			return errors.Errorf("spout pipelines (without a service) must not have an input")
		}
	}
	return nil
}

func branchProvenance(input *pps.Input) []*pfs.Branch {
	var result []*pfs.Branch
	pps.VisitInput(input, func(input *pps.Input) error {
		if input.Pfs != nil {
			result = append(result, client.NewBranch(input.Pfs.Repo, input.Pfs.Branch))
		}
		if input.Cron != nil {
			result = append(result, client.NewBranch(input.Cron.Repo, "master"))
		}
		if input.Git != nil {
			result = append(result, client.NewBranch(input.Git.Name, input.Git.Branch))
		}
		return nil
	})
	return result
}

var (
	// superUserToken is the cached auth token used by PPS to write to the spec
	// repo, create pipeline subjects, and
	superUserToken string

	// superUserTokenOnce ensures that ppsToken is only read from etcd once. These are
	// read/written by apiServer#sudo()
	superUserTokenOnce sync.Once
)

// sudo is a helper function that copies 'pachClient' grants it PPS's superuser
// token, and calls 'f' with the superuser client. This helps isolate PPS's use
// of its superuser token so that it's not widely copied and is unlikely to
// leak authority to parts of the code that aren't supposed to have it.
//
// Note that because the argument to 'f' is a superuser client, it should not
// be used to make any calls with unvalidated user input. Any such use could be
// exploited to make PPS a confused deputy
func (a *apiServer) sudo(ctx context.Context, f func(*client.APIClient) error) error {
	// Get PPS auth token
	superUserTokenOnce.Do(func() {
		b := backoff.NewExponentialBackOff()
		b.MaxElapsedTime = 60 * time.Second
		b.MaxInterval = 5 * time.Second
		if err := backoff.Retry(func() error {
			superUserTokenCol := col.NewEtcdCollection(a.env.GetEtcdClient(), ppsconsts.PPSTokenKey, nil, &types.StringValue{}, nil, nil).ReadOnly(ctx)
			var result types.StringValue
			if err := superUserTokenCol.Get("", &result); err != nil {
				return err
			}
			superUserToken = result.Value
			return nil
		}, b); err != nil {
			panic(fmt.Sprintf("couldn't get PPS superuser token: %v", err))
		}
	})

	// Copy pach client, but keep ctx (to propagate cancellation). Replace token
	// with superUserToken
	pachClient := a.env.GetPachClient(ctx)
	superUserClient := pachClient.WithCtx(ctx)
	superUserClient.SetAuthToken(superUserToken)
	return f(superUserClient)
}

// sudoTransaction is a convenience wrapper around sudo for api calls in a transaction
func (a *apiServer) sudoTransaction(txnCtx *txncontext.TransactionContext, f func(*txncontext.TransactionContext) error) error {
	return a.sudo(txnCtx.ClientContext, func(superUserClient *client.APIClient) error {
		superCtx := *txnCtx
		// simulate the GRPC setting outgoing as incoming - this should only change the auth token
		outMD, _ := metadata.FromOutgoingContext(superUserClient.Ctx())
		superCtx.ClientContext = metadata.NewIncomingContext(superUserClient.Ctx(), outMD)
		return f(&superCtx)
	})
}

// writePipelineInfo is a helper for CreatePipeline that creates a commit with
// 'pipelineInfo' in SpecRepo (in PFS). It's called in both the case where a
// user is updating a pipeline and the case where a user is creating a new
// pipeline.
// pipelineInfo.SpecCommit is used to ensure that (in the absence of
// transactionality) the new pipelineInfo is only applied on top of the previous
// pipelineInfo - if the head of the spec branch has changed, this operation
// will error out.
func (a *apiServer) writePipelineInfo(ctx context.Context, pipelineInfo *pps.PipelineInfo) (*pfs.Commit, error) {
	filesetID, err := a.writePipelineInfoToFileset(ctx, pipelineInfo)
	if err != nil {
		return nil, err
	}

	var commit *pfs.Commit
	if err := a.txnEnv.WithWriteContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
		var err error
		commit, err = a.commitPipelineInfoFromFileset(txnCtx, pipelineInfo.Pipeline.Name, filesetID, pipelineInfo.SpecCommit)
		return err
	}); err != nil {
		return nil, err
	}

	return commit, nil
}

func (a *apiServer) writePipelineInfoToFileset(ctx context.Context, pipelineInfo *pps.PipelineInfo) (string, error) {
	data, err := pipelineInfo.Marshal()
	if err != nil {
		return "", errors.Wrapf(err, "could not marshal PipelineInfo")
	}
	pachClient := a.env.GetPachClient(ctx)
	resp, err := pachClient.WithCreateFilesetClient(func(mf client.ModifyFile) error {
		err := mf.PutFile(ppsconsts.SpecFile, bytes.NewReader(data))
		return err
	})
	if err != nil {
		return "", err
	}
	return resp.FilesetId, nil
}

func (a *apiServer) commitPipelineInfoFromFileset(
	txnCtx *txncontext.TransactionContext,
	pipelineName string,
	filesetID string,
	prevSpecCommit *pfs.Commit,
) (*pfs.Commit, error) {
	var commit *pfs.Commit
	if err := a.sudoTransaction(txnCtx, func(superCtx *txncontext.TransactionContext) error {
		var err error
		commit, err = a.env.PfsServer().StartCommitInTransaction(superCtx, &pfs.StartCommitRequest{
			Parent: prevSpecCommit,
			Branch: client.NewBranch(ppsconsts.SpecRepo, pipelineName),
		}, nil)
		if err != nil {
			return errors.Wrapf(err, "could not marshal PipelineInfo")
		}

		if err := a.env.PfsServer().AddFilesetInTransaction(superCtx, &pfs.AddFilesetRequest{
			Commit:    commit,
			FilesetId: filesetID,
		}); err != nil {
			return err
		}

		return a.env.PfsServer().FinishCommitInTransaction(superCtx, &pfs.FinishCommitRequest{
			Commit: commit,
		})
	}); err != nil {
		return nil, err
	}

	return commit, nil
}

func (a *apiServer) fixPipelineInputRepoACLs(ctx context.Context, pipelineInfo *pps.PipelineInfo, prevPipelineInfo *pps.PipelineInfo) error {
	return a.txnEnv.WithWriteContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
		return a.fixPipelineInputRepoACLsInTransaction(txnCtx, pipelineInfo, prevPipelineInfo)
	})
}

func (a *apiServer) fixPipelineInputRepoACLsInTransaction(txnCtx *txncontext.TransactionContext, pipelineInfo *pps.PipelineInfo, prevPipelineInfo *pps.PipelineInfo) (retErr error) {
	add := make(map[string]struct{})
	remove := make(map[string]struct{})
	var pipelineName string
	// Figure out which repos 'pipeline' might no longer be using
	if prevPipelineInfo != nil {
		pipelineName = prevPipelineInfo.Pipeline.Name
		pps.VisitInput(prevPipelineInfo.Input, func(input *pps.Input) error {
			var repo string
			switch {
			case input.Pfs != nil:
				repo = input.Pfs.Repo
			case input.Cron != nil:
				repo = input.Cron.Repo
			case input.Git != nil:
				repo = input.Git.Name
			default:
				return nil // no scope to set: input is not a repo
			}
			remove[repo] = struct{}{}
			return nil
		})
	}

	// Figure out which repos 'pipeline' is using
	if pipelineInfo != nil {
		// also check that pipeline name is consistent
		if pipelineName == "" {
			pipelineName = pipelineInfo.Pipeline.Name
		} else if pipelineInfo.Pipeline.Name != pipelineName {
			return errors.Errorf("pipelineInfo (%s) and prevPipelineInfo (%s) do not "+
				"belong to matching pipelines; this is a bug",
				pipelineInfo.Pipeline.Name, prevPipelineInfo.Pipeline.Name)
		}

		// collect inputs (remove redundant inputs from 'remove', but don't
		// bother authorizing 'pipeline' twice)
		pps.VisitInput(pipelineInfo.Input, func(input *pps.Input) error {
			var repo string
			switch {
			case input.Pfs != nil:
				repo = input.Pfs.Repo
			case input.Cron != nil:
				repo = input.Cron.Repo
			case input.Git != nil:
				repo = input.Git.Name
			default:
				return nil // no scope to set: input is not a repo
			}
			if _, ok := remove[repo]; ok {
				delete(remove, repo)
			} else {
				add[repo] = struct{}{}
			}
			return nil
		})
	}
	if pipelineName == "" {
		return errors.Errorf("fixPipelineInputRepoACLs called with both current and " +
			"previous pipelineInfos == to nil; this is a bug")
	}

	// make sure we don't touch the pipeline's permissions on its output repo
	delete(remove, pipelineName)
	delete(add, pipelineName)

	defer func() {
		retErr = errors.Wrapf(retErr, "error fixing ACLs on \"%s\"'s input repos", pipelineName)
	}()

	// Remove pipeline from old, unused inputs
	for repo := range remove {
		// If we get an `ErrNoRoleBinding` that means the input repo no longer exists - we're removing it anyways, so we don't care.
		if err := a.env.AuthServer().RemovePipelineReaderFromRepoInTransaction(txnCtx, repo, pipelineName); err != nil && !auth.IsErrNoRoleBinding(err) {
			return err
		}
	}
	// Add pipeline to every new input's ACL as a READER
	for repo := range add {
		// This raises an error if the input repo doesn't exist, or if the user doesn't have permissions to add a pipeline as a reader on the input repo
		if err := a.env.AuthServer().AddPipelineReaderToRepoInTransaction(txnCtx, repo, pipelineName); err != nil {
			return err
		}
	}
	// Add pipeline to its output repo's ACL as a WRITER if it's new
	if prevPipelineInfo == nil {
		if err := a.env.AuthServer().AddPipelineWriterToRepoInTransaction(txnCtx, pipelineName); err != nil {
			return err
		}
	}
	return nil
}

// getExpectedNumWorkers is a helper function for CreatePipeline that transforms
// the parallelism spec in CreatePipelineRequest.Parallelism into a constant
// that can be stored in StoredPipelineInfo.Parallelism
func getExpectedNumWorkers(kc *kube.Clientset, pipelineInfo *pps.PipelineInfo) (int, error) {
	switch pspec := pipelineInfo.ParallelismSpec; {
	case pspec == nil, pspec.Constant == 0 && pspec.Coefficient == 0:
		return 1, nil
	case pspec.Constant > 0 && pspec.Coefficient == 0:
		return int(pspec.Constant), nil
	case pspec.Constant == 0 && pspec.Coefficient > 0:
		// Start ('coefficient' * 'nodes') workers. Determine number of workers
		nodeList, err := kc.CoreV1().Nodes().List(metav1.ListOptions{})
		if err != nil {
			return 0, errors.Wrapf(err, "unable to retrieve node list from k8s to determine parallelism")
		}
		if len(nodeList.Items) == 0 {
			return 0, errors.Errorf("unable to determine parallelism for %q: no k8s nodes found",
				pipelineInfo.Pipeline.Name)
		}
		numNodes := len(nodeList.Items)
		floatParallelism := math.Floor(pspec.Coefficient * float64(numNodes))
		return int(math.Max(floatParallelism, 1)), nil
	default:
		return 0, errors.Errorf("unable to interpret ParallelismSpec %+v", pspec)
	}
}

// CreatePipeline implements the protobuf pps.CreatePipeline RPC
//
// Implementation note:
// - CreatePipeline always creates pipeline output branches such that the
//   pipeline's spec branch is in the pipeline output branch's provenance
// - CreatePipeline will always create a new output commit, but that's done
//   by CreateBranch at the bottom of the function, which sets the new output
//   branch provenance, rather than commitPipelineInfoFromFileset higher up.
// - This is because CreatePipeline calls hardStopPipeline towards the top,
// 	 breaking the provenance connection from the spec branch to the output branch
// - For straightforward pipeline updates (e.g. new pipeline image)
//   stopping + updating + starting the pipeline isn't necessary
// - However it is necessary in many slightly atypical cases  (e.g. the
//   pipeline input changed: if the spec commit is created while the
//   output branch has its old provenance, or the output branch gets new
//   provenance while the old spec commit is the HEAD of the spec branch,
//   then an output commit will be created with provenance that doesn't
//   match its spec's PipelineInfo.Input. Another example is when
//   request.Reprocess == true).
// - Rather than try to enumerate every case where we can't create a spec
//   commit without stopping the pipeline, we just always stop the pipeline
func (a *apiServer) CreatePipeline(ctx context.Context, request *pps.CreatePipelineRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "CreatePipeline")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	if request.Pipeline == nil {
		return nil, errors.New("request.Pipeline cannot be nil")
	}

	// Annotate current span with pipeline & persist any extended trace to etcd
	span := opentracing.SpanFromContext(ctx)
	tracing.TagAnySpan(span, "pipeline", request.Pipeline.Name)
	defer func() {
		tracing.TagAnySpan(span, "err", retErr)
	}()
	extended.PersistAny(ctx, a.env.GetEtcdClient(), request.Pipeline.Name)

	// Don't provide a fileset and the transaction env will generate it
	filesetID := ""
	var prevSpecCommit *pfs.Commit
	if err := a.txnEnv.WithTransaction(ctx, func(txn txnenv.Transaction) error {
		return txn.CreatePipeline(request, &filesetID, &prevSpecCommit)
	}); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

func (a *apiServer) initializePipelineInfo(request *pps.CreatePipelineRequest, oldPipelineInfo *pps.PipelineInfo) (*pps.PipelineInfo, error) {
	if err := a.validatePipelineRequest(request); err != nil {
		return nil, err
	}

	// Reprocess overrides the salt in the request
	if request.Salt == "" || request.Reprocess {
		request.Salt = uuid.NewWithoutDashes()
	}

	pipelineInfo := &pps.PipelineInfo{
		Pipeline:              request.Pipeline,
		Version:               1,
		Transform:             request.Transform,
		TFJob:                 request.TFJob,
		ParallelismSpec:       request.ParallelismSpec,
		Input:                 request.Input,
		OutputBranch:          request.OutputBranch,
		Egress:                request.Egress,
		CreatedAt:             now(),
		ResourceRequests:      request.ResourceRequests,
		ResourceLimits:        request.ResourceLimits,
		SidecarResourceLimits: request.SidecarResourceLimits,
		Description:           request.Description,
		CacheSize:             request.CacheSize,
		EnableStats:           request.EnableStats,
		Salt:                  request.Salt,
		MaxQueueSize:          request.MaxQueueSize,
		Service:               request.Service,
		Spout:                 request.Spout,
		ChunkSpec:             request.ChunkSpec,
		DatumTimeout:          request.DatumTimeout,
		JobTimeout:            request.JobTimeout,
		Standby:               request.Standby,
		DatumTries:            request.DatumTries,
		SchedulingSpec:        request.SchedulingSpec,
		PodSpec:               request.PodSpec,
		PodPatch:              request.PodPatch,
		S3Out:                 request.S3Out,
		Metadata:              request.Metadata,
		ReprocessSpec:         request.ReprocessSpec,
	}

	if err := setPipelineDefaults(pipelineInfo); err != nil {
		return nil, err
	}
	// Validate final PipelineInfo (now that defaults have been populated)
	if err := a.validatePipeline(pipelineInfo); err != nil {
		return nil, err
	}

	pps.SortInput(pipelineInfo.Input) // Makes datum hashes comparable

	if oldPipelineInfo != nil {
		// Modify pipelineInfo (increment Version, and *preserve Stopped* so
		// that updating a pipeline doesn't restart it)
		pipelineInfo.Version = oldPipelineInfo.Version + 1
		if oldPipelineInfo.Stopped {
			pipelineInfo.Stopped = true
		}
		if !request.Reprocess {
			pipelineInfo.Salt = oldPipelineInfo.Salt
		}
		// Cannot disable stats after it has been enabled.
		if oldPipelineInfo.EnableStats && !pipelineInfo.EnableStats {
			return nil, newErrPipelineUpdate(pipelineInfo.Pipeline.Name, "cannot disable stats")
		}
	}

	return pipelineInfo, nil
}

// latestPipelineInfo doesn't actually work transactionally because
// ppsutil.GetPipelineInfo needs to use pfs.GetFile to read the spec commit,
// which does not support transactions.
func (a *apiServer) latestPipelineInfo(txnCtx *txncontext.TransactionContext, pipelineName string) (*pps.PipelineInfo, error) {
	pipelinePtr := pps.StoredPipelineInfo{}
	if err := a.pipelines.ReadWrite(txnCtx.SqlTx).Get(pipelineName, &pipelinePtr); err != nil {
		if col.IsErrNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	// the spec commit must already exist outside of the transaction, so we can retrieve it normally
	return ppsutil.GetPipelineInfo(a.env.GetPachClient(txnCtx.ClientContext), &pipelinePtr)
}

func (a *apiServer) pipelineInfosForUpdate(txnCtx *txncontext.TransactionContext, request *pps.CreatePipelineRequest) (*pps.PipelineInfo, *pps.PipelineInfo, error) {
	if request.Pipeline == nil {
		return nil, nil, errors.New("invalid pipeline spec: request.Pipeline cannot be nil")
	}

	// We can't recover from an incomplete pipeline info here because
	// modifying the spec repo depends on being able to access the previous
	// commit. We therefore use `GetPipelineInfo` which will error if the
	// spec commit isn't working.
	oldPipelineInfo, err := a.latestPipelineInfo(txnCtx, request.Pipeline.Name)
	if err != nil {
		return nil, nil, err
	}

	newPipelineInfo, err := a.initializePipelineInfo(request, oldPipelineInfo)
	if err != nil {
		return nil, nil, err
	}

	return oldPipelineInfo, newPipelineInfo, nil
}

func (a *apiServer) CreatePipelineInTransaction(
	txnCtx *txncontext.TransactionContext,
	request *pps.CreatePipelineRequest,
	specFilesetID *string,
	prevSpecCommit **pfs.Commit,
) error {
	oldPipelineInfo, newPipelineInfo, err := a.pipelineInfosForUpdate(txnCtx, request)
	if err != nil {
		return err
	}
	pipelineName := request.Pipeline.Name

	if *specFilesetID != "" {
		// If we already have a fileset, try to renew it - if that fails, invalidate it
		if err := a.env.GetPachClient(txnCtx.ClientContext).RenewFileSet(*specFilesetID, 600*time.Second); err != nil {
			*specFilesetID = ""
		}
	}

	// If the expected spec commit doesn't match up with oldPipelineInfo, we need
	// to recreate the fileset
	staleFileset := false
	if oldPipelineInfo == nil {
		staleFileset = (*prevSpecCommit != nil)
	} else {
		staleFileset = !proto.Equal(oldPipelineInfo.SpecCommit, *prevSpecCommit)
	}

	if staleFileset || *specFilesetID == "" {
		// No existing fileset or the old one expired, create a new fileset - the
		// pipeline spec to be written into a fileset outside of the transaction.
		*specFilesetID, err = a.writePipelineInfoToFileset(txnCtx.ClientContext, newPipelineInfo)
		if err != nil {
			return err
		}
		if oldPipelineInfo != nil {
			*prevSpecCommit = oldPipelineInfo.SpecCommit
		}

		// The transaction cannot continue because it cannot see the fileset - abort and retry
		return &col.ErrTransactionConflict{}
	}

	// Verify that all input repos exist (create cron and git repos if necessary)
	if visitErr := pps.VisitInput(newPipelineInfo.Input, func(input *pps.Input) error {
		if input.Pfs != nil {
			if _, err := a.env.PfsServer().InspectRepoInTransaction(txnCtx,
				&pfs.InspectRepoRequest{
					Repo: client.NewSystemRepo(input.Pfs.Repo, input.Pfs.RepoType),
				},
			); err != nil {
				return err
			}
		}
		if input.Cron != nil {
			if err := a.env.PfsServer().CreateRepoInTransaction(txnCtx,
				&pfs.CreateRepoRequest{
					Repo:        client.NewRepo(input.Cron.Repo),
					Description: fmt.Sprintf("Cron tick repo for pipeline %s.", request.Pipeline.Name),
				},
			); err != nil && !isAlreadyExistsErr(err) {
				return err
			}
		}
		if input.Git != nil {
			if err := a.env.PfsServer().CreateRepoInTransaction(txnCtx,
				&pfs.CreateRepoRequest{
					Repo:        client.NewRepo(input.Git.Name),
					Description: fmt.Sprintf("Git input repo for pipeline %s.", request.Pipeline.Name),
				},
			); err != nil && !isAlreadyExistsErr(err) {
				return err
			}
		}
		return nil
	}); visitErr != nil {
		return visitErr
	}

	// Authorize pipeline creation
	operation := pipelineOpCreate
	if request.Update {
		operation = pipelineOpUpdate
	}
	if err := a.authorizePipelineOpInTransaction(txnCtx, operation, newPipelineInfo.Input, newPipelineInfo.Pipeline.Name); err != nil {
		return err
	}
	update := false
	if request.Update {
		// inspect the pipeline to see if this is a real update
		if _, err := a.inspectPipelineInTransaction(txnCtx, request.Pipeline.Name); err == nil {
			update = true
		}
	}
	var (
		// provenance for the pipeline's output branch (includes the spec branch)
		provenance = append(branchProvenance(newPipelineInfo.Input),
			client.NewBranch(ppsconsts.SpecRepo, pipelineName))
		outputBranch     = client.NewBranch(pipelineName, newPipelineInfo.OutputBranch)
		statsBranch      = client.NewSystemRepo(pipelineName, pfs.MetaRepoType).NewBranch("master")
		outputBranchHead *pfs.Commit
		statsBranchHead  *pfs.Commit
		specCommit       *pfs.Commit
	)

	// Get the expected number of workers for this pipeline
	parallelism, err := getExpectedNumWorkers(a.env.GetKubeClient(), newPipelineInfo)
	if err != nil {
		return err
	}

	if update {
		// Help user fix inconsistency if previous UpdatePipeline call failed
		if ci, err := a.env.PfsServer().InspectCommitInTransaction(txnCtx, &pfs.InspectCommitRequest{
			Commit: client.NewCommit(ppsconsts.SpecRepo, pipelineName, ""),
		}); err != nil {
			return err
		} else if ci.Finished == nil {
			return errors.Errorf("the HEAD commit of this pipeline's spec branch " +
				"is open. Either another CreatePipeline call is running or a previous " +
				"call crashed. If you're sure no other CreatePipeline commands are " +
				"running, you can run 'pachctl update pipeline --clean' which will " +
				"delete this open commit")
		}
	}

	if request.SpecCommit != nil {
		if update {
			// request.SpecCommit indicates we're restoring from an extracted cluster
			// state and should not do any updates
			return errors.New("Cannot update a pipeline and provide a spec commit at the same time")
		}
		commitInfo, err := a.env.PfsServer().InspectCommitInTransaction(txnCtx, &pfs.InspectCommitRequest{
			Commit: request.SpecCommit,
		})
		if err != nil {
			return errors.Wrap(err, "error inspecting spec commit")
		}
		// It does, so we use that as the spec commit, rather than making a new one
		specCommit = commitInfo.Commit
		// We also use the existing head for the branches, rather than making a new one.
		outputBranchHead = client.NewCommit(pipelineName, newPipelineInfo.OutputBranch, "")
		statsBranchHead = client.NewSystemRepo(pipelineName, pfs.MetaRepoType).NewCommit("master", "")
	} else {
		specCommit, err = a.commitPipelineInfoFromFileset(txnCtx, pipelineName, *specFilesetID, *prevSpecCommit)
		if err != nil {
			return err
		}
	}

	if update {
		// finish any open output commits at the end of the transaction
		txnCtx.FinishPipelineCommits(client.NewBranch(pipelineName, newPipelineInfo.OutputBranch))

		// Update the existing StoredPipelineInfo in postgres
		pipelinePtr := &pps.StoredPipelineInfo{}
		if err := a.pipelines.ReadWrite(txnCtx.SqlTx).Update(pipelineName, pipelinePtr, func() error {
			if oldPipelineInfo.Stopped {
				provenance = nil // CreateBranch() below shouldn't create new output
			}

			// Update pipelinePtr to point to new commit
			pipelinePtr.SpecCommit = specCommit
			// Reset pipeline state (PPS master/pipeline controller recreates RC)
			pipelinePtr.State = pps.PipelineState_PIPELINE_STARTING
			// Clear any failure reasons
			pipelinePtr.Reason = ""
			// Update pipeline parallelism
			pipelinePtr.Parallelism = uint64(parallelism)

			// Generate new pipeline auth token (added due to & add pipeline to the ACLs of input/output repos
			if err := func() error {
				oldAuthToken := pipelinePtr.AuthToken
				token, err := a.env.AuthServer().GetPipelineAuthTokenInTransaction(txnCtx, request.Pipeline.Name)
				if err != nil {
					if auth.IsErrNotActivated(err) {
						return nil // no auth work to do
					}
					return grpcutil.ScrubGRPC(err)
				}
				pipelinePtr.AuthToken = token

				// If getting a new auth token worked, we should revoke the old one
				if oldAuthToken != "" {
					_, err := a.env.AuthServer().RevokeAuthTokenInTransaction(txnCtx,
						&auth.RevokeAuthTokenRequest{
							Token: oldAuthToken,
						})
					if err != nil {
						if auth.IsErrNotActivated(err) {
							return nil // no auth work to do
						}
						return grpcutil.ScrubGRPC(err)
					}
				}
				return nil
			}(); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return err
		}

		if !request.Reprocess {
			// don't branch the output/stats commit chain from the old pipeline (re-use old branch HEAD)
			// However it's valid to set request.Update == true even if no pipeline exists, so only
			// set outputBranchHead if there's an old pipeline to update
			_, err := a.env.PfsServer().InspectBranchInTransaction(txnCtx, &pfs.InspectBranchRequest{Branch: outputBranch})
			if err != nil && !isNotFoundErr(err) {
				return err
			} else if err == nil {
				outputBranchHead = client.NewCommit(pipelineName, newPipelineInfo.OutputBranch, "")
			}

			_, err = a.env.PfsServer().InspectBranchInTransaction(txnCtx, &pfs.InspectBranchRequest{Branch: statsBranch})
			if err != nil && !isNotFoundErr(err) {
				return err
			} else if err == nil {
				statsBranchHead = client.NewSystemRepo(pipelineName, pfs.MetaRepoType).NewCommit("master", "")
			}
		}

		if pipelinePtr.AuthToken != "" {
			if err := a.fixPipelineInputRepoACLsInTransaction(txnCtx, newPipelineInfo, oldPipelineInfo); err != nil {
				return err
			}
		}
	} else {
		// Create output repo, pipeline output, and stats
		if err := a.env.PfsServer().CreateRepoInTransaction(txnCtx,
			&pfs.CreateRepoRequest{
				Repo:        client.NewRepo(pipelineName),
				Description: fmt.Sprintf("Output repo for pipeline %s.", request.Pipeline.Name),
			}); err != nil && !isAlreadyExistsErr(err) {
			return errors.Wrapf(err, "error creating output repo for %s", pipelineName)

		}
		// pipelinePtr will be written to etcd, pointing at 'commit'. May include an
		// auth token
		pipelinePtr := &pps.StoredPipelineInfo{
			Pipeline:    request.Pipeline,
			SpecCommit:  specCommit,
			State:       pps.PipelineState_PIPELINE_STARTING,
			Parallelism: uint64(parallelism),
		}

		// Generate pipeline's auth token & add pipeline to the ACLs of input/output
		// repos

		if err := func() error {
			token, err := a.env.AuthServer().GetPipelineAuthTokenInTransaction(txnCtx, request.Pipeline.Name)
			if err != nil {
				if auth.IsErrNotActivated(err) {
					return nil // no auth work to do
				}
				return grpcutil.ScrubGRPC(err)
			}

			pipelinePtr.AuthToken = token
			return nil
		}(); err != nil {
			return err
		}
		// Put a pointer to the new PipelineInfo commit into etcd
		err := a.pipelines.ReadWrite(txnCtx.SqlTx).Create(pipelineName, pipelinePtr)
		if isAlreadyExistsErr(err) {
			return newErrPipelineExists(pipelineName)
		} else if err != nil {
			return err
		}

		if pipelinePtr.AuthToken != "" {
			if err := a.fixPipelineInputRepoACLsInTransaction(txnCtx, newPipelineInfo, nil); err != nil {
				return err
			}
		}
	}

	// A stopped pipeline should not have its provenance restored until it is
	// restarted.
	if oldPipelineInfo != nil && oldPipelineInfo.Stopped {
		provenance = nil
	}

	// Create/update output branch (creating new output commit for the pipeline
	// and restarting the pipeline)
	if err := a.env.PfsServer().CreateBranchInTransaction(txnCtx, &pfs.CreateBranchRequest{
		Branch:     outputBranch,
		Provenance: provenance,
		Head:       outputBranchHead,
	}); err != nil {
		return errors.Wrapf(err, "could not create/update output branch")
	}
	if visitErr := pps.VisitInput(request.Input, func(input *pps.Input) error {
		if input.Pfs != nil && input.Pfs.Trigger != nil {
			_, err = a.env.PfsServer().InspectBranchInTransaction(txnCtx, &pfs.InspectBranchRequest{
				Branch: client.NewBranch(input.Pfs.Repo, input.Pfs.Branch),
			})

			if err != nil && !isNotFoundErr(err) {
				return err
			} else {
				var prevHead *pfs.Commit
				if err == nil {
					prevHead = client.NewCommit(input.Pfs.Repo, input.Pfs.Branch, "")
				}
				return a.env.PfsServer().CreateBranchInTransaction(txnCtx, &pfs.CreateBranchRequest{
					Branch:  client.NewBranch(input.Pfs.Repo, input.Pfs.Branch),
					Head:    prevHead,
					Trigger: input.Pfs.Trigger,
				})
			}
		}
		return nil
	}); visitErr != nil {
		return errors.Wrapf(visitErr, "could not create/update trigger branch")
	}
	if newPipelineInfo.EnableStats {
		if err := a.env.PfsServer().CreateRepoInTransaction(txnCtx, &pfs.CreateRepoRequest{
			Repo:        statsBranch.Repo,
			Description: fmt.Sprint("Meta repo for", pipelineName),
			Update:      true, // don't error if it already exists
		}); err != nil {
			return errors.Wrap(err, "could not create meta repo")
		}
		if err := a.env.PfsServer().CreateBranchInTransaction(txnCtx, &pfs.CreateBranchRequest{
			Branch:     statsBranch,
			Provenance: []*pfs.Branch{outputBranch},
			Head:       statsBranchHead,
		}); err != nil {
			return errors.Wrapf(err, "could not create/update meta branch")
		}
	}
	return nil
}

// setPipelineDefaults sets the default values for a pipeline info
func setPipelineDefaults(pipelineInfo *pps.PipelineInfo) error {
	if pipelineInfo.Transform.Image == "" {
		pipelineInfo.Transform.Image = DefaultUserImage
	}
	setInputDefaults(pipelineInfo.Pipeline.Name, pipelineInfo.Input)
	if pipelineInfo.OutputBranch == "" {
		// Output branches default to master
		pipelineInfo.OutputBranch = "master"
	}
	if pipelineInfo.CacheSize == "" {
		pipelineInfo.CacheSize = "64M"
	}
	if pipelineInfo.MaxQueueSize < 1 {
		pipelineInfo.MaxQueueSize = 1
	}
	if pipelineInfo.DatumTries == 0 {
		pipelineInfo.DatumTries = DefaultDatumTries
	}
	if pipelineInfo.Service != nil {
		if pipelineInfo.Service.Type == "" {
			pipelineInfo.Service.Type = string(v1.ServiceTypeNodePort)
		}
	}
	if pipelineInfo.Spout != nil && pipelineInfo.Spout.Service != nil && pipelineInfo.Spout.Service.Type == "" {
		pipelineInfo.Spout.Service.Type = string(v1.ServiceTypeNodePort)
	}
	if pipelineInfo.ReprocessSpec == "" {
		pipelineInfo.ReprocessSpec = client.ReprocessSpecUntilSuccess
	}
	return nil
}

func setInputDefaults(pipelineName string, input *pps.Input) {
	now := time.Now()
	nCreatedBranches := make(map[string]int)
	pps.VisitInput(input, func(input *pps.Input) error {
		if input.Pfs != nil {
			if input.Pfs.Branch == "" {
				if input.Pfs.Trigger != nil {
					// We start counting trigger branches at 1
					nCreatedBranches[input.Pfs.Repo]++
					input.Pfs.Branch = fmt.Sprintf("%s-trigger-%d", pipelineName, nCreatedBranches[input.Pfs.Repo])
					if input.Pfs.Trigger.Branch == "" {
						input.Pfs.Trigger.Branch = "master"
					}
				} else {
					input.Pfs.Branch = "master"
				}
			}
			if input.Pfs.Name == "" {
				input.Pfs.Name = input.Pfs.Repo
			}
			if input.Pfs.RepoType == "" {
				input.Pfs.RepoType = pfs.UserRepoType
			}
		}
		if input.Cron != nil {
			if input.Cron.Start == nil {
				start, _ := types.TimestampProto(now)
				input.Cron.Start = start
			}
			if input.Cron.Repo == "" {
				input.Cron.Repo = fmt.Sprintf("%s_%s", pipelineName, input.Cron.Name)
			}
		}
		if input.Git != nil {
			if input.Git.Branch == "" {
				input.Git.Branch = "master"
			}
			if input.Git.Name == "" {
				// We know URL looks like:
				// "https://github.com/sjezewski/testgithook.git",
				tokens := strings.Split(path.Base(input.Git.URL), ".")
				input.Git.Name = tokens[0]
			}
		}
		return nil
	})
}

// InspectPipeline implements the protobuf pps.InspectPipeline RPC
func (a *apiServer) InspectPipeline(ctx context.Context, request *pps.InspectPipelineRequest) (response *pps.PipelineInfo, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	return a.inspectPipeline(ctx, request.Pipeline.Name)
}

// inspectPipeline contains the functional implementation of InspectPipeline.
// Many functions (GetLogs, ListPipeline, CreatePipelineJob) need to inspect a
// pipeline, so they call this instead of making an RPC
func (a *apiServer) inspectPipeline(ctx context.Context, name string) (*pps.PipelineInfo, error) {
	var response *pps.PipelineInfo
	if err := a.txnEnv.WithReadContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
		var err error
		response, err = a.inspectPipelineInTransaction(txnCtx, name)
		return err
	}); err != nil {
		return nil, err
	}
	return response, nil
}

func (a *apiServer) inspectPipelineInTransaction(txnCtx *txncontext.TransactionContext, name string) (*pps.PipelineInfo, error) {
	kubeClient := a.env.GetKubeClient()
	name, ancestors, err := ancestry.Parse(name)
	if err != nil {
		return nil, err
	}
	pipelinePtr := pps.StoredPipelineInfo{}
	if err := a.pipelines.ReadWrite(txnCtx.SqlTx).Get(name, &pipelinePtr); err != nil {
		if col.IsErrNotFound(err) {
			return nil, errors.Errorf("pipeline \"%s\" not found", name)
		}
		return nil, err
	}
	pipelinePtr.SpecCommit.ID = ancestry.Add(pipelinePtr.SpecCommit.ID, ancestors)
	// the spec commit must already exist outside of the transaction, so we can retrieve it normally
	pachClient := a.env.GetPachClient(txnCtx.ClientContext)
	pipelineInfo, err := ppsutil.GetPipelineInfo(pachClient, &pipelinePtr)
	if err != nil {
		return nil, err
	}
	if pipelineInfo.Service != nil {
		rcName := ppsutil.PipelineRcName(pipelineInfo.Pipeline.Name, pipelineInfo.Version)
		if err != nil {
			return nil, err
		}
		service, err := kubeClient.CoreV1().Services(a.namespace).Get(fmt.Sprintf("%s-user", rcName), metav1.GetOptions{})
		if err != nil {
			if !isNotFoundErr(err) {
				return nil, err
			}
		} else {
			pipelineInfo.Service.IP = service.Spec.ClusterIP
		}
	}
	var hasGitInput bool
	pps.VisitInput(pipelineInfo.Input, func(input *pps.Input) error {
		if input.Git != nil {
			hasGitInput = true
			return errutil.ErrBreak
		}
		return nil
	})
	if hasGitInput {
		pipelineInfo.GithookURL = "pending"
		svc, err := getGithookService(kubeClient, a.namespace)
		if err != nil {
			return pipelineInfo, nil
		}
		numIPs := len(svc.Status.LoadBalancer.Ingress)
		if numIPs == 0 {
			// When running locally, no external IP is set
			return pipelineInfo, nil
		}
		if numIPs != 1 {
			return nil, errors.Errorf("unexpected number of external IPs set for githook service")
		}
		ingress := svc.Status.LoadBalancer.Ingress[0]
		if ingress.IP != "" {
			// GKE load balancing
			pipelineInfo.GithookURL = githook.URLFromDomain(ingress.IP)
		} else if ingress.Hostname != "" {
			// AWS load balancing
			pipelineInfo.GithookURL = githook.URLFromDomain(ingress.Hostname)
		}
	}

	workerPoolID := ppsutil.PipelineRcName(pipelineInfo.Pipeline.Name, pipelineInfo.Version)
	workerStatus, err := workerserver.Status(txnCtx.ClientContext, workerPoolID, a.env.GetEtcdClient(), a.etcdPrefix, a.workerGrpcPort)
	if err != nil {
		logrus.Errorf("failed to get worker status with err: %s", err.Error())
	} else {
		pipelineInfo.WorkersAvailable = int64(len(workerStatus))
		pipelineInfo.WorkersRequested = int64(pipelinePtr.Parallelism)
	}
	return pipelineInfo, nil
}

// ListPipeline implements the protobuf pps.ListPipeline RPC
func (a *apiServer) ListPipeline(ctx context.Context, request *pps.ListPipelineRequest) (response *pps.PipelineInfos, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) {
		if response != nil && len(response.PipelineInfo) > client.MaxListItemsLog {
			logrus.Infof("Response contains %d objects; logging the first %d", len(response.PipelineInfo), client.MaxListItemsLog)
			a.Log(request, &pps.PipelineInfos{PipelineInfo: response.PipelineInfo[:client.MaxListItemsLog]}, retErr, time.Since(start))
		} else {
			a.Log(request, response, retErr, time.Since(start))
		}
	}(time.Now())
	pipelineInfos := &pps.PipelineInfos{}
	if err := a.listPipeline(ctx, request, func(pi *pps.PipelineInfo) error {
		pipelineInfos.PipelineInfo = append(pipelineInfos.PipelineInfo, pi)
		return nil
	}); err != nil {
		return nil, err
	}
	return pipelineInfos, nil
}

func (a *apiServer) listPipeline(ctx context.Context, request *pps.ListPipelineRequest, f func(*pps.PipelineInfo) error) error {
	var jqCode *gojq.Code
	var enc serde.Encoder
	var jsonBuffer bytes.Buffer
	if request.JqFilter != "" {
		jqQuery, err := gojq.Parse(request.JqFilter)
		if err != nil {
			return err
		}
		jqCode, err = gojq.Compile(jqQuery)
		if err != nil {
			return err
		}
		// ensure field names and enum values match with --raw output
		enc = serde.NewJSONEncoder(&jsonBuffer, serde.WithOrigName(true))
	}
	filterPipeline := func(pipelineInfo *pps.PipelineInfo) bool {
		if jqCode != nil {
			jsonBuffer.Reset()
			// convert pipelineInfo to a map[string]interface{} for use with gojq
			enc.EncodeProto(pipelineInfo)
			var pipelineInterface interface{}
			json.Unmarshal(jsonBuffer.Bytes(), &pipelineInterface)
			iter := jqCode.Run(pipelineInterface)
			// treat either jq false-y value as rejection
			if v, _ := iter.Next(); v == false || v == nil {
				return false
			}
		}
		return true
	}
	// the mess below is so we can lookup the PFS info for each pipeline concurrently.
	eg, ctx := errgroup.WithContext(ctx)
	infos := make(chan *pps.StoredPipelineInfo)
	// stream these out of etcd
	eg.Go(func() error {
		defer close(infos)
		return a.listPipelinePtr(ctx, request.Pipeline, request.History, func(ptr *pps.StoredPipelineInfo) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case infos <- proto.Clone(ptr).(*pps.StoredPipelineInfo):
				return nil
			}
		})
	})
	// spin up goroutines to get the PFS info, and then use the mutex to call f all synchronized like.
	var mu sync.Mutex
	var fHasErrored bool
	for i := 0; i < 20; i++ {
		eg.Go(func() error {
			for info := range infos {
				pinfo, err := a.resolvePipelineInfo(ctx, request.AllowIncomplete, info)
				if err != nil {
					return err
				}
				if err := func() error {
					mu.Lock()
					defer mu.Unlock()
					if fHasErrored {
						return nil
					}
					// the filtering shares that buffer thing, and it's CPU bound so why not do it with the lock
					if !filterPipeline(pinfo) {
						return nil
					}
					if err := f(pinfo); err != nil {
						fHasErrored = true
						return err
					}
					return nil
				}(); err != nil {
					return err
				}
			}
			return nil
		})
	}
	return eg.Wait()
}

// resolvePipelineInfo looks up additional pipeline info in PFS needed to turn a StoredPipelineInfo into a PipelineInfo
func (a *apiServer) resolvePipelineInfo(ctx context.Context, allowIncomplete bool, ptr *pps.StoredPipelineInfo) (*pps.PipelineInfo, error) {
	pachClient := a.env.GetPachClient(ctx)
	if allowIncomplete {
		return ppsutil.GetPipelineInfoAllowIncomplete(pachClient, ptr)
	}
	return ppsutil.GetPipelineInfo(pachClient, ptr)
}

// listPipelinePtr enumerates all PPS pipelines in etcd, filters them based on
// 'request', and then calls 'f' on each value
func (a *apiServer) listPipelinePtr(ctx context.Context,
	pipeline *pps.Pipeline, history int64, f func(*pps.StoredPipelineInfo) error) error {
	p := &pps.StoredPipelineInfo{}
	forEachPipeline := func() error {
		for i := int64(0); ; i++ {
			// call f() if i <= history (esp. if history == 0, call f() once)
			if err := f(p); err != nil {
				return err
			}
			// however, only call InspectCommit if i < history (i.e. don't call it on
			// the last iteration, and if history == 0, don't call it at all)
			if history >= 0 && i >= history {
				return nil
			}
			// Get parent commit
			pachClient := a.env.GetPachClient(ctx)
			ci, err := pachClient.InspectCommit(ppsconsts.SpecRepo, p.SpecCommit.Branch.Name, p.SpecCommit.ID)
			switch {
			case err != nil:
				return err
			case ci.ParentCommit == nil:
				return nil
			default:
				p.SpecCommit = ci.ParentCommit
			}
		}
	}
	if pipeline == nil {
		if err := a.pipelines.ReadOnly(ctx).List(p, col.DefaultOptions(), func(string) error {
			return forEachPipeline()
		}); err != nil {
			return err
		}
	} else {
		if err := a.pipelines.ReadOnly(ctx).Get(pipeline.Name, p); err != nil {
			if col.IsErrNotFound(err) {
				return errors.Errorf("pipeline \"%s\" not found", pipeline.Name)
			}
			return err
		}
		if err := forEachPipeline(); err != nil {
			return err
		}
	}
	return nil
}

// DeletePipeline implements the protobuf pps.DeletePipeline RPC
func (a *apiServer) DeletePipeline(ctx context.Context, request *pps.DeletePipelineRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	// Possibly list pipelines in etcd (skip PFS read--don't need it) and delete them
	if request.All {
		request.Pipeline = &pps.Pipeline{}
		pipelinePtr := &pps.StoredPipelineInfo{}
		if err := a.pipelines.ReadOnly(ctx).List(pipelinePtr, col.DefaultOptions(), func(string) error {
			request.Pipeline.Name = pipelinePtr.Pipeline.Name
			_, err := a.deletePipeline(ctx, request)
			return err
		}); err != nil {
			return nil, err
		}
		return &types.Empty{}, nil
	}

	// Otherwise delete single pipeline from request
	return a.deletePipeline(ctx, request)
}

// cleanUpSpecBranch handles the corner case where a spec branch was created for
// a new pipeline, but the etcdPipelineInfo was never created successfully (and
// the pipeline is in an inconsistent state). It's called if a pipeline's
// etcdPipelineInfo wasn't found, checks if an orphaned branch exists, and if
// so, deletes the orphaned branch.
func (a *apiServer) cleanUpSpecBranch(ctx context.Context, pipeline string) error {
	pachClient := a.env.GetPachClient(ctx)
	specBranchInfo, err := pachClient.InspectBranch(ppsconsts.SpecRepo, pipeline)
	if err != nil && (specBranchInfo != nil && specBranchInfo.Head != nil) {
		// No spec branch (and no etcd pointer) => the pipeline doesn't exist
		return errors.Wrapf(err, "pipeline %v was not found", pipeline)
	}
	// branch exists but head is nil => pipeline creation never finished/
	// pps state is invalid. Delete nil branch
	return grpcutil.ScrubGRPC(a.sudo(ctx, func(superUserClient *client.APIClient) error {
		return superUserClient.DeleteBranch(ppsconsts.SpecRepo, pipeline, true)
	}))
}

func (a *apiServer) deletePipeline(ctx context.Context, request *pps.DeletePipelineRequest) (response *types.Empty, retErr error) {
	// Check if there's an StoredPipelineInfo for this pipeline. If not, we can't
	// authorize, and must return something here
	pipelinePtr := pps.StoredPipelineInfo{}
	if err := a.pipelines.ReadOnly(ctx).Get(request.Pipeline.Name, &pipelinePtr); err != nil {
		if col.IsErrNotFound(err) {
			if err := a.cleanUpSpecBranch(ctx, request.Pipeline.Name); err != nil {
				return nil, err
			}
			return &types.Empty{}, nil

		}
		return nil, err
	}

	// Get current pipeline info from:
	// - etcdPipelineInfo
	// - spec commit in etcdPipelineInfo (which may not be the HEAD of the
	//   pipeline's spec branch)
	// - kubernetes services (for service pipelines, githook pipelines, etc)
	pipelineInfo, err := a.inspectPipeline(ctx, request.Pipeline.Name)
	if err != nil {
		logrus.Errorf("error inspecting pipeline: %v", err)
		pipelineInfo = &pps.PipelineInfo{Pipeline: request.Pipeline, OutputBranch: "master"}
	}

	pachClient := a.env.GetPachClient(ctx)

	// check if the output repo exists--if not, the pipeline is non-functional and
	// the rest of the delete operation continues without any auth checks
	if _, err := pachClient.InspectRepo(request.Pipeline.Name); err != nil && !isNotFoundErr(err) && !auth.IsErrNoRoleBinding(err) {
		return nil, err
	} else if err == nil {
		// Check if the caller is authorized to delete this pipeline. This must be
		// done after cleaning up the spec branch HEAD commit, because the
		// authorization condition depends on the pipeline's PipelineInfo
		if err := a.authorizePipelineOp(ctx, pipelineOpDelete, pipelineInfo.Input, pipelineInfo.Pipeline.Name); err != nil {
			return nil, err
		}
		if request.KeepRepo {
			// Remove branch provenance (pass branch twice so that it continues to point
			// at the same commit, but also pass empty provenance slice)
			if err := pachClient.CreateBranch(
				request.Pipeline.Name,
				pipelineInfo.OutputBranch,
				pipelineInfo.OutputBranch,
				"",
				nil,
			); err != nil {
				return nil, err
			}
		} else {
			// delete the pipeline's output repo
			if err := pachClient.DeleteRepo(request.Pipeline.Name, request.Force); err != nil {
				return nil, err
			}
		}
	}

	// If necessary, revoke the pipeline's auth token and remove it from its
	// inputs' ACLs
	if pipelinePtr.AuthToken != "" {
		// If auth was deactivated after the pipeline was created, don't bother
		// revoking
		if _, err := pachClient.WhoAmI(pachClient.Ctx(), &auth.WhoAmIRequest{}); err == nil {
			// 'pipelineInfo' == nil => remove pipeline from all input repos
			if err := a.fixPipelineInputRepoACLs(ctx, nil, pipelineInfo); err != nil {
				return nil, grpcutil.ScrubGRPC(err)
			}
			if _, err := pachClient.RevokeAuthToken(pachClient.Ctx(),
				&auth.RevokeAuthTokenRequest{
					Token: pipelinePtr.AuthToken,
				}); err != nil {

				return nil, grpcutil.ScrubGRPC(err)
			}
		}
	}

	// Kill or delete all of the pipeline's jobs
	// TODO(msteffen): a job may be created by the worker master after this step
	// but before the pipeline RC is deleted. Check for orphaned jobs in
	// pollPipelines.
	var eg errgroup.Group
	pipelinJobPtr := &pps.StoredPipelineJobInfo{}
	if err := a.pipelineJobs.ReadOnly(ctx).GetByIndex(ppsdb.PipelineJobsPipelineIndex, request.Pipeline.Name, pipelinJobPtr, col.DefaultOptions(), func(pipelineJobID string) error {
		eg.Go(func() error {
			_, err := a.DeletePipelineJob(ctx, &pps.DeletePipelineJobRequest{PipelineJob: client.NewPipelineJob(pipelineJobID)})
			if isNotFoundErr(err) || auth.IsErrNoRoleBinding(err) {
				return nil
			}
			return err
		})
		return nil
	}); err != nil {
		return nil, err
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	eg = errgroup.Group{}
	// Delete pipeline branch in SpecRepo (leave commits, to preserve downstream
	// commits)
	eg.Go(func() error {
		return a.sudo(ctx, func(superUserClient *client.APIClient) error {
			return grpcutil.ScrubGRPC(superUserClient.DeleteBranch(ppsconsts.SpecRepo, request.Pipeline.Name, request.Force))
		})
	})
	// Delete cron input repos
	if !request.KeepRepo {
		pps.VisitInput(pipelineInfo.Input, func(input *pps.Input) error {
			if input.Cron != nil {
				eg.Go(func() error {
					return pachClient.DeleteRepo(input.Cron.Repo, request.Force)
				})
			}
			return nil
		})
	}
	// Delete StoredPipelineInfo
	eg.Go(func() error {
		if err := col.NewSQLTx(ctx, a.env.GetDBClient(), func(sqlTx *sqlx.Tx) error {
			return a.pipelines.ReadWrite(sqlTx).Delete(request.Pipeline.Name)
		}); err != nil {
			return errors.Wrapf(err, "collection.Delete")
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

// StartPipeline implements the protobuf pps.StartPipeline RPC
func (a *apiServer) StartPipeline(ctx context.Context, request *pps.StartPipelineRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	if request.Pipeline == nil {
		return nil, errors.New("request.Pipeline cannot be nil")
	}

	// Get request.Pipeline's info
	var pipelineInfo *pps.PipelineInfo
	if err := a.txnEnv.WithReadContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
		var err error
		pipelineInfo, err = a.latestPipelineInfo(txnCtx, request.Pipeline.Name)
		return err
	}); err != nil {
		return nil, err
	}

	// check if the caller is authorized to update this pipeline
	if err := a.authorizePipelineOp(ctx, pipelineOpUpdate, pipelineInfo.Input, pipelineInfo.Pipeline.Name); err != nil {
		return nil, err
	}

	// Remove 'Stopped' from the pipeline spec
	pipelineInfo.Stopped = false
	commit, err := a.writePipelineInfo(ctx, pipelineInfo)
	if err != nil {
		return nil, err
	}
	if a.updatePipelineSpecCommit(ctx, request.Pipeline.Name, commit); err != nil {
		return nil, err
	}

	pachClient := a.env.GetPachClient(ctx)
	// Replace missing branch provenance (removed by StopPipeline)
	provenance := append(branchProvenance(pipelineInfo.Input),
		client.NewBranch(ppsconsts.SpecRepo, pipelineInfo.Pipeline.Name))
	if err := pachClient.CreateBranch(
		request.Pipeline.Name,
		pipelineInfo.OutputBranch,
		pipelineInfo.OutputBranch,
		"",
		provenance,
	); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

// StopPipeline implements the protobuf pps.StopPipeline RPC
func (a *apiServer) StopPipeline(ctx context.Context, request *pps.StopPipelineRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	if request.Pipeline == nil {
		return nil, errors.New("request.Pipeline cannot be nil")
	}

	var pipelineInfo *pps.PipelineInfo
	if err := a.txnEnv.WithReadContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
		var err error
		pipelineInfo, err = a.latestPipelineInfo(txnCtx, request.Pipeline.Name)
		return err
	}); err != nil {
		return nil, err
	}

	// check if the caller is authorized to update this pipeline
	if err := a.authorizePipelineOp(ctx, pipelineOpUpdate, pipelineInfo.Input, pipelineInfo.Pipeline.Name); err != nil {
		return nil, err
	}

	pachClient := a.env.GetPachClient(ctx)
	// Remove branch provenance (pass branch twice so that it continues to point
	// at the same commit, but also pass empty provenance slice)
	if err := pachClient.CreateBranch(
		request.Pipeline.Name,
		pipelineInfo.OutputBranch,
		pipelineInfo.OutputBranch,
		"",
		nil,
	); err != nil {
		return nil, err
	}

	// Update PipelineInfo with new state
	pipelineInfo.Stopped = true
	commit, err := a.writePipelineInfo(ctx, pipelineInfo)
	if err != nil {
		return nil, err
	}
	if a.updatePipelineSpecCommit(ctx, request.Pipeline.Name, commit); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

func (a *apiServer) RunPipeline(ctx context.Context, request *pps.RunPipelineRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	pachClient := a.env.GetPachClient(ctx)
	ctx = ctx // pachClient will propagate auth info
	pfsClient := pachClient.PfsAPIClient
	ppsClient := pachClient.PpsAPIClient

	pipelineInfo, err := a.inspectPipeline(ctx, request.Pipeline.Name)
	if err != nil {
		return nil, err
	}
	// make sure the user isn't trying to run pipeline on an empty branch
	branchInfo, err := pfsClient.InspectBranch(ctx, &pfs.InspectBranchRequest{
		Branch: client.NewBranch(request.Pipeline.Name, pipelineInfo.OutputBranch),
	})
	if err != nil {
		return nil, err
	}
	if branchInfo.Head == nil {
		return nil, errors.Errorf("run pipeline needs a pipeline with existing data to run\nnew commits will trigger the pipeline automatically, so this only needs to be used if you need to run the pipeline on an old version of the data, or to rerun an job")
	}

	key := pfsdb.BranchKey
	branchProvMap := make(map[string]bool)

	// include the branch and its provenance in the branch provenance map
	branchProvMap[key(branchInfo.Branch)] = true
	for _, b := range branchInfo.Provenance {
		branchProvMap[key(b)] = true
	}
	if branchInfo.Head != nil {
		headCommitInfo, err := pfsClient.InspectCommit(ctx, &pfs.InspectCommitRequest{
			Commit: branchInfo.Branch.NewCommit(branchInfo.Head.ID),
		})
		if err != nil {
			return nil, err
		}
		for _, prov := range headCommitInfo.Provenance {
			branchProvMap[key(prov.Commit.Branch)] = true
		}
	}

	// we need to inspect the commit in order to resolve the commit ID, which may have an ancestry tag
	specCommit, err := pfsClient.InspectCommit(ctx, &pfs.InspectCommitRequest{
		Commit: pipelineInfo.SpecCommit,
	})
	if err != nil {
		return nil, err
	}
	provenance := request.Provenance
	provenanceMap := make(map[string]*pfs.CommitProvenance)

	if request.PipelineJobID != "" {
		pipelineJobInfo, err := ppsClient.InspectPipelineJob(ctx, &pps.InspectPipelineJobRequest{
			PipelineJob: client.NewPipelineJob(request.PipelineJobID),
		})
		if err != nil {
			return nil, err
		}
		jobOutputCommit, err := pfsClient.InspectCommit(ctx, &pfs.InspectCommitRequest{
			Commit: pipelineJobInfo.OutputCommit,
		})
		if err != nil {
			return nil, err
		}
		provenance = append(provenance, jobOutputCommit.Provenance...)
	}

	for _, prov := range provenance {
		if prov == nil {
			return nil, errors.Errorf("request should not contain nil provenance")
		}
		// Resolve the provenance commit and populate any missing fields
		provCommitInfo, err := pfsClient.InspectCommit(ctx, &pfs.InspectCommitRequest{
			Commit: prov.Commit,
		})
		if err != nil {
			return nil, err
		}
		prov.Commit = provCommitInfo.Commit

		// ensure the commit provenance is consistent with the branch provenance
		branch := prov.Commit.Branch
		if len(branchProvMap) != 0 {
			if branch.Repo.Name != ppsconsts.SpecRepo && !branchProvMap[key(branch)] {
				return nil, errors.Errorf("the commit provenance contains a branch which the pipeline's branch is not provenant on")
			}
		}
		provenanceMap[key(branch)] = prov
	}

	// fill in the provenance from branches in the provenance that weren't explicitly set in the request
	for _, branchProv := range append(branchInfo.Provenance, branchInfo.Branch) {
		if _, ok := provenanceMap[key(branchProv)]; !ok {
			branchInfo, err := pfsClient.InspectBranch(ctx, &pfs.InspectBranchRequest{
				Branch: branchProv,
			})
			if err != nil {
				return nil, err
			}
			if branchInfo.Head == nil {
				continue
			}
			headCommit, err := pfsClient.InspectCommit(ctx, &pfs.InspectCommitRequest{Commit: branchInfo.Head})
			if err != nil {
				return nil, err
			}
			for _, headProv := range headCommit.Provenance {
				if _, ok := provenanceMap[key(headProv.Commit.Branch)]; !ok {
					provenance = append(provenance, headProv)
				}
			}
		}
	}
	// we need to include the spec commit in the provenance, so that the new job is represented by the correct spec commit
	specProvenance := specCommit.Commit.NewProvenance()
	if _, ok := provenanceMap[key(specProvenance.Commit.Branch)]; !ok {
		provenance = append(provenance, specProvenance)
	}
	if _, err := pachClient.ExecuteInTransaction(func(txnClient *client.APIClient) error {
		newCommit, err := txnClient.PfsAPIClient.StartCommit(txnClient.Ctx(), &pfs.StartCommitRequest{
			Branch:     branchInfo.Branch,
			Provenance: provenance,
		})
		if err != nil {
			return err
		}

		// if stats are enabled, then create a stats commit for the job as well
		if pipelineInfo.EnableStats {
			// it needs to additionally be provenant on the commit we just created
			_, err = txnClient.PfsAPIClient.StartCommit(txnClient.Ctx(), &pfs.StartCommitRequest{
				Branch:     client.NewSystemRepo(request.Pipeline.Name, pfs.MetaRepoType).NewBranch("master"),
				Provenance: append(provenance, newCommit.NewProvenance()),
			})
			if err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return &types.Empty{}, nil
}

func (a *apiServer) RunCron(ctx context.Context, request *pps.RunCronRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	pipelineInfo, err := a.inspectPipeline(ctx, request.Pipeline.Name)
	if err != nil {
		return nil, err
	}

	if pipelineInfo.Input == nil {
		return nil, errors.Errorf("pipeline must have a cron input")
	}

	// find any cron inputs
	var crons []*pps.CronInput
	pps.VisitInput(pipelineInfo.Input, func(in *pps.Input) error {
		if in.Cron != nil {
			crons = append(crons, in.Cron)
		}
		return nil
	})

	if len(crons) < 1 {
		return nil, errors.Errorf("pipeline must have a cron input")
	}

	pachClient := a.env.GetPachClient(ctx)
	txn, err := pachClient.StartTransaction()
	if err != nil {
		return nil, err
	}

	// We need all the DeleteFile and the PutFile requests to happen atomicly
	txnClient := pachClient.WithTransaction(txn)

	// make a tick on each cron input
	for _, cron := range crons {
		// TODO: This isn't transactional, we could support a transactional modify file through the fileset API though.
		if err := txnClient.WithModifyFileClient(client.NewCommit(cron.Repo, "master", ""), func(mf client.ModifyFile) error {
			if cron.Overwrite {
				// get rid of any files, so the new file "overwrites" previous runs
				err = mf.DeleteFile("/")
				if err != nil && !isNotFoundErr(err) && !pfsServer.IsNoHeadErr(err) {
					return errors.Wrapf(err, "delete error")
				}
			}
			// Put in an empty file named by the timestamp
			if err := mf.PutFile(time.Now().Format(time.RFC3339), strings.NewReader("")); err != nil {
				return errors.Wrapf(err, "put error")
			}
			return nil
		}); err != nil {
			return nil, err
		}
	}

	_, err = txnClient.FinishTransaction(txn)
	if err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

// CreateSecret implements the protobuf pps.CreateSecret RPC
func (a *apiServer) CreateSecret(ctx context.Context, request *pps.CreateSecretRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "CreateSecret")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	var s v1.Secret
	if err := json.Unmarshal(request.GetFile(), &s); err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal secret")
	}

	labels := s.GetLabels()
	if labels["suite"] != "" && labels["suite"] != "pachyderm" {
		return nil, errors.Errorf("invalid suite label set on secret: suite=%s", labels["suite"])
	}
	if labels == nil {
		labels = map[string]string{}
	}
	labels["suite"] = "pachyderm"
	labels["secret-source"] = "pachyderm-user"
	s.SetLabels(labels)

	if _, err := a.env.GetKubeClient().CoreV1().Secrets(a.namespace).Create(&s); err != nil {
		return nil, errors.Wrapf(err, "failed to create secret")
	}
	return &types.Empty{}, nil
}

// DeleteSecret implements the protobuf pps.DeleteSecret RPC
func (a *apiServer) DeleteSecret(ctx context.Context, request *pps.DeleteSecretRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "DeleteSecret")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	if err := a.env.GetKubeClient().CoreV1().Secrets(a.namespace).Delete(request.Secret.Name, &metav1.DeleteOptions{}); err != nil {
		return nil, errors.Wrapf(err, "failed to delete secret")
	}
	return &types.Empty{}, nil
}

// InspectSecret implements the protobuf pps.InspectSecret RPC
func (a *apiServer) InspectSecret(ctx context.Context, request *pps.InspectSecretRequest) (response *pps.SecretInfo, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "InspectSecret")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	secret, err := a.env.GetKubeClient().CoreV1().Secrets(a.namespace).Get(request.Secret.Name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get secret")
	}
	creationTimestamp, err := ptypes.TimestampProto(secret.GetCreationTimestamp().Time)
	if err != nil {
		return nil, errors.Errorf("failed to parse creation timestamp")
	}
	return &pps.SecretInfo{
		Secret: &pps.Secret{
			Name: secret.Name,
		},
		Type: string(secret.Type),
		CreationTimestamp: &types.Timestamp{
			Seconds: creationTimestamp.GetSeconds(),
			Nanos:   creationTimestamp.GetNanos(),
		},
	}, nil
}

// ListSecret implements the protobuf pps.ListSecret RPC
func (a *apiServer) ListSecret(ctx context.Context, in *types.Empty) (response *pps.SecretInfos, retErr error) {
	func() { a.Log(nil, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(nil, response, retErr, time.Since(start)) }(time.Now())
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "ListSecret")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	secrets, err := a.env.GetKubeClient().CoreV1().Secrets(a.namespace).List(metav1.ListOptions{
		LabelSelector: "secret-source=pachyderm-user",
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to list secrets")
	}
	secretInfos := []*pps.SecretInfo{}
	for _, s := range secrets.Items {
		creationTimestamp, err := ptypes.TimestampProto(s.GetCreationTimestamp().Time)
		if err != nil {
			return nil, errors.Errorf("failed to parse creation timestamp")
		}
		secretInfos = append(secretInfos, &pps.SecretInfo{
			Secret: &pps.Secret{
				Name: s.Name,
			},
			Type: string(s.Type),
			CreationTimestamp: &types.Timestamp{
				Seconds: creationTimestamp.GetSeconds(),
				Nanos:   creationTimestamp.GetNanos(),
			},
		})
	}

	return &pps.SecretInfos{
		SecretInfo: secretInfos,
	}, nil
}

// DeleteAll implements the protobuf pps.DeleteAll RPC
func (a *apiServer) DeleteAll(ctx context.Context, request *types.Empty) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	pachClient := a.env.GetPachClient(ctx)
	ctx = ctx // pachClient will propagate auth info

	if _, err := a.DeletePipeline(ctx, &pps.DeletePipelineRequest{All: true, Force: true}); err != nil {
		return nil, err
	}

	if err := a.env.GetKubeClient().CoreV1().Secrets(a.namespace).DeleteCollection(&metav1.DeleteOptions{}, metav1.ListOptions{
		LabelSelector: "secret-source=pachyderm-user",
	}); err != nil {
		return nil, err
	}

	// PFS doesn't delete the spec repo, so do it here
	if err := pachClient.DeleteRepo(ppsconsts.SpecRepo, true); err != nil && !isNotFoundErr(err) {
		return nil, err
	}
	if _, err := pachClient.PfsAPIClient.CreateRepo(
		ctx,
		&pfs.CreateRepoRequest{
			Repo:        client.NewRepo(ppsconsts.SpecRepo),
			Update:      true,
			Description: ppsconsts.SpecRepoDesc,
		},
	); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

// ActivateAuth implements the protobuf pps.ActivateAuth RPC
func (a *apiServer) ActivateAuth(ctx context.Context, req *pps.ActivateAuthRequest) (resp *pps.ActivateAuthResponse, retErr error) {
	func() { a.Log(req, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(req, resp, retErr, time.Since(start)) }(time.Now())
	pachClient := a.env.GetPachClient(ctx)

	// Set the permissions on the spec repo so anyone can read it
	if err := pachClient.ModifyRepoRoleBinding(ppsconsts.SpecRepo, auth.AllClusterUsersSubject, []string{auth.RepoReaderRole}); err != nil {
		return nil, errors.Wrapf(grpcutil.ScrubGRPC(err), "cannot configure role binding on spec repo")
	}

	// Set the permissions on the spec repo so the PPS user can write to it
	if err := pachClient.ModifyRepoRoleBinding(ppsconsts.SpecRepo, auth.PpsUser, []string{auth.RepoWriterRole}); err != nil {
		return nil, errors.Wrapf(grpcutil.ScrubGRPC(err), "cannot configure role binding on spec repo")
	}

	// Unauthenticated users can't create new pipelines or repos, and users can't
	// log in while auth is in an intermediate state, so 'pipelines' is exhaustive
	var pipelines []*pps.PipelineInfo
	pipelines, err := pachClient.ListPipeline()
	if err != nil {
		return nil, errors.Wrapf(grpcutil.ScrubGRPC(err), "cannot get list of pipelines to update")
	}

	var eg errgroup.Group
	for _, pipeline := range pipelines {
		pipeline := pipeline
		pipelineName := pipeline.Pipeline.Name
		// 1) Create a new auth token for 'pipeline' and attach it, so that the
		// pipeline can authenticate as itself when it needs to read input data
		eg.Go(func() error {
			return a.txnEnv.WithWriteContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
				token, err := a.env.AuthServer().GetPipelineAuthTokenInTransaction(txnCtx, pipelineName)
				if err != nil {
					return errors.Wrapf(grpcutil.ScrubGRPC(err), "could not generate pipeline auth token")
				}

				var pipelinePtr pps.StoredPipelineInfo
				if err := a.pipelines.ReadWrite(txnCtx.SqlTx).Update(pipelineName, &pipelinePtr, func() error {
					pipelinePtr.AuthToken = token
					return nil
				}); err != nil {
					return errors.Wrapf(err, "could not update \"%s\" with new auth token", pipelineName)
				}
				return nil
			})
		})
		// put 'pipeline' on relevant ACLs
		if err := a.fixPipelineInputRepoACLs(ctx, pipeline, nil); err != nil {
			return nil, err
		}
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	return &pps.ActivateAuthResponse{}, nil
}

func isAlreadyExistsErr(err error) bool {
	return err != nil && strings.Contains(err.Error(), "already exists")
}

func isNotFoundErr(err error) bool {
	return err != nil && strings.Contains(err.Error(), "not found")
}

func (a *apiServer) updatePipelineSpecCommit(ctx context.Context, pipelineName string, commit *pfs.Commit) error {
	err := col.NewSQLTx(ctx, a.env.GetDBClient(), func(sqlTx *sqlx.Tx) error {
		pipelines := a.pipelines.ReadWrite(sqlTx)
		pipelinePtr := &pps.StoredPipelineInfo{}
		if err := pipelines.Get(pipelineName, pipelinePtr); err != nil {
			return err
		}
		pipelinePtr.SpecCommit = commit
		return pipelines.Put(pipelineName, pipelinePtr)
	})
	if isNotFoundErr(err) {
		return newErrPipelineNotFound(pipelineName)
	}
	return err
}

// RepoNameToEnvString is a helper which uppercases a repo name for
// use in environment variable names.
func RepoNameToEnvString(repoName string) string {
	return strings.ToUpper(repoName)
}

func (a *apiServer) rcPods(rcName string) ([]v1.Pod, error) {
	podList, err := a.env.GetKubeClient().CoreV1().Pods(a.namespace).List(metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ListOptions",
			APIVersion: "v1",
		},
		LabelSelector: metav1.FormatLabelSelector(metav1.SetAsLabelSelector(map[string]string{"app": rcName})),
	})
	if err != nil {
		return nil, err
	}
	return podList.Items, nil
}

func (a *apiServer) resolveCommit(ctx context.Context, commit *pfs.Commit) (*pfs.Commit, error) {
	pachClient := a.env.GetPachClient(ctx)
	ci, err := pachClient.PfsAPIClient.InspectCommit(
		pachClient.Ctx(),
		&pfs.InspectCommitRequest{
			Commit:     commit,
			BlockState: pfs.CommitState_STARTED,
		})
	if err != nil {
		return nil, grpcutil.ScrubGRPC(err)
	}
	return ci.Commit, nil
}

func labels(app string) map[string]string {
	return map[string]string{
		"app":       app,
		"suite":     suite,
		"component": "worker",
	}
}
