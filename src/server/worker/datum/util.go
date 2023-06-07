package datum

import (
	"os"
	"path/filepath"

	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"google.golang.org/protobuf/types/known/durationpb"
)

// MergeStats merges two stats.
func MergeStats(x, y *Stats) {
	MergeProcessStats(x.ProcessStats, y.ProcessStats)
	x.Processed += y.Processed
	x.Skipped += y.Skipped
	x.Failed += y.Failed
	x.Recovered += y.Recovered
	if x.FailedId == "" {
		x.FailedId = y.FailedId
	}
}

// MergeProcessStats merges two process stats.
func MergeProcessStats(x, y *pps.ProcessStats) {
	x.DownloadTime = durationpb.New(x.DownloadTime.AsDuration() + y.DownloadTime.AsDuration())
	x.ProcessTime = durationpb.New(x.ProcessTime.AsDuration() + y.ProcessTime.AsDuration())
	x.UploadTime = durationpb.New(x.UploadTime.AsDuration() + y.UploadTime.AsDuration())
	x.DownloadBytes += y.DownloadBytes
	x.UploadBytes += y.UploadBytes
}

func WithCreateFileSet(pachClient *client.APIClient, name string, cb func(*Set) error) (string, error) {
	resp, err := pachClient.WithCreateFileSetClient(func(mf client.ModifyFile) error {
		storageRoot := filepath.Join(os.TempDir(), name, uuid.NewWithoutDashes())
		return WithSet(nil, storageRoot, cb, WithMetaOutput(mf))
	})
	if err != nil {
		return "", err
	}
	return resp.FileSetId, nil
}

func CreateEmptyFileSet(pachClient *client.APIClient) (string, error) {
	return WithCreateFileSet(pachClient, "pachyderm-datums-empty", func(_ *Set) error { return nil })
}
