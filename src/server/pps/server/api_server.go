package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"path"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/itchyny/gojq"
	opentracing "github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	enterpriseclient "github.com/pachyderm/pachyderm/v2/src/enterprise"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	enterpriselimits "github.com/pachyderm/pachyderm/v2/src/server/enterprise/limits"
	enterprisemetrics "github.com/pachyderm/pachyderm/v2/src/server/enterprise/metrics"
	enterprisetext "github.com/pachyderm/pachyderm/v2/src/server/enterprise/text"
	pfsServer "github.com/pachyderm/pachyderm/v2/src/server/pfs"
	ppsServer "github.com/pachyderm/pachyderm/v2/src/server/pps"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/common"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/datum"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/driver"
	workerserver "github.com/pachyderm/pachyderm/v2/src/server/worker/server"
	taskapi "github.com/pachyderm/pachyderm/v2/src/task"

	"github.com/pachyderm/pachyderm/v2/src/internal/ancestry"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/cronutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/lokiutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/metrics"
	"github.com/pachyderm/pachyderm/v2/src/internal/miscutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachtmpl"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsfile"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsload"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/serde"
	"github.com/pachyderm/pachyderm/v2/src/internal/task"
	"github.com/pachyderm/pachyderm/v2/src/internal/tracing"
	"github.com/pachyderm/pachyderm/v2/src/internal/tracing/extended"
	txnenv "github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/internal/watch"
)

const (
	// DefaultUserImage is the image used for jobs when the user does not specify
	// an image.
	DefaultUserImage = "ubuntu:20.04"
	// DefaultDatumTries is the default number of times a datum will be tried
	// before we give up and consider the job failed.
	DefaultDatumTries = 3

	// DefaultLogsFrom is the default duration to return logs from, i.e. by
	// default we return logs from up to 24 hours ago.
	DefaultLogsFrom = time.Hour * 24
	// jobClockSkew is how much earlier than the job start time to look for
	// logs.  It is an attempt to account for clock skew.
	jobClockSkew     = time.Hour
	ppsTaskNamespace = "/pps"

	maxPipelineNameLength = 51

	// dnsLabelLimit is the maximum length of a ReplicationController
	// or Service name.
	dnsLabelLimit = 63
)

var (
	suite = "pachyderm"
)

// apiServer implements the public interface of the Pachyderm Pipeline System,
// including all RPCs defined in the protobuf spec.
type apiServer struct {
	pps.UnimplementedAPIServer

	etcdPrefix            string
	env                   Env
	txnEnv                *txnenv.TransactionEnv
	namespace             string
	workerImage           string
	workerSidecarImage    string
	workerImagePullPolicy string
	storageRoot           string
	storageBackend        string
	storageHostPath       string
	imagePullSecrets      string
	reporter              *metrics.Reporter
	workerUsesRoot        bool
	workerGrpcPort        uint16
	port                  uint16
	peerPort              uint16
	gcPercent             int
	// collections
	pipelines       col.PostgresCollection
	jobs            col.PostgresCollection
	clusterDefaults col.PostgresCollection
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
		if err := validateName(input.Pfs.Name); err != nil {
			return err
		}
		if names[input.Pfs.Name] {
			return errors.Errorf(`name "%s" was used more than once`, input.Pfs.Name)
		}
		names[input.Pfs.Name] = true
	case input.Cron != nil:
		if err := validateName(input.Cron.Name); err != nil {
			return err
		}
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
	}
	return nil
}

func validateName(name string) error {
	if name == "" {
		return errors.Errorf("input must specify a name")
	}
	switch name {
	case common.OutputPrefix, common.EnvFileName:
		return errors.Errorf("input cannot be named %v", name)
	}
	return nil
}

