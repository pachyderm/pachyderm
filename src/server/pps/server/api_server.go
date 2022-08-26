package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"path"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"github.com/itchyny/gojq"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/types/known/timestamppb"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/client"
	enterpriseclient "github.com/pachyderm/pachyderm/v2/src/enterprise"
	"github.com/pachyderm/pachyderm/v2/src/internal/ancestry"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/lokiutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/metrics"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachtmpl"
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
)

var (
	suite                 = "pachyderm"
	errIncompleteDeletion = errors.Errorf("pipeline may not be fully deleted, provenance may still be intact")
)

// apiServer implements the public interface of the Pachyderm Pipeline System,
// including all RPCs defined in the protobuf spec.
type apiServer struct {
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
	pipelines col.PostgresCollection
	jobs      col.PostgresCollection
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
	errors := false
	kubeClient := a.env.KubeClient
	_, err := kubeClient.CoreV1().Pods(a.namespace).Watch(ctx, metav1.ListOptions{Watch: true})
	if err != nil {
		errors = true
		logrus.Errorf("unable to access kubernetes pods, Pachyderm will continue to work but certain pipeline errors will result in pipelines being stuck indefinitely in \"starting\" state. error: %v", err)
	}
	pods, err := a.pachdPods(ctx)
	if err != nil || len(pods) == 0 {
		errors = true
		logrus.Errorf("unable to access kubernetes pods, Pachyderm will continue to work but 'pachctl logs' will not work. error: %v", err)
	} else {
		// No need to check all pods since we're just checking permissions.
		pod := pods[0]
		_, err = kubeClient.CoreV1().Pods(a.namespace).GetLogs(
			pod.ObjectMeta.Name, &v1.PodLogOptions{
				Container: "pachd",
			}).Timeout(10 * time.Second).Do(ctx).Raw()
		if err != nil {
			errors = true
			logrus.Errorf("unable to access kubernetes logs, Pachyderm will continue to work but 'pachctl logs' will not work. error: %v", err)
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
	if _, err := kubeClient.CoreV1().ReplicationControllers(a.namespace).Create(ctx, rc, metav1.CreateOptions{}); err != nil {
		if err != nil {
			errors = true
			logrus.Errorf("unable to create kubernetes replication controllers, Pachyderm will not function properly until this is fixed. error: %v", err)
		}
	}
	if err := kubeClient.CoreV1().ReplicationControllers(a.namespace).Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
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
	// pipelineOpStartStop is required for StartPipeline and StopPipeline
	pipelineOpStartStop
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
			err := a.env.AuthServer.CheckRepoIsAuthorizedInTransaction(txnCtx, &pfs.Repo{Type: pfs.UserRepoType, Name: repo}, auth.Permission_REPO_READ)
			return errors.EnsureStack(err)
		}); err != nil {
			return err
		}
	}

	// Check that the user is authorized to write to the output repo
	if output != "" {
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
			if _, err := a.env.PFSServer.InspectRepoInTransaction(txnCtx, &pfs.InspectRepoRequest{
				Repo: client.NewRepo(output),
			}); errutil.IsNotFoundError(err) {
				// special case: the pipeline output repo has been deleted (so the
				// pipeline is now invalid). It should be possible to delete the pipeline.
				return nil
			}
			required = auth.Permission_REPO_DELETE
		default:
			return errors.Errorf("internal error, unrecognized operation %v", operation)
		}
		if err := a.env.AuthServer.CheckRepoIsAuthorizedInTransaction(txnCtx, &pfs.Repo{Type: pfs.UserRepoType, Name: output}, required); err != nil {
			return errors.EnsureStack(err)
		}
	}
	return nil
}

