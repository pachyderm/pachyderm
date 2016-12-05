package server

import (
	"context"

	ppsclient "github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pps/persist"

	"go.pedge.io/pb/go/google/protobuf"
)

var (
	// DefaultGCPolicy is the default GC policy used by a pipeline if one is not
	// specified.
	DefaultGCPolicy = &ppsclient.GCPolicy{
		// a day
		Success: &google_protobuf.Duration{
			Seconds: 24 * 60 * 60,
		},
		// a week
		Failure: &google_protobuf.Duration{
			Seconds: 7 * 24 * 60 * 60,
		},
	}
)

func runGC(ctx context.Context, client persist.APIClient, pipelineInfo *ppsclient.PipelineInfo) {
	for {
	}
	panic("unreachable")
}