func (a *apiServer) validateInput(pipeline *pps.Pipeline, input *pps.Input) error {
	if err := validateNames(make(map[string]bool), input); err != nil {
		return err
	}
	return pps.VisitInput(input, func(input *pps.Input) error {
		set := false
		if input.Pfs != nil {
			if input.Pfs.Project == "" {
				input.Pfs.Project = pipeline.Project.GetName()
			}
			set = true
			switch {
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
			case input.Pfs.Commit != "":
				return errors.Errorf("input cannot come from a commit; use a branch with head pointing to the commit")
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
			if input.Cron.Project == "" {
				input.Cron.Project = pipeline.Project.GetName()
			}
			if set {
				return errors.Errorf("multiple input types set")
			}
			set = true
			if _, err := cronutil.ParseCronExpression(input.Cron.Spec); err != nil {
				return errors.Wrapf(err, "error parsing cron-spec")
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

func (a *apiServer) validateKube(ctx context.Context) {
	ctx = pctx.Child(ctx, "validateKube")
	errors := false
	kubeClient := a.env.KubeClient

	log.Debug(ctx, "checking if a pod watch can be created")
	_, err := kubeClient.CoreV1().Pods(a.namespace).Watch(ctx, metav1.ListOptions{Watch: true})
	if err != nil {
		errors = true
		log.Error(ctx, "unable to access kubernetes pods, Pachyderm will continue to work but certain pipeline errors will result in pipelines being stuck indefinitely in \"starting\" state.", zap.Error(err))
	} else {
		log.Debug(ctx, "pod watch ok")
	}

	log.Debug(ctx, "checking if pachd pods can be read")
	pods, err := a.pachdPods(ctx)
	if err != nil || len(pods) == 0 {
		errors = true
		log.Error(ctx, "unable to access kubernetes pods, Pachyderm will continue to work but 'pachctl logs' may not work", zap.Error(err))
	} else {
		log.Debug(ctx, "pachd pods found ok", zap.Int("count", len(pods)))
		// No need to check all pods since we're just checking permissions.
		pod := pods[0]
		log.Debug(ctx, "checking if pod logs can be read", zap.String("target", pod.ObjectMeta.Name))
		_, err = kubeClient.CoreV1().Pods(a.namespace).GetLogs(
			pod.ObjectMeta.Name, &v1.PodLogOptions{
				Container: "pachd",
			}).Timeout(10 * time.Second).Do(ctx).Raw()
		if err != nil {
			errors = true
			log.Error(ctx, "unable to access kubernetes logs, Pachyderm will continue to work but 'pachctl logs' may not", zap.Error(err))
		} else {
			log.Debug(ctx, "pod logs read ok")
		}
	}
	name := uuid.NewWithoutDashes()
	labels := map[string]string{appLabel: name}
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
	log.Debug(ctx, "checking if replication controllers can be created", zap.String("rc", rc.ObjectMeta.Name))
	if _, err := kubeClient.CoreV1().ReplicationControllers(a.namespace).Create(ctx, rc, metav1.CreateOptions{}); err != nil {
		errors = true
		log.Error(ctx, "unable to create kubernetes replication controllers, Pachyderm will not function properly until this is fixed", zap.Error(err))
	} else {
		log.Debug(ctx, "rc creation ok")
	}

	log.Debug(ctx, "checking if replication controllers can be deleted", zap.String("rc", rc.ObjectMeta.Name))
	if err := kubeClient.CoreV1().ReplicationControllers(a.namespace).Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
		errors = true
		log.Error(ctx, "unable to delete kubernetes replication controllers, Pachyderm function properly but pipeline cleanup will not work", zap.Error(err))
	} else {
		log.Debug(ctx, "rc deletion ok")
	}
	if !errors {
		log.Info(ctx, "validated kubernetes access ok")
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
	// pipelineOpStartStop is required for StartPipeline and StopPipeline
	pipelineOpStartStop
)

// authorizePipelineOp checks if the user indicated by 'ctx' is authorized
// to perform 'operation' on the pipeline in 'info'
func (a *apiServer) authorizePipelineOp(ctx context.Context, operation pipelineOperation, input *pps.Input, projectName, outputName string) error {
	return a.txnEnv.WithReadContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
		return a.authorizePipelineOpInTransaction(ctx, txnCtx, operation, input, projectName, outputName)
	})
}

// authorizePipelineOpInTransaction is identical to authorizePipelineOp, but runs in the provided transaction
func (a *apiServer) authorizePipelineOpInTransaction(ctx context.Context, txnCtx *txncontext.TransactionContext, operation pipelineOperation, input *pps.Input, projectName, outputName string) error {
	_, err := txnCtx.WhoAmI()
	if auth.IsErrNotActivated(err) {
		return nil // Auth isn't activated, skip authorization completely
	} else if err != nil {
		return err
	}

	if input != nil && operation != pipelineOpDelete && operation != pipelineOpStartStop {
		// Check that the user is authorized to read all input repos, and write to the
		// output repo (which the pipeline needs to be able to do on the user's
		// behalf)
		done := make(map[string]struct{}) // don't double-authorize repos
		if err := pps.VisitInput(input, func(in *pps.Input) error {
			var project, repo string
			if in.Pfs != nil {
				if in.Pfs.Project == "" {
					project = projectName
				} else {
					project = in.Pfs.Project
				}
				repo = in.Pfs.Repo
			} else {
				return nil
			}

			if _, ok := done[repo]; ok {
				return nil
			}
			done[repo] = struct{}{}
			err := a.env.AuthServer.CheckRepoIsAuthorizedInTransaction(txnCtx, &pfs.Repo{Type: pfs.UserRepoType, Project: &pfs.Project{Name: project}, Name: repo}, auth.Permission_REPO_READ)
			return errors.EnsureStack(err)
		}); err != nil {
			return err
		}
	}

	// Check that the user is authorized to write to the output repo
	if outputName != "" {
		var required auth.Permission
		switch operation {
		case pipelineOpCreate:
			// no permissions needed, we will error later if the repo already exists
			return nil
		case pipelineOpListDatum, pipelineOpGetLogs:
			required = auth.Permission_REPO_READ
		case pipelineOpUpdate, pipelineOpStartStop:
			required = auth.Permission_REPO_WRITE
		case pipelineOpDelete:
			if _, err := a.env.PFSServer.InspectRepoInTransaction(ctx, txnCtx, &pfs.InspectRepoRequest{
				Repo: client.NewRepo(projectName, outputName),
			}); errutil.IsNotFoundError(err) {
				// special case: the pipeline output repo has been deleted (so the
				// pipeline is now invalid). It should be possible to delete the pipeline.
				return nil
			}
			required = auth.Permission_REPO_DELETE
		default:
			return errors.Errorf("internal error, unrecognized operation %v", operation)
		}
		if err := a.env.AuthServer.CheckRepoIsAuthorizedInTransaction(txnCtx, &pfs.Repo{Type: pfs.UserRepoType, Project: &pfs.Project{Name: projectName}, Name: outputName}, required); err != nil {
			return errors.EnsureStack(err)
		}
	}
	return nil
}

func (a *apiServer) UpdateJobState(ctx context.Context, request *pps.UpdateJobStateRequest) (response *emptypb.Empty, retErr error) {
	if err := a.txnEnv.WithTransaction(ctx, func(txn txnenv.Transaction) error {
		return errors.EnsureStack(txn.UpdateJobState(request))
	}); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (a *apiServer) UpdateJobStateInTransaction(txnCtx *txncontext.TransactionContext, request *pps.UpdateJobStateRequest) error {
	jobs := a.jobs.ReadWrite(txnCtx.SqlTx)
	jobInfo := &pps.JobInfo{}
	if err := jobs.Get(ppsdb.JobKey(request.Job), jobInfo); err != nil {
		return errors.EnsureStack(err)
	}
	if pps.IsTerminal(jobInfo.State) {
		return ppsServer.ErrJobFinished{Job: jobInfo.Job}
	}

	jobInfo.Restart = request.Restart
	jobInfo.DataProcessed = request.DataProcessed
	jobInfo.DataSkipped = request.DataSkipped
	jobInfo.DataFailed = request.DataFailed
	jobInfo.DataRecovered = request.DataRecovered
	jobInfo.DataTotal = request.DataTotal
	jobInfo.Stats = request.Stats

	return ppsutil.UpdateJobState(a.pipelines.ReadWrite(txnCtx.SqlTx), jobs, jobInfo, request.State, request.Reason)
}

// InspectJob implements the protobuf pps.InspectJob RPC
func (a *apiServer) InspectJob(ctx context.Context, request *pps.InspectJobRequest) (response *pps.JobInfo, retErr error) {
	if request.Job == nil {
		return nil, errors.Errorf("must specify a job")
	}
	ensurePipelineProject(request.Job.Pipeline)
	jobs := a.jobs.ReadOnly(ctx)
	// Make sure the job exists
	// TODO: there's a race condition between this check and the watch below where
	// a deleted job could make this block forever.
	jobInfo := &pps.JobInfo{}
	if err := jobs.Get(ppsdb.JobKey(request.Job), jobInfo); err != nil {
		return nil, errors.EnsureStack(err)
	}
	if request.Wait {
		watcher, err := jobs.WatchOne(ppsdb.JobKey(request.Job))
		if err != nil {
			return nil, errors.EnsureStack(err)
		}
		defer watcher.Close()

	watchLoop:
		for {
			ev, ok := <-watcher.Watch()
			if !ok {
				return nil, errors.Errorf("the stream for job updates closed unexpectedly")
			}
			switch ev.Type {
			case watch.EventError:
				return nil, ev.Err
			case watch.EventDelete:
				return nil, errors.Errorf("job %s was deleted", request.Job.Id)
			case watch.EventPut:
				var jobID string
				if err := ev.Unmarshal(&jobID, jobInfo); err != nil {
					return nil, err
				}
				if pps.IsTerminal(jobInfo.State) {
					break watchLoop
				}
			}
		}
	}

	if request.Details {
		if err := a.getJobDetails(ctx, jobInfo); err != nil {
			return nil, err
		}
	}
	return jobInfo, nil
}

// InspectJobSet implements the protobuf pps.InspectJobSet RPC
func (a *apiServer) InspectJobSet(request *pps.InspectJobSetRequest, server pps.API_InspectJobSetServer) (retErr error) {
	ctx := server.Context()
	pachClient := a.env.GetPachClient(ctx)

	cb := func(projectName, pipelineName string) error {
		jobInfo, err := a.InspectJob(ctx, &pps.InspectJobRequest{
			Job: &pps.Job{
				Pipeline: &pps.Pipeline{
					Project: &pfs.Project{
						Name: projectName,
					},
					Name: pipelineName,
				},
				Id: request.JobSet.Id,
			},
			Details: request.Details,
		})
		if err != nil {
			// Not all commits are guaranteed to have an associated job - skip over it
			if errutil.IsNotFoundError(err) {
				return nil
			}
			return err
		}
		return errors.EnsureStack(server.Send(jobInfo))
	}

	if err := forEachCommitInJob(pachClient, request.JobSet.Id, request.Wait, func(ci *pfs.CommitInfo) error {
		if ci.Commit.Repo.Type != pfs.UserRepoType {
			return nil
		}
		return cb(ci.Commit.Repo.Project.GetName(), ci.Commit.Repo.Name)
	}); err != nil {
		if pfsServer.IsCommitSetNotFoundErr(err) {
			// There are no commits for this ID, but there may still be jobs, query
			// the jobs table directly and don't worry about the topological sort
			// Load all the jobs eagerly to avoid a nested query
			pipelines := []*pps.Pipeline{}
			jobInfo := &pps.JobInfo{}
			if err := a.jobs.ReadOnly(pachClient.Ctx()).GetByIndex(ppsdb.JobsJobSetIndex, request.JobSet.Id, jobInfo, col.DefaultOptions(), func(string) error {
				pipelines = append(pipelines, jobInfo.Job.Pipeline)
				return nil
			}); err != nil {
				return errors.EnsureStack(err)
			}
			for _, pipeline := range pipelines {
				if err := cb(pipeline.Project.GetName(), pipeline.Name); err != nil {
					return err
				}
			}
			return nil
		}

		return err
	}
	return nil
}

func forEachCommitInJob(pachClient *client.APIClient, jobID string, wait bool, cb func(*pfs.CommitInfo) error) error {
	if wait {
		// Note that while this will return jobs in the same topological sort as the
		// commitset, it will block on commits that don't have a job associated with
		// them (aliases and input commits, for example).
		return pachClient.WaitCommitSet(jobID, cb)
	}

	commitInfos, err := pachClient.InspectCommitSet(jobID)
	if err != nil {
		return err
	}

	for _, ci := range commitInfos {
		if err := cb(ci); err != nil {
			return err
		}
	}
	return nil
}

// ListJobSet implements the protobuf pps.ListJobSet RPC
func (a *apiServer) ListJobSet(request *pps.ListJobSetRequest, serv pps.API_ListJobSetServer) (retErr error) {
	pachClient := a.env.GetPachClient(serv.Context())

	// Track the jobsets we've already processed
	seen := map[string]struct{}{}

	filterJob, err := newMessageFilterFunc(request.GetJqFilter(), request.GetProjects())
	if err != nil {
		return errors.Wrap(err, "error creating message filter function")
	}

	number := request.Number
	if number < 0 {
		return errors.Errorf("number must be non-negative")
	}
	if number == 0 {
		number = math.MaxInt64
	}
	paginationMarker := request.PaginationMarker

	// Return jobsets by the newest job in each set (which can be at a different
	// timestamp due to triggers or deferred processing)
	jobInfo := &pps.JobInfo{}
	opts := &col.Options{Target: col.SortByCreateRevision, Order: col.SortDescend}
	if request.Reverse {
		opts.Order = col.SortAscend
	}
	if err := a.jobs.ReadOnly(serv.Context()).List(jobInfo, opts, func(string) error {
		if number == 0 {
			return errutil.ErrBreak
		}
		id := jobInfo.GetJob().GetId()
		if _, ok := seen[id]; ok {
			return nil
		}
		seen[id] = struct{}{}
		if paginationMarker != nil {
			createdAt := time.Unix(int64(jobInfo.Created.GetSeconds()), int64(jobInfo.Created.GetNanos())).UTC()
			fromTime := time.Unix(int64(paginationMarker.GetSeconds()), int64(paginationMarker.GetNanos())).UTC()
			if createdAt.Equal(fromTime) || !request.Reverse && createdAt.After(fromTime) || request.Reverse && createdAt.Before(fromTime) {
				return nil
			}
		}
		jobInfos, err := pachClient.InspectJobSet(id, request.Details)
		if err != nil {
			return err
		}
		// jobInfos can contain jobs that belong in the same project or different projects due to commit sets.
		var jobInfosFiltered []*pps.JobInfo
		for _, ji := range jobInfos {
			if ok, err := filterJob(serv.Context(), ji); err != nil {
				return errors.Wrap(err, "error filtering job")
			} else if !ok {
				continue
			}
			jobInfosFiltered = append(jobInfosFiltered, ji)
		}
		if len(jobInfosFiltered) == 0 {
			return nil
		}
		if err := serv.Send(&pps.JobSetInfo{
			JobSet: client.NewJobSet(id),
			Jobs:   jobInfosFiltered,
		}); err != nil {
			return errors.Wrap(err, "error sending JobSet")
		}
		number--
		return nil
	}); err != nil && err != errutil.ErrBreak {
		return errors.EnsureStack(err)
	}
	return nil
}

// TODO(provenance): rewrite in terms of CommitSubvenance.
// intersectCommitSets finds all commitsets which involve the specified commits
func (a *apiServer) intersectCommitSets(ctx context.Context, commits []*pfs.Commit) (map[string]struct{}, error) {
	intersection := map[string]struct{}{}
	if len(commits) == 0 {
		return intersection, nil
	}
	listClient, err := a.env.GetPachClient(ctx).ListCommitSet(ctx, &pfs.ListCommitSetRequest{})
	if err != nil {
		return nil, err
	}
	if err := grpcutil.ForEach[*pfs.CommitSetInfo](listClient, func(cs *pfs.CommitSetInfo) error {
		provCommits := map[string]struct{}{}
		for _, c := range cs.Commits {
			for _, p := range c.DirectProvenance {
				provCommits[p.String()] = struct{}{}
			}
		}
		allCommits := true
		for _, c := range commits {
			if _, ok := provCommits[c.String()]; !ok {
				allCommits = false
				break
			}
		}
		if allCommits {
			intersection[cs.CommitSet.Id] = struct{}{}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return intersection, nil
}

func (a *apiServer) getJobDetails(ctx context.Context, jobInfo *pps.JobInfo) error {
	projectName := jobInfo.Job.Pipeline.Project.GetName()
	pipelineName := jobInfo.Job.Pipeline.Name

	if err := a.env.AuthServer.CheckRepoIsAuthorized(ctx, &pfs.Repo{Type: pfs.UserRepoType, Project: &pfs.Project{Name: projectName}, Name: pipelineName}, auth.Permission_PIPELINE_LIST_JOB); err != nil && !auth.IsErrNotActivated(err) {
		return errors.EnsureStack(err)
	}

	// Override the SpecCommit for the pipeline to be what it was when this job
	// was created, this prevents races between updating a pipeline and
	// previous jobs running.
	pipelineInfo := &pps.PipelineInfo{}
	if err := a.pipelines.ReadOnly(ctx).GetUniqueByIndex(
		ppsdb.PipelinesVersionIndex,
		ppsdb.VersionKey(jobInfo.Job.Pipeline, jobInfo.PipelineVersion),
		pipelineInfo); err != nil {
		return errors.EnsureStack(err)
	}
	details := &pps.JobInfo_Details{}
	details.Transform = pipelineInfo.Details.Transform
	details.ParallelismSpec = pipelineInfo.Details.ParallelismSpec
	details.Egress = pipelineInfo.Details.Egress
	details.Service = pipelineInfo.Details.Service
	details.Spout = pipelineInfo.Details.Spout
	details.ResourceRequests = pipelineInfo.Details.ResourceRequests
	details.ResourceLimits = pipelineInfo.Details.ResourceLimits
	details.SidecarResourceLimits = pipelineInfo.Details.SidecarResourceLimits
	details.SidecarResourceRequests = pipelineInfo.Details.SidecarResourceRequests
	details.Input = ppsutil.JobInput(pipelineInfo, jobInfo.OutputCommit)
	details.Salt = pipelineInfo.Details.Salt
	details.DatumSetSpec = pipelineInfo.Details.DatumSetSpec
	details.DatumTimeout = pipelineInfo.Details.DatumTimeout
	details.JobTimeout = pipelineInfo.Details.JobTimeout
	details.DatumTries = pipelineInfo.Details.DatumTries
	details.SchedulingSpec = pipelineInfo.Details.SchedulingSpec
	details.PodSpec = pipelineInfo.Details.PodSpec
	details.PodPatch = pipelineInfo.Details.PodPatch

	// If the job is running, we fill in WorkerStatus field, otherwise
	// we just return the jobInfo.
	if jobInfo.State == pps.JobState_JOB_RUNNING {
		var pi = &pps.PipelineInfo{
			Pipeline: &pps.Pipeline{
				Project: &pfs.Project{Name: projectName},
				Name:    pipelineName,
			},
			Version: jobInfo.PipelineVersion,
		}
		workerStatus, err := workerserver.Status(ctx, pi, a.env.EtcdClient, a.etcdPrefix, a.workerGrpcPort)
		if err != nil {
			log.Error(ctx, "failed to get worker status", zap.Error(err))
		} else {
			// It's possible that the workers might be working on datums for other
			// jobs, we omit those since they're not part of the status for this
			// job.
			for _, status := range workerStatus {
				if status.JobId == jobInfo.Job.Id {
					details.WorkerStatus = append(details.WorkerStatus, status)
				}
			}
		}
	}

	jobInfo.Details = details
	return nil
}

// ListJob implements the protobuf pps.ListJob RPC
func (a *apiServer) ListJob(request *pps.ListJobRequest, resp pps.API_ListJobServer) (retErr error) {
	filterJob, err := newMessageFilterFunc(request.GetJqFilter(), request.GetProjects())
	if err != nil {
		return errors.Wrap(err, "error creating message filter function")
	}

	ctx := resp.Context()
	pipeline := request.GetPipeline()
	if pipeline != nil {
		ensurePipelineProject(pipeline)
		// If 'pipeline is set, check that caller has access to the pipeline's
		// output repo; currently, that's all that's required for ListJob.
		//
		// If 'pipeline' isn't set, then we don't return an error (otherwise, a
		// caller without access to a single pipeline's output repo couldn't run
		// `pachctl list job` at all) and instead silently skip jobs where the user
		// doesn't have access to the job's output repo.
		if err := a.env.AuthServer.CheckRepoIsAuthorized(ctx, &pfs.Repo{Type: pfs.UserRepoType, Project: pipeline.Project, Name: pipeline.Name}, auth.Permission_PIPELINE_LIST_JOB); err != nil && !auth.IsErrNotActivated(err) {
			return errors.EnsureStack(err)
		}
	}

	number := request.Number
	// If number is not set, return all jobs that match the query
	if number == 0 {
		number = math.MaxInt64
	}
	// pipelineVersions holds the versions of pipelines that we're interested in
	pipelineVersions := make(map[string]bool)
	if err := ppsutil.ListPipelineInfo(ctx, a.pipelines, pipeline, request.GetHistory(),
		func(ptr *pps.PipelineInfo) error {
			pipelineVersions[ppsdb.VersionKey(ptr.Pipeline, ptr.Version)] = true
			return nil
		}); err != nil {
		return err
	}

	jobs := a.jobs.ReadOnly(ctx)
	jobInfo := &pps.JobInfo{}
	_f := func(string) error {
		if number == 0 {
			return errutil.ErrBreak
		}
		if request.PaginationMarker != nil {
			createdAt := time.Unix(int64(jobInfo.Created.GetSeconds()), int64(jobInfo.Created.GetNanos())).UTC()
			fromTime := time.Unix(int64(request.PaginationMarker.GetSeconds()), int64(request.PaginationMarker.GetNanos())).UTC()
			if createdAt.Equal(fromTime) || !request.Reverse && createdAt.After(fromTime) || request.Reverse && createdAt.Before(fromTime) {
				return nil
			}
		}
		if request.GetDetails() {
			if err := a.getJobDetails(ctx, jobInfo); err != nil {
				if auth.IsErrNotAuthorized(err) {
					return nil // skip job--see note at top of function
				}
				return err
			}
		}
		if len(request.GetInputCommit()) > 0 {
			// Only include the job if it's in the set of intersected commitset IDs
			commitsets, err := a.intersectCommitSets(ctx, request.GetInputCommit())
			if err != nil {
				return err
			}
			if _, ok := commitsets[jobInfo.Job.Id]; !ok {
				return nil
			}
		}
		if !pipelineVersions[ppsdb.VersionKey(jobInfo.Job.Pipeline, jobInfo.PipelineVersion)] {
			return nil
		}

		if ok, err := filterJob(ctx, jobInfo); err != nil {
			return errors.Wrap(err, "error filtering job")
		} else if !ok {
			return nil
		}

		// Erase any AuthToken - this shouldn't be returned to anyone (the workers
		// won't use this function to get their auth token)
		jobInfo.AuthToken = ""

		if err := resp.Send(jobInfo); err != nil {
			return errors.Wrap(err, "error sending job")
		}
		number--
		return nil
	}
	opts := &col.Options{Target: col.SortByCreateRevision, Order: col.SortDescend}
	if request.Reverse {
		opts.Order = col.SortAscend
	}
	if pipeline != nil {
		err = jobs.GetByIndex(ppsdb.JobsPipelineIndex, ppsdb.JobsPipelineKey(pipeline), jobInfo, opts, _f)
	} else {
		err = jobs.List(jobInfo, opts, _f)
	}
	if err != nil && err != errutil.ErrBreak {
		if errors.Is(err, context.DeadlineExceeded) {
			return status.Error(codes.DeadlineExceeded, err.Error())
		}
		return errors.EnsureStack(err)
	}
	return nil
}

// SubscribeJob implements the protobuf pps.SubscribeJob RPC
func (a *apiServer) SubscribeJob(request *pps.SubscribeJobRequest, stream pps.API_SubscribeJobServer) (retErr error) {
	ensurePipelineProject(request.GetPipeline())
	ctx := stream.Context()

	// Validate arguments
	if request.Pipeline == nil || request.Pipeline.Name == "" {
		return errors.New("pipeline must be specified")
	}

	if err := a.env.AuthServer.CheckRepoIsAuthorized(ctx, &pfs.Repo{Type: pfs.UserRepoType, Project: request.Pipeline.Project, Name: request.Pipeline.Name}, auth.Permission_PIPELINE_LIST_JOB); err != nil && !auth.IsErrNotActivated(err) {
		return errors.EnsureStack(err)
	}

	// keep track of the jobs that have been sent
	seen := map[string]struct{}{}

	err := a.jobs.ReadOnly(ctx).WatchByIndexF(ppsdb.JobsTerminalIndex, ppsdb.JobsTerminalKey(request.Pipeline, false), func(ev *watch.Event) error {
		var key string
		jobInfo := &pps.JobInfo{}
		if err := ev.Unmarshal(&key, jobInfo); err != nil {
			return errors.Wrapf(err, "unmarshal")
		}

		if _, ok := seen[key]; ok {
			return nil
		}
		seen[key] = struct{}{}

		if request.Details {
			if err := a.getJobDetails(ctx, jobInfo); err != nil {
				return err
			}
		}
		return errors.EnsureStack(stream.Send(jobInfo))
	}, watch.WithSort(col.SortByCreateRevision, col.SortAscend), watch.IgnoreDelete)
	return errors.EnsureStack(err)
}

// DeleteJob implements the protobuf pps.DeleteJob RPC
func (a *apiServer) DeleteJob(ctx context.Context, request *pps.DeleteJobRequest) (response *emptypb.Empty, retErr error) {
	if request.GetJob() == nil {
		return nil, errors.New("job cannot be nil")
	}
	ensurePipelineProject(request.Job.Pipeline)
	if err := a.txnEnv.WithWriteContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
		return a.deleteJobInTransaction(txnCtx, request)
	}); err != nil {
		return nil, err
	}
	clearJobCache(a.env.GetPachClient(ctx), ppsdb.JobKey(request.Job))
	return &emptypb.Empty{}, nil
}

func (a *apiServer) deleteJobInTransaction(txnCtx *txncontext.TransactionContext, request *pps.DeleteJobRequest) error {
	if err := a.stopJob(txnCtx, request.Job, "job deleted"); err != nil {
		return err
	}
	return errors.EnsureStack(a.jobs.ReadWrite(txnCtx.SqlTx).Delete(ppsdb.JobKey(request.Job)))
}

// StopJob implements the protobuf pps.StopJob RPC
func (a *apiServer) StopJob(ctx context.Context, request *pps.StopJobRequest) (response *emptypb.Empty, retErr error) {
	ensurePipelineProject(request.GetJob().GetPipeline())
	if err := a.txnEnv.WithTransaction(ctx, func(txn txnenv.Transaction) error {
		return errors.EnsureStack(txn.StopJob(request))
	}); err != nil {
		return nil, err
	}
	clearJobCache(a.env.GetPachClient(ctx), ppsdb.JobKey(request.Job))
	return &emptypb.Empty{}, nil
}

// TODO: Remove when job state transition operations are handled by a background process.
func clearJobCache(pachClient *client.APIClient, tagPrefix string) {
	if _, err := pachClient.PfsAPIClient.ClearCache(pachClient.Ctx(), &pfs.ClearCacheRequest{
		TagPrefix: tagPrefix,
	}); err != nil {
		log.Error(pachClient.Ctx(), "errored clearing job cache", zap.Error(err))
	}
}

// StopJobInTransaction is identical to StopJob except that it can run inside an
// existing postgres transaction.  This is not an RPC.
func (a *apiServer) StopJobInTransaction(txnCtx *txncontext.TransactionContext, request *pps.StopJobRequest) error {
	reason := request.Reason
	if reason == "" {
		reason = "job stopped"
	}
	return a.stopJob(txnCtx, request.Job, reason)
}

func (a *apiServer) stopJob(txnCtx *txncontext.TransactionContext, job *pps.Job, reason string) error {
	jobs := a.jobs.ReadWrite(txnCtx.SqlTx)
	if job == nil {
		return errors.New("Job must be specified")
	}

	jobInfo := &pps.JobInfo{}
	if err := jobs.Get(ppsdb.JobKey(job), jobInfo); err != nil {
		return errors.EnsureStack(err)
	}

	commitInfo, err := a.env.PFSServer.InspectCommitInTransaction(txnCtx, &pfs.InspectCommitRequest{
		Commit: jobInfo.OutputCommit,
	})
	if err != nil && !pfsServer.IsCommitNotFoundErr(err) && !pfsServer.IsCommitDeletedErr(err) {
		return errors.EnsureStack(err)
	}

	// TODO: Leaning on the reason rather than state for commit errors seems a bit sketchy, but we don't
	// store commit states.
	if commitInfo != nil {
		if err := a.env.PFSServer.FinishCommitInTransaction(txnCtx, &pfs.FinishCommitRequest{
			Commit: ppsutil.MetaCommit(commitInfo.Commit),
			Error:  reason,
			Force:  true,
		}); err != nil && !pfsServer.IsCommitNotFoundErr(err) && !pfsServer.IsCommitDeletedErr(err) && !pfsServer.IsCommitFinishedErr(err) {
			return errors.EnsureStack(err)
		}
		if err := a.env.PFSServer.FinishCommitInTransaction(txnCtx, &pfs.FinishCommitRequest{
			Commit: commitInfo.Commit,
			Error:  reason,
			Force:  true,
		}); err != nil && !pfsServer.IsCommitNotFoundErr(err) && !pfsServer.IsCommitDeletedErr(err) && !pfsServer.IsCommitFinishedErr(err) {
			return errors.EnsureStack(err)
		}
	}

	// TODO: We can still not update a job's state if we fail here. This is
	// probably fine for now since we are likely to have a more comprehensive
	// solution to this with global ids.
	if err := ppsutil.UpdateJobState(a.pipelines.ReadWrite(txnCtx.SqlTx), jobs, jobInfo, pps.JobState_JOB_KILLED, reason); err != nil && !ppsServer.IsJobFinishedErr(err) {
		return err
	}
	return nil
}

// RestartDatum implements the protobuf pps.RestartDatum RPC
func (a *apiServer) RestartDatum(ctx context.Context, request *pps.RestartDatumRequest) (response *emptypb.Empty, retErr error) {
	ensurePipelineProject(request.GetJob().GetPipeline())
	jobInfo, err := a.InspectJob(ctx, &pps.InspectJobRequest{
		Job: request.Job,
	})
	if err != nil {
		return nil, err
	}
	var pi = &pps.PipelineInfo{
		Pipeline: jobInfo.Job.Pipeline,
		Version:  jobInfo.PipelineVersion,
	}
	if err := workerserver.Cancel(ctx, pi, a.env.EtcdClient, a.etcdPrefix, a.workerGrpcPort, request.Job.Id, request.DataFilters); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (a *apiServer) InspectDatum(ctx context.Context, request *pps.InspectDatumRequest) (response *pps.DatumInfo, retErr error) {
	if request.Datum == nil || request.Datum.Id == "" {
		return nil, errors.New("must specify a datum")
	}
	if request.Datum.Job == nil {
		return nil, errors.New("must specify a job")
	}
	ensurePipelineProject(request.Datum.Job.Pipeline)
	// TODO: Auth?
	if err := a.collectDatums(ctx, request.Datum.Job, func(meta *datum.Meta, pfsState *pfs.File) error {
		if common.DatumID(meta.Inputs) == request.Datum.Id {
			response = convertDatumMetaToInfo(meta, request.Datum.Job)
			response.PfsState = pfsState
		}
		return nil
	}); err != nil {
		return nil, err
	}
	if response == nil {
		return nil, errors.Errorf("datum %s not found in job %s", request.Datum.Id, request.Datum.Job)
	}
	return response, nil
}

func (a *apiServer) ListDatum(request *pps.ListDatumRequest, server pps.API_ListDatumServer) (retErr error) {
	// TODO: Auth?
	ensurePipelineProject(request.GetJob().GetPipeline())
	if request.Job != nil && request.Input != nil {
		return errors.Errorf("cannot specify both job and input")
	}
	number := request.Number
	if number == 0 && request.Reverse {
		return errors.Errorf("number must be > 0 when reverse is set")
	}
	if request.Reverse {
		return a.listDatumReverse(server.Context(), request, server)
	}
	if number == 0 {
		number = math.MaxInt64
	}
	if request.Input != nil {
		if err := a.listDatumInput(server.Context(), request.Input, func(meta *datum.Meta) error {
			if number == 0 {
				return errutil.ErrBreak
			}
			info := convertDatumMetaToInfo(meta, nil)
			info.State = pps.DatumState_UNKNOWN
			if (request.PaginationMarker != "" && info.Datum.Id <= request.PaginationMarker) || !request.Filter.Allow(info) {
				return nil
			}
			number--
			return errors.EnsureStack(server.Send(info))
		}); err != nil && !errors.Is(err, errutil.ErrBreak) {
			return errors.EnsureStack(err)
		}
		return nil
	}
	if err := a.collectDatums(server.Context(), request.Job, func(meta *datum.Meta, _ *pfs.File) error {
		if number == 0 {
			return errutil.ErrBreak
		}
		info := convertDatumMetaToInfo(meta, request.Job)
		if (request.PaginationMarker != "" && info.Datum.Id <= request.PaginationMarker) || !request.Filter.Allow(info) {
			return nil
		}
		number--
		return errors.EnsureStack(server.Send(info))
	}); err != nil && !errors.Is(err, errutil.ErrBreak) {
		return errors.EnsureStack(err)
	}
	return nil
}

func (a *apiServer) listDatumReverse(ctx context.Context, request *pps.ListDatumRequest, server pps.API_ListDatumServer) error {
	dis := make([]*pps.DatumInfo, request.Number)
	index := 0
	if request.Input != nil {
		if err := a.listDatumInput(server.Context(), request.Input, func(meta *datum.Meta) error {
			info := convertDatumMetaToInfo(meta, nil)
			info.State = pps.DatumState_UNKNOWN
			if request.PaginationMarker != "" && info.Datum.Id >= request.PaginationMarker {
				return errutil.ErrBreak
			}
			if !request.Filter.Allow(info) {
				return nil
			}
			// wrap around the buffer
			if index == int(request.Number) {
				index = 0
			}
			dis[index] = info
			index++
			return nil
		}); err != nil && !errors.Is(err, errutil.ErrBreak) {
			return errors.EnsureStack(err)
		}
	} else {
		if err := a.collectDatums(server.Context(), request.Job, func(meta *datum.Meta, _ *pfs.File) error {
			info := convertDatumMetaToInfo(meta, request.Job)
			if request.PaginationMarker != "" && info.Datum.Id >= request.PaginationMarker {
				return errutil.ErrBreak
			}
			if !request.Filter.Allow(info) {
				return nil
			}
			if index == int(request.Number) {
				index = 0
			}
			dis[index] = info
			index++
			return nil
		}); err != nil && !errors.Is(err, errutil.ErrBreak) {
			return errors.EnsureStack(err)
		}
	}
	if index == 0 {
		return nil
	}
	// move the index marker to the last populated datum
	index--
	for i := 0; i < len(dis); i++ {
		if dis[index] == nil {
			break
		}
		if err := server.Send(dis[index]); err != nil {
			return errors.EnsureStack(err)
		}
		index--
		// wrap around to the end of the slice
		if index < 0 {
			index = len(dis) - 1
		}
	}
	return nil
}

func (a *apiServer) listDatumInput(ctx context.Context, input *pps.Input, cb func(*datum.Meta) error) error {
	setInputDefaults("", input)
	if visitErr := pps.VisitInput(input, func(input *pps.Input) error {
		if input.Pfs != nil {
			pachClient := a.env.GetPachClient(ctx)
			ci, err := pachClient.InspectCommit(input.Pfs.Project, input.Pfs.Repo, input.Pfs.Branch, "")
			if err != nil {
				return err
			}
			input.Pfs.Commit = ci.Commit.Id
		}
		if input.Cron != nil {
			return errors.Errorf("can't list datums with a cron input, there will be no datums until the pipeline is created")
		}
		return nil
	}); visitErr != nil {
		return visitErr
	}
	pachClient := a.env.GetPachClient(ctx)
	// TODO: Add cache?
	taskDoer := a.env.TaskService.NewDoer(ppsTaskNamespace, uuid.NewWithoutDashes(), nil)
	di, err := datum.NewIterator(pachClient, taskDoer, input)
	if err != nil {
		return err
	}
	return errors.EnsureStack(di.Iterate(func(meta *datum.Meta) error {
		return cb(meta)
	}))
}

func convertDatumMetaToInfo(meta *datum.Meta, sourceJob *pps.Job) *pps.DatumInfo {
	di := &pps.DatumInfo{
		Datum: &pps.Datum{
			Job: meta.Job,
			Id:  common.DatumID(meta.Inputs),
		},
		State:   convertDatumState(meta.State),
		Stats:   meta.Stats,
		ImageId: meta.ImageId,
	}
	for _, input := range meta.Inputs {
		di.Data = append(di.Data, input.FileInfo)
	}
	if meta.Job != nil && !proto.Equal(meta.Job, sourceJob) {
		di.State = pps.DatumState_SKIPPED
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

func (a *apiServer) collectDatums(ctx context.Context, job *pps.Job, cb func(*datum.Meta, *pfs.File) error) error {
	jobInfo, err := a.InspectJob(ctx, &pps.InspectJobRequest{
		Job: job,
	})
	if err != nil {
		return err
	}
	pachClient := a.env.GetPachClient(ctx)
	metaCommit := ppsutil.MetaCommit(jobInfo.OutputCommit)
	fsi := datum.NewCommitIterator(pachClient, metaCommit, nil)
	err = fsi.Iterate(func(meta *datum.Meta) error {
		// TODO: Potentially refactor into datum package (at least the path).
		pfsState := &pfs.File{
			Commit: metaCommit,
			Path:   "/" + path.Join(common.PFSPrefix, common.DatumID(meta.Inputs)),
		}
		return cb(meta, pfsState)
	})
	return errors.EnsureStack(err)
}

func (a *apiServer) GetKubeEvents(request *pps.LokiRequest, apiGetKubeEventsServer pps.API_GetKubeEventsServer) (retErr error) {
	loki, err := a.env.GetLokiClient()
	if err != nil {
		return errors.EnsureStack(err)
	}
	since := time.Time{}
	if request.Since != nil {
		since = time.Now().Add(-time.Duration(request.Since.Seconds) * time.Second)
	}
	return lokiutil.QueryRange(apiGetKubeEventsServer.Context(), loki, `{app="pachyderm-kube-event-tail"}`, since, time.Time{}, false, func(t time.Time, line string) error {
		return errors.EnsureStack(apiGetKubeEventsServer.Send(&pps.LokiLogMessage{
			Message: strings.TrimSuffix(line, "\n"),
		}))
	})
}

func (a *apiServer) QueryLoki(request *pps.LokiRequest, apiQueryLokiServer pps.API_QueryLokiServer) (retErr error) {
	if err := a.env.AuthServer.CheckClusterIsAuthorized(apiQueryLokiServer.Context(), auth.Permission_CLUSTER_GET_LOKI_LOGS); err != nil {
		return errors.EnsureStack(err)
	}
	loki, err := a.env.GetLokiClient()
	if err != nil {
		return errors.EnsureStack(err)
	}
	since := time.Time{}
	if request.Since != nil {
		since = time.Now().Add(-time.Duration(request.Since.Seconds) * time.Second)
	}
	return lokiutil.QueryRange(apiQueryLokiServer.Context(), loki, request.Query, since, time.Time{}, false, func(t time.Time, line string) error {
		return errors.EnsureStack(apiQueryLokiServer.Send(&pps.LokiLogMessage{
			Message: strings.TrimSuffix(line, "\n"),
		}))
	})
}

func (a *apiServer) GetLogs(request *pps.GetLogsRequest, apiGetLogsServer pps.API_GetLogsServer) (retErr error) {
	ctx := apiGetLogsServer.Context()

	ensurePipelineProject(request.GetPipeline())
	ensurePipelineProject(request.GetJob().GetPipeline())
	ensurePipelineProject(request.GetDatum().GetJob().GetPipeline())

	if a.env.Config.LokiLogging || request.UseLokiBackend {
		return a.getLogsLoki(ctx, request, apiGetLogsServer)
	}

	// Authorize request and get list of pods containing logs we're interested in
	// (based on pipeline and job filters)
	var (
		rcName, containerName string
		pods                  []v1.Pod
		err                   error
		since                 *time.Duration
	)
	// Convert request.Since to a usable timestamp.
	if request.Since != nil {
		since = new(time.Duration)
		*since = request.Since.AsDuration()
		if *since == 0 {
			since = nil
		}
	}
	if request.Pipeline == nil && request.Job == nil {
		if len(request.DataFilters) > 0 || request.Datum != nil {
			return errors.Errorf("must specify the Job or Pipeline that the datum is from to get logs for it")
		}
		if err := a.env.AuthServer.CheckClusterIsAuthorized(apiGetLogsServer.Context(), auth.Permission_CLUSTER_GET_PACHD_LOGS); err != nil {
			return errors.EnsureStack(err)
		}
		containerName, rcName = "pachd", "pachd"
		pods, err = a.pachdPods(apiGetLogsServer.Context())
	} else if request.Job != nil && request.Job.GetPipeline().GetName() == "" {
		return errors.Errorf("pipeline must be specified for the given job")
	} else if request.Job != nil && request.Pipeline != nil && !proto.Equal(request.Job.Pipeline, request.Pipeline) {
		return errors.Errorf("job is from the wrong pipeline")
	} else {
		containerName = client.PPSWorkerUserContainerName

		// 1) Lookup the PipelineInfo for this pipeline/job, for auth and to get the
		// RC name
		var pipeline *pps.Pipeline
		if request.Pipeline != nil && request.Job == nil {
			pipeline = request.Pipeline
		} else if request.Job != nil {
			// If user provides a job, lookup the pipeline from the JobInfo, and then
			// get the pipeline RC
			jobInfo := &pps.JobInfo{}
			err = a.jobs.ReadOnly(apiGetLogsServer.Context()).Get(ppsdb.JobKey(request.Job), jobInfo)
			if err != nil {
				return errors.Wrapf(err, "could not get job information for \"%s\"", request.Job.Id)
			}
			if created := jobInfo.GetCreated(); since == nil && created != nil {
				t := created.AsTime()
				since = new(time.Duration)
				*since = time.Since(t) + jobClockSkew
			}
			pipeline = jobInfo.Job.Pipeline
		} // not possible for request.{Pipeline,Job} to both be nil due to previous check above
		pipelineInfo, err := a.inspectPipeline(apiGetLogsServer.Context(), pipeline, true)
		if err != nil {
			return errors.Wrapf(err, "could not get pipeline information for %s", pipeline)
		}

		// 2) Check whether the caller is authorized to get logs from this pipeline/job
		if err := a.authorizePipelineOp(apiGetLogsServer.Context(), pipelineOpGetLogs, pipelineInfo.Details.Input, pipelineInfo.Pipeline.Project.GetName(), pipelineInfo.Pipeline.Name); err != nil {
			return err
		}

		// 3) Get pods for this pipeline
		rcName = ppsutil.PipelineRcName(pipelineInfo)
		pods, err = a.rcPods(apiGetLogsServer.Context(), pipelineInfo)
		if err != nil {
			return err
		}
	}

	if err != nil {
		return errors.Wrapf(err, "could not get pods in rc \"%s\" containing logs", rcName)
	}
	if len(pods) == 0 {
		return errors.Errorf("no pods belonging to the rc \"%s\" were found", rcName)
	}

	// Spawn one goroutine per pod. Each goro writes its pod's logs to a channel
	// and channels are read into the output server in a stable order.
	// (sort the pods to make sure that the order of log lines is stable)
	sort.Sort(podSlice(pods))
	logCh := make(chan *pps.LogMessage)
	var eg errgroup.Group
	var mu sync.Mutex
	if since == nil {
		since = new(time.Duration)
		*since = DefaultLogsFrom
	}
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
				sinceSeconds := new(int64)
				*sinceSeconds = int64(*since / time.Second)
				stream, err := a.env.KubeClient.CoreV1().Pods(a.namespace).GetLogs(
					pod.ObjectMeta.Name, &v1.PodLogOptions{
						Container:    containerName,
						Follow:       request.Follow,
						TailLines:    tailLines,
						SinceSeconds: sinceSeconds,
					}).Timeout(10 * time.Second).Stream(apiGetLogsServer.Context())
				if err != nil {
					return errors.EnsureStack(err)
				}
				defer func() {
					if err := stream.Close(); err != nil && retErr == nil {
						retErr = err
					}
				}()

				// Parse pods' log lines, and filter out irrelevant ones
				return ppsutil.FilterLogLines(request, stream, containerName == "pachd", func(msg *pps.LogMessage) error {
					select {
					case logCh <- msg:
						return nil
					case <-apiGetLogsServer.Context().Done():
						return errutil.ErrBreak
					}
				})
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
			return errors.EnsureStack(err)
		}
	}
	return errors.EnsureStack(egErr)
}

func (a *apiServer) getLogsLoki(ctx context.Context, request *pps.GetLogsRequest, apiGetLogsServer pps.API_GetLogsServer) (retErr error) {
	// Authorize request and get list of pods containing logs we're interested in
	// (based on pipeline and job filters)
	loki, err := a.env.GetLokiClient()
	if err != nil {
		return err
	}
	var from time.Time
	if request.Since != nil {
		since := request.Since.AsDuration()
		if since != 0 {
			from = time.Now().Add(-since)
		}
	}
	if request.Pipeline == nil && request.Job == nil {
		if len(request.DataFilters) > 0 || request.Datum != nil {
			return errors.Errorf("must specify the Job or Pipeline that the datum is from to get logs for it")
		}
		if err := a.env.AuthServer.CheckClusterIsAuthorized(apiGetLogsServer.Context(), auth.Permission_CLUSTER_GET_PACHD_LOGS); err != nil {
			return errors.EnsureStack(err)
		}
		return lokiutil.QueryRange(apiGetLogsServer.Context(), loki, `{app="pachd"}`, from, time.Time{}, request.Follow, func(t time.Time, line string) error {
			return errors.EnsureStack(apiGetLogsServer.Send(&pps.LogMessage{
				Message: strings.TrimSuffix(line, "\n"),
			}))
		})
	} else if request.Job != nil && request.Pipeline != nil && !proto.Equal(request.Job.Pipeline, request.Pipeline) {
		return errors.Errorf("job is from the wrong pipeline")
	}

	// 1) Lookup the PipelineInfo for this pipeline/job, for auth and to get the
	// RC name
	var pipelineInfo *pps.PipelineInfo

	if request.Pipeline != nil && request.Job == nil {
		pipelineInfo, err = a.inspectPipeline(apiGetLogsServer.Context(), request.Pipeline, true)
		if err != nil {
			return errors.Wrapf(err, "could not get pipeline information for %s", request.Pipeline)
		}
	} else if request.Job != nil {
		// If user provides a job, lookup the pipeline from the JobInfo, and then
		// get the pipeline RC
		jobInfo := &pps.JobInfo{}
		err = a.jobs.ReadOnly(apiGetLogsServer.Context()).Get(ppsdb.JobKey(request.Job), jobInfo)
		if err != nil {
			return errors.Wrapf(err, "could not get job information for \"%s\"", request.Job.Id)
		}
		if created := jobInfo.GetCreated(); from.IsZero() && created != nil {
			from = created.AsTime()
			from = from.Add(-jobClockSkew)
		}
		pipelineInfo, err = a.inspectPipeline(apiGetLogsServer.Context(), jobInfo.Job.Pipeline, true)
		if err != nil {
			return errors.Wrapf(err, "could not get pipeline information for %s", jobInfo.Job.Pipeline)
		}
	}

	// 2) Check whether the caller is authorized to get logs from this pipeline/job
	if err := a.authorizePipelineOp(apiGetLogsServer.Context(), pipelineOpGetLogs, pipelineInfo.Details.Input, pipelineInfo.Pipeline.Project.GetName(), pipelineInfo.Pipeline.Name); err != nil {
		return err
	}
	query := fmt.Sprintf(`{pipelineProject=%q, pipelineName=%q, container="user"}`, pipelineInfo.Pipeline.Project.Name, pipelineInfo.Pipeline.Name)
	if request.Master {
		query += contains("master")
	}
	if request.Job != nil {
		query += contains(request.Job.Id)
	}
	if request.Datum != nil {
		query += contains(request.Datum.Id)
	}
	for _, filter := range request.DataFilters {
		query += contains(filter)
	}
	if from.IsZero() {
		from = time.Now().Add(-DefaultLogsFrom)
	}
	return lokiutil.QueryRange(ctx, loki, query, from, time.Time{}, request.Follow, func(t time.Time, line string) error {
		msg := new(pps.LogMessage)
		if err := ParseLokiLine(line, msg); err != nil {
			log.Debug(ctx, "get logs (loki): unparseable log line", zap.String("line", line), zap.Error(err))
			return nil
		}

		// These filters are almost always unnecessary because we apply
		// them in the Loki request, but many of them are just done with
		// string matching so there technically could be some false
		// positive matches (although it's pretty unlikely), checking here
		// just makes sure we don't accidentally intersperse unrelated log
		// messages.
		if request.Pipeline != nil && request.Pipeline.Name != msg.PipelineName {
			return nil
		}
		if request.Job != nil && (request.Job.Id != msg.JobId || request.Job.Pipeline.Name != msg.PipelineName) {
			return nil
		}
		if request.Datum != nil && request.Datum.Id != msg.DatumId {
			return nil
		}
		if request.Master != msg.Master {
			return nil
		}
		if !common.MatchDatum(request.DataFilters, msg.Data) {
			return nil
		}
		msg.Message = strings.TrimSuffix(msg.Message, "\n")
		return errors.EnsureStack(apiGetLogsServer.Send(msg))
	})
}

// parseNativeLineWithDuplicate key is like parseNativeLine, but tries to convert to a
// map[string]any and back to JSON if unmarshaling fails because of a duplicate key.
func parseNativeLineWithDuplicateKey(s string, msg *pps.LogMessage) error {
	if err := parseNativeLine(s, msg); err != nil {
		if !strings.Contains(err.Error(), "duplicate field") {
			return errors.Wrap(err, "parseNativeLine")
		}
		var result map[string]any
		if err := json.Unmarshal([]byte(s), &result); err != nil {
			return errors.Wrap(err, "unmarshal into map[string]any")
		}
		bs, err := json.Marshal(result)
		if err != nil {
			return errors.Wrap(err, "marshal map[string]any back to JSON")
		}
		if err := parseNativeLine(string(bs), msg); err != nil {
			return errors.Wrap(err, "parseNativeLine(cleaned)")
		}
		return nil
	}
	return nil
}

// parseNativeLine receives a raw chunk of JSON from Loki, which is our logged
// pps.LogMessage object.
func parseNativeLine(s string, msg *pps.LogMessage) error {
	m := protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: true,
	}
	if err := m.Unmarshal([]byte(s), msg); err != nil {
		return errors.EnsureStack(err)
	}
	if proto.Equal(msg, &pps.LogMessage{}) {
		// NOTE(jonathan): If the parsed proto is equal to an empty proto, we have to assume
		// that the JSON is not native format.  A usable Docker-format message looks empty
		// and valid to the Unmarshaler in this mode.  This is not strictly correct, as
		// `{"master":false}` is a valid pps.LogMessage proto, but appears empty here, so we
		// incorrectly reject it.  This is the price we pay for DisardUnknown, of which zap
		// logs lots of.  (If we didn't allow unknown fields then we couldn't ever log
		// fields in the worker, which is too unfortunate to consider.)
		//
		// An alternative considered was to make zap put all of its fields in a
		// zap.Namespace, and then put that key into the proto so it would be recognized
		// here.  That would work, but it's difficult to then add optional fields to the zap
		// logger, like workerId, jobId, etc.; we'd have to know the values in advance of
		// creating the namespace.  The worker kind of incrementally learns about these
		// things and adds them to log messages when it knows about them, not at the start
		// of logging, so that doesn't work either.
		//
		// As an example:
		// l = l.With(zap.String("known_field", ""), zap.Namespace("extraFields"))
		// l.Info("1")
		// l.With(zap.String("known_field", "the_value")).Info("2")
		//
		// logs:
		// 1 {"known_field":""}
		// 2 {"known_field":"", "extraFields":{"known_field":"the_value"}}
		//
		// Unfortunate.  If it knew how to "promote" known fields, we wouldn't have this
		// problem.
		return errors.New("parsed message is completely empty; assume invalid")
	}
	return nil
}

// parseDockerLine receives the raw Docker log line, which is a JSON object
// containing a time, stream, and log field.  The log contains what our program
// actually logged, so we parse that with parseNative after extracting it.
func parseDockerLine(s string, msg *pps.LogMessage) error {
	var result struct{ Log string }
	if err := json.Unmarshal([]byte(s), &result); err != nil {
		return errors.Errorf("outer json: %v", err)
	}
	if result.Log == "" {
		return errors.New("log field is empty")
	}
	if err := parseNativeLineWithDuplicateKey(result.Log, msg); err != nil {
		return errors.Errorf("native json (%q): %v", result.Log, err)
	}
	return nil
}

// parseCRILine receives the raw CRI-format (used by containerd and cri-o) log
// line, which is a line of text formatted like: <RFC3339 time> <stream name>
// <maybe flags> <log message>.  Because this format is not actually parseable
// (the log message could start with a flag, but there are no flags), we just
// seek to the first { and feed that to the native parser.
func parseCRILine(s string, msg *pps.LogMessage) error {
	b := []byte(s)
	i := bytes.IndexByte(b, '{')
	if i < 0 {
		return errors.New("line does not contain {")
	}
	l := string(b[i:])
	if err := parseNativeLineWithDuplicateKey(l, msg); err != nil {
		return errors.Errorf("native json (%q): %v", l, err)
	}
	return nil
}

func ParseLokiLine(inputLine string, msg *pps.LogMessage) error {
	// There are three possible log formats in Loki, depending on the underlying formatting done
	// by the container runtime.  Under ideal circumstances, the user of Loki has configured
	// promtail to remove the runtime-specific logs, but it's quite easy to miss this
	// configuration aspect, so we correct for it here.  (See:
	// https://grafana.com/docs/loki/latest/clients/promtail/stages/cri/ and
	// https://grafana.com/docs/loki/latest/clients/promtail/stages/docker/)

	// Try each driver; if one results in a valid message, then we're done.
	errs := make(map[string]error)
	parsers := []struct {
		driver string
		parser func(string, *pps.LogMessage) error
	}{
		// The order here is important, because native and docker messages are ambiguous.
		// It is unfortunate that we can't happy-path native.
		{"cri", parseCRILine},
		{"docker", parseDockerLine},
		{"native", parseNativeLineWithDuplicateKey},
	}
	for _, item := range parsers {
		if err := item.parser(inputLine, msg); err != nil {
			errs[item.driver] = err
			continue
		}
		// This driver worked, so give up!
		return nil
	}

	if len(errs) == 0 {
		// This can't happen, because there are 3 iterations of the loop and each iteration
		// either returns or populates errs.
		return errors.EnsureStack(errors.New("impossible: did not succeed, but also did not error"))
	}

	// None of the drivers worked; explain why each failed.
	errMsg := new(strings.Builder)
	errMsg.WriteString("interpret loki line as json:")
	for parser, err := range errs {
		errMsg.WriteString("\n\t")
		errMsg.WriteString(parser)
		errMsg.WriteString(": ")
		errMsg.WriteString(err.Error())
	}
	return errors.EnsureStack(errors.New(errMsg.String()))
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

func (a *apiServer) validatePipelineRequest(request *pps.CreatePipelineRequest) error {
	if request.Pipeline == nil {
		return errors.New("invalid pipeline spec: request.Pipeline cannot be nil")
	}
	if request.Pipeline.Name == "" {
		return errors.New("invalid pipeline spec: request.Pipeline.Name cannot be empty")
	}
	if err := ancestry.ValidateName(request.Pipeline.Name); err != nil {
		return errors.Wrapf(err, "invalid pipeline name")
	}
	if len(request.Pipeline.Name) > maxPipelineNameLength {
		return errors.Errorf("pipeline name %q is %d characters longer than the %d max",
			request.Pipeline.Name, len(request.Pipeline.Name)-maxPipelineNameLength, maxPipelineNameLength)
	}
	// TODO(CORE-1489): Remove dependency of name length on Kubernetes
	// resource naming convention.
	if k8sName := ppsutil.PipelineRcName(&pps.PipelineInfo{Pipeline: &pps.Pipeline{Project: &pfs.Project{Name: request.Pipeline.GetProject().GetName()}, Name: request.Pipeline.Name}, Version: 99}); len(k8sName) > dnsLabelLimit {
		return errors.Errorf("Kubernetes name %q is %d characters longer than the %d max", k8sName, len(k8sName)-dnsLabelLimit, dnsLabelLimit)
	}
	// TODO(msteffen) eventually TFJob and Transform will be alternatives, but
	// currently TFJob isn't supported
	if request.TfJob != nil {
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
	if request.Spout != nil && request.Autoscaling {
		return errors.Errorf("autoscaling can't be used with spouts (spouts aren't triggered externally)")
	}
	var tolErrs error
	for i, t := range request.GetTolerations() {
		if _, err := transformToleration(t); err != nil {
			errors.JoinInto(&tolErrs, errors.Errorf("toleration %d/%d: %v", i+1, len(request.GetTolerations()), err))
		}
	}
	if tolErrs != nil {
		return tolErrs
	}
	return nil
}

func (a *apiServer) validateEnterpriseChecks(ctx context.Context, req *pps.CreatePipelineRequest) error {
	if _, err := a.inspectPipeline(ctx, req.Pipeline, false); err == nil {
		// Pipeline already exists so we allow people to update it even if
		// they're over the limits.
		return nil
	} else if !errutil.IsNotFoundError(err) {
		return err
	}
	pachClient := a.env.GetPachClient(ctx)
	resp, err := pachClient.Enterprise.GetState(pachClient.Ctx(),
		&enterpriseclient.GetStateRequest{})
	if err != nil {
		return errors.Wrapf(grpcutil.ScrubGRPC(err), "could not get enterprise status")
	}
	if resp.State == enterpriseclient.State_ACTIVE {
		// Enterprise is enabled so anything goes.
		return nil
	}
	var info pps.PipelineInfo
	seen := make(map[string]struct{})
	if err := a.pipelines.ReadOnly(ctx).List(&info, col.DefaultOptions(), func(_ string) error {
		seen[info.Pipeline.Name] = struct{}{}
		return nil
	}); err != nil {
		return errors.EnsureStack(err)
	}
	if len(seen) >= enterpriselimits.Pipelines {
		enterprisemetrics.IncEnterpriseFailures()
		return errors.Errorf("%s requires an activation key to create more than %d total pipelines (you have %d). %s\n\n%s",
			enterprisetext.OpenSourceProduct, enterpriselimits.Pipelines, len(seen), enterprisetext.ActivateCTA, enterprisetext.RegisterCTA)
	}
	if req.ParallelismSpec != nil && req.ParallelismSpec.Constant > enterpriselimits.Parallelism {
		enterprisemetrics.IncEnterpriseFailures()
		return errors.Errorf("%s requires an activation key to create pipelines with parallelism more than %d. %s\n\n%s",
			enterprisetext.OpenSourceProduct, enterpriselimits.Parallelism, enterprisetext.ActivateCTA, enterprisetext.RegisterCTA)
	}
	return nil
}

func (a *apiServer) validateSecret(ctx context.Context, req *pps.CreatePipelineRequest) error {
	for _, s := range req.GetTransform().GetSecrets() {
		if s.EnvVar != "" && s.Key == "" {
			return errors.Errorf("secret %s has env_var set but is missing key", s.Name)
		}
		ss, err := a.env.KubeClient.CoreV1().Secrets(a.namespace).Get(ctx, s.Name, metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				return errors.Errorf("missing Kubernetes secret %s", s.Name)
			}
			return errors.Wrapf(err, "could not get Kubernetes secret %s", s.Name)
		}
		if s.EnvVar == "" {
			continue
		}
		if _, ok := ss.Data[s.Key]; !ok {
			return errors.Errorf("Kubernetes secret %s missing key %s", s.Name, s.Key)
		}
	}
	return nil
}

// validateEgress validates the egress field.
func (a *apiServer) validateEgress(pipelineName string, egress *pps.Egress) error {
	if egress == nil {
		return nil
	}
	return pfsServer.ValidateSQLDatabaseEgress(egress.GetSqlDatabase())
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
	if err := validateTransform(pipelineInfo.Details.Transform); err != nil {
		return errors.Wrapf(err, "invalid transform")
	}
	if err := a.validateInput(pipelineInfo.Pipeline, pipelineInfo.Details.Input); err != nil {
		return err
	}
	if err := a.validateEgress(pipelineInfo.Pipeline.Name, pipelineInfo.Details.Egress); err != nil {
		return err
	}
	if pipelineInfo.Details.ParallelismSpec != nil {
		if pipelineInfo.Details.Service != nil && pipelineInfo.Details.ParallelismSpec.Constant != 1 {
			return errors.New("services can only be run with a constant parallelism of 1")
		}
	}
	if pipelineInfo.Details.OutputBranch == "" {
		return errors.New("pipeline needs to specify an output branch")
	}
	if pipelineInfo.Details.PodSpec != "" && !json.Valid([]byte(pipelineInfo.Details.PodSpec)) {
		return errors.Errorf("malformed PodSpec")
	}
	if pipelineInfo.Details.PodPatch != "" && !json.Valid([]byte(pipelineInfo.Details.PodPatch)) {
		return errors.Errorf("malformed PodPatch")
	}
	if pipelineInfo.Details.Service != nil {
		validServiceTypes := map[v1.ServiceType]bool{
			v1.ServiceTypeClusterIP:    true,
			v1.ServiceTypeLoadBalancer: true,
			v1.ServiceTypeNodePort:     true,
		}

		if !validServiceTypes[v1.ServiceType(pipelineInfo.Details.Service.Type)] {
			return errors.Errorf("the following service type %s is not allowed", pipelineInfo.Details.Service.Type)
		}
	}
	if pipelineInfo.Details.Spout != nil {
		if pipelineInfo.Details.Spout.Service == nil && pipelineInfo.Details.Input != nil {
			return errors.Errorf("spout pipelines (without a service) must not have an input")
		}
	}
	return nil
}

func branchProvenance(project *pfs.Project, input *pps.Input) []*pfs.Branch {
	var result []*pfs.Branch
	pps.VisitInput(input, func(input *pps.Input) error { //nolint:errcheck
		if input.Pfs != nil {
			var projectName = input.Pfs.Project
			if projectName == "" {
				projectName = project.GetName()
			}
			result = append(result, client.NewBranch(projectName, input.Pfs.Repo, input.Pfs.Branch))
		}
		if input.Cron != nil {
			var projectName = input.Cron.Project
			if projectName == "" {
				projectName = project.GetName()
			}
			result = append(result, client.NewBranch(projectName, input.Cron.Repo, "master"))
		}
		return nil
	})
	return result
}

func (a *apiServer) fixPipelineInputRepoACLsInTransaction(txnCtx *txncontext.TransactionContext, pipelineInfo *pps.PipelineInfo, prevPipelineInfo *pps.PipelineInfo) (retErr error) {
	addRead := make(map[string]*pfs.Repo)
	addWrite := make(map[string]*pfs.Repo)
	remove := make(map[string]*pfs.Repo)
	var pipeline *pps.Pipeline
	// Figure out which repos 'pipeline' might no longer be using
	if prevPipelineInfo != nil {
		pipeline = prevPipelineInfo.Pipeline
		if err := pps.VisitInput(prevPipelineInfo.Details.Input, func(input *pps.Input) error {
			var repo *pfs.Repo
			switch {
			case input.Pfs != nil:
				repo = new(pfs.Repo)
				if input.Pfs.Project == "" {
					repo.Project = prevPipelineInfo.Pipeline.Project
				} else {
					repo.Project = &pfs.Project{Name: input.Pfs.Project}
				}
				repo.Name = input.Pfs.Repo
			case input.Cron != nil:
				repo = new(pfs.Repo)
				repo.Name = input.Cron.Repo
				if input.Cron.Project == "" {
					repo.Project = prevPipelineInfo.Pipeline.Project
				} else {
					repo.Project = &pfs.Project{Name: input.Cron.Project}
				}
			default:
				return nil // no scope to set: input is not a repo
			}
			remove[repo.String()] = repo
			return nil
		}); err != nil {
			return errors.EnsureStack(err)
		}
	}

	// Figure out which repos 'pipeline' is using
	if pipelineInfo != nil {
		// also check that pipeline name is consistent
		if pipeline == nil {
			pipeline = pipelineInfo.Pipeline
		} else if !(pipelineInfo.Pipeline.Project.GetName() == pipeline.Project.GetName() && pipelineInfo.Pipeline.Name == pipeline.Name) {
			return errors.Errorf("pipelineInfo (%s) and prevPipelineInfo (%s) do not "+
				"belong to matching pipelines; this is a bug",
				pipelineInfo.Pipeline.Name, prevPipelineInfo.Pipeline.Name)
		}

		// collect inputs (remove redundant inputs from 'remove', but don't
		// bother authorizing 'pipeline' twice)
		if err := pps.VisitInput(pipelineInfo.Details.Input, func(input *pps.Input) error {
			var repo *pfs.Repo
			switch {
			case input.Pfs != nil:
				repo = new(pfs.Repo)
				if input.Pfs.Project == "" {
					repo.Project = pipelineInfo.Pipeline.Project
				} else {
					repo.Project = &pfs.Project{Name: input.Pfs.Project}
				}
				repo.Name = input.Pfs.Repo
			case input.Cron != nil:
				repo = new(pfs.Repo)
				if input.Cron.Project == "" {
					repo.Project = pipelineInfo.Pipeline.Project
				} else {
					repo.Project = &pfs.Project{Name: input.Cron.Project}
				}
				repo.Name = input.Cron.Repo
			default:
				return nil // no scope to set: input is not a repo
			}
			if _, ok := remove[repo.String()]; ok {
				delete(remove, repo.String())
			} else {
				addRead[repo.String()] = repo
				if input.Cron != nil {
					addWrite[repo.String()] = repo
				}
			}
			return nil
		}); err != nil {
			return errors.EnsureStack(err)
		}
	}
	if pipeline == nil {
		return errors.Errorf("fixPipelineInputRepoACLs called with both current and " +
			"previous pipelineInfos == to nil; this is a bug")
	}

	// make sure we don't touch the pipeline's permissions on its output repo
	delete(remove, pipeline.String())
	delete(addRead, pipeline.String())
	delete(addWrite, pipeline.String())

	defer func() {
		retErr = errors.Wrapf(retErr, "error fixing ACLs on \"%s\"'s input repos", pipeline)
	}()

	// Remove pipeline from old, unused inputs
	for _, repo := range remove {
		// If we get an `ErrNoRoleBinding` that means the input repo no longer exists - we're removing it anyways, so we don't care.
		if err := a.env.AuthServer.RemovePipelineReaderFromRepoInTransaction(txnCtx, repo, pipeline); err != nil && !auth.IsErrNoRoleBinding(err) {
			return errors.EnsureStack(err)
		}
	}
	// Add pipeline to every new input's ACL as a READER
	for _, repo := range addRead {
		// This raises an error if the input repo doesn't exist, or if the user doesn't have permissions to add a pipeline as a reader on the input repo
		if err := a.env.AuthServer.AddPipelineReaderToRepoInTransaction(txnCtx, repo, pipeline); err != nil {
			return errors.EnsureStack(err)
		}
	}

	for _, repo := range addWrite {
		// This raises an error if the input repo doesn't exist, or if the user doesn't have permissions to add a pipeline as a writer on the input repo
		if err := a.env.AuthServer.AddPipelineWriterToSourceRepoInTransaction(txnCtx, repo, pipeline); err != nil {
			return errors.EnsureStack(err)
		}
	}

	// Add pipeline to its output repo's ACL as a WRITER if it's new
	if prevPipelineInfo == nil {
		if err := a.env.AuthServer.AddPipelineWriterToRepoInTransaction(txnCtx, pipeline); err != nil {
			return errors.EnsureStack(err)
		}
	}
	return nil
}

// getExpectedNumWorkers is a helper function for CreatePipeline that transforms
// the parallelism spec in CreatePipelineRequest.Parallelism into a constant
// that can be stored in PipelineInfo.Parallelism
func getExpectedNumWorkers(pipelineInfo *pps.PipelineInfo) (int, error) {
	switch pspec := pipelineInfo.Details.ParallelismSpec; {
	case pspec == nil, pspec.Constant == 0:
		return 1, nil
	case pspec.Constant > 0:
		return int(pspec.Constant), nil
	default:
		return 0, errors.Errorf("unable to interpret ParallelismSpec %+v", pspec)
	}
}

// CreatePipeline implements the protobuf pps.CreatePipeline RPC
//
// Implementation note:
//   - CreatePipeline always creates pipeline output branches such that the
//     pipeline's spec branch is in the pipeline output branch's provenance.
//   - CreatePipeline will always create a new output commit, but that's done
//     by CreateBranch at the bottom of the function, which sets the new output
//     branch provenance, rather than commitPipelineInfoFromFileSet higher up.
//   - This is because CreatePipeline calls hardStopPipeline towards the top,
//     breaking the provenance connection from the spec branch to the output
//     branch.
//   - For straightforward pipeline updates (e.g. new pipeline image)
//     stopping + updating + starting the pipeline isn't necessary.
//   - However it is necessary in many slightly atypical cases  (e.g. the
//     pipeline input changed: if the spec commit is created while the
//     output branch has its old provenance, or the output branch gets new
//     provenance while the old spec commit is the HEAD of the spec branch,
//     then an output commit will be created with provenance that doesn't
//     match its spec's PipelineInfo.Details.Input. Another example is when
//     request.Reprocess == true).
//   - Rather than try to enumerate every case where we can't create a spec
//     commit without stopping the pipeline, we just always stop the pipeline.
func (a *apiServer) CreatePipeline(ctx context.Context, request *pps.CreatePipelineRequest) (response *emptypb.Empty, retErr error) {
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "CreatePipeline")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	if request.Pipeline == nil {
		return nil, errors.New("request.Pipeline cannot be nil")
	}
	ensurePipelineProject(request.GetPipeline())

	// Annotate current span with pipeline & persist any extended trace to etcd
	span := opentracing.SpanFromContext(ctx)
	tracing.TagAnySpan(span, "project", request.Pipeline.Project.GetName(), "pipeline", request.Pipeline.Name)
	defer func() {
		tracing.TagAnySpan(span, "err", retErr)
	}()
	extended.PersistAny(ctx, a.env.EtcdClient, request.Pipeline)

	if err := a.validateEnterpriseChecks(ctx, request); err != nil {
		return nil, err
	}

	if err := a.validateSecret(ctx, request); err != nil {
		return nil, err
	}
	if request.Determined != nil {
		if err := a.CreateDetPipelineSideEffects(ctx, request.Pipeline, request.Determined.Workspaces); err != nil {
			return nil, errors.Wrap(err, "create det pipeline side effects")
		}
	}
	if err := a.txnEnv.WithTransaction(ctx, func(txn txnenv.Transaction) error {
		return errors.EnsureStack(txn.CreatePipeline(request))
	}); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// CreateDetPipelineSideEffects modifies state outside pachyderm's database involved in running determined/pachyderm pipelines.
// Provisions a determined user on representing the pipeline, and stores its password in a kubernetes secret named "{project}/{pipeline}-det"
//
// Implementation Notes:
// - This method must be idempotent, as it interfaces with Pachyderm's Transaction API that may run this multiple times
//
// TODO: set up garbage collection for the records stored outside the DB
func (a *apiServer) CreateDetPipelineSideEffects(ctx context.Context, pipeline *pps.Pipeline, workspaces []string) error {
	// check if pipeline's creds secret exists
	secretName := pipeline.Project.Name + "-" + pipeline.Name + "-det"
	password := uuid.NewWithoutDashes()
	whoAmI, err := a.env.AuthServer.WhoAmI(ctx, &auth.WhoAmIRequest{})
	if err != nil {
		return errors.Wrap(err, "who am i")
	}
	splits := strings.Split(whoAmI.Username, ":")
	if len(splits) != 2 {
		return errors.Errorf("subject %q expected to be segmented by one ':'", whoAmI.Username)
	}
	username := splits[1]
	if err := a.hookDeterminedPipeline(ctx, pipeline, workspaces, password, username); err != nil {
		return errors.Wrapf(err, "failed to connect pipeline %q to determined", pipeline.String())
	}
	if err := a.env.KubeClient.CoreV1().Secrets(a.namespace).Delete(ctx, secretName, metav1.DeleteOptions{}); err != nil {
		if !errutil.IsNotFoundError(err) {
			return errors.Wrapf(err, "clear pipeline's determined secret")
		}
	}
	s := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: a.namespace,
		},
		StringData: map[string]string{
			"password": password,
		},
	}
	s.SetLabels(map[string]string{
		"suite": "pachyderm",
	})
	if _, err := a.env.KubeClient.CoreV1().Secrets(a.namespace).Create(ctx, s, metav1.CreateOptions{}); err != nil {
		return errors.Wrapf(err, "failed to create pipeline's determined secret")
	}
	return nil
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
		Pipeline: request.Pipeline,
		Version:  1,
		Details: &pps.PipelineInfo_Details{
			Transform:               request.Transform,
			TfJob:                   request.TfJob,
			ParallelismSpec:         request.ParallelismSpec,
			Input:                   request.Input,
			OutputBranch:            request.OutputBranch,
			Egress:                  request.Egress,
			CreatedAt:               timestamppb.Now(),
			ResourceRequests:        request.ResourceRequests,
			ResourceLimits:          request.ResourceLimits,
			SidecarResourceLimits:   request.SidecarResourceLimits,
			SidecarResourceRequests: request.SidecarResourceRequests,
			Description:             request.Description,
			Salt:                    request.Salt,
			Service:                 request.Service,
			Spout:                   request.Spout,
			DatumSetSpec:            request.DatumSetSpec,
			DatumTimeout:            request.DatumTimeout,
			JobTimeout:              request.JobTimeout,
			DatumTries:              request.DatumTries,
			SchedulingSpec:          request.SchedulingSpec,
			PodSpec:                 request.PodSpec,
			PodPatch:                request.PodPatch,
			S3Out:                   request.S3Out,
			Metadata:                request.Metadata,
			ReprocessSpec:           request.ReprocessSpec,
			Autoscaling:             request.Autoscaling,
			Tolerations:             request.Tolerations,
			Determined:              request.Determined,
		},
	}
	if err := setPipelineDefaults(pipelineInfo); err != nil {
		return nil, err
	}
	// Validate final PipelineInfo (now that defaults have been populated)
	if err := a.validatePipeline(pipelineInfo); err != nil {
		return nil, err
	}
	if oldPipelineInfo != nil {
		// Modify pipelineInfo (increment Version, and *preserve Stopped* so
		// that updating a pipeline doesn't restart it)
		pipelineInfo.Version = oldPipelineInfo.Version + 1
		if oldPipelineInfo.Stopped {
			pipelineInfo.Stopped = true
		}
		if !request.Reprocess {
			pipelineInfo.Details.Salt = oldPipelineInfo.Details.Salt
		}
	}

	return pipelineInfo, nil
}

func (a *apiServer) CreatePipelineInTransaction(ctx context.Context, txnCtx *txncontext.TransactionContext, request *pps.CreatePipelineRequest) error {
	var (
		projectName          = request.Pipeline.Project.GetName()
		pipelineName         = request.Pipeline.Name
		oldPipelineInfo, err = a.InspectPipelineInTransaction(txnCtx, request.Pipeline)
	)
	if err != nil && !errutil.IsNotFoundError(err) {
		// silently ignore pipeline not found, old info will be nil
		return err
	}
	if oldPipelineInfo != nil && !request.Update {
		return ppsServer.ErrPipelineAlreadyExists{
			Pipeline: request.Pipeline,
		}
	}
	newPipelineInfo, err := a.initializePipelineInfo(request, oldPipelineInfo)
	if err != nil {
		return err
	}
	// Verify that all input repos exist (create cron repos if necessary).
	if visitErr := pps.VisitInput(newPipelineInfo.Details.Input, func(input *pps.Input) error {
		if input.Pfs != nil {
			if input.Pfs.Project == "" {
				input.Pfs.Project = projectName
			}
			if _, err := a.env.PFSServer.InspectRepoInTransaction(ctx, txnCtx,
				&pfs.InspectRepoRequest{
					Repo: client.NewSystemRepo(input.Pfs.Project, input.Pfs.Repo, input.Pfs.RepoType),
				},
			); err != nil {
				return errors.EnsureStack(err)
			}
		}
		if input.Cron != nil {
			if err := a.env.PFSServer.CreateRepoInTransaction(ctx, txnCtx,
				&pfs.CreateRepoRequest{
					Repo:        client.NewRepo(projectName, input.Cron.Repo),
					Description: fmt.Sprintf("Cron tick repo for pipeline %s.", request.Pipeline),
				},
			); err != nil && !errutil.IsAlreadyExistError(err) {
				return errors.EnsureStack(err)
			}
		}
		return nil
	}); visitErr != nil {
		return visitErr
	}
	update := request.Update && oldPipelineInfo != nil
	// Authorize pipeline creation
	operation := pipelineOpCreate
	if update {
		operation = pipelineOpUpdate
	}
	if err := a.authorizePipelineOpInTransaction(ctx, txnCtx, operation, newPipelineInfo.Details.Input, newPipelineInfo.Pipeline.Project.GetName(), newPipelineInfo.Pipeline.Name); err != nil {
		return err
	}

	var (
		// provenance for the pipeline's output branch (includes the spec branch)
		provenance = append(branchProvenance(newPipelineInfo.Pipeline.Project, newPipelineInfo.Details.Input),
			client.NewSystemRepo(projectName, pipelineName, pfs.SpecRepoType).NewBranch("master"))
		outputBranch = client.NewBranch(projectName, pipelineName, newPipelineInfo.Details.OutputBranch)
		metaBranch   = client.NewSystemRepo(projectName, pipelineName, pfs.MetaRepoType).NewBranch(newPipelineInfo.Details.OutputBranch)
	)

	// Get the expected number of workers for this pipeline
	if parallelism, err := getExpectedNumWorkers(newPipelineInfo); err != nil {
		return err
	} else {
		newPipelineInfo.Parallelism = uint64(parallelism)
	}

	newPipelineInfo.State = pps.PipelineState_PIPELINE_STARTING
	newPipelineInfo.Type = pipelineTypeFromInfo(newPipelineInfo)

	if !update {
		// Create output and spec repos
		if err := a.env.PFSServer.CreateRepoInTransaction(ctx, txnCtx,
			&pfs.CreateRepoRequest{
				Repo:        client.NewRepo(projectName, pipelineName),
				Description: fmt.Sprintf("Output repo for pipeline %s.", request.Pipeline),
			}); err != nil && !errutil.IsAlreadyExistError(err) {
			return errors.Wrapf(err, "error creating output repo for %s", pipelineName)
		} else if errutil.IsAlreadyExistError(err) {
			return errors.Errorf("pipeline %q cannot be created because a repo with the same name already exists", pipelineName)
		}
		if err := a.env.PFSServer.CreateRepoInTransaction(ctx, txnCtx,
			&pfs.CreateRepoRequest{
				Repo:        client.NewSystemRepo(projectName, pipelineName, pfs.SpecRepoType),
				Description: fmt.Sprintf("Spec repo for pipeline %s.", request.Pipeline),
				Update:      true,
			}); err != nil && !errutil.IsAlreadyExistError(err) {
			return errors.Wrapf(err, "error creating spec repo for %s", request.Pipeline)
		}
	}

	if request.SpecCommit != nil {
		if update {
			// request.SpecCommit indicates we're restoring from an extracted cluster
			// state and should not do any updates
			return errors.New("Cannot update a pipeline and provide a spec commit at the same time")
		}
		// Check if there is an existing spec commit
		commitInfo, err := a.env.PFSServer.InspectCommitInTransaction(txnCtx, &pfs.InspectCommitRequest{
			Commit: request.SpecCommit,
		})
		if err != nil {
			return errors.Wrap(err, "error inspecting spec commit")
		}
		// There is, so we use that as the spec commit, rather than making a new one
		newPipelineInfo.SpecCommit = commitInfo.Commit
	} else {
		// create an empty spec commit to mark the update
		newPipelineInfo.SpecCommit, err = a.env.PFSServer.StartCommitInTransaction(ctx, txnCtx, &pfs.StartCommitRequest{
			Branch: client.NewSystemRepo(projectName, pipelineName, pfs.SpecRepoType).NewBranch("master"),
		})
		if err != nil {
			return errors.EnsureStack(err)
		}
		if err := a.env.PFSServer.FinishCommitInTransaction(txnCtx, &pfs.FinishCommitRequest{
			Commit: newPipelineInfo.SpecCommit,
		}); err != nil {
			return errors.EnsureStack(err)
		}
	}
	// Generate new pipeline auth token (added due to & add pipeline to the ACLs of input/output repos
	if err := func() error {
		token, err := a.env.AuthServer.GetPipelineAuthTokenInTransaction(txnCtx, request.Pipeline)
		if err != nil {
			if auth.IsErrNotActivated(err) {
				return nil // no auth work to do
			}
			return errors.EnsureStack(err)
		}
		newPipelineInfo.AuthToken = token
		return nil
	}(); err != nil {
		return err
	}
	// store the new PipelineInfo in the collection
	if err := a.pipelines.ReadWrite(txnCtx.SqlTx).Create(newPipelineInfo.SpecCommit, newPipelineInfo); err != nil {
		return errors.EnsureStack(err)
	}

	if newPipelineInfo.AuthToken != "" {
		if err := a.fixPipelineInputRepoACLsInTransaction(txnCtx, newPipelineInfo, oldPipelineInfo); err != nil {
			return err
		}
	}

	if update {
		// Kill all unfinished jobs (as those are for the previous version and will
		// no longer be completed)
		if err := a.stopAllJobsInPipeline(txnCtx, request.Pipeline, "all jobs killed because pipeline was updated"); err != nil {
			return err
		}

		if newPipelineInfo.AuthToken != "" {
			// delete old auth token
			// refetch because inspect clears it
			var oldWithAuth pps.PipelineInfo
			if err := a.pipelines.ReadWrite(txnCtx.SqlTx).Get(oldPipelineInfo.SpecCommit, &oldWithAuth); err != nil {
				return errors.EnsureStack(err)
			}
			if _, err := a.env.AuthServer.RevokeAuthTokenInTransaction(txnCtx,
				&auth.RevokeAuthTokenRequest{Token: oldWithAuth.AuthToken}); err != nil {
				return errors.EnsureStack(err)
			}
		}
	}

	// A stopped pipeline should not have its provenance restored until it is
	// restarted.
	if newPipelineInfo.Stopped {
		provenance = nil
	}

	// Create or update the output branch (creating new output commit for the pipeline
	// and restarting the pipeline)
	if err := a.env.PFSServer.CreateBranchInTransaction(ctx, txnCtx, &pfs.CreateBranchRequest{
		Branch:     outputBranch,
		Provenance: provenance,
	}); err != nil {
		return errors.Wrapf(err, "could not create/update output branch")
	}

	if visitErr := pps.VisitInput(request.Input, func(input *pps.Input) error {
		if input.Pfs != nil && input.Pfs.Trigger != nil {
			var prevHead *pfs.Commit
			if branchInfo, err := a.env.PFSServer.InspectBranchInTransaction(txnCtx, &pfs.InspectBranchRequest{
				Branch: client.NewBranch(input.Pfs.Project, input.Pfs.Repo, input.Pfs.Branch),
			}); err != nil {
				if !errutil.IsNotFoundError(err) {
					return errors.EnsureStack(err)
				}
			} else {
				prevHead = branchInfo.Head
			}

			err := a.env.PFSServer.CreateBranchInTransaction(ctx, txnCtx, &pfs.CreateBranchRequest{
				Branch:  client.NewBranch(input.Pfs.Project, input.Pfs.Repo, input.Pfs.Branch),
				Head:    prevHead,
				Trigger: input.Pfs.Trigger,
			})
			return errors.EnsureStack(err)
		}
		return nil
	}); visitErr != nil {
		return errors.Wrapf(visitErr, "could not create/update trigger branch")
	}

	if request.Service == nil && request.Spout == nil {
		if err := a.env.PFSServer.CreateRepoInTransaction(ctx, txnCtx, &pfs.CreateRepoRequest{
			Repo:        metaBranch.Repo,
			Description: fmt.Sprint("Meta repo for pipeline ", pipelineName),
		}); err != nil && !errutil.IsAlreadyExistError(err) {
			return errors.Wrap(err, "could not create meta repo")
		}
		if err := a.env.PFSServer.CreateBranchInTransaction(ctx, txnCtx, &pfs.CreateBranchRequest{
			Branch:     metaBranch,
			Provenance: provenance, // same provenance as output branch
		}); err != nil {
			return errors.Wrapf(err, "could not create/update meta branch")
		}
	}
	return nil
}

func pipelineTypeFromInfo(pipelineInfo *pps.PipelineInfo) pps.PipelineInfo_PipelineType {
	if pipelineInfo.Details.Spout != nil {
		return pps.PipelineInfo_PIPELINE_TYPE_SPOUT
	} else if pipelineInfo.Details.Service != nil {
		return pps.PipelineInfo_PIPELINE_TYPE_SERVICE
	}
	return pps.PipelineInfo_PIPELINE_TYPE_TRANSFORM
}

// setPipelineDefaults sets the default values for a pipeline info
func setPipelineDefaults(pipelineInfo *pps.PipelineInfo) error {
	if pipelineInfo.Details.Transform.Image == "" {
		pipelineInfo.Details.Transform.Image = DefaultUserImage
	}
	setInputDefaults(pipelineInfo.Pipeline.Name, pipelineInfo.Details.Input)
	if pipelineInfo.Details.OutputBranch == "" {
		// Output branches default to master
		pipelineInfo.Details.OutputBranch = "master"
	}
	if pipelineInfo.Details.DatumTries == 0 {
		pipelineInfo.Details.DatumTries = DefaultDatumTries
	}
	if pipelineInfo.Details.Service != nil {
		if pipelineInfo.Details.Service.Type == "" {
			pipelineInfo.Details.Service.Type = string(v1.ServiceTypeNodePort)
		}
	}
	if pipelineInfo.Details.Spout != nil && pipelineInfo.Details.Spout.Service != nil && pipelineInfo.Details.Spout.Service.Type == "" {
		pipelineInfo.Details.Spout.Service.Type = string(v1.ServiceTypeNodePort)
	}
	if pipelineInfo.Details.ReprocessSpec == "" {
		pipelineInfo.Details.ReprocessSpec = client.ReprocessSpecUntilSuccess
	}
	return nil
}

func setInputDefaults(pipelineName string, input *pps.Input) {
	pps.SortInput(input)
	nCreatedBranches := make(map[string]int)
	pps.VisitInput(input, func(input *pps.Input) error { //nolint:errcheck
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
			if input.Pfs.Glob != "" {
				input.Pfs.Glob = pfsfile.CleanPath(input.Pfs.Glob)
			}
		}
		if input.Cron != nil {
			if input.Cron.Start == nil {
				input.Cron.Start = timestamppb.Now()
			}
			if input.Cron.Repo == "" {
				input.Cron.Repo = fmt.Sprintf("%s_%s", pipelineName, input.Cron.Name)
			}
		}
		return nil
	})
}

func (a *apiServer) stopAllJobsInPipeline(txnCtx *txncontext.TransactionContext, pipeline *pps.Pipeline, reason string) error {
	// Using ReadWrite here may load a large number of jobs inline in the
	// transaction, but doing an inconsistent read outside of the transaction
	// would be pretty sketchy (and we'd have to worry about trying to get another
	// postgres connection and possibly deadlocking).
	jobInfo := &pps.JobInfo{}
	sort := &col.Options{Target: col.SortByCreateRevision, Order: col.SortAscend}
	username := "unknown_username"
	if whoami, err := txnCtx.WhoAmI(); err == nil {
		username = whoami.Username
	}
	reason += " for user " + username
	err := a.jobs.ReadWrite(txnCtx.SqlTx).GetByIndex(ppsdb.JobsTerminalIndex, ppsdb.JobsTerminalKey(pipeline, false), jobInfo, sort, func(string) error {
		return a.stopJob(txnCtx, jobInfo.Job, reason)
	})
	return errors.EnsureStack(err)
}

func (a *apiServer) updatePipeline(
	txnCtx *txncontext.TransactionContext,
	pipeline *pps.Pipeline,
	info *pps.PipelineInfo,
	cb func() error) error {

	// get most recent pipeline key
	key, err := ppsutil.FindPipelineSpecCommitInTransaction(txnCtx, a.env.PFSServer, pipeline, "")
	if err != nil {
		return err
	}

	return errors.EnsureStack(a.pipelines.ReadWrite(txnCtx.SqlTx).Update(key, info, cb))
}

// InspectPipeline implements the protobuf pps.InspectPipeline RPC
func (a *apiServer) InspectPipeline(ctx context.Context, request *pps.InspectPipelineRequest) (response *pps.PipelineInfo, retErr error) {
	ensurePipelineProject(request.GetPipeline())
	return a.inspectPipeline(ctx, request.Pipeline, request.Details)
}

// inspectPipeline contains the functional implementation of InspectPipeline.
// Many functions (GetLogs, ListPipeline) need to inspect a pipeline, so they
// call this instead of making an RPC.
func (a *apiServer) inspectPipeline(ctx context.Context, pipeline *pps.Pipeline, details bool) (*pps.PipelineInfo, error) {
	var info *pps.PipelineInfo
	if err := a.txnEnv.WithReadContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
		var err error
		info, err = a.InspectPipelineInTransaction(txnCtx, pipeline)
		return err
	}); err != nil {
		return nil, err
	}

	if err := a.getLatestJobState(ctx, info); err != nil {
		return nil, err
	}
	if !details {
		info.Details = nil // preserve old behavior
	} else {
		kubeClient := a.env.KubeClient
		if info.Details.Service != nil {
			rcName := ppsutil.PipelineRcName(info)
			service, err := kubeClient.CoreV1().Services(a.namespace).Get(ctx, fmt.Sprintf("%s-user", rcName), metav1.GetOptions{})
			if err != nil {
				if !errutil.IsNotFoundError(err) {
					return nil, errors.EnsureStack(err)
				}
			} else {
				info.Details.Service.Ip = service.Spec.ClusterIP
			}
		}

		workerStatus, err := workerserver.Status(ctx, info, a.env.EtcdClient, a.etcdPrefix, a.workerGrpcPort)
		if err != nil {
			log.Error(ctx, "failed to get worker status", zap.Error(err))
		} else {
			info.Details.WorkersAvailable = int64(len(workerStatus))
			info.Details.WorkersRequested = int64(info.Parallelism)
		}
		tasks, claims, err := task.Count(ctx, a.env.TaskService, driver.TaskNamespace(info), "")
		if err != nil {
			return nil, errors.EnsureStack(err)
		}
		info.Details.UnclaimedTasks = tasks - claims
	}

	return info, nil
}

// InspectPipelineInTransaction implements the APIServer interface.
func (a *apiServer) InspectPipelineInTransaction(txnCtx *txncontext.TransactionContext, pipeline *pps.Pipeline) (*pps.PipelineInfo, error) {
	// The pipeline pipelineName arrived in ancestry format; need to turn it into a simple pipelineName.
	pipelineName, ancestors, err := ancestry.Parse(pipeline.Name)
	if err != nil {
		return nil, err
	}

	if ancestors < 0 {
		return nil, errors.New("cannot inspect future pipelines")
	}

	pipeline.Name = pipelineName
	commit, err := ppsutil.FindPipelineSpecCommitInTransaction(txnCtx, a.env.PFSServer, pipeline, "")
	if err != nil {
		return nil, errors.Wrapf(err, "pipeline was not inspected: couldn't find up to date spec for pipeline %q", pipeline)
	}

	pipelineInfo := &pps.PipelineInfo{}
	if err := a.pipelines.ReadWrite(txnCtx.SqlTx).Get(commit, pipelineInfo); err != nil {
		if col.IsErrNotFound(err) {
			return nil, ppsServer.ErrPipelineNotFound{Pipeline: pipeline}
		}
		return nil, errors.EnsureStack(err)
	}
	if ancestors > 0 {
		targetVersion := int(pipelineInfo.Version) - ancestors
		if targetVersion < 1 {
			return nil, errors.Errorf("pipeline %q has only %d versions, not enough to find ancestor %d", pipeline, pipelineInfo.Version, ancestors)
		}
		if err := a.pipelines.ReadWrite(txnCtx.SqlTx).GetUniqueByIndex(ppsdb.PipelinesVersionIndex, ppsdb.VersionKey(pipeline, uint64(targetVersion)), pipelineInfo); err != nil {
			return nil, errors.EnsureStack(err)
		}
	}

	// Erase any AuthToken - this shouldn't be returned to anyone (the workers
	// won't use this function to get their auth token)
	pipelineInfo.AuthToken = ""
	return pipelineInfo, nil
}

// ListPipeline implements the protobuf pps.ListPipeline RPC
func (a *apiServer) ListPipeline(request *pps.ListPipelineRequest, srv pps.API_ListPipelineServer) (retErr error) {
	ensurePipelineProject(request.GetPipeline())
	return a.listPipeline(srv.Context(), request, srv.Send)
}

func (a *apiServer) getLatestJobState(ctx context.Context, info *pps.PipelineInfo) error {
	// fill in state of most-recently-created job (the first shown in list job)
	opts := col.DefaultOptions()
	opts.Limit = 1
	var job pps.JobInfo
	err := a.jobs.ReadOnly(ctx).GetByIndex(
		ppsdb.JobsPipelineIndex,
		ppsdb.JobsPipelineKey(info.Pipeline),
		&job,
		opts, func(_ string) error {
			info.LastJobState = job.State
			return errutil.ErrBreak // not strictly necessary because we are limiting to 1 already
		})
	return errors.EnsureStack(err)
}

func (a *apiServer) listPipeline(ctx context.Context, request *pps.ListPipelineRequest, f func(*pps.PipelineInfo) error) error {
	filterPipeline, err := newMessageFilterFunc(request.GetJqFilter(), request.GetProjects())
	if err != nil {
		return errors.Wrap(err, "error creating message filter function")
	}

	// Helper func to check whether a user is allowed to see the given pipeline in the result.
	// Cache the project level access because it applies to every pipeline within the same project.
	checkProjectAccess := miscutil.CacheFunc(func(project string) error {
		return a.env.AuthServer.CheckProjectIsAuthorized(ctx, &pfs.Project{Name: project}, auth.Permission_PROJECT_LIST_REPO)
	}, 100 /* size */)
	checkAccess := func(ctx context.Context, pipeline *pps.Pipeline) error {
		if err := checkProjectAccess(pipeline.Project.GetName()); err != nil {
			if !errors.As(err, &auth.ErrNotAuthorized{}) {
				return err
			}
			// Use the pipeline's output repo as the reference to the pipeline itself
			return a.env.AuthServer.CheckRepoIsAuthorized(ctx, &pfs.Repo{Name: pipeline.Name, Type: pfs.UserRepoType, Project: pipeline.Project}, auth.Permission_REPO_READ)
		}
		return nil
	}

	loadPipelineAtCommit := func(pi *pps.PipelineInfo, commitSetID string) error { // mutates pi
		return a.txnEnv.WithReadContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
			key, err := ppsutil.FindPipelineSpecCommitInTransaction(txnCtx, a.env.PFSServer, pi.Pipeline, commitSetID)
			if err != nil {
				return errors.Wrapf(err, "couldn't find up to date spec for pipeline %q", pi.Pipeline)
			}
			if err := a.pipelines.ReadWrite(txnCtx.SqlTx).Get(key, pi); err != nil {
				return errors.EnsureStack(err)
			}
			var ji pps.JobInfo
			return a.jobs.ReadWrite(txnCtx.SqlTx).GetByIndex(ppsdb.JobsPipelineIndex, pi.Pipeline.String(), &ji, col.DefaultOptions(), func(string) error {
				if ji.PipelineVersion > pi.Version {
					return nil
				}
				pi.LastJobState = ji.State
				return errutil.ErrBreak
			})
		})
	}
	if err := ppsutil.ListPipelineInfo(ctx, a.pipelines, request.Pipeline, request.History, func(pi *pps.PipelineInfo) error {
		if ok, err := filterPipeline(ctx, pi); err != nil {
			return errors.Wrap(err, "error filtering pipeline")
		} else if !ok {
			return nil
		}
		if err := checkAccess(ctx, pi.Pipeline); err != nil {
			if !errors.As(err, &auth.ErrNotAuthorized{}) {
				return errors.Wrapf(err, "could not check user is authorized to list pipeline, problem with pipeline %s", pi.Pipeline)
			}
			return nil
		}
		if request.GetCommitSet().GetId() != "" {
			// load pipeline as it was at the commit set along with its associated job state
			if err := loadPipelineAtCommit(pi, request.GetCommitSet().GetId()); err != nil {
				if pfsServer.IsCommitNotFoundErr(err) || ppsServer.IsPipelineNotFoundErr(err) || col.IsErrNotFound(err) {
					return nil
				}
				return err
			}
		} else {
			if err := a.getLatestJobState(ctx, pi); err != nil {
				return err
			}
		}
		return f(pi)
	}); err != nil && err != errutil.ErrBreak {
		if errors.Is(err, context.DeadlineExceeded) {
			return status.Error(codes.DeadlineExceeded, err.Error())
		}
		return errors.Wrap(err, "could not list pipeline info")
	}
	return nil
}