func (a *apiServer) UpdateJobState(ctx context.Context, request *pps.UpdateJobStateRequest) (response *types.Empty, retErr error) {
	if err := a.txnEnv.WithTransaction(ctx, func(txn txnenv.Transaction) error {
		return errors.EnsureStack(txn.UpdateJobState(request))
	}, nil); err != nil {
		return nil, err
	}

	return &types.Empty{}, nil
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
				return nil, errors.Errorf("job %s was deleted", request.Job.ID)
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

	cb := func(pipeline string) error {
		jobInfo, err := pachClient.InspectJob(pipeline, request.JobSet.ID, request.Details)
		if err != nil {
			// Not all commits are guaranteed to have an associated job - skip over it
			if errutil.IsNotFoundError(err) {
				return nil
			}
			return err
		}
		return errors.EnsureStack(server.Send(jobInfo))
	}

	if err := forEachCommitInJob(pachClient, request.JobSet.ID, request.Wait, func(ci *pfs.CommitInfo) error {
		if ci.Commit.Branch.Repo.Type != pfs.UserRepoType || ci.Origin.Kind == pfs.OriginKind_ALIAS {
			return nil
		}
		return cb(ci.Commit.Branch.Repo.Name)
	}); err != nil {
		if pfsServer.IsCommitSetNotFoundErr(err) {
			// There are no commits for this ID, but there may still be jobs, query
			// the jobs table directly and don't worry about the topological sort
			// Load all the jobs eagerly to avoid a nested query
			pipelines := []string{}
			jobInfo := &pps.JobInfo{}
			if err := a.jobs.ReadOnly(pachClient.Ctx()).GetByIndex(ppsdb.JobsJobSetIndex, request.JobSet.ID, jobInfo, col.DefaultOptions(), func(string) error {
				pipelines = append(pipelines, jobInfo.Job.Pipeline.Name)
				return nil
			}); err != nil {
				return errors.EnsureStack(err)
			}
			for _, pipeline := range pipelines {
				if err := cb(pipeline); err != nil {
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

	// Return jobsets by the newest job in each set (which can be at a different
	// timestamp due to triggers or deferred processing)
	jobInfo := &pps.JobInfo{}
	err := a.jobs.ReadOnly(serv.Context()).List(jobInfo, col.DefaultOptions(), func(string) error {
		if _, ok := seen[jobInfo.Job.ID]; ok {
			return nil
		}
		seen[jobInfo.Job.ID] = struct{}{}

		jobInfos, err := pachClient.InspectJobSet(jobInfo.Job.ID, request.Details)
		if err != nil {
			return err
		}

		return errors.EnsureStack(serv.Send(&pps.JobSetInfo{
			JobSet: client.NewJobSet(jobInfo.Job.ID),
			Jobs:   jobInfos,
		}))
	})
	return errors.EnsureStack(err)
}

// intersectCommitSets finds all commitsets which involve the specified commits
// (or aliases of the specified commits)
// TODO(global ids): this assumes that all aliases are equivalent to their first
// ancestor non-alias commit, but that may not be true if the ancestor has been
// squashed.  We may need to recursively squash commitsets to prevent this.
func (a *apiServer) intersectCommitSets(ctx context.Context, commits []*pfs.Commit) (map[string]struct{}, error) {
	walkCommits := func(startCommit *pfs.Commit) (map[string]struct{}, error) {
		result := map[string]struct{}{} // key is the commitset id
		queue := []*pfs.Commit{}

		// Walk upwards until finding a concrete commit
		cursor := startCommit
		for {
			commitInfo, err := a.resolveCommit(ctx, cursor)
			if err != nil {
				return nil, err
			}
			if commitInfo.Origin.Kind != pfs.OriginKind_ALIAS || commitInfo.ParentCommit == nil {
				result[cursor.ID] = struct{}{}
				queue = append(queue, commitInfo.ChildCommits...)
				break
			}
			cursor = commitInfo.ParentCommit
		}

		// Now find all descendent aliases
		for len(queue) > 0 {
			cursor = queue[0]
			queue = queue[1:]

			commitInfo, err := a.resolveCommit(ctx, cursor)
			if err != nil {
				return nil, err
			}
			if commitInfo.Origin.Kind == pfs.OriginKind_ALIAS {
				result[cursor.ID] = struct{}{}
				queue = append(queue, commitInfo.ChildCommits...)
			}
		}
		return result, nil
	}

	var intersection map[string]struct{}
	for _, commit := range commits {
		result, err := walkCommits(commit)
		if err != nil {
			return nil, err
		}
		if intersection == nil {
			intersection = result
		} else {
			newIntersection := map[string]struct{}{}
			for commitsetID := range result {
				if _, ok := intersection[commitsetID]; ok {
					newIntersection[commitsetID] = struct{}{}
				}
			}
			intersection = newIntersection
		}
	}
	return intersection, nil
}

// listJob is the internal implementation of ListJob shared between ListJob and
// ListJobStream. When ListJob is removed, this should be inlined into
// ListJobStream.
func (a *apiServer) listJob(
	ctx context.Context,
	pipeline *pps.Pipeline,
	inputCommits []*pfs.Commit,
	history int64,
	details bool,
	jqFilter string,
	f func(*pps.JobInfo) error,
) error {
	if pipeline != nil {
		// If 'pipeline is set, check that caller has access to the pipeline's
		// output repo; currently, that's all that's required for ListJob.
		//
		// If 'pipeline' isn't set, then we don't return an error (otherwise, a
		// caller without access to a single pipeline's output repo couldn't run
		// `pachctl list job` at all) and instead silently skip jobs where the user
		// doesn't have access to the job's output repo.
		if err := a.env.AuthServer.CheckRepoIsAuthorized(ctx, &pfs.Repo{Type: pfs.UserRepoType, Name: pipeline.Name}, auth.Permission_PIPELINE_LIST_JOB); err != nil && !auth.IsErrNotActivated(err) {
			return errors.EnsureStack(err)
		}
	}

	// For each specified input commit, build the set of commitset IDs which
	// belong to all of them.
	commitsets, err := a.intersectCommitSets(ctx, inputCommits)
	if err != nil {
		return err
	}

	var jqCode *gojq.Code
	var enc serde.Encoder
	var jsonBuffer bytes.Buffer
	if jqFilter != "" {
		jqQuery, err := gojq.Parse(jqFilter)
		if err != nil {
			return errors.EnsureStack(err)
		}
		jqCode, err = gojq.Compile(jqQuery)
		if err != nil {
			return errors.EnsureStack(err)
		}
		// ensure field names and enum values match with --raw output
		enc = serde.NewJSONEncoder(&jsonBuffer, serde.WithOrigName(true))
	}

	// pipelineVersions holds the versions of pipelines that we're interested in
	versionKey := func(name string, version uint64) string {
		return fmt.Sprintf("%s-%v", name, version)
	}
	pipelineVersions := make(map[string]bool)
	if err := ppsutil.ListPipelineInfo(ctx, a.pipelines, pipeline, history,
		func(ptr *pps.PipelineInfo) error {
			pipelineVersions[versionKey(ptr.Pipeline.Name, ptr.Version)] = true
			return nil
		}); err != nil {
		return err
	}

	jobs := a.jobs.ReadOnly(ctx)
	jobInfo := &pps.JobInfo{}
	_f := func(string) error {
		if details {
			if err := a.getJobDetails(ctx, jobInfo); err != nil {
				if auth.IsErrNotAuthorized(err) {
					return nil // skip job--see note at top of function
				}
				return err
			}
		}

		if len(inputCommits) > 0 {
			// Only include the job if it's in the set of intersected commitset IDs
			if _, ok := commitsets[jobInfo.Job.ID]; !ok {
				return nil
			}
		}

		if !pipelineVersions[versionKey(jobInfo.Job.Pipeline.Name, jobInfo.PipelineVersion)] {
			return nil
		}

		if jqCode != nil {
			jsonBuffer.Reset()
			// convert jobInfo to a map[string]interface{} for use with gojq
			if err := enc.EncodeProto(jobInfo); err != nil {
				return errors.EnsureStack(err)
			}
			var jobInterface interface{}
			if err := json.Unmarshal(jsonBuffer.Bytes(), &jobInterface); err != nil {
				return errors.EnsureStack(err)
			}
			iter := jqCode.Run(jobInterface)
			// treat either jq false-y value as rejection
			if v, _ := iter.Next(); v == false || v == nil {
				return nil
			}
		}

		return f(jobInfo)
	}
	if pipeline != nil {
		err := jobs.GetByIndex(ppsdb.JobsPipelineIndex, pipeline.Name, jobInfo, col.DefaultOptions(), _f)
		return errors.EnsureStack(err)
	} else {
		err := jobs.List(jobInfo, col.DefaultOptions(), _f)
		return errors.EnsureStack(err)
	}
}

func (a *apiServer) getJobDetails(ctx context.Context, jobInfo *pps.JobInfo) error {
	pipelineName := jobInfo.Job.Pipeline.Name

	if err := a.env.AuthServer.CheckRepoIsAuthorized(ctx, &pfs.Repo{Type: pfs.UserRepoType, Name: pipelineName}, auth.Permission_PIPELINE_LIST_JOB); err != nil && !auth.IsErrNotActivated(err) {
		return errors.EnsureStack(err)
	}

	// Override the SpecCommit for the pipeline to be what it was when this job
	// was created, this prevents races between updating a pipeline and
	// previous jobs running.
	pipelineInfo := &pps.PipelineInfo{}
	if err := a.pipelines.ReadOnly(ctx).GetUniqueByIndex(
		ppsdb.PipelinesVersionIndex,
		ppsdb.VersionKey(pipelineName, jobInfo.PipelineVersion),
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
		workerStatus, err := workerserver.Status(ctx, jobInfo.Job.Pipeline.Name, jobInfo.PipelineVersion, a.env.EtcdClient, a.etcdPrefix, a.workerGrpcPort)
		if err != nil {
			logrus.Errorf("failed to get worker status with err: %s", err.Error())
		} else {
			// It's possible that the workers might be working on datums for other
			// jobs, we omit those since they're not part of the status for this
			// job.
			for _, status := range workerStatus {
				if status.JobID == jobInfo.Job.ID {
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
	return a.listJob(resp.Context(), request.Pipeline, request.InputCommit, request.History, request.Details, request.JqFilter, func(ji *pps.JobInfo) error {
		return errors.EnsureStack(resp.Send(ji))
	})
}

// SubscribeJob implements the protobuf pps.SubscribeJob RPC
func (a *apiServer) SubscribeJob(request *pps.SubscribeJobRequest, stream pps.API_SubscribeJobServer) (retErr error) {
	ctx := stream.Context()

	// Validate arguments
	if request.Pipeline == nil || request.Pipeline.Name == "" {
		return errors.New("pipeline must be specified")
	}

	if err := a.env.AuthServer.CheckRepoIsAuthorized(ctx, &pfs.Repo{Type: pfs.UserRepoType, Name: request.Pipeline.Name}, auth.Permission_PIPELINE_LIST_JOB); err != nil && !auth.IsErrNotActivated(err) {
		return errors.EnsureStack(err)
	}

	// keep track of the jobs that have been sent
	seen := map[string]struct{}{}

	err := a.jobs.ReadOnly(ctx).WatchByIndexF(ppsdb.JobsTerminalIndex, ppsdb.JobTerminalKey(request.Pipeline, false), func(ev *watch.Event) error {
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
func (a *apiServer) DeleteJob(ctx context.Context, request *pps.DeleteJobRequest) (response *types.Empty, retErr error) {
	if request.Job == nil {
		return nil, errors.New("job cannot be nil")
	}
	if err := a.txnEnv.WithWriteContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
		return a.deleteJobInTransaction(txnCtx, request)
	}); err != nil {
		return nil, err
	}
	clearJobCache(a.env.GetPachClient(ctx), ppsdb.JobKey(request.Job))
	return &types.Empty{}, nil
}

func (a *apiServer) deleteJobInTransaction(txnCtx *txncontext.TransactionContext, request *pps.DeleteJobRequest) error {
	if err := a.stopJob(txnCtx, request.Job, "job deleted"); err != nil {
		return err
	}
	return errors.EnsureStack(a.jobs.ReadWrite(txnCtx.SqlTx).Delete(ppsdb.JobKey(request.Job)))
}

// StopJob implements the protobuf pps.StopJob RPC
func (a *apiServer) StopJob(ctx context.Context, request *pps.StopJobRequest) (response *types.Empty, retErr error) {
	if err := a.txnEnv.WithTransaction(ctx, func(txn txnenv.Transaction) error {
		return errors.EnsureStack(txn.StopJob(request))
	}, nil); err != nil {
		return nil, err
	}
	clearJobCache(a.env.GetPachClient(ctx), ppsdb.JobKey(request.Job))
	return &types.Empty{}, nil
}

// TODO: Remove when job state transition operations are handled by a background process.
func clearJobCache(pachClient *client.APIClient, tagPrefix string) {
	if _, err := pachClient.PfsAPIClient.ClearCache(pachClient.Ctx(), &pfs.ClearCacheRequest{
		TagPrefix: tagPrefix,
	}); err != nil {
		logrus.Errorf("errored clearing job cache: %v", err)
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
func (a *apiServer) RestartDatum(ctx context.Context, request *pps.RestartDatumRequest) (response *types.Empty, retErr error) {
	jobInfo, err := a.InspectJob(ctx, &pps.InspectJobRequest{
		Job: request.Job,
	})
	if err != nil {
		return nil, err
	}
	if err := workerserver.Cancel(ctx, jobInfo.Job.Pipeline.Name, jobInfo.PipelineVersion, a.env.EtcdClient, a.etcdPrefix, a.workerGrpcPort, request.Job.ID, request.DataFilters); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

func (a *apiServer) InspectDatum(ctx context.Context, request *pps.InspectDatumRequest) (response *pps.DatumInfo, retErr error) {
	if request.Datum == nil || request.Datum.ID == "" {
		return nil, errors.New("must specify a datum")
	}
	if request.Datum.Job == nil {
		return nil, errors.New("must specify a job")
	}
	// TODO: Auth?
	if err := a.collectDatums(ctx, request.Datum.Job, func(meta *datum.Meta, pfsState *pfs.File) error {
		if common.DatumID(meta.Inputs) == request.Datum.ID {
			response = convertDatumMetaToInfo(meta, request.Datum.Job)
			response.PfsState = pfsState
		}
		return nil
	}); err != nil {
		return nil, err
	}
	if response == nil {
		return nil, errors.Errorf("datum %s not found in job %s", request.Datum.ID, request.Datum.Job)
	}
	return response, nil
}

func (a *apiServer) ListDatum(request *pps.ListDatumRequest, server pps.API_ListDatumServer) (retErr error) {
	// TODO: Auth?
	if request.Input != nil {
		return a.listDatumInput(server.Context(), request.Input, func(meta *datum.Meta) error {
			di := convertDatumMetaToInfo(meta, nil)
			di.State = pps.DatumState_UNKNOWN
			if !request.Filter.Allow(di) {
				return nil
			}
			return errors.EnsureStack(server.Send(di))
		})
	}
	return a.collectDatums(server.Context(), request.Job, func(meta *datum.Meta, _ *pfs.File) error {
		info := convertDatumMetaToInfo(meta, request.Job)
		if !request.Filter.Allow(info) {
			return nil
		}
		return errors.EnsureStack(server.Send(info))
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
	return errors.EnsureStack(di.Iterate(func(meta *datum.Meta) error {
		return cb(meta)
	}))
}

func convertDatumMetaToInfo(meta *datum.Meta, sourceJob *pps.Job) *pps.DatumInfo {
	di := &pps.DatumInfo{
		Datum: &pps.Datum{
			Job: meta.Job,
			ID:  common.DatumID(meta.Inputs),
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
	fsi := datum.NewCommitIterator(pachClient, metaCommit)
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

func (a *apiServer) GetLogs(request *pps.GetLogsRequest, apiGetLogsServer pps.API_GetLogsServer) (retErr error) {
	// Set the default for the `Since` field.
	if request.Since == nil || (request.Since.Seconds == 0 && request.Since.Nanos == 0) {
		request.Since = types.DurationProto(DefaultLogsFrom)
	}
	if a.env.Config.LokiLogging || request.UseLokiBackend {
		return a.getLogsLoki(request, apiGetLogsServer)
	}

	// Authorize request and get list of pods containing logs we're interested in
	// (based on pipeline and job filters)
	var (
		rcName, containerName string
		pods                  []v1.Pod
		err                   error
	)
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
		var pipelineInfo *pps.PipelineInfo
		if request.Pipeline != nil && request.Job == nil {
			pipelineInfo, err = a.inspectPipeline(apiGetLogsServer.Context(), request.Pipeline.Name, true)
			if err != nil {
				return errors.Wrapf(err, "could not get pipeline information for %s", request.Pipeline.Name)
			}
		} else if request.Job != nil {
			// If user provides a job, lookup the pipeline from the JobInfo, and then
			// get the pipeline RC
			jobInfo := &pps.JobInfo{}
			err = a.jobs.ReadOnly(apiGetLogsServer.Context()).Get(ppsdb.JobKey(request.Job), jobInfo)
			if err != nil {
				return errors.Wrapf(err, "could not get job information for \"%s\"", request.Job.ID)
			}
			pipelineInfo, err = a.inspectPipeline(apiGetLogsServer.Context(), jobInfo.Job.Pipeline.Name, true)
			if err != nil {
				return errors.Wrapf(err, "could not get pipeline information for %s", jobInfo.Job.Pipeline.Name)
			}
		}

		// 2) Check whether the caller is authorized to get logs from this pipeline/job
		if err := a.authorizePipelineOp(apiGetLogsServer.Context(), pipelineOpGetLogs, pipelineInfo.Details.Input, pipelineInfo.Pipeline.Name); err != nil {
			return err
		}

		// 3) Get pods for this pipeline
		rcName = ppsutil.PipelineRcName(pipelineInfo.Pipeline.Name, pipelineInfo.Version)
		pods, err = a.rcPods(apiGetLogsServer.Context(), pipelineInfo.Pipeline.Name, pipelineInfo.Version)
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
				stream, err := a.env.KubeClient.CoreV1().Pods(a.namespace).GetLogs(
					pod.ObjectMeta.Name, &v1.PodLogOptions{
						Container:    containerName,
						Follow:       request.Follow,
						TailLines:    tailLines,
						SinceSeconds: &sinceSeconds,
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

func (a *apiServer) getLogsLoki(request *pps.GetLogsRequest, apiGetLogsServer pps.API_GetLogsServer) (retErr error) {
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
	if request.Pipeline == nil && request.Job == nil {
		if len(request.DataFilters) > 0 || request.Datum != nil {
			return errors.Errorf("must specify the Job or Pipeline that the datum is from to get logs for it")
		}
		if err := a.env.AuthServer.CheckClusterIsAuthorized(apiGetLogsServer.Context(), auth.Permission_CLUSTER_GET_PACHD_LOGS); err != nil {
			return errors.EnsureStack(err)
		}
		return lokiutil.QueryRange(apiGetLogsServer.Context(), loki, `{app="pachd"}`, time.Now().Add(-since), time.Time{}, request.Follow, func(t time.Time, line string) error {
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
		pipelineInfo, err = a.inspectPipeline(apiGetLogsServer.Context(), request.Pipeline.Name, true)
		if err != nil {
			return errors.Wrapf(err, "could not get pipeline information for %s", request.Pipeline.Name)
		}
	} else if request.Job != nil {
		// If user provides a job, lookup the pipeline from the JobInfo, and then
		// get the pipeline RC
		jobInfo := &pps.JobInfo{}
		err = a.jobs.ReadOnly(apiGetLogsServer.Context()).Get(ppsdb.JobKey(request.Job), jobInfo)
		if err != nil {
			return errors.Wrapf(err, "could not get job information for \"%s\"", request.Job.ID)
		}
		pipelineInfo, err = a.inspectPipeline(apiGetLogsServer.Context(), jobInfo.Job.Pipeline.Name, true)
		if err != nil {
			return errors.Wrapf(err, "could not get pipeline information for %s", jobInfo.Job.Pipeline.Name)
		}
	}

	// 2) Check whether the caller is authorized to get logs from this pipeline/job
	if err := a.authorizePipelineOp(apiGetLogsServer.Context(), pipelineOpGetLogs, pipelineInfo.Details.Input, pipelineInfo.Pipeline.Name); err != nil {
		return err
	}
	query := fmt.Sprintf(`{pipelineName=%q, container="user"}`, pipelineInfo.Pipeline.Name)
	if request.Master {
		query += contains("master")
	}
	if request.Job != nil {
		query += contains(request.Job.ID)
	}
	if request.Datum != nil {
		query += contains(request.Datum.ID)
	}
	for _, filter := range request.DataFilters {
		query += contains(filter)
	}
	return lokiutil.QueryRange(apiGetLogsServer.Context(), loki, query, time.Now().Add(-since), time.Time{}, request.Follow, func(t time.Time, line string) error {
		msg := new(pps.LogMessage)
		if err := parseLokiLine(line, msg); err != nil {
			logrus.WithField("line", line).WithError(err).Debug("get logs (loki): unparseable log line")
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
		if request.Job != nil && (request.Job.ID != msg.JobID || request.Job.Pipeline.Name != msg.PipelineName) {
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
		return errors.EnsureStack(apiGetLogsServer.Send(msg))
	})
}

// parseNativeLine receives a raw chunk of JSON from Loki, which is our logged
// pps.LogMessage object.
func parseNativeLine(s string, msg *pps.LogMessage) error {
	return errors.EnsureStack(jsonpb.UnmarshalString(s, msg))
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
	if err := parseNativeLine(result.Log, msg); err != nil {
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
	if err := parseNativeLine(l, msg); err != nil {
		return errors.Errorf("native json (%q): %v", l, err)
	}
	return nil
}

func parseLokiLine(inputLine string, msg *pps.LogMessage) error {
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
		{"native", parseNativeLine},
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

func now() *types.Timestamp {
	t, err := types.TimestampProto(time.Now())
	if err != nil {
		panic(err)
	}
	return t
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
	if request.Spout != nil && request.Autoscaling {
		return errors.Errorf("autoscaling can't be used with spouts (spouts aren't triggered externally)")
	}
	return nil
}

func (a *apiServer) validateEnterpriseChecks(ctx context.Context, req *pps.CreatePipelineRequest) error {
	if _, err := a.inspectPipeline(ctx, req.Pipeline.Name, false); err == nil {
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
	if err := a.validateInput(pipelineInfo.Pipeline.Name, pipelineInfo.Details.Input); err != nil {
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
	if pipelineInfo.Details.JobTimeout != nil {
		_, err := types.DurationFromProto(pipelineInfo.Details.JobTimeout)
		if err != nil {
			return errors.EnsureStack(err)
		}
	}
	if pipelineInfo.Details.DatumTimeout != nil {
		_, err := types.DurationFromProto(pipelineInfo.Details.DatumTimeout)
		if err != nil {
			return errors.EnsureStack(err)
		}
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

func branchProvenance(input *pps.Input) []*pfs.Branch {
	var result []*pfs.Branch
	pps.VisitInput(input, func(input *pps.Input) error { //nolint:errcheck
		if input.Pfs != nil {
			result = append(result, client.NewBranch(input.Pfs.Repo, input.Pfs.Branch))
		}
		if input.Cron != nil {
			result = append(result, client.NewBranch(input.Cron.Repo, "master"))
		}
		return nil
	})
	return result
}

func (a *apiServer) fixPipelineInputRepoACLsInTransaction(txnCtx *txncontext.TransactionContext, pipelineInfo *pps.PipelineInfo, prevPipelineInfo *pps.PipelineInfo) (retErr error) {
	addRead := make(map[string]struct{})
	addWrite := make(map[string]struct{})
	remove := make(map[string]struct{})
	var pipelineName string
	// Figure out which repos 'pipeline' might no longer be using
	if prevPipelineInfo != nil {
		pipelineName = prevPipelineInfo.Pipeline.Name
		if err := pps.VisitInput(prevPipelineInfo.Details.Input, func(input *pps.Input) error {
			var repo string
			switch {
			case input.Pfs != nil:
				repo = input.Pfs.Repo
			case input.Cron != nil:
				repo = input.Cron.Repo
			default:
				return nil // no scope to set: input is not a repo
			}
			remove[repo] = struct{}{}
			return nil
		}); err != nil {
			return errors.EnsureStack(err)
		}
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
		if err := pps.VisitInput(pipelineInfo.Details.Input, func(input *pps.Input) error {
			var repo string
			switch {
			case input.Pfs != nil:
				repo = input.Pfs.Repo
			case input.Cron != nil:
				repo = input.Cron.Repo
			default:
				return nil // no scope to set: input is not a repo
			}
			if _, ok := remove[repo]; ok {
				delete(remove, repo)
			} else {
				addRead[repo] = struct{}{}
				if input.Cron != nil {
					addWrite[repo] = struct{}{}
				}
			}
			return nil
		}); err != nil {
			return errors.EnsureStack(err)
		}
	}
	if pipelineName == "" {
		return errors.Errorf("fixPipelineInputRepoACLs called with both current and " +
			"previous pipelineInfos == to nil; this is a bug")
	}

	// make sure we don't touch the pipeline's permissions on its output repo
	delete(remove, pipelineName)
	delete(addRead, pipelineName)
	delete(addWrite, pipelineName)

	defer func() {
		retErr = errors.Wrapf(retErr, "error fixing ACLs on \"%s\"'s input repos", pipelineName)
	}()

	// Remove pipeline from old, unused inputs
	for repo := range remove {
		// If we get an `ErrNoRoleBinding` that means the input repo no longer exists - we're removing it anyways, so we don't care.
		if err := a.env.AuthServer.RemovePipelineReaderFromRepoInTransaction(txnCtx, repo, pipelineName); err != nil && !auth.IsErrNoRoleBinding(err) {
			return errors.EnsureStack(err)
		}
	}
	// Add pipeline to every new input's ACL as a READER
	for repo := range addRead {
		// This raises an error if the input repo doesn't exist, or if the user doesn't have permissions to add a pipeline as a reader on the input repo
		if err := a.env.AuthServer.AddPipelineReaderToRepoInTransaction(txnCtx, repo, pipelineName); err != nil {
			return errors.EnsureStack(err)
		}
	}

	for repo := range addWrite {
		// This raises an error if the input repo doesn't exist, or if the user doesn't have permissions to add a pipeline as a writer on the input repo
		if err := a.env.AuthServer.AddPipelineWriterToSourceRepoInTransaction(txnCtx, repo, pipelineName); err != nil {
			return errors.EnsureStack(err)
		}
	}

	// Add pipeline to its output repo's ACL as a WRITER if it's new
	if prevPipelineInfo == nil {
		if err := a.env.AuthServer.AddPipelineWriterToRepoInTransaction(txnCtx, pipelineName); err != nil {
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
// - CreatePipeline always creates pipeline output branches such that the
//   pipeline's spec branch is in the pipeline output branch's provenance
// - CreatePipeline will always create a new output commit, but that's done
//   by CreateBranch at the bottom of the function, which sets the new output
//   branch provenance, rather than commitPipelineInfoFromFileSet higher up.
// - This is because CreatePipeline calls hardStopPipeline towards the top,
// 	 breaking the provenance connection from the spec branch to the output branch
// - For straightforward pipeline updates (e.g. new pipeline image)
//   stopping + updating + starting the pipeline isn't necessary
// - However it is necessary in many slightly atypical cases  (e.g. the
//   pipeline input changed: if the spec commit is created while the
//   output branch has its old provenance, or the output branch gets new
//   provenance while the old spec commit is the HEAD of the spec branch,
//   then an output commit will be created with provenance that doesn't
//   match its spec's PipelineInfo.Details.Input. Another example is when
//   request.Reprocess == true).
// - Rather than try to enumerate every case where we can't create a spec
//   commit without stopping the pipeline, we just always stop the pipeline
func (a *apiServer) CreatePipeline(ctx context.Context, request *pps.CreatePipelineRequest) (response *types.Empty, retErr error) {
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
	extended.PersistAny(ctx, a.env.EtcdClient, request.Pipeline.Name)

	if err := a.validateEnterpriseChecks(ctx, request); err != nil {
		return nil, err
	}

	if err := a.validateSecret(ctx, request); err != nil {
		return nil, err
	}

	if err := a.txnEnv.WithTransaction(ctx, func(txn txnenv.Transaction) error {
		return errors.EnsureStack(txn.CreatePipeline(request))
	}, nil); err != nil {
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
		Pipeline: request.Pipeline,
		Version:  1,
		Details: &pps.PipelineInfo_Details{
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
			Salt:                  request.Salt,
			Service:               request.Service,
			Spout:                 request.Spout,
			DatumSetSpec:          request.DatumSetSpec,
			DatumTimeout:          request.DatumTimeout,
			JobTimeout:            request.JobTimeout,
			DatumTries:            request.DatumTries,
			SchedulingSpec:        request.SchedulingSpec,
			PodSpec:               request.PodSpec,
			PodPatch:              request.PodPatch,
			S3Out:                 request.S3Out,
			Metadata:              request.Metadata,
			ReprocessSpec:         request.ReprocessSpec,
			Autoscaling:           request.Autoscaling,
		},
	}

	if err := setPipelineDefaults(pipelineInfo); err != nil {
		return nil, err
	}
	// Validate final PipelineInfo (now that defaults have been populated)
	if err := a.validatePipeline(pipelineInfo); err != nil {
		return nil, err
	}

	pps.SortInput(pipelineInfo.Details.Input) // Makes datum hashes comparable

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

func (a *apiServer) CreatePipelineInTransaction(
	txnCtx *txncontext.TransactionContext,
	request *pps.CreatePipelineRequest,
) error {
	pipelineName := request.Pipeline.Name
	oldPipelineInfo, err := a.InspectPipelineInTransaction(txnCtx, pipelineName)
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
	// Verify that all input repos exist (create cron and git repos if necessary)
	if visitErr := pps.VisitInput(newPipelineInfo.Details.Input, func(input *pps.Input) error {
		if input.Pfs != nil {
			if _, err := a.env.PFSServer.InspectRepoInTransaction(txnCtx,
				&pfs.InspectRepoRequest{
					Repo: client.NewSystemRepo(input.Pfs.Repo, input.Pfs.RepoType),
				},
			); err != nil {
				return errors.EnsureStack(err)
			}
		}
		if input.Cron != nil {
			if err := a.env.PFSServer.CreateRepoInTransaction(txnCtx,
				&pfs.CreateRepoRequest{
					Repo:        client.NewRepo(input.Cron.Repo),
					Description: fmt.Sprintf("Cron tick repo for pipeline %s.", request.Pipeline.Name),
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
	if err := a.authorizePipelineOpInTransaction(txnCtx, operation, newPipelineInfo.Details.Input, newPipelineInfo.Pipeline.Name); err != nil {
		return err
	}

	var (
		// provenance for the pipeline's output branch (includes the spec branch)
		provenance = append(branchProvenance(newPipelineInfo.Details.Input),
			client.NewSystemRepo(pipelineName, pfs.SpecRepoType).NewBranch("master"))
		outputBranch = client.NewBranch(pipelineName, newPipelineInfo.Details.OutputBranch)
		metaBranch   = client.NewSystemRepo(pipelineName, pfs.MetaRepoType).NewBranch(newPipelineInfo.Details.OutputBranch)
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
		if err := a.env.PFSServer.CreateRepoInTransaction(txnCtx,
			&pfs.CreateRepoRequest{
				Repo:        client.NewRepo(pipelineName),
				Description: fmt.Sprintf("Output repo for pipeline %s.", request.Pipeline.Name),
			}); err != nil && !errutil.IsAlreadyExistError(err) {
			return errors.Wrapf(err, "error creating output repo for %s", pipelineName)
		} else if errutil.IsAlreadyExistError(err) {
			return errors.Errorf("pipeline %q cannot be created because a repo with the same name already exists", pipelineName)
		}
		if err := a.env.PFSServer.CreateRepoInTransaction(txnCtx,
			&pfs.CreateRepoRequest{
				Repo:        client.NewSystemRepo(pipelineName, pfs.SpecRepoType),
				Description: fmt.Sprintf("Spec repo for pipeline %s.", request.Pipeline.Name),
				Update:      true,
			}); err != nil && !errutil.IsAlreadyExistError(err) {
			return errors.Wrapf(err, "error creating spec repo for %s", pipelineName)
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
		newPipelineInfo.SpecCommit, err = a.env.PFSServer.StartCommitInTransaction(txnCtx, &pfs.StartCommitRequest{
			Branch: client.NewSystemRepo(pipelineName, pfs.SpecRepoType).NewBranch("master"),
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
		token, err := a.env.AuthServer.GetPipelineAuthTokenInTransaction(txnCtx, request.Pipeline.Name)
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
		if err := a.stopAllJobsInPipeline(txnCtx, request.Pipeline); err != nil {
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
	if err := a.env.PFSServer.CreateBranchInTransaction(txnCtx, &pfs.CreateBranchRequest{
		Branch:     outputBranch,
		Provenance: provenance,
	}); err != nil {
		return errors.Wrapf(err, "could not create/update output branch")
	}

	if visitErr := pps.VisitInput(request.Input, func(input *pps.Input) error {
		if input.Pfs != nil && input.Pfs.Trigger != nil {
			var prevHead *pfs.Commit
			if branchInfo, err := a.env.PFSServer.InspectBranchInTransaction(txnCtx, &pfs.InspectBranchRequest{
				Branch: client.NewBranch(input.Pfs.Repo, input.Pfs.Branch),
			}); err != nil {
				if !errutil.IsNotFoundError(err) {
					return errors.EnsureStack(err)
				}
			} else {
				prevHead = branchInfo.Head
			}

			err := a.env.PFSServer.CreateBranchInTransaction(txnCtx, &pfs.CreateBranchRequest{
				Branch:  client.NewBranch(input.Pfs.Repo, input.Pfs.Branch),
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
		if err := a.env.PFSServer.CreateRepoInTransaction(txnCtx, &pfs.CreateRepoRequest{
			Repo:        metaBranch.Repo,
			Description: fmt.Sprint("Meta repo for pipeline ", pipelineName),
		}); err != nil && !errutil.IsAlreadyExistError(err) {
			return errors.Wrap(err, "could not create meta repo")
		}
		if err := a.env.PFSServer.CreateBranchInTransaction(txnCtx, &pfs.CreateBranchRequest{
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
	now := time.Now()
	nCreatedBranches := make(map[string]int)
	if err := pps.VisitInput(input, func(input *pps.Input) error {
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
				start, _ := types.TimestampProto(now)
				input.Cron.Start = start
			}
			if input.Cron.Repo == "" {
				input.Cron.Repo = fmt.Sprintf("%s_%s", pipelineName, input.Cron.Name)
			}
		}
		return nil
	}); err != nil {
		logrus.Errorf("error while visiting inputs: %v", err)
	}
}

func (a *apiServer) stopAllJobsInPipeline(txnCtx *txncontext.TransactionContext, pipeline *pps.Pipeline) error {
	// Using ReadWrite here may load a large number of jobs inline in the
	// transaction, but doing an inconsistent read outside of the transaction
	// would be pretty sketchy (and we'd have to worry about trying to get another
	// postgres connection and possibly deadlocking).
	jobInfo := &pps.JobInfo{}
	sort := &col.Options{Target: col.SortByCreateRevision, Order: col.SortAscend}
	err := a.jobs.ReadWrite(txnCtx.SqlTx).GetByIndex(ppsdb.JobsTerminalIndex, ppsdb.JobTerminalKey(pipeline, false), jobInfo, sort, func(string) error {
		return a.stopJob(txnCtx, jobInfo.Job, "pipeline updated")
	})
	return errors.EnsureStack(err)
}

func (a *apiServer) updatePipeline(
	txnCtx *txncontext.TransactionContext,
	pipeline string,
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
	return a.inspectPipeline(ctx, request.Pipeline.Name, request.Details)
}

// inspectPipeline contains the functional implementation of InspectPipeline.
// Many functions (GetLogs, ListPipeline) need to inspect a pipeline, so they
// call this instead of making an RPC
func (a *apiServer) inspectPipeline(ctx context.Context, name string, details bool) (*pps.PipelineInfo, error) {
	var info *pps.PipelineInfo
	if err := a.txnEnv.WithReadContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
		var err error
		info, err = a.InspectPipelineInTransaction(txnCtx, name)
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
			rcName := ppsutil.PipelineRcName(info.Pipeline.Name, info.Version)
			service, err := kubeClient.CoreV1().Services(a.namespace).Get(ctx, fmt.Sprintf("%s-user", rcName), metav1.GetOptions{})
			if err != nil {
				if !errutil.IsNotFoundError(err) {
					return nil, errors.EnsureStack(err)
				}
			} else {
				info.Details.Service.IP = service.Spec.ClusterIP
			}
		}

		workerStatus, err := workerserver.Status(ctx, info.Pipeline.Name, info.Version, a.env.EtcdClient, a.etcdPrefix, a.workerGrpcPort)
		if err != nil {
			logrus.Errorf("failed to get worker status with err: %s", err.Error())
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

func (a *apiServer) InspectPipelineInTransaction(txnCtx *txncontext.TransactionContext, name string) (*pps.PipelineInfo, error) {
	name, ancestors, err := ancestry.Parse(name)
	if err != nil {
		return nil, err
	}

	if ancestors < 0 {
		return nil, errors.New("cannot inspect future pipelines")
	}

	key, err := ppsutil.FindPipelineSpecCommitInTransaction(txnCtx, a.env.PFSServer, name, "")
	if err != nil {
		return nil, errors.Wrapf(err, "pipeline not found: couldn't find up to date spec for pipeline %q", name)
	}

	pipelineInfo := &pps.PipelineInfo{}
	if err := a.pipelines.ReadWrite(txnCtx.SqlTx).Get(key, pipelineInfo); err != nil {
		if col.IsErrNotFound(err) {
			return nil, errors.Errorf("pipeline \"%s\" not found", name)
		}
		return nil, errors.EnsureStack(err)
	}
	if ancestors > 0 {
		targetVersion := int(pipelineInfo.Version) - ancestors
		if targetVersion < 1 {
			return nil, errors.Errorf("pipeline %q has only %d versions, not enough to find ancestor %d", name, pipelineInfo.Version, ancestors)
		}
		if err := a.pipelines.ReadWrite(txnCtx.SqlTx).GetUniqueByIndex(ppsdb.PipelinesVersionIndex, ppsdb.VersionKey(name, uint64(targetVersion)), pipelineInfo); err != nil {
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
	return a.listPipeline(srv.Context(), request, srv.Send)
}

func (a *apiServer) getLatestJobState(ctx context.Context, info *pps.PipelineInfo) error {
	// fill in state of most-recently-created job (the first shown in list job)
	opts := col.DefaultOptions()
	opts.Limit = 1
	var job pps.JobInfo
	err := a.jobs.ReadOnly(ctx).GetByIndex(
		ppsdb.JobsPipelineIndex,
		info.Pipeline.Name,
		&job,
		opts, func(_ string) error {
			info.LastJobState = job.State
			return errutil.ErrBreak // not strictly necessary because we are limiting to 1 already
		})
	return errors.EnsureStack(err)
}

func (a *apiServer) listPipeline(ctx context.Context, request *pps.ListPipelineRequest, f func(*pps.PipelineInfo) error) error {
	var jqCode *gojq.Code
	var enc serde.Encoder
	var jsonBuffer bytes.Buffer
	if request.JqFilter != "" {
		jqQuery, err := gojq.Parse(request.JqFilter)
		if err != nil {
			return errors.EnsureStack(err)
		}
		jqCode, err = gojq.Compile(jqQuery)
		if err != nil {
			return errors.EnsureStack(err)
		}
		// ensure field names and enum values match with --raw output
		enc = serde.NewJSONEncoder(&jsonBuffer, serde.WithOrigName(true))
	}
	// get all pipelines at once to avoid holding the list query open
	// this should be fine with numbers of pipelines pachyderm can actually run
	var infos []*pps.PipelineInfo
	if err := ppsutil.ListPipelineInfo(ctx, a.pipelines, request.Pipeline, request.History, func(ptr *pps.PipelineInfo) error {
		infos = append(infos, proto.Clone(ptr).(*pps.PipelineInfo))
		return nil
	}); err != nil {
		return err
	}

	filterPipeline := func(pipelineInfo *pps.PipelineInfo) bool {
		if jqCode != nil {
			jsonBuffer.Reset()
			// convert pipelineInfo to a map[string]interface{} for use with gojq
			if err := enc.EncodeProto(pipelineInfo); err != nil {
				logrus.Errorf("error encoding pipelineInfo to JSON: %v", err)
			}
			var pipelineInterface interface{}
			if err := json.Unmarshal(jsonBuffer.Bytes(), &pipelineInterface); err != nil {
				logrus.Errorf("error parsing JSON encoded pipeline info: %v", err)
			}
			iter := jqCode.Run(pipelineInterface)
			// treat either jq false-y value as rejection
			if v, _ := iter.Next(); v == false || v == nil {
				return false
			}
		}
		return true
	}

	for i := range infos {
		if filterPipeline(infos[i]) {
			if err := a.getLatestJobState(ctx, infos[i]); err != nil {
				return err
			}
			if err := f(infos[i]); err != nil {
				return err
			}
		}
	}
	return nil
}

// DeletePipeline implements the protobuf pps.DeletePipeline RPC
func (a *apiServer) DeletePipeline(ctx context.Context, request *pps.DeletePipelineRequest) (response *types.Empty, retErr error) {
	if request.All {
		request.Pipeline = &pps.Pipeline{}
		pipelineInfo := &pps.PipelineInfo{}
		deleted := make(map[string]struct{})
		if err := a.pipelines.ReadOnly(ctx).List(pipelineInfo, col.DefaultOptions(), func(string) error {
			if _, ok := deleted[pipelineInfo.Pipeline.Name]; ok {
				// while the delete pipeline call will delete historical versions,
				// they could still show up in the list. Ignore them
				return nil
			}
			request.Pipeline.Name = pipelineInfo.Pipeline.Name
			err := a.deletePipeline(ctx, request)
			if err == nil {
				deleted[pipelineInfo.Pipeline.Name] = struct{}{}
			}
			return err
		}); err != nil {
			return nil, errors.EnsureStack(err)
		}
	} else if err := a.deletePipeline(ctx, request); err != nil {
		return nil, err
	}

	return &types.Empty{}, nil
}

func (a *apiServer) deletePipeline(ctx context.Context, request *pps.DeletePipelineRequest) error {
	pipelineName := request.Pipeline.Name

	// stop the pipeline to avoid interference from new jobs
	if _, err := a.StopPipeline(ctx,
		&pps.StopPipelineRequest{Pipeline: request.Pipeline}); err != nil && errutil.IsNotFoundError(err) {
		logrus.Errorf("failed to stop pipeline, continuing with delete: %v", err)
	} else if err != nil {
		return errors.Wrapf(err, "error stopping pipeline %s", pipelineName)
	}
	// perform the rest of the deletion in a transaction
	var deleteErr error
	if err := a.txnEnv.WithWriteContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
		deleteErr = a.deletePipelineInTransaction(txnCtx, request)
		// we still want deletion to succeed if it was merely incomplete, but warn the caller
		if errors.Is(deleteErr, errIncompleteDeletion) {
			return nil
		}
		return deleteErr
	}); err != nil {
		return err
	}
	clearJobCache(a.env.GetPachClient(ctx), pipelineName)
	return deleteErr
}

func (a *apiServer) deletePipelineInTransaction(txnCtx *txncontext.TransactionContext, request *pps.DeletePipelineRequest) error {
	pipelineName := request.Pipeline.Name

	// make sure the pipeline exists
	var foundPipeline bool
	if err := a.pipelines.ReadWrite(txnCtx.SqlTx).GetByIndex(
		ppsdb.PipelinesNameIndex,
		pipelineName,
		&pps.PipelineInfo{},
		col.DefaultOptions(),
		func(_ string) error {
			foundPipeline = true
			return nil
		}); err != nil {
		return errors.Wrapf(err, "error checking if pipeline %s exists", pipelineName)
	}
	if !foundPipeline {
		// nothing to delete
		return nil
	}

	pipelineInfo := &pps.PipelineInfo{}
	// Try to retrieve PipelineInfo for this pipeline. If we see a not found error,
	// we will still try to delete what we can because we know there is a pipeline
	if specCommit, err := ppsutil.FindPipelineSpecCommitInTransaction(txnCtx, a.env.PFSServer, pipelineName, ""); err == nil {
		if err := a.pipelines.ReadWrite(txnCtx.SqlTx).Get(specCommit, pipelineInfo); err != nil && !col.IsErrNotFound(err) {
			return errors.EnsureStack(err)
		}
	} else if !errutil.IsNotFoundError(err) && !auth.IsErrNoRoleBinding(err) {
		return err
	}
	var missingRepo bool
	// check if the output repo exists--if not, the pipeline is non-functional and
	// the rest of the delete operation continues without any auth checks
	if _, err := a.env.PFSServer.InspectRepoInTransaction(txnCtx, &pfs.InspectRepoRequest{
		Repo: client.NewRepo(pipelineName)}); err != nil && !errutil.IsNotFoundError(err) && !auth.IsErrNoRoleBinding(err) {
		return errors.EnsureStack(err)
	} else if err == nil {
		// Check if the caller is authorized to delete this pipeline
		if err := a.authorizePipelineOpInTransaction(txnCtx, pipelineOpDelete, pipelineInfo.GetDetails().GetInput(), pipelineName); err != nil {
			return err
		}
	} else {
		missingRepo = true
	}

	// If necessary, revoke the pipeline's auth token and remove it from its inputs' ACLs
	// If auth is deactivated, don't bother doing either
	if _, err := txnCtx.WhoAmI(); err == nil && pipelineInfo.AuthToken != "" {
		// 'pipelineInfo' == nil => remove pipeline from all input repos
		if pipelineInfo.Details != nil {
			// need details for acls
			if err := a.fixPipelineInputRepoACLsInTransaction(txnCtx, nil, pipelineInfo); err != nil {
				return errors.Wrapf(err, "error fixing repo ACLs for pipeline %q", pipelineName)
			}
		}
		if _, err := a.env.AuthServer.RevokeAuthTokenInTransaction(txnCtx,
			&auth.RevokeAuthTokenRequest{
				Token: pipelineInfo.AuthToken,
			}); err != nil {
			return errors.EnsureStack(err)
		}
	}

	// Delete all of the pipeline's jobs - we shouldn't need to worry about any
	// new jobs since the pipeline has already been stopped.
	jobInfo := &pps.JobInfo{}
	if err := a.jobs.ReadWrite(txnCtx.SqlTx).GetByIndex(ppsdb.JobsPipelineIndex, pipelineName, jobInfo, col.DefaultOptions(), func(string) error {
		job := proto.Clone(jobInfo.Job).(*pps.Job)
		err := a.deleteJobInTransaction(txnCtx, &pps.DeleteJobRequest{Job: job})
		if errutil.IsNotFoundError(err) || auth.IsErrNoRoleBinding(err) {
			return nil
		}
		return err
	}); err != nil {
		return errors.EnsureStack(err)
	}

	// Delete all past PipelineInfos
	if err := a.pipelines.ReadWrite(txnCtx.SqlTx).DeleteByIndex(ppsdb.PipelinesNameIndex, pipelineName); err != nil {
		return errors.Wrapf(err, "collection.Delete")
	}

	if !missingRepo {
		if !request.KeepRepo {
			// delete the pipeline's output repo
			if err := a.env.PFSServer.DeleteRepoInTransaction(txnCtx, &pfs.DeleteRepoRequest{
				Repo:  client.NewRepo(pipelineName),
				Force: request.Force,
			}); err != nil && !errutil.IsNotFoundError(err) {
				return errors.Wrap(err, "error deleting pipeline repo")
			}
		} else {
			// Remove branch provenance from output and then delete meta and spec repos
			// this leaves the repo as a source repo, eliminating pipeline metadata
			// need details for output branch, presumably if we don't have them the spec repo is gone, anyway
			if pipelineInfo.Details != nil {
				if err := a.env.PFSServer.CreateBranchInTransaction(txnCtx, &pfs.CreateBranchRequest{
					Branch: client.NewBranch(pipelineName, pipelineInfo.Details.OutputBranch),
				}); err != nil {
					return errors.EnsureStack(err)
				}
			}
			if err := a.env.PFSServer.DeleteRepoInTransaction(txnCtx, &pfs.DeleteRepoRequest{
				Repo:  client.NewSystemRepo(pipelineName, pfs.SpecRepoType),
				Force: request.Force,
			}); err != nil && !col.IsErrNotFound(err) && !auth.IsErrNoRoleBinding(err) {
				return errors.EnsureStack(err)
			}
			if err := a.env.PFSServer.DeleteRepoInTransaction(txnCtx, &pfs.DeleteRepoRequest{
				Repo:  client.NewSystemRepo(pipelineName, pfs.MetaRepoType),
				Force: request.Force,
			}); err != nil && !col.IsErrNotFound(err) && !auth.IsErrNoRoleBinding(err) {
				return errors.EnsureStack(err)
			}
		}
	}
	// delete cron after main repo is deleted or has provenance removed
	// cron repos are only used to trigger jobs, so don't keep them even with KeepRepo
	if pipelineInfo.Details != nil {
		if err := pps.VisitInput(pipelineInfo.Details.Input, func(input *pps.Input) error {
			if input.Cron != nil {
				err := a.env.PFSServer.DeleteRepoInTransaction(txnCtx, &pfs.DeleteRepoRequest{
					Repo:  client.NewRepo(input.Cron.Repo),
					Force: request.Force,
				})
				return errors.EnsureStack(err)
			}
			return nil
		}); err != nil {
			return err
		}
	}

	if request.KeepRepo && !missingRepo && pipelineInfo.Details == nil {
		// warn about being unable to delete provenance, caller can ignore
		return errIncompleteDeletion
	}

	return nil
}

// StartPipeline implements the protobuf pps.StartPipeline RPC
func (a *apiServer) StartPipeline(ctx context.Context, request *pps.StartPipelineRequest) (response *types.Empty, retErr error) {
	if request.Pipeline == nil {
		return nil, errors.New("request.Pipeline cannot be nil")
	}

	if err := a.txnEnv.WithWriteContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
		pipelineInfo, err := a.InspectPipelineInTransaction(txnCtx, request.Pipeline.Name)
		if err != nil {
			return err
		}

		// check if the caller is authorized to update this pipeline
		if err := a.authorizePipelineOpInTransaction(txnCtx, pipelineOpStartStop, pipelineInfo.Details.Input, pipelineInfo.Pipeline.Name); err != nil {
			return err
		}

		// Restore branch provenance, which may create a new output commit/job
		provenance := append(branchProvenance(pipelineInfo.Details.Input),
			client.NewSystemRepo(pipelineInfo.Pipeline.Name, pfs.SpecRepoType).NewBranch("master"))
		if err := a.env.PFSServer.CreateBranchInTransaction(txnCtx, &pfs.CreateBranchRequest{
			Branch:     client.NewBranch(pipelineInfo.Pipeline.Name, pipelineInfo.Details.OutputBranch),
			Provenance: provenance,
		}); err != nil {
			return errors.EnsureStack(err)
		}
		// restore same provenance to meta repo
		if err := a.env.PFSServer.CreateBranchInTransaction(txnCtx, &pfs.CreateBranchRequest{
			Branch:     client.NewSystemRepo(pipelineInfo.Pipeline.Name, pfs.MetaRepoType).NewBranch(pipelineInfo.Details.OutputBranch),
			Provenance: provenance,
		}); err != nil {
			return errors.EnsureStack(err)
		}

		newPipelineInfo := &pps.PipelineInfo{}
		return a.updatePipeline(txnCtx, pipelineInfo.Pipeline.Name, newPipelineInfo, func() error {
			newPipelineInfo.Stopped = false
			return nil
		})
	}); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

// StopPipeline implements the protobuf pps.StopPipeline RPC
func (a *apiServer) StopPipeline(ctx context.Context, request *pps.StopPipelineRequest) (response *types.Empty, retErr error) {
	if request.Pipeline == nil {
		return nil, errors.New("request.Pipeline cannot be nil")
	}

	if err := a.txnEnv.WithWriteContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
		pipelineInfo, err := a.InspectPipelineInTransaction(txnCtx, request.Pipeline.Name)
		if err == nil {
			// check if the caller is authorized to update this pipeline
			// don't pass in the input - stopping the pipeline means they won't be read anymore,
			// so we don't need to check any permissions
			if err := a.authorizePipelineOpInTransaction(txnCtx, pipelineOpStartStop, pipelineInfo.Details.Input, pipelineInfo.Pipeline.Name); err != nil {
				return err
			}

			// Remove branch provenance to prevent new output and meta commits from being created
			if err := a.env.PFSServer.CreateBranchInTransaction(txnCtx, &pfs.CreateBranchRequest{
				Branch:     client.NewBranch(pipelineInfo.Pipeline.Name, pipelineInfo.Details.OutputBranch),
				Provenance: nil,
			}); err != nil {
				return errors.EnsureStack(err)
			}
			if pipelineInfo.Details.Spout == nil && pipelineInfo.Details.Service == nil {
				if err := a.env.PFSServer.CreateBranchInTransaction(txnCtx, &pfs.CreateBranchRequest{
					Branch:     client.NewSystemRepo(pipelineInfo.Pipeline.Name, pfs.MetaRepoType).NewBranch(pipelineInfo.Details.OutputBranch),
					Provenance: nil,
				}); err != nil {
					return errors.EnsureStack(err)
				}
			}

			newPipelineInfo := &pps.PipelineInfo{}
			if err := a.updatePipeline(txnCtx, pipelineInfo.Pipeline.Name, newPipelineInfo, func() error {
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
		return a.stopAllJobsInPipeline(txnCtx, request.Pipeline)
	}); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

func (a *apiServer) RunPipeline(ctx context.Context, request *pps.RunPipelineRequest) (response *types.Empty, retErr error) {
	return nil, errors.New("unimplemented")
}

func (a *apiServer) RunCron(ctx context.Context, request *pps.RunCronRequest) (response *types.Empty, retErr error) {
	pipelineInfo, err := a.inspectPipeline(ctx, request.Pipeline.Name, true)
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

	return &types.Empty{}, nil
}

func (a *apiServer) propagateJobs(txnCtx *txncontext.TransactionContext) error {
	commitInfos, err := a.env.PFSServer.InspectCommitSetInTransaction(txnCtx, client.NewCommitSet(txnCtx.CommitSetID))
	if err != nil {
		return errors.EnsureStack(err)
	}

	for _, commitInfo := range commitInfos {
		// Skip alias commits and any commits which have already been finished
		if commitInfo.Origin.Kind == pfs.OriginKind_ALIAS || commitInfo.Finishing != nil {
			continue
		}

		// Skip commits from system repos
		if commitInfo.Commit.Branch.Repo.Type != pfs.UserRepoType {
			continue
		}

		// Skip commits from repos that have no associated pipeline
		var pipelineInfo *pps.PipelineInfo
		if pipelineInfo, err = a.InspectPipelineInTransaction(txnCtx, commitInfo.Commit.Branch.Repo.Name); err != nil {
			if col.IsErrNotFound(err) {
				continue
			}
			return err
		}

		// Don't create jobs for spouts
		if pipelineInfo.Type == pps.PipelineInfo_PIPELINE_TYPE_SPOUT {
			continue
		}

		// Check if there is an existing job for the output commit
		job := client.NewJob(pipelineInfo.Pipeline.Name, txnCtx.CommitSetID)
		jobInfo := &pps.JobInfo{}
		if err := a.jobs.ReadWrite(txnCtx.SqlTx).Get(ppsdb.JobKey(job), jobInfo); err == nil {
			continue // Job already exists, skip it
		} else if !col.IsErrNotFound(err) {
			return errors.EnsureStack(err)
		}

		pipelines := a.pipelines.ReadWrite(txnCtx.SqlTx)
		jobs := a.jobs.ReadWrite(txnCtx.SqlTx)
		jobPtr := &pps.JobInfo{
			Job:             job,
			PipelineVersion: pipelineInfo.Version,
			OutputCommit:    commitInfo.Commit,
			Stats:           &pps.ProcessStats{},
			Created:         types.TimestampNow(),
		}
		if err := ppsutil.UpdateJobState(pipelines, jobs, jobPtr, pps.JobState_JOB_CREATED, ""); err != nil {
			return err
		}
	}
	return nil
}

// CreateSecret implements the protobuf pps.CreateSecret RPC
func (a *apiServer) CreateSecret(ctx context.Context, request *pps.CreateSecretRequest) (response *types.Empty, retErr error) {
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
	return &types.Empty{}, nil
}

// DeleteSecret implements the protobuf pps.DeleteSecret RPC
func (a *apiServer) DeleteSecret(ctx context.Context, request *pps.DeleteSecretRequest) (response *types.Empty, retErr error) {
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "DeleteSecret")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	if err := a.env.KubeClient.CoreV1().Secrets(a.namespace).Delete(ctx, request.Secret.Name, metav1.DeleteOptions{}); err != nil {
		return nil, errors.Wrapf(err, "failed to delete secret")
	}
	return &types.Empty{}, nil
}

// InspectSecret implements the protobuf pps.InspectSecret RPC
func (a *apiServer) InspectSecret(ctx context.Context, request *pps.InspectSecretRequest) (response *pps.SecretInfo, retErr error) {
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "InspectSecret")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	secret, err := a.env.KubeClient.CoreV1().Secrets(a.namespace).Get(ctx, request.Secret.Name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get secret")
	}
	creationTimestamp := timestamppb.New(secret.GetCreationTimestamp().Time)
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
		creationTimestamp := timestamppb.New(s.GetCreationTimestamp().Time)

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
	if _, err := a.DeletePipeline(ctx, &pps.DeletePipelineRequest{All: true, Force: true}); err != nil {
		return nil, err
	}

	if err := a.env.KubeClient.CoreV1().Secrets(a.namespace).DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{
		LabelSelector: "secret-source=pachyderm-user",
	}); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return &types.Empty{}, nil
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
		token, err := a.env.AuthServer.GetPipelineAuthTokenInTransaction(txnCtx, pi.Pipeline.Name)
		if err != nil {
			return errors.Wrapf(grpcutil.ScrubGRPC(err), "could not generate pipeline auth token")
		}
		if err := a.updatePipeline(txnCtx, pi.Pipeline.Name, pi, func() error {
			pi.AuthToken = token
			return nil
		}); err != nil {
			return errors.Wrapf(err, "could not update %q with new auth token", pi.Pipeline.Name)
		}
		// put 'pipeline' on relevant ACLs
		if err := a.fixPipelineInputRepoACLsInTransaction(txnCtx, pi, nil); err != nil {
			return errors.Wrapf(err, "fix repo ACLs for pipeline %q", pi.Pipeline.Name)
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
func (a *apiServer) RunLoadTestDefault(ctx context.Context, _ *types.Empty) (_ *pps.RunLoadTestResponse, retErr error) {
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

func (a *apiServer) rcPods(ctx context.Context, pipelineName string, pipelineVersion uint64) ([]v1.Pod, error) {
	pp, err := a.listPods(ctx, map[string]string{
		appLabel:             "pipeline",
		pipelineNameLabel:    pipelineName,
		pipelineVersionLabel: fmt.Sprint(pipelineVersion),
	})
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	if len(pp) == 0 {
		// Look for the pre-2.4–style pods labelled with the long name.
		// This long name could exceed 63 characters, which is why 2.4
		// and later use separate labels for each component.
		return a.listPods(ctx, map[string]string{
			appLabel: ppsutil.PipelineRcName(pipelineName, pipelineVersion),
		})
	}
	return pp, nil
}

func (a *apiServer) resolveCommit(ctx context.Context, commit *pfs.Commit) (*pfs.CommitInfo, error) {
	pachClient := a.env.GetPachClient(ctx)
	ci, err := pachClient.PfsAPIClient.InspectCommit(
		pachClient.Ctx(),
		&pfs.InspectCommitRequest{
			Commit: commit,
			Wait:   pfs.CommitState_STARTED,
		})
	if err != nil {
		return nil, grpcutil.ScrubGRPC(err)
	}
	return ci, nil
}

func pipelineLabels(pipelineName string, pipelineVersion uint64) map[string]string {
	return map[string]string{
		appLabel:             "pipeline",
		pipelineNameLabel:    pipelineName,
		pipelineVersionLabel: fmt.Sprint(pipelineVersion),
		"suite":              suite,
		"component":          "worker",
	}
}

func spoutLabels(pipelineName string) map[string]string {
	return map[string]string{
		appLabel:          "spout",
		pipelineNameLabel: pipelineName,
		"suite":           suite,
		"component":       "worker",
	}
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
		if err := jsonpb.Unmarshal(strings.NewReader(jsonResult), &spec); err != nil {
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
