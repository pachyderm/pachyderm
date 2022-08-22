// Package ppsutil contains utilities for various PPS-related tasks, which are
// shared by both the PPS API and the worker binary. These utilities include:
// - Getting the RC name and querying k8s reguarding pipelines
// - Reading and writing pipeline resource requests and limits
// - Reading and writing PipelineInfos[1]
//
// [1] Note that PipelineInfo in particular is complicated because it contains
// fields that are not always set or are stored in multiple places.  The
// 'Details' field is not stored in the database and must be fetched from the
// PFS spec commit, and a few fields in 'Details' may depend on current
// kubernetes state.
package ppsutil

import (
	"bufio"
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/collection"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/tracing"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	pfsServer "github.com/pachyderm/pachyderm/v2/src/server/pfs"
	ppsServer "github.com/pachyderm/pachyderm/v2/src/server/pps"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/common"
)

// PipelineRepo creates a pfs repo for a given pipeline.
func PipelineRepo(pipeline *pps.Pipeline) *pfs.Repo {
	return client.NewRepo(pipeline.Name)
}

// PipelineRcName generates the name of the k8s replication controller that
// manages a pipeline's workers
func PipelineRcName(name string, version uint64) string {
	// k8s won't allow RC names that contain upper-case letters
	// or underscores
	// TODO: deal with name collision
	name = strings.ReplaceAll(name, "_", "-")
	return fmt.Sprintf("pipeline-%s-v%d", strings.ToLower(name), version)
}

// GetRequestsResourceListFromPipeline returns a list of resources that the pipeline,
// minimally requires.
func GetRequestsResourceListFromPipeline(pipelineInfo *pps.PipelineInfo) (*v1.ResourceList, error) {
	return getResourceListFromSpec(pipelineInfo.Details.ResourceRequests)
}

func getResourceListFromSpec(resources *pps.ResourceSpec) (*v1.ResourceList, error) {
	result := make(v1.ResourceList)

	if resources.Cpu != 0 {
		cpuStr := fmt.Sprintf("%f", resources.Cpu)
		cpuQuantity, err := resource.ParseQuantity(cpuStr)
		if err != nil {
			log.Warnf("error parsing cpu string: %s: %+v", cpuStr, err)
		} else {
			result[v1.ResourceCPU] = cpuQuantity
		}
	}

	if resources.Memory != "" {
		memQuantity, err := resource.ParseQuantity(resources.Memory)
		if err != nil {
			log.Warnf("error parsing memory string: %s: %+v", resources.Memory, err)
		} else {
			result[v1.ResourceMemory] = memQuantity
		}
	}

	if resources.Disk != "" { // needed because not all versions of k8s support disk resources
		diskQuantity, err := resource.ParseQuantity(resources.Disk)
		if err != nil {
			log.Warnf("error parsing disk string: %s: %+v", resources.Disk, err)
		} else {
			result[v1.ResourceEphemeralStorage] = diskQuantity
		}
	}

	if resources.Gpu != nil {
		gpuStr := fmt.Sprintf("%d", resources.Gpu.Number)
		gpuQuantity, err := resource.ParseQuantity(gpuStr)
		if err != nil {
			log.Warnf("error parsing gpu string: %s: %+v", gpuStr, err)
		} else {
			result[v1.ResourceName(resources.Gpu.Type)] = gpuQuantity
		}
	}

	return &result, nil
}

// GetLimitsResourceList returns a list of resources from a pipeline
// ResourceSpec that it is maximally limited to.
func GetLimitsResourceList(limits *pps.ResourceSpec) (*v1.ResourceList, error) {
	return getResourceListFromSpec(limits)
}

// FailPipeline updates the pipeline's state to failed and sets the failure reason
func FailPipeline(ctx context.Context, db *pachsql.DB, pipelinesCollection col.PostgresCollection, specCommit *pfs.Commit, reason string) error {
	return SetPipelineState(ctx, db, pipelinesCollection, specCommit,
		nil, pps.PipelineState_PIPELINE_FAILURE, reason)
}

// CrashingPipeline updates the pipeline's state to crashing and sets the reason
func CrashingPipeline(ctx context.Context, db *pachsql.DB, pipelinesCollection col.PostgresCollection, specCommit *pfs.Commit, reason string) error {
	return SetPipelineState(ctx, db, pipelinesCollection, specCommit,
		nil, pps.PipelineState_PIPELINE_CRASHING, reason)
}