// DeletePipeline implements the protobuf pps.DeletePipeline RPC
func (a *apiServer) DeletePipeline(ctx context.Context, request *pps.DeletePipelineRequest) (response *emptypb.Empty, retErr error) {
	if request != nil && request.Pipeline != nil {
		ensurePipelineProject(request.GetPipeline())
	}
	if request.All { //nolint:staticcheck
		_, err := a.DeletePipelines(ctx, &pps.DeletePipelinesRequest{
			KeepRepo: request.KeepRepo,
		})
		return &emptypb.Empty{}, errors.Wrap(err, "delete all pipelines")
	} else if err := a.deletePipeline(ctx, request); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (a *apiServer) deletePipeline(ctx context.Context, request *pps.DeletePipelineRequest) error {
	pipelineName := request.Pipeline.Name
	// stop the pipeline to avoid interference from new jobs
	if _, err := a.StopPipeline(ctx,
		&pps.StopPipelineRequest{Pipeline: request.Pipeline}); err != nil && errutil.IsNotFoundError(err) {
		log.Error(ctx, "failed to stop pipeline, continuing with delete", zap.Error(err))
	} else if err != nil {
		return errors.Wrapf(err, "error stopping pipeline %s", request.Pipeline)
	}
	// perform the rest of the deletion in a transaction
	var deleteErr error
	if err := a.txnEnv.WithWriteContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
		var deleteRepos []*pfs.Repo
		deleteRepos, deleteErr = a.deletePipelineInTransaction(ctx, txnCtx, request)
		if deleteErr != nil {
			return deleteErr
		}
		return a.env.PFSServer.DeleteReposInTransaction(ctx, txnCtx, deleteRepos, request.Force)
	}); err != nil {
		return err
	}
	// FIXME(1101): make project-aware
	clearJobCache(a.env.GetPachClient(ctx), pipelineName)
	clearJobCache(a.env.GetPachClient(ctx), request.Pipeline.String())
	return deleteErr
}

