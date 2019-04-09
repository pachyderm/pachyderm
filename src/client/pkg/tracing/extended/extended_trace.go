package extended

import (
	"context"
	"encoding/base64"
	"fmt"

	etcd "github.com/coreos/etcd/clientv3"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
	"google.golang.org/grpc/metadata"
)

const (
	// TraceCtxKey is the grpc metadata key whose value is a ExtendedTrace
	// identifying the current RPC/commit
	TraceCtxKey = "commit-trace"

	// TracesCollectionPrefix is the prefix associated with the 'traces'
	// collection in etcd (which maps pipelines and commits to extended traces)
	TracesCollectionPrefix = "commit_traces"
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