// PipelineTransitionError represents an error transitioning a pipeline from
// one state to another.
type PipelineTransitionError struct {
	Pipeline        string
	Expected        []pps.PipelineState
	Target, Current pps.PipelineState
}

func (p PipelineTransitionError) Error() string {
	var froms bytes.Buffer
	for i, state := range p.Expected {
		if i > 0 {
			froms.WriteString(", ")
		}
		froms.WriteString(state.String())
	}
	return fmt.Sprintf("could not transition %q from any of [%s] -> %s, as it is in %s",
		p.Pipeline, froms.String(), p.Target, p.Current)
}

// SetPipelineState does a lot of conditional logging, and converts 'from' and
// 'to' to strings, so the construction of its log message is factored into this
// helper.
func logSetPipelineState(pipeline string, from []pps.PipelineState, to pps.PipelineState, reason string) {
	var logMsg strings.Builder
	logMsg.Grow(300) // approx. max length of this log msg if len(from) <= ~2
	logMsg.WriteString("SetPipelineState attempting to move \"")
	logMsg.WriteString(pipeline)
	logMsg.WriteString("\" ")
	if len(from) > 0 {
		logMsg.WriteString("from one of {")
		logMsg.WriteString(from[0].String())
		for _, s := range from[1:] {
			logMsg.WriteByte(',')
			logMsg.WriteString(s.String())
		}
		logMsg.WriteString("} ")
	}
	logMsg.WriteString("to ")
	logMsg.WriteString(to.String())
	if reason != "" {
		logMsg.WriteString(" (reason: \"")
		logMsg.WriteString(reason)
		logMsg.WriteString("\")")
	}
	log.Info(logMsg.String())
}

// SetPipelineState is a helper that moves the state of 'pipeline' from any of
// the states in 'from' (if not nil) to 'to'. It will annotate any trace in
// 'ctx' with information about 'pipeline' that it reads.
//
// This function logs a lot for a library function, but it's mostly (maybe
// exclusively?) called by the PPS master
func SetPipelineState(ctx context.Context, db *pachsql.DB, pipelinesCollection col.PostgresCollection, specCommit *pfs.Commit, from []pps.PipelineState, to pps.PipelineState, reason string) (retErr error) {
	pipeline := specCommit.Branch.Repo.Name
	logSetPipelineState(pipeline, from, to, reason)
	var resultMessage string
	var warn bool
	err := dbutil.WithTx(ctx, db, func(sqlTx *pachsql.Tx) error {
		resultMessage = ""
		warn = false
		pipelines := pipelinesCollection.ReadWrite(sqlTx)
		pipelineInfo := &pps.PipelineInfo{}
		if err := pipelines.Get(specCommit, pipelineInfo); err != nil {
			return errors.EnsureStack(err)
		}
		tracing.TagAnySpan(ctx, "old-state", pipelineInfo.State)
		// Only UpdatePipeline can bring a pipeline out of failure
		// TODO(msteffen): apply the same logic for CRASHING?
		if pipelineInfo.State == pps.PipelineState_PIPELINE_FAILURE {
			if to != pps.PipelineState_PIPELINE_FAILURE {
				resultMessage = fmt.Sprintf("cannot move pipeline %q to %s when it is already in FAILURE", pipeline, to)
				warn = true
			}
			return nil
		}
		// Don't allow a transition from STANDBY to CRASHING if we receive events out of order
		if pipelineInfo.State == pps.PipelineState_PIPELINE_STANDBY && to == pps.PipelineState_PIPELINE_CRASHING {
			resultMessage = fmt.Sprintf("cannot move pipeline %q to CRASHING when it is in STANDBY", pipeline)
			warn = true
			return nil
		}

		// transitionPipelineState case: error if pipeline is in an unexpected
		// state.
		//
		// allow transitionPipelineState to send a pipeline state to its target
		// repeatedly (thus pipelineInfo.State == to yields no error). This will
		// trigger additional etcd write events, but will not trigger an error.
		if len(from) > 0 {
			var isInFromState bool
			for _, fromState := range from {
				if pipelineInfo.State == fromState {
					isInFromState = true
					break
				}
			}
			if !isInFromState && pipelineInfo.State != to {
				return PipelineTransitionError{
					Pipeline: pipeline,
					Expected: from,
					Target:   to,
					Current:  pipelineInfo.State,
				}
			}
		}
		resultMessage = fmt.Sprintf("SetPipelineState moved pipeline %s from %s to %s", pipeline, pipelineInfo.State, to)
		pipelineInfo.State = to
		pipelineInfo.Reason = reason
		return errors.EnsureStack(pipelines.Put(specCommit, pipelineInfo))
	})
	if resultMessage != "" {
		if warn {
			log.Warn(resultMessage)
		} else {
			log.Info(resultMessage)
		}
	}
	return err
}

