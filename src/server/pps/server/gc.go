package server

import (
	"github.com/gogo/protobuf/types"
	ppsclient "github.com/pachyderm/pachyderm/src/client/pps"
)

var (
	// DefaultGCPolicy is the default GC policy used by a pipeline if one is not
	// specified.
	DefaultGCPolicy = &ppsclient.GCPolicy{
		// a day
		Success: &types.Duration{
			Seconds: 24 * 60 * 60,
		},
		// a week
		Failure: &types.Duration{
			Seconds: 7 * 24 * 60 * 60,
		},
	}
)
