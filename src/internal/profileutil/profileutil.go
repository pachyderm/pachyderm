// Profileutil contains functionality to export performance information to external systems.
package profileutil

import (
	"context"

	"cloud.google.com/go/profiler"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachconfig"
	"github.com/pachyderm/pachyderm/v2/src/version"
	"go.uber.org/zap"
)

// StartCloudProfiler enables Google Cloud Profiler and begins exporting profiles, if
// GoogleCloudProfilerProject is set in the provided configuration.  Configuration information is
// read from the instance's GCE metadata server, so this requires running on GCP.  Docs:
// https://cloud.google.com/profiler/docs
//
// Service is the name of this binary (pachd, worker, etc.).
//
// If there is a problem starting the cloud profiler, it logs a message but we continue.
func StartCloudProfiler(ctx context.Context, service string, config *pachconfig.Configuration) {
	if config == nil {
		log.Error(ctx, "nil configuration passed to StartCloudProfiler; profiling not enabled")
		return
	}
	p := config.GoogleCloudProfilerProject
	if p == "" {
		return
	}
	log.Info(ctx, "enabling google cloud profiler", zap.String("project", p))
	err := profiler.Start(profiler.Config{
		Service:        service,
		ServiceVersion: version.PrettyPrintVersion(version.Version),
		MutexProfiling: true,
		DebugLogging:   false,
		ProjectID:      p,
	})
	if err != nil {
		log.Error(ctx, "failed to start cloud profiler; profiling not enabled", zap.Error(err))
	}
}