// JobInput fills in the commits for an Input
func JobInput(pipelineInfo *pps.PipelineInfo, outputCommit *pfs.Commit) *pps.Input {
	commitsetID := outputCommit.ID
	jobInput := proto.Clone(pipelineInfo.Details.Input).(*pps.Input)
	pps.VisitInput(jobInput, func(input *pps.Input) error { //nolint:errcheck
		if input.Pfs != nil {
			input.Pfs.Commit = commitsetID
		}
		if input.Cron != nil {
			input.Cron.Commit = commitsetID
		}
		return nil
	})
	return jobInput
}

// PipelineReqFromInfo converts a PipelineInfo into a CreatePipelineRequest.
func PipelineReqFromInfo(pipelineInfo *pps.PipelineInfo) *pps.CreatePipelineRequest {
	return &pps.CreatePipelineRequest{
		Pipeline:              pipelineInfo.Pipeline,
		Transform:             pipelineInfo.Details.Transform,
		ParallelismSpec:       pipelineInfo.Details.ParallelismSpec,
		Egress:                pipelineInfo.Details.Egress,
		OutputBranch:          pipelineInfo.Details.OutputBranch,
		ResourceRequests:      pipelineInfo.Details.ResourceRequests,
		ResourceLimits:        pipelineInfo.Details.ResourceLimits,
		SidecarResourceLimits: pipelineInfo.Details.SidecarResourceLimits,
		Input:                 pipelineInfo.Details.Input,
		Description:           pipelineInfo.Details.Description,
		Service:               pipelineInfo.Details.Service,
		DatumSetSpec:          pipelineInfo.Details.DatumSetSpec,
		DatumTimeout:          pipelineInfo.Details.DatumTimeout,
		JobTimeout:            pipelineInfo.Details.JobTimeout,
		Salt:                  pipelineInfo.Details.Salt,
		PodSpec:               pipelineInfo.Details.PodSpec,
		PodPatch:              pipelineInfo.Details.PodPatch,
		Spout:                 pipelineInfo.Details.Spout,
		SchedulingSpec:        pipelineInfo.Details.SchedulingSpec,
		DatumTries:            pipelineInfo.Details.DatumTries,
		S3Out:                 pipelineInfo.Details.S3Out,
		Metadata:              pipelineInfo.Details.Metadata,
		ReprocessSpec:         pipelineInfo.Details.ReprocessSpec,
		Autoscaling:           pipelineInfo.Details.Autoscaling,
	}
}

// UpdateJobState performs the operations involved with a job state transition.
func UpdateJobState(pipelines col.PostgresReadWriteCollection, jobs col.ReadWriteCollection, jobInfo *pps.JobInfo, state pps.JobState, reason string) error {
	// Check if this is a new job
	if jobInfo.State != pps.JobState_JOB_STATE_UNKNOWN {
		if pps.IsTerminal(jobInfo.State) {
			return ppsServer.ErrJobFinished{Job: jobInfo.Job}
		}
	}

	// Update job info
	var err error
	if jobInfo.State == pps.JobState_JOB_STARTING && state == pps.JobState_JOB_RUNNING {
		jobInfo.Started, err = types.TimestampProto(time.Now())
		if err != nil {
			return errors.EnsureStack(err)
		}
	} else if pps.IsTerminal(state) {
		if jobInfo.Started == nil {
			jobInfo.Started, err = types.TimestampProto(time.Now())
			if err != nil {
				return errors.EnsureStack(err)
			}
		}
		jobInfo.Finished, err = types.TimestampProto(time.Now())
		if err != nil {
			return errors.EnsureStack(err)
		}
	}
	jobInfo.State = state
	jobInfo.Reason = reason
	return errors.EnsureStack(jobs.Put(ppsdb.JobKey(jobInfo.Job), jobInfo))
}

