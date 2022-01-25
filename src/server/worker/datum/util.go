package datum

import (
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/pps"
)

// MergeStats merges two stats.
func MergeStats(x, y *Stats) error {
	if err := MergeProcessStats(x.ProcessStats, y.ProcessStats); err != nil {
		return err
	}
	x.Processed += y.Processed
	x.Skipped += y.Skipped
	x.Failed += y.Failed
	x.Recovered += y.Recovered
	if x.FailedID == "" {
		x.FailedID = y.FailedID
	}
	return nil
}

// MergeProcessStats merges two process stats.
func MergeProcessStats(x, y *pps.ProcessStats) error {
	var err error
	if x.DownloadTime, err = plusDuration(x.DownloadTime, y.DownloadTime); err != nil {
		return err
	}
	if x.ProcessTime, err = plusDuration(x.ProcessTime, y.ProcessTime); err != nil {
		return err
	}
	if x.UploadTime, err = plusDuration(x.UploadTime, y.UploadTime); err != nil {
		return err
	}
	x.DownloadBytes += y.DownloadBytes
	x.UploadBytes += y.UploadBytes
	return nil
}

func plusDuration(x *types.Duration, y *types.Duration) (*types.Duration, error) {
	var xd time.Duration
	var yd time.Duration
	var err error
	if x != nil {
		xd, err = types.DurationFromProto(x)
		if err != nil {
			return nil, errors.EnsureStack(err)
		}
	}
	if y != nil {
		yd, err = types.DurationFromProto(y)
		if err != nil {
			return nil, errors.EnsureStack(err)
		}
	}
	return types.DurationProto(xd + yd), nil
}
