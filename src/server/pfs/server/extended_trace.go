package server

import (
	etcd "github.com/coreos/etcd/clientv3"
	log "github.com/sirupsen/logrus"

	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/tracing"
	"github.com/pachyderm/pachyderm/src/client/pkg/tracing/extended"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
	"github.com/pachyderm/pachyderm/src/server/pkg/pfsdb"
)

// attachNewCommitsToAnyExtendedTrace:
// (1) extracts the extended trace (if any) from stm.Context()
// (2) gets the extended trace's target branch
// (3) finds all commits provenant on the HEAD of (2)
// (4) attaches those commits to the extended trace from (1)
//
// Note: this must be called as soon as possible after downstream commits have
// been created for a new input commit, so that workers can retrieve the span
// and use it. Currently, attachNewCommitsToAnyExtendedTrace is called in
// propagateCommit, which creates these downstream commits and also
// conveniently pre-loads all of the read results below into 'stm's internal
// cache
//
// Note: 'c' should be the client associated with 'stm'
// TODO(msteffen): make it possible to extract an etcd client from an stm
func attachNewCommitsToAnyExtendedTrace(c *etcd.Client, etcdPrefix string, stm col.STM) {
	if !tracing.IsActive() {
		return
	}

	// get extended trace from ctx
	extendedTrace, err := extended.GetTraceFromCtx(stm.Context())
	if extendedTrace == nil {
		if err != nil {
			log.Errorf(err.Error())
		}
		return
	}
	traceID := extendedTrace.SerializedTrace["uber-trace-id"]

	// Delete any stale traces for the trace's pipeline, if it has one, before
	// writing it to etcd (can happen if a CreatePipeline RPC errored and PPS
	// never deleted it)
	tracesCol := extended.TracesCol(c).ReadWrite(stm)
	if extendedTrace.Pipeline != "" {
		var oldExtendedTrace extended.TraceProto
		// var spanCtx opentracing.SpanContext
		// options can't be nil, but there should only be at most one trace attached
		// to 'pipeline' so the options shouldn't matter
		if err := extended.TracesCol(c).ReadOnly(stm.Context()).GetByIndex(
			extended.PipelineIndex, extendedTrace.Pipeline, &oldExtendedTrace, extended.TraceGetOpts,
			func(key string) error {
				if key != traceID {
					return tracesCol.Delete(key)
				}
				return nil
			}); err != nil && !col.IsErrNotFound(err) {
			log.Errorf("could not delete old trace for %q: %v", extendedTrace.Pipeline, err)
			return
		}
	}

	// Attach HEAD of each branch in extendedTrace.Branch's provenance to
	// extendedTrace
	var (
		targetBranch     = extendedTrace.Branch
		targetBranchInfo pfs.BranchInfo
		branchesCol      = pfsdb.Branches(c, etcdPrefix, targetBranch.Repo.Name).ReadWrite(stm)
	)
	if err := branchesCol.Get(targetBranch.Name, &targetBranchInfo); err != nil {
		log.Errorf("error getting branch info for extended trace target branch \"%s/%s\": %v",
			targetBranch.Repo.Name, targetBranch.Name, err)
		// Note: Don't return, so we can still write extended trace to etcd. This
		// usually happens when the pipeline's output branch hasn't been created
		// yet.  Initially the extended trace will be written without any commit
		// IDs, and will only be indexed by its pipeline. In this case, it'll be
		// re-written later, after the pipeline's output branch is created.
	}
	for _, provBranch := range targetBranchInfo.Provenance {
		var provBranchInfo pfs.BranchInfo
		if err := branchesCol.Get(provBranch.Name, &provBranchInfo); err != nil {
			log.Errorf("error getting branch info for extended trace provenant branch \"%s/%s\": %v",
				provBranch.Repo.Name, provBranch.Name, err)
			continue
		}
		if provBranchInfo.Head != nil {
			extendedTrace.CommitIDs = append(extendedTrace.CommitIDs, provBranchInfo.Head.ID)
		}
	}

	// copy extendedTrace into etcd
	if err := tracesCol.Create(traceID, extendedTrace); err != nil {
		log.Errorf("error writing extended trace to etcd: %v", err)
	}
}
