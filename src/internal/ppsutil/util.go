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
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/pachyderm/pachyderm/v2/src/client"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsconsts"
	"github.com/pachyderm/pachyderm/v2/src/internal/tracing"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	ppsServer "github.com/pachyderm/pachyderm/v2/src/server/pps"
)

// JobKey is the string representation of a Job suitable for use as an indexing key
func JobKey(job *pps.Job) string {
	return job.String()
}

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
	name = strings.Replace(name, "_", "-", -1)
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

// GetPipelineDetails retrieves the pipeline spec for the given PipelineInfo from PFS,
// then fills in the Details field in the given PipelineInfo from the spec.
func GetPipelineDetails(pachClient *client.APIClient, pipelineInfo *pps.PipelineInfo) error {
	// ensure we are authorized to read the pipeline's spec commit, but don't propagate that back out
	pachClient = pachClient.WithCtx(pachClient.Ctx())
	pachClient.SetAuthToken(pipelineInfo.AuthToken)

	buf := bytes.Buffer{}
	if err := pachClient.GetFile(pipelineInfo.SpecCommit, ppsconsts.SpecFile, &buf); err != nil {
		return errors.Wrapf(err, "could not retrieve pipeline spec file from PFS for pipeline '%s'", pipelineInfo.Pipeline.Name)
	}

	loadedPipelineInfo := &pps.PipelineInfo{}
	if err := loadedPipelineInfo.Unmarshal(buf.Bytes()); err != nil {
		return errors.Wrapf(err, "could not unmarshal PipelineInfo bytes from PFS")
	}
	pipelineInfo.Version = loadedPipelineInfo.Version
	pipelineInfo.Details = loadedPipelineInfo.Details
	return nil
}

// FailPipeline updates the pipeline's state to failed and sets the failure reason
func FailPipeline(ctx context.Context, db *sqlx.DB, pipelinesCollection col.PostgresCollection, pipelineName string, reason string) error {
	return SetPipelineState(ctx, db, pipelinesCollection, pipelineName,
		nil, pps.PipelineState_PIPELINE_FAILURE, reason)
}