func (a *apiServer) deletePipelineInTransaction(ctx context.Context, txnCtx *txncontext.TransactionContext, request *pps.DeletePipelineRequest) ([]*pfs.Repo, error) {
	projectName := request.Pipeline.Project.GetName()
	pipelineName := request.Pipeline.Name
	pipelinesNameKey := ppsdb.PipelinesNameKey(request.Pipeline)
	var deleteRepos []*pfs.Repo
	// make sure the pipeline exists
	var foundPipeline bool
	if err := a.pipelines.ReadWrite(txnCtx.SqlTx).GetByIndex(
		ppsdb.PipelinesNameIndex,
		pipelinesNameKey,
		&pps.PipelineInfo{},
		col.DefaultOptions(),
		func(_ string) error {
			foundPipeline = true
			return nil
		}); err != nil {
		return nil, errors.Wrapf(err, "error checking if pipeline %s/%s exists", projectName, pipelineName)
	}
	if !foundPipeline {
		// nothing to delete
		return nil, nil
	}
	pipelineInfo := &pps.PipelineInfo{}
	// Try to retrieve PipelineInfo for this pipeline. If we see a not found error,
	// we will still try to delete what we can because we know there is a pipeline
	if specCommit, err := ppsutil.FindPipelineSpecCommitInTransaction(txnCtx, a.env.PFSServer, request.Pipeline, ""); err == nil {
		if err := a.pipelines.ReadWrite(txnCtx.SqlTx).Get(specCommit, pipelineInfo); err != nil && !col.IsErrNotFound(err) {
			return nil, errors.EnsureStack(err)
		}
	} else if !errutil.IsNotFoundError(err) && !auth.IsErrNoRoleBinding(err) {
		return nil, err
	}
	// check if the output repo exists--if not, the pipeline is non-functional and
	// the rest of the delete operation continues without any auth checks
	if _, err := a.env.PFSServer.InspectRepoInTransaction(ctx, txnCtx, &pfs.InspectRepoRequest{
		Repo: client.NewRepo(projectName, pipelineName)}); err != nil && !auth.IsErrNoRoleBinding(err) {
		return nil, errors.EnsureStack(err)
	} else if err == nil {
		// Check if the caller is authorized to delete this pipeline
		if err := a.authorizePipelineOpInTransaction(ctx, txnCtx, pipelineOpDelete, pipelineInfo.GetDetails().GetInput(), projectName, pipelineName); err != nil {
			return nil, err
		}
	}
	// If necessary, revoke the pipeline's auth token and remove it from its inputs' ACLs
	// If auth is deactivated, don't bother doing either
	if _, err := txnCtx.WhoAmI(); err == nil && pipelineInfo.AuthToken != "" {
		// 'pipelineInfo' == nil => remove pipeline from all input repos
		if pipelineInfo.Details != nil {
			// need details for acls
			if err := a.fixPipelineInputRepoACLsInTransaction(txnCtx, nil, pipelineInfo); err != nil {
				return nil, errors.Wrapf(err, "error fixing repo ACLs for pipeline %q", pipelineName)
			}
		}
		if _, err := a.env.AuthServer.RevokeAuthTokenInTransaction(txnCtx,
			&auth.RevokeAuthTokenRequest{
				Token: pipelineInfo.AuthToken,
			}); err != nil {
			return nil, errors.EnsureStack(err)
		}
	}
	// Delete all of the pipeline's jobs - we shouldn't need to worry about any
	// new jobs since the pipeline has already been stopped.
	jobInfo := &pps.JobInfo{}
	if err := a.jobs.ReadWrite(txnCtx.SqlTx).GetByIndex(ppsdb.JobsPipelineIndex, ppsdb.JobsPipelineKey(request.Pipeline), jobInfo, col.DefaultOptions(), func(string) error {
		job := proto.Clone(jobInfo.Job).(*pps.Job)
		err := a.deleteJobInTransaction(txnCtx, &pps.DeleteJobRequest{Job: job})
		if errutil.IsNotFoundError(err) || auth.IsErrNoRoleBinding(err) {
			return nil
		}
		return err
	}); err != nil {
		return nil, errors.EnsureStack(err)
	}
	// Delete all past PipelineInfos
	if err := a.pipelines.ReadWrite(txnCtx.SqlTx).DeleteByIndex(ppsdb.PipelinesNameIndex, pipelinesNameKey); err != nil {
		return nil, errors.Wrapf(err, "collection.Delete")
	}
	if !request.KeepRepo {
		// delete the pipeline's output repo
		deleteRepos = append(deleteRepos, client.NewRepo(projectName, pipelineName))
	} else {
		// Remove branch provenance from output and then delete meta and spec repos
		// this leaves the repo as a source repo, eliminating pipeline metadata
		// need details for output branch, presumably if we don't have them the spec repo is gone, anyway
		if pipelineInfo.Details != nil {
			if err := a.env.PFSServer.CreateBranchInTransaction(ctx, txnCtx, &pfs.CreateBranchRequest{
				Branch: client.NewBranch(projectName, pipelineName, pipelineInfo.Details.OutputBranch),
			}); err != nil {
				return nil, errors.EnsureStack(err)
			}
		}
	}
	if _, err := a.env.PFSServer.InspectRepoInTransaction(ctx, txnCtx, &pfs.InspectRepoRequest{Repo: client.NewSystemRepo(projectName, pipelineName, pfs.SpecRepoType)}); err == nil {
		deleteRepos = append(deleteRepos, client.NewSystemRepo(projectName, pipelineName, pfs.SpecRepoType))
	} else if !errutil.IsNotFoundError(err) {
		return nil, errors.Wrapf(err, "inspect spec repo for pipeline %q", pipelineName)
	}
	if _, err := a.env.PFSServer.InspectRepoInTransaction(ctx, txnCtx, &pfs.InspectRepoRequest{Repo: client.NewSystemRepo(projectName, pipelineName, pfs.MetaRepoType)}); err == nil {
		deleteRepos = append(deleteRepos, client.NewSystemRepo(projectName, pipelineName, pfs.MetaRepoType))
	} else if !errutil.IsNotFoundError(err) {
		return nil, errors.Wrapf(err, "inspect meta repo for pipeline %q", pipelineName)
	}
	// delete cron after main repo is deleted or has provenance removed
	// cron repos are only used to trigger jobs, so don't keep them even with KeepRepo
	if pipelineInfo.Details != nil {
		if err := pps.VisitInput(pipelineInfo.Details.Input, func(input *pps.Input) error {
			if input.Cron != nil {
				deleteRepos = append(deleteRepos, client.NewRepo(projectName, input.Cron.Repo))
			}
			return nil
		}); err != nil {
			return nil, err
		}
	}
	return deleteRepos, nil
}