func FinishJob(pachClient *client.APIClient, jobInfo *pps.JobInfo, state pps.JobState, reason string) error {
	jobInfo.State = state
	jobInfo.Reason = reason
	// TODO: Leaning on the reason rather than state for commit errors seems a bit sketchy, but we don't
	// store commit states.

	// only try to close meta commits for transform pipelines. We can't simply ignore a NotFound error
	// because the real error happens on the server and is returned by RunBatchInTransaction itself
	hasMeta := jobInfo.GetDetails().GetSpout() == nil && jobInfo.GetDetails().GetService() == nil
	_, err := pachClient.RunBatchInTransaction(func(builder *client.TransactionBuilder) error {
		if hasMeta {
			if _, err := builder.PfsAPIClient.FinishCommit(pachClient.Ctx(), &pfs.FinishCommitRequest{
				Commit: MetaCommit(jobInfo.OutputCommit),
				Error:  reason,
				Force:  true,
			}); err != nil {
				return errors.EnsureStack(err)
			}
		}
		if _, err := builder.PfsAPIClient.FinishCommit(pachClient.Ctx(), &pfs.FinishCommitRequest{
			Commit: jobInfo.OutputCommit,
			Error:  reason,
			Force:  true,
		}); err != nil {
			return errors.EnsureStack(err)
		}
		return WriteJobInfo(&builder.APIClient, jobInfo)
	})
	return err
}

func WriteJobInfo(pachClient *client.APIClient, jobInfo *pps.JobInfo) error {
	_, err := pachClient.PpsAPIClient.UpdateJobState(pachClient.Ctx(), &pps.UpdateJobStateRequest{
		Job:           jobInfo.Job,
		State:         jobInfo.State,
		Reason:        jobInfo.Reason,
		Restart:       jobInfo.Restart,
		DataProcessed: jobInfo.DataProcessed,
		DataSkipped:   jobInfo.DataSkipped,
		DataTotal:     jobInfo.DataTotal,
		DataFailed:    jobInfo.DataFailed,
		DataRecovered: jobInfo.DataRecovered,
		Stats:         jobInfo.Stats,
	})
	return errors.EnsureStack(err)
}

func MetaCommit(commit *pfs.Commit) *pfs.Commit {
	return client.NewSystemRepo(commit.Branch.Repo.Name, pfs.MetaRepoType).NewCommit(commit.Branch.Name, commit.ID)
}

// ContainsS3Inputs returns 'true' if 'in' is or contains any PFS inputs with
// 'S3' set to true. Any pipelines with s3 inputs lj
func ContainsS3Inputs(in *pps.Input) bool {
	var found bool
	pps.VisitInput(in, func(in *pps.Input) error { //nolint:errcheck
		if in.Pfs != nil && in.Pfs.S3 {
			found = true
			return errutil.ErrBreak
		}
		return nil
	})
	return found
}

// SidecarS3GatewayService returns the name of the kubernetes service created
// for the job 'jobID' to hand sidecar s3 gateway requests. This helper
// is in ppsutil because both PPS (which creates the service, in the s3 gateway
// sidecar server) and the worker (which passes the endpoint to the user code)
// need to know it.
func SidecarS3GatewayService(pipeline, commitSetId string) string {
	hash := md5.New()
	hash.Write([]byte(pipeline))
	hash.Write([]byte(commitSetId))
	return "s3-" + pfs.EncodeHash(hash.Sum(nil))
}

// ErrorState returns true if s is an error state for a pipeline, that is, a
// state that users should be aware of and one which will have a "Reason" set
// for why it's in this state.
func ErrorState(s pps.PipelineState) bool {
	return map[pps.PipelineState]bool{
		pps.PipelineState_PIPELINE_FAILURE:    true,
		pps.PipelineState_PIPELINE_CRASHING:   true,
		pps.PipelineState_PIPELINE_RESTARTING: true,
	}[s]
}

// GetWorkerPipelineInfo gets the PipelineInfo proto describing the pipeline that this
// worker is part of.
// getPipelineInfo has the side effect of adding auth to the passed pachClient
func GetWorkerPipelineInfo(pachClient *client.APIClient, db *pachsql.DB, l collection.PostgresListener, pipelineName, specCommitID string) (*pps.PipelineInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	pipelines := ppsdb.Pipelines(db, l)
	pipelineInfo := &pps.PipelineInfo{}
	// Notice we use the SpecCommitID from our env, not from postgres. This is
	// because the value in postgres might get updated while the worker pod is
	// being created and we don't want to run the transform of one version of
	// the pipeline in the image of a different verison.
	specCommit := client.NewSystemRepo(pipelineName, pfs.SpecRepoType).
		NewCommit("master", specCommitID)
	if err := pipelines.ReadOnly(ctx).Get(specCommit, pipelineInfo); err != nil {
		return nil, errors.EnsureStack(err)
	}
	pachClient.SetAuthToken(pipelineInfo.AuthToken)

	return pipelineInfo, nil
}

