// Profileutil contains functionality to export performance information to external systems.
package profileutil

import (
	"cloud.google.com/go/profiler"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	"github.com/pachyderm/pachyderm/v2/src/version"
	log "github.com/sirupsen/logrus"
)

// StartCloudProfiler enables Google Cloud Profiler and begins exporting profiles, if
// GoogleCloudProfilerProject is set in the provided configuration.  Configuration information is
// read from the instance's GCE metadata server, so this requires running on GCP.  Docs:
// https://cloud.google.com/profiler/docs
//
// Service is the name of this binary (pachd, worker, etc.).
func StartCloudProfiler(service string, config *serviceenv.Configuration) error {
	if config == nil {
		log.Warn("nil configuration passed to StartCloudProfiler; profiling not enabled")
		return nil
	}
	p := config.GoogleCloudProfilerProject
	if p == "" {
		return nil
	}
	log.Debugf("enabling google cloud profiler; sending profiles to project %q", p)
	return profiler.Start(profiler.Config{
		Service:        service,
		ServiceVersion: version.PrettyPrintVersion(version.Version),
		MutexProfiling: true,
		DebugLogging:   false,
		ProjectID:      p,
	})
}
