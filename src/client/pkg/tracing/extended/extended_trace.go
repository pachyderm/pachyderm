package extended

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/tracing"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"

	etcd "github.com/coreos/etcd/clientv3"
	opentracing "github.com/opentracing/opentracing-go"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/metadata"
)

const (
	// TraceCtxKey is the grpc metadata key whose value is a ExtendedTrace
	// identifying the current RPC/commit
	TraceCtxKey = "commit-trace"

	// TracesCollectionPrefix is the prefix associated with the 'traces'
	// collection in etcd (which maps pipelines and commits to extended traces)
	TracesCollectionPrefix = "commit_traces"

	// TargetRepoEnvVar determines how long extended traces are updated until
	// they're deleted from the cluster
	TargetRepoEnvVar = "PACH_TRACING_TARGET_REPO"
)

var (
	// CommitIDIndex is a secondary index for extended traces by the set of
	// commit IDs watched by the trace
	CommitIDIndex = &col.Index{
		Field: "CommitIDs",
		Multi: true,
	}

	// PipelineIndex is a secondary index for extended traces by the pipelint
	// watched by the trace
	PipelineIndex = &col.Index{
		Field: "Pipeline",
	}

	// TraceGetOpts are the default options for retrieving a trace from
	// 'TracesCol'
	TraceGetOpts = &col.Options{Target: etcd.SortByKey, Order: etcd.SortNone, SelfSort: false}
)

// TracesCol returns the etcd collection of extended traces
func TracesCol(c *etcd.Client) col.Collection {
	return col.NewCollection(c,
		TracesCollectionPrefix,
		[]*col.Index{CommitIDIndex, PipelineIndex},
		&TraceProto{},
		nil,
		nil)
}

// GetTraceFromCtx extracts any extended trace embeded in 'ctx's metadata
func GetTraceFromCtx(ctx context.Context) (*TraceProto, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, nil // no trace
	}
	vals := md.Get(TraceCtxKey)
	if len(vals) < 1 {
		return nil, nil // no trace
	}
	if len(vals) > 1 {
		return nil, fmt.Errorf("ctx should have at most one extended trace, but found %d", len(vals))
	}
	marshalledProto, err := base64.URLEncoding.DecodeString(vals[0])
	if err != nil {
		return nil, fmt.Errorf("error base64-decoding marshalled ExtendedTrace proto: %v", err)
	}
	extendedTrace := &TraceProto{}
	if err := extendedTrace.Unmarshal(marshalledProto); err != nil {
		return nil, fmt.Errorf("error unmarshalling extended trace from ctx: %v", err)

	}
	return extendedTrace, nil
}

// TraceIn2Out copies any extended traces from the incoming RPC context in 'ctx'
// into the outgoing RPC context in 'ctx'. Currently, this is only called by
// CreatePipeline, when it forwards extended contexts to the PutFile RPC with
// the new commit info.
func TraceIn2Out(ctx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx // no trace
	}

	// Expected len('vals') is 1, but don't bother verifying
	vals := md.Get(TraceCtxKey)
	pairs := make([]string, 2*len(vals))
	for i := 0; i < len(pairs); i += 2 {
		pairs[i], pairs[i+1] = TraceCtxKey, vals[i/2]
	}
	return metadata.AppendToOutgoingContext(ctx, pairs...)
}

func (t *TraceProto) isValid() bool {
	return len(t.SerializedTrace) > 0
}

// AddPipelineSpanToAnyTrace finds any extended traces associated with
// 'pipeline', and if any such trace exists, it creates a new span associated
// with that trace and returns it
func AddPipelineSpanToAnyTrace(ctx context.Context, c *etcd.Client,
	pipeline, operation string, kvs ...interface{}) (opentracing.Span, context.Context) {
	if !tracing.IsActive() {
		return nil, ctx // no Jaeger instance to send trace info to
	}

	var tracesFound int
	var extendedTrace TraceProto
	tracesCol := TracesCol(c).ReadOnly(ctx)
	if err := tracesCol.GetByIndex(PipelineIndex, pipeline, &extendedTrace, TraceGetOpts,
		func(key string) error {
			if tracesFound++; tracesFound > 1 {
				log.Errorf("found second, unexpected span with key %q (using new key)", key)
			}
			return nil
		}); err != nil {
		log.Errorf("error getting trace via pipeline %q: %v", pipeline, err)
		return nil, ctx
	}
	if !extendedTrace.isValid() {
		return nil, ctx // no trace found
	}

	// Deserialize opentracing span from 'extendedTrace'
	spanCtx, err := opentracing.GlobalTracer().Extract(opentracing.TextMap,
		opentracing.TextMapCarrier(extendedTrace.SerializedTrace))
	if err != nil {
		log.Errorf("could not extract span context from ExtendedTrace proto: %v", err)
		return nil, ctx
	}

	// Construct new span's options (follows from extended span, + tags)
	options := []opentracing.StartSpanOption{
		opentracing.FollowsFrom(spanCtx),
		opentracing.Tag{"pipeline", pipeline},
	}
	for i := 0; i < len(kvs); i += 2 {
		if len(kvs) == i+1 {
			// likely forgot key or value--best effort
			options = append(options, opentracing.Tag{"extra", fmt.Sprintf("%v", kvs[i])})
			break
		}
		key, ok := kvs[i].(string) // common case--key is string
		if !ok {
			key = fmt.Sprintf("%v", kvs[i])
		}
		val, ok := kvs[i+1].(string) // common case--val is string
		if !ok {
			val = fmt.Sprintf("%v", kvs[i+1])
		}
		options = append(options, opentracing.Tag{key, val})
	}

	// return new span
	return opentracing.StartSpanFromContext(ctx,
		operation, options...)
}