// CrashingPipeline updates the pipeline's state to crashing and sets the reason
func CrashingPipeline(ctx context.Context, db *sqlx.DB, pipelinesCollection col.PostgresCollection, pipelineName string, reason string) error {
	return SetPipelineState(ctx, db, pipelinesCollection, pipelineName,
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
func SetPipelineState(ctx context.Context, db *sqlx.DB, pipelinesCollection col.PostgresCollection, pipeline string, from []pps.PipelineState, to pps.PipelineState, reason string) (retErr error) {
	logSetPipelineState(pipeline, from, to, reason)
	err := col.NewSQLTx(ctx, db, func(sqlTx *sqlx.Tx) error {
		pipelines := pipelinesCollection.ReadWrite(sqlTx)
		pipelineInfo := &pps.PipelineInfo{}
		if err := pipelines.Get(pipeline, pipelineInfo); err != nil {
			return err
		}
		tracing.TagAnySpan(ctx, "old-state", pipelineInfo.State)
		// Only UpdatePipeline can bring a pipeline out of failure
		// TODO(msteffen): apply the same logic for CRASHING?
		if pipelineInfo.State == pps.PipelineState_PIPELINE_FAILURE {
			if to != pps.PipelineState_PIPELINE_FAILURE {
				log.Warningf("cannot move pipeline %q to %s when it is already in FAILURE", pipeline, to)
			}
			return nil
		}
		// Don't allow a transition from STANDBY to CRASHING if we receive events out of order
		if pipelineInfo.State == pps.PipelineState_PIPELINE_STANDBY && to == pps.PipelineState_PIPELINE_CRASHING {
			log.Warningf("cannot move pipeline %q to CRASHING when it is in STANDBY", pipeline)
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
		log.Infof("SetPipelineState moving pipeline %s from %s to %s", pipeline, pipelineInfo.State, to)
		pipelineInfo.State = to
		pipelineInfo.Reason = reason
		return pipelines.Put(pipeline, pipelineInfo)
	})
	return err
}

// JobInput fills in the commits for an Input
func JobInput(pipelineInfo *pps.PipelineInfo, outputCommit *pfs.Commit) *pps.Input {
	commitsetID := outputCommit.ID
	jobInput := proto.Clone(pipelineInfo.Details.Input).(*pps.Input)
	pps.VisitInput(jobInput, func(input *pps.Input) error {
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
		CacheSize:             pipelineInfo.Details.CacheSize,
		MaxQueueSize:          pipelineInfo.Details.MaxQueueSize,
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

// IsTerminal returns 'true' if 'state' indicates that the job is done (i.e.
// the state will not change later: SUCCESS, FAILURE, KILLED) and 'false'
// otherwise.
func IsTerminal(state pps.JobState) bool {
	switch state {
	case pps.JobState_JOB_SUCCESS, pps.JobState_JOB_FAILURE, pps.JobState_JOB_KILLED:
		return true
	case pps.JobState_JOB_CREATED, pps.JobState_JOB_STARTING, pps.JobState_JOB_RUNNING, pps.JobState_JOB_EGRESSING:
		return false
	default:
		panic(fmt.Sprintf("unrecognized job state: %s", state))
	}
}

// UpdateJobState performs the operations involved with a job state transition.
func UpdateJobState(pipelines col.ReadWriteCollection, jobs col.ReadWriteCollection, jobInfo *pps.JobInfo, state pps.JobState, reason string) error {
	// Check if this is a new job
	if jobInfo.State != pps.JobState_JOB_STATE_UNKNOWN {
		if IsTerminal(jobInfo.State) {
			return ppsServer.ErrJobFinished{Job: jobInfo.Job}
		}
	}

	// Update pipeline
	pipelineInfo := &pps.PipelineInfo{}
	if err := pipelines.Get(jobInfo.Job.Pipeline.Name, pipelineInfo); err != nil {
		return err
	}
	if pipelineInfo.JobCounts == nil {
		pipelineInfo.JobCounts = make(map[int32]int32)
	}
	// If the old state is UNKNOWN, this is a new job, don't decrement the count
	if jobInfo.State != pps.JobState_JOB_STATE_UNKNOWN {
		if pipelineInfo.JobCounts[int32(jobInfo.State)] != 0 {
			pipelineInfo.JobCounts[int32(jobInfo.State)]--
		}
	}
	pipelineInfo.JobCounts[int32(state)]++
	pipelineInfo.LastJobState = state
	if err := pipelines.Put(jobInfo.Job.Pipeline.Name, pipelineInfo); err != nil {
		return err
	}

	// Update job info
	var err error
	if state == pps.JobState_JOB_STARTING {
		jobInfo.Started, err = types.TimestampProto(time.Now())
		if err != nil {
			return err
		}
	} else if IsTerminal(state) {
		if jobInfo.Started == nil {
			jobInfo.Started, err = types.TimestampProto(time.Now())
			if err != nil {
				return err
			}
		}
		jobInfo.Finished, err = types.TimestampProto(time.Now())
		if err != nil {
			return err
		}
	}
	jobInfo.State = state
	jobInfo.Reason = reason
	return jobs.Put(JobKey(jobInfo.Job), jobInfo)
}

func FinishJob(pachClient *client.APIClient, jobInfo *pps.JobInfo, state pps.JobState, reason string) error {
	jobInfo.State = state
	jobInfo.Reason = reason
	var empty bool
	if state == pps.JobState_JOB_FAILURE || state == pps.JobState_JOB_KILLED {
		empty = true
	}
	_, err := pachClient.RunBatchInTransaction(func(builder *client.TransactionBuilder) error {
		if _, err := builder.PfsAPIClient.FinishCommit(pachClient.Ctx(), &pfs.FinishCommitRequest{
			Commit: jobInfo.OutputCommit,
			Empty:  empty,
		}); err != nil {
			return err
		}
		if _, err := builder.PfsAPIClient.FinishCommit(pachClient.Ctx(), &pfs.FinishCommitRequest{
			Commit: MetaCommit(jobInfo.OutputCommit),
			Empty:  empty,
		}); err != nil {
			return err
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
	return err
}

func MetaCommit(commit *pfs.Commit) *pfs.Commit {
	return client.NewSystemRepo(commit.Branch.Repo.Name, pfs.MetaRepoType).NewCommit(commit.Branch.Name, commit.ID)
}

// ContainsS3Inputs returns 'true' if 'in' is or contains any PFS inputs with
// 'S3' set to true. Any pipelines with s3 inputs lj
func ContainsS3Inputs(in *pps.Input) bool {
	var found bool
	pps.VisitInput(in, func(in *pps.Input) error {
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
func SidecarS3GatewayService(jobID string) string {
	return "s3-" + jobID
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