// DeletePipelines implements the protobuf pps.DeletePipelines RPC.  It deletes
// multiple pipelines at once.  If all is set, then all pipelines in all
// projects are deleted; otherwise the requested pipelines are deleted.
func (a *apiServer) DeletePipelines(ctx context.Context, request *pps.DeletePipelinesRequest) (response *pps.DeletePipelinesResponse, retErr error) {
	var (
		projects = make(map[string]bool)
		dr       = &pps.DeletePipelineRequest{
			Force:    request.Force,
			KeepRepo: request.KeepRepo,
		}
		ps []*pps.Pipeline
	)
	for _, p := range request.GetProjects() {
		projects[p.String()] = true
	}
	dr.Pipeline = &pps.Pipeline{}
	pipelineInfo := &pps.PipelineInfo{}
	deleted := make(map[string]struct{})
	if err := a.pipelines.ReadOnly(ctx).List(pipelineInfo, col.DefaultOptions(), func(string) error {
		if _, ok := deleted[pipelineInfo.Pipeline.String()]; ok {
			// while the delete pipeline call will delete historical versions,
			// they could still show up in the list.  Ignore them.
			return nil
		}
		if !request.GetAll() && !projects[pipelineInfo.GetPipeline().GetProject().GetName()] {
			return nil
		}
		deleted[pipelineInfo.Pipeline.String()] = struct{}{}
		ps = append(ps, pipelineInfo.Pipeline)
		return nil
	}); err != nil {
		return nil, errors.EnsureStack(err)
	}
	for _, p := range ps {
		if _, err := a.StopPipeline(ctx, &pps.StopPipelineRequest{Pipeline: p}); err != nil {
			return nil, errors.Wrapf(err, "stop pipeline %q", p.String())
		}

	}
	if err := a.env.TxnEnv.WithWriteContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
		var rs []*pfs.Repo
		for _, p := range ps {
			deleteRepos, err := a.deletePipelineInTransaction(ctx, txnCtx, &pps.DeletePipelineRequest{Pipeline: p, KeepRepo: request.KeepRepo})
			if err != nil {
				return err
			}
			rs = append(rs, deleteRepos...)
		}
		return a.env.PFSServer.DeleteReposInTransaction(ctx, txnCtx, rs, request.Force)
	}); err != nil {
		return nil, err
	}
	return &pps.DeletePipelinesResponse{Pipelines: ps}, nil
}

