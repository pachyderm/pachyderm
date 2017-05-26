// Package rc provides helper functions for managing replication controllers
// for PPS pipelines.
package rc

import (
	"fmt"
	"strings"
)

// PipelineRcName generates the name of the k8s replication controller that
// manages a pipeline's workers
func PipelineRcName(name string, version uint64) string {
	// k8s won't allow RC names that contain upper-case letters
	// or underscores
	// TODO: deal with name collision
	name = strings.Replace(name, "_", "-", -1)
	return fmt.Sprintf("pipeline-%s-v%d", strings.ToLower(name), version)
}

// JobRcName generates the name of the k8s replication controller that manages
// an orphan job's workers
func JobRcName(id string) string {
	// k8s won't allow RC names that contain upper-case letters
	// or underscores
	// TODO: deal with name collision
	id = strings.Replace(id, "_", "-", -1)
	return fmt.Sprintf("job-%s", strings.ToLower(id))
}
