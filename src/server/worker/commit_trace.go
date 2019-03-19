package worker

import (
	"context"
	"fmt"

	etcd "github.com/coreos/etcd/clientv3"
	opentracing "github.com/opentracing/opentracing-go"
	log "github.com/sirupsen/logrus"

	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/tracing"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
)

// commitIDIndex is the secondary index for commit traces by the set of commit
// IDs watched by the trace
var commitIDIndex = &col.Index{
	Field: "CommitIDs",
	Multi: true,
}

func addJobSpanToAnyCommitTrace(ctx context.Context, c *etcd.Client, commit *pfs.CommitInfo) (opentracing.Span, context.Context) {
	// use key from 'ctx' to read full CommitTrace reference from etcd
	tracesCol := col.NewCollection(c,
		pfs.TracesCollectionPrefix,
		[]*col.Index{commitIDIndex},
		&pfs.CommitTrace{},
		nil,
		nil).ReadOnly(ctx)

	var allCommitsHaveBeenTraced = true
	var commitTraceID string
	var commitTrace pfs.CommitTrace
	var spanCtx opentracing.SpanContext
	// options can't be nil, but there should only be at most one trace attached
	// to 'commit' so the options shouldn't matter
	fmt.Printf(">>> [worker] commit: %s\n", commit.Commit.ID)
	if err := tracesCol.GetByIndex(commitIDIndex, commit.Commit.ID, &commitTrace,
		&col.Options{
			Target:   etcd.SortByKey,
			Order:    etcd.SortNone,
			SelfSort: false,
		},
		func(key string) error {
			fmt.Printf(">>> [worker] commitTrace:\n%v\n", commitTrace)
			if spanCtx != nil {
				return fmt.Errorf("second, unexpected span associated with commit %q: %s",
					commit.Commit.ID, key)
			}
			commitTraceID = key

			// See if 'commitTrace's commit IDs are all in 'commit's provenance. If so,
			// remove 'commitTrace' from etcd after this is done
			provCommits := make(map[string]struct{})
			for _, c := range commit.Provenance {
				provCommits[c.ID] = struct{}{}
			}
			for _, c := range commitTrace.CommitIDs {
				if _, ok := provCommits[c]; !ok {
					allCommitsHaveBeenTraced = false
					break
				}
			}

			// Create new opentracing span from 'commitTrace'
			if !tracing.IsActive() {
				return nil // no Jaeger instance to send trace info to
			}
			var err error
			spanCtx, err = opentracing.GlobalTracer().Extract(opentracing.TextMap,
				opentracing.TextMapCarrier(commitTrace.Value))
			if err != nil {
				return fmt.Errorf("could not extract span context from CommitTrace proto: %v", err)
			}
			return nil
		}); err != nil {
		log.Errorf("tracing error: %v", err)
		return nil, nil
	}
	if spanCtx == nil {
		return nil, nil
	}

	// Delete commitTrace from etcd if it's no longer needed
	if allCommitsHaveBeenTraced {
		if _, err := col.NewSTM(ctx, c, func(stm col.STM) error {
			return col.NewCollection(c,
				pfs.TracesCollectionPrefix,
				[]*col.Index{commitIDIndex},
				&pfs.CommitTrace{},
				nil,
				nil).ReadWrite(stm).Delete(commitTraceID)
		}); err != nil {
			log.Errorf("error deleting trace for %q: %v", commit.Commit.ID, err)
		}
	}

	// return new span
	fmt.Printf(">>> [worker] spanCtx: %v\n", spanCtx)
	return opentracing.StartSpanFromContext(ctx,
		"worker.SpawnJob", opentracing.FollowsFrom(spanCtx))
}