// StartPipeline implements the protobuf pps.StartPipeline RPC
func (a *apiServer) StartPipeline(ctx context.Context, request *pps.StartPipelineRequest) (response *emptypb.Empty, retErr error) {
	if request.Pipeline == nil {
		return nil, errors.New("request.Pipeline cannot be nil")
	}
	ensurePipelineProject(request.Pipeline)

	if err := a.txnEnv.WithWriteContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
		pipelineInfo, err := a.InspectPipelineInTransaction(txnCtx, request.Pipeline)
		if err != nil {
			return err
		}

		// check if the caller is authorized to update this pipeline
		if err := a.authorizePipelineOpInTransaction(ctx, txnCtx, pipelineOpStartStop, pipelineInfo.Details.Input, pipelineInfo.Pipeline.Project.GetName(), pipelineInfo.Pipeline.Name); err != nil {
			return err
		}

		// Restore branch provenance, which may create a new output commit/job
		provenance := append(branchProvenance(pipelineInfo.Pipeline.Project, pipelineInfo.Details.Input),
			client.NewSystemRepo(pipelineInfo.Pipeline.Project.GetName(), pipelineInfo.Pipeline.Name, pfs.SpecRepoType).NewBranch("master"))
		if err := a.env.PFSServer.CreateBranchInTransaction(ctx, txnCtx, &pfs.CreateBranchRequest{
			Branch:     client.NewBranch(pipelineInfo.Pipeline.Project.GetName(), pipelineInfo.Pipeline.Name, pipelineInfo.Details.OutputBranch),
			Provenance: provenance,
		}); err != nil {
			return errors.EnsureStack(err)
		}
		// restore same provenance to meta repo
		if pipelineInfo.Details.Spout == nil && pipelineInfo.Details.Service == nil {
			if err := a.env.PFSServer.CreateBranchInTransaction(ctx, txnCtx, &pfs.CreateBranchRequest{
				Branch:     client.NewSystemRepo(pipelineInfo.Pipeline.Project.GetName(), pipelineInfo.Pipeline.Name, pfs.MetaRepoType).NewBranch(pipelineInfo.Details.OutputBranch),
				Provenance: provenance,
			}); err != nil {
				return errors.EnsureStack(err)
			}
		}

		newPipelineInfo := &pps.PipelineInfo{}
		return a.updatePipeline(txnCtx, pipelineInfo.Pipeline, newPipelineInfo, func() error {
			newPipelineInfo.Stopped = false
			return nil
		})
	}); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// StopPipeline implements the protobuf pps.StopPipeline RPC
