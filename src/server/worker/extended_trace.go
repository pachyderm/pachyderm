package worker

import (
	"context"
	"fmt"

	etcd "github.com/coreos/etcd/clientv3"
	opentracing "github.com/opentracing/opentracing-go"
	log "github.com/sirupsen/logrus"

	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/tracing"
	"github.com/pachyderm/pachyderm/src/client/pkg/tracing/extended"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
)

func addJobSpanToAnyExtendedTrace(ctx context.Context, c *etcd.Client, commit *pfs.CommitInfo) (opentracing.Span, context.Context) {
	if !tracing.IsActive() {
		return nil, ctx // no Jaeger instance to send trace info to
	}

	// use key from 'ctx' to read full ExtendedTrace reference from etcd
	tracesCol := extended.TracesCol(c).ReadOnly(ctx)
	var allCommitsHaveBeenTraced bool
	var extendedTraceID string
	var extendedTrace extended.TraceProto
	var spanCtx opentracing.SpanContext
	// options can't be nil, but there should only be at most one trace attached
	// to 'commit' so the options shouldn't matter
	if err := tracesCol.GetByIndex(extended.CommitIDIndex, commit.Commit.ID, &extendedTrace, extended.TraceGetOpts,
		func(key string) error {
			if spanCtx != nil {
				return fmt.Errorf("second, unexpected span associated with commit %q: %s",
					commit.Commit.ID, key)
			}
			extendedTraceID = key

			// See if 'extendedTrace's commit IDs are all in 'commit's provenance. If
			// so, remove 'extendedTrace' from etcd after this is done
			allCommitsHaveBeenTraced = true
			provCommits := make(map[string]struct{})
			for _, c := range commit.Provenance {
				provCommits[c.ID] = struct{}{}
			}
			for _, c := range extendedTrace.CommitIDs {
				if _, ok := provCommits[c]; !ok {
					allCommitsHaveBeenTraced = false
					break
				}
			}

			// Create new opentracing span from 'extendedTrace'
			var err error
			spanCtx, err = opentracing.GlobalTracer().Extract(opentracing.TextMap,
				opentracing.TextMapCarrier(extendedTrace.SerializedTrace))
			if err != nil {
				return fmt.Errorf("could not extract span context from ExtendedTrace proto: %v", err)
			}
			return nil
		}); err != nil {
		log.Errorf("tracing error: %v", err)
		return nil, nil
	}
	if spanCtx != nil {
		// Delete 'extendedTrace' from etcd if 'commit' is the last commit that it
		// covers.
		if allCommitsHaveBeenTraced {
			if _, err := col.NewSTM(ctx, c, func(stm col.STM) error {
				return extended.TracesCol(c).ReadWrite(stm).Delete(extendedTraceID)
			}); err != nil {
				log.Errorf("error deleting trace for %q: %v", commit.Commit.ID, err)
			}
		}

		// return new span
		return opentracing.StartSpanFromContext(ctx,
			"worker.SpawnJob", opentracing.FollowsFrom(spanCtx))
	}

	// If there's no trace for 'commit', see if there's a trace for the pipeline
	// (indicating that the pipeline was restarted, the new pipeline spec has an
	// output commit, and the startup is being traced). If it has a commit
	// that's NOT equal to 'commit', then we're scanning through all ancestors
	// of 'commit' and we want to attach that to the pipeline startup trace
	var pipeline = commit.Commit.Repo.Name
	if err := tracesCol.GetByIndex(extended.PipelineIndex, pipeline, &extendedTrace, extended.TraceGetOpts,
		func(key string) error {
			if spanCtx != nil {
				return fmt.Errorf("second, unexpected span associated with pipeline %q: %s",
					pipeline, key)
			}
			extendedTraceID = key

			// if extendedTrace has no commits, but we know this pipeline has input
			// data (because worker/master is getting commits) then this is a stale
			// startup trace -- delete it
			if len(extendedTrace.CommitIDs) == 0 {
				if _, err := col.NewSTM(ctx, c, func(stm col.STM) error {
					return extended.TracesCol(c).ReadWrite(stm).Delete(extendedTraceID)
				}); err != nil {
					log.Errorf("error deleting trace for %q: %v", commit.Commit.ID, err)
				}
				return nil
			}

			// This is a startup trace that's tied to a future output commit. Create
			// new opentracing span from 'extendedTrace'
			var err error
			spanCtx, err = opentracing.GlobalTracer().Extract(opentracing.TextMap,
				opentracing.TextMapCarrier(extendedTrace.SerializedTrace))
			if err != nil {
				return fmt.Errorf("could not extract span context from ExtendedTrace proto: %v", err)
			}
			return nil
		}); err != nil {
		log.Errorf("tracing error: %v", err)
		return nil, ctx
	}
	if spanCtx == nil {
		return nil, ctx
	}
	return opentracing.StartSpanFromContext(ctx,
		"worker.ProcessCommit", opentracing.FollowsFrom(spanCtx),
		opentracing.Tag{"commit", commit.Commit.ID})
}

func addStartupSpanToAnyExtendedTrace(ctx context.Context, c *etcd.Client, pipeline string) (opentracing.Span, context.Context) {
	if !tracing.IsActive() {
		return nil, ctx // no Jaeger instance to send trace info to
	}

	// use key from 'ctx' to read full ExtendedTrace reference from etcd
	tracesCol := extended.TracesCol(c).ReadOnly(ctx)
	var extendedTraceID string
	var extendedTrace extended.TraceProto
	var spanCtx opentracing.SpanContext
	if err := tracesCol.GetByIndex(extended.PipelineIndex, pipeline, &extendedTrace, extended.TraceGetOpts,
		func(key string) error {
			if spanCtx != nil {
				return fmt.Errorf("second, unexpected span associated with pipeline %q: %s",
					pipeline)
			}
			extendedTraceID = key

			// Create new opentracing span from 'extendedTrace'
			var err error
			spanCtx, err = opentracing.GlobalTracer().Extract(opentracing.TextMap,
				opentracing.TextMapCarrier(extendedTrace.SerializedTrace))
			if err != nil {
				return fmt.Errorf("could not extract span context from ExtendedTrace proto: %v", err)
			}
			return nil
		}); err != nil {
		log.Errorf("tracing error: %v", err)
		return nil, nil
	}
	if spanCtx == nil {
		return nil, nil
	}

	// return new span
	return opentracing.StartSpanFromContext(ctx,
		"worker.Start", opentracing.FollowsFrom(spanCtx))
}