func FindPipelineSpecCommit(ctx context.Context, pfsServer pfsServer.APIServer, txnEnv transactionenv.TransactionEnv, pipeline string) (*pfs.Commit, error) {
	var commit *pfs.Commit
	if err := txnEnv.WithReadContext(ctx, func(txnCtx *txncontext.TransactionContext) (err error) {
		commit, err = FindPipelineSpecCommitInTransaction(txnCtx, pfsServer, pipeline, "")
		return
	}); err != nil {
		return nil, err
	}
	return commit, nil
}

// FindPipelineSpecCommitInTransaction finds the spec commit corresponding to the pipeline version present in the commit given
// by startID. If startID is blank, find the current pipeline version
func FindPipelineSpecCommitInTransaction(txnCtx *txncontext.TransactionContext, pfsServer pfsServer.APIServer, pipeline, startID string) (*pfs.Commit, error) {
	curr := client.NewSystemRepo(pipeline, pfs.SpecRepoType).NewCommit("master", startID)
	commitInfo, err := pfsServer.InspectCommitInTransaction(txnCtx,
		&pfs.InspectCommitRequest{Commit: curr})
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	for commitInfo.Origin.Kind != pfs.OriginKind_USER {
		curr = commitInfo.ParentCommit
		if curr == nil {
			return nil, errors.Errorf("spec commit for pipeline %s not found", pipeline)
		}
		if commitInfo, err = pfsServer.InspectCommitInTransaction(txnCtx,
			&pfs.InspectCommitRequest{Commit: curr}); err != nil {
			return nil, errors.EnsureStack(err)
		}
	}

	return curr, nil
}

// ListPipelineInfo enumerates all PPS pipelines in the database, filters them
// based on 'request', and then calls 'f' on each value
func ListPipelineInfo(ctx context.Context,
	pipelines collection.PostgresCollection,
	pipeline *pps.Pipeline,
	history int64,
	f func(*pps.PipelineInfo) error) error {
	p := &pps.PipelineInfo{}
	versionMap := make(map[string]uint64)
	checkPipelineVersion := func(_ string) error {
		// Erase any AuthToken - this shouldn't be returned to anyone (the workers
		// won't use this function to get their auth token)
		p.AuthToken = ""
		// TODO: this is kind of silly - callers should just make a version range for each pipeline?
		if last, ok := versionMap[p.Pipeline.Name]; ok {
			if p.Version < last {
				// don't send, exit early
				return nil
			}
		} else {
			// we haven't seen this pipeline yet, rely on sort order and assume this is latest
			var lastVersionToSend uint64
			if history < 0 || uint64(history) >= p.Version {
				lastVersionToSend = 1
			} else {
				lastVersionToSend = p.Version - uint64(history)
			}
			versionMap[p.Pipeline.Name] = lastVersionToSend
		}

		return f(p)
	}
	if pipeline != nil {
		if err := pipelines.ReadOnly(ctx).GetByIndex(
			ppsdb.PipelinesNameIndex,
			pipeline.Name,
			p,
			col.DefaultOptions(),
			checkPipelineVersion); err != nil {
			return errors.EnsureStack(err)
		}
		if len(versionMap) == 0 {
			// pipeline didn't exist after all
			return ppsServer.ErrPipelineNotFound{Pipeline: pipeline}
		}
		return nil
	}
	return errors.EnsureStack(pipelines.ReadOnly(ctx).List(p, col.DefaultOptions(), checkPipelineVersion))
}

func FilterLogLines(request *pps.GetLogsRequest, r io.Reader, plainText bool, send func(*pps.LogMessage) error) error {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		msg := new(pps.LogMessage)
		if plainText {
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
			if request.Job != nil && (request.Job.ID != msg.JobID || request.Job.Pipeline.Name != msg.PipelineName) {
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
		// Log message passes all filters -- return it
		msg.Message = strings.TrimSuffix(msg.Message, "\n")
		if err := send(msg); err != nil {
			if errors.Is(err, errutil.ErrBreak) {
				return nil
			}
			return err
		}
	}
	return nil
}