func (a *apiServer) StopPipeline(ctx context.Context, request *pps.StopPipelineRequest) (response *emptypb.Empty, retErr error) {
	if request.Pipeline == nil {
		return nil, errors.New("request.Pipeline cannot be nil")
	}
	ensurePipelineProject(request.Pipeline)
	if err := a.txnEnv.WithWriteContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
		if pipelineInfo, err := a.InspectPipelineInTransaction(txnCtx, request.Pipeline); err == nil {
			// check if the caller is authorized to update this pipeline
			// don't pass in the input - stopping the pipeline means they won't be read anymore,
			// so we don't need to check any permissions
			if err := a.authorizePipelineOpInTransaction(ctx, txnCtx, pipelineOpStartStop, pipelineInfo.Details.Input, pipelineInfo.Pipeline.Project.GetName(), pipelineInfo.Pipeline.Name); err != nil {
				return err
			}

			// Remove branch provenance to prevent new output and meta commits from being created
			if err := a.env.PFSServer.CreateBranchInTransaction(ctx, txnCtx, &pfs.CreateBranchRequest{
				Branch:     client.NewBranch(pipelineInfo.Pipeline.Project.GetName(), pipelineInfo.Pipeline.Name, pipelineInfo.Details.OutputBranch),
				Provenance: nil,
			}); err != nil {
				return errors.EnsureStack(err)
			}
			if pipelineInfo.Details.Spout == nil && pipelineInfo.Details.Service == nil {
				if err := a.env.PFSServer.CreateBranchInTransaction(ctx, txnCtx, &pfs.CreateBranchRequest{
					Branch:     client.NewSystemRepo(pipelineInfo.Pipeline.Project.GetName(), pipelineInfo.Pipeline.Name, pfs.MetaRepoType).NewBranch(pipelineInfo.Details.OutputBranch),
					Provenance: nil,
				}); err != nil {
					return errors.EnsureStack(err)
				}
			}

			newPipelineInfo := &pps.PipelineInfo{}
			if err := a.updatePipeline(txnCtx, pipelineInfo.Pipeline, newPipelineInfo, func() error {
				newPipelineInfo.Stopped = true
				return nil
			}); err != nil {
				return err
			}
		} else if !errutil.IsNotFoundError(err) {
			return err
		}

		// Kill any remaining jobs
		// if the pipeline output repo doesn't exist, we technically run this without authorization,
		// but it's not clear what authorization means in that case, and those jobs are doomed, anyway
		return a.stopAllJobsInPipeline(txnCtx, request.Pipeline, "all jobs killed because pipeline was stopped")
	}); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (a *apiServer) RunPipeline(ctx context.Context, request *pps.RunPipelineRequest) (response *emptypb.Empty, retErr error) {
	return nil, errors.New("unimplemented")
}

func (a *apiServer) RunCron(ctx context.Context, request *pps.RunCronRequest) (response *emptypb.Empty, retErr error) {
	ensurePipelineProject(request.GetPipeline())
	pipelineInfo, err := a.inspectPipeline(ctx, request.Pipeline, true)
	if err != nil {
		return nil, err
	}

	if pipelineInfo.Details.Input == nil {
		return nil, errors.Errorf("pipeline doesn't have a cron input")
	}

	// find any cron inputs
	var crons []*pps.CronInput
	if err := pps.VisitInput(pipelineInfo.Details.Input, func(in *pps.Input) error {
		if in.Cron != nil {
			crons = append(crons, in.Cron)
		}
		return nil
	}); err != nil {
		return nil, errors.Errorf("error visiting pps inputs: %v", err)
	}

	if len(crons) < 1 {
		return nil, errors.Errorf("pipeline doesn't have a cron input")
	}

	// put the same time for all ticks
	now := time.Now()

	// add all the ticks. These will be in separate transactions if there are more than one
	for _, c := range crons {
		if err := cronTick(a.env.GetPachClient(ctx), now, c); err != nil {
			return nil, err
		}
	}

	return &emptypb.Empty{}, nil
}

func (a *apiServer) propagateJobs(txnCtx *txncontext.TransactionContext) error {
	commitInfos, err := a.env.PFSServer.InspectCommitSetInTransaction(txnCtx, client.NewCommitSet(txnCtx.CommitSetID), false)
	if err != nil {
		return errors.EnsureStack(err)
	}
	for _, commitInfo := range commitInfos {
		// Skip alias commits and any commits which have already been finished
		if commitInfo.Finishing != nil {
			continue
		}
		// Skip commits from system repos
		if commitInfo.Commit.Repo.Type != pfs.UserRepoType {
			continue
		}
		// Skip commits from repos that have no associated pipeline
		var pipelineInfo *pps.PipelineInfo
		if pipelineInfo, err = a.InspectPipelineInTransaction(txnCtx, pps.RepoPipeline(commitInfo.Commit.Repo)); err != nil {
			if col.IsErrNotFound(err) {
				continue
			}
			return err
		}
		// Don't create jobs for spouts
		if pipelineInfo.Type == pps.PipelineInfo_PIPELINE_TYPE_SPOUT {
			continue
		}
		// Don't create jobs for commits not on the output branch.
		if commitInfo.Commit.Branch.Name != pipelineInfo.Details.OutputBranch {
			continue
		}
		// Check if there is an existing job for the output commit
		job := client.NewJob(pipelineInfo.Pipeline.Project.GetName(), pipelineInfo.Pipeline.Name, txnCtx.CommitSetID)
		jobInfo := &pps.JobInfo{}
		if err := a.jobs.ReadWrite(txnCtx.SqlTx).Get(ppsdb.JobKey(job), jobInfo); err == nil {
			continue // Job already exists, skip it
		} else if !col.IsErrNotFound(err) {
			return errors.EnsureStack(err)
		}

		token := ""
		if _, err := txnCtx.WhoAmI(); err == nil {
			// If auth is active, generate an auth token for the job
			token, err = a.env.AuthServer.GetPipelineAuthTokenInTransaction(txnCtx, pipelineInfo.Pipeline)
			if err != nil {
				return errors.EnsureStack(err)
			}
		}

		pipelines := a.pipelines.ReadWrite(txnCtx.SqlTx)
		jobs := a.jobs.ReadWrite(txnCtx.SqlTx)
		jobPtr := &pps.JobInfo{
			Job:             job,
			PipelineVersion: pipelineInfo.Version,
			OutputCommit:    commitInfo.Commit,
			Stats:           &pps.ProcessStats{},
			AuthToken:       token,
			Created:         timestamppb.Now(),
		}
		if err := ppsutil.UpdateJobState(pipelines, jobs, jobPtr, pps.JobState_JOB_CREATED, ""); err != nil {
			return err
		}
	}
	return nil
}