// AddJobSpanToAnyTrace is like AddPipelineSpanToAnyTrace but looks for traces
// associated with the output commit 'commit'. Unlike AddPipelineSpan, this also
// may delete extended traces that it finds.
//
// In general, AddJobSpan will
// - Retrieve any trace associated with the pipeline identified with 'commit's
//   repo
// - Delete (from etcd, but not Jaeger) any extended trace that it finds where
//   all commits associated with the trace are in 'commit's provenance. This
//   covers the following two cases:
//   1. A job is being traced through 'commit's output commit. Once commit is
//      finished, no more spans should be added to the trace
//   2. The trace has no commits associated with it. This means the trace is for
//      initial pipeline creation, but a worker is now starting jobs for the
//      pipeline, so the pipeline is up and it's safe to stop tracing it
// - Add a span to any trace associated with 'commit's ID (case 2. above)
func AddJobSpanToAnyTrace(ctx context.Context, c *etcd.Client, commit *pfs.CommitInfo) (opentracing.Span, context.Context) {
	if !tracing.IsActive() {
		return nil, ctx // no Jaeger instance to send trace info to
	}

	// copy 'commit's provenance into a map, to make subset computation simpler
	provCommits := make(map[string]struct{})
	for _, c := range commit.Provenance {
		provCommits[c.Commit.ID] = struct{}{}
	}

	// read all traces associated with 'commit's pipeline (i.e. repo)
	tracesCol := TracesCol(c).ReadOnly(ctx)
	var xt, curXT TraceProto
	if err := tracesCol.GetByIndex(PipelineIndex, commit.Commit.Repo.Name, &curXT, TraceGetOpts,
		func(key string) error {
			// 1. if 'commit's ID is in 'curXT's provenance, then we'll add the new
			//    span to this trace
			// 2. if 'curXT's commit IDs are all in 'commit's provenance, then remove
			//    'curXT' from etcd after this is done
			allCommitsHaveBeenTraced := true
			for _, c := range curXT.CommitIDs {
				if commit.Commit.ID == c {
					if xt.isValid() {
						log.Errorf("multiple traces found for commit %q", commit.Commit.ID)
					}
					xt = curXT
				}
				if _, ok := provCommits[c]; !ok {
					allCommitsHaveBeenTraced = false
				}
			}

			// Delete 'curXT' if all of its commits are in 'commit's provenance
			if allCommitsHaveBeenTraced {
				if _, err := col.NewSTM(ctx, c, func(stm col.STM) error {
					return TracesCol(c).ReadWrite(stm).Delete(key)
				}); err != nil {
					log.Errorf("error deleting trace for %q: %v", commit.Commit.ID, err)
				}
			}
			return nil
		}); err != nil {
		log.Errorf("error getting trace via pipeline %q: %v", commit.Commit.Repo.Name, err)
		return nil, ctx
	}
	if !xt.isValid() {
		return nil, ctx // no trace found
	}

	// Create new opentracing span from 'xt'
	spanCtx, err := opentracing.GlobalTracer().Extract(opentracing.TextMap,
		opentracing.TextMapCarrier(xt.SerializedTrace))
	if err != nil {
		log.Errorf("could not extract span context from ExtendedTrace proto: %v", err)
		return nil, ctx
	}

	// return new span
	return opentracing.StartSpanFromContext(ctx,
		"worker.ProcessCommit", opentracing.FollowsFrom(spanCtx),
		opentracing.Tag{"commit", commit.Commit.ID})
}
