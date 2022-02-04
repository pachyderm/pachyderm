package taskprotos

import (
	// storage and compaction tasks
	_ "github.com/pachyderm/pachyderm/v2/src/server/pfs/server"
	// worker datum tasks
	_ "github.com/pachyderm/pachyderm/v2/src/server/worker/pipeline/transform"
)