// CreateSecret implements the protobuf pps.CreateSecret RPC
func (a *apiServer) CreateSecret(ctx context.Context, request *pps.CreateSecretRequest) (response *emptypb.Empty, retErr error) {
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

	if _, err := a.env.KubeClient.CoreV1().Secrets(a.namespace).Create(ctx, &s, metav1.CreateOptions{}); err != nil {
		return nil, errors.Wrapf(err, "failed to create secret")
	}
	return &emptypb.Empty{}, nil
}

// DeleteSecret implements the protobuf pps.DeleteSecret RPC
func (a *apiServer) DeleteSecret(ctx context.Context, request *pps.DeleteSecretRequest) (response *emptypb.Empty, retErr error) {
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "DeleteSecret")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	if err := a.env.KubeClient.CoreV1().Secrets(a.namespace).Delete(ctx, request.Secret.Name, metav1.DeleteOptions{}); err != nil {
		return nil, errors.Wrapf(err, "failed to delete secret")
	}
	return &emptypb.Empty{}, nil
}

// InspectSecret implements the protobuf pps.InspectSecret RPC
func (a *apiServer) InspectSecret(ctx context.Context, request *pps.InspectSecretRequest) (response *pps.SecretInfo, retErr error) {
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "InspectSecret")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	secret, err := a.env.KubeClient.CoreV1().Secrets(a.namespace).Get(ctx, request.Secret.Name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get secret")
	}
	return &pps.SecretInfo{
		Secret: &pps.Secret{
			Name: secret.Name,
		},
		Type:              string(secret.Type),
		CreationTimestamp: timestamppb.New(secret.GetCreationTimestamp().Time),
	}, nil
}

// ListSecret implements the protobuf pps.ListSecret RPC
func (a *apiServer) ListSecret(ctx context.Context, in *emptypb.Empty) (response *pps.SecretInfos, retErr error) {
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "ListSecret")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	secrets, err := a.env.KubeClient.CoreV1().Secrets(a.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: "secret-source=pachyderm-user",
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to list secrets")
	}
	secretInfos := []*pps.SecretInfo{}
	for _, s := range secrets.Items {
		secretInfos = append(secretInfos, &pps.SecretInfo{
			Secret: &pps.Secret{
				Name: s.Name,
			},
			Type:              string(s.Type),
			CreationTimestamp: timestamppb.New(s.GetCreationTimestamp().Time),
		})
	}

	return &pps.SecretInfos{
		SecretInfo: secretInfos,
	}, nil
}

// DeleteAll implements the protobuf pps.DeleteAll RPC
func (a *apiServer) DeleteAll(ctx context.Context, request *emptypb.Empty) (response *emptypb.Empty, retErr error) {
	if _, err := a.DeletePipelines(ctx, &pps.DeletePipelinesRequest{All: true, Force: true}); err != nil {
		return nil, err
	}

	if err := a.env.KubeClient.CoreV1().Secrets(a.namespace).DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{
		LabelSelector: "secret-source=pachyderm-user",
	}); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return &emptypb.Empty{}, nil
}

// ActivateAuth implements the protobuf pps.ActivateAuth RPC
func (a *apiServer) ActivateAuth(ctx context.Context, req *pps.ActivateAuthRequest) (*pps.ActivateAuthResponse, error) {
	var resp *pps.ActivateAuthResponse
	if err := a.txnEnv.WithWriteContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
		var err error
		resp, err = a.ActivateAuthInTransaction(txnCtx, req)
		return err
	}); err != nil {
		return nil, err
	}
	return resp, nil
}

func (a *apiServer) ActivateAuthInTransaction(txnCtx *txncontext.TransactionContext, req *pps.ActivateAuthRequest) (resp *pps.ActivateAuthResponse, retErr error) {
	// Unauthenticated users can't create new pipelines or repos, and users can't
	// log in while auth is in an intermediate state, so 'pipelines' is exhaustive
	pi := &pps.PipelineInfo{}
	if err := a.pipelines.ReadWrite(txnCtx.SqlTx).List(pi, col.DefaultOptions(), func(string) error {
		// 1) Create a new auth token for 'pipeline' and attach it, so that the
		// pipeline can authenticate as itself when it needs to read input data
		token, err := a.env.AuthServer.GetPipelineAuthTokenInTransaction(txnCtx, pi.Pipeline)
		if err != nil {
			return errors.Wrapf(grpcutil.ScrubGRPC(err), "could not generate pipeline auth token")
		}
		if err := a.updatePipeline(txnCtx, pi.Pipeline, pi, func() error {
			pi.AuthToken = token
			return nil
		}); err != nil {
			return errors.Wrapf(err, "could not update %q with new auth token", pi.Pipeline)
		}
		// put 'pipeline' on relevant ACLs
		if err := a.fixPipelineInputRepoACLsInTransaction(txnCtx, pi, nil); err != nil {
			return errors.Wrapf(err, "fix repo ACLs for pipeline %q", pi.Pipeline)
		}
		return nil
	}); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return &pps.ActivateAuthResponse{}, nil
}

// RunLoadTest implements the pps.RunLoadTest RPC
func (a *apiServer) RunLoadTest(ctx context.Context, req *pps.RunLoadTestRequest) (_ *pps.RunLoadTestResponse, retErr error) {
	pachClient := a.env.GetPachClient(ctx)
	if req.DagSpec == "" {
		req.DagSpec = defaultDagSpecs[0]
	}
	return ppsload.Pipeline(pachClient, req)
}

// RunLoadTestDefault implements the pps.RunLoadTestDefault RPC
func (a *apiServer) RunLoadTestDefault(ctx context.Context, _ *emptypb.Empty) (_ *pps.RunLoadTestResponse, retErr error) {
	pachClient := a.env.GetPachClient(ctx)
	return ppsload.Pipeline(pachClient, &pps.RunLoadTestRequest{
		DagSpec:  defaultDagSpecs[0],
		LoadSpec: defaultLoadSpecs[0],
	})
}

var defaultDagSpecs = []string{`
default-load-test-source:
default-load-test-pipeline: default-load-test-source
`}

var defaultLoadSpecs = []string{`
count: 5
modifications:
  - count: 5
    putFile:
      count: 5
      source: "random"
fileSources:
  - name: "random"
    random:
      directory:
        depth:
          min: 0
          max: 3
        run: 3
      sizes:
        - min: 1000
          max: 10000
          prob: 30
        - min: 10000
          max: 100000
          prob: 30
        - min: 1000000
          max: 10000000
          prob: 30
        - min: 10000000
          max: 100000000
          prob: 10
validator: {}
`}

// RepoNameToEnvString is a helper which uppercases a repo name for
// use in environment variable names.
func RepoNameToEnvString(repoName string) string {
	return strings.ToUpper(repoName)
}

func (a *apiServer) listPods(ctx context.Context, labels labels.Set) ([]v1.Pod, error) {
	podList, err := a.env.KubeClient.CoreV1().Pods(a.namespace).List(ctx, metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ListOptions",
			APIVersion: "v1",
		},
		LabelSelector: metav1.FormatLabelSelector(metav1.SetAsLabelSelector(labels)),
	})
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	return podList.Items, nil
}

func (a *apiServer) pachdPods(ctx context.Context) ([]v1.Pod, error) {
	return a.listPods(ctx, map[string]string{appLabel: "pachd"})
}

func (a *apiServer) rcPods(ctx context.Context, pi *pps.PipelineInfo) ([]v1.Pod, error) {
	labels := map[string]string{
		appLabel:             "pipeline",
		pipelineNameLabel:    pi.Pipeline.Name,
		pipelineVersionLabel: fmt.Sprint(pi.Version),
	}
	if projectName := pi.Pipeline.Project.GetName(); projectName != "" {
		labels[pipelineProjectLabel] = projectName
	}
	pp, err := a.listPods(ctx, labels)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	if len(pp) == 0 {
		// Look for the pre-2.4–style pods labelled with the long name.
		// This long name could exceed 63 characters, which is why 2.4
		// and later use separate labels for each component.
		return a.listPods(ctx, map[string]string{
			appLabel: ppsutil.PipelineRcName(pi),
		})
	}
	return pp, nil
}

func pipelineLabels(projectName, pipelineName string, pipelineVersion uint64) map[string]string {
	labels := map[string]string{
		appLabel:             "pipeline",
		pipelineNameLabel:    pipelineName,
		pipelineVersionLabel: fmt.Sprint(pipelineVersion),
		"suite":              suite,
		"component":          "worker",
	}
	if projectName != "" {
		labels[pipelineProjectLabel] = projectName
	}
	return labels
}

func spoutLabels(pipeline *pps.Pipeline) map[string]string {
	m := map[string]string{
		appLabel:          "spout",
		pipelineNameLabel: pipeline.Name,
		"suite":           suite,
		"component":       "worker",
	}
	if projectName := pipeline.Project.GetName(); projectName != "" {
		m[pipelineProjectLabel] = projectName
	}
	return m
}

func (a *apiServer) RenderTemplate(ctx context.Context, req *pps.RenderTemplateRequest) (*pps.RenderTemplateResponse, error) {
	jsonResult, err := pachtmpl.RenderTemplate(req.Template, req.Args)
	if err != nil {
		return nil, err
	}
	var specs []*pps.CreatePipelineRequest
	switch jsonResult[0] {
	case '[':
		if err := json.Unmarshal([]byte(jsonResult), &specs); err != nil {
			return nil, errors.EnsureStack(err)
		}
	case '{':
		var spec pps.CreatePipelineRequest
		if err := protojson.Unmarshal([]byte(jsonResult), &spec); err != nil {
			return nil, errors.EnsureStack(err)
		}
		specs = append(specs, &spec)
	default:
		return nil, errors.Errorf("not a json object or list: %v", jsonResult)
	}
	return &pps.RenderTemplateResponse{
		Json:  jsonResult,
		Specs: specs,
	}, nil
}

func (a *apiServer) ListTask(req *taskapi.ListTaskRequest, server pps.API_ListTaskServer) error {
	return task.List(server.Context(), a.env.TaskService, req, server.Send)
}

// ensurePipelineProject ensures that a pipeline’s repo is valid.  It does nothing
// if pipeline is nil.
func ensurePipelineProject(p *pps.Pipeline) {
	if p == nil {
		return
	}
	if p.Project.GetName() == "" {
		p.Project = &pfs.Project{Name: pfs.DefaultProjectName}
	}
}

// newMessageFilterFunc returns a function that filters out messages that don't satisify either jq filter or projects filter.
func newMessageFilterFunc(jqFilter string, projects []*pfs.Project) (func(context.Context, proto.Message) (bool, error), error) {
	projectsFilter := make(map[string]bool, len(projects))
	for _, project := range projects {
		projectsFilter[project.GetName()] = true
	}
	var (
		jqCode *gojq.Code
		enc    serde.Encoder
	)
	if jqFilter != "" {
		jqQuery, err := gojq.Parse(jqFilter)
		if err != nil {
			return nil, errors.Wrap(err, "error parsing jq filter")
		}
		jqCode, err = gojq.Compile(jqQuery)
		if err != nil {
			return nil, errors.Wrap(err, "error compiling jq filter")
		}
	}
	return func(ctx context.Context, m proto.Message) (bool, error) {
		// Assume that empty projectsFilter means no filter.
		if len(projectsFilter) > 0 {
			switch v := m.(type) {
			case *pps.PipelineInfo:
				if !projectsFilter[v.Pipeline.Project.GetName()] {
					return false, nil
				}
			case *pps.JobInfo:
				if !projectsFilter[v.Job.Pipeline.Project.GetName()] {
					return false, nil
				}
			default:
				return false, errors.Errorf("unknown proto message type: %T\n", v)
			}
		}

		if jqFilter != "" {
			var jsonBuffer bytes.Buffer
			// ensure field names and enum values match with --raw output
			enc = serde.NewJSONEncoder(&jsonBuffer, serde.WithOrigName(true))
			if err := enc.EncodeProto(m); err != nil {
				return false, err
			}
			var v any
			if err := json.Unmarshal(jsonBuffer.Bytes(), &v); err != nil {
				return false, errors.Wrap(err, "error unmarshalling proto message")
			}
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second) // assume that no filter will ever take more than 10 seconds to run
			defer cancel()
			iter := jqCode.RunWithContext(ctx, v)
			v, ok := iter.Next()
			if !ok {
				return false, nil
			}
			if err, ok := v.(error); ok && err != nil {
				return false, errors.Wrap(err, "error applying jq filter on proto message")
			}
			if v == false || v == nil {
				return false, nil
			}
		}
		return true, nil
	}, nil
}

func (a *apiServer) GetClusterDefaults(ctx context.Context, req *pps.GetClusterDefaultsRequest) (*pps.GetClusterDefaultsResponse, error) {
	var clusterDefaults ppsdb.ClusterDefaultsWrapper
	if err := a.clusterDefaults.ReadOnly(ctx).Get("", &clusterDefaults); err != nil {
		if !errors.As(err, &col.ErrNotFound{}) {
			return nil, unknownError(ctx, "could not read cluster defaults", err)
		}
		clusterDefaults.Json = "{}"
	}
	return &pps.GetClusterDefaultsResponse{ClusterDefaultsJson: clusterDefaults.Json}, nil
}

// jsonMergePatch merges a JSON patch in string form with a JSON target, also in
// string form.
func jsonMergePatch(target, patch string) (string, error) {
	var targetObject, patchObject any
	if err := json.Unmarshal([]byte(target), &targetObject); err != nil {
		return "", errors.Wrap(err, "could not unmarshal target JSON")
	}
	if err := json.Unmarshal([]byte(patch), &patchObject); err != nil {
		return "", errors.Wrap(err, "could not unmarshal patch JSON")
	}
	result, err := json.Marshal(mergePatch(targetObject, patchObject))
	if err != nil {
		return "", errors.Wrap(err, "could not marshal merge patch result")
	}
	return string(result), nil
}

// mergePatch implements the RFC 7396 algorithm.  To quote the RFC “If the patch
// is anything other than an object, the result will always be to replace the
// entire target with the entire patch.  Also, it is not possible to patch part
// of a target that is not an object, such as to replace just some of the values
// in an array.”  If the patch _is_ an object, then non-null values replace
// target values, and null values delete target values.
func mergePatch(target, patch any) any {
	switch patch := patch.(type) {
	case map[string]any:
		var targetMap map[string]any
		switch t := target.(type) {
		case map[string]any:
			targetMap = t
		default:
			targetMap = make(map[string]any)
		}
		for name, value := range patch {
			if value == nil {
				delete(targetMap, name)
			} else {
				targetMap[name] = mergePatch(targetMap[name], value)
			}
		}
		return targetMap
	default:
		return patch
	}
}

func badRequest(ctx context.Context, msg string, violations []*errdetails.BadRequest_FieldViolation) error {
	s, err := status.New(codes.InvalidArgument, msg).WithDetails(&errdetails.BadRequest{
		FieldViolations: violations,
	})
	if err != nil {
		log.Error(ctx, "could not add bad-request details", zap.Error(err))
		s = status.New(codes.Internal, "could not add bad-request details")
	}
	return s.Err()
}

func unknownError(ctx context.Context, msg string, err error) error {
	var stack []string
	errors.ForEachStackFrame(err, func(f errors.Frame) {
		stack = append(stack, fmt.Sprintf("%s:%d (%n)", f, f, f))
	})
	s, err := status.Newf(codes.Unknown, "unknown error: %s", msg).WithDetails(&errdetails.DebugInfo{
		StackEntries: stack,
		Detail:       err.Error(),
	})
	if err != nil {
		log.Error(ctx, "could not add debug info details", zap.Error(err))
		s = status.New(codes.Internal, "could not add unknown-error details")
	}
	return s.Err()
}

func (a *apiServer) SetClusterDefaults(ctx context.Context, req *pps.SetClusterDefaultsRequest) (*pps.SetClusterDefaultsResponse, error) {
	var (
		cd pps.ClusterDefaults
	)
	if err := protojson.Unmarshal([]byte(req.GetClusterDefaultsJson()), &cd); err != nil {
		return nil, badRequest(ctx, "invalid cluster defaults JSON", []*errdetails.BadRequest_FieldViolation{
			{Field: "cluster_defaults_json", Description: err.Error()},
		})
	}

	if err := a.txnEnv.WithWriteContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
		if err := a.clusterDefaults.ReadWrite(txnCtx.SqlTx).Put("", &ppsdb.ClusterDefaultsWrapper{Json: req.GetClusterDefaultsJson()}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, unknownError(ctx, "could not write cluster defaults", err)
	}
	// TODO(CORE-1708): add affected pipelines
	return &pps.SetClusterDefaultsResponse{}, nil
}
